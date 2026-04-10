# FINDINGS_2: Full macOS Sysctl Landscape

This document extends the MIB mapping research
([cmd/darwin-name2int/FINDINGS.md](cmd/darwin-name2int/FINDINGS.md)) to
cover **every sysctl on the system**, not just our 250 registered metrics.
It exists because the Darwin encoding decision (sequential enum vs MIB-packed
IDs) can only be made with a complete picture of the namespace.

## The numbers

| | Count |
|---|---|
| Total sysctl entries on this machine | 1,695 |
| We register | 250 (243 non-computed + 7 computed) |
| We don't register | 1,445 |
| MIB-resolvable | 1,675 (all except user.*) |
| Not resolvable | 20 (all user.* — POSIX constants, special kernel handler) |

**Our 250 metrics cover 15% of the sysctl namespace.**

## Coverage by top-level namespace

| Namespace | MIB[0] | System total | Ours | Missing | Coverage |
|-----------|--------|-------------|------|---------|----------|
| kern | 1 (CTL_KERN) | 425 | 60 | 365 | 14% |
| vm | 2 (CTL_VM) | 256 | 64 | 192 | 25% |
| vfs | 3 (CTL_VFS) | 113 | 7 | 106 | 6% |
| net | 4 (CTL_NET) | 598 | 28 | 570 | 5% |
| debug | 5 (CTL_DEBUG) | 56 | 3 | 53 | 5% |
| hw | 6 (CTL_HW) | 124 | 60 | 64 | 48% |
| machdep | 7 (CTL_MACHDEP) | 18 | 11 | 61% |
| user | 8 (CTL_USER) | 20 | 0 | 20 | 0% |
| ktrace | 100 | 5 | 0 | 5 | 0% |
| kperf | 101 | 6 | 1 | 5 | 17% |
| kpc | 102 | 1 | 1 | 0 | 100% |
| security | 103 | 68 | 5 | 63 | 7% |
| iogpu | 104 | 5 | 3 | 2 | 60% |

The biggest gaps are **net** (570 missing) and **kern** (365 missing). These
aren't exotic — they include tunables, counters, and protocol stats that any
serious monitoring tool would want.

## What we're missing: the 30 largest gaps

