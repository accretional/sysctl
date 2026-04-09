#!/usr/bin/env bash
# test.sh — Run ALL tests (unit, integration, e2e).
#
# IDEMPOTENCY CONTRACT:
#   Tests are stateless reads of sysctl values. Safe to run repeatedly.
set -euo pipefail

cd "$(dirname "$0")"

echo "=== test.sh ==="

echo "  Running all tests with verbose output..."
go test -v -count=1 ./...

echo "=== test.sh complete ==="
