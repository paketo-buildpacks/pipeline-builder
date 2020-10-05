#!/usr/bin/env bash

set -euo pipefail

PAYLOAD=$(cat "${GITHUB_EVENT_PATH}")

RELEASE_ID=$(jq -n -r --argjson PAYLOAD "${PAYLOAD}" '$PAYLOAD.release.id')
RELEASE_TAG_NAME=$(jq -n -r --argjson PAYLOAD "${PAYLOAD}" '$PAYLOAD.release.tag_name')
RELEASE_NAME=$(jq -n -r --argjson PAYLOAD "${PAYLOAD}" '$PAYLOAD.release.name')
RELEASE_BODY=$(jq -n -r --argjson PAYLOAD "${PAYLOAD}" '$PAYLOAD.release.body')

gh api \
  --method PATCH \
  "/repos/:owner/:repo/releases/${RELEASE_ID}" \
  --field "tag_name=${RELEASE_TAG_NAME}" \
  --field "name=${RELEASE_NAME}" \
  --field "body=${RELEASE_BODY//<!-- DIGEST PLACEHOLDER -->/\`${DIGEST}\`}"
