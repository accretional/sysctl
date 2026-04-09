# hw — Cache

CPU cache hierarchy. Top-level keys return **8 bytes** on ARM64. Per-perflevel keys are 4 bytes.

| Key | Bytes | Type | Example | Description |
|-----|-------|------|---------|-------------|
| `hw.cachelinesize` | 8 | int64 | 128 | Cache line size in bytes |
| `hw.l1icachesize` | 8 | int64 | 65536 | L1 instruction cache (system-wide, smallest) |
| `hw.l1dcachesize` | 8 | int64 | 65536 | L1 data cache (system-wide, smallest) |
| `hw.l2cachesize` | 8 | int64 | 4194304 | L2 cache (system-wide, largest) |
| `hw.tbfrequency` | 8 | int64 | 24000000 | Timebase frequency in Hz |
| `hw.cacheconfig` | 80 | raw | [10]uint64 | CPUs per cache level (index 0 = all, 1 = L1, 2 = L2, ...) |
| `hw.cachesize` | 80 | raw | [10]uint64 | Cache size in bytes per level |

## Notes

- `hw.cachelinesize` is 128 bytes on Apple Silicon (vs 64 on Intel).
- Top-level L1/L2 sizes report the **smallest** across P/E cores. For per-core-type sizes, see [hw-perflevel/](../hw-perflevel/).
- On Apple M4: P-cores have 192 KB L1I, 128 KB L1D, 16 MB shared L2. E-cores have 128 KB L1I, 64 KB L1D, 4 MB shared L2.
- `hw.cacheconfig` and `hw.cachesize` are arrays of 10 uint64 values (one per cache level, unused levels are 0).
- `hw.tbfrequency` = 24 MHz on all current Apple Silicon. Used for mach_absolute_time conversion.
