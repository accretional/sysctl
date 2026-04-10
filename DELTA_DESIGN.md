# Delta Stream Optimization Design

## Current State (implemented)

Subscribe streams `MetricDelta` messages: `{string name, bytes value}`, sent only when the raw sysctl bytes change from the previous push. The server clamps the client's requested interval to the maximum recommended TTL (monotonic access chain).

**Audit finding:** All POLLED metrics are fixed-size. The 17 variable-length metrics (15 strings + 2 raw) are all STATIC recommended — they're sent once in the first snapshot and never again. Struct types (timeval=16B, loadavg=24B, swap=32B, clock=20B) are also fixed-size. This means every delta after the initial snapshot contains only fixed-size integer values.

## Phase 1: Darwin Proto Subpackage

Create `proto/darwin/` with platform-specific enums:

```protobuf
package sysctl.darwin;

enum Category {
  HW_CPU = 0;
  HW_MEMORY = 1;
  HW_CACHE = 2;
  // ... all 32 categories
}

enum Name {
  HW_NCPU = 0;
  HW_ACTIVECPU = 1;
  HW_PHYSICALCPU = 2;
  // ... all 250 metrics
  // Ideally: use the actual MIB integer the kernel uses for sysctl(),
  // so Darwin.Name IS the kernel's metric ID.
}
```

This replaces string names with compact int32 enums in wire format.

## Phase 2: MetricDescriptor

Rename `Metric` to `MetricDescriptor`. It describes the metric (name, type, category, access pattern) but does **not** carry a value. This is the registry/metadata type.

Remove `MetricInfo` (redundant with `KernelMetric` which already carries identity + access patterns).

```protobuf
message MetricDescriptor {
  darwin.Name id = 1;
  darwin.Category category = 2;
  string description = 3;
  ValueType type = 4;  // enum: INT32, INT64, UINT32, UINT64, TIMEVAL, LOADAVG, etc.
  AccessConfig kernel_access_pattern = 5;
  AccessConfig recommended_access_pattern = 6;
}
```

## Phase 3: Compact Delta Encoding

With fixed-size values confirmed and enum IDs in place, delta messages compress to:

```protobuf
message CompactDelta {
  darwin.Name id = 1;   // int32 enum — 1-2 bytes varint
  fixed64 value = 2;    // all values fit in 8 bytes (largest: uint64, timeval fields)
}

message CompactSubscribeResponse {
  repeated CompactDelta deltas = 1;
}
```

For struct types (timeval, loadavg, swap, clock), either:
- Emit multiple CompactDelta entries (one per struct field), or
- Use a `bytes value` field with known fixed size per type

The first approach is simpler and keeps everything as a single int.

### Wire size comparison

Current (string name + bytes value):
- `"vm.page_free_count"` = 22 bytes name + 4 bytes value + protobuf overhead ≈ 30 bytes/delta

Compact (enum id + fixed64):
- 1-2 bytes id + 8 bytes value + protobuf overhead ≈ 12 bytes/delta

**~60% reduction per delta.**

## Phase 4: MIB-Based IDs

The Darwin kernel identifies each sysctl by an integer MIB array (e.g., `kern.ostype` = `{1, 1}`). We already resolve these in the MIB cache. If `darwin.Name` enum values match the kernel's MIB encoding (or a flattened version of it), then the enum IS the kernel's metric ID — zero translation cost.

This requires a stable mapping from MIB arrays to flat enum values. The `nametomib()` function already does this; the generator can emit enum values from the resolved MIBs.

## Non-Goals

- Variable-length values in deltas: all POLLED metrics are fixed-size integers
- Computed metrics in Subscribe: these are server-side aggregates, not direct sysctl reads; use GetMetrics for those
- Backward compatibility with string-based names in Subscribe: the compact format is a new wire protocol, not an evolution of the current one
