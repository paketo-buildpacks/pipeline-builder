#!/usr/bin/env bash

set -euo pipefail

go install -ldflags="-s -w" github.com/paketo-buildpacks/libpak/v2/cmd/update-build-image-dependency@${PAKETO_LIBPAK_COMMIT:=v2.0.0-alpha}
