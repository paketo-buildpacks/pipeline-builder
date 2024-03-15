#!/usr/bin/env bash
# this is coming from a copy of https://github.com/buildpacks/pack/actions/runs/8118576298 stored on box
# TODO to revisit when the official one is out
set -euo pipefail

echo "Installing pack experimental"

mkdir -p "${HOME}"/bin
echo "${HOME}/bin" >> "${GITHUB_PATH}"

curl -L "https://ent.box.com/shared/static/j4d1bfe9uk1sb0i7zjvci0md9xmy41u4" -o ${HOME}/bin/pack
chmod +x "${HOME}"/bin/pack
