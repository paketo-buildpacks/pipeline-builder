#!/usr/bin/env bash

set -euo pipefail

echo "::group::Building ${TARGET}:${VERSION}"
  docker build \
    --file actions/Dockerfile \
    --build-arg "SOURCE=${SOURCE}" \
    --tag "${TARGET}:${VERSION}" \
    .
echo "::endgroup::"

if [[ "${PUSH}" == "true" ]]; then
  echo "::group::Pushing ${TARGET}:${VERSION}"
    docker push "${TARGET}:${VERSION}"
  echo "::endgroup::"
else
  echo "Skipping push"
fi
