# macOS Sysctl Inventory — Triage & Implementation Guide

## Overview

macOS exposes ~1,695 sysctl keys. We implement 250 across 32 categories (including 7 server-side computed aggregates). Each metric has a kernel access pattern classification in `internal/metrics/darwin/24.6.0.textproto` — see `CACHE_DESIGN.md` for the access pattern model. This document records triage findings: what to implement, what to skip, and the reasoning behind each decision.

## Batch / Aggregate API Findings

### There Is No Batch Read API

macOS sysctl has **no** wildcard or batch read mechanism. Each key must be read individually via `sysctlbyname()` or the MIB-based `sysctl()`.

### Tree Traversal: CTL_SYSCTL_NEXT (OID [0,2,...])

The kernel provides `CTL_SYSCTL_NEXT` (MIB `{0, 2}`) for enumerating the entire sysctl tree. This is how `sysctl -a` works internally. It returns the next valid MIB after the one you pass. This lets you discover all keys but still requires individual reads for values.

### MIB Pre-Resolution: sysctlnametomib

`sysctlnametomib()` resolves a dotted name (e.g., `vm.page_free_count`) to its integer MIB array once. Subsequent reads via `sysctl()` with the MIB are ~3x faster than `sysctlbyname()` because they skip name resolution. **Implemented**: The server pre-resolves all 243 sysctl MIBs at startup via `MIBCache.Warm()` and uses MIB-based reads for all requests.

### Computed Aggregate Metrics (Server-Side)

Since no kernel-side aggregation exists, the server computes these from individual reads (implemented as `computed.*` metrics):

| Aggregate Metric | Formula | Source Keys |
|---|---|---|
| Memory utilization % | `(total_pages - free_pages) / total_pages * 100` | `vm.pages`, `vm.page_free_count` |
| Compression ratio | `input_bytes / compressed_bytes` | `vm.compressor_input_bytes`, `vm.compressor_compressed_bytes` |
| Swap utilization % | `used / total * 100` | `vm.swapusage` (struct fields) |
| Compressor pressure | `bytes_used / pool_size * 100` | `vm.compressor_bytes_used`, `vm.compressor_pool_size` |
| Total connections | `tcp + udp + unix` | `net.inet.tcp.pcbcount`, `net.inet.udp.pcbcount`, `net.local.pcbcount` |
| Uptime (seconds) | `now - boottime` | `kern.boottime` |
| VFS reclamation rate | `recycled / total` | `vfs.vnstats.num_recycledvnodes`, `vfs.vnstats.num_vnodes` |

## Tier Classification

All ~1,695 keys are classified into four tiers based on empirical dynamic/static analysis (two snapshots 10 seconds apart, ~80 keys changed).

### Tier 1 — Active Polling (10-30s)

Dynamic counters that change frequently under normal load. These are the core telemetry metrics.

**Currently implemented (correct):**
- `vm.page_free_count`, `vm.page_speculative_count`, `vm.page_cleaned_count` — page state
- `vm.page_pageable_internal_count`, `vm.page_pageable_external_count` — memory classification
- `vm.page_purgeable_count`, `vm.page_reusable_count` — reclaimable pages
- `vm.compressor_input_bytes`, `vm.compressor_compressed_bytes`, `vm.compressor_bytes_used` — compressor activity
- `vm.pages_grabbed`, `vm.copied_on_read` — allocation events
- `vm.pageout_*` counters — pageout activity
- `vm.cs_blob_count`, `vm.cs_blob_size` — code signing activity
- `net.inet.tcp.pcbcount`, `net.inet.udp.pcbcount`, `net.local.pcbcount` — connection counts
- `net.inet.tcp.sack_globalholes` — SACK state
- `net.inet.tcp.cubic_sockets` — congestion control distribution
- `kern.num_files`, `kern.num_vnodes`, `kern.free_vnodes`, `kern.num_recycledvnodes` — VFS/FD counters
- `vfs.vnstats.*` — vnode statistics
- `kern.memorystatus_level` — memory pressure percentage
- `kern.monotonicclock_usecs` — monotonic time

