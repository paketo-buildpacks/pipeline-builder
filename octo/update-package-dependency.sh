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
elif [[ -e package.toml ]]; then
  OLD_VERSION=$(yj -tj < package.toml | jq -r ".dependencies[].uri | capture(\".*${DEPENDENCY}:(?<version>.+)\") | .version")

  update-package-dependency \
    --buildpack-toml buildpack.toml \
    --package-toml package.toml \
    --id "${DEPENDENCY}" \
    --version "${NEW_VERSION}"

  git add buildpack.toml package.toml
fi

git checkout -- .

echo "::set-output name=old-version::${OLD_VERSION}"
echo "::set-output name=new-version::${NEW_VERSION}"
