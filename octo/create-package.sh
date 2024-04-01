#!/usr/bin/env bash

set -xeuo pipefail

# With Go 1.20, we need to set this so that we produce statically compiled binaries
#
# Starting with Go 1.20, Go will produce binaries that are dynamically linked against libc
#   which can cause compatibility issues. The compiler links against libc on the build system
#   but that may be newer than on the stacks we support.
export CGO_ENABLED=0

if [[ "${INCLUDE_DEPENDENCIES}" == "true" ]]; then
  create-package \
    --source "${SOURCE_PATH:-.}" \
    --cache-location "${HOME}"/carton-cache \
    --destination "${HOME}"/buildpack \
    --include-dependencies \
    --version "${VERSION}"
else
  create-package \
    --source "${SOURCE_PATH:-.}" \
    --destination "${HOME}"/buildpack \
    --version "${VERSION}"
fi

PACKAGE_FILE="${SOURCE_PATH:-.}/package.toml"
if [ -f "${PACKAGE_FILE}" ]; then
  cp "${PACKAGE_FILE}" "${HOME}/buildpack/package.toml"
  printf '[buildpack]\nuri = "%s"\n\n[platform]\nos = "%s"\n' "${HOME}/buildpack" "${OS}" >> "${HOME}/buildpack/package.toml"
fi
