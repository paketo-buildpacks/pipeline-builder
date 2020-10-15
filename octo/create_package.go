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
	"strconv"
	"strings"

	"github.com/paketo-buildpacks/pipeline-builder/octo/actions"
	"github.com/paketo-buildpacks/pipeline-builder/octo/actions/event"
)

func ContributeCreatePackage(descriptor Descriptor) (*Contribution, error) {
	if descriptor.Package == nil {
		return nil, nil
	}

	w := actions.Workflow{
		Name: "Create Package",
		On: map[event.Type]event.Event{
			event.ReleaseType: event.Release{
				Types: []event.ReleaseActivityType{
					event.ReleasePublished,
				},
			},
		},
		Jobs: map[string]actions.Job{
			"create-package": {
				Name:   "Create Package",
				RunsOn: []actions.VirtualEnvironment{actions.UbuntuLatest},
				Steps: []actions.Step{
					{
						Uses: "actions/checkout@v2",
					},
					{
						Uses: "actions/cache@v2",
						If:   fmt.Sprintf("${{ %t }}", descriptor.Package.IncludeDependencies),
						With: map[string]interface{}{
							"path": strings.Join([]string{
								"${{ env.HOME }}/.pack",
								"${{ env.HOME }}/carton-cache",
							}, "\n"),
							"key":          "${{ runner.os }}-go-${{ hashFiles('**/buildpack.toml', '**/package.toml') }}",
							"restore-keys": "${{ runner.os }}-go-",
						},
					},
					{
						Uses: "actions/setup-go@v2",
						With: map[string]interface{}{"go-version": GoVersion},
					},
					{
						Name: "Install crane",
						Run:  statikString("/install-crane.sh"),
					},
					{
						Name: "Install create-package",
						Run:  statikString("/install-create-package.sh"),
					},
					{
						Name: "Install pack",
						Run:  statikString("/install-pack.sh"),
						Env:  map[string]string{"PACK_VERSION": PackVersion},
					},
					{
						Id:   "version",
						Name: "Compute Version",
						Run:  statikString("/compute-version.sh"),
					},
					{
						Name: "Create Package",
						Run:  statikString("/create-package.sh"),
						Env: map[string]string{
							"INCLUDE_DEPENDENCIES": strconv.FormatBool(descriptor.Package.IncludeDependencies),
							"VERSION":              "${{ steps.version.outputs.version }}",
						},
					},
					{
						Id:   "package",
						Name: "Package Buildpack",
						Run:  statikString("/package-buildpack.sh"),
						Env: map[string]string{
							"PACKAGE": descriptor.Package.Repository,
							"PUBLISH": "true",
							"VERSION": "${{ steps.version.outputs.version }}",
						},
					},
					{
						Name: "Update release with digest",
						Run:  statikString("/update-release-digest.sh"),
						Env: map[string]string{
							"DIGEST":       "${{ steps.package.outputs.digest }}",
							"GITHUB_TOKEN": "${{ secrets.GITHUB_TOKEN }}",
						},
					},
				},
			},
		},
	}

	j := w.Jobs["create-package"]
	j.Steps = append(NewDockerCredentialActions(descriptor.DockerCredentials), j.Steps...)
	j.Steps = append(NewHttpCredentialActions(descriptor.HttpCredentials), j.Steps...)
	w.Jobs["create-package"] = j

	c, err := NewActionContribution(w)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
