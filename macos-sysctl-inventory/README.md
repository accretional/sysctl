# macOS Sysctl Inventory — Triage & Implementation Guide

## Overview

macOS exposes ~1,695 sysctl keys. We currently implement 243 across 31 categories. This document records our triage findings: what to implement, what to skip, batch/aggregate API capabilities, and the reasoning behind each decision.

## Batch / Aggregate API Findings

### There Is No Batch Read API

macOS sysctl has **no** wildcard or batch read mechanism. Each key must be read individually via `sysctlbyname()` or the MIB-based `sysctl()`.

### Tree Traversal: CTL_SYSCTL_NEXT (OID [0,2,...])

The kernel provides `CTL_SYSCTL_NEXT` (MIB `{0, 2}`) for enumerating the entire sysctl tree. This is how `sysctl -a` works internally. It returns the next valid MIB after the one you pass. This lets you discover all keys but still requires individual reads for values.

### MIB Pre-Resolution: sysctlnametomib

`sysctlnametomib()` resolves a dotted name (e.g., `vm.page_free_count`) to its integer MIB array once. Subsequent reads via `sysctl()` with the MIB are ~3x faster than `sysctlbyname()` because they skip name resolution. **Recommendation**: For tier-1 metrics polled at 10-30s intervals, pre-resolve MIBs at startup and use the MIB-based path.

### Computed Aggregate Metrics (Server-Side)

Since no kernel-side aggregation exists, the server should compute these from individual reads:

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

**Should add (confirmed dynamic, byte sizes verified):**
- `net.soflow.count` — socket flow count (int64, 8 bytes) — changes with network activity
- `security.mac.vnode_label_count` — vnode MAC labels (int32, 4 bytes) — tracks labeling activity
- `vfs.vnstats.num_dead_vnodes` — dead vnodes (int64, 8 bytes) — reclamation pressure indicator
- `security.mac.asp.stats.exec_hook_count` — exec hook invocations (int64, 8 bytes) — process launch rate proxy
- `security.mac.asp.stats.library_hook_count` — library hook invocations (int64, 8 bytes) — dylib load rate
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

### Tier 3 — Startup Inventory (Read Once)

Static hardware/software configuration. Read at server start, serve from cache.

**Currently implemented (correct):**
- All `hw.*` keys (CPU topology, memory sizes, cache hierarchy, perflevel, ARM features)
- `kern.os*`, `kern.version`, `kern.hostname`, `kern.uuid`, `kern.bootuuid` — identity
- `kern.maxproc`, `kern.maxfiles`, `kern.maxvnodes`, etc. — limits
- `kern.boottime`, `kern.sleeptime`, `kern.waketime` — boot timestamps
- `kern.clockrate` — clock info struct
- `kern.hv_support`, `kern.secure_kernel`, `kern.safeboot` — boot config
- `machdep.cpu.*` — CPU identification
- All `net.inet.tcp.*` config (mssdflt, keepidle, sendspace, etc.)
- `debug.*`, `kpc.*`, `iogpu.*`

### Tier 4 — Skip / Noise

Keys that should NOT be implemented because they return no useful data, are static limits masquerading as counters, or are irrelevant to performance telemetry.

**Registry entries to demote/fix:**

| Key | Problem | Action |
|---|---|---|
| `kern.num_tasks` | Returns 4096 always — it's the static task limit, NOT a live count | Fix description: "Task limit (not live count)" and move to `kern.limits` |
| `kern.num_threads` | Returns 20480 always — static thread limit | Fix description: "Thread limit (not live count)" and move to `kern.limits` |
| `kern.num_taskthreads` | Returns 4096 always — static limit | Fix description: "Threads-per-task limit (not live count)" and move to `kern.limits` |
| `vm.wk_compressions` | Always 0 on Apple Silicon M4 | Remove — kernel uses different compressor |
| `vm.wk_compressed_bytes_total` | Always 0 | Remove |
| `vm.wk_decompressions` | Always 0 | Remove |
| `vm.wk_decompressed_bytes` | Always 0 | Remove |
| `vm.lz4_compressions` | Always 0 on Apple Silicon M4 | Remove |
| `vm.lz4_compressed_bytes` | Always 0 | Remove |
| `vm.lz4_decompressions` | Always 0 | Remove |
| `vm.lz4_compression_failures` | Always 0 | Remove |
| `vm.loadavg` | Miscategorized under `vm.swap` | Move to `kern.time` or `vm.misc` — it's not swap-related |

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

### Phase 5: Background Polling Loop
The server currently reads sysctl values on-demand per gRPC request (using MIB-cached reads for speed). The next step is a background polling loop with tiered intervals:
- **Tier 1** (10-30s): Dynamic counters (VM pages, compressor, pageout, connections, VFS, security hooks) — serve from a hot snapshot cache
- **Tier 2** (60-300s): Slow-changing config (pressure thresholds, compressor state, wire limits, idle level)
- **Tier 3** (read once at startup): Static hardware/software config (all hw.*, kern identity, limits, net config)

This would decouple read latency from gRPC response time and enable consistent point-in-time snapshots.

### Phase 6: Export / Observability
- Prometheus metrics endpoint
- Streaming gRPC (server-push on change)
- TSDB push integration

## Data Sources

- Dynamic analysis: Two snapshots 10 seconds apart, diffed to identify ~80 changing keys
- Type verification: Manual `sysctl` reads checking byte lengths on M4 Apple Silicon (macOS 15.x)
- Static limit verification: `kern.num_tasks` confirmed returning 4096 across multiple reads
- WKdm/LZ4 verification: All 8 counters confirmed zero across multiple reads on M4
