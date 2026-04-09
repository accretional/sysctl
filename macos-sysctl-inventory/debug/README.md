# debug — Debug

Kernel debug settings. 56 keys, all `int32`.

## BPF (Berkeley Packet Filter)

| Key | Example | Description |
|-----|---------|-------------|
| `debug.bpf_bufsize` | 4096 | BPF buffer size |
| `debug.bpf_maxbufsize` | 524288 | Max BPF buffer size |
| `debug.bpf_maxdevices` | 256 | Max BPF devices |
| `debug.bpf_debug` | 0 | BPF debug mode |
| `debug.bpf_wantpktap` | 0 | BPF pktap mode |
| `debug.bpf_hdr_comp_enable` | 0 | BPF header compression |
| `debug.bpf_trunc_overflow` | 0 | BPF truncation overflow |
| `debug.bpf_bufsize_cap` | 0 | BPF buffer size cap |

## I/O Throttling

| Key | Example | Description |
|-----|---------|-------------|
| `debug.lowpri_throttle_enabled` | 1 | Low-priority I/O throttling |
| `debug.lowpri_throttle_tier1_io_period_msecs` | 200 | Tier 1 I/O period |
| `debug.lowpri_throttle_tier1_io_period_ssd_msecs` | 100 | Tier 1 I/O period (SSD) |
| `debug.lowpri_throttle_tier1_window_msecs` | 200 | Tier 1 throttle window |
| `debug.lowpri_throttle_tier2_io_period_msecs` | 50 | Tier 2 I/O period |
| `debug.lowpri_throttle_tier2_io_period_ssd_msecs` | 25 | Tier 2 I/O period (SSD) |
| `debug.lowpri_throttle_tier3_io_period_msecs` | 15 | Tier 3 I/O period |
| `debug.lowpri_throttle_tier3_io_period_ssd_msecs` | 10 | Tier 3 I/O period (SSD) |

## Disk I/O Device (debug.didevice_*)

| Key | Description |
|-----|-------------|
| `debug.didevice_cache_size_default` | Default cache size |
| `debug.didevice_enable_cache` | Cache enabled |
| `debug.didevice_queue_depth` | Queue depth |
| `debug.didevice_bounce_count` | Bounce buffer count |
| `debug.didevice_no_bounce_count` | Non-bounce count |
| `debug.didevice_read_by_kernel_bytes` | Kernel read bytes |
| `debug.didevice_write_by_kernel_bytes` | Kernel write bytes |
| `debug.didevice_cache_fully_satisfied` | Requests served from cache |
| `debug.didevice_cache_spared_bytes` | Bytes spared by cache |

## Scheduler

| Key | Description |
|-----|-------------|
| `debug.sched` | Scheduler debug flags |
| `debug.sched_hygiene_debug_available` | Scheduler hygiene debug |

## Miscellaneous

| Key | Description |
|-----|-------------|
| `debug.noidle` | Prevent idle sleep |
| `debug.kextlog` | Kext logging level |
| `debug.iotrace` | I/O tracing |
| `debug.toggle_address_reuse` | Address reuse toggle |
| `debug.swd_timeout` | Software watchdog timeout |
| `debug.swd_panic` | SWD panic enabled |
| `debug.swd_sleep_timeout` | SWD sleep timeout |
| `debug.swd_wake_timeout` | SWD wake timeout |

## Notes

- I/O throttle tiers control macOS "low priority" I/O (Time Machine, Spotlight indexing).
- SSD-specific periods are shorter because SSDs have lower latency.
- `debug.didevice_*` counters are useful for disk I/O performance analysis.
- `debug.swd_*` = software watchdog (kernel panic on hang).
