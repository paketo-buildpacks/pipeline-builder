#!/usr/bin/env bash

set -euo pipefail

contains() {
  local TAG="${1}"
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

for IMG in $(crane ls "${SOURCE}" | grep -v "latest" | sort -r); do
  if contains "${IMG}" "${IMAGES}"; then
    echo "Found ${IMG}. Skipping."
    echo "::set-output name=skip::true"
    exit
  fi

  echo "::set-output name=version::${IMG}"
  exit
done
