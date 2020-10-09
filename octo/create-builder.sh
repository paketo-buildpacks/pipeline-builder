#!/usr/bin/env bash

set -euo pipefail

if [[ -n "${PUBLISH+x}" ]]; then
  pack create-builder \
    "${BUILDER}:${VERSION}" \
    --config builder.toml \
    --publish
else
  pack create-builder \
    "${BUILDER}:${VERSION}" \
    --config builder.toml
fi
