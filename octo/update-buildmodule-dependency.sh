#!/usr/bin/env bash

set -euo pipefail

if [[ "${EXTENSION:-x}" == "true" ]]; then
  export PACK_MODULE_TYPE=extension
else
  export PACK_MODULE_TYPE=buildpack
fi

OLD_VERSION=$(yj -tj < ${PACK_MODULE_TYPE}.toml | jq -r "
  .metadata.dependencies[] |
  select( .id == env.ID ) |
  select( .version | test( env.VERSION_PATTERN ) ) |
  .version")

update-buildmodule-dependency \
  --buildmodule-toml ${PACK_MODULE_TYPE}.toml \
  --id "${ID}" \
  --version-pattern "${VERSION_PATTERN}" \
  --version "${VERSION}" \
  --cpe-pattern "${CPE_PATTERN:-}" \
  --cpe "${CPE:-}" \
  --purl-pattern "${PURL_PATTERN:-}" \
  --purl "${PURL:-}" \
  --uri "${URI}" \
  --sha256 "${SHA256}"

git add ${PACK_MODULE_TYPE}.toml
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
