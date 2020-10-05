#!/usr/bin/env bash

set -euo pipefail

if [[ ${GITHUB_REF} =~ refs/tags/v([0-9]+\.[0-9]+\.[0-9]+) ]]; then
  VERSION=${BASH_REMATCH[1]}
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
