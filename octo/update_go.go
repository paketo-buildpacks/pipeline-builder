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

	"github.com/paketo-buildpacks/pipeline-builder/octo/actions"
	"github.com/paketo-buildpacks/pipeline-builder/octo/actions/event"
)

func ContributeUpdateGo(descriptor Descriptor) (*Contribution, error) {
	entries, err := os.ReadDir(descriptor.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to Find go.mod or go.sum files in %s\n%w", descriptor.Path, err)
	}

	goModFiles := []string{}
	for _, entry := range entries {
		if entry.Name() == "go.mod" || entry.Name() == "go.sum" {
			goModFiles = append(goModFiles, entry.Name())
		}
	}

	if len(goModFiles) != 2 {
		return nil, nil
	}

	w := actions.Workflow{
		Name: "Update Go",
		On: map[event.Type]event.Event{
			event.ScheduleType:         event.Schedule{{Minute: "0", Hour: "2", DayOfWeek: "1"}},
			event.WorkflowDispatchType: event.WorkflowDispatch{},
		},
		Jobs: map[string]actions.Job{
			"update": {
				Name:   "Update Go",
				RunsOn: []actions.VirtualEnvironment{actions.UbuntuLatest},
				Steps: []actions.Step{
					{
						Uses: "actions/setup-go@v3",
						With: map[string]interface{}{"go-version": GoVersion},
					},
					{
						Uses: "actions/checkout@v3",
					},
					{
						Id:   "update-go",
						Name: "Update Go Version",
						Run:  StatikString("/update-go.sh"),
						Env: map[string]string{
							"GO_VERSION": GoVersion,
						},
					},
					{
						Uses: "peter-evans/create-pull-request@v4",
						With: map[string]interface{}{
							"token":  descriptor.GitHub.Token,
							"author": fmt.Sprintf("%[1]s <%[1]s@users.noreply.github.com>", descriptor.GitHub.Username),
							"commit-message": `Bump Go from ${{ steps.update-go.outputs.old-go-version }} to ${{ steps.update-go.outputs.go-version }}

Bumps Go from ${{ steps.update-go.outputs.old-go-version }} to ${{ steps.update-go.outputs.go-version }}.`,
							"signoff":       true,
							"branch":        "update/go",
							"delete-branch": true,
							"title":         "Bump Go from ${{ steps.update-go.outputs.old-go-version }} to ${{ steps.update-go.outputs.go-version }}",
							"body":          "Bumps Go from `${{ steps.update-go.outputs.old-go-version }}` to `${{ steps.update-go.outputs.go-version }}`.\n\n<details>\n<summary>Release Notes</summary>\n${{ steps.pipeline.outputs.release-notes }}\n</details>",
							"labels":        "semver:minor, type:task",
						},
					},
				},
			},
		},
	}

	c, err := NewActionContributionWithNamespace(Namespace, w)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
