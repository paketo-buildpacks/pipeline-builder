#!/usr/bin/env bash

set -euo pipefail

if [[ -n "${PUBLISH+x}" ]]; then
  pack buildpack package \
    "${PACKAGE}:${VERSION}" \
    --config "${HOME}"/package.toml \
    --publish

  crane tag "${PACKAGE}:${VERSION}" latest
  echo "::set-output name=digest::$(crane digest "${PACKAGE}:${VERSION}")"
else
  pack buildpack package \
    "${PACKAGE}:${VERSION}" \
    --config "${HOME}"/package.toml \
    --format "${FORMAT}"
fi
