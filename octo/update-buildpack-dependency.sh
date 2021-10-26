#!/usr/bin/env bash

set -euo pipefail

OLD_VERSION=$(yj -tj < buildpack.toml | jq -r "
  .metadata.dependencies[] |
  select( .id == env.ID ) |
  select( .version | test( env.VERSION_PATTERN ) ) |
  .version")

update-buildpack-dependency \
  --buildpack-toml buildpack.toml \
  --id "${ID}" \
  --version-pattern "${VERSION_PATTERN}" \
  --version "${VERSION}" \
  --uri "${URI}" \
  --sha256 "${SHA256}"

git add buildpack.toml
git checkout -- .

if [ "$(echo "$OLD_VERSION" | awk -F '.' '{print $1}')" != "$(echo "$VERSION" | awk -F '.' '{print $1}')" ]; then
  LABEL="semver:major"
elif [ "$(echo "$OLD_VERSION" | awk -F '.' '{print $2}')" != "$(echo "$VERSION" | awk -F '.' '{print $2}')" ]; then
  LABEL="semver:minor"
else
  LABEL="semver:patch"
fi

echo "::set-output name=old-version::${OLD_VERSION}"
echo "::set-output name=new-version::${VERSION}"
echo "::set-output name=version-label::${LABEL}"