#!/usr/bin/env bash

set -euo pipefail

if [[ -n "${PUBLISH+x}" ]]; then
  pack buildpack package \
    "${PACKAGE}:${VERSION}" \
    --config "${HOME}"/package.toml \
    --publish

  if [ ! -z ${VERSION_MINOR} ] && [ ! -z ${VERSION_MAJOR} ]; then
    crane tag "${PACKAGE}:${VERSION}" "${VERSION_MINOR}"
    crane tag "${PACKAGE}:${VERSION}" "${VERSION_MAJOR}"
  fi
  crane tag "${PACKAGE}:${VERSION}" latest
  echo "::set-output name=digest::$(crane digest "${PACKAGE}:${VERSION}")"
else
  pack buildpack package \
    "${PACKAGE}:${VERSION}" \
    --config "${HOME}"/package.toml \
    --format "${FORMAT}"
fi
