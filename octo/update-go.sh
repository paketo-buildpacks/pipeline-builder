#!/usr/bin/env bash

set -euo pipefail

if [ -z "${GO_VERSION:-}" ]; then
    echo "No go version set"
    exit 1
fi

OLD_GO_VERSION=$(grep -P '^go \d\.\d+' go.mod | cut -d ' ' -f 2)

go mod edit -go="$GO_VERSION"
go mod tidy
go get -u all
go mod tidy

git add go.mod go.sum
git checkout -- .

echo "::set-output name=old-go-version::${OLD_GO_VERSION}"
echo "::set-output name=go-version::${GO_VERSION}"
