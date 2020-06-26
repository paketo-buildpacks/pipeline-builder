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

type Dependency struct {
	Name           string                 `yaml:"name"`
	VersionPattern string                 `yaml:"version_pattern"`
	Resource       string                 `yaml:"resource"`
	Type           string                 `yaml:"type"`
	Icon           string                 `yaml:"icon"`
	Source         map[string]interface{} `yaml:"source"`
	Params         map[string]interface{} `yaml:"params"`
}

type Builder struct {
	Repository string `yaml:"repository"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`

	BuildImage string `yaml:"build_image"`
	RunImage   string `yaml:"run_image"`
}

type Package struct {
	Repository string `yaml:"repository"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`

	IncludeDependencies bool `yaml:"include_dependencies"`
}

type Descriptor struct {
	Name                   string       `yaml:"name"`
	Builder                *Builder     `yaml:"builder"`
	Package                *Package     `yaml:"package"`
	Dependencies           []Dependency `yaml:"dependencies"`
	SkipAdditional         bool         `yaml:"skip_additional"`
	SkipModuleDependencies bool         `yaml:"skip_module_dependencies"`
	SkipRelease            bool         `yaml:"skip_release"`
	SkipTest               bool         `yaml:"skip_test"`

	ResourceTypes []ResourceType `yaml:"resource_types"`
	Resources     []Resource     `yaml:"resources"`
	Groups        []Group        `yaml:"groups"`
	Jobs          []Job          `yaml:"jobs"`
}

func NewDescriptor(path string) (Descriptor, error) {
	in, err := os.Open(path)
	if err != nil {
		return Descriptor{}, fmt.Errorf("unable to open %s\n%w", path, err)
	}
	defer in.Close()

	var d Descriptor
	if err := yaml.NewDecoder(in).Decode(&d); err != nil {
		return Descriptor{}, fmt.Errorf("unable to decode descriptor from %s\n%w", path, err)
	}

	for i, dep := range d.Dependencies {
		if dep.Resource == "" {
			d.Dependencies[i].Resource = dep.Name
		}

		if dep.VersionPattern == "" {
			d.Dependencies[i].VersionPattern = `[\d]+(?:\.[\d]+(?:\.[\d]+)?)?`
		}
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
