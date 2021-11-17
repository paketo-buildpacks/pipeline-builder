#!/usr/bin/env bash

set -euo pipefail

if [[ ${GITHUB_REF} =~ refs/tags/v([0-9]+\.[0-9]+\.[0-9]+) ]]; then
  VERSION=${BASH_REMATCH[1]}
elif [[ ${GITHUB_REF} =~ refs/heads/(.+) ]]; then
  VERSION=${BASH_REMATCH[1]}
else
  VERSION=$(git rev-parse --short HEAD)
fi

MAJOR_VERSION="$(echo "${VERSION}" | awk -F '.' '{print $1 }')"
MINOR_VERSION="$(echo "${VERSION}" | awk -F '.' '{print $1 "." $2 }')"

echo "::set-output name=version_major::${MAJOR_VERSION}"
echo "::set-output name=version_minor::${MINOR_VERSION}"
echo "::set-output name=version::${VERSION}"
echo "Selected ${VERSION} from
  * ref: ${GITHUB_REF}
  * sha: ${GITHUB_SHA}
"
