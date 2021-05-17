#!/usr/bin/env bash

set -euo pipefail

if [[ -n "${PUBLISH+x}" ]]; then
  pack buildpack package \
    "${PACKAGE}:${VERSION}${SUFFIX}" \
    --config "${HOME}"/package.toml \
    --publish

  crane tag "${PACKAGE}:${VERSION}${SUFFIX}" "latest${SUFFIX}"
  echo "::set-output name=digest::$(crane digest "${PACKAGE}:${VERSION}${SUFFIX}")"
else
  pack buildpack package \
    "${PACKAGE}:${VERSION}${SUFFIX}" \
    --config "${HOME}"/package.toml \
    --format "${FORMAT}"
fi
