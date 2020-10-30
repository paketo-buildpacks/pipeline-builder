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
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/disiqueira/gotree"
	"github.com/iancoleman/strcase"
	"gopkg.in/yaml.v3"

	"github.com/paketo-buildpacks/pipeline-builder/octo/actions"
	"github.com/paketo-buildpacks/pipeline-builder/octo/actions/event"
	"github.com/paketo-buildpacks/pipeline-builder/octo/dependabot"
	"github.com/paketo-buildpacks/pipeline-builder/octo/labels"
	"github.com/paketo-buildpacks/pipeline-builder/octo/release"
)

//go:generate statik -src . -include *.sh

const (
	GoVersion   = "1.15"
	PackVersion = "0.14.1"
	YJVersion   = "5.0.0"
)

type Octo struct {
	DescriptorPath string
}

func (o Octo) Contribute() error {
	descriptor, err := NewDescriptor(o.DescriptorPath)
	if err != nil {
		return fmt.Errorf("unable to read descriptor\n%w", err)
	}

	var contributions []Contribution

	contributions = append(contributions, ContributeCodeOwners(descriptor))

	if c, err := ContributeActions(descriptor); err != nil {
		return err
	} else {
		contributions = append(contributions, c...)
	}

	if c, err := ContributeBuilderDependencies(descriptor); err != nil {
		return err
	} else {
		contributions = append(contributions, c...)
	}

	if c, err := ContributeCreateBuilder(descriptor); err != nil {
		return err
	} else if c != nil {
		contributions = append(contributions, *c)
	}

	if c, err := ContributeCreatePackage(descriptor); err != nil {
		return err
	} else if c != nil {
		contributions = append(contributions, *c)
	}

	if c, err := ContributeDependabot(descriptor); err != nil {
		return err
	} else {
		contributions = append(contributions, c)
	}

	if c, err := ContributeBuildpackDependencies(descriptor); err != nil {
		return err
	} else {
		contributions = append(contributions, c...)
	}

	if c, err := ContributeDraftRelease(descriptor); err != nil {
		return err
	} else {
		contributions = append(contributions, c...)
	}

	if c, err := ContributeLabels(); err != nil {
		return err
	} else {
		contributions = append(contributions, c...)
	}

	if c, err := ContributeOfflinePackages(descriptor); err != nil {
		return err
	} else {
		contributions = append(contributions, c...)
	}

	if c, err := ContributePackageDependencies(descriptor); err != nil {
		return err
	} else {
		contributions = append(contributions, c...)
	}

	if c, err := ContributeTest(descriptor); err != nil {
		return err
	} else if c != nil {
		contributions = append(contributions, *c)
	}

	return o.Write(descriptor, contributions)
}

func (Octo) Write(descriptor Descriptor, contributions []Contribution) error {
	t := gotree.New(descriptor.Path)

	sort.Slice(contributions, func(i, j int) bool {
		return contributions[i].Structure.Text() < contributions[j].Structure.Text()
	})

	for _, c := range contributions {
		t.AddTree(c.Structure)

		file := filepath.Join(descriptor.Path, c.Path)

		if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
			return fmt.Errorf("unable to create %s\n%w", filepath.Dir(file), err)
		}

		if err := ioutil.WriteFile(file, c.Content, c.Permissions); err != nil {
			return fmt.Errorf("unable to write %s\n%w", file, err)
		}
	}

	fmt.Println(t.Print())
	return nil
}

type Contribution struct {
	Path        string
	Permissions os.FileMode
	Structure   gotree.Tree
	Content     []byte
}

func NewActionContribution(workflow actions.Workflow) (Contribution, error) {
	var (
		c   Contribution
		err error
	)
	c.Path = filepath.Join(".github", "workflows", fmt.Sprintf("%s.yml", strcase.ToKebab(workflow.Name)))
	c.Permissions = 0644

	var t []event.Type
	for k, _ := range workflow.On {
		t = append(t, k)
	}

	c.Structure = gotree.New(fmt.Sprintf("%s %s", c.Path, t))
	for _, j := range workflow.Jobs {
		c.Structure.Add(j.Name)
	}

	if c.Content, err = yaml.Marshal(workflow); err != nil {
		return Contribution{}, fmt.Errorf("unable to marshal workflow\n%w", err)
	}

	return c, err
}

