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
)

//go:generate statik -src . -include *.sh

const (
	GoVersion   = "1.15"
	PackVersion = "0.14.1"
	YJVersion   = "5.0.0"
)

var RemovedFiles []string

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

	if c, err := ContributeUpdatePipeline(); err != nil {
		return err
	} else {
		contributions = append(contributions, c)
	}

	if err := o.Remove(descriptor, RemovedFiles); err != nil {
		return err
	}

	return o.Write(descriptor, contributions)
}

func (Octo) Remove(descriptor Descriptor, removals []string) error {
	for _, r := range removals {
		file := filepath.Join(descriptor.Path, r)
		fmt.Printf("Removing %s\n", r)

		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("unable to remove %s\n%w", r, err)
		}
	}

	return nil
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
