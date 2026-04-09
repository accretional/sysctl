#!/usr/bin/env bash
# LET_IT_RIP.sh — Full end-to-end flow: setup, build, test, run.
#
# IDEMPOTENCY CONTRACT:
#   Every sub-script is idempotent. This script is safe to run at any time,
#   from a clean checkout or mid-development. It will:
#   1. setup.sh  — ensure tools, generate protos, tidy deps (skips if done)
#   2. build.sh  — compile binaries (Go cache makes this fast)
#   3. test.sh   — run ALL tests (unit + integration + e2e)
#   4. Smoke test — start server, query via client, verify output, stop
#
# Run this before every push. If it passes, the project is healthy.
set -euo pipefail

cd "$(dirname "$0")"

echo "========================================"
echo "  LET IT RIP — Full E2E Flow"
echo "========================================"
echo ""

# 1. Setup.
bash setup.sh
echo ""

# 2. Build.
bash build.sh
echo ""

# 3. Test.
bash test.sh
echo ""

# 4. Smoke test: run server + client e2e.
echo "=== Smoke test: server + client ==="

PORT=50099  # Use a non-default port to avoid conflicts.

# Start server in background.
echo "  Starting server on port $PORT..."
bin/server -port "$PORT" &
SERVER_PID=$!

# Give server time to start.
sleep 1

# Verify server is running.
if ! kill -0 "$SERVER_PID" 2>/dev/null; then
    echo "  ERROR: Server failed to start"
    exit 1
fi

cleanup() {
    echo "  Stopping server (PID $SERVER_PID)..."
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
}
trap cleanup EXIT

# Run client: list metrics.
echo "  Client: listing known metrics..."
bin/client -addr "localhost:$PORT" -list
echo ""

# Run client: fetch all metrics.
echo "  Client: fetching all metrics..."
bin/client -addr "localhost:$PORT" -all
echo ""

# Run client: fetch specific metric and verify.
echo "  Client: verifying kern.ostype = Darwin..."
OUTPUT=$(bin/client -addr "localhost:$PORT" kern.ostype)
echo "  $OUTPUT"
if echo "$OUTPUT" | grep -q "Darwin"; then
    echo "  ✓ kern.ostype verified"
else
    echo "  ✗ kern.ostype verification FAILED"
    exit 1
fi

echo ""
echo "========================================"
echo "  ALL CHECKS PASSED"
echo "========================================"
