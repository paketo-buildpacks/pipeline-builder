#!/usr/bin/env bash

set -euo pipefail

NEW_VERSION=$(crane ls "${DEPENDENCY}" | grep -v latest | sort -V | tail -n 1)

if [[ -e builder.toml ]]; then
  OLD_VERSION=$(yj -tj < builder.toml | jq -r ".buildpacks[].uri | capture(\".*${DEPENDENCY}:(?<version>.+)\") | .version")

  update-package-dependency \
    --builder-toml builder.toml \
    --id "${DEPENDENCY}" \
    --version "${NEW_VERSION}"

  git add builder.toml
fi

if [[ -e package.toml ]]; then
  OLD_VERSION=$(yj -tj < package.toml | jq -r ".dependencies[].uri | capture(\".*${DEPENDENCY}:(?<version>.+)\") | .version")

  update-package-dependency \
    --buildpack-toml buildpack.toml \
    --id "${BP_DEPENDENCY:-$DEPENDENCY}" \
    --version "${NEW_VERSION}"

  update-package-dependency \
    --package-toml package.toml \
    --id "${PKG_DEPENDENCY:-$DEPENDENCY}" \
    --version "${NEW_VERSION}"

  git add buildpack.toml package.toml
fi

git checkout -- .

if [ "$(echo "$OLD_VERSION" | awk -F '.' '{print $1}')" != "$(echo "$NEW_VERSION" | awk -F '.' '{print $1}')" ]; then
  LABEL="semver:major"
elif [ "$(echo "$OLD_VERSION" | awk -F '.' '{print $2}')" != "$(echo "$NEW_VERSION" | awk -F '.' '{print $2}')" ]; then
  LABEL="semver:minor"
else
  LABEL="semver:patch"
fi

echo "::set-output name=old-version::${OLD_VERSION}"
echo "::set-output name=new-version::${NEW_VERSION}"
echo "::set-output name=version-label::${LABEL}"
