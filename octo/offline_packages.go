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
						Uses: fmt.Sprintf("buildpacks/github-actions/setup-tools@v%s", BuildpackActionsVersion),
						With: map[string]interface{}{
							"crane-version": CraneVersion,
							"yj-version":    YJVersion,
						},
					},
					{
						Uses: "actions/checkout@v4",
						With: map[string]interface{}{
							"repository":  offlinePackage.Source,
							"fetch-depth": 0,
						},
					},
					{
						Id:   "version",
						Name: "Checkout next version",
						Run:  StatikString("/checkout-next-version.sh"),
						Env: map[string]string{
							"SOURCE":     offlinePackage.Source,
							"TARGET":     offlinePackage.Target,
							"TAG_PREFIX": offlinePackage.TagPrefix,
						},
					},
					{
						Uses: "actions/setup-go@v5",
						If:   "${{ ! steps.version.outputs.skip }}",
						With: map[string]interface{}{"go-version": GoVersion},
					},
					{
						Uses: "actions/cache@v4",
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
						Run:  StatikString("/install-create-package.sh"),
					},
					{
						If:   "${{ ! steps.version.outputs.skip }}",
						Uses: fmt.Sprintf("buildpacks/github-actions/setup-pack@v%s", BuildpackActionsVersion),
						With: map[string]interface{}{
							"pack-version": PackVersion,
						},
					},
					{
						Name: "Enable pack Experimental",
						If:   fmt.Sprintf("${{ %t }}", offlinePackage.Platform.OS == PlatformWindows),
						Run:  StatikString("/enable-pack-experimental.sh"),
					},
					{
						Name: "Create Package",
						If:   "${{ ! steps.version.outputs.skip }}",
						Run:  StatikString("/create-package.sh"),
						Env: map[string]string{
							"INCLUDE_DEPENDENCIES": "true",
							"OS":                   offlinePackage.Platform.OS,
							"VERSION":              "${{ steps.version.outputs.version }}",
							"SOURCE_PATH":          offlinePackage.SourcePath,
						},
					},
					{
						Name: "Package Buildpack",
						If:   "${{ ! steps.version.outputs.skip }}",
						Run:  StatikString("/package-buildpack.sh"),
						Env: map[string]string{
							"PACKAGES": offlinePackage.Target,
							"PUBLISH":  "true",
							"VERSION":  "${{ steps.version.outputs.version }}",
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

	return NewActionContributionWithNamespace(Namespace, w)
}
