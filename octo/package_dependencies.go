/*
 * Copyright 2018-2020 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package octo

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/pelletier/go-toml"

	"github.com/paketo-buildpacks/pipeline-builder/octo/actions"
	"github.com/paketo-buildpacks/pipeline-builder/octo/actions/event"
	_package "github.com/paketo-buildpacks/pipeline-builder/octo/package"
)

func ContributePackageDependencies(descriptor Descriptor) ([]Contribution, error) {
	var contributions []Contribution

	file := filepath.Join(descriptor.Path, "package.toml")
	b, err := ioutil.ReadFile(file)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to read %s\n%w", file, err)
	}
	var p _package.Package
	if err := toml.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("unable to decode %s\n%w", file, err)
	}

	re := regexp.MustCompile(`^(.+):[^:]+$`)
	for _, d := range p.Dependencies {
		if g := re.FindStringSubmatch(d.Image); g == nil {
			return nil, fmt.Errorf("unable to parse image coordinates from %s", d.Image)
		} else {
			if c, err := contributePackageDependency(descriptor, g[1]); err != nil {
				return nil, err
			} else {
				contributions = append(contributions, c)
			}
		}
	}

	return contributions, nil
}

func contributePackageDependency(descriptor Descriptor, name string) (Contribution, error) {
	w := actions.Workflow{
		Name: fmt.Sprintf("Update %s", filepath.Base(name)),
		On: map[event.Type]event.Event{
			event.ScheduleType:         event.Schedule{{Minute: "0", Hour: "12-23", DayOfWeek: "1-5"}},
			event.WorkflowDispatchType: event.WorkflowDispatch{},
		},
		Jobs: map[string]actions.Job{
			"update": {
				Name:   "Update Package Dependency",
				RunsOn: []actions.VirtualEnvironment{actions.UbuntuLatest},
				Steps: []actions.Step{
					{
						Uses: "actions/setup-go@v2",
						With: map[string]interface{}{"go-version": GoVersion},
					},
					{
						Name: "Install update-package-dependency",
						Run:  statikString("/install-update-package-dependency.sh"),
					},
					{
						Name: "Install crane",
						Run:  statikString("/install-crane.sh"),
						Env:  map[string]string{"CraneVersion": CraneVersion},
					},
					{
						Name: "Install yj",
						Run:  statikString("/install-yj.sh"),
						Env:  map[string]string{"YJ_VERSION": YJVersion},
					},
					{
						Uses: "actions/checkout@v2",
					},
					{
						Id:   "package",
						Name: "Update Package Dependency",
						Run:  statikString("/update-package-dependency.sh"),
						Env:  map[string]string{"DEPENDENCY": name},
					},
					{
						Uses: "peter-evans/create-pull-request@v3",
						With: map[string]interface{}{
							"token": descriptor.GitHubToken,
							"commit-message": fmt.Sprintf(`Bump %[1]s from ${{ steps.package.outputs.old-version }} to ${{ steps.package.outputs.new-version }}

Bumps %[1]s from ${{ steps.package.outputs.old-version }} to ${{ steps.package.outputs.new-version }}.`, name),
							"signoff":       true,
							"branch":        fmt.Sprintf("update/package/%s", filepath.Base(name)),
							"delete-branch": true,
							"title":         fmt.Sprintf("Bump %s from ${{ steps.package.outputs.old-version }} to ${{ steps.package.outputs.new-version }}", name),
							"body":          fmt.Sprintf("Bumps [`%[1]s`](https://%[1]s) from [`${{ steps.package.outputs.old-version }}`](https://%[1]s:${{ steps.package.outputs.old-version }}) to [`${{ steps.package.outputs.new-version }}`](https://%[1]s:${{ steps.package.outputs.new-version }}).", name),
							"labels":        "semver:minor, type:dependency-upgrade",
						},
					},
				},
			},
		},
	}

	j := w.Jobs["update"]
	j.Steps = append(NewDockerCredentialActions(descriptor.DockerCredentials), j.Steps...)
	w.Jobs["update"] = j

	return NewActionContribution(w)
}
