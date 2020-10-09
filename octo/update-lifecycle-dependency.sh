#!/usr/bin/env bash

set -euo pipefail

OLD_VERSION=$(yj -tj < builder.toml | jq -r '.lifecycle.uri | capture(".+/v(?<version>[\\d]+\\.[\\d]+\\.[\\d]+)/.+") | .version')

update-lifecycle-dependency \
  --builder-toml builder.toml \
  --version "${VERSION}"

git add builder.toml
git checkout -- .

echo "::set-output name=old-version::${OLD_VERSION}"
echo "::set-output name=new-version::${VERSION}"
