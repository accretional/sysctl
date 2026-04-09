#!/usr/bin/env bash
# setup.sh — Idempotent project setup.
#
# IDEMPOTENCY CONTRACT:
#   This script checks before acting. Running it 100 times costs ~1 second
#   after the first run. It will:
#   - Verify Go 1.26.x is installed (does NOT install it)
#   - Install protoc via brew if missing
#   - Install protoc-gen-go and protoc-gen-go-grpc if missing
#   - Generate proto stubs if proto has changed or stubs are missing
#   - Run go mod tidy (cheap, idempotent)
set -euo pipefail

cd "$(dirname "$0")"

echo "=== setup.sh ==="

# 1. Verify Go version.
REQUIRED_GO_MINOR="1.26"
GO_VERSION=$(go version 2>/dev/null | grep -oE 'go[0-9]+\.[0-9]+' | head -1)
if [[ -z "$GO_VERSION" ]]; then
    echo "ERROR: Go is not installed. Install Go ${REQUIRED_GO_MINOR}.x first."
    exit 1
fi
if [[ "$GO_VERSION" != "go${REQUIRED_GO_MINOR}" ]]; then
    echo "ERROR: Go ${REQUIRED_GO_MINOR}.x required, found $GO_VERSION"
    exit 1
fi
echo "  Go version OK: $(go version)"

# 2. Install protoc if missing.
if ! command -v protoc &>/dev/null; then
    echo "  Installing protoc via brew..."
    brew install protobuf
else
    echo "  protoc OK: $(protoc --version)"
fi

# 3. Install protoc Go plugins if missing.
GOBIN=$(go env GOBIN)
if [[ -z "$GOBIN" ]]; then
    GOBIN=$(go env GOPATH)/bin
fi

if [[ ! -x "$GOBIN/protoc-gen-go" ]]; then
    echo "  Installing protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
else
    echo "  protoc-gen-go OK"
fi

if [[ ! -x "$GOBIN/protoc-gen-go-grpc" ]]; then
    echo "  Installing protoc-gen-go-grpc..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
else
    echo "  protoc-gen-go-grpc OK"
fi

export PATH="$GOBIN:$PATH"

# 4. Generate proto stubs if needed.
PROTO_SRC="proto/sysctlpb/sysctl.proto"
PROTO_OUT="proto/sysctlpb"

NEED_REGEN=false
if [[ ! -f "$PROTO_OUT/sysctl.pb.go" ]] || [[ ! -f "$PROTO_OUT/sysctl_grpc.pb.go" ]]; then
    NEED_REGEN=true
elif [[ "$PROTO_SRC" -nt "$PROTO_OUT/sysctl.pb.go" ]]; then
    NEED_REGEN=true
fi

if $NEED_REGEN; then
    echo "  Generating protobuf stubs..."
    protoc \
        --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        "$PROTO_SRC"
    echo "  Proto stubs generated."
else
    echo "  Proto stubs up to date"
fi

# 5. go mod tidy.
echo "  Running go mod tidy..."
go mod tidy
echo "  go mod tidy done"

echo "=== setup.sh complete ==="
