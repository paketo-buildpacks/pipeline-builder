#!/usr/bin/env bash

set -euo pipefail

if [[ -n "${PUBLISH+x}" ]]; then
  pack package-buildpack \
    "${PACKAGE}:${VERSION}" \
    --config "${HOME}"/package.toml \
    --publish

  crane tag "${PACKAGE}:${VERSION}" latest
  echo "::set-output name=digest::$(crane digest "${PACKAGE}:${VERSION}")"
else
  pack package-buildpack \
    "${PACKAGE}:${VERSION}" \
    --config "${HOME}"/package.toml
fi
