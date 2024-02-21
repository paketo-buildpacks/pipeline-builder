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
						Uses: "actions/setup-go@v5",
						With: map[string]interface{}{"go-version": GoVersion},
					},
					{
						Uses: fmt.Sprintf("buildpacks/github-actions/setup-tools@v%s", BuildpackActionsVersion),
						With: map[string]interface{}{
							"crane-version": CraneVersion,
							"yj-version":    YJVersion,
						},
					},
					{
						Uses: fmt.Sprintf("buildpacks/github-actions/setup-pack@v%s", BuildpackActionsVersion),
						With: map[string]interface{}{
							"pack-version": PackVersion,
						},
					},
					{
						Uses: "actions/checkout@v4",
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
