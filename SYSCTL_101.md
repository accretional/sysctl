# Sysctl 101

A deep dive on how macOS sysctl works, what it exposes, how we use it, and
what we've learned building this project.

## What is sysctl?

Sysctl is the kernel interface for reading (and sometimes writing) system
parameters on BSD-derived operating systems, including macOS. It's the
canonical way to query hardware properties, kernel configuration, memory
statistics, network tunables, and process counts without parsing files or
calling specialized APIs.

On Linux, this information is scattered across `/proc`, `/sys`, and various
system calls. On macOS, sysctl is the single unified interface. There is no
procfs.

## The two syscalls

There are two ways to read a sysctl value:

### sysctlbyname(name, ...) — the string API

```c
// Read kern.ostype by name
char buf[256];
size_t len = sizeof(buf);
sysctlbyname("kern.ostype", buf, &len, NULL, 0);
// buf = "Darwin"
```

This is the easy API. You pass a dotted string name like `"hw.memsize"` and get
back raw bytes. Internally, the kernel resolves the name to an integer MIB array
on every call.

### sysctl(mib, ...) — the integer API

```c
// Read kern.ostype by MIB: CTL_KERN=1, KERN_OSTYPE=1
int mib[2] = {1, 1};
char buf[256];
size_t len = sizeof(buf);
sysctl(mib, 2, buf, &len, NULL, 0);
// buf = "Darwin"
```

This is the fast API. You pass an integer array (the MIB) directly. No name
resolution overhead. But you need to know the MIB values.

### The performance difference

Name resolution adds overhead on every call. In our benchmarks, pre-resolved
MIB reads are ~3x faster than `sysctlbyname`. For a telemetry service reading
hundreds of metrics many times per second, this matters.

Our approach: resolve names to MIBs once at startup (using the kernel's
`nametomib` facility, sysctl `{0,3}`), cache the MIB arrays, and use the
integer API for all subsequent reads. This is what `mibcache.go` does.

## MIB addressing

MIB stands for Management Information Base, borrowed from SNMP. Each sysctl
metric is identified by an array of integers forming a hierarchical path,
similar to a filesystem path but with integers instead of strings.

### Top-level categories

The first integer selects a major subsystem, defined in XNU's `bsd/sys/sysctl.h`:

| MIB[0] | Constant | Subsystem |
|--------|----------|-----------|
| 1 | CTL_KERN | Kernel (OS info, limits, IPC, scheduling, time) |
| 2 | CTL_VM | Virtual memory (pages, compressor, swap, pressure) |
| 3 | CTL_VFS | Virtual filesystem (vnodes, mount ops, APFS) |
| 4 | CTL_NET | Networking (TCP, UDP, IP config and counters) |
| 5 | CTL_DEBUG | Debug (BPF config, throttle) |
| 6 | CTL_HW | Hardware (CPU, memory, cache, ARM features) |
| 7 | CTL_MACHDEP | Machine-dependent (CPU brand, timing, idle level) |
| 8 | CTL_USER | User-level (POSIX limits, path config) |

These have been stable since 4.4BSD (1993). They will not change.

### Extended categories

Modern macOS adds subsystems beyond the original 9:

| MIB[0] | Subsystem | Metrics |
|--------|-----------|---------|
| 101 | kperf | Performance sampling |
| 102 | kpc | Performance counter config |
| 103 | security | MAC framework, ASP stats |
| 104 | iogpu | GPU memory limits |

These use OID values ≥100 and are not defined in the public XNU headers.
They are assigned by the kernel's dynamic OID registration system.

### Subtree structure

Each category is a tree. The second integer (and sometimes third, fourth)
selects a specific metric within the tree:

```
kern.ostype          → {1, 1}          CTL_KERN / KERN_OSTYPE
kern.maxproc         → {1, 6}          CTL_KERN / KERN_MAXPROC
kern.ipc.maxsockbuf  → {1, 32, 1}     CTL_KERN / KERN_IPC / KIPC_MAXSOCKBUF
net.inet.tcp.mssdflt → {4, 2, 6, 3}   CTL_NET / PF_INET / IPPROTO_TCP / TCPCTL_MSSDFLT
```

Network metrics have the deepest trees: CTL_NET / protocol family / protocol /
sub-ID. Our deepest metrics are 5 levels: `security.mac.asp.stats.*`.

### How nametomib works

The kernel exposes a special sysctl `{0, 3}` that converts a string name to
its MIB array. This is what `sysctlbyname` uses internally, and what we call
once per metric at startup:

```go
// nametomib("kern.ostype") → []uint32{1, 1}
mib := []uint32{0, 3}  // magic "name2oid" sysctl
rawSysctl(mib, &resultBuf, &resultLen, &nameBytes, nameLen)
```

