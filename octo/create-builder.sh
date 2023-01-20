#!/usr/bin/env bash

set -euo pipefail

if [[ -n "${PUBLISH+x}" ]]; then
  pack builder create \
    "${BUILDER}:${VERSION}" \
    --config builder.toml \
    --publish

    echo "digest=$(crane digest "${BUILDER}:${VERSION}")" >> "$GITHUB_OUTPUT"
else
  pack builder create \
    "${BUILDER}:${VERSION}" \
    --config builder.toml
fi
