#!/usr/bin/env bash

set -euo pipefail

echo "Enabling pack experimental features"

mkdir -p "${HOME}"/.pack
echo "experimental = true" >> "${HOME}"/.pack/config.toml
