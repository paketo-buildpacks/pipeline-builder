#!/usr/bin/env bash

set -euo pipefail

go install -ldflags="-s -w" github.com/paketo-buildpacks/pipeline-builder/v2/cmd/octo@${PAKETO_PIPELINEBUILDER_COMMIT:=latest}