**Added (confirmed dynamic, byte sizes verified):**
- `net.soflow.count` — socket flow count (int64, 8 bytes)
- `security.mac.vnode_label_count` — vnode MAC labels (int32, 4 bytes)
- `vfs.vnstats.num_dead_vnodes` — dead vnodes (int64, 8 bytes)
- `security.mac.asp.stats.exec_hook_count` — exec hook invocations (int64, 8 bytes)
- `security.mac.asp.stats.library_hook_count` — library hook invocations (int64, 8 bytes)
- `security.mac.asp.stats.exec_hook_work_time` — exec hook time (int64, 8 bytes)
- `security.mac.asp.stats.library_hook_time` — library hook time (int64, 8 bytes)
- `net.inet.mptcp.pcbcount` — MPTCP connections (int32, 4 bytes)

### Tier 2 — Periodic Sampling (60-300s)

Slow-changing configuration or low-frequency counters.

**Currently implemented (correct):**
- `vm.memory_pressure`, `vm.page_free_wanted`, `vm.vm_page_free_target` — pressure thresholds
- `vm.compressor_mode`, `vm.compressor_is_active`, `vm.compressor_available` — compressor state
- `vm.swap_enabled`, `vm.swapusage` — swap configuration
- `vm.global_user_wire_limit`, `vm.user_wire_limit` — wire limits
- `vm.add_wire_count_over_*_limit` — wire limit violations (slow accumulators)
- `kern.memorystatus_purge_on_*` — purge thresholds
- `machdep.user_idle_level` — idle state
- `machdep.time_since_reset`, `machdep.wake_abstime` — timing
- `kern.hibernatecount` — hibernate events

### Tier 3 — Startup Inventory (recommended: read once / poll infrequently)

Hardware immutables and rarely-changed tunables. Recommended as STATIC or POLLED@60s.

Note: many of these (limits, IPC, sched, net config) are writable tunables — the kernel CAN change them at runtime — so their `kernel_access_pattern` is DYNAMIC. But in practice they rarely change, so `recommended_access_pattern` is STATIC or POLLED@60s.

**Implemented:**
- All `hw.*` keys (CPU topology, memory sizes, cache hierarchy, perflevel, ARM features) — truly STATIC
- `kern.os*`, `kern.version`, `kern.uuid`, `kern.bootuuid` — STATIC (immutable per boot)
- `kern.hostname` — DYNAMIC at kernel level (writable), recommended STATIC
- `kern.maxproc`, `kern.maxfiles`, `kern.maxvnodes`, etc. — writable tunables, recommended POLLED@60s
- `kern.boottime`, `kern.clockrate` — STATIC
- `kern.hv_support`, `kern.secure_kernel`, `kern.safeboot` — STATIC (boot-time flags)
- `machdep.cpu.*` — STATIC (CPU identification)
- `net.inet.tcp.*` config (mssdflt, keepidle, sendspace, etc.) — writable, recommended POLLED@60s
- `debug.*`, `kpc.*`, `iogpu.*` — writable, recommended POLLED@60s

### Tier 4 — Skip / Noise

Keys that should NOT be implemented because they return no useful data, are static limits masquerading as counters, or are irrelevant to performance telemetry.

**Registry entries demoted/fixed (all done):**

| Key | Problem | Resolution |
|---|---|---|
| `kern.num_tasks` | Returns 4096 always — static task limit | Moved to `kern.limits`, description updated |
| `kern.num_threads` | Returns 20480 always — static thread limit | Moved to `kern.limits`, description updated |
| `kern.num_taskthreads` | Returns 4096 always — static limit | Moved to `kern.limits`, description updated |
| `vm.wk_*` (4 keys) | Always 0 on Apple Silicon M4 | Removed — kernel uses different compressor |
| `vm.lz4_*` (4 keys) | Always 0 on Apple Silicon M4 | Removed |
| `vm.loadavg` | Miscategorized under `vm.swap` | Moved to `vm.misc` |

**Entire subtrees to skip (~1,400 keys):**
- `kern.tty.*` — terminal driver internals (irrelevant to server telemetry)
- `kern.ipc.extbk*`, `kern.ipc.mb_*` — deep mbuf internals
- `kern.dtrace.*` — DTrace configuration (not telemetry)
- `kern.kdbg.*` — kernel debug tracing
- `net.key.*` — IPsec key management internals
- `net.inet.ipsec.*`, `net.inet6.ipsec6.*` — IPsec counters (usually all zero)
- `net.route.*` — routing table metadata
- `net.link.*` — link-layer interface details
- `security.mac.endpointsecurity.*` — ES framework internals
- `hw.optional.arm.FEAT_*` beyond the 17 we already have — there are ~50 more feature flags, most unrelated to perf
- `kern.skywalk.*` — networking stack internals (very deep, ~200 keys)
- `kern.timer.*`, `kern.pmgt.*` — power management internals
- `kern.entropy.*` — entropy pool statistics
- `vfs.generic.hfs.*`, `vfs.generic.nfs.*` — filesystem-specific config
- All `sysctl.*` (OID 0) — metadata about the sysctl system itself

