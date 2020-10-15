#!/usr/bin/env bash

set -euo pipefail

if [[ -n "${PUBLISH+x}" ]]; then
  pack create-builder \
    "${BUILDER}:${VERSION}" \
    --config builder.toml \
    --publish

    echo "::set-output name=digest::$(crane digest "${BUILDER}:${VERSION}")"
else
  pack create-builder \
    "${BUILDER}:${VERSION}" \
    --config builder.toml
fi