func NewDependabotContribution(dependabot dependabot.Dependabot) (Contribution, error) {
	var (
		c   Contribution
		err error
	)
	c.Path = filepath.Join(".github", "dependabot.yml")
	c.Permissions = 0644

	c.Structure = gotree.New(c.Path)
	for _, u := range dependabot.Updates {
		c.Structure.Add(fmt.Sprintf("%s: %s", u.PackageEcosystem, u.Schedule.Interval))
	}

	if c.Content, err = yaml.Marshal(dependabot); err != nil {
		return Contribution{}, fmt.Errorf("unable to marshal dependabot\n%w", err)
	}

	return c, err
}

func NewDockerCredentialActions(credentials []DockerCredentials) []actions.Step {
	var s []actions.Step

	for _, c := range credentials {
		s = append(s, actions.Step{
			Name: fmt.Sprintf("Docker login %s", c.Registry),
			If:   "${{ github.event_name != 'pull_request' || ! github.event.pull_request.head.repo.fork }}",
			Uses: "docker/login-action@v1",
			With: map[string]interface{}{
				"registry": c.Registry,
				"username": c.Username,
				"password": c.Password,
			},
		})
	}

	return s
}

func NewHttpCredentialActions(credentials []HTTPCredentials) []actions.Step {
	var s []actions.Step

	for _, c := range credentials {
		s = append(s, actions.Step{
			Name: fmt.Sprintf("HTTP login %s", c.Host),
			If:   "${{ github.event_name != 'pull_request' || ! github.event.pull_request.head.repo.fork }}",
			Run:  statikString("/update-netrc.sh"),
			Env: map[string]string{
				"HOST":     c.Host,
				"USERNAME": c.Username,
				"PASSWORD": c.Password,
			},
		})
	}

	return s
}

func NewDrafterContribution(drafter release.Drafter) (Contribution, error) {
	var (
		c   Contribution
		err error
	)
	c.Path = filepath.Join(".github", "release-drafter.yml")
	c.Permissions = 0644

	c.Structure = gotree.New(c.Path)
	for _, a := range drafter.Categories {
		c.Structure.Add(fmt.Sprintf("%s: %s", a.Title, a.Labels))
	}

	if c.Content, err = yaml.Marshal(drafter); err != nil {
		return Contribution{}, fmt.Errorf("unable to marshal release drafter\n%w", err)
	}

	return c, err
}

func NewLabelsContribution(labels []labels.Label) (Contribution, error) {
	var (
		c   Contribution
		err error
	)
	c.Path = filepath.Join(".github", "labels.yml")
	c.Permissions = 0644

	c.Structure = gotree.New(c.Path)
	for _, l := range labels {
		c.Structure.Add(l.Name)
	}

	if c.Content, err = yaml.Marshal(labels); err != nil {
		return Contribution{}, fmt.Errorf("unable to marshal labels\n%w", err)
	}

	return c, err
}

type Descriptor struct {
	Path              string
	CodeOwners        []CodeOwner
	Builder           *Builder
	Package           *Package
	DockerCredentials []DockerCredentials `yaml:"docker_credentials"`
	HttpCredentials   []HTTPCredentials   `yaml:"http_credentials"`
	OfflinePackages   []OfflinePackage    `yaml:"offline_packages"`
	Actions           []Action
	Dependencies      []Dependency
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
	Uses           string
	With           map[string]interface{}
}

type OfflinePackage struct {
	Source string
	Target string
}

type Package struct {
	Repository          string
	IncludeDependencies bool `yaml:"include_dependencies"`
	Register            bool
	RegistryToken       string `yaml:"registry_token"`
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

	if !filepath.IsAbs(d.Path) {
		if d.Path, err = filepath.Abs(filepath.Join(filepath.Dir(path), d.Path)); err != nil {
			return Descriptor{}, fmt.Errorf("unable to find absolute path\n%w", err)
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

	return d, nil
}
