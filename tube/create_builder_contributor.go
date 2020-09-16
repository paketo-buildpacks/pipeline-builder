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

type CreateBuilderContributor struct {
	Descriptor Descriptor
	Salt       string
}

func (CreateBuilderContributor) Group() string {
	return "build"
}

func (c CreateBuilderContributor) Job() Job {
	b := NewBuilderResource(c.Descriptor)
	d := NewBuildCommonResource()
	k := NewPackResource()
	s := NewBuilderSourceResource(c.Descriptor, c.Salt)

	return Job{
		Name:   "create-builder",
		Public: true,
		Plan: []map[string]interface{}{
			{
				"in_parallel": []map[string]interface{}{
					{
						"get":      "build-common",
						"resource": d.Name,
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
				"task": "create-builder",
				"file": "build-common/create-builder.yml",
				"params": map[string]interface{}{
					"DOCKER_REGISTRY_CREDENTIALS":    "((docker-registry-credentials))",
					"GOOGLE_APPLICATION_CREDENTIALS": "((artifact-gcs-json-key))",
				},
			},
			{
				"put": b.Name,
				"params": map[string]interface{}{
					"image":           "image/image.tar",
					"additional_tags": "image/tags",
				},
			},
		},
	}
}

func (c CreateBuilderContributor) Resources() []Resource {
	return []Resource{
		NewBuilderResource(c.Descriptor),
		NewBuildCommonResource(),
		NewPackResource(),
		NewBuilderSourceResource(c.Descriptor, c.Salt),
	}
}
