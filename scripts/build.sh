#!/usr/bin/env bash

set -euo pipefail

ROOT=$(realpath "$(dirname "${BASH_SOURCE[0]}")"/../..)

if [[ -d "${ROOT}"/go-cache ]]; then
  export GOPATH="${ROOT}"/go-cache
  export PATH="${ROOT}"/go-cache/bin:${PATH}
fi

cd "${ROOT}"/pipeline-builder

for RESOURCE in $(find "${ROOT}"/pipeline-builder/resources/cmd -maxdepth 1 -mindepth 1 -type d -exec basename {} \; | sort) ; do
  printf "Building %s\n" "$RESOURCE"

  cp "${ROOT}"/pipeline-builder/resources/Dockerfile "${ROOT}"/"${RESOURCE}"-resource
  GOOS="linux" go build -ldflags='-s -w' -o "${ROOT}"/"${RESOURCE}"-resource/check github.com/paketoio/pipeline-builder/resources/cmd/"${RESOURCE}"/check
  GOOS="linux" go build -ldflags='-s -w' -o "${ROOT}"/"${RESOURCE}"-resource/in    github.com/paketoio/pipeline-builder/resources/cmd/"${RESOURCE}"/in
  GOOS="linux" go build -ldflags='-s -w' -o "${ROOT}"/"${RESOURCE}"-resource/out   github.com/paketoio/pipeline-builder/resources/cmd/"${RESOURCE}"/out
done