The result is the MIB array that can be used with `sysctl()` directly.

## Value types

Sysctl returns raw bytes. The caller is responsible for knowing the type.
There is no type metadata in the response (though the kernel does track types
internally for the `sysctl -a` formatter).

### Integer types

Most metrics are integers. On ARM64 macOS:

| Type | Size | Examples |
|------|------|---------|
| int32 | 4 bytes | hw.ncpu, kern.maxproc, most counters |
| uint32 | 4 bytes | vm.page_free_count, load average components |
| int64 | 8 bytes | hw.pagesize, hw.cachelinesize (widened on ARM64) |
| uint64 | 8 bytes | hw.memsize, vm.compressor_bytes_used |

227 of our 243 non-computed metrics are integers (≤8 bytes).

### Struct types

Some metrics return fixed-size C structs:

| Type | Size | Structure | Examples |
|------|------|-----------|---------|
| timeval | 16 bytes | `{uint64 sec, uint64 usec}` | kern.boottime, kern.sleeptime |
| loadavg | 24 bytes | `{uint32 ldavg[3], pad, int64 fscale}` | vm.loadavg |
| swap | 32 bytes | `{uint64 total, avail, used, flags}` | vm.swapusage |
| clockinfo | 20 bytes | `{int32 hz, tick, spare, profhz, stathz}` | kern.clockrate |

These are all STATIC (read once). They never appear in delta streams.

### String types

15 metrics return NUL-terminated strings:

```
kern.ostype       → "Darwin"
kern.version      → "Darwin Kernel Version 24.6.0: ..." (104 bytes)
kern.hostname     → "Freds-Mac-mini.local"
kern.uuid         → "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
machdep.cpu.brand_string → "Apple M4"
```

All strings are STATIC (read once) except `kern.hostname` which is CONSTRAINED
(writable, polled at 60s).

### Raw types

2 metrics return raw byte arrays:

```
hw.cacheconfig → 80 bytes (10 × uint64: CPUs sharing each cache level)
hw.cachesize   → 80 bytes (10 × uint64: size of each cache level)
```

Both are STATIC.

### The key insight for delta streaming

**Every metric that changes at runtime is ≤8 bytes (a single integer).**

The 16 metrics larger than uint64 are all STATIC — they appear in the initial
snapshot but never in subsequent deltas. This means the delta stream after the
first push contains only fixed-size integer values, enabling a compact
`{int32 id, fixed64 value}` encoding with ~60% wire size reduction vs string
names.

## ARM64 type widening

This is one of the most important findings from our MIB mapping research
(see [cmd/darwin-name2int/FINDINGS.md](cmd/darwin-name2int/FINDINGS.md)).

When Apple moved from x86 to ARM64, several hardware metrics that were
traditionally 32-bit were re-registered as 64-bit values under new OIDs:

```
hw.pagesize:
  Legacy MIB {6, 7}  (HW_PAGESIZE from XNU header) → 4 bytes, value 16384
  ARM64  MIB {6,115} (new-style OID)                → 8 bytes, value 16384
```

Both MIBs work. Both return the correct value. But they differ in width:

- `sysctl({6,7}, ...)` returns `int32(16384)` — 4 bytes
- `sysctlbyname("hw.pagesize", ...)` resolves to `{6,115}` and returns
  `int64(16384)` — 8 bytes

This affects any code that hardcodes MIB integers from the XNU headers. The
XNU header constant `HW_PAGESIZE = 7` still works but returns a narrower type
than what `sysctlbyname` would give you.

### Scope of widening

Of 60 hw.* metrics in our registry:

| OID range | Count | Notes |
|-----------|-------|-------|
| Legacy (MIB[1] ≤ 27) | 4 | hw.ncpu, hw.byteorder, hw.memsize, hw.activecpu |
| New-style (MIB[1] ≥ 100) | 56 | Everything else, including widened types |

Widened metrics (4 bytes legacy → 8 bytes ARM64):
- `hw.pagesize` ({6,7} → {6,115})
- `hw.cachelinesize` → {6,123}
- `hw.l1icachesize` → {6,124}
- `hw.l1dcachesize` → {6,125}
- `hw.l2cachesize` → {6,126}
- `hw.tbfrequency` → {6,128}

Metrics that were always 64-bit:
- `hw.memsize` ({6,24}) — `HW_MEMSIZE` was defined as 64-bit from the start

Metrics that don't exist on ARM64:
- `hw.cpufrequency` (`HW_CPU_FREQ = 15`) — Apple Silicon doesn't expose this
- `hw.busfrequency` (`HW_BUS_FREQ = 14`) — same

### Why this matters

If you're building a sysctl reader:

1. **Use `sysctlbyname` for name resolution** — it gives you the correct
   modern OID. Don't hardcode MIB integers from XNU headers.
