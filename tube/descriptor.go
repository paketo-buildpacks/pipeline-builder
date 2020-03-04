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
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Descriptor struct {
	Name string `yaml:"name"`
}

func NewDescriptor(path string) (Descriptor, error) {
	in, err := os.Open(path)
	if err != nil {
		return Descriptor{}, fmt.Errorf("unable to open %s: %w", path, err)
	}
	defer in.Close()

	var d Descriptor
	if err := yaml.NewDecoder(in).Decode(&d); err != nil {
		return Descriptor{}, fmt.Errorf("unable to decode descriptor from %s: %w", path, err)
	}

	return d, nil
}

func (d Descriptor) GitRepository() string {
	return fmt.Sprintf("https://%s.git", d.Name)
}

func (d Descriptor) Owner() string {
	s := strings.Split(d.Name, "/")
	return s[1]
}

func (d Descriptor) Repository() string {
	s := strings.Split(d.Name, "/")
	return s[2]
}

func (d Descriptor) ShortName() string {
	s := strings.Split(d.Name, "/")
	return s[len(s)-1]
}

/*
type Descriptor struct {
	Bucket             string       `yaml:"bucket"`
	Builders           []Builder    `yaml:"builders"`
	Name               string       `yaml:"name"`
	Dependencies       []Dependency `yaml:"dependencies"`
	Upstream           string       `yaml:"upstream"`
}

func (d Descriptor) HasBucket() bool {
	return d.Bucket != ""
}

func (d Descriptor) HasBuilders() bool {
	return len(d.Builders) != 0
}

func (d Descriptor) HasDependencies() bool {
	return len(d.Dependencies) != 0
}

func (d Descriptor) HasModuleDependencies() bool {
	return len(d.ModuleDependencies) != 0
}

func (d Descriptor) HasPackage() bool {
	return d.HasBucket() || d.HasBuilders()
}

func (d Descriptor) HasUpstream() bool {
	return d.Upstream != ""
}

func (d Descriptor) UpstreamDescriptor() Descriptor {
	return Descriptor{Name: d.Upstream}
}

*/
