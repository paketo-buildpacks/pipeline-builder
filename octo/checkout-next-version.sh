#!/usr/bin/env bash

set -euo pipefail

contains() {
  local TAG="${1}"
  TAG=${TAG#v}
  local SUFFIX="${2}"
  local IMAGES="${3}"

  for IMAGE in ${IMAGES}; do
    if [[ "${TAG}${SUFFIX}" == "${IMAGE}" ]]; then
      echo "::debug::Found ${TAG}${SUFFIX}"
      return 0
    fi
  done

  return 1
}

SUFFIX=${SUFFIX:-}

IMAGES=$(crane ls "${TARGET}")

for GIT in $(git tag | sort -V -r ); do
  if contains "${GIT}" "${SUFFIX}" "${IMAGES}"; then
    echo "Found ${GIT}. Skipping."
    echo "::set-output name=skip::true"
    exit
  fi

  echo "::group::Checking out ${GIT}"
    git checkout -- .
    git checkout "${GIT}"
  echo "::endgroup::"
  echo "::set-output name=version::${GIT#v}"
  exit
done
