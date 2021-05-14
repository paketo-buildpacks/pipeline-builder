#!/usr/bin/env bash

set -euo pipefail

if [[ -n "${PUBLISH+x}" ]]; then
  pack builder create \
    "${BUILDER}:${VERSION}" \
    --config builder.toml \
    --publish

    echo "::set-output name=digest::$(crane digest "${BUILDER}:${VERSION}")"
else
  pack builder create \
    "${BUILDER}:${VERSION}" \
    --config builder.toml
fi
