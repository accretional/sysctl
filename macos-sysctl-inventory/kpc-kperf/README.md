# kpc/kperf/ktrace — Performance Counters & Tracing

Kernel performance counter and profiling infrastructure.

## kpc.* — Kernel Performance Counters

| Key | Bytes | Type | Example | Description |
|-----|-------|------|---------|-------------|
| `kpc.pc_capture_supported` | 4 | int32 | 1 | PC capture supported |

## kperf.* — Kernel Performance Framework

| Key | Bytes | Type | Description |
|-----|-------|------|-------------|
| `kperf.debug_level` | 4 | int32 | Debug level |
| `kperf.limits.timer_min_period_ns` | 8 | int64 | Min sampling timer period (ns) |
| `kperf.limits.timer_min_bg_period_ns` | 8 | int64 | Min background sampling period |
| `kperf.limits.timer_min_pet_period_ns` | 8 | int64 | Min PET sampling period |
| `kperf.limits.max_action_count` | 4 | int32 | Max sampling actions |

## ktrace.* — Kernel Tracing

| Key | Bytes | Type | Description |
|-----|-------|------|-------------|
| `ktrace.state` | 4 | int32 | Tracing state (0=off) |
| `ktrace.active_mask` | 4 | int32 | Active trace mask |
| `ktrace.owning_pid` | 4 | int32 | PID owning the trace session |
| `ktrace.background_pid` | 4 | int32 | Background trace PID |
| `ktrace.configured_by` | 4 | int32 | Configured by (0=none, 1=ktrace, 2=kdebug) |

## Notes

- `kpc.*` is the low-level interface to ARM PMU counters. Requires entitlements on macOS.
- `kperf.*` is the higher-level sampling framework (used by Instruments).
- `kperf.limits.timer_min_period_ns` sets the minimum sampling interval — typically ~1ms.
- `ktrace.state` = 0 means no active kernel tracing. Non-zero when Instruments or dtrace is running.
- Most kpc/kperf operations require root or specific entitlements.
