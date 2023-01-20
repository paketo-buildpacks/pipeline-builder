#!/usr/bin/env bash

set -euo pipefail

if [[ -f .github/pipeline-version ]]; then
  OLD_VERSION=$(cat .github/pipeline-version)
else
  OLD_VERSION="0.0.0"
fi

rm .github/workflows/pb-*.yml || true
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

echo "old-version=${OLD_VERSION}" >> "$GITHUB_OUTPUT"
echo "new-version=${NEW_VERSION}" >> "$GITHUB_OUTPUT"

DELIMITER=$(openssl rand -hex 16) # roughly the same entropy as uuid v4 used in https://github.com/actions/toolkit/blob/b36e70495fbee083eb20f600eafa9091d832577d/packages/core/src/file-command.ts#L28
printf "release-notes<<%s\n%s\n%s\n" "${DELIMITER}" "${RELEASE_NOTES}" "${DELIMITER}" >> "${GITHUB_OUTPUT}" # see https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#multiline-strings
