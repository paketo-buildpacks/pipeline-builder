#!/usr/bin/env bash

set -euo pipefail

echo "Installing yj ${YJ_VERSION}"

mkdir -p "${HOME}"/bin
echo "${HOME}/bin" >> "${GITHUB_PATH}"

curl \
  --location \
  --show-error \
  --silent \
  --output "${HOME}"/bin/yj \
  "https://github.com/sclevine/yj/releases/download/v${YJ_VERSION}/yj-linux-amd64"

chmod +x "${HOME}"/bin/yj
