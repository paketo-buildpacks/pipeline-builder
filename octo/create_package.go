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
	"github.com/paketo-buildpacks/pipeline-builder/octo/actions"
	"github.com/paketo-buildpacks/pipeline-builder/octo/actions/event"
	"github.com/paketo-buildpacks/pipeline-builder/octo/internal"
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
						Id:   "version",
						Name: "Compute Version",
						Run:  internal.StatikString("/compute-version.sh"),
					},
					{
						Uses: "actions/setup-go@v2",
						With: map[string]interface{}{"go-version": GoVersion},
					},
					{
						Name: "Install Create Package",
						Run:  internal.StatikString("/install-create-package.sh"),
					},
					{
						Name: "Create Package",
						Run:  internal.StatikString("/create-package.sh"),
						Env:  map[string]string{"VERSION": "${{ steps.version.outputs.version }}"},
					},
					{
						Name: "Install Crane",
						Run:  internal.StatikString("/install-crane.sh"),
					},
					{
						Name: "Install pack",
						Run:  internal.StatikString("/install-pack.sh"),
						Env:  map[string]string{"PACK_VERSION": PackVersion},
					},
					{
						Uses: "GoogleCloudPlatform/github-actions/setup-gcloud@master",
						With: map[string]interface{}{
							"service_account_key": "${{ secrets.JAVA_GCLOUD_SERVICE_ACCOUNT_KEY }}",
						},
					},
					{
						Name: "Configure gcloud docker credentials",
						Run: "gcloud auth configure-docker",
					},
					{
						Id:   "package",
						Name: "Package Buildpack",
						Run:  internal.StatikString("/package-buildpack.sh"),
						Env: map[string]string{
							"PACKAGE": descriptor.Package.Repository,
							"PUBLISH": "true",
							"VERSION": "${{ steps.version.outputs.version }}",
						},
					},
					{
						Name: "Update release with digest",
						Run:  internal.StatikString("/update-release-digest.sh"),
						Env: map[string]string{
							"DIGEST":       "${{ steps.package.outputs.digest }}",
							"GITHUB_TOKEN": "${{ secrets.GITHUB_TOKEN }}",
						},
					},
				},
			},
		},
	}

	c, err := NewActionContribution(w)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
