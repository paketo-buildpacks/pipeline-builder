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

type TestContributor struct {
	Descriptor Descriptor
	Salt       string
}

func (TestContributor) Group() string {
	return "main"
}

func (t TestContributor) Job() Job {
	b := NewBuildCommonResource()
	s := NewSourceResource(t.Descriptor, t.Salt)

	return Job{
		Name:   "test",
		Public: true,
		Plan: []map[string]interface{}{
			{
				"in_parallel": []map[string]interface{}{
					{
						"get":      "build-common",
						"resource": b.Name,
					},
					{
						"get":      "source",
						"resource": s.Name,
						"trigger":  true,
					},
				},
			},
			{
				"task": "test",
				"file": "build-common/test.yml",
			},
		},
	}
}

func (t TestContributor) Resources() []Resource {
	return []Resource{
		NewBuildCommonResource(),
		NewSourceResource(t.Descriptor, t.Salt),
	}
}
