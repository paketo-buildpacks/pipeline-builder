#!/usr/bin/env bash

set -euo pipefail

PATTERN="refs/tags/v([0-9]+\.[0-9]+\.[0-9]+)"
if [[ ${GITHUB_REF} =~ ${PATTERN} ]]; then
  VERSION=${BASH_REMATCH[1]}
else
  VERSION=$(git rev-parse --short HEAD)
fi

echo "::set-output name=version::${VERSION}"
echo "Selected ${VERSION} from
  * ref: ${GITHUB_REF}
  * sha: ${GITHUB_SHA}
"
