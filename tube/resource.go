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
	"sort"
)

type Resource struct {
	Name       string                 `yaml:"name,omitempty"`
	Type       string                 `yaml:"type,omitempty"`
	Icon       string                 `yaml:"icon,omitempty"`
	Source     map[string]interface{} `yaml:"source,omitempty"`
	CheckEvery string                 `yaml:"check_every,omitempty"`
	WebHook    WebHook                `yaml:"webhook_token,omitempty"`
}

func NewBuildCommonResource() Resource {
	d := Descriptor{Name: "github.com/paketo-buildpacks/build-common"}

	return Resource{
		Name: d.ShortName(),
		Type: "git",
		Icon: "github-circle",
		Source: map[string]interface{}{
			"uri":      d.GitRepository(),
			"branch":   "master",
			"username": "((github-username))",
			"password": "((github-access-token))",
		},
	}
}

func NewDependencyResource(dependency Dependency) Resource {
	return Resource{
		Name:   fmt.Sprintf("dependency:%s", dependency.Resource),
		Type:   dependency.Type,
		Icon:   dependency.Icon,
		Source: dependency.Source,
	}
}

func NewModuleResource(name string) Resource {
	d := Descriptor{Name: name}

	return Resource{
		Name: fmt.Sprintf("module:%s", name),
		Type: "git",
		Icon: "github-circle",
		Source: map[string]interface{}{
			"uri":        d.GitRepository(),
			"tag_filter": "v*",
			"username":   "((github-username))",
			"password":   "((github-access-token))",
		},
	}
}

func NewPackResource() Resource {
	return Resource{
		Name:       "pack",
		Type:       "github-release",
		Icon:       "package-variant-closed",
		CheckEvery: "4h",
		Source: map[string]interface{}{
			"owner":        "buildpacks",
			"repository":   "pack",
			"access_token": "((github-access-token))",
		},
	}
}

func NewPackageResource(descriptor Descriptor) Resource {
	return Resource{
		Name: fmt.Sprintf("package:%s", descriptor.Package.Repository),
		Type: "registry-image",
		Icon: "docker",
		Source: map[string]interface{}{
			"repository": descriptor.Package.Repository,
			"username":   descriptor.Package.Username,
			"password":   descriptor.Package.Password,
		},
	}
}

func NewPackageSourceResource(descriptor Descriptor, salt string) Resource {
	return Resource{
		Name:       fmt.Sprintf("package-source:%s", descriptor.ShortName()),
		Type:       "git",
		Icon:       "github-circle",
		CheckEvery: "4h",
		WebHook:    NewWebHook(salt, descriptor.Owner(), descriptor.Repository()),
		Source: map[string]interface{}{
			"uri":        descriptor.GitRepository(),
			"tag_filter": "v*",
			"username":   "((github-username))",
			"password":   "((github-access-token))",
		},
	}
}

func NewSourceResource(descriptor Descriptor, salt string) Resource {
	return Resource{
		Name:       fmt.Sprintf("source:%s", descriptor.ShortName()),
		Type:       "git",
		Icon:       "github-circle",
		CheckEvery: "4h",
		WebHook:    NewWebHook(salt, descriptor.Owner(), descriptor.Repository()),
		Source: map[string]interface{}{
			"uri":      descriptor.GitRepository(),
			"branch":   "master",
			"username": "((github-username))",
			"password": "((github-access-token))",
		},
	}
}

func NewVersionResource(descriptor Descriptor) Resource {
	return Resource{
		Name: fmt.Sprintf("version:%s", descriptor.ShortName()),
		Type: "semver",
		Icon: "tag",
		Source: map[string]interface{}{
			"driver":   "gcs",
			"bucket":   "repository-versions",
			"key":      descriptor.Name,
			"json_key": "((gcs-json-key))",
		},
	}
}

type Resources map[string]Resource

func (r Resources) Add(resources ...Resource) {
	for _, s := range resources {
		r[s.Name] = s
	}
}

func (r Resources) MarshalYAML() (interface{}, error) {
	var s []Resource

	for _, v := range r {
		s = append(s, v)
	}

	sort.Slice(s, func(i, j int) bool {
		return s[i].Name < s[j].Name
	})

	return s, nil
}
