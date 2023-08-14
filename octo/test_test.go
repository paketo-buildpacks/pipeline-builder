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

package octo_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	octo "github.com/paketo-buildpacks/pipeline-builder/v2/octo"
	"github.com/paketo-buildpacks/pipeline-builder/v2/octo/actions"
	"github.com/sclevine/spec"
	"gopkg.in/yaml.v3"
)

// use our own type as the events cannot be unmarshalled
type jobs struct {
	Jobs map[string]*actions.Job `yaml:"jobs"`
}

func testTestGeneration(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		dir        string
		descriptor octo.Descriptor
	)

	context("Generate Tests for Buildpack", func() {
		it.Before(func() {
			var err error
			dir, err = ioutil.TempDir("", "main-package")
			Expect(err).To(Not(HaveOccurred()))

			Expect(os.Mkdir(filepath.Join(dir, ".github"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(dir, ".github", "pipeline-descriptor.yaml"), []byte(`---
github:
  username: ${{ secrets.JAVA_GITHUB_USERNAME }}
  token:    ${{ secrets.JAVA_GITHUB_TOKEN }}

package:
  repository:     gcr.io/paketo-buildpacks/dummy
  register:       true
  registry_token: ${{ secrets.JAVA_GITHUB_TOKEN }}
`), 0644)).To(Succeed())

			descriptor, err = octo.NewDescriptor(filepath.Join(dir, ".github", "pipeline-descriptor.yaml"))
			Expect(err).To(Not(HaveOccurred()))

			Expect(ioutil.WriteFile(filepath.Join(dir, "main.go"), []byte{}, 0644)).To(Succeed())

			Expect(ioutil.WriteFile(filepath.Join(dir, "buildpack.toml"), []byte{}, 0644)).To(Succeed())
		})

		it.After(func() {
			Expect(os.RemoveAll(dir)).To(Succeed())
		})

		it("will contribute a unit test pipeline", func() {
			contribution, err := octo.ContributeTest(descriptor)
			Expect(err).To(Not(HaveOccurred()))

			Expect(contribution.Path).To(Equal(".github/workflows/pb-tests.yml"))

			var workflow jobs
			Expect(yaml.Unmarshal(contribution.Content, &workflow)).To(Succeed())

			Expect(len(workflow.Jobs)).To(Equal(2))
			Expect(workflow.Jobs["unit"]).To(Not(BeNil()))
			Expect(workflow.Jobs["integration"]).To(BeNil())

			steps := workflow.Jobs["unit"].Steps
			Expect(steps[len(steps)-1].Run).Should(ContainSubstring("richgo test ./..."))
		})
	})

	context("Generate Tests for Extension", func() {
		it.Before(func() {
			var err error
			dir, err = ioutil.TempDir("", "main-package")
			Expect(err).To(Not(HaveOccurred()))

			Expect(os.Mkdir(filepath.Join(dir, ".github"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(dir, ".github", "pipeline-descriptor.yaml"), []byte(`---
github:
  username: ${{ secrets.JAVA_GITHUB_USERNAME }}
  token:    ${{ secrets.JAVA_GITHUB_TOKEN }}

package:
  repository:     gcr.io/paketo-buildpacks/dummy
  register:       true
  registry_token: ${{ secrets.JAVA_GITHUB_TOKEN }}
`), 0644)).To(Succeed())

			descriptor, err = octo.NewDescriptor(filepath.Join(dir, ".github", "pipeline-descriptor.yaml"))
			Expect(err).To(Not(HaveOccurred()))

			Expect(ioutil.WriteFile(filepath.Join(dir, "main.go"), []byte{}, 0644)).To(Succeed())

			Expect(ioutil.WriteFile(filepath.Join(dir, "extension.toml"), []byte{}, 0644)).To(Succeed())
		})

		it.After(func() {
			Expect(os.RemoveAll(dir)).To(Succeed())
		})

		it("will contribute a unit test pipeline", func() {
			contribution, err := octo.ContributeTest(descriptor)
			Expect(err).To(Not(HaveOccurred()))

			Expect(contribution.Path).To(Equal(".github/workflows/pb-tests.yml"))

			var workflow jobs
			Expect(yaml.Unmarshal(contribution.Content, &workflow)).To(Succeed())

			Expect(len(workflow.Jobs)).To(Equal(2))
			Expect(workflow.Jobs["unit"]).To(Not(BeNil()))
			Expect(workflow.Jobs["integration"]).To(BeNil())

			steps := workflow.Jobs["unit"].Steps
			Expect(steps[len(steps)-1].Run).Should(ContainSubstring("richgo test ./..."))
		})

		context("there are integration tests", func() {
			it.Before(func() {
				Expect(os.Mkdir(filepath.Join(dir, "integration"), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(dir, "integration", "main.go"), []byte{}, 0644)).To(Succeed())
			})

			it("will contribute a unit test and an integration test pipeline", func() {
				contribution, err := octo.ContributeTest(descriptor)
				Expect(err).To(Not(HaveOccurred()))

				Expect(contribution.Path).To(Equal(".github/workflows/pb-tests.yml"))

				var workflow jobs
				Expect(yaml.Unmarshal(contribution.Content, &workflow)).To(Succeed())

				Expect(len(workflow.Jobs)).To(Equal(3))
				Expect(workflow.Jobs["unit"]).To(Not(BeNil()))
				Expect(workflow.Jobs["integration"]).To(Not(BeNil()))

				unitSteps := workflow.Jobs["unit"].Steps
				Expect(unitSteps[len(unitSteps)-1].Run).Should(ContainSubstring("richgo test ./... -run Unit"))

				integrationSteps := workflow.Jobs["integration"].Steps
				Expect(integrationSteps[len(integrationSteps)-1].Run).Should(ContainSubstring("richgo test ./integration/... -run Integration"))
			})
		})

	})

}
