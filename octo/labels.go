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
	"strings"

	"github.com/paketo-buildpacks/pipeline-builder/octo/actions"
	"github.com/paketo-buildpacks/pipeline-builder/octo/actions/event"
	"github.com/paketo-buildpacks/pipeline-builder/octo/labels"
)

func ContributeLabels(descriptor Descriptor) ([]Contribution, error) {
	var contributions []Contribution

	l := []labels.Label{
		{
			Name:        "semver:major",
			Description: "A change requiring a major version bump",
			Color:       "f9d0c4",
		},
		{
			Name:        "semver:minor",
			Description: "A change requiring a minor version bump",
			Color:       "f9d0c4",
		},
		{
			Name:        "semver:patch",
			Description: "A change requiring a patch version bump",
			Color:       "f9d0c4",
		},
		{
			Name:        "type:bug",
			Description: "A general bug",
			Color:       "e3d9fc",
		},
		{
			Name:        "type:dependency-upgrade",
			Description: "A dependency upgrade",
			Color:       "e3d9fc",
		},
		{
			Name:        "type:documentation",
			Description: "A documentation update",
			Color:       "e3d9fc",
		},
		{
			Name:        "type:enhancement",
			Description: "A general enhancement",
			Color:       "e3d9fc",
		},
		{
			Name:        "type:question",
			Description: "A user question",
			Color:       "e3d9fc",
		},
		{
			Name:        "type:task",
			Description: "A general task",
			Color:       "e3d9fc",
		},
	}

	if c, err := NewLabelsContribution(l); err != nil {
		return nil, err
	} else {
		contributions = append(contributions, c)
	}

	w := actions.Workflow{
		Name: "Synchronize Labels",
		On: map[event.Type]event.Event{
			event.PushType: event.Push{
				Branches: []string{"main"},
				Paths:    []string{filepath.Join(".github", "labels.yml")},
			},
		},
		Jobs: map[string]actions.Job{
			"synchronize": {
				Name:   "Synchronize Labels",
				RunsOn: []actions.VirtualEnvironment{actions.UbuntuLatest},
				Steps: []actions.Step{
					{
						Uses: "actions/checkout@v3",
					},
					{
						Uses: "micnncim/action-label-syncer@v1",
						Env:  map[string]string{"GITHUB_TOKEN": descriptor.GitHub.Token},
					},
				},
			},
		},
	}

	if c, err := NewActionContribution(w); err != nil {
		return nil, err
	} else {
		contributions = append(contributions, c)
	}

	w = actions.Workflow{
		Name: "Minimal Labels",
		On: map[event.Type]event.Event{
			event.PullRequestType: event.PullRequest{
				Types: []event.PullRequestActivityType{
					event.PullRequestSynchronize,
					event.PullRequestReopened,
					event.PullRequestLabeled,
					event.PullRequestUnlabeled,
				},
			},
		},
		Jobs: map[string]actions.Job{
			"semver": {
				Name:   "Minimal Semver Labels",
				RunsOn: []actions.VirtualEnvironment{actions.UbuntuLatest},
				Steps: []actions.Step{
					{
						Uses: "mheap/github-action-required-labels@v2",
						With: map[string]interface{}{
							"mode":  "exactly",
							"count": 1,
							"labels": strings.Join([]string{
								"semver:major",
								"semver:minor",
								"semver:patch",
							}, ", "),
						},
					},
				},
			},
			"type": {
				Name:   "Minimal Type Labels",
				RunsOn: []actions.VirtualEnvironment{actions.UbuntuLatest},
				Steps: []actions.Step{
					{
						Uses: "mheap/github-action-required-labels@v2",
						With: map[string]interface{}{
							"mode":  "exactly",
							"count": 1,
							"labels": strings.Join([]string{
								"type:bug",
								"type:dependency-upgrade",
								"type:documentation",
								"type:enhancement",
								"type:question",
								"type:task",
							}, ", "),
						},
					},
				},
			},
		},
	}

	if c, err := NewActionContribution(w); err != nil {
		return nil, err
	} else {
		contributions = append(contributions, c)
	}

	return contributions, nil
}
