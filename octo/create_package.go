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
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/buildpacks/libcnb/v2"

	"github.com/paketo-buildpacks/pipeline-builder/octo/actions"
	"github.com/paketo-buildpacks/pipeline-builder/octo/actions/event"
)

func ContributeCreatePackage(descriptor Descriptor) (*Contribution, error) {
	if !descriptor.Package.Enabled {
		return nil, nil
	}

	if descriptor.Package.Repository == "" && len(descriptor.Package.Repositories) > 0 {
		descriptor.Package.Repository = descriptor.Package.Repositories[0]
	}

	if len(descriptor.Package.Repositories) == 0 && descriptor.Package.Repository != "" {
		descriptor.Package.Repositories = append(descriptor.Package.Repositories, descriptor.Package.Repository)
	}

	// Is this a buildpack or an extension?
	bpfile := filepath.Join(descriptor.Path, "buildpack.toml")
	extnfile := filepath.Join(descriptor.Path, "extension.toml")
	id := ""
	extension := false
	if _, err := os.Stat(bpfile); err == nil {
		s, err := os.ReadFile(bpfile)
		if err != nil {
			return nil, fmt.Errorf("unable to read buildpack.toml %s\n%w", bpfile, err)
		}
		var b libcnb.Buildpack
		if err := toml.Unmarshal(s, &b); err != nil {
			return nil, fmt.Errorf("unable to decode %s\n%w", bpfile, err)
		}
		id = b.Info.ID
	} else if _, err := os.Stat(extnfile); err == nil {
		s, err := os.ReadFile(extnfile)
		if err != nil {
			return nil, fmt.Errorf("unable to read extension.toml %s\n%w", extnfile, err)
		}
		var e libcnb.Extension
		if err := toml.Unmarshal(s, &e); err != nil {
			return nil, fmt.Errorf("unable to decode %s\n%w", extnfile, err)
		}
		id = e.Info.ID
		extension = true
	} else {
		return nil, fmt.Errorf("unable to read buildpack/extension.toml at %s\n", descriptor.Path)
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
						Uses: "actions/setup-go@v4",
						With: map[string]interface{}{"go-version": GoVersion},
					},
					{
						Name: "Install create-package",
						Run:  StatikString("/install-create-package.sh"),
						Env:  map[string]string{"PAKETO_LIBPAK_COMMIT": "${{ vars.PAKETO_LIBPAK_COMMIT }}"},
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
						Name: "Enable pack Experimental",
						If:   fmt.Sprintf("${{ %t }}", descriptor.Package.Platform.OS == PlatformWindows),
						Run:  StatikString("/enable-pack-experimental.sh"),
					},
					{
						Uses: "actions/checkout@v3",
					},
					{
						Uses: "actions/cache@v3",
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
							"EXTENSION":            strconv.FormatBool(extension),
							"INCLUDE_DEPENDENCIES": strconv.FormatBool(descriptor.Package.IncludeDependencies),
							"OS":                   descriptor.Package.Platform.OS,
							"VERSION":              "${{ steps.version.outputs.version }}",
							"SOURCE_PATH":          descriptor.Package.SourcePath,
						},
					},
					{
						Id:   "package",
						Name: "Package Buildpack/Extension",
						Run:  StatikString("/package-buildpack.sh"),
						Env: map[string]string{
							"EXTENSION":     strconv.FormatBool(extension),
							"PACKAGES":      strings.Join(descriptor.Package.Repositories, " "),
							"PUBLISH":       "true",
							"VERSION":       "${{ steps.version.outputs.version }}",
							"VERSION_MAJOR": "${{ steps.version.outputs.version-major }}",
							"VERSION_MINOR": "${{ steps.version.outputs.version-minor }}",
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
					{
						Uses: "docker://ghcr.io/buildpacks/actions/registry/request-add-entry:4.0.1",
						If:   fmt.Sprintf("${{ %t && %t }}", descriptor.Package.Register, !extension), //ignore registration for extensions
						With: map[string]interface{}{
							"token":   descriptor.Package.RegistryToken,
							"id":      id,
							"version": "${{ steps.version.outputs.version }}",
							"address": fmt.Sprintf("%s@${{ steps.package.outputs.digest }}", descriptor.Package.Repository),
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

	c, err := NewActionContributionWithNamespace(Namespace, w)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
