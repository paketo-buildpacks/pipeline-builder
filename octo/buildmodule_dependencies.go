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
	"strconv"

	"github.com/iancoleman/strcase"

	"github.com/paketo-buildpacks/pipeline-builder/octo/actions"
	"github.com/paketo-buildpacks/pipeline-builder/octo/actions/event"
)

func ContributeBuildModuleDependencies(descriptor Descriptor) ([]Contribution, error) {
	var contributions []Contribution

	for _, d := range descriptor.Dependencies {

		// Is this a buildpack or an extension?
		bpfile := filepath.Join(descriptor.Path, "buildpack.toml")
		extnfile := filepath.Join(descriptor.Path, "extension.toml")
		extension := false
		if _, err := os.Stat(bpfile); err == nil {
			extension = false
		} else if _, err := os.Stat(extnfile); err == nil {
			extension = true
		} else {
			return nil, fmt.Errorf("unable to read buildpack/extension.toml at %s\n", descriptor.Path)
		}

		w := actions.Workflow{
			Name: fmt.Sprintf("Update %s", d.Name),
			On: map[event.Type]event.Event{
				event.ScheduleType:         event.Schedule{{Minute: "0", Hour: "5", DayOfWeek: "1-5"}},
				event.WorkflowDispatchType: event.WorkflowDispatch{},
			},
			Jobs: map[string]actions.Job{
				"update": {
					Name:   "Update Build Module Dependency",
					RunsOn: []actions.VirtualEnvironment{actions.UbuntuLatest},
					Steps: []actions.Step{
						{
							Uses: "actions/setup-go@v4",
							With: map[string]interface{}{"go-version": GoVersion},
						},
						{
							Name: "Install update-buildmodule-dependency",
							Run:  StatikString("/install-update-buildmodule-dependency.sh"),
						},
						{
							Name: "Install yj",
							Run:  StatikString("/install-yj.sh"),
							Env:  map[string]string{"YJ_VERSION": YJVersion},
						},
						{
							Uses: "actions/checkout@v3",
						},
						{
							Id:   "dependency",
							Uses: d.Uses,
							With: d.With,
						},
						{
							Id:   "buildmodule",
							Name: "Update BuildModule Dependency",
							Run:  StatikString("/update-buildmodule-dependency.sh"),
							Env: map[string]string{
								"EXTENSION":       strconv.FormatBool(extension),
								"ID":              d.Id,
								"SHA256":          "${{ steps.dependency.outputs.sha256 }}",
								"URI":             "${{ steps.dependency.outputs.uri }}",
								"VERSION":         "${{ steps.dependency.outputs.version }}",
								"VERSION_PATTERN": d.VersionPattern,
								"CPE":             "${{ steps.dependency.outputs.cpe }}",
								"CPE_PATTERN":     d.CPEPattern,
								"PURL":            "${{ steps.dependency.outputs.purl }}",
								"PURL_PATTERN":    d.PURLPattern,
							},
						}, {
							Uses: "peter-evans/create-pull-request@v5",
							With: map[string]interface{}{
								"token":  descriptor.GitHub.Token,
								"author": fmt.Sprintf("%[1]s <%[1]s@users.noreply.github.com>", descriptor.GitHub.Username),
								"commit-message": fmt.Sprintf(`Bump %[1]s from ${{ steps.buildmodule.outputs.old-version }} to ${{ steps.buildmodule.outputs.new-version }}

Bumps %[1]s from ${{ steps.buildmodule.outputs.old-version }} to ${{ steps.buildmodule.outputs.new-version }}.`, d.Name),
								"signoff":       true,
								"branch":        fmt.Sprintf("update/buildmodule/%s", strcase.ToKebab(d.Name)),
								"delete-branch": true,
								"title":         fmt.Sprintf("Bump %s from ${{ steps.buildmodule.outputs.old-version }} to ${{ steps.buildmodule.outputs.new-version }}", d.Name),
								"body":          fmt.Sprintf("Bumps `%[1]s` from `${{ steps.buildpack.outputs.old-version }}` to `${{ steps.buildmodule.outputs.new-version }}`.", d.Name),
								"labels":        "${{ steps.buildmodule.outputs.version-label }}, type:dependency-upgrade",
							},
						},
					},
				},
			},
		}

		if c, err := NewActionContributionWithNamespace(Namespace, w); err != nil {
			return nil, err
		} else {
			contributions = append(contributions, c)
		}
	}

	return contributions, nil
}
