/*
 * Copyright 2018-2025 the original author or authors.
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

package integration_test

import (
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-github/v43/github"
	"github.com/paketo-buildpacks/pipeline-builder/drafts"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDrafts(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	context("registry buildpack loader", func() {
		it("fetches buildpack.toml from a remote buildpack", func() {
			bp, err := drafts.RegistryBuildpackLoader{}.LoadBuildpack("docker.io/paketobuildpacks/bellsoft-liberica:latest")
			Expect(err).ToNot(HaveOccurred())

			Expect(bp.Info.ID).To(Equal("paketo-buildpacks/bellsoft-liberica"))
			Expect(bp.Dependencies).ToNot(BeEmpty())
			Expect(bp.OrderGroups).To(BeEmpty())
			Expect(bp.Stacks).ToNot(BeEmpty())
		})

		it("fails fetching an image that does not exist", func() {
			_, err := drafts.RegistryBuildpackLoader{}.LoadBuildpack("lasjdflaksdjfl")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(HavePrefix("unable to load lasjdflaksdjfl")))
		})
	})

	context("github buildpack loader", func() {
		it("fetches buildpack.toml from a remote buildpack", func() {
			bp, err := drafts.GithubBuildpackLoader{
				GithubClient: github.NewClient(http.DefaultClient),
				RegexMappers: []string{
					`|paketobuildpacks|paketo-buildpacks|`,
				},
			}.LoadBuildpack("docker.io/paketobuildpacks/bellsoft-liberica:main")
			Expect(err).ToNot(HaveOccurred())

			Expect(bp.Info.ID).To(Equal("paketo-buildpacks/bellsoft-liberica"))
			Expect(bp.Info.Version).ToNot(ContainSubstring("{{.version}}"))
			Expect(bp.Dependencies).ToNot(BeEmpty())
			Expect(bp.OrderGroups).To(BeEmpty())
			Expect(bp.Stacks).ToNot(BeEmpty())
		})

		it("fetches buildpack.toml from a mapped remote buildpack", func() {
			bp, err := drafts.GithubBuildpackLoader{
				GithubClient: github.NewClient(http.DefaultClient),
				RegexMappers: []string{
					`|foo|bar|`,
					`|foo\/me|paketo-buildpacks/bellsoft-liberica|`,
					`|foo|baz|`,
				},
			}.LoadBuildpack("docker.io/foo/me:main")
			Expect(err).ToNot(HaveOccurred())

			Expect(bp.Info.ID).To(Equal("paketo-buildpacks/bellsoft-liberica"))
			Expect(bp.Info.Version).ToNot(ContainSubstring("{{.version}}"))
			Expect(bp.Dependencies).ToNot(BeEmpty())
			Expect(bp.OrderGroups).To(BeEmpty())
			Expect(bp.Stacks).ToNot(BeEmpty())
		})

		it("fetches buildpack.toml from a multiple remote buildpack", func() {
			bpList := []string{
				"docker.io/paketobuildpacks/upx:main",
				"docker.io/paketobuildpacks/azul-zulu:main",
				"docker.io/paketobuildpacks/watchexec:main",
				"docker.io/paketobuildpacks/bellsoft-liberica:main",
			}

			bps, err := drafts.GithubBuildpackLoader{
				GithubClient: github.NewClient(http.DefaultClient),
				RegexMappers: []string{
					`|paketobuildpacks|paketo-buildpacks|`,
				},
			}.LoadBuildpacks(bpList)
			Expect(err).ToNot(HaveOccurred())

			bpIds := []string{
				"paketo-buildpacks/azul-zulu",
				"paketo-buildpacks/bellsoft-liberica",
				"paketo-buildpacks/upx",
				"paketo-buildpacks/watchexec",
			}

			sort.Strings(bpList)
			for i := range bpList {
				Expect(bps[i].Info.ID).To(Equal(strings.TrimSuffix(bpIds[i], ":main")))
				Expect(bps[i].Info.Version).ToNot(ContainSubstring("{{.version}}"))
				Expect(bps[i].Dependencies).ToNot(BeEmpty())
				Expect(bps[i].OrderGroups).To(BeEmpty())
				Expect(bps[i].Stacks).ToNot(BeEmpty())
			}
		})

		it("fails fetching an image that does not exist", func() {
			_, err := drafts.GithubBuildpackLoader{
				GithubClient: github.NewClient(http.DefaultClient),
			}.LoadBuildpack("lasjdflaksdjfl")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ContainSubstring("unable to parse lasjdflaksdjfl, found []")))

			_, err = drafts.GithubBuildpackLoader{
				GithubClient: github.NewClient(http.DefaultClient),
			}.LoadBuildpack("docker.io/paketobuildpacks/does-not-exist:main")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ContainSubstring("unable to load buildpack.toml for docker.io/paketobuildpacks/does-not-exist:main")))
		})
	})
}
