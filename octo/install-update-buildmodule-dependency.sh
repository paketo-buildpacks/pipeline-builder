#!/usr/bin/env bash

set -euo pipefail

go install -ldflags="-s -w" github.com/paketo-buildpacks/libpak/v2/cmd/update-buildmodule-dependency@${PAKETO_LIBPAK_COMMIT:=latest}
