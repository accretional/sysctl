# Cache & Config Design

## Access Patterns

Every metric has an access pattern controlling how it's read:

| Pattern | Behavior |
|---|---|
| `STATIC` | Read once at startup, served from cache forever |
| `POLLED` | Background refresh at TTL interval, serve from cache |
| `CACHED` | Read live, cache for TTL, re-read after expiry |
| `DYNAMIC` | Read live on every request (no caching) |
| `DISABLED` | Never read, return error if requested |

## Proto Messages

```protobuf
enum AccessPattern { STATIC, POLLED, CACHED, DYNAMIC, DISABLED }

message AccessConfig {
  AccessPattern pattern = 1;
  google.protobuf.Duration ttl = 2;  // Used by POLLED and CACHED
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
   DYNAMIC            POLLED@10s       CACHED@30s     CACHED@60s
```

**The invariant:** if a layer upstream says a metric is `STATIC`, all downstream layers must treat it as `STATIC` or `DISABLED`. If upstream says `POLLED@10s`, downstream can serve `POLLED@30s`, `CACHED@60s`, or `STATIC` (snapshot) — but never `POLLED@5s` or `DYNAMIC` (which would claim more freshness than the source provides).

Access patterns are ordered by decreasing frequency:

```
DYNAMIC > POLLED > CACHED > STATIC > DISABLED
```

A downstream layer's effective pattern must be ≤ its upstream's pattern. Within the same pattern, TTL must be ≥ upstream's TTL.

### Current behavior

Today, the service does a direct pass-through: `ListKnownMetrics` returns the kernel registry's `recommended_access_pattern` unchanged. All reads are actually `DYNAMIC` (live on every request via MIB cache). The access patterns are metadata only — they tell clients what's *recommended*, not what's *enforced*.

### Future: tiered access enforcement

When the service implements background polling, the chain becomes real:

1. **Service reads kernel at `kernel_access_pattern` rate** (or the recommended rate)
   - STATIC metrics: read once at startup, frozen
   - POLLED@10s: background goroutine refreshes every 10s
   - DYNAMIC: read live per request

2. **Service exposes `recommended_access_pattern` to clients**
   - Clients that poll the service should respect these patterns
   - A client polling a POLLED@10s metric every 1s gets the same 10s-stale data

3. **Service-level overrides (future)**
   - Operator config can override the kernel registry's recommendations
   - Overrides must respect the monotonic invariant (can only decrease, never increase)
   - Attempting to set DYNAMIC on a STATIC metric is a config error

4. **Client-level masks (future)**
   - Clients request a "mask" — their desired access patterns
   - Service clamps each metric to `min(service_pattern, client_request)`
   - This lets a lightweight dashboard client say "I only need CACHED@60s for everything"

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
- All access patterns are metadata only — actual reads are all DYNAMIC (live via MIB cache)

## Next Steps

### Phase 5: Background Polling Loop
Implement the access patterns as real behavior:
- STATIC: read once at startup into frozen cache
- POLLED: background goroutines refresh at TTL interval
- CACHED: lazy read with expiry timestamp
- DYNAMIC: pass through to live MIB-cached read (current behavior)

### Phase 6: Service-Level Overrides
- `--config` flag loads a `ServiceConfig` textproto with per-metric overrides
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
