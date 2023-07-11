#!/usr/bin/env bash

set -euo pipefail

go install -ldflags="-s -w" github.com/paketo-buildpacks/libpak/cmd/update-build-image-dependency@${PAKETO_LIBPAK_COMMIT:=latest}
