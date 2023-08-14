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

	"github.com/paketo-buildpacks/pipeline-builder/v2/octo/actions"
	"github.com/paketo-buildpacks/pipeline-builder/v2/octo/actions/event"
	"github.com/paketo-buildpacks/pipeline-builder/v2/octo/release"
)

func ContributeDraftRelease(descriptor Descriptor) ([]Contribution, error) {
	if descriptor.OfflinePackages != nil {
		return nil, nil
	}

	var contributions []Contribution

	d := release.Drafter{
		TagTemplate:  "v$RESOLVED_VERSION",
		NameTemplate: "$RESOLVED_VERSION",
		Template:     "$CHANGES",
		Categories: []release.Category{
			{
				Title:  "⭐️ Enhancements",
				Labels: []string{"type:enhancement"},
			},
			{
				Title:  "🐞 Bug Fixes",
				Labels: []string{"type:bug"},
			},
			{
				Title:  "📔 Documentation",
				Labels: []string{"type:documentation"},
			},
			{
				Title:  "⛏ Dependency Upgrades",
				Labels: []string{"type:dependency-upgrade"},
			},
			{
				Title:  "🚧 Tasks",
				Labels: []string{"type:task"},
			},
		},
		ExcludeLabels: []string{"type:question"},
		VersionResolver: release.VersionResolver{
			Major:   release.Version{Labels: []string{"semver:major"}},
			Minor:   release.Version{Labels: []string{"semver:minor"}},
			Patch:   release.Version{Labels: []string{"semver:patch"}},
			Default: release.PatchVersionComponent,
		},
	}

	if c, err := NewDrafterContribution(d); err != nil {
		return nil, err
	} else {
		contributions = append(contributions, c)
	}

	w := actions.Workflow{
		Name: "Update Draft Release",
		On: map[event.Type]event.Event{
			event.PushType: event.Push{
				Branches: []string{"main"},
			},
		},
		Jobs: map[string]actions.Job{
			"update": {
				Name:   "Update Draft Release",
				RunsOn: []actions.VirtualEnvironment{actions.UbuntuLatest},
				Steps: []actions.Step{
					{
						Id:   "release-drafter",
						Uses: "release-drafter/release-drafter@v5",
						Env:  map[string]string{"GITHUB_TOKEN": descriptor.GitHub.Token},
					},
				},
			},
		},
	}

	file := filepath.Join(descriptor.Path, "builder.toml")
	builderExists, err := Exists(file)
	if err != nil {
		return nil, fmt.Errorf("unable to determine if %s Exists\n%w", file, err)
	}

	file = filepath.Join(descriptor.Path, "buildpack.toml")
	buildpackExists, err := Exists(file)
	if err != nil {
		return nil, fmt.Errorf("unable to determine if %s Exists\n%w", file, err)
	}

	file = filepath.Join(descriptor.Path, "package.toml")
	packageExists, err := Exists(file)
	if err != nil {
		return nil, fmt.Errorf("unable to determine if %s Exists\n%w", file, err)
	}

	if builderExists || buildpackExists {
		j := w.Jobs["update"]

		if builderExists || packageExists {
			j.Steps = append(j.Steps, NewDockerCredentialActions(descriptor.DockerCredentials)...)
		}

		draftReleaseContext := map[string]interface{}{
			"github_token":     descriptor.GitHub.Token,
			"release_id":       "${{ steps.release-drafter.outputs.id }}",
			"release_tag_name": "${{ steps.release-drafter.outputs.tag_name }}",
			"release_name":     "${{ steps.release-drafter.outputs.name }}",
			"release_body":     "${{ steps.release-drafter.outputs.body }}",
		}

		for i, mapper := range descriptor.GitHub.Mappers {
			draftReleaseContext[fmt.Sprintf("mapper_%d", i+1)] = mapper
		}

		draftReleaseEnv := map[string]string{}
		for key, val := range descriptor.GitHub.BuildpackTOMLPaths {
			draftReleaseEnv[fmt.Sprintf("BP_TOML_PATH_%s", toEnvVar(key))] = val
		}

		j.Steps = append(j.Steps,
			actions.Step{
				Uses: "actions/checkout@v3",
			},
			actions.Step{
				Name: "Update draft release with buildpack information",
				Uses: "docker://ghcr.io/paketo-buildpacks/actions/draft-release:main",
				With: draftReleaseContext,
				Env:  draftReleaseEnv,
			},
		)
		w.Jobs["update"] = j
	}

	if c, err := NewActionContributionWithNamespace(Namespace, w); err != nil {
		return nil, err
	} else {
		contributions = append(contributions, c)
	}

	return contributions, nil
}

func toEnvVar(key string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(key, "-", "_"), "/", "_"), " ", "_"))
}
