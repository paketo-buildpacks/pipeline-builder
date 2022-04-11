package integration_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip integration tests in short mode")
	}

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Default", testDrafts)
	suite.Run(t)
}
