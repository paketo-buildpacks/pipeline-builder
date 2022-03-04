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
	"path/filepath"

	"github.com/paketo-buildpacks/pipeline-builder/octo/actions"
	"github.com/paketo-buildpacks/pipeline-builder/octo/actions/event"
)

func ContributeLitePackages(descriptor Descriptor) ([]Contribution, error) {
	var contributions []Contribution

	for _, o := range descriptor.RepublishImages {
		if c, err := contributeLitePackage(descriptor, o); err != nil {
			return nil, err
		} else {
			contributions = append(contributions, c)
		}
	}

	return contributions, nil
}

func contributeLitePackage(descriptor Descriptor, republishImage RepublishImage) (Contribution, error) {
	w := actions.Workflow{
		Name: fmt.Sprintf("Republish Image %s", filepath.Base(republishImage.Target)),
		On: map[event.Type]event.Event{
			event.ScheduleType:         event.Schedule{{Minute: "15", Hour: "12-23", DayOfWeek: "1-5"}},
			event.WorkflowDispatchType: event.WorkflowDispatch{},
		},
		Jobs: map[string]actions.Job{
			"republish-image": {
				Name:   "Republish Image with new ID",
				RunsOn: []actions.VirtualEnvironment{actions.UbuntuLatest},
				Steps: []actions.Step{
					{
						Name: "Install crane",
						Run:  StatikString("/install-crane.sh"),
						Env:  map[string]string{"CRANE_VERSION": CraneVersion},
					},
					{
						Uses: "actions/checkout@v3",
						With: map[string]interface{}{
							"repository":  republishImage.TargetRepo,
							"fetch-depth": 0,
						},
					},
					{
						Id:   "version",
						Name: "Check for next version",
						Run:  StatikString("/check-republish-version.sh"),
						Env: map[string]string{
							"SOURCE":     republishImage.Source,
							"TARGET":     republishImage.Target,
							"TAG_PREFIX": republishImage.TagPrefix,
						},
					},
					{
						Uses: "actions/setup-go@v2",
						If:   "${{ ! steps.version.outputs.skip }}",
						With: map[string]interface{}{"go-version": GoVersion},
					},
					{
						Name: "Install update-buildpack-image-id",
						If:   "${{ ! steps.version.outputs.skip }}",
						Run:  StatikString("/install-update-buildpack-image-id.sh"),
					},
					{
						Name: "Republish Image",
						If:   "${{ ! steps.version.outputs.skip }}",
						Run:  StatikString("/republish-image.sh"),
						Env: map[string]string{
							"SOURCE":         republishImage.Source,
							"SOURCE_VERSION": "${{ steps.version.outputs.source }}",
							"TARGET":         republishImage.Target,
							"TARGET_VERSION": "${{ steps.version.outputs.target }}",
							"NEWID":          republishImage.ID,
						},
					},
				},
			},
		},
	}

	j := w.Jobs["republish-image"]
	j.Steps = append(NewDockerCredentialActions(descriptor.DockerCredentials), j.Steps...)
	j.Steps = append(NewHttpCredentialActions(descriptor.HttpCredentials), j.Steps...)
	w.Jobs["republish-image"] = j

	return NewActionContribution(w)
}
