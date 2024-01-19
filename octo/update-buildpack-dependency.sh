#!/usr/bin/env bash

set -euo pipefail

OLD_VERSION=$(yj -tj < buildpack.toml | \
  jq -r ".metadata.dependencies[] | select( .id == env.ID ) | select( .version | test( env.VERSION_PATTERN ) ) | select( .purl | contains( env.ARCH ) ) | .version")

update-buildpack-dependency \
  --buildpack-toml buildpack.toml \
  --id "${ID}" \
  --version-pattern "${VERSION_PATTERN}" \
  --version "${VERSION}" \
  --cpe-pattern "${CPE_PATTERN:-}" \
  --cpe "${CPE:-}" \
  --purl-pattern "${PURL_PATTERN:-}" \
  --purl "${PURL:-}" \
  --uri "${URI}" \
  --sha256 "${SHA256}" \
  --source "${SOURCE_URI}" \
  --source-sha256 "${SOURCE_SHA256}"

git add buildpack.toml
git checkout -- .

if [ "$(echo "$OLD_VERSION" | awk -F '.' '{print $1}')" != "$(echo "$VERSION" | awk -F '.' '{print $1}')" ]; then
  LABEL="semver:major"
elif [ "$(echo "$OLD_VERSION" | awk -F '.' '{print $2}')" != "$(echo "$VERSION" | awk -F '.' '{print $2}')" ]; then
  LABEL="semver:minor"
else
  LABEL="semver:patch"
fi

echo "old-version=${OLD_VERSION}" >> "$GITHUB_OUTPUT"
echo "new-version=${VERSION}" >> "$GITHUB_OUTPUT"
echo "version-label=${LABEL}" >> "$GITHUB_OUTPUT"
