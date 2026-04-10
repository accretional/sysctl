# darwin-name2int: MIB Mapping Findings

## What This Tool Does

Maps all 243 non-computed sysctl metric names to their integer MIB (Management
Information Base) arrays using the kernel's `nametomib()` facility (sysctl
`{0,3}`). Cross-references results against XNU header constants and validates
by comparing reads via `sysctl(mib)` vs `sysctlbyname(name)`.

## Key Findings

### 1. All 243 metrics resolve to unique MIB arrays

Every non-computed metric resolves successfully. No two metrics share a MIB.
This means MIB arrays are a viable unique identifier for metrics.

### 2. MIB structure is hierarchical (2-5 levels deep)

| Depth | Count | Examples |
|-------|-------|---------|
| 2     | 161   | `hw.ncpu` → `{6,3}`, `kern.ostype` → `{1,1}` |
| 3     | 35    | `kern.ipc.maxsockbuf` → `{1,32,1}`, `vfs.vnstats.num_vnodes` → `{3,101,114}` |
| 4     | 43    | `net.inet.tcp.mssdflt` → `{4,2,6,3}`, `hw.perflevel0.physicalcpu` → `{6,131,100,104}` |
| 5     | 4     | `security.mac.asp.stats.*` → `{103,101,124,104,*}` |

### 3. Top-level MIB[0] values: mostly well-known, some extended

| MIB[0] | XNU Constant | Count | Notes |
|--------|-------------|-------|-------|
| 1      | CTL_KERN    | 60    | Kernel subsystem |
| 2      | CTL_VM      | 64    | Virtual memory |
| 3      | CTL_VFS     | 7     | Filesystem |
| 4      | CTL_NET     | 28    | Networking |
| 5      | CTL_DEBUG   | 3     | Debug |
| 6      | CTL_HW      | 60    | Hardware |
| 7      | CTL_MACHDEP | 11    | Machine-dependent |
| 101    | *(extended)* | 1    | kperf (performance counters) |
| 102    | *(extended)* | 1    | kpc (perf counter config) |
| 103    | *(extended)* | 5    | security.mac |
| 104    | *(extended)* | 3    | iogpu |

233 metrics (96%) use well-known CTL_* constants. 10 metrics (4%) use
extended OID space (101-104) for subsystems added after the original sysctl
header.

### 4. XNU header constants: 21 verified, 1 documented divergence

21 metrics match their XNU header constants exactly. One known divergence:

**`hw.pagesize`**: XNU header defines `HW_PAGESIZE = 7`, but `sysctlbyname`
resolves to `{6, 115}` on modern macOS ARM64. The kernel registers `hw.pagesize`
as a new-style OID rather than using the legacy constant. The value reads
correctly (16384 on Apple Silicon).

### 5. Network MIB hierarchy follows protocol families

Network metrics follow `CTL_NET / PF_* / IPPROTO_* / sub-ID`:

```
net.inet.tcp.mssdflt  → {4, 2, 6, 3}     = CTL_NET / PF_INET / IPPROTO_TCP / TCPCTL_MSSDFLT
net.inet.udp.maxdgram → {4, 2, 17, 3}    = CTL_NET / PF_INET / IPPROTO_UDP / UDPCTL_MAXDGRAM
net.inet.ip.forwarding → {4, 2, 0, 1}    = CTL_NET / PF_INET / IPPROTO_IP / 1
net.local.pcbcount    → {4, 1, 102}       = CTL_NET / PF_LOCAL / 102
net.inet.mptcp.*      → {4, 2, 256, ...}  = CTL_NET / PF_INET / IPPROTO_MPTCP / ...
```

### 6. Flattenability to uint64: 239/243 succeed

Using 16-bits-per-level encoding (max 4 levels), 239 of 243 metrics flatten to
unique uint64 values with zero collisions. The 4 unflattenable metrics are
`security.mac.asp.stats.*` (depth 5).

**Max values per MIB level:**
| Level | Max Value | Bits Needed |
|-------|-----------|-------------|
| 0     | 104       | 7           |
| 1     | 327       | 9           |
| 2     | 256       | 9           |
| 3     | 202       | 8           |
| 4     | 108       | 7           |

Total bits needed: 7+9+9+8+7 = 40 bits. All MIB arrays fit in a uint64 with
room to spare. A 40-bit packed encoding would cover all 243 metrics.

### 7. Volatile metrics: only 4 change between consecutive reads

| Metric | Size | Notes |
|--------|------|-------|
| kern.monotonicclock_usecs | 8 bytes | Microsecond counter |
| vm.page_pageable_internal_count | 4 bytes | Memory page counter |
| vm.pageout_inactive_dirty_internal | 8 bytes | Pageout counter |
| machdep.time_since_reset | 8 bytes | Time counter |

All other 239 metrics read identically via `sysctl(mib)` and `sysctlbyname(name)`
in a single test pass. The 4 volatile ones simply changed between the two reads —
not a correctness issue.

## Implications for DELTA_DESIGN.md

### Phase 4: MIB-Based IDs — Feasible

The original design proposed using kernel MIB values as the `Darwin.Name` enum
values. The findings confirm this is viable:

1. **All MIBs are unique** — no collisions.
2. **Most fit in 32 bits** — max 4 levels × 9 bits = 36 bits. Even with
   5-level security metrics, 40 bits suffices.
3. **Stable across boots** — MIB values are determined by kernel OID
   registration order, which is deterministic within a kernel version.

**Recommended encoding for `Darwin.Name`:**

```
// Pack MIB array into a single uint64:
// [depth:4][level0:9][level1:9][level2:9][level3:9][level4:7]  = 47 bits
//
// Or simpler: sequential enum 0..242, with a lookup table to MIB arrays.
// The sequential approach is simpler and equally compact for wire encoding
// (varint for values <128 = 1-2 bytes).
```

The sequential enum is simpler and just as wire-efficient. The MIB-based
encoding is clever but unnecessary unless we need to construct MIB arrays
from the enum value without a lookup table.

### However: MIBs May Change Between OS Versions

MIB integers come from OID registration order in the kernel. While the
well-known constants (CTL_KERN=1, KERN_OSTYPE=1, etc.) are stable across
versions, the extended OIDs (>100) and dynamic registrations could change
between macOS versions. The `Darwin.Name` enum should be versioned per
OS release, not assumed stable across versions.

## Files

- `main.go` — CLI tool for MIB resolution and analysis
- `name2int_test.go` — Comprehensive validation tests
- `mib_mappings.json` — Full MIB mappings for all 243 metrics (generated)
- `FINDINGS.md` — This document
