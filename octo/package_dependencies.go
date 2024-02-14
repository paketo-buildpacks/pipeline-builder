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
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pelletier/go-toml"

	"github.com/paketo-buildpacks/pipeline-builder/octo/actions"
	"github.com/paketo-buildpacks/pipeline-builder/octo/actions/event"
	_package "github.com/paketo-buildpacks/pipeline-builder/octo/package"
)

func ContributePackageDependencies(descriptor Descriptor) ([]Contribution, error) {
	var contributions []Contribution

	file := filepath.Join(descriptor.Path, "package.toml")
	b, err := os.ReadFile(file)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to read %s\n%w", file, err)
	}
	var p _package.Package
	if err := toml.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("unable to decode %s\n%w", file, err)
	}

	file = filepath.Join(descriptor.Path, "buildpack.toml")
	b, err = os.ReadFile(file)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to read %s\n%w", file, err)
	}
	var bpGroups _package.BuildpackOrderGroups
	if err := toml.Unmarshal(b, &bpGroups); err != nil {
		return nil, fmt.Errorf("unable to decode %s\n%w", file, err)
	}

	for _, d := range p.Dependencies {
		pkgId, bpId, err := findIds(bpGroups, d)
		if err != nil {
			return nil, err
		}

		if c, err := contributePackageDependency(descriptor, pkgId, bpId); err != nil {
			return nil, err
		} else {
			contributions = append(contributions, c)
		}
	}

	return contributions, nil
}

// findIds will return the pkg group id (used in package.toml) and the buildpack id (used in buildpack.toml)
//
//	  it will do some fuzzy matching because there is not a strict relationship between the two
//	    - often pkg group id is `repo.io/buildpack/id:version`, so you can extract `buildpack/id`
//	    - if that does not exist as a build pack id, then we search for just the last part of
//	        buildpack id and a matching version
//
//	        for example, `gcr.io/tanzu-buildpacks/bellsoft-liberica:1.2.3`, we'll look for `bellsoft-liberica`
//	           and version `1.2.3` in buildpack.toml
//
//	           if there is a match, then we return the found buildpack id, let's say it finds `paketo-buildpacks/bellsoft-liberica`
//				  then we return `gcr.io/paketo-buildpacks/bellsoft-liberica:1.2.3`
func findIds(bpOrders _package.BuildpackOrderGroups, dep _package.Dependency) (string, string, error) {
	re := regexp.MustCompile(`^(?:.+://)?(.+?)/(.+):([^:]+)$`)
	if g := re.FindStringSubmatch(dep.URI); g == nil {
		return "", "", fmt.Errorf("unable to parse image coordinates from %s", dep.URI)
	} else {
		registry := g[1]
		possibleBpId := g[2]
		version := g[3]

		// search for a direct match
		for _, order := range bpOrders.Orders {
			for _, group := range order.Groups {
				if group.ID == possibleBpId && group.Version == version {
					foundId := fmt.Sprintf("%s/%s", registry, possibleBpId)
					return foundId, foundId, nil
				}
			}
		}

		// search for a fuzzy match
		for _, order := range bpOrders.Orders {
			for _, group := range order.Groups {
				endOfId := strings.Split(possibleBpId, "/")[1]
				// fmt.Println("group.Id", group.ID, "endOfId", endOfId, "group.Version", group.Version, "version", version)
				if strings.HasSuffix(group.ID, endOfId) && group.Version == version {
					pkgId := fmt.Sprintf("%s/%s", registry, possibleBpId)
					bpId := fmt.Sprintf("%s/%s", registry, group.ID)
					// fmt.Println("pkgId", pkgId, "bpId", bpId)
					return pkgId, bpId, nil
				}
			}
		}

		return "", "", fmt.Errorf("unable to match image coordinates from [%s, %s, %s]", registry, possibleBpId, version)
	}
}

func contributePackageDependency(descriptor Descriptor, name string, bpId string) (Contribution, error) {
	w := actions.Workflow{
		Name: fmt.Sprintf("Update %s", filepath.Base(name)),
		On: map[event.Type]event.Event{
			event.ScheduleType:         event.Schedule{{Minute: "0", Hour: "4", DayOfWeek: "4-5"}},
			event.WorkflowDispatchType: event.WorkflowDispatch{},
		},
		Jobs: map[string]actions.Job{
			"update": {
				Name:   "Update Package Dependency",
				RunsOn: []actions.VirtualEnvironment{actions.UbuntuLatest},
				Steps: []actions.Step{
					{
						Uses: "actions/setup-go@v5",
						With: map[string]interface{}{"go-version": GoVersion},
					},
					{
						Name: "Install update-package-dependency",
						Run:  StatikString("/install-update-package-dependency.sh"),
					},
					{
						Name: "Install crane",
						Run:  StatikString("/install-crane.sh"),
						Env:  map[string]string{"CRANE_VERSION": CraneVersion},
					},
					{
						Name: "Install yj",
						Run:  StatikString("/install-yj.sh"),
						Env:  map[string]string{"YJ_VERSION": YJVersion},
					},
					{
						Uses: "actions/checkout@v4",
					},
					{
						Id:   "package",
						Name: "Update Package Dependency",
						Run:  StatikString("/update-package-dependency.sh"),
						Env:  map[string]string{"DEPENDENCY": name},
					},
					{
						Uses: "peter-evans/create-pull-request@v6",
						With: map[string]interface{}{
							"token":  descriptor.GitHub.Token,
							"author": fmt.Sprintf("%[1]s <%[1]s@users.noreply.github.com>", descriptor.GitHub.Username),
							"commit-message": fmt.Sprintf(`Bump %[1]s from ${{ steps.package.outputs.old-version }} to ${{ steps.package.outputs.new-version }}

Bumps %[1]s from ${{ steps.package.outputs.old-version }} to ${{ steps.package.outputs.new-version }}.`, name),
							"signoff":       true,
							"branch":        fmt.Sprintf("update/package/%s", filepath.Base(name)),
							"delete-branch": true,
							"title":         fmt.Sprintf("Bump %s from ${{ steps.package.outputs.old-version }} to ${{ steps.package.outputs.new-version }}", name),
							"body":          fmt.Sprintf("Bumps [`%[1]s`](https://%[1]s) from [`${{ steps.package.outputs.old-version }}`](https://%[1]s:${{ steps.package.outputs.old-version }}) to [`${{ steps.package.outputs.new-version }}`](https://%[1]s:${{ steps.package.outputs.new-version }}).", name),
							"labels":        "${{ steps.package.outputs.version-label }}, type:dependency-upgrade",
						},
					},
				},
			},
		},
	}

	j := w.Jobs["update"]
	j.Steps = append(NewDockerCredentialActions(descriptor.DockerCredentials), j.Steps...)
	w.Jobs["update"] = j

	if name != bpId {
		for i := 0; i < len(w.Jobs["update"].Steps); i++ {
			if w.Jobs["update"].Steps[i].Id == "package" {
				w.Jobs["update"].Steps[i].Env["BP_DEPENDENCY"] = bpId
				break
			}
		}
	}

	return NewActionContributionWithNamespace(Namespace, w)
}
