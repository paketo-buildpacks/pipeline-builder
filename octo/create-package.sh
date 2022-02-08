#!/usr/bin/env bash

set -euo pipefail

if [[ "${INCLUDE_DEPENDENCIES}" == "true" ]]; then
  create-package \
    --source ${SOURCE_PATH:-.} \
    --cache-location "${HOME}"/carton-cache \
    --destination "${HOME}"/buildpack \
    --include-dependencies \
    --version "${VERSION}"
else
  create-package \
    --source ${SOURCE_PATH:-.} \
    --destination "${HOME}"/buildpack \
    --version "${VERSION}"
fi

PACKAGE_FILE=${SOURCE_PATH:-.}/package.toml
[[ -e ${PACKAGE_FILE} ]] && cp ${PACKAGE_FILE} "${HOME}"/package.toml
printf '[buildpack]\nuri = "%s"\n\n[platform]\nos = "%s"\n' "${HOME}"/buildpack "${OS}" >> "${HOME}"/package.toml
