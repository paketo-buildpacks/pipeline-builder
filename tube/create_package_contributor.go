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

package tube

type CreatePackageContributor struct {
	Descriptor Descriptor
	Salt       string
}

func (c CreatePackageContributor) Group() string {
	return "build"
}

func (c CreatePackageContributor) Job() Job {
	b := NewBuildCommonResource()
	k := NewPackResource()
	p := NewPackageResource(c.Descriptor)
	s := NewPackageSourceResource(c.Descriptor, c.Salt)

	params := map[string]interface{}{"GOOGLE_APPLICATION_CREDENTIALS": "((artifact-gcs-json-key))"}
	if c.Descriptor.Package.IncludeDependencies {
		params["INCLUDE_DEPENDENCIES"] = true
	}

	return Job{
		Name:   "create-package",
		Public: true,
		Plan: []map[string]interface{}{
			{
				"in_parallel": []map[string]interface{}{
					{
						"get":      "build-common",
						"resource": b.Name,
					},
					{
						"get":      "pack",
						"resource": k.Name,
						"params": map[string]interface{}{
							"globs": []string{"pack-*-linux.tgz"},
						},
					},
					{
						"get":      "source",
						"resource": s.Name,
						"trigger":  true,
					},
				},
			},
			{
				"task":   "create-package",
				"file":   "build-common/create-package.yml",
				"params": params,
			},
			{
				"put": p.Name,
				"params": map[string]interface{}{
					"image":           "image/image.tar",
					"additional_tags": "image/tags",
				},
			},
		},
	}
}

func (c CreatePackageContributor) Resources() []Resource {
	return []Resource{
		NewBuildCommonResource(),
		NewPackResource(),
		NewPackageResource(c.Descriptor),
		NewPackageSourceResource(c.Descriptor, c.Salt),
	}
}
