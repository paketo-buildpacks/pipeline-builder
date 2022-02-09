#!/usr/bin/env bash

set -euo pipefail

contains() {
  local TAG="${1}"
  TAG=${TAG#v}
  local IMAGES="${2}"

  for IMAGE in ${IMAGES}; do
    if [[ "${TAG}" == "${IMAGE}" ]]; then
      echo "::debug::Found ${TAG}"
      return 0
    fi
  done

  return 1
}

IMAGES=$(crane ls "${TARGET}")

for TAG in $(git tag -l "${TAG_PREFIX}*" | sort -V -r ); do
  VERSION=${TAG#${TAG_PREFIX}}
  VERSION=${VERSION#v}

  if contains "${VERSION}" "${IMAGES}"; then
    echo "Found ${TAG}. Skipping."
    echo "::set-output name=skip::true"
    exit
  fi

  echo "::group::Checking out ${TAG}"
    git checkout -- .
    git checkout "${TAG}"
  echo "::endgroup::"
  echo "::set-output name=version::${VERSION}"
  exit
done
