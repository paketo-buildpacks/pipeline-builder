#!/usr/bin/env bash

set -euo pipefail

if [ -z "${GO_VERSION:-}" ]; then
    echo "No go version set"
    exit 1
fi

OLD_GO_VERSION=$(grep -P '^go \d\.\d+' go.mod | cut -d ' ' -f 2)

go mod edit -go="$GO_VERSION"
go mod tidy
go get -u -t ./...
go mod tidy

git add go.mod go.sum
git checkout -- .

if [ "$OLD_GO_VERSION" == "$GO_VERSION" ]; then
    COMMIT_TITLE="Bump Go Modules"
    COMMIT_BODY="Bumps Go modules used by the project. See the commit for details on what modules were updated."
    COMMIT_SEMVER="semver:patch"
else
    COMMIT_TITLE="Bump Go from ${OLD_GO_VERSION} to ${GO_VERSION}"
    COMMIT_BODY="Bumps Go from ${OLD_GO_VERSION} to ${GO_VERSION} and update Go modules used by the project. See the commit for details on what modules were updated."
    COMMIT_SEMVER="semver:minor"
fi

echo "commit-title=${COMMIT_TITLE}" >> "$GITHUB_OUTPUT"
echo "commit-body=${COMMIT_BODY}" >> "$GITHUB_OUTPUT"
echo "commit-semver=${COMMIT_SEMVER}" >> "$GITHUB_OUTPUT"
