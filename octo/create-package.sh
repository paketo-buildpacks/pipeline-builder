#!/usr/bin/env bash

set -euo pipefail

if [[ -n "${INCLUDE_DEPENDENCIES+x}" ]]; then
  create-package \
    --cache-location "${HOME}"/carton-cache \
    --destination "${HOME}"/buildpack \
    --include-dependencies \
    --version "${VERSION}"
else
  create-package \
    --destination "${HOME}"/buildpack \
    --version "${VERSION}"
fi

[[ -e package.toml ]] && cp package.toml "${HOME}"/package.toml
printf '[buildpack]\nuri = "%s"' "${HOME}"/buildpack >> "${HOME}"/package.toml
