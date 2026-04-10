# Delta Encoding Implementation Plan

This supersedes the original [DELTA_DESIGN.md](DELTA_DESIGN.md) with concrete
decisions based on the MIB mapping research
([cmd/darwin-name2int/FINDINGS.md](cmd/darwin-name2int/FINDINGS.md)).

## Current state

Subscribe streams `MetricDelta {string name, bytes value}` — changed values
only, raw sysctl bytes, with string metric names on the wire. This works but
is inefficient: a typical delta like `vm.page_free_count` is ~30 bytes
(22-byte name + 4-byte value + protobuf framing). Most of that is the name.

## What we now know

From the darwin-name2int research:

1. **All 243 non-computed metrics resolve to unique MIB arrays.** No collisions.
2. **All metrics that change at runtime are ≤8 bytes** (single integers). The
   16 metrics larger than uint64 (strings, structs, raw arrays) are all STATIC —
   they appear once in the initial snapshot and never in subsequent deltas.
3. **MIB arrays fit in 40 bits** but are hierarchical (2-5 levels, max value
   327 per level). Packing them into a flat integer is possible but pointless —
   a sequential enum is simpler and equally compact on the wire.
4. **MIB values may change between OS versions.** The well-known CTL_* constants
   are stable, but new-style OIDs (≥100) are assigned by kernel registration
   order. Any ID scheme must be versioned.
5. **hw.pagesize type widening** shows that even "stable" metrics can change
   their OID and type width across architectures. Runtime resolution is safer
   than hardcoded constants.

## Design decisions

### Use sequential enum, not MIB-based IDs

The original DELTA_DESIGN.md proposed using kernel MIB integers as enum values.
This would make the enum "free" (no lookup table needed to call `sysctl`).

**We're not doing this.** Reasons:

- MIBs are variable-length arrays (2-5 ints), not single ints. Packing them
  requires an encoding scheme that buys nothing over a lookup table.
- MIBs can change between OS versions. A sequential enum versioned by OS
  release is more predictable.
- We already resolve names to MIBs at startup via `nametomib`. The MIB cache
  is the lookup table — the enum just needs to index into it.
- Sequential enum values 0..242 encode as 1-2 byte varints. MIB-packed uint64s
  would also be 1-5 bytes. No wire savings.

### Two-phase Subscribe: descriptor table + compact deltas

The core optimization: send metric metadata once, then reference by integer ID.

**Phase 1 response** (first push on the stream):

```protobuf
message SubscribeDescriptor {
  uint32 id = 1;          // 0..N, sequential, assigned by server for this stream
  string name = 2;        // "vm.page_free_count"
  string value_type = 3;  // "uint32", "int64", etc.
  uint32 size = 4;        // byte width of value (4 or 8 for all post-snapshot deltas)
}

message SubscribeInit {
  repeated SubscribeDescriptor descriptors = 1;
  repeated string errors = 2;            // unknown names, computed metrics, etc.
  int64 min_interval_ns = 3;             // server's minimum interval
}
```

**Phase 2+ responses** (subsequent pushes):

```protobuf
message CompactDelta {
  uint32 id = 1;          // references SubscribeDescriptor.id
  fixed64 value = 2;      // raw value, zero-extended to 8 bytes
}

message CompactSubscribeResponse {
  int64 timestamp_ns = 1;
  repeated CompactDelta deltas = 2;
}
```

The client uses the descriptor table from the init message to interpret IDs
and values. No strings on the wire after the first push.

### Don't create a darwin proto subpackage

The original plan proposed `proto/darwin/` with platform-specific `Name` and
`Category` enums. This adds complexity for marginal benefit:

- The sequential IDs are stream-scoped (assigned per Subscribe call), not
  global. Different streams can have different ID assignments.
- Categories are only useful for `ListKnownMetrics` filtering, which already
  works fine with strings. Encoding them as enums saves nothing in the
  streaming path.
- A global `Darwin.Name` enum would need to be regenerated for every OS
  version. Stream-scoped IDs avoid this entirely.

If we later want a global enum for non-streaming RPCs, we can add it
independently. It's not on the critical path.

### Don't rename Metric to MetricDescriptor yet

The original plan proposed renaming `Metric` (which carries a value) to
`MetricDescriptor` (metadata only). This is a good cleanup but it's a
breaking change to the existing RPCs (GetMetrics, GetMetricsByCategory).

