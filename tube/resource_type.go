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

var KnownResourceTypes = map[string]ResourceType{
	"adopt-openjdk-resource": {
		Name: "adopt-openjdk-resource",
		Type: "registry-image",
		Source: map[string]interface{}{
			"repository": "paketoio/adopt-openjdk-resource",
		},
	},
	"corretto-resource": {
		Name: "corretto-resource",
		Type: "registry-image",
		Source: map[string]interface{}{
			"repository": "paketoio/amazon-corretto-resource",
		},
	},
	"git": {
		Name: "git",
		Type: "registry-image",
		Source: map[string]interface{}{
			"repository": "concourse/git-resource",
		},
	},
	"github-release": {
		Name: "github-release",
		Type: "registry-image",
		Source: map[string]interface{}{
			"repository": "concourse/github-release-resource",
		},
	},
	"gradle-resource": {
		Name: "gradle-resource",
		Type: "registry-image",
		Source: map[string]interface{}{
			"repository": "paketoio/gradle-resource",
		},
	},
	"jprofiler-resource": {
		Name: "jprofiler-resource",
		Type: "registry-image",
		Source: map[string]interface{}{
			"repository": "paketo/jprofiler-resource",
		},
	},
	"maven-resource": {
		Name: "maven-resource",
		Type: "registry-image",
		Source: map[string]interface{}{
			"repository": "paketo/maven-resource",
		},
	},
	"npm-resource": {
		Name: "npm-resource",
		Type: "registry-image",
		Source: map[string]interface{}{
			"repository": "paketo/npm-resource",
		},
	},
	"registry-image": {
		Name: "registry-image",
		Type: "registry-image",
		Source: map[string]interface{}{
			"repository": "concourse/registry-image-resource",
		},
	},
	"repository-resource": {
		Name: "repository-resource",
		Type: "registry-image",
		Source: map[string]interface{}{
			"repository": "paketo/repository-resource",
		},
	},
	"semver": {
		Name: "semver",
		Type: "registry-image",
		Source: map[string]interface{}{
			"repository": "concourse/semver-resource",
		},
	},
	"sky-walking-resource": {
		Name: "sky-walking-resource",
		Type: "registry-image",
		Source: map[string]interface{}{
			"repository": "paketo/sky-walking-resource",
		},
	},
	"tomcat-resource": {
		Name: "tomcat-resource",
		Type: "registry-image",
		Source: map[string]interface{}{
			"repository": "paketo/tomcat-resource",
		},
	},
	"zulu-resource": {
		Name: "zulu-resource",
		Type: "registry-image",
		Source: map[string]interface{}{
			"repository": "paketo/azul-zulu-resource",
		},
	},
}

type ResourceType struct {
	Name   string                 `yaml:"name,omitempty"`
	Type   string                 `yaml:"type,omitempty"`
	Source map[string]interface{} `yaml:"source,omitempty"`
}

type ResourceTypes map[string]ResourceType

func (r ResourceTypes) Add(resourceTypes ...ResourceType) {
	for _, s := range resourceTypes {
		r[s.Name] = s
	}
}

func (r ResourceTypes) MarshalYAML() (interface{}, error) {
	var s []ResourceType

	for _, v := range r {
		s = append(s, v)
	}

	sort.Slice(s, func(i, j int) bool {
		return s[i].Name < s[j].Name
	})

	return s, nil
}
