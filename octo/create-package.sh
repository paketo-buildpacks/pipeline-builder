#!/usr/bin/env bash

set -euo pipefail

# With Go 1.20, we need to set this so that we produce statically compiled binaries
#
# Starting with Go 1.20, Go will produce binaries that are dynamically linked against libc
#   which can cause compatibility issues. The compiler links against libc on the build system
#   but that may be newer than on the stacks we support.
export CGO_ENABLED=0

if [[ "${EXTENSION:-x}" == "true" ]]; then
  export PACK_MODULE_TYPE=extension
else
  export PACK_MODULE_TYPE=buildpack
fi

if [[ "${INCLUDE_DEPENDENCIES}" == "true" ]]; then
  create-package \
    --source ${SOURCE_PATH:-.} \
    --cache-location "${HOME}"/carton-cache \
    --destination "${HOME}"/${PACK_MODULE_TYPE} \
    --include-dependencies \
    --version "${VERSION}"
else
  create-package \
    --source ${SOURCE_PATH:-.} \
    --destination "${HOME}"/${PACK_MODULE_TYPE} \
    --version "${VERSION}"
fi

PACKAGE_FILE=${SOURCE_PATH:-.}/package.toml
[[ -e ${PACKAGE_FILE} ]] && cp ${PACKAGE_FILE} "${HOME}"/package.toml
printf '['${PACK_MODULE_TYPE}']\nuri = "%s"\n\n[platform]\nos = "%s"\n' "${HOME}"/${PACK_MODULE_TYPE} "${OS}" >> "${HOME}"/package.toml
