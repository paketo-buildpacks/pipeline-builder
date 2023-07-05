#!/usr/bin/env bash

set -euo pipefail

PACKAGE_LIST=($PACKAGES)
# Extract first repo (Docker Hub) as the main to package & register
PACKAGE=${PACKAGE_LIST[0]}

if [[ "${EXTENSION:-x}" == "true" ]]; then
  export PACK_MODULE_TYPE=extension
else
  export PACK_MODULE_TYPE=buildpack
fi

if [[ "${PUBLISH:-x}" == "true" ]]; then
  pack ${PACK_MODULE_TYPE} package \
    "${PACKAGE}:${VERSION}" \
    --config "${HOME}"/package.toml \
    --publish

  if [[ -n ${VERSION_MINOR:-} && -n ${VERSION_MAJOR:-} ]]; then
    crane tag "${PACKAGE}:${VERSION}" "${VERSION_MINOR}"
    crane tag "${PACKAGE}:${VERSION}" "${VERSION_MAJOR}"
  fi
  crane tag "${PACKAGE}:${VERSION}" latest
  echo "digest=$(crane digest "${PACKAGE}:${VERSION}")" >> "$GITHUB_OUTPUT"

  # copy to other repositories specified
  for P in "${PACKAGE_LIST[@]}"
    do
      if [ "$P" != "$PACKAGE" ]; then
        crane copy "${PACKAGE}:${VERSION}" "${P}:${VERSION}"
        if [[ -n ${VERSION_MINOR:-} && -n ${VERSION_MAJOR:-} ]]; then
           crane tag "${P}:${VERSION}" "${VERSION_MINOR}"
           crane tag "${P}:${VERSION}" "${VERSION_MAJOR}"
        fi
        crane tag "${P}:${VERSION}" latest
      fi
    done

else
  pack ${PACK_MODULE_TYPE} package \
    "${PACKAGE}:${VERSION}" \
    --config "${HOME}"/package.toml \
    --format "${FORMAT}"
fi