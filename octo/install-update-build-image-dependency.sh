#!/usr/bin/env bash

set -euo pipefail

go install -ldflags="-s -w" github.com/paketo-buildpacks/libpak/cmd/update-build-image-dependency@cd6875914a49b022b2dd073cac568f1a780fccf3
