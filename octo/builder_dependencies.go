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

	"github.com/BurntSushi/toml"

	"github.com/paketo-buildpacks/pipeline-builder/octo/actions"
	"github.com/paketo-buildpacks/pipeline-builder/octo/actions/event"
)

func ContributeBuilderDependencies(descriptor Descriptor) ([]Contribution, error) {
	file := filepath.Join(descriptor.Path, "builder.toml")
	if e, err := exists(file); err != nil {
		return nil, fmt.Errorf("unable to determine if %s exists\n%w", file, err)
	} else if !e {
		return nil, nil
	}

	var b RawBuilder
	if _, err := toml.DecodeFile(file, &b); err != nil {
		return nil, fmt.Errorf("unable to decode %s\n%w", file, err)
	}

	var contributions []Contribution

	re := regexp.MustCompile(`(.+):[^:]+`)
	for _, b := range b.Buildpacks {
		if g := re.FindStringSubmatch(b.Image); g != nil {
			if c, err := contributePackageDependency(descriptor, g[1]); err != nil {
				return nil, err
			} else {
				contributions = append(contributions, c)
			}
		}
	}

	re = regexp.MustCompile(`(.+):[\d.]+-(.+)`)
	if g := re.FindStringSubmatch(b.Stack.BuildImage); g != nil {
		if c, err := contributeBuildImage(descriptor, g[1], g[2]); err != nil {
			return nil, err
		} else {
			contributions = append(contributions, c)
		}
	}

	if c, err := contributeLifecycle(); err != nil {
		return nil, err
	} else {
		contributions = append(contributions, c)
	}

	return contributions, nil
}

func contributeBuildImage(descriptor Descriptor, image string, classifier string) (Contribution, error) {
	w := actions.Workflow{
		Name: "Update Build Image",
		On: map[event.Type]event.Event{
			event.ScheduleType:         event.Schedule{{Minute: "15"}},
			event.WorkflowDispatchType: event.WorkflowDispatch{},
		},
		Jobs: map[string]actions.Job{
			"update": {
				Name:   "Update Build Image",
				RunsOn: []actions.VirtualEnvironment{actions.UbuntuLatest},
				Steps: []actions.Step{
					{
						Uses: "actions/checkout@v2",
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
						Name: "Install update-build-image-dependency",
						Run:  statikString("/install-update-build-image-dependency.sh"),
					},
					{
						Name: "Install yj",
						Run:  statikString("/install-yj.sh"),
						Env:  map[string]string{"YJ_VERSION": YJVersion},
					},
					{
						Id:   "build-image",
						Name: "Update Build Image Dependency",
						Run:  statikString("/update-build-image-dependency.sh"),
						Env: map[string]string{
							"IMAGE":      image,
							"CLASSIFIER": classifier,
						},
					},
					{
						Uses: "peter-evans/create-pull-request@v3",
						With: map[string]interface{}{
							"token": "${{ secrets.GITHUB_TOKEN }}",
							"commit-message": fmt.Sprintf(`Bump %[1]s from ${{ steps.build-image.outputs.old-version }} to ${{ steps.build-image.outputs.new-version }}

Bumps %[1]s from ${{ steps.build-image.outputs.old-version }} to ${{ steps.build-image.outputs.new-version }}.`, image),
							"signoff":       true,
							"branch":        fmt.Sprintf("update/build-image/%s", filepath.Base(image)),
							"delete-branch": true,
							"title":         fmt.Sprintf("Bump %s from ${{ steps.build-image.outputs.old-version }} to ${{ steps.build-image.outputs.new-version }}", image),
							"body":          fmt.Sprintf("Bumps [`%[1]s`](https://%[1]s) from [`${{ steps.build-image.outputs.old-version }}`](https://%[1]s:${{ steps.build-image.outputs.old-version }}) to [`${{ steps.build-image.outputs.new-version }}`](https://%[1]s:${{ steps.build-image.outputs.new-version }}).", image),
							"labels":        "semver:minor, type:dependency-upgrade",
						},
					},
				},
			},
		},
	}

	j := w.Jobs["update"]
	j.Steps = append(NewDockerCredentialActions(descriptor.DockerCredentials), j.Steps...)
	w.Jobs["update"] = j

	return NewActionContribution(w)
}

func contributeLifecycle() (Contribution, error) {
	w := actions.Workflow{
		Name: "Update Lifecycle",
		On: map[event.Type]event.Event{
			event.ScheduleType:         event.Schedule{{Minute: "15"}},
			event.WorkflowDispatchType: event.WorkflowDispatch{},
		},
		Jobs: map[string]actions.Job{
			"update": {
				Name:   "Update Lifecycle",
				RunsOn: []actions.VirtualEnvironment{actions.UbuntuLatest},
				Steps: []actions.Step{
					{
						Uses: "actions/checkout@v2",
					},
					{
						Uses: "actions/setup-go@v2",
						With: map[string]interface{}{"go-version": GoVersion},
					},
					{
						Name: "Install update-lifecycle-dependency",
						Run:  statikString("/install-update-lifecycle-dependency.sh"),
					},
					{
						Name: "Install yj",
						Run:  statikString("/install-yj.sh"),
						Env:  map[string]string{"YJ_VERSION": YJVersion},
					},
					{
						Id:   "dependency",
						Uses: "docker://ghcr.io/paketo-buildpacks/actions/github-release-dependency:main",
						With: map[string]interface{}{
							"glob":       `lifecycle-v[^+]+\+linux\.x86-64\.tgz`,
							"owner":      "buildpacks",
							"repository": "lifecycle",
							"token":      "${{ secrets.GITHUB_TOKEN }}",
						},
					},
					{
						Id:   "lifecycle",
						Name: "Update Lifecycle Dependency",
						Run:  statikString("/update-lifecycle-dependency.sh"),
						Env: map[string]string{
							"VERSION": "${{ steps.dependency.outputs.version }}",
						},
					},
					{
						Uses: "peter-evans/create-pull-request@v3",
						With: map[string]interface{}{
							"token": "${{ secrets.GITHUB_TOKEN }}",
							"commit-message": `Bump lifecycle from ${{ steps.lifecycle.outputs.old-version }} to ${{ steps.lifecycle.outputs.new-version }}

Bumps lifecycle from ${{ steps.lifecycle.outputs.old-version }} to ${{ steps.lifecycle.outputs.new-version }}.`,
							"signoff":       true,
							"branch":        "update/package/lifecycle",
							"delete-branch": true,
							"title":         "Bump lifecycle from ${{ steps.lifecycle.outputs.old-version }} to ${{ steps.lifecycle.outputs.new-version }}",
							"body":          "Bumps `lifecycle` from `${{ steps.lifecycle.outputs.old-version }}` to `${{ steps.lifecycle.outputs.new-version }}`.",
							"labels":        "semver:minor, type:dependency-upgrade",
						},
					},
				},
			},
		},
	}

	return NewActionContribution(w)
}

type RawBuilder struct {
	Buildpacks []RawBuildpack
	Stack      RawStack
}

type RawBuildpack struct {
	Image string
}

type RawStack struct {
	BuildImage string `toml:"build-image"`
}
