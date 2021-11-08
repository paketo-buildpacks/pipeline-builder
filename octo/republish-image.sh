#!/usr/bin/env bash

set -euo pipefail

if [ "$SOURCE" == "" ] || [ "$TARGET" == "" ] || [ "$VERSION" == "" ] || [ "$NEWID" == "" ]; then
    echo "Missing source, target, version and/or newid:"
    echo "   Source: $SOURCE"
    echo "   Target: $TARGET"
    echo "  Version: $VERSION"
    echo "    NewID: $NEWID"
    exit -1
fi

update-buildpack-image-id \
    --image "$SOURCE:$VERSION" \
    --id "$NEWID" \
    --new-image "$TARGET:$VERSION"
