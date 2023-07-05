/*
 * Copyright 2018-2022 the original author or authors.
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

package drafts_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/pipeline-builder/actions"
	"github.com/paketo-buildpacks/pipeline-builder/drafts"
	"github.com/paketo-buildpacks/pipeline-builder/drafts/mocks"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/mock"
)

func testDrafts(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		dir string

		bpLoader mocks.BuildModuleLoader
	)

	context("draft a release", func() {
		it.Before(func() {
			var err error
			dir, err = ioutil.TempDir("", "drafts")
			Expect(err).To(Not(HaveOccurred()))
		})

		it.After(func() {
			Expect(os.RemoveAll(dir)).To(Succeed())
		})

		context("payloads", func() {
			it("creates a payload for a component buildpack", func() {
				d := drafts.Drafter{Loader: &bpLoader}

				p, err := d.CreatePayload(actions.Inputs{
					"release_id":       "foo-id",
					"release_name":     "foo-name",
					"release_body":     "foo-body",
					"release_tag_name": "foo-tag",
				}, "testdata/component")
				Expect(err).ToNot(HaveOccurred())

				Expect(p.Release).To(Equal(drafts.Release{
					ID:   "foo-id",
					Name: "foo-name",
					Body: "foo-body",
					Tag:  "foo-tag",
				}))
				Expect(p.NestedBuildpacks).To(HaveLen(0))
				Expect(p.Builder.OrderGroups).To(HaveLen(0))
				Expect(p.Builder.Stack.ID).To(BeEmpty())
				Expect(p.Builder.Buildpacks).To(HaveLen(0))
				Expect(p.PrimaryBuildpack.OrderGroups).To(HaveLen(0))
				Expect(p.PrimaryBuildpack.Info.ID).To(Equal("example/component"))
				Expect(p.PrimaryBuildpack.Info.Version).To(Equal("2.1.1"))
				Expect(p.PrimaryBuildpack.Info.Name).To(Equal("Example Component Buildpack"))
				Expect(p.PrimaryBuildpack.Stacks).To(Equal(
					[]libcnb.BuildpackStack{
						{ID: "*"},
						{ID: "stack1"},
						{ID: "stack2"},
					}))
				Expect(p.PrimaryBuildpack.Dependencies).To(HaveLen(2))
				Expect(p.PrimaryBuildpack.Dependencies[0].ID).To(Equal("dep"))
				Expect(p.PrimaryBuildpack.Dependencies[0].Version).To(Equal("8.5.78"))
				Expect(p.PrimaryBuildpack.Dependencies[0].Version).To(Equal("8.5.78"))
				Expect(p.PrimaryBuildpack.Dependencies[1].Version).To(Equal("9.0.62"))
				Expect(p.PrimaryBuildpack.Dependencies[0].Stacks).To(Equal([]string{"stack1", "stack2", "*"}))
				Expect(p.PrimaryBuildpack.Dependencies[1].CPEs).To(Equal([]string{"cpe:2.3:a:example:dep:9.0.62:*:*:*:*:*:*:*"}))
				Expect(p.PrimaryBuildpack.Dependencies[0].Licenses).To(HaveLen(1))
				Expect(p.PrimaryBuildpack.Dependencies[0].Licenses[0]).To(
					Equal(libpak.BuildModuleDependencyLicense{Type: "Apache-2.0", URI: "https://www.apache.org/licenses/"}))
			})

			it("creates a payload for a component buildpack without dependencies", func() {
				d := drafts.Drafter{Loader: &bpLoader}

				p, err := d.CreatePayload(actions.Inputs{
					"release_id":       "foo-id",
					"release_name":     "foo-name",
					"release_body":     "foo-body",
					"release_tag_name": "foo-tag",
				}, "testdata/component-no-deps")
				Expect(err).ToNot(HaveOccurred())

				Expect(p.Release).To(Equal(drafts.Release{
					ID:   "foo-id",
					Name: "foo-name",
					Body: "foo-body",
					Tag:  "foo-tag",
				}))
				Expect(p.NestedBuildpacks).To(HaveLen(0))
				Expect(p.Builder.OrderGroups).To(HaveLen(0))
				Expect(p.Builder.Stack.ID).To(BeEmpty())
				Expect(p.Builder.Buildpacks).To(HaveLen(0))
				Expect(p.PrimaryBuildpack.OrderGroups).To(HaveLen(0))
				Expect(p.PrimaryBuildpack.Info.ID).To(Equal("example/component"))
				Expect(p.PrimaryBuildpack.Info.Version).To(Equal("2.1.1"))
				Expect(p.PrimaryBuildpack.Info.Name).To(Equal("Example Component Buildpack"))
				Expect(p.PrimaryBuildpack.Stacks).To(Equal(
					[]libcnb.BuildpackStack{
						{ID: "*"},
						{ID: "stack1"},
						{ID: "stack2"},
					}))
				Expect(p.PrimaryBuildpack.Dependencies).To(HaveLen(0))
			})

			it("creates a payload for a composite buildpack", func() {
				bpLoader.On("LoadBuildpacks", mock.Anything).Return([]drafts.Buildpack{
					{Buildpack: libcnb.Buildpack{Info: libcnb.BuildpackInfo{ID: "bp1"}}},
					{Buildpack: libcnb.Buildpack{Info: libcnb.BuildpackInfo{ID: "bp2"}}},
					{Buildpack: libcnb.Buildpack{Info: libcnb.BuildpackInfo{ID: "bp3"}}},
					{Buildpack: libcnb.Buildpack{Info: libcnb.BuildpackInfo{ID: "bp4"}}},
				}, nil)

				d := drafts.Drafter{Loader: &bpLoader}

				p, err := d.CreatePayload(actions.Inputs{
					"release_id":       "foo-id",
					"release_name":     "foo-name",
					"release_body":     "foo-body",
					"release_tag_name": "foo-tag",
				}, "testdata/composite")
				Expect(err).ToNot(HaveOccurred())

				Expect(p.Release).To(Equal(drafts.Release{
					ID:   "foo-id",
					Name: "foo-name",
					Body: "foo-body",
					Tag:  "foo-tag",
				}))
				Expect(p.Builder.OrderGroups).To(HaveLen(0))
				Expect(p.Builder.Stack.ID).To(BeEmpty())
				Expect(p.Builder.Buildpacks).To(HaveLen(0))

				Expect(p.PrimaryBuildpack.Info.ID).To(Equal("example/composite"))
				Expect(p.PrimaryBuildpack.Info.Version).To(Equal("1.1.8"))
				Expect(p.PrimaryBuildpack.Info.Name).To(Equal("Example Composite Buildpack"))
				Expect(p.PrimaryBuildpack.Stacks).To(HaveLen(0))
				Expect(p.PrimaryBuildpack.OrderGroups).To(HaveLen(2))
				Expect(p.PrimaryBuildpack.OrderGroups[0].Groups).To(Equal([]libcnb.BuildpackOrderBuildpack{
					{ID: "example/bp1", Version: "3.1.0", Optional: true},
					{ID: "example/bp2", Version: "9.3.1", Optional: false},
					{ID: "example/bp3", Version: "1.10.0", Optional: true},
				}))
				Expect(p.PrimaryBuildpack.OrderGroups[1].Groups).To(Equal([]libcnb.BuildpackOrderBuildpack{
					{ID: "example/bp1", Version: "3.1.0", Optional: true},
					{ID: "example/bp3", Version: "1.10.0", Optional: true},
					{ID: "example/bp4", Version: "6.5.0", Optional: true},
				}))

				Expect(p.NestedBuildpacks).To(HaveLen(4))
				for i, expectedBp := range []string{"bp1", "bp2", "bp3", "bp4"} {
					Expect(p.NestedBuildpacks[i].Info.ID).To(Equal(expectedBp))
				}
			})

			it("creates a payload for a builder", func() {
				bpLoader.On("LoadBuildpacks", mock.Anything).Return([]drafts.Buildpack{
					{Buildpack: libcnb.Buildpack{Info: libcnb.BuildpackInfo{ID: "bp1"}}},
					{Buildpack: libcnb.Buildpack{Info: libcnb.BuildpackInfo{ID: "bp2"}}},
					{Buildpack: libcnb.Buildpack{Info: libcnb.BuildpackInfo{ID: "bp3"}}},
					{Buildpack: libcnb.Buildpack{Info: libcnb.BuildpackInfo{ID: "bp4"}}},
				}, nil)

				d := drafts.Drafter{Loader: &bpLoader}

				p, err := d.CreatePayload(actions.Inputs{
					"release_id":       "foo-id",
					"release_name":     "foo-name",
					"release_body":     "foo-body",
					"release_tag_name": "foo-tag",
				}, "testdata/builder")
				Expect(err).ToNot(HaveOccurred())

				Expect(p.Release).To(Equal(drafts.Release{
					ID:   "foo-id",
					Name: "foo-name",
					Body: "foo-body",
					Tag:  "foo-tag",
				}))

				Expect(p.PrimaryBuildpack.Info.ID).To(BeEmpty())
				Expect(p.PrimaryBuildpack.Info.Name).To(BeEmpty())
				Expect(p.PrimaryBuildpack.Info.Version).To(BeEmpty())
				Expect(p.PrimaryBuildpack.Stacks).To(HaveLen(0))
				Expect(p.PrimaryBuildpack.OrderGroups).To(HaveLen(0))

				Expect(p.Builder.Stack).To(Equal(drafts.BuilderStack{
					ID:         "stack1",
					BuildImage: "example.com/example/build:base-cnb",
					RunImage:   "example.com/example/run:base-cnb",
				}))

				Expect(p.Builder.Buildpacks).To(HaveLen(4))
				for i, expected := range []string{
					"docker://example.com/example/bp1:1.1.2",
					"docker://example.com/example/bp2:1.4.1",
					"docker://example.com/example/bp3:0.2.1",
					"docker://example.com/example/bp4:5.0.3",
				} {
					Expect(p.Builder.Buildpacks[i].URI).To(Equal(expected))
				}

				Expect(p.Builder.OrderGroups).To(HaveLen(1))
				Expect(p.Builder.OrderGroups[0].Groups).To(Equal([]libcnb.BuildpackOrderBuildpack{
					{ID: "example/bp1"},
					{ID: "example/bp2"},
					{ID: "example/bp3"},
					{ID: "example/bp4", Optional: true},
				}))

				Expect(p.NestedBuildpacks).To(HaveLen(4))
				for i, expectedBp := range []string{"bp1", "bp2", "bp3", "bp4"} {
					Expect(p.NestedBuildpacks[i].Info.ID).To(Equal(expectedBp))
				}
			})
		})

		context("generates body", func() {
			var templateContent string

			it.Before(func() {
				tc, err := ioutil.ReadFile("../actions/draft-release/draft.template")
				Expect(err).ToNot(HaveOccurred())
				templateContent = string(tc)
			})

			it("generates component", func() {
				d := drafts.Drafter{Loader: &bpLoader}

				p, err := d.CreatePayload(actions.Inputs{
					"release_id":       "foo-id",
					"release_name":     "foo-name",
					"release_body":     "foo-body",
					"release_tag_name": "foo-tag",
					"dry_run":          "true",
				}, "testdata/component")
				Expect(err).ToNot(HaveOccurred())

				buf := &bytes.Buffer{}
				err = d.BuildAndWriteReleaseDraftFromTemplate(buf, templateContent, p)
				Expect(err).ToNot(HaveOccurred())

				body := buf.String()
				Expect(body).To(ContainSubstring("**ID**: `example/component`"))
				Expect(body).To(ContainSubstring("**Digest**: <!-- DIGEST PLACEHOLDER -->"))
				Expect(body).To(ContainSubstring("#### Supported Stacks:"))
				Expect(body).To(ContainSubstring("- `stack1`"))
				Expect(body).To(ContainSubstring("- `stack2`"))
				Expect(body).To(ContainSubstring("- `*`"))
				Expect(body).To(ContainSubstring("#### Dependencies:"))
				Expect(body).To(ContainSubstring("Example Dep 8 | `8.5.78` | `84c7707db0ce495473df2efdc93da21b6d47bf25cd0a79de52e5472ff9e5f094`"))
				Expect(body).To(ContainSubstring("Example Dep 9 | `9.0.62` | `03157728a832cf9c83048cdc28d09600cbb3e4fa087f8b97d74c8b4f34cd89bb`"))
				Expect(body).To(ContainSubstring("foo-body"))
			})

			it("generates composite", func() {
				bpLoader.On("LoadBuildpacks", mock.Anything).Return([]drafts.Buildpack{
					{Buildpack: libcnb.Buildpack{Info: libcnb.BuildpackInfo{ID: "bp1", Name: "BP 1", Version: "3.1.0"}}},
					{Buildpack: libcnb.Buildpack{Info: libcnb.BuildpackInfo{ID: "bp2", Name: "BP 2", Version: "9.3.1"}}},
					{Buildpack: libcnb.Buildpack{Info: libcnb.BuildpackInfo{ID: "bp3", Name: "BP 3", Version: "1.10.0"}}},
					{Buildpack: libcnb.Buildpack{Info: libcnb.BuildpackInfo{ID: "bp4", Name: "BP 4", Version: "6.5.0"}}},
				}, nil)

				d := drafts.Drafter{Loader: &bpLoader}

				p, err := d.CreatePayload(actions.Inputs{
					"release_id":       "foo-id",
					"release_name":     "foo-name",
					"release_body":     "foo-body",
					"release_tag_name": "foo-tag",
					"dry_run":          "true",
				}, "testdata/composite")
				Expect(err).ToNot(HaveOccurred())

				buf := &bytes.Buffer{}
				err = d.BuildAndWriteReleaseDraftFromTemplate(buf, templateContent, p)
				Expect(err).ToNot(HaveOccurred())

				body := buf.String()
				Expect(body).To(ContainSubstring("**ID**: `example/composite`"))
				Expect(body).To(ContainSubstring("**Digest**: <!-- DIGEST PLACEHOLDER -->"))
				Expect(body).ToNot(ContainSubstring("#### Supported Stacks:"))
				Expect(body).To(ContainSubstring("#### Included Buildpackages:"))
				Expect(body).To(ContainSubstring("BP 1 | `bp1` | `3.1.0`"))
				Expect(body).To(ContainSubstring("BP 2 | `bp2` | `9.3.1`"))
				Expect(body).To(ContainSubstring("BP 3 | `bp3` | `1.10.0`"))
				Expect(body).To(ContainSubstring("BP 4 | `bp4` | `6.5.0`"))
				Expect(body).To(ContainSubstring("<summary>Order Groupings</summary>"))
				Expect(body).To(ContainSubstring("`example/bp1` | `3.1.0` | `true`"))
				Expect(body).To(ContainSubstring("`example/bp2` | `9.3.1` | `false`"))
				Expect(body).To(ContainSubstring("`example/bp3` | `1.10.0` | `true`"))
				Expect(body).To(ContainSubstring("<summary>BP 1 3.1.0</summary>"))
				Expect(body).To(ContainSubstring("**ID**: `bp1`"))
				Expect(body).To(ContainSubstring("<summary>BP 2 9.3.1</summary>"))
				Expect(body).To(ContainSubstring("**ID**: `bp2`"))
				Expect(body).To(ContainSubstring("<summary>BP 3 1.10.0</summary>"))
				Expect(body).To(ContainSubstring("**ID**: `bp3`"))
				Expect(body).To(ContainSubstring("<summary>BP 4 6.5.0</summary>"))
				Expect(body).To(ContainSubstring("**ID**: `bp4`"))
			})

			it("generates builder", func() {
				bpLoader.On("LoadBuildpacks", mock.Anything).Return([]drafts.Buildpack{
					{Buildpack: libcnb.Buildpack{Info: libcnb.BuildpackInfo{ID: "bp1", Name: "BP 1", Version: "3.1.0"}}},
					{Buildpack: libcnb.Buildpack{Info: libcnb.BuildpackInfo{ID: "bp2", Name: "BP 2", Version: "9.3.1"}}},
					{Buildpack: libcnb.Buildpack{Info: libcnb.BuildpackInfo{ID: "bp3", Name: "BP 3", Version: "1.10.0"}}},
					{Buildpack: libcnb.Buildpack{Info: libcnb.BuildpackInfo{ID: "bp4", Name: "BP 4", Version: "6.5.0"}}},
				}, nil)

				d := drafts.Drafter{Loader: &bpLoader}

				p, err := d.CreatePayload(actions.Inputs{
					"release_id":       "foo-id",
					"release_name":     "foo-name",
					"release_body":     "foo-body",
					"release_tag_name": "foo-tag",
					"dry_run":          "true",
				}, "testdata/builder")
				Expect(err).ToNot(HaveOccurred())

				buf := &bytes.Buffer{}
				err = d.BuildAndWriteReleaseDraftFromTemplate(buf, templateContent, p)
				Expect(err).ToNot(HaveOccurred())

				body := buf.String()
				Expect(body).ToNot(ContainSubstring("**ID**: ``"))
				Expect(body).To(ContainSubstring("**Digest**: <!-- DIGEST PLACEHOLDER -->"))
				Expect(body).ToNot(ContainSubstring("#### Supported Stacks:"))
				Expect(body).To(ContainSubstring("#### Included Buildpackages:"))
				Expect(body).To(ContainSubstring("BP 1 | `bp1` | `3.1.0`"))
				Expect(body).To(ContainSubstring("BP 2 | `bp2` | `9.3.1`"))
				Expect(body).To(ContainSubstring("BP 3 | `bp3` | `1.10.0`"))
				Expect(body).To(ContainSubstring("BP 4 | `bp4` | `6.5.0`"))
				Expect(body).To(ContainSubstring("<summary>Order Groupings</summary>"))
				Expect(body).To(ContainSubstring("`example/bp1` | `` | `false`"))
				Expect(body).To(ContainSubstring("`example/bp2` | `` | `false`"))
				Expect(body).To(ContainSubstring("`example/bp3` | `` | `false`"))
				Expect(body).To(ContainSubstring("`example/bp4` | `` | `true`"))
				Expect(body).To(ContainSubstring("<summary>BP 1 3.1.0</summary>"))
				Expect(body).To(ContainSubstring("**ID**: `bp1`"))
				Expect(body).To(ContainSubstring("<summary>BP 2 9.3.1</summary>"))
				Expect(body).To(ContainSubstring("**ID**: `bp2`"))
				Expect(body).To(ContainSubstring("<summary>BP 3 1.10.0</summary>"))
				Expect(body).To(ContainSubstring("**ID**: `bp3`"))
				Expect(body).To(ContainSubstring("<summary>BP 4 6.5.0</summary>"))
				Expect(body).To(ContainSubstring("**ID**: `bp4`"))
			})
		})
	})

}
