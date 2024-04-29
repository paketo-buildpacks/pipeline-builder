#!/usr/bin/env bash

set -euo pipefail

COMPILED_BUILDPACK="${HOME}/buildpack"

# create-package puts the buildpack here, we need to run from that directory
#   for component buildpacks so that pack doesn't need a package.toml
cd "${COMPILED_BUILDPACK}"
CONFIG=""
if [ -f "${COMPILED_BUILDPACK}/package.toml" ]; then
  CONFIG="--config ${COMPILED_BUILDPACK}/package.toml"
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
    --format "${FORMAT}" $([ -n "$TTL_SH_PUBLISH" ] && [ "$TTL_SH_PUBLISH" = "true" ] && echo "--publish")
fi