Deferred until we have a reason to break the API (e.g., v2 proto package).
The Subscribe path uses its own `SubscribeDescriptor` message and doesn't
touch `Metric`.

## Wire size analysis

Current delta for `vm.page_free_count` (a 4-byte uint32):

```
MetricDelta:
  name:  "vm.page_free_count"  = 22 bytes + 2 bytes framing = 24 bytes
  value: <4 bytes raw>         = 4 bytes + 2 bytes framing  = 6 bytes
  total:                                                     ≈ 30 bytes
```

Compact delta for the same metric:

```
CompactDelta:
  id:    <varint, 0..242>      = 1-2 bytes + 1 byte tag     = 2-3 bytes
  value: <fixed64>             = 8 bytes + 1 byte tag        = 9 bytes
  total:                                                     ≈ 11-12 bytes
```

**~60% reduction per delta**, matching the original DELTA_DESIGN.md estimate.

For a typical Subscribe push with 20 changed metrics:
- Current: ~600 bytes
- Compact: ~240 bytes

The initial descriptor table is a one-time cost of ~40 bytes per metric
(name + type + id). For 168 POLLED+CONSTRAINED metrics, that's ~6.7 KB once.

## Implementation plan

### Step 1: Add CompactSubscribe RPC

New RPC alongside the existing Subscribe (no breaking changes):

```protobuf
rpc CompactSubscribe(CompactSubscribeRequest) returns (stream CompactSubscribeResponse);
```

The first response is a `CompactSubscribeResponse` with an `init` field
containing the descriptor table. Subsequent responses contain only
`CompactDelta` entries.

### Step 2: Server implementation

- On stream open: validate names, assign sequential IDs 0..N, build
  descriptor table, send init message with initial values.
- On tick: read via MIB cache, compare raw bytes, for changed values
  zero-extend to uint64 and emit `CompactDelta{id, value}`.
- The existing string-based Subscribe stays unchanged.

### Step 3: Client support

- Client reads the init message, builds `id → name` and `id → type` maps.
- For each `CompactDelta`, looks up the name and type, interprets the
  fixed64 according to type (truncate to 4 bytes for uint32/int32).
- Client CLI (`cmd/client/`) gets a `-compact` flag for the new path.

### Step 4: Validation

- Test that CompactSubscribe produces identical values to Subscribe for
  the same metrics and interval.
- Benchmark wire sizes: log bytes per response for both paths.
- Verify that STATIC metrics are excluded from post-init deltas (same as
  current Subscribe behavior).

## What about the initial snapshot?

The current Subscribe sends STATIC metrics (strings, structs, raw arrays)
in the first push as `MetricDelta{name, bytes}`. CompactSubscribe can do
the same via the descriptor table: include STATIC values in the init message
as a separate `repeated bytes initial_values` field, indexed by descriptor ID.

Alternatively, don't include STATIC values in CompactSubscribe at all — the
client can fetch them once via GetMetrics. The compact stream is optimized
for the steady-state delta path, not the one-time snapshot.

## Non-goals

- **Variable-length values in compact deltas**: all post-snapshot values are
  ≤8 bytes. Confirmed by audit.
- **Computed metrics in Subscribe**: server-side aggregates, not raw sysctl
  reads. Use GetMetrics.
- **Cross-version ID stability**: IDs are stream-scoped. No global registry
  needed.
- **Backward compatibility with string Subscribe**: it stays as-is. The
  compact path is a separate RPC.

## Relationship to DELTA_DESIGN.md

| DELTA_DESIGN.md | This plan | Status |
|-----------------|-----------|--------|
| Phase 1: Darwin proto subpackage | **Skipped** — stream-scoped IDs avoid global enums | Not needed |
| Phase 2: Rename Metric → MetricDescriptor | **Deferred** — separate from Subscribe optimization | Future cleanup |
| Phase 3: CompactDelta encoding | **Adopted** with modifications — `uint32 id` + `fixed64 value` | Step 1-2 |
| Phase 4: MIB-based IDs | **Rejected** — sequential enum is simpler, equally compact | Decided against |

The original design's core insight (compact integer deltas replace string
names + variable bytes) is correct. The implementation details changed based
on what we learned from the MIB research.
