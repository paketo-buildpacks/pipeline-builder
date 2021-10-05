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

package actions

import (
	"github.com/paketo-buildpacks/pipeline-builder/octo/actions/event"
)

type Workflow struct {
	Name     string                     `yaml:",omitempty"`
	On       map[event.Type]event.Event `yaml:",omitempty"`
	Env      map[string]string          `yaml:",omitempty"`
	Defaults Defaults                   `yaml:",omitempty"`
	Jobs     map[string]Job             `yaml:",omitempty"`
}

type Job struct {
	Name            string               `yaml:",omitempty"`
	If              string               `yaml:",omitempty"`
	Needs           []string             `yaml:",omitempty"`
	RunsOn          []VirtualEnvironment `yaml:"runs-on,omitempty"`
	Outputs         map[string]string    `yaml:",omitempty"`
	Env             map[string]string    `yaml:",omitempty"`
	Defaults        Defaults             `yaml:",omitempty"`
	Steps           []Step               `yaml:",omitempty"`
	TimeoutMinutes  int                  `yaml:",omitempty"`
	Strategy        Strategy             `yaml:",omitempty"`
	ContinueOnError bool                 `yaml:"continue-on-error,omitempty"`
	Container       Container            `yaml:",omitempty"`
	Services        map[string]Service   `yaml:",omitempty"`
}

type Run struct {
	Shell            Shell  `yaml:",omitempty"`
	WorkingDirectory string `yaml:"working-directory,omitempty"`
}

type Step struct {
	Name             string            `yaml:",omitempty"`
	If               string            `yaml:",omitempty"`
	Id               string            `yaml:",omitempty"`
	Uses             string            `yaml:",omitempty"`
	Run              string            `yaml:",omitempty"`
	WorkingDirectory string            `yaml:"working-directory,omitempty"`
	Shell            Shell             `yaml:",omitempty"`
	With             With              `yaml:",omitempty"`
	Env              map[string]string `yaml:",omitempty"`
	ContinueOnError  bool              `yaml:"continue-on-error,omitempty"`
	TimeoutMinutes   int               `yaml:",omitempty"`
}

type Strategy struct {
	Matrix      Matrix `yaml:",omitempty"`
	FailFast    bool   `yaml:"fail-fast,omitempty"`
	MaxParallel int    `yaml:"max-parallel:omitempty"`
}

type Container struct {
	Image   string            `yaml:",omitempty"`
	Env     map[string]string `yaml:",omitempty"`
	Ports   []string          `yaml:",omitempty"`
	Volumes []string          `yaml:",omitempty"`
	Options string            `yaml:",omitempty"`
}

type Service struct {
	Image   string            `yaml:",omitempty"`
	Env     map[string]string `yaml:",omitempty"`
	Ports   []string          `yaml:",omitempty"`
	Volumes []string          `yaml:",omitempty"`
	Options string            `yaml:",omitempty"`
}
