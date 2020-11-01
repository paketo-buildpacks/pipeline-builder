#!/usr/bin/env bash

set -euo pipefail

if [[ -f .github/pipeline-version ]]; then
  OLD_VERSION=$(cat .github/pipeline-version)
else
  OLD_VERSION="0.0.0"
fi

octo --descriptor "${DESCRIPTOR}"

PAYLOAD=$(gh api /repos/paketo-buildpacks/pipeline-builder/releases/latest)

NEW_VERSION=$(jq -n -r --argjson PAYLOAD "${PAYLOAD}" '$PAYLOAD.name')
echo "${NEW_VERSION}" > .github/pipeline-version

RELEASE_NOTES=$(
  gh api \
    -F text="$(jq -n -r --argjson PAYLOAD "${PAYLOAD}" '$PAYLOAD.body')" \
    -F mode="gfm" \
    -F context="paketo-buildpacks/pipeline-builder" \
    -X POST /markdown
)

git add .github/
git checkout -- .

echo "::set-output name=old-version::${OLD_VERSION}"
echo "::set-output name=new-version::${NEW_VERSION}"
echo "::set-output name=release-notes::${RELEASE_NOTES//$'\n'/%0A}"
