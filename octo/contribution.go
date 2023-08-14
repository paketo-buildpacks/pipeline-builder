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

	"github.com/disiqueira/gotree"
	"github.com/iancoleman/strcase"
	"gopkg.in/yaml.v3"

	"github.com/paketo-buildpacks/pipeline-builder/v2/octo/actions"
	"github.com/paketo-buildpacks/pipeline-builder/v2/octo/actions/event"
	"github.com/paketo-buildpacks/pipeline-builder/v2/octo/dependabot"
	"github.com/paketo-buildpacks/pipeline-builder/v2/octo/labels"
	"github.com/paketo-buildpacks/pipeline-builder/v2/octo/release"
)

type Contribution struct {
	Path        string
	Permissions os.FileMode
	Structure   gotree.Tree
	Content     []byte
	Namespace   string
}

func NewActionContributionWithNamespace(namespace string, workflow actions.Workflow) (Contribution, error) {
	var (
		c   Contribution
		err error
	)

	var workflowName string
	if namespace == "" {
		workflowName = fmt.Sprintf("%s.yml", strcase.ToKebab(workflow.Name))
	} else {
		workflowName = fmt.Sprintf("%s-%s.yml", namespace, strcase.ToKebab(workflow.Name))
	}

	c.Path = filepath.Join(".github", "workflows", workflowName)
	c.Permissions = 0644

	var t []event.Type
	for k := range workflow.On {
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

// Deprecated: use NewActionContributionWithNamespace instead
func NewActionContribution(workflow actions.Workflow) (Contribution, error) {
	return NewActionContributionWithNamespace("", workflow)
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
			If:   "${{ (github.event_name != 'pull_request' || ! github.event.pull_request.head.repo.fork) && (github.actor != 'dependabot[bot]') }}",
			Uses: "docker/login-action@v2",
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
			Run:  StatikString("/update-netrc.sh"),
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
