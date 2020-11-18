#!/usr/bin/env bash

set -euo pipefail

if [[ "${INCLUDE_DEPENDENCIES}" == "true" ]]; then
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
printf '[buildpack]\nuri = "%s"\n\n[platform]\nos = "%s"\n' "${HOME}"/buildpack "${OS}" >> "${HOME}"/package.toml
