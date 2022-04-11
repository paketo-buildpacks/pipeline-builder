package integration_test

import (
	"testing"

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
			bp, err := drafts.RegistryBuildpackLoader{}.LoadBuildpack("gcr.io/paketo-buildpacks/bellsoft-liberica:latest")
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
}
