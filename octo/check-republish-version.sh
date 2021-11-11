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

for GIT in $(git tag | sort -V -r ); do
  if contains "${GIT}" "${IMAGES}"; then
    echo "Found ${GIT}. Skipping."
    echo "::set-output name=skip::true"
    exit
  fi

  echo "::group::Checking out ${GIT}"
    git checkout -- .
    git checkout "${GIT}"
    echo "::endgroup::"
    echo "::set-output name=target::${GIT#v}"
  break
done

for IMG in $(crane ls "${SOURCE}" | grep -v "latest" | sort -V -r); do
  echo "::set-output name=source::${IMG}"
  exit
done
