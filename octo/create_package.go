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
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/buildpacks/libcnb"

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
						Uses: "actions/setup-go@v2",
						With: map[string]interface{}{"go-version": GoVersion},
					},
					{
						Name: "Install create-package",
						Run:  StatikString("/install-create-package.sh"),
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
						Id:   "version",
						Name: "Compute Version",
						Run:  StatikString("/compute-version.sh"),
					},
					{
						Name: "Create Package",
						Run:  StatikString("/create-package.sh"),
						Env: map[string]string{
							"INCLUDE_DEPENDENCIES": strconv.FormatBool(descriptor.Package.IncludeDependencies),
							"VERSION":              "${{ steps.version.outputs.version }}",
						},
					},
					{
						Id:   "package",
						Name: "Package Buildpack",
						Run:  StatikString("/package-buildpack.sh"),
						Env: map[string]string{
							"PACKAGE": descriptor.Package.Repository,
							"PUBLISH": "true",
							"VERSION": "${{ steps.version.outputs.version }}",
						},
					},
					{
						Name: "Update release with digest",
						Run:  StatikString("/update-release-digest.sh"),
						Env: map[string]string{
							"DIGEST":       "${{ steps.package.outputs.digest }}",
							"GITHUB_TOKEN": descriptor.GitHub.Token,
						},
					},
				},
			},
		},
	}

	j := w.Jobs["create-package"]
	j.Steps = append(NewDockerCredentialActions(descriptor.DockerCredentials), j.Steps...)
	j.Steps = append(NewHttpCredentialActions(descriptor.HttpCredentials), j.Steps...)

	if descriptor.Package.Register {
		file := filepath.Join(descriptor.Path, "buildpack.toml")
		s, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("unable to read %s\n%w", file, err)
		}

		var b libcnb.Buildpack
		if err := toml.Unmarshal(s, &b); err != nil {
			return nil, fmt.Errorf("unable to decode %s\n%w", file, err)
		}

		j.Steps = append(j.Steps, actions.Step{
			Uses: "docker://ghcr.io/buildpacks/actions/registry:main",
			With: map[string]interface{}{
				"token":   descriptor.Package.RegistryToken,
				"id":      b.Info.ID,
				"version": "${{ steps.version.outputs.version }}",
				"address": fmt.Sprintf("%s@${{ steps.package.outputs.digest }}", descriptor.Package.Repository),
			},
		})
	}

	w.Jobs["create-package"] = j

	c, err := NewActionContribution(w)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
