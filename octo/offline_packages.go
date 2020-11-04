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
	"strings"

	"github.com/paketo-buildpacks/pipeline-builder/octo/actions"
	"github.com/paketo-buildpacks/pipeline-builder/octo/actions/event"
)

func ContributeOfflinePackages(descriptor Descriptor) ([]Contribution, error) {
	var contributions []Contribution

	for _, o := range descriptor.OfflinePackages {
		if c, err := contributeOfflinePackage(descriptor, o); err != nil {
			return nil, err
		} else {
			contributions = append(contributions, c)
		}
	}

	return contributions, nil
}

func contributeOfflinePackage(descriptor Descriptor, offlinePackage OfflinePackage) (Contribution, error) {
	w := actions.Workflow{
		Name: fmt.Sprintf("Create Package %s", filepath.Base(offlinePackage.Target)),
		On: map[event.Type]event.Event{
			event.ScheduleType:         event.Schedule{{Minute: "0", Hour: "12-23", DayOfWeek: "1-5"}},
			event.WorkflowDispatchType: event.WorkflowDispatch{},
		},
		Jobs: map[string]actions.Job{
			"offline-package": {
				Name:   "Create Offline Package",
				RunsOn: []actions.VirtualEnvironment{actions.UbuntuLatest},
				Steps: []actions.Step{
					{
						Name: "Install crane",
						Run:  statikString("/install-crane.sh"),
						Env:  map[string]string{"CRANE_VERSION": CraneVersion},
					},
					{
						Uses: "actions/checkout@v2",
						With: map[string]interface{}{
							"repository":  offlinePackage.Source,
							"fetch-depth": 0,
						},
					},
					{
						Id:   "version",
						Name: "Checkout next version",
						Run:  statikString("/checkout-next-version.sh"),
						Env: map[string]string{
							"SOURCE": offlinePackage.Source,
							"TARGET": offlinePackage.Target,
						},
					},
					{
						Uses: "actions/setup-go@v2",
						If:   "${{ ! steps.version.outputs.skip }}",
						With: map[string]interface{}{"go-version": GoVersion},
					},
					{
						Uses: "actions/cache@v2",
						If:   "${{ ! steps.version.outputs.skip }}",
						With: map[string]interface{}{
							"path": strings.Join([]string{
								"${{ env.HOME }}/carton-cache",
							}, "\n"),
							"key":          "${{ runner.os }}-go-${{ hashFiles('**/buildpack.toml') }}",
							"restore-keys": "${{ runner.os }}-go-",
						},
					},
					{
						Name: "Install create-package",
						If:   "${{ ! steps.version.outputs.skip }}",
						Run:  statikString("/install-create-package.sh"),
					},
					{
						Name: "Install pack",
						If:   "${{ ! steps.version.outputs.skip }}",
						Run:  statikString("/install-pack.sh"),
						Env:  map[string]string{"PACK_VERSION": PackVersion},
					},
					{
						Name: "Create Package",
						If:   "${{ ! steps.version.outputs.skip }}",
						Run:  statikString("/create-package.sh"),
						Env: map[string]string{
							"INCLUDE_DEPENDENCIES": "true",
							"VERSION":              "${{ steps.version.outputs.version }}",
						},
					},
					{
						Name: "Package Buildpack",
						If:   "${{ ! steps.version.outputs.skip }}",
						Run:  statikString("/package-buildpack.sh"),
						Env: map[string]string{
							"PACKAGE": offlinePackage.Target,
							"PUBLISH": "true",
							"VERSION": "${{ steps.version.outputs.version }}",
						},
					},
				},
			},
		},
	}

	j := w.Jobs["offline-package"]
	j.Steps = append(NewDockerCredentialActions(descriptor.DockerCredentials), j.Steps...)
	j.Steps = append(NewHttpCredentialActions(descriptor.HttpCredentials), j.Steps...)
	w.Jobs["offline-package"] = j

	return NewActionContribution(w)
}
