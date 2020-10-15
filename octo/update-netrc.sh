#!/usr/bin/env bash

set -euo pipefail

echo "machine ${HOST} login ${USERNAME} password ${PASSWORD}" >> "${HOME}"/.netrc