| Namespace | Missing | What's in it |
|-----------|---------|-------------|
| net.inet | 215 | TCP stats (cwnd, rto, retransmits), UDP stats, IP stats, IPSEC, raw IP, ICMP, IGMP — the bulk of network telemetry |
| net.link | 126 | Link-layer stats, interface management, bonding, wake-on-LAN, bridge, link heuristics |
| kern.prng | 99 | PRNG pool stats per entropy source (pools 0-9 × {sample_count, drain_count, max_sample_count}) |
| net.inet6 | 94 | IPv6 protocol stats, ICMPv6, ND, scope handling |
| vfs.generic | 89 | APFS stats, NFS stats, mount table, filesystem-level config |
| security.mac | 57 | MAC framework, sandbox stats, AMFI, proc hooks, vnode labels |
| hw.optional | 57 | Feature flags beyond ARM — FP, SIMD, crypto, atomics, memory model caps |
| kern.skywalk | 38 | Network stack internals (Apple's userspace networking framework), per-interface flowswitch stats |
| net.necp | 35 | Network Extension Control Policy — rules, sessions, clients |
| kern.ipc | 32 | IPC tunables beyond maxsockbuf/somaxconn — socket buffers, mbuf, pipe, message queue config |
| net.key | 14 | IPSec key management (PF_KEY) stats |
| net.cfil | 12 | Content filter stats (for network extensions) |
| kern.timer | 12 | Timer coalescing, deadline tracking, longterm timer queue |
| vfs.vnstats | 12 | Vnode stats we missed (lookups, reclaims, async I/O) |
| kern.entropy | 11 | Entropy estimation, health tests, filter stats |
| net.classq | 11 | Packet scheduler class queue stats |
| kern.sysv | 10 | System V IPC limits (shmmax, semmax, msgmax, etc.) |
| kern.dtrace | 9 | DTrace config (buffer size, speculations, helpers) |
| kern.crypto | 8 | Crypto framework stats |

**None of these require root access.** They are all readable by unprivileged
processes. The 20 `user.*` POSIX constants are the only entries that can't
be MIB-resolved (they use a special kernel handler, not the OID tree), but
they can still be read via `sysctlbyname`.

## Value size distribution (all 1,695 sysctls)

| Size | Count | Notes |
|------|-------|-------|
| 0 bytes | 20 | user.* (POSIX constants, resolve via different path) |
| 1 byte | 10 | Boolean flags |
| 4 bytes | 1,214 | **71.6%** — int32, the dominant type |
| 5-7 bytes | 10 | Short strings |
| 8 bytes | 397 | **23.4%** — int64/uint64 |
| 9-104 bytes | 38 | Strings, small structs |
| 140-306 bytes | 3 | Large structs (stats blobs) |

**95% of all sysctls are 4 or 8 bytes.** The fixed64 encoding from
DELTA_PLAN.md is not just viable for our curated 250 — it covers nearly the
entire sysctl namespace.

The 3 large blobs (140, 156, 306 bytes) are aggregate stats structures
(TCP stats, IP stats) that pack dozens of counters into a single sysctl.
These would need special handling (parse and decompose) or be served as
opaque blobs.

## MIB structure (all 1,675 resolvable sysctls)

### Depth distribution

| Depth | Count | % | Examples |
|-------|-------|---|---------|
| 2 | 544 | 32% | `kern.maxproc` → {1,6} |
| 3 | 304 | 18% | `kern.ipc.maxsockbuf` → {1,32,1} |
| 4 | 607 | 36% | `net.inet.tcp.mssdflt` → {4,2,6,3} |
| 5 | 182 | 11% | `net.inet.tcp.stats` → {4,2,6,4,0} |
| 6 | 38 | 2% | `kern.skywalk.flowswitch.en0.ipfm.frag_count` → {1,113,103,112,100,101} |

Depth 6 is new — our curated 250 only went to depth 5. The depth-6 entries
are all `kern.skywalk.flowswitch.<interface>.ipfm.*` per-interface
fragmentation counters.

### Max MIB values per level

| Level | Max value | Bits needed |
|-------|-----------|-------------|
| 0 | 104 | 7 |
| 1 | 360 | 9 |
| 2 | 257 | 9 |
| 3 | 208 | 8 |
| 4 | 150 | 8 |
| 5 | 107 | 7 |

Total: 7+9+9+8+8+7 = **48 bits** to pack all 6 levels. Still fits in uint64.

### MIB uniqueness

**No collisions across all 1,675 resolvable sysctls.** Every metric has a
unique MIB array, including the 1,445 we don't register.

## Implications for the Darwin encoding decision

### The full namespace is 1,675 unique IDs, not 243

DELTA_PLAN.md designed for 243 metrics (sequential enum 0..242). If we want
the encoding to cover the full sysctl namespace — which we should, since we
may add metrics incrementally and clients shouldn't need proto changes to
subscribe to new ones — the ID space needs to accommodate ~1,700 values.

A sequential enum 0..1674 still encodes as 1-2 byte varints (values <16384
fit in 2 bytes). Wire cost is identical to the 243-metric case.

### MIB-packed encoding still works but is less attractive

A 48-bit packed MIB fits in uint64 and uniquely identifies every sysctl.
But:

- It requires knowing the packing scheme to decode.
- It doesn't self-describe (is `0x0001_0020_0001_0000` kern.ipc.maxsockbuf
  or something else? You need the depth to know where the levels end).
- The depth-6 entries push us from 40 to 48 bits. More network subtrees
  in future macOS versions could push it further.
- A sequential enum is 1-2 bytes on the wire; a packed uint64 is 5-7 bytes
  varint-encoded. **The sequential enum is 3-5x more compact on the wire.**

### Recommendation: versioned sequential registry, full namespace

1. **Assign IDs 0..N covering all 1,675 resolvable sysctls**, not just our
   curated 250. The ID → name → MIB mapping is generated per OS version.

2. **Server sends the full ID table on stream init** (CompactSubscribe).
   Client caches it for the session. ~1,675 × ~30 bytes = ~50 KB one-time
   cost.

3. **Clients subscribe by name** (strings in the request). Server maps to
   IDs internally. Client never needs to know the ID assignment — it uses
   the descriptor table from the init message.

