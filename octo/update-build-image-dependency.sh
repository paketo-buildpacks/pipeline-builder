#!/usr/bin/env bash

set -euo pipefail

OLD_VERSION=$(yj -tj < builder.toml | jq -r ".stack.\"build-image\" | capture(\"${IMAGE}:(?<version>.+-${CLASSIFIER})\") | .version")
NEW_VERSION=$(crane ls "${IMAGE}" | grep ".*-${CLASSIFIER}" | sort -V | tail -n 1)

update-build-image-dependency \
  --builder-toml builder.toml \
  --version "${NEW_VERSION}"

git add builder.toml
git checkout -- .

echo "::set-output name=old-version::${OLD_VERSION}"
echo "::set-output name=new-version::${NEW_VERSION}"
