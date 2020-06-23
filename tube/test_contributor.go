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

import (
	"context"
	"fmt"

	"github.com/google/go-github/v30/github"
)

type TestContributor struct {
	Code       bool
	Descriptor Descriptor
	Salt       string
}

func NewTestContributor(descriptor Descriptor, salt string, gh *github.Client) (TestContributor, error) {
	t := TestContributor{
		Descriptor: descriptor,
		Salt:       salt,
	}

	_, _, resp, err := gh.Repositories.GetContents(context.Background(), descriptor.Owner(), descriptor.Repository(), "go.mod", nil)
	if resp != nil && resp.StatusCode == 404 {
		return t, nil
	} else if err != nil {
		return TestContributor{}, fmt.Errorf("unable to get %s/go.mod\n%w", descriptor.Name, err)
	}

	t.Code = true
	return t, nil
}

func (TestContributor) Group() string {
	return "build"
}

func (t TestContributor) Job() Job {
	b := NewBuildCommonResource()
	s := NewSourceResource(t.Descriptor, t.Salt)

	inputs := []map[string]interface{}{
		{
			"get":      "build-common",
			"resource": b.Name,
		},
		{
			"get":      "source",
			"resource": s.Name,
			"trigger":  true,
		},
	}

	var jobs []map[string]interface{}

	if t.Code {
		jobs = append(jobs, map[string]interface{}{
			"task": "test",
			"file": "build-common/test.yml",
		})
	}

	if t.Descriptor.Package != nil {
		inputs = append(inputs, map[string]interface{}{
			"get":      "pack",
			"resource": NewPackResource().Name,
			"params": map[string]interface{}{
				"globs": []string{"pack-*-linux.tgz"},
			},
		})

		jobs = append(jobs, map[string]interface{}{
			"task": "create-package",
			"file": "build-common/create-package.yml",
			"params": map[string]interface{}{
				"GOOGLE_APPLICATION_CREDENTIALS": "((artifact-gcs-json-key))",
				"INCLUDE_DEPENDENCIES":           true,
			},
		})
	}

	if t.Descriptor.Builder != nil {
		inputs = append(inputs, map[string]interface{}{
			"get":      "pack",
			"resource": NewPackResource().Name,
			"params": map[string]interface{}{
				"globs": []string{"pack-*-linux.tgz"},
			},
		})

		jobs = append(jobs, map[string]interface{}{
			"task": "create-builder",
			"file": "build-common/create-builder.yml",
			"params": map[string]interface{}{
				"GOOGLE_APPLICATION_CREDENTIALS": "((artifact-gcs-json-key))",
			},
		})
	}

	return Job{
		Name:   "test",
		Public: true,
		Serial: true,
		Plan: []map[string]interface{}{
			{"in_parallel": inputs},
			{"in_parallel": jobs},
		},
	}
}

func (t TestContributor) Resources() []Resource {
	r := []Resource{
		NewBuildCommonResource(),
		NewSourceResource(t.Descriptor, t.Salt),
	}

	if t.Descriptor.Package != nil {
		r = append(r, NewPackResource())
	}

	return r
}
