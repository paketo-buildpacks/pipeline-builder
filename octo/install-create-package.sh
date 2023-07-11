#!/usr/bin/env bash

set -euo pipefail

go install -ldflags="-s -w" github.com/paketo-buildpacks/libpak/cmd/create-package@${PAKETO_LIBPAK_COMMIT:=latest}
