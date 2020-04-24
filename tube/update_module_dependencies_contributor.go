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
	"regexp"
	"sort"

	"github.com/google/go-github/v30/github"
)

type UpdateModuleDependenciesContributor struct {
	Descriptor Descriptor
	Modules    []string
	Salt       string
}

func NewUpdateModuleDependenciesContributor(descriptor Descriptor, salt string, gh *github.Client) (*UpdateModuleDependenciesContributor, error) {
	file, _, resp, err := gh.Repositories.GetContents(context.Background(), descriptor.Owner(), descriptor.Repository(), "go.mod", nil)
	if resp != nil && resp.StatusCode == 404 {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to get %s/go.mod\n%w", descriptor.Name, err)
	}

	s, err := file.GetContent()
	if err != nil {
		return nil, fmt.Errorf("unable to get %s/go.mod content\n%w", descriptor.Name, err)
	}

	m := UpdateModuleDependenciesContributor{
		Descriptor: descriptor,
		Salt:       salt,
	}

	re := regexp.MustCompile(`(?mU)^	([\S]+)(?:/v[\d]+)? v[^-\r\n\t\f\v ]+$`)
	for _, s := range re.FindAllStringSubmatch(s, -1) {
		m.Modules = append(m.Modules, s[1])
	}

	sort.Strings(m.Modules)

	return &m, nil
}

func (UpdateModuleDependenciesContributor) Group() string {
	return "module-dependencies"
}

func (u UpdateModuleDependenciesContributor) Job() Job {
	b := NewBuildCommonResource()
	s := NewSourceResource(u.Descriptor, u.Salt)

	inputs := []map[string]interface{}{
		{
			"get":      "build-common",
			"resource": b.Name,
		},
		{
			"get":      "source",
			"resource": s.Name,
		},
	}

	for _, m := range u.Modules {
		inputs = append(inputs, map[string]interface{}{
			"get":     NewModuleResource(m).Name,
			"trigger": true,
		})
	}

	return Job{
		Name:   "update-module-dependencies",
		Public: true,
		Plan: []map[string]interface{}{
			{"in_parallel": inputs},
			{
				"task": "update-module-dependencies",
				"file": "build-common/update-module-dependencies.yml",
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

func (u UpdateModuleDependenciesContributor) Resources() []Resource {
	r := []Resource{
		NewBuildCommonResource(),
		NewSourceResource(u.Descriptor, u.Salt),
	}

	for _, m := range u.Modules {
		r = append(r, NewModuleResource(m))
	}

	return r
}
