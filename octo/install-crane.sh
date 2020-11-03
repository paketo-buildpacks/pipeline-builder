#!/usr/bin/env bash

set -euo pipefail

echo "Installing crane ${CRANE_VERSION}"

mkdir -p "${HOME}"/bin
echo "${HOME}/bin" >> "${GITHUB_PATH}"

curl \
  --show-error \
  --silent \
  --location \
  "https://github.com/google/go-containerregistry/releases/download/v${CRANE_VERSION}/go-containerregistry_Linux_x86_64.tar.gz" \
| tar -C "${HOME}/bin" -xz crane