2. **Resolve names to MIBs once, then use `sysctl`** — this is what our MIB
   cache does. You get the speed of integer MIBs with the correctness of
   runtime resolution.
3. **Handle variable-width integers** — `hw.pagesize` might be 4 or 8 bytes
   depending on how you access it. Read the size from the first `sysctl` call.

## Writable sysctls

Most sysctl metrics are read-only, but 58 of our 250 registered metrics are
writable (`sysctl -aW`). These are admin-configurable tunables:

**Kernel limits**: `kern.maxproc`, `kern.maxfiles`, `kern.maxvnodes`,
`kern.maxfilesperproc`, `kern.maxprocperuid`, `kern.maxnbuf`

**IPC**: `kern.ipc.maxsockbuf`, `kern.ipc.somaxconn`

**Network**: `net.inet.tcp.mssdflt`, `net.inet.tcp.keepidle`,
`net.inet.tcp.sendspace`, `net.inet.tcp.recvspace`, `net.inet.tcp.sack`,
`net.inet.tcp.fastopen`, `net.inet.ip.forwarding`, `net.inet.ip.ttl`, ...

**Other**: `kern.hostname`, `debug.bpf_bufsize`, `iogpu.wired_limit_mb`

We classify these as CONSTRAINED — they're polled at 60s TTL since they only
change by explicit admin action (`sysctl -w` or equivalent). This is distinct
from POLLED counters that change continuously.

## What this project does with sysctl

### 1. Direct kernel access (no cgo)

We call `sysctl(3)` and `sysctlbyname(3)` via ARM64 assembly trampolines into
`/usr/lib/libSystem.B.dylib`, the same way Go's runtime calls libc functions.
No cgo, no external dependencies, no build complexity.

### 2. MIB caching for speed

At startup, we resolve all 243 metric names to MIB arrays using `nametomib`.
Subsequent reads use the integer `sysctl()` path, which is ~3x faster than
`sysctlbyname()` because it skips name resolution.

### 3. Access pattern classification

We classify every metric's mutability and recommended poll frequency:

- **Kernel access pattern**: can the kernel change this value? (STATIC vs
  CONSTRAINED vs DYNAMIC)
- **Recommended access pattern**: how often should a monitoring client poll?
  (STATIC, POLLED@10s, POLLED@60s, CONSTRAINED@60s)

These classifications live in a textproto file generated from the registry and
verified against live kernel behavior.

### 4. Background poller

A single goroutine refreshes POLLED and CONSTRAINED metrics at their TTL
intervals. STATIC metrics are read once and frozen. DYNAMIC metrics (none
currently, but the pattern exists) would be read live on every request.

### 5. Delta streaming

The Subscribe RPC streams raw sysctl bytes to clients. After the initial
snapshot, only values that changed since the last push are sent. Combined with
the access pattern classifications, this means:

- STATIC metrics are sent once (first push) and never again
- POLLED metrics are re-read at their TTL and sent only if changed
- The wire format after the first push is exclusively small fixed-size integers

### 6. gRPC API

Typed protobuf messages with `oneof` value types (string, uint64, int64,
uint32, int32, raw bytes, struct). The kernel registry (all metrics + access
patterns) is exposed to clients so they know what's available and how fresh
each metric is.

## Platform notes

### macOS only (for now)

The assembly trampolines, MIB cache, and metric registry are all
Darwin/ARM64-specific. The proto definitions and gRPC server structure are
platform-independent. Adding Linux support would mean implementing a new
reader backend (probably reading from `/proc` and `/sys` instead of sysctl)
and a new metric registry.

### Apple Silicon only (for now)

The assembly is ARM64. Adding x86_64 macOS support would require x86 assembly
trampolines and handling the type width differences (4-byte vs 8-byte values
for the same metrics).

### Security / SIP

All metrics we read are available to unprivileged processes. Some sysctl
metrics (not in our registry) require root or specific entitlements. We don't
write any sysctl values — the CONSTRAINED classification describes writability,
but the server is read-only.

## Further reading

- [CACHE_DESIGN.md](CACHE_DESIGN.md) — Access patterns, monotonic chain, poller
- [DELTA_DESIGN.md](DELTA_DESIGN.md) — Compact delta encoding roadmap
- [cmd/darwin-name2int/FINDINGS.md](cmd/darwin-name2int/FINDINGS.md) — Full MIB research
- XNU source: `bsd/sys/sysctl.h` for MIB constants
  ([apple-oss-distributions/xnu](https://github.com/apple-oss-distributions/xnu))
- Apple docs: [sysctl(3)](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man3/sysctl.3.html),
  [sysctlbyname(3)](https://developer.apple.com/documentation/kernel/1387446-sysctlbyname)
