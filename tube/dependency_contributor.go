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

type DependencyContributor struct {
	Descriptor Descriptor
	Dependency Dependency
	Salt       string
}

func (d DependencyContributor) Group() string {
	return "dependency"
}

func (d DependencyContributor) Job() Job {
	b := NewBuildCommonResource()
	e := NewDependencyResource(d.Dependency)
	s := NewSourceResource(d.Descriptor, d.Salt)

	return Job{
		Name:   e.Name,
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
						"resource": e.Name,
						"trigger":  true,
						"params":   d.Dependency.Params,
					},
					{
						"get":      "source",
						"resource": s.Name,
					},
				},
			},
			{
				"task": "update-dependency",
				"file": "build-common/update-dependency.yml",
				"params": map[string]interface{}{
					"DEPENDENCY":      d.Dependency.Name,
					"VERSION_PATTERN": d.Dependency.VersionPattern,
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

func (d DependencyContributor) Resources() []Resource {
	return []Resource{
		NewBuildCommonResource(),
		NewDependencyResource(d.Dependency),
		NewSourceResource(d.Descriptor, d.Salt),
	}
}
