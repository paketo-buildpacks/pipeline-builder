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
	"github.com/paketo-buildpacks/pipeline-builder/octo/internal"
	"github.com/paketo-buildpacks/pipeline-builder/octo/release"
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
						Env:  map[string]string{"GITHUB_TOKEN": "${{ secrets.GITHUB_TOKEN }}"},
					},
				},
			},
		},
	}

	file := filepath.Join(descriptor.Path, "buildpack.toml")
	if e, err := internal.Exists(file); err != nil {
		return nil, fmt.Errorf("unable to determine if %s exists\n%w", file, err)
	} else if e {
		j := w.Jobs["update"]

		j.Steps = append(j.Steps, NewDockerLoginActions(descriptor.Credentials)...)

		file := filepath.Join(descriptor.Path, "package.toml")
		if e, err := internal.Exists(file); err != nil {
			return nil, fmt.Errorf("unable to determine if %s exists\n%w", file, err)
		} else if e {
			j.Steps = append(j.Steps,
				actions.Step{
					Uses: "actions/setup-go@v2",
					With: map[string]interface{}{"go-version": GoVersion},
				},
				actions.Step{
					Name: "Install crane",
					Run:  internal.StatikString("/install-crane.sh"),
				},
			)
		}

		j.Steps = append(j.Steps,
			actions.Step{
				Uses: "actions/checkout@v2",
			},
			actions.Step{
				Name: "Install yj",
				Run:  internal.StatikString("/install-yj.sh"),
				Env:  map[string]string{"YJ_VERSION": YJVersion},
			},
			actions.Step{
				Name: "Update draft release with buildpack information",
				Run:  internal.StatikString("/update-draft-release-buildpack.sh"),
				Env: map[string]string{
					"GITHUB_TOKEN":     "${{ secrets.GITHUB_TOKEN }}",
					"RELEASE_ID":       "${{ steps.release-drafter.outputs.id }}",
					"RELEASE_TAG_NAME": "${{ steps.release-drafter.outputs.tag_name }}",
					"RELEASE_NAME":     "${{ steps.release-drafter.outputs.name }}",
					"RELEASE_BODY":     "${{ steps.release-drafter.outputs.body }}",
				},
			},
		)
		w.Jobs["update"] = j
	}

	if c, err := NewActionContribution(w); err != nil {
		return nil, err
	} else {
		contributions = append(contributions, c)
	}

	return contributions, nil
}
