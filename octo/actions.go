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

	"github.com/paketo-buildpacks/pipeline-builder/octo/actions"
	"github.com/paketo-buildpacks/pipeline-builder/octo/actions/event"
)

func ContributeActions(descriptor Descriptor) ([]Contribution, error) {
	var contributions []Contribution

	for _, a := range descriptor.Actions {
		w := actions.Workflow{
			Name: fmt.Sprintf("Create Action %s", a.Source),
			On: map[event.Type]event.Event{
				event.PullRequestType: event.PullRequest{
					Paths: []string{
						"actions/*",
						fmt.Sprintf("actions/%s/*", a.Source),
					},
				},
				event.PushType: event.Push{
					Branches: []string{"main"},
					Paths: []string{
						"actions/*",
						fmt.Sprintf("actions/%s/*", a.Source),
					},
				},
				event.ReleaseType: event.Release{
					Types: []event.ReleaseActivityType{event.ReleasePublished},
				},
			},
			Jobs: map[string]actions.Job{
				"create-action": {
					Name:   "Create Action",
					RunsOn: []actions.VirtualEnvironment{actions.UbuntuLatest},
					Steps: []actions.Step{
						{
							Uses: "actions/checkout@v3",
						},
						{
							Name: "Create Action",
							Run:  StatikString("/create-action.sh"),
							Env: map[string]string{
								"PUSH":    "${{ github.event_name != 'pull_request' }}",
								"SOURCE":  a.Source,
								"TARGET":  a.Target,
								"VERSION": "main",
							},
						},
					},
				},
			},
		}

		j := w.Jobs["create-action"]
		j.Steps = append(NewDockerCredentialActions(descriptor.DockerCredentials), j.Steps...)
		w.Jobs["create-action"] = j

		if c, err := NewActionContribution(w); err != nil {
			return nil, err
		} else {
			contributions = append(contributions, c)
		}
	}

	return contributions, nil
}
