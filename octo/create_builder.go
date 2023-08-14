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
	"github.com/paketo-buildpacks/pipeline-builder/v2/octo/actions"
	"github.com/paketo-buildpacks/pipeline-builder/v2/octo/actions/event"
)

func ContributeCreateBuilder(descriptor Descriptor) (*Contribution, error) {
	if descriptor.Builder == nil {
		return nil, nil
	}

	w := actions.Workflow{
		Name: "Create Builder",
		On: map[event.Type]event.Event{
			event.ReleaseType: event.Release{
				Types: []event.ReleaseActivityType{
					event.ReleasePublished,
				},
			},
		},
		Jobs: map[string]actions.Job{
			"create-builder": {
				Name:   "Create Builder",
				RunsOn: []actions.VirtualEnvironment{actions.UbuntuLatest},
				Steps: []actions.Step{
					{
						Uses: "actions/setup-go@v4",
						With: map[string]interface{}{"go-version": GoVersion},
					},
					{
						Name: "Install crane",
						Run:  StatikString("/install-crane.sh"),
						Env:  map[string]string{"CRANE_VERSION": CraneVersion},
					},
					{
						Name: "Install pack",
						Run:  StatikString("/install-pack.sh"),
						Env:  map[string]string{"PACK_VERSION": PackVersion},
					},
					{
						Uses: "actions/checkout@v3",
					},
					{
						Id:   "version",
						Name: "Compute Version",
						Run:  StatikString("/compute-version.sh"),
					},
					{
						Id:   "builder",
						Name: "Create Builder",
						Run:  StatikString("/create-builder.sh"),
						Env: map[string]string{
							"BUILDER": descriptor.Builder.Repository,
							"PUBLISH": "true",
							"VERSION": "${{ steps.version.outputs.version }}",
						},
					},
					{
						Name: "Update release with digest",
						Run:  StatikString("/update-release-digest.sh"),
						Env: map[string]string{
							"DIGEST":       "${{ steps.builder.outputs.digest }}",
							"GITHUB_TOKEN": descriptor.GitHub.Token,
						},
					},
				},
			},
		},
	}

	j := w.Jobs["create-builder"]
	j.Steps = append(NewDockerCredentialActions(descriptor.DockerCredentials), j.Steps...)
	w.Jobs["create-builder"] = j

	c, err := NewActionContributionWithNamespace(Namespace, w)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
