#!/usr/bin/env bash

set -euo pipefail

GO111MODULE=on go get -u -ldflags="-s -w" github.com/paketo-buildpacks/libpak/cmd/update-package-dependency
