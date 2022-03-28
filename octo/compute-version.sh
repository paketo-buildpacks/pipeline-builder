#!/usr/bin/env bash

set -euo pipefail

if [ -z "${GITHUB_REF+set}" ]; then
  echo "GITHUB_REF set to [${GITHUB_REF-<unset>}], but should never be empty or unset"
  exit 255
fi

if [[ ${GITHUB_REF} =~ refs/tags/v([0-9]+\.[0-9]+\.[0-9]+) ]]; then
  VERSION=${BASH_REMATCH[1]}

  MAJOR_VERSION="$(echo "${VERSION}" | awk -F '.' '{print $1 }')"
  MINOR_VERSION="$(echo "${VERSION}" | awk -F '.' '{print $1 "." $2 }')"

  echo "::set-output name=version-major::${MAJOR_VERSION}"
  echo "::set-output name=version-minor::${MINOR_VERSION}"
elif [[ ${GITHUB_REF} =~ refs/heads/(.+) ]]; then
  VERSION=${BASH_REMATCH[1]}
else
  VERSION=$(git rev-parse --short HEAD)
fi

echo "::set-output name=version::${VERSION}"
echo "Selected ${VERSION} from
  * ref: ${GITHUB_REF}
  * sha: ${GITHUB_SHA}
"
