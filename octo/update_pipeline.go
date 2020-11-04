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
	"path/filepath"

	"github.com/paketo-buildpacks/pipeline-builder/octo/actions"
	"github.com/paketo-buildpacks/pipeline-builder/octo/actions/event"
)

func ContributeUpdatePipeline(descriptor Descriptor) (Contribution, error) {
	w := actions.Workflow{
		Name: "Update Pipeline",
		On: map[event.Type]event.Event{
			event.ScheduleType:         event.Schedule{{Minute: "0", Hour: "5", DayOfWeek: "1-5"}},
			event.WorkflowDispatchType: event.WorkflowDispatch{},
		},
		Jobs: map[string]actions.Job{
			"update": {
				Name:   "Update Pipeline",
				RunsOn: []actions.VirtualEnvironment{actions.UbuntuLatest},
				Steps: []actions.Step{
					{
						Uses: "actions/setup-go@v2",
						With: map[string]interface{}{"go-version": GoVersion},
					},
					{
						Name: "Install octo",
						Run:  StatikString("/install-octo.sh"),
					},
					{
						Uses: "actions/checkout@v2",
					},
					{
						Id:   "pipeline",
						Name: "Update Pipeline",
						Run:  StatikString("/update-pipeline.sh"),
						Env: map[string]string{
							"GITHUB_TOKEN": descriptor.GitHubToken,
							"DESCRIPTOR":   filepath.Join(".github", "pipeline-descriptor.yml"),
						},
					},
					{
						Uses: "peter-evans/create-pull-request@v3",
						With: map[string]interface{}{
							"token": descriptor.GitHubToken,
							"commit-message": `Bump pipeline from ${{ steps.pipeline.outputs.old-version }} to ${{ steps.pipeline.outputs.new-version }}

Bumps pipeline from ${{ steps.pipeline.outputs.old-version }} to ${{ steps.pipeline.outputs.new-version }}.`,
							"signoff":       true,
							"branch":        "update/pipeline",
							"delete-branch": true,
							"title":         "Bump pipeline from ${{ steps.pipeline.outputs.old-version }} to ${{ steps.pipeline.outputs.new-version }}",
							"body":          "Bumps pipeline from `${{ steps.pipeline.outputs.old-version }}` to `${{ steps.pipeline.outputs.new-version }}`.\n\n<details>\n<summary>Release Notes</summary>\n${{ steps.pipeline.outputs.release-notes }}\n</details>",
							"labels":        "semver:patch, type:task",
						},
					},
				},
			},
		},
	}

	return NewActionContribution(w)
}
