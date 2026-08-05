#!/usr/bin/env bash

set -euo pipefail

docker run -d --name local-registry -p 5000:5000 registry:2

for i in $(seq 1 30); do
  if curl -sf http://localhost:5000/v2/ >/dev/null 2>&1; then
    echo "Local registry is ready (waited ${i}s)"
    exit 0
  fi
  sleep 1
done

echo "ERROR: local registry did not become ready" >&2
docker logs local-registry >&2 || true
exit 1
