#!/usr/bin/env bash

set -euo pipefail

go install -ldflags="-s -w" github.com/paketo-buildpacks/pipeline-builder/v2/cmd/update-buildpack-image-id@${PAKETO_PIPELINEBUILDER_COMMIT:=v2.0.0-alpha.1}
