#!/usr/bin/env bash

set -euo pipefail

echo "Installing richgo ${RICHGO_VERSION}"

mkdir -p "${HOME}"/bin
echo "${HOME}/bin" >> "${GITHUB_PATH}"

curl \
  --location \
  --show-error \
  --silent \
  "https://github.com/kyoh86/richgo/releases/download/v${RICHGO_VERSION}/richgo_${RICHGO_VERSION}_linux_amd64.tar.gz" \
| tar -C "${HOME}"/bin -xz richgo
