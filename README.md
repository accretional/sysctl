# sysctl

Kernel performance telemetry for macOS via gRPC. Reads 250 sysctl metrics
directly from the Darwin kernel using ARM64 assembly trampolines (no cgo),
serves them over a typed protobuf API with background polling, delta streaming,
and per-metric access pattern classifications.

## Quick start

```bash
bash LET_IT_RIP.sh          # build + test + smoke test (idempotent)
bin/server --port 50051      # start server
bin/client -all              # fetch all 250 metrics
bin/client kern.ostype hw.memsize vm.loadavg   # fetch specific metrics
bin/client -cats             # list categories
bin/client -cat hw.cpu       # fetch all hw.cpu metrics
bin/client -list             # list known metrics with access patterns
```

## What it does

The server reads macOS kernel metrics through the `sysctl` syscall and
exposes them via gRPC. It pre-resolves metric names to integer MIB arrays
at startup for ~3x faster reads, classifies each metric by how often it
changes, and runs a background poller that keeps frequently-accessed metrics
fresh in cache.

### Metric coverage

250 metrics across 32 categories:

| Domain | Categories | Examples |
|--------|-----------|---------|
| Hardware | hw.cpu, hw.memory, hw.cache, hw.perflevel, hw.arm, hw.misc | CPU topology, memory size, cache hierarchy, P/E core info, ARM feature flags |
| Kernel | kern.identity, kern.process, kern.memory, kern.sched, kern.time, kern.limits, kern.ipc, kern.misc | OS version, process counts, memory pressure, scheduler config, boot time, file descriptor limits |
| Virtual memory | vm.pressure, vm.pages, vm.compressor, vm.pageout, vm.swap, vm.wire, vm.misc | Page counts, compressor stats, swap usage, wire limits, load averages |
| Network | net.tcp, net.udp, net.ip, net.misc | TCP/UDP config tunables, connection counts, IP settings |
| Filesystem | vfs | Vnode stats, mount ops, APFS allocation |
| Other | machdep, debug, kpc, iogpu, security | CPU timing, BPF config, perf counters, GPU limits, MAC stats |
| Computed | computed | Memory utilization %, compression ratio, uptime, total connections |

### Access patterns

Every metric is classified with two access patterns (see [CACHE_DESIGN.md](CACHE_DESIGN.md)):

| Pattern | Count | Behavior |
|---------|-------|---------|
| STATIC | 81 | Read once at startup, cached forever (hardware properties, OS identity) |
| POLLED | 118 | Background refresh at 10-60s TTL (counters, gauges) |
| CONSTRAINED | 51 | Background refresh at 60s TTL (writable admin tunables like `kern.maxproc`, `net.inet.tcp.keepidle`) |

CONSTRAINED metrics were identified by cross-referencing all 250 metrics
against `sysctl -aW` (writable sysctls) on Darwin ARM64. They change only by
explicit admin action, not continuously. See [CACHE_DESIGN.md](CACHE_DESIGN.md)
for the monotonic access pattern chain design.

### API

```protobuf
service SysctlService {
  rpc GetMetrics(GetMetricsRequest) returns (GetMetricsResponse);
  rpc GetMetricsByCategory(GetMetricsByCategoryRequest) returns (GetMetricsResponse);
  rpc Subscribe(SubscribeRequest) returns (stream SubscribeResponse);
  rpc ListKnownMetrics(ListKnownMetricsRequest) returns (ListKnownMetricsResponse);
  rpc ListCategories(ListCategoriesRequest) returns (ListCategoriesResponse);
  rpc GetKernelRegistry(GetKernelRegistryRequest) returns (GetKernelRegistryResponse);
}
```

**Subscribe** streams raw sysctl byte deltas at a client-requested interval.
After the initial full snapshot, only changed values are sent. The server
rejects intervals faster than `min_interval_ns` (derived from the fastest
recommended TTL in the registry). Computed metrics are not supported in
Subscribe — use GetMetrics for those.

See [DELTA_DESIGN.md](DELTA_DESIGN.md) for the compact encoding roadmap.

## Architecture

```
cmd/server/                  Server binary (--port, --os-version, --poll-interval)
cmd/client/                  CLI client (-list, -all, -cats, -cat, metric names)
cmd/darwin-name2int/         MIB mapping research tool (see below)
internal/macosasmsysctl/     ARM64 asm + Go: direct sysctl via libSystem trampolines
                             MIB cache for ~3x faster reads (mibcache.go)
internal/metrics/            Registry of 250 metrics + kernel access pattern
                             classifications (darwin/24.6.0.textproto)
internal/server/             gRPC server: polled cache, computed aggregates,
                             delta streaming, kernel registry
proto/sysctlpb/              Protobuf definitions + generated Go code
```

