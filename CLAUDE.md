# CLAUDE.md — Project conventions for sysctl

## Quick start

```bash
bash LET_IT_RIP.sh   # Setup + build + test + smoke test. Run this before every push.
```

## Scripts (all idempotent)

| Script | What it does |
|---|---|
| `setup.sh` | Verifies Go 1.26.x, installs protoc/plugins if missing, regenerates proto stubs if changed, runs `go mod tidy`. Skips work already done. |
| `build.sh` | Builds `bin/server` and `bin/client`. Go's build cache makes repeated runs fast. |
| `test.sh` | Runs ALL tests: unit (asm sysctl), integration (gRPC server), and e2e (full client-server flow). Uses `-count=1` to avoid cached results. |
| `LET_IT_RIP.sh` | Calls setup → build → test → smoke test (starts server, queries via client, verifies output). **Always run this before pushing.** |

### Idempotency contract

All scripts are designed to be cheap on repeat runs:
- `setup.sh` checks before acting (skips installs if present, skips proto gen if unchanged)
- `build.sh` relies on Go's build cache
- `test.sh` is stateless (reads kernel metrics, no mutations)
- `LET_IT_RIP.sh` starts/stops its own server on port 50099

Running `LET_IT_RIP.sh` 10 times in a row is fine and expected.

## Architecture

```
internal/macosasmsysctl/     ARM64 assembly + Go: direct sysctl via libSystem trampolines
internal/metrics/            Registry of known sysctl metric names and types
internal/server/             gRPC server implementation
proto/sysctlpb/              Protobuf definition + generated Go code
cmd/server/                  Server binary (default port 50051)
cmd/client/                  Client binary (-list, -all, or specific metric names)
```

### Assembly approach

The `macosasmsysctl` package calls `sysctl(3)` and `sysctlbyname(3)` directly via ARM64 assembly trampolines into `/usr/lib/libSystem.B.dylib`. No cgo, no external dependencies. This is the same mechanism Go's runtime and `x/sys/unix` use internally:

1. `cgo_import.go` — `//go:cgo_import_dynamic` imports libc symbols
2. `sysctl_darwin_arm64.s` — Assembly `JMP` trampolines to those symbols
3. `linkname.go` — `//go:linkname` links to `syscall.syscall6` for dispatch
4. `sysctl.go` — Go wrapper functions (GetString, GetUint64, GetRaw, etc.)

### Type notes (ARM64 macOS)

Some sysctl values that are traditionally `int32` on x86 return 8 bytes on ARM64 (e.g., `hw.pagesize`, `hw.cachelinesize`, cache sizes). The metrics registry uses `int64` for these. `hw.cpufrequency` and `hw.busfrequency` don't exist on Apple Silicon — they were removed from the registry.

## Workflow rules

- **Never push without a green `LET_IT_RIP.sh`** — not just the specific test for what you changed
- **Never run one-off tests as sufficient** — always validate the full flow
- Proto changes: edit `proto/sysctlpb/sysctl.proto`, then `setup.sh` regenerates stubs
- New metrics: add to `internal/metrics/registry.go`
