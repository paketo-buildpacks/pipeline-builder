#!/usr/bin/env bash

set -euo pipefail

if [[ "${PUBLISH:-x}" == "true" ]]; then
  pack buildpack package \
    "${PACKAGE}:${VERSION}" \
    --config "${HOME}"/package.toml \
    --publish

  if [[ -n ${VERSION_MINOR:-} && -n ${VERSION_MAJOR:-} ]]; then
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
