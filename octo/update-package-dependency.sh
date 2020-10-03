#!/usr/bin/env bash

set -euo pipefail

OLD_VERSION=$(yj -tj < package.toml | jq -r ".dependencies[].image | select(. | startswith(\"${DEPENDENCY}\"))")
OLD_VERSION=${OLD_VERSION#*:}
NEW_VERSION=$(crane ls "${DEPENDENCY}" | grep -v latest | sort -V | tail -n1)

update-package-dependency \
  --buildpack-toml buildpack.toml \
  --package-toml package.toml \
  --id "${DEPENDENCY}" \
  --version "${NEW_VERSION}"

echo "::set-output name=old-version::${OLD_VERSION}"
echo "::set-output name=new-version::${NEW_VERSION}"
