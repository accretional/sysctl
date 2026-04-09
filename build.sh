#!/usr/bin/env bash
# build.sh — Build all binaries.
#
# IDEMPOTENCY CONTRACT:
#   Always rebuilds (Go's build cache makes repeated builds fast).
#   Produces: bin/server, bin/client
set -euo pipefail

cd "$(dirname "$0")"

echo "=== build.sh ==="

mkdir -p bin

echo "  Building server..."
go build -o bin/server ./cmd/server

echo "  Building client..."
go build -o bin/client ./cmd/client

echo "  Binaries:"
ls -lh bin/
echo "=== build.sh complete ==="
