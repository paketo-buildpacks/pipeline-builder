#!/usr/bin/env bash

set -euo pipefail

mkdir -p "${HOME}"/bin
echo "::add-path::${HOME}/bin"

curl \
	--location \
	--show-error \
	--silent \
	--output "${HOME}"/bin/yj \
	"https://github.com/sclevine/yj/releases/download/v${YJ_VERSION}/yj-linux"

chmod +x "${HOME}"/bin/yj
