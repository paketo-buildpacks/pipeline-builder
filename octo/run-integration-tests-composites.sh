#!/usr/bin/env bash

set -euo pipefail

BP_UNDER_TEST=ttl.sh/${PACKAGE}-${VERSION}:1h make integration