4. **The ID table is versioned by `os_registry`** (e.g., "darwin-arm64") and
   `os_version` (e.g., "Darwin 24.6.0 / xnu-11417"). Different OS versions
   may have different ID assignments.

This makes the encoding scheme trivially simple:
- Server: `name → lookup table → sequential ID → varint on wire`
- Client: `varint on wire → sequential ID → lookup table → name`
- No packing, no bit manipulation, no depth tracking.

### What about the user.* namespace?

The 20 `user.*` entries (CTL_USER, POSIX constants) cannot be resolved via
`nametomib` — the kernel uses a special handler that doesn't register them
as OIDs. They can be read via `sysctlbyname` or by hardcoding MIBs from the
XNU header (`{8, 1}` through `{8, 20}`).

These are static POSIX configuration values (path lengths, locale limits).
They're not performance metrics. For the encoding:

- **Option A**: Exclude them from the ID table. Read via `sysctlbyname` on
  demand. (Simple, and they'd never appear in a delta stream anyway.)
- **Option B**: Assign them hardcoded IDs using the XNU constants. They're
  stable — USER_CS_PATH=1 through USER_TZNAME_MAX=20 have been unchanged
  since 4.4BSD.

Option A is fine. These aren't telemetry.

## The case against compressing this data

An LLM working on this codebase might be inclined to represent the 1,695
sysctls as a compact data structure — a trie, a hash map, a compressed
enum. **Don't.** The purpose of this inventory is to make the sysctl
namespace legible to human developers and operators.

When someone asks "what sysctls are available for network monitoring?", the
answer should be a readable table they can scan, not a compressed binary blob
they need tooling to decode. When someone asks "what MIB does
`kern.ipc.maxsockbuf` resolve to?", the answer should be in a CSV they can
grep, not buried in a generated proto enum.

The full inventory CSV (`cmd/darwin-name2int/full_inventory.csv`) is 1,695
rows. That's ~100 KB. It should be checked into the repo, human-readable,
and diffable across OS versions. The MIB mappings JSON
(`cmd/darwin-name2int/mib_mappings.json`) serves the same purpose for our
curated 243 metrics.

**Ergonomics for human readers is a feature.** The Darwin encoding should
make it *easier* to understand the mapping from names to IDs, not harder.
A sequential enum with a readable lookup table achieves this. A packed
MIB encoding does not.

## Process: how this was discovered

1. **Started with 250 curated metrics** (registry.go). Built the MIB
   resolver (`cmd/darwin-name2int/`). Found all 243 non-computed metrics
   resolve to unique MIBs. Documented in FINDINGS.md.

2. **Asked: are there metrics we can't read?** Cross-referenced against
   `sysctl -aW` for writable tunables and checked for root-only metrics.
   Found: **zero root-only metrics.** Everything in `sysctl -a` is readable
   by unprivileged processes.

3. **Asked: what's the full namespace?** Ran `sysctl -a` (1,695 entries),
   resolved every name to a MIB, cross-referenced against our registry.
   Found we cover 15% of the namespace. The remaining 85% is dominated by
   network stats (net.inet, net.inet6, net.link) and kernel internals
   (kern.prng, kern.skywalk, kern.ipc).

4. **Asked: does the full namespace change the encoding decision?**
   Yes. The full namespace has 1,675 unique MIBs, depth up to 6, max MIB
   value 360. A sequential enum 0..1674 (1-2 byte varints) is more compact
   than packed MIBs (5-7 bytes) and far simpler. The recommendation in
   DELTA_PLAN.md to use sequential IDs is reinforced.

## Files

- `cmd/darwin-name2int/full_inventory.csv` — All 1,695 sysctls with MIB, depth, size, type, registration status
- `cmd/darwin-name2int/full_inventory_test.go` — Test that generates the inventory and summary stats
- `cmd/darwin-name2int/mib_mappings.json` — MIB mappings for our 243 curated metrics
- `cmd/darwin-name2int/FINDINGS.md` — Original 243-metric research
- `FINDINGS_2.md` — This document (full namespace)
- `DELTA_PLAN.md` — Encoding plan (informed by both findings documents)
