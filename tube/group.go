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
	"sort"
)

type Group struct {
	Name string   `yaml:"name,omitempty"`
	Jobs []string `yaml:"jobs,omitempty"`
}

type Groups map[string]Group

func (g Groups) Add(name string, job string) {
	h, ok := g[name]
	if !ok {
		h = Group{Name: name}
	}

	for _, j := range h.Jobs {
		if j == job {
			return
		}
	}
	h.Jobs = append(h.Jobs, job)

	g[name] = h
}

func (g Groups) MarshalYAML() (interface{}, error) {
	var h []Group

	for _, v := range g {
		sort.Strings(v.Jobs)
		h = append(h, v)
	}

	sort.Slice(h, func(i, j int) bool {
		return h[i].Name < h[j].Name
	})

	return h, nil
}
