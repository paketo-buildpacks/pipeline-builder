#!/usr/bin/env bash

set -euo pipefail

echo "Installing pack ${PACK_VERSION}"

mkdir -p "${HOME}"/bin
echo "${HOME}/bin" >> "${GITHUB_PATH}"

curl \
  --location \
  --show-error \
  --silent \
  "https://github.com/buildpacks/pack/releases/download/v${PACK_VERSION}/pack-v${PACK_VERSION}-linux.tgz" \
| tar -C "${HOME}"/bin -xz pack
