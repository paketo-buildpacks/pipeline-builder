#!/usr/bin/env bash

set -euo pipefail

go install -ldflags="-s -w" github.com/paketo-buildpacks/libpak/cmd/update-lifecycle-dependency@${PAKETO_LIBPAK_COMMIT:=latest}
