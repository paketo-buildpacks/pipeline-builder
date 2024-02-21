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
	"regexp"
	"strings"

	"github.com/paketo-buildpacks/pipeline-builder/octo/actions"
	"github.com/paketo-buildpacks/pipeline-builder/octo/actions/event"
)

const (
	FormatFile  = "file"
	FormatImage = "image"
)

func ContributeTest(descriptor Descriptor) (*Contribution, error) {
	if descriptor.OfflinePackages != nil {
		return nil, nil
	}

	if descriptor.Package.Repository == "" && len(descriptor.Package.Repositories) > 0 {
		descriptor.Package.Repository = descriptor.Package.Repositories[0]
	}

	if len(descriptor.Package.Repositories) == 0 && descriptor.Package.Repository != "" {
		descriptor.Package.Repositories = append(descriptor.Package.Repositories, descriptor.Package.Repository)
	}

	w := actions.Workflow{
		Name: "Tests",
		On: map[event.Type]event.Event{
			event.PullRequestType: event.PullRequest{},
			event.PushType: event.Push{
				Branches: []string{"main"},
			},
			event.MergeGroupType: event.MergeGroup{
				Types:    []event.MergeGroupActivityType{event.MergeGroupChecksRequested},
				Branches: []string{"main"},
			},
		},
		Jobs: map[string]actions.Job{},
	}

	goFiles, err := Find(descriptor.Path, regexp.MustCompile(`.+\.go`).MatchString)
	if err != nil {
		return nil, fmt.Errorf("unable to Find .go files in %s\n%w", descriptor.Path, err)
	}

	integrationTestFiles, err := Find(descriptor.Path, regexp.MustCompile(`integration/.+\.go`).MatchString)
	if err != nil {
		return nil, fmt.Errorf("unable to Find .go files in %s\n%w", filepath.Join(descriptor.Path, "integration"), err)
	}

	if len(goFiles) > 0 {
		j := actions.Job{
			Name:   "Unit Test",
			RunsOn: []actions.VirtualEnvironment{actions.UbuntuLatest},
			Steps: []actions.Step{
				{
					Uses: "actions/checkout@v4",
				},
				{
					Uses: "actions/cache@v4",
					With: map[string]interface{}{
						"path":         "${{ env.HOME }}/go/pkg/mod",
						"key":          "${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}",
						"restore-keys": "${{ runner.os }}-go-",
					},
				},
				{
					Uses: "actions/setup-go@v5",
					With: map[string]interface{}{"go-version": GoVersion},
				},
			},
		}

		if len(integrationTestFiles) == 0 {
			j.Steps = append(j.Steps, descriptor.Test.Steps...)
		} else {
			j.Steps = append(j.Steps, actions.Step{
				Name: "Install richgo",
				Run:  StatikString("/install-richgo.sh"),
				Env:  map[string]string{"RICHGO_VERSION": RichGoVersion},
			},
				actions.Step{
					Name: "Run Tests",
					Run:  StatikString("/run-unit-tests.sh"),
					Env:  map[string]string{"RICHGO_FORCE_COLOR": "1"},
				})
		}

		w.Jobs["unit"] = j
	}

	if len(integrationTestFiles) > 0 {
		j := actions.Job{
			Name:   "Integration Test",
			RunsOn: []actions.VirtualEnvironment{actions.UbuntuLatest},
			Steps: []actions.Step{
				{
					Uses: "actions/checkout@v4",
				},
				{
					Uses: "actions/cache@v4",
					With: map[string]interface{}{
						"path":         "${{ env.HOME }}/go/pkg/mod",
						"key":          "${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}",
						"restore-keys": "${{ runner.os }}-go-",
					},
				},
				{
					Uses: "actions/setup-go@v5",
					With: map[string]interface{}{"go-version": GoVersion},
				},
				{
					Name: "Install create-package",
					Run:  StatikString("/install-create-package.sh"),
				},
				{
					Uses: fmt.Sprintf("buildpacks/github-actions/setup-pack@v%s", BuildpackActionsVersion),
					With: map[string]interface{}{
						"pack-version": PackVersion,
					},
				},
				{
					Name: "Enable pack Experimental",
					If:   fmt.Sprintf("${{ %t }}", descriptor.Package.Platform.OS == PlatformWindows),
					Run:  StatikString("/enable-pack-experimental.sh"),
				},
				{
					Name: "Install richgo",
					Run:  StatikString("/install-richgo.sh"),
					Env:  map[string]string{"RICHGO_VERSION": RichGoVersion},
				},
				{
					Name: "Run Tests",
					Run:  StatikString("/run-integration-tests.sh"),
					Env:  map[string]string{"RICHGO_FORCE_COLOR": "1"},
				},
			},
		}

		w.Jobs["integration"] = j
	}

	if descriptor.Builder != nil {
		j := actions.Job{
			Name:   "Create Builder Test",
			RunsOn: []actions.VirtualEnvironment{actions.UbuntuLatest},
			Steps: []actions.Step{
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
					Name: "Create Builder",
					Run:  StatikString("/create-builder.sh"),
					Env: map[string]string{
						"BUILDER": "test",
						"VERSION": "${{ steps.version.outputs.version }}",
					},
				},
			},
		}

		j.Steps = append(NewDockerCredentialActions(descriptor.DockerCredentials), j.Steps...)

		w.Jobs["create-builder"] = j
	}

	if descriptor.Package.Enabled {
		format := FormatImage
		if descriptor.Package.Platform.OS == PlatformWindows {
			format = FormatFile
		}

		j := actions.Job{
			Name:   "Create Package Test",
			RunsOn: []actions.VirtualEnvironment{actions.UbuntuLatest},
			Steps: []actions.Step{
				{
					Uses: "actions/setup-go@v5",
					With: map[string]interface{}{"go-version": GoVersion},
				},
				{
					Name: "Install create-package",
					Run:  StatikString("/install-create-package.sh"),
				},
				{
					Uses: fmt.Sprintf("buildpacks/github-actions/setup-pack@v%s", BuildpackActionsVersion),
					With: map[string]interface{}{
						"pack-version": PackVersion,
					},
				},
				{
					Name: "Enable pack Experimental",
					If:   fmt.Sprintf("${{ %t }}", descriptor.Package.Platform.OS == PlatformWindows),
					Run:  StatikString("/enable-pack-experimental.sh"),
				},
				{
					Uses: "actions/checkout@v4",
				},
				{
					Uses: "actions/cache@v4",
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
						"INCLUDE_DEPENDENCIES": "true",
						"OS":                   descriptor.Package.Platform.OS,
						"VERSION":              "${{ steps.version.outputs.version }}",
					},
				},
				{
					Name: "Package Buildpack",
					Run:  StatikString("/package-buildpack.sh"),
					Env: map[string]string{
						"FORMAT":   format,
						"PACKAGES": "test",
						"VERSION":  "${{ steps.version.outputs.version }}",
					},
				},
			},
		}

		skipPrefixes := []string{
			"paketo-buildpacks",
			"paketobuildpacks",
			"paketo-community",
			"paketocommunity",
		}

		for _, repo := range descriptor.Package.Repositories {
			skipMatch := false
			for _, skipPrefix := range skipPrefixes {
				if strings.Contains(repo, skipPrefix) {
					skipMatch = true
					break
				}
			}
			if !skipMatch {
				j.Steps = append(NewDockerCredentialActions(descriptor.DockerCredentials), j.Steps...)
			}
		}
		j.Steps = append(NewHttpCredentialActions(descriptor.HttpCredentials), j.Steps...)

		w.Jobs["create-package"] = j
	}

	c, err := NewActionContributionWithNamespace(Namespace, w)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
