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

package octo

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/paketo-buildpacks/pipeline-builder/octo/actions"
)

type Descriptor struct {
	GitHub            *GitHub
	Path              string
	CodeOwners        []CodeOwner
	Builder           *Builder
	Package           *Package
	DockerCredentials []DockerCredentials `yaml:"docker_credentials"`
	HttpCredentials   []HTTPCredentials   `yaml:"http_credentials"`
	OfflinePackages   []OfflinePackage    `yaml:"offline_packages"`
	RepublishImages   []RepublishImage    `yaml:"republish_images"`
	Actions           []Action
	Dependencies      []Dependency
	Test              Test
	PackageMatcher    string
}

type GitHub struct {
	Username           string
	Token              string
	Mappers            []string
	BuildpackTOMLPaths map[string]string `yaml:"buildpack_toml_paths"`
}

type Action struct {
	Source string
	Target string
}

type Builder struct {
	Repository string
}

type CodeOwner struct {
	Path  string
	Owner string
}

type DockerCredentials struct {
	Registry string
	Username string
	Password string
}

type HTTPCredentials struct {
	Host     string
	Username string
	Password string
}

type Dependency struct {
	Name           string
	Id             string
	VersionPattern string `yaml:"version_pattern"`
	PURLPattern    string `yaml:"purl_pattern"`
	CPEPattern     string `yaml:"cpe_pattern"`
	Uses           string
	With           map[string]interface{}
}

type OfflinePackage struct {
	Source     string
	Target     string
	SourcePath string `yaml:"source_path"`
	TagPrefix  string `yaml:"tag_prefix"`
	Platform   Platform
}

type RepublishImage struct {
	Source     string
	Target     string
	ID         string
	TagPrefix  string `yaml:"tag_prefix"`
	TargetRepo string `yaml:"target_repo"`
}

type Package struct {
	Repositories        []string
	Repository          string
	IncludeDependencies bool `yaml:"include_dependencies"`
	Register            bool
	RegistryToken       string `yaml:"registry_token"`
	Platform            Platform
	SourcePath          string `yaml:"source_path"`
	Enabled             bool
}

const (
	PlatformLinux   = "linux"
	PlatformWindows = "windows"
)

type Platform struct {
	OS string
}

type Test struct {
	Steps []actions.Step
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

	if d.GitHub == nil {
		return Descriptor{}, fmt.Errorf("github is required")
	}

	if d.Path == "" {
		d.Path = ".."
	}

	if !filepath.IsAbs(d.Path) {
		if d.Path, err = filepath.Abs(filepath.Join(filepath.Dir(path), d.Path)); err != nil {
			return Descriptor{}, fmt.Errorf("unable to Find absolute path\n%w", err)
		}
	}

	for i, e := range d.Dependencies {
		if e.Name == "" {
			e.Name = e.Id
			d.Dependencies[i] = e
		}

		if e.VersionPattern == "" {
			e.VersionPattern = `[\d]+\.[\d]+\.[\d]+`
			d.Dependencies[i] = e
		}
	}

	if d.Package == nil {
		d.Package = &Package{}
	} else {
		d.Package.Enabled = true
	}

	if d.Package.Platform.OS == "" {
		d.Package.Platform.OS = PlatformLinux
	}

	for i, o := range d.OfflinePackages {
		if o.Platform.OS == "" {
			o.Platform.OS = PlatformLinux
			d.OfflinePackages[i] = o
		}
	}

	if d.Test.Steps == nil {
		d.Test.Steps = []actions.Step{
			{
				Name: "Install richgo",
				Run:  StatikString("/install-richgo.sh"),
				Env:  map[string]string{"RICHGO_VERSION": RichGoVersion},
			},
			{
				Name: "Run Tests",
				Run:  StatikString("/run-tests.sh"),
				Env:  map[string]string{"RICHGO_FORCE_COLOR": "1"},
			},
		}
	}

	return d, nil
}
