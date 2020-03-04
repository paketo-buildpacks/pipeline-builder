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
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

type ModuleDependencies struct {
	Descriptor Descriptor
	Modules    []string
	Salt       string
}

func NewModuleDependencies(descriptor Descriptor, salt string) (ModuleDependencies, error) {
	m := ModuleDependencies{
		Descriptor: descriptor,
		Salt:       salt,
	}

	uri := strings.ReplaceAll(fmt.Sprintf("https://%s/master/go.mod", descriptor.Name), "github.com", "raw.githubusercontent.com")
	resp, err := http.Get(uri)
	if err != nil {
		return ModuleDependencies{}, fmt.Errorf("unable to read %s: %w", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return ModuleDependencies{}, fmt.Errorf("could not download %s: %d", uri, resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ModuleDependencies{}, fmt.Errorf("unable to read %s: %w", uri, err)
	}

	re := regexp.MustCompile(`(?mU)^	([\S]+)(?:/v[\d]+)? v[^-\r\n\t\f\v ]+$`)
	for _, s := range re.FindAllStringSubmatch(string(b), -1) {
		m.Modules = append(m.Modules, s[1])
	}

	sort.Strings(m.Modules)

	return m, nil
}

func (m ModuleDependencies) Group() string {
	return "module-dependencies"
}

func (m ModuleDependencies) Job() Job {
	b := NewBuildCommonResource()
	s := NewSourceResource(m.Descriptor, m.Salt)

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

	for _, m := range m.Modules {
		inputs = append(inputs, map[string]interface{}{
			"get":     NewModuleResource(m).Name,
			"trigger": true,
		})
	}

	return Job{
		Name:   "module-dependencies",
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

func (m ModuleDependencies) Resources() []Resource {
	r := []Resource{
		NewBuildCommonResource(),
		NewSourceResource(m.Descriptor, m.Salt),
	}

	for _, m := range m.Modules {
		r = append(r, NewModuleResource(m))
	}

	return r
}
