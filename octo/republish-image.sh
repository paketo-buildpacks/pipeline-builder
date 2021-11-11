#!/usr/bin/env bash

set -euo pipefail

if [ "$SOURCE" == "" ] || [ "$TARGET" == "" ] || [ "$SOURCE_VERSION" == "" ] || [ "$TARGET_VERSION" == "" ] || [ "$NEWID" == "" ]; then
    echo "Missing source, target, version and/or newid:"
    echo "     Source: $SOURCE"
    echo "    Version: $SOURCE_VERSION"
    echo "     Target: $TARGET"
    echo "    Version: $TARGET_VERSION"
    echo "     New ID: $NEWID"
    exit 255
fi

update-buildpack-image-id \
    --image "$SOURCE:$SOURCE_VERSION" \
    --id "$NEWID" \
    --version "$TARGET_VERSION" \
    --new-image "$TARGET:$TARGET_VERSION"
