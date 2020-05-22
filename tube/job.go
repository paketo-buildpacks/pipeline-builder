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

type Job struct {
	Name         string                   `yaml:"name,omitempty"`
	Public       bool                     `yaml:"public,omitempty"`
	Plan         []map[string]interface{} `yaml:"plan,omitempty"`
	Serial       bool                     `yaml:"serial,omitempty"`
	SerialGroups []string                 `yaml:"serial_groups,omitempty"`
}

type Jobs map[string]Job

func (j Jobs) Add(jobs ...Job) {
	for _, k := range jobs {
		j[k.Name] = k
	}
}

func (j Jobs) MarshalYAML() (interface{}, error) {
	var k []Job

	for _, v := range j {
		k = append(k, v)
	}

	sort.Slice(k, func(i, j int) bool {
		return k[i].Name < k[j].Name
	})

	return k, nil
}