### How it calls the kernel

The `macosasmsysctl` package calls `sysctl(3)` and `sysctlbyname(3)` directly
via ARM64 assembly trampolines into `/usr/lib/libSystem.B.dylib`. No cgo, no
external dependencies. This is the same mechanism Go's runtime and `x/sys/unix`
use internally. See [SYSCTL_101.md](SYSCTL_101.md) for a deep dive on how
sysctl works.

Files in `internal/macosasmsysctl/`:

1. `cgo_import.go` — `//go:cgo_import_dynamic` imports libc symbols
2. `sysctl_darwin_arm64.s` — Assembly `JMP` trampolines to those symbols
3. `linkname.go` — `//go:linkname` links to `syscall.syscall6` for dispatch
4. `sysctl.go` — Go wrappers (`GetString`, `GetUint64`, `GetRaw`, etc.)
5. `mibcache.go` — MIB pre-resolution cache

### ARM64 type widening

Some sysctl values that are `int32` on x86 return 8 bytes on ARM64. Apple
re-registered these as new-style OIDs with wider types:

| Metric | Legacy MIB | Legacy size | ARM64 MIB | ARM64 size |
|--------|-----------|-------------|-----------|------------|
| hw.pagesize | {6,7} | 4 bytes (int32) | {6,115} | 8 bytes (int64) |
| hw.cachelinesize | — | 4 bytes | {6,123} | 8 bytes |
| hw.l1icachesize | — | 4 bytes | {6,124} | 8 bytes |
| hw.l1dcachesize | — | 4 bytes | {6,125} | 8 bytes |
| hw.l2cachesize | — | 4 bytes | {6,126} | 8 bytes |
| hw.tbfrequency | — | 4 bytes | {6,128} | 8 bytes |

The legacy MIB `{6,7}` for `hw.pagesize` still works and returns 16384, but as
a 4-byte int32. `sysctlbyname("hw.pagesize")` resolves to the new OID `{6,115}`
which returns the same value as an 8-byte int64. Our MIB cache uses runtime
name resolution, so it automatically gets the wider types. Only 4 of 60 hw.*
metrics still use legacy OIDs (≤27); the other 56 use new-style OIDs (≥100).

See [cmd/darwin-name2int/FINDINGS.md](cmd/darwin-name2int/FINDINGS.md) for the
full investigation.

`hw.cpufrequency` and `hw.busfrequency` don't exist on Apple Silicon — they
were removed from the registry.

## MIB mapping research (cmd/darwin-name2int/)

The `darwin-name2int` tool resolves all 243 non-computed metric names to their
integer MIB arrays and cross-references them against XNU kernel header constants
(`bsd/sys/sysctl.h`). Key findings:

- **All 243 resolve to unique MIB arrays** — viable as metric identifiers
- **96% use well-known CTL_* constants** (1=kern, 2=vm, 4=net, 6=hw, etc.)
- **4% use extended OID space** (101=kperf, 102=kpc, 103=security, 104=iogpu)
- **21 of 22 XNU header constants match** — only hw.pagesize diverges (type widening, see above)
- **All MIBs fit in 40 bits** — max depth 5, max value per level 327 (9 bits)
- **Sequential enum recommended** over MIB-based encoding for wire protocol

```bash
go run ./cmd/darwin-name2int/ -validate   # resolve + validate all metrics
go run ./cmd/darwin-name2int/ -json       # full MIB mappings as JSON
go test -v ./cmd/darwin-name2int/         # run all validation tests
```

See [cmd/darwin-name2int/FINDINGS.md](cmd/darwin-name2int/FINDINGS.md) for
detailed analysis.

## Design documents

| Document | Contents |
|----------|---------|
| [SYSCTL_101.md](SYSCTL_101.md) | How sysctl works, MIB addressing, type system, what this project does with it |
| [CACHE_DESIGN.md](CACHE_DESIGN.md) | Access patterns, monotonic chain, poller implementation, future tiered access |
| [DELTA_DESIGN.md](DELTA_DESIGN.md) | Compact delta encoding roadmap (enum IDs, MetricDescriptor, CompactDelta) |
| [CLAUDE.md](CLAUDE.md) | Development conventions, scripts, workflow rules |
| [cmd/darwin-name2int/FINDINGS.md](cmd/darwin-name2int/FINDINGS.md) | MIB mapping research results |

## Development

```bash
bash LET_IT_RIP.sh   # always run before pushing (setup + build + test + smoke)
```

See [CLAUDE.md](CLAUDE.md) for full development conventions.
