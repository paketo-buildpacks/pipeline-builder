#!/usr/bin/env bash

set -euo pipefail

OLD_VERSION=$(yj -tj < builder.toml | jq -r ".stack.\"build-image\" | capture(\"${IMAGE}:(?<version>.+-${CLASSIFIER})\") | .version")
NEW_VERSION=$(crane ls "${IMAGE}" | grep ".*-${CLASSIFIER}" | sort -V | tail -n 1)

update-build-image-dependency \
  --builder-toml builder.toml \
  --version "${NEW_VERSION}"

git add builder.toml
git checkout -- .

if [ "$(echo "$OLD_VERSION" | awk -F '.' '{print $1}')" != "$(echo "$NEW_VERSION" | awk -F '.' '{print $1}')" ]; then
  LABEL="semver:major"
elif [ "$(echo "$OLD_VERSION" | awk -F '.' '{print $2}')" != "$(echo "$NEW_VERSION" | awk -F '.' '{print $2}')" ]; then
  LABEL="semver:minor"
else
  LABEL="semver:patch"
fi

echo "old-version=${OLD_VERSION}" >> "$GITHUB_OUTPUT"
echo "new-version=${NEW_VERSION}" >> "$GITHUB_OUTPUT"
echo "version-label=${LABEL}" >> "$GITHUB_OUTPUT"
