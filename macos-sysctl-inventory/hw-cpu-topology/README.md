# hw — CPU Topology

CPU identification and topology. All 4 bytes / `int32` unless noted.

| Key | Type | Example | Description |
|-----|------|---------|-------------|
| `hw.ncpu` | int32 | 10 | Total number of CPUs (logical) |
| `hw.activecpu` | int32 | 10 | Currently active CPUs |
| `hw.physicalcpu` | int32 | 10 | Physical CPU cores |
| `hw.physicalcpu_max` | int32 | 10 | Max physical CPU cores |
| `hw.logicalcpu` | int32 | 10 | Logical CPU cores |
| `hw.logicalcpu_max` | int32 | 10 | Max logical CPU cores |
| `hw.packages` | int32 | 1 | Number of CPU packages |
| `hw.nperflevels` | int32 | 2 | Number of performance levels (P+E cores) |
| `hw.cpu64bit_capable` | int32 | 1 | 64-bit capable (always 1 on ARM64) |
| `hw.cputype` | int32 | 16777228 | Mach-O CPU type (CPU_TYPE_ARM64) |
| `hw.cpusubtype` | int32 | 2 | Mach-O CPU subtype |
| `hw.cpufamily` | int32 | 1867590060 | CPU family identifier |
| `hw.cpusubfamily` | int32 | 2 | CPU subfamily |
| `hw.byteorder` | int32 | 1234 | Byte order (1234 = little-endian) |

## Notes

- `hw.nperflevels` = 2 on all Apple Silicon with P/E cores. See [hw-perflevel/](../hw-perflevel/) for per-level detail.
- `hw.cpufamily` uniquely identifies the CPU generation (M1, M2, M3, M4 each have different values).
- `hw.activecpu` can change at runtime based on power management.
