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

	"github.com/iancoleman/strcase"

	"github.com/paketo-buildpacks/pipeline-builder/octo/actions"
	"github.com/paketo-buildpacks/pipeline-builder/octo/actions/event"
)

func ContributeBuildpackDependencies(descriptor Descriptor) ([]Contribution, error) {
	var contributions []Contribution

	for _, d := range descriptor.Dependencies {
		w := actions.Workflow{
			Name: fmt.Sprintf("Update %s", d.Name),
			On: map[event.Type]event.Event{
				event.ScheduleType:         event.Schedule{{Minute: "0", Hour: "5", DayOfWeek: "1-5"}},
				event.WorkflowDispatchType: event.WorkflowDispatch{},
			},
			Jobs: map[string]actions.Job{
				"update": {
					Name:   "Update Buildpack Dependency",
					RunsOn: []actions.VirtualEnvironment{actions.UbuntuLatest},
					Steps: []actions.Step{
						{
							Uses: "actions/setup-go@v4",
							With: map[string]interface{}{"go-version": GoVersion},
						},
						{
							Name: "Install update-buildpack-dependency",
							Run:  StatikString("/install-update-buildpack-dependency.sh"),
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
							Id:   "buildpack",
							Name: "Update Buildpack Dependency",
							Run:  StatikString("/update-buildpack-dependency.sh"),
							Env: map[string]string{
								"ID":              d.Id,
								"ARCH":            safeGetArch(d),
								"SHA256":          "${{ steps.dependency.outputs.sha256 }}",
								"URI":             "${{ steps.dependency.outputs.uri }}",
								"VERSION":         "${{ steps.dependency.outputs.version }}",
								"VERSION_PATTERN": d.VersionPattern,
								"CPE":             "${{ steps.dependency.outputs.cpe }}",
								"CPE_PATTERN":     d.CPEPattern,
								"PURL":            "${{ steps.dependency.outputs.purl }}",
								"PURL_PATTERN":    d.PURLPattern,
								"SOURCE_URI":      "${{ steps.dependency.outputs.source }}",
								"SOURCE_SHA256":      "${{ steps.dependency.outputs.source_sha256 }}",
							},
						}, {
							Uses: "peter-evans/create-pull-request@v5",
							With: map[string]interface{}{
								"token":  descriptor.GitHub.Token,
								"author": fmt.Sprintf("%[1]s <%[1]s@users.noreply.github.com>", descriptor.GitHub.Username),
								"commit-message": fmt.Sprintf(`Bump %[1]s from ${{ steps.buildpack.outputs.old-version }} to ${{ steps.buildpack.outputs.new-version }}

Bumps %[1]s from ${{ steps.buildpack.outputs.old-version }} to ${{ steps.buildpack.outputs.new-version }}.`, d.Name),
								"signoff":       true,
								"branch":        fmt.Sprintf("update/buildpack/%s", strcase.ToKebab(d.Name)),
								"delete-branch": true,
								"title":         fmt.Sprintf("Bump %s from ${{ steps.buildpack.outputs.old-version }} to ${{ steps.buildpack.outputs.new-version }}", d.Name),
								"body":          fmt.Sprintf("Bumps `%[1]s` from `${{ steps.buildpack.outputs.old-version }}` to `${{ steps.buildpack.outputs.new-version }}`.", d.Name),
								"labels":        "${{ steps.buildpack.outputs.version-label }}, type:dependency-upgrade",
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

func safeGetArch(d Dependency) string {
	if arch, ok := d.With["arch"]; ok {
		return arch.(string)
	}

	return ""
}
