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

type PackageContributor struct {
	Descriptor Descriptor
	Salt       string
}

func (p PackageContributor) Group() string {
	return "build"
}

func (p PackageContributor) Job() Job {
	b := NewBuildCommonResource()
	q := NewPackResource()
	r := NewPackageResource(p.Descriptor)
	s := NewPackageSourceResource(p.Descriptor, p.Salt)

	inputs := []map[string]interface{}{
		{
			"get":      "build-common",
			"resource": b.Name,
		},
		{
			"get":      "pack",
			"resource": q.Name,
			"params": map[string]interface{}{
				"globs": []string{"pack-*-linux.tgz"},
			},
		},
		{
			"get":      "source",
			"resource": s.Name,
			"trigger":  true,
		},
	}

	return Job{
		Name:   "package",
		Public: true,
		Plan: []map[string]interface{}{
			{"in_parallel": inputs},
			{
				"task": "create-package",
				"file": "build-common/create-package.yml",
			},
			{
				"put": r.Name,
				"params": map[string]interface{}{
					"image":           "image/image.tar",
					"additional_tags": "image/tags",
				},
			},
		},
	}
}

func (p PackageContributor) Resources() []Resource {
	return []Resource{
		NewBuildCommonResource(),
		NewPackResource(),
		NewPackageResource(p.Descriptor),
		NewPackageSourceResource(p.Descriptor, p.Salt),
	}
}
