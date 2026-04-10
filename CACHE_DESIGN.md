# Cache & Config Design

## Access Patterns

Every metric has an access pattern controlling how it's read:

| Pattern | Behavior |
|---|---|
| `STATIC` | Read once at startup, served from cache forever |
| `POLLED` | Background refresh at TTL interval, serve from cache |
| `CONSTRAINED` | Mutable via OS/sandbox/process configuration, poll at TTL |
| `DYNAMIC` | Read live on every request (no caching) |
| `DISABLED` | Never read, return error if requested |

## Proto Messages

```protobuf
enum AccessPattern { STATIC, POLLED, CONSTRAINED, DYNAMIC, DISABLED }

message AccessConfig {
  AccessPattern pattern = 1;
  google.protobuf.Duration ttl = 2;  // Used by POLLED and CONSTRAINED
}

message KernelMetric {
  string name = 1;
  string description = 5;
  string value_type = 6;
  string category = 7;
  AccessConfig kernel_access_pattern = 2;
  AccessConfig recommended_access_pattern = 3;
  string notes = 4;
}

message KernelMetricRegistry {
  string os_registry = 1;
  string os_version = 2;
  repeated KernelMetric metrics = 3;
}
```

`ListKnownMetrics` returns `KernelMetricRegistry` — the single unified view of all metrics, their identities, and their access patterns.

## Monotonic Access Pattern Chain

The system models metric services as a chain where each layer can only *decrease* access frequency, never increase it:

```
kernel (source)  →  sysctl service  →  client  →  downstream client ...
   DYNAMIC            POLLED@10s       CONSTRAINED@30s     CONSTRAINED@60s
```

**The invariant:** if a layer upstream says a metric is `STATIC`, all downstream layers must treat it as `STATIC` or `DISABLED`. If upstream says `POLLED@10s`, downstream can serve `POLLED@30s`, `CONSTRAINED@60s`, or `STATIC` (snapshot) — but never `POLLED@5s` or `DYNAMIC` (which would claim more freshness than the source provides).

Access patterns are ordered by decreasing frequency:

```
DYNAMIC > POLLED > CONSTRAINED > STATIC > DISABLED
```

A downstream layer's effective pattern must be ≤ its upstream's pattern. Within the same pattern, TTL must be ≥ upstream's TTL.

### Current behavior (implemented)

The service implements per-host background polling:

1. **STATIC metrics** are read once at startup and frozen in the poller store.
2. **POLLED metrics** are refreshed by a single background goroutine on a configurable tick interval (`--poll-interval`, default 500ms). Each POLLED metric tracks its TTL from the registry and a `nextGatherNs` timestamp. On each tick, the poller iterates all POLLED metrics and re-reads those whose `nextGatherNs ≤ now`.
3. **DYNAMIC metrics** (not STATIC or POLLED) fall through to a live MIB-cached read on every request.

`GetMetric` / `GetMetrics` check the poller store first. If a metric is in the store (STATIC or POLLED), it's served from cache. Otherwise, `readMetricLive()` does a fresh kernel read.

`ListKnownMetrics` returns the kernel registry's `recommended_access_pattern` unchanged — the patterns describe what the service actually does.

### Future: tiered access enforcement

1. **Service-level overrides**
   - `--config` flag loads a config textproto with per-metric overrides
   - Overrides clamped to monotonic invariant against kernel registry
   - Attempting to set DYNAMIC on a STATIC metric is a config error

2. **Client-level masks**
   - Clients request a "mask" — their desired access patterns
   - Service clamps each metric to `min(service_pattern, client_request)`
   - This lets a lightweight dashboard client say "I only need CONSTRAINED@60s for everything"

### Why this matters

A metric service that consumes this service (e.g., a Prometheus exporter, a fleet aggregator) is itself a link in the chain. It can:
- Cache our POLLED@10s metrics at 30s for its own consumers
- Disable metrics it doesn't need
- But never claim to have fresher data than we provide

This prevents stale-data bugs (client thinks it's getting live data but it's 10s old) and over-polling (hammering the kernel for data that hasn't changed).

## File Layout

```
internal/metrics/darwin/
  24.6.0.textproto              # KernelMetricRegistry for Darwin 24.6.0 ARM64
  generate_textproto.go         # Generator (go run, reads from registry.go)
```

The textproto is embedded via `//go:embed` and loaded at server startup. It contains only access pattern classifications — description, value_type, and category are merged from `registry.go` at serve time.

## Current State (implemented)

- `KernelMetric` carries full identity (description, value_type, category) + access patterns
- `ListKnownMetrics` returns `KernelMetricRegistry` — unified view for clients
- `GetKernelRegistry` also returns the full merged registry
- Category filter still works on `ListKnownMetrics`
- Kernel registry textproto for Darwin 24.6.0 with 250 metrics classified
- Kernel patterns: 85 STATIC (hardware immutables), 165 DYNAMIC (everything that can change)
- Recommended patterns: 80 STATIC (read once), 170 POLLED (10-60s TTLs)
- `ValidateRegistry()` ensures 1:1 match between textproto and Known registry
- **Per-host polling loop**: single goroutine with configurable tick interval
  - STATIC metrics read once at startup, frozen in poller store
  - POLLED metrics refreshed per TTL via `nextGatherNs` tracking
  - Non-polled metrics fall through to live MIB-cached reads
  - `--poll-interval` flag (default 500ms), pass 0 to disable polling
  - Mutex-locked `metricStore` for concurrent read/write safety

## Next Steps

### Phase 6: Service-Level Overrides
- `--config` flag loads a config textproto with per-metric overrides
- Overrides clamped to monotonic invariant against kernel registry
- Warnings if config tries to increase frequency beyond kernel pattern

### Phase 7: Client Access Masks
- Client sends desired access patterns in request
- Service clamps to `min(service_pattern, client_request)`
- Response includes effective pattern so client knows actual freshness

### Phase 8: Export / Observability
- Prometheus metrics endpoint
- Streaming gRPC (server-push on change, respecting poll intervals)
- TSDB push integration
