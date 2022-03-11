#!/usr/bin/env bash

set -euo pipefail

richgo test ./integration/... -run Integration