## ARM64 Type Quirks (Apple Silicon)

These were discovered empirically and are critical for correct implementation:

| Key Pattern | Expected (x86) | Actual (ARM64) | Notes |
|---|---|---|---|
| `hw.pagesize`, `hw.cachelinesize` | 4 bytes (int32) | 8 bytes (int64) | ARM64 returns 64-bit values |
| `hw.l1icachesize`, `hw.l1dcachesize`, `hw.l2cachesize` | 4 bytes | 8 bytes (int64) | Same |
| `hw.tbfrequency` | 4 bytes | 8 bytes (int64) | Same |
| `hw.cpufrequency`, `hw.busfrequency` | present | ENOENT | Don't exist on Apple Silicon |
| `vm.pageout_inactive_clean/used`, `vm.pageout_protect_realtime` | 8 bytes | 4 bytes (int32) | Opposite direction! |
| `vm.pageout_inactive_dirty_*`, `vm.pageout_freed_*`, `vm.pageout_speculative_clean` | 4 bytes | 8 bytes (int64) | 64-bit counters |
| `kern.num_recycledvnodes` | 4 bytes | 8 bytes (int64) | Counter grew to 64-bit |

## Category Summary

| Category | Keys | Tier | Status |
|---|---|---|---|
| hw.cpu | 14 | 3 | Implemented |
| hw.memory | 4 | 3 | Implemented |
| hw.cache | 7 | 3 | Implemented |
| hw.perflevel | 18 | 3 | Implemented |
| hw.arm | 17 | 3 | Implemented |
| hw.misc | 0 | — | Empty (removed non-existent keys) |
| kern.identity | 11 | 3 | Implemented |
| kern.process | 4* | 1 | Implemented (*3 moved to kern.limits) |
| kern.memory | 4 | 1-2 | Implemented |
| kern.sched | 6 | 3 | Implemented |
| kern.time | 9 | 1-3 | Implemented |
| kern.limits | 14* | 3 | Implemented (*+3 from kern.process) |
| kern.ipc | 3 | 3 | Implemented |
| kern.misc | 9 | 3 | Implemented |
| vm.pressure | 3 | 1 | Implemented |
| vm.pages | 16 | 1 | Implemented |
| vm.compressor | 7* | 1-2 | Implemented (*-8 removed WKdm/LZ4 zeros) |
| vm.pageout | 13 | 1 | Implemented |
| vm.swap | 3* | 2-3 | Implemented (*-1 vm.loadavg moved) |
| vm.wire | 5 | 2 | Implemented |
| vm.misc | 8* | 1-2 | Implemented (*+1 vm.loadavg moved here) |
| machdep | 11 | 2-3 | Implemented |
| net.tcp | 18 | 1-3 | Implemented |
| net.udp | 4 | 1-3 | Implemented |
| net.ip | 3 | 3 | Implemented |
| net.misc | 3 | 1 | Implemented |
| vfs | 7 | 1-3 | Implemented |
| debug | 3 | 3 | Implemented |
| kpc | 2 | 3 | Implemented |
| iogpu | 3 | 2-3 | Implemented |
| security | 5 | 1 | Implemented |
| computed | 7 | — | Implemented (server-side aggregates) |

## Next Steps

See `CACHE_DESIGN.md` for the full roadmap. Summary:

- **Phase 5**: Background polling loop — enforce STATIC/POLLED/DYNAMIC access patterns with real caching
- **Phase 6**: Service-level config overrides (operator `--config` textproto)
- **Phase 7**: Client access masks (monotonic decrease in access frequency)
- **Phase 8**: Export / observability (Prometheus, streaming gRPC, TSDB)

## Data Sources

- Dynamic analysis: Two snapshots 10 seconds apart, diffed to identify ~80 changing keys
- Type verification: Manual `sysctl` reads checking byte lengths on M4 Apple Silicon (macOS 15.x)
- Static limit verification: `kern.num_tasks` confirmed returning 4096 across multiple reads
- WKdm/LZ4 verification: All 8 counters confirmed zero across multiple reads on M4
