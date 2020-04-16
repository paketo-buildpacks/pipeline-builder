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

type UpdatePackageDependencyContributor struct {
	Descriptor Descriptor
	Dependency Dependency
	Salt       string
}

func (UpdatePackageDependencyContributor) Group() string {
	return "dependencies"
}

func (u UpdatePackageDependencyContributor) Job() Job {
	b := NewBuildCommonResource()
	p := NewPackageDependencyResource(u.Dependency)
	s := NewSourceResource(u.Descriptor, u.Salt)

	return Job{
		Name:   p.Name,
		Public: true,
		Plan: []map[string]interface{}{
			{
				"in_parallel": []map[string]interface{}{
					{
						"get":      "build-common",
						"resource": b.Name,
					},
					{
						"get":      "dependency",
						"resource": p.Name,
						"trigger":  true,
						"params":   u.Dependency.Params,
					},
					{
						"get":      "source",
						"resource": s.Name,
					},
				},
			},
			{
				"task": "update-package-dependency",
				"file": "build-common/update-package-dependency.yml",
				"params": map[string]interface{}{
					"DEPENDENCY":      u.Dependency.Name,
					"VERSION_PATTERN": u.Dependency.VersionPattern,
				},
			},
			{
				"put": s.Name,
				"params": map[string]interface{}{
					"repository": "source",
					"rebase":     true,
				},
			},
		},
	}

}

func (u UpdatePackageDependencyContributor) Resources() []Resource {
	return []Resource{
		NewBuildCommonResource(),
		NewPackageDependencyResource(u.Dependency),
		NewSourceResource(u.Descriptor, u.Salt),
	}
}
