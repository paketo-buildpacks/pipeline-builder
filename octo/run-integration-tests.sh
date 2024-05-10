#!/usr/bin/env bash

set -euo pipefail

go test ./integration/... -run Integration
