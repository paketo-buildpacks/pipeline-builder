package integration_test

import (
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-github/v43/github"
	drafts "github.com/paketo-buildpacks/pipeline-builder/v2/drafts"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDrafts(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	// context("registry buildpack loader", func() {
	// 	it("fetches buildpack.toml from a remote buildpack", func() {
	// 		bp, err := drafts.RegistryBuildpackLoader{}.LoadBuildpack("gcr.io/paketo-buildpacks/bellsoft-liberica:latest")
	// 		Expect(err).ToNot(HaveOccurred())

	// 		Expect(bp.Info.ID).To(Equal("paketo-buildpacks/bellsoft-liberica"))
	// 		Expect(bp.Dependencies).ToNot(BeEmpty())
	// 		Expect(bp.OrderGroups).To(BeEmpty())
	// 		Expect(bp.Stacks).ToNot(BeEmpty())
	// 	})

	// 	it("fails fetching an image that does not exist", func() {
	// 		_, err := drafts.RegistryBuildpackLoader{}.LoadBuildpack("lasjdflaksdjfl")
	// 		Expect(err).To(HaveOccurred())
	// 		Expect(err).To(MatchError(HavePrefix("unable to load lasjdflaksdjfl")))
	// 	})
	// })

	context("github buildpack loader", func() {
		it("fetches buildpack.toml from a remote buildpack", func() {
			bp, err := drafts.GithubBuildModuleLoader{
				GithubClient: github.NewClient(http.DefaultClient),
			}.LoadBuildpack("gcr.io/paketo-buildpacks/bellsoft-liberica:main")
			Expect(err).ToNot(HaveOccurred())

			Expect(bp.Info.ID).To(Equal("paketo-buildpacks/bellsoft-liberica"))
			Expect(bp.Info.Version).ToNot(ContainSubstring("{{.version}}"))
			Expect(bp.Dependencies).ToNot(BeEmpty())
			Expect(bp.OrderGroups).To(BeEmpty())
			Expect(bp.Stacks).ToNot(BeEmpty())
		})

		it("fetches buildpack.toml from a mapped remote buildpack", func() {
			bp, err := drafts.GithubBuildModuleLoader{
				GithubClient: github.NewClient(http.DefaultClient),
				RegexMappers: []string{
					`|foo|bar|`,
					`|foo\/me|paketo-buildpacks/bellsoft-liberica|`,
					`|foo|baz|`,
				},
			}.LoadBuildpack("gcr.io/foo/me:main")
			Expect(err).ToNot(HaveOccurred())

			Expect(bp.Info.ID).To(Equal("paketo-buildpacks/bellsoft-liberica"))
			Expect(bp.Info.Version).ToNot(ContainSubstring("{{.version}}"))
			Expect(bp.Dependencies).ToNot(BeEmpty())
			Expect(bp.OrderGroups).To(BeEmpty())
			Expect(bp.Stacks).ToNot(BeEmpty())
		})

		it("fetches buildpack.toml from a multiple remote buildpack", func() {
			bpList := []string{
				"gcr.io/paketo-buildpacks/upx:main",
				"gcr.io/paketo-buildpacks/azul-zulu:main",
				"gcr.io/paketo-buildpacks/watchexec:main",
				"gcr.io/paketo-buildpacks/bellsoft-liberica:main",
			}

			bps, err := drafts.GithubBuildModuleLoader{
				GithubClient: github.NewClient(http.DefaultClient),
			}.LoadBuildpacks(bpList)
			Expect(err).ToNot(HaveOccurred())

			sort.Strings(bpList)
			for i := range bpList {
				Expect(bps[i].Info.ID).To(Equal(strings.TrimSuffix(strings.TrimPrefix(bpList[i], "gcr.io/"), ":main")))
				Expect(bps[i].Info.Version).ToNot(ContainSubstring("{{.version}}"))
				Expect(bps[i].Dependencies).ToNot(BeEmpty())
				Expect(bps[i].OrderGroups).To(BeEmpty())
				Expect(bps[i].Stacks).ToNot(BeEmpty())
			}
		})

		it("fails fetching an image that does not exist", func() {
			_, err := drafts.GithubBuildModuleLoader{
				GithubClient: github.NewClient(http.DefaultClient),
			}.LoadBuildpack("lasjdflaksdjfl")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ContainSubstring("unable to parse lasjdflaksdjfl, found []")))

			_, err = drafts.GithubBuildModuleLoader{
				GithubClient: github.NewClient(http.DefaultClient),
			}.LoadBuildpack("gcr.io/paketo-buildpacks/does-not-exist:main")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ContainSubstring("unable to load buildpack.toml for gcr.io/paketo-buildpacks/does-not-exist:main")))
		})
	})
}
