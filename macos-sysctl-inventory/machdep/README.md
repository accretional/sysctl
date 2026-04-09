# machdep — Machine-Dependent

Machine-dependent kernel parameters. 18 keys total.

## CPU Information

| Key | Bytes | Type | Example | Description |
|-----|-------|------|---------|-------------|
| `machdep.cpu.brand_string` | varies | string | "Apple M4" | CPU brand string |
| `machdep.cpu.core_count` | 4 | int32 | 10 | Physical core count |
| `machdep.cpu.thread_count` | 4 | int32 | 10 | Thread count (= core count on Apple Silicon) |
| `machdep.cpu.cores_per_package` | 4 | int32 | 10 | Cores per package |
| `machdep.cpu.logical_per_package` | 4 | int32 | 10 | Logical CPUs per package |

## Timing & Wake

| Key | Bytes | Type | Example | Description |
|-----|-------|------|---------|-------------|
| `machdep.time_since_reset` | 8 | uint64 | | Time since CPU reset (abs time) |
| `machdep.wake_abstime` | 8 | uint64 | | Absolute time of last wake |
| `machdep.wake_conttime` | 8 | uint64 | | Continuous time of last wake |

## Security & Debug

| Key | Bytes | Type | Example | Description |
|-----|-------|------|---------|-------------|
| `machdep.ptrauth_enabled` | 4 | int32 | 1 | Pointer authentication enabled |
| `machdep.virtual_address_size` | 4 | int32 | 47 | Virtual address bits |
| `machdep.user_idle_level` | 4 | int32 | 0 | User idle level |
| `machdep.deferred_ipi_timeout` | 4 | int32 | 64000 | Deferred IPI timeout |

## PHY Delay Reporting

| Key | Bytes | Type | Description |
|-----|-------|------|-------------|
| `machdep.report_phy_read_delay` | 8 | int64 | PHY read delay reporting threshold |
| `machdep.report_phy_write_delay` | 8 | int64 | PHY write delay reporting threshold |
| `machdep.trace_phy_read_delay` | 4 | int32 | Trace PHY read delays |
| `machdep.trace_phy_write_delay` | 4 | int32 | Trace PHY write delays |
| `machdep.phy_read_delay_panic` | 4 | int32 | Panic on PHY read delay |
| `machdep.phy_write_delay_panic` | 4 | int32 | Panic on PHY write delay |

## Notes

- `machdep.virtual_address_size` = 47 on Apple Silicon (128 TB virtual address space).
- `machdep.ptrauth_enabled` = 1 confirms PAC is active (pointer authentication codes for security).
- Apple Silicon has no SMT, so `thread_count` = `core_count`.
