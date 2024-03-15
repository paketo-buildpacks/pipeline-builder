#!/usr/bin/env bash

set -euo pipefail

CONFIG="--config "${HOME}"/package.toml"
#TODO with this, we don't need to use the package.toml, because pack exp. does not support it with multi arch yet
if ! [ -f "${PWD}/package.toml" ]; then
  cd ~/buildpack
  CONFIG=""
fi

PACKAGE_LIST=($PACKAGES)
# Extract first repo (Docker Hub) as the main to package & register
PACKAGE=${PACKAGE_LIST[0]}

if [[ "${PUBLISH:-x}" == "true" ]]; then
  pack -v buildpack package \
    "${PACKAGE}:${VERSION}" ${CONFIG} \
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
  pack -v buildpack package \
    "${PACKAGE}:${VERSION}" ${CONFIG} \
    --format "${FORMAT}"
fi