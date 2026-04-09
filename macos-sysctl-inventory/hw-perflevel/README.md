# hw — Performance Levels (Apple Silicon P/E Cores)

Apple Silicon exposes per-performance-level details. `perflevel0` = **P-cores** (Performance),
`perflevel1` = **E-cores** (Efficiency). All values are 4 bytes / `int32` except `name` (string).

## perflevel0 — Performance Cores

| Key | Type | Example | Description |
|-----|------|---------|-------------|
| `hw.perflevel0.name` | string | "Performance" | Human-readable name |
| `hw.perflevel0.physicalcpu` | int32 | 4 | Physical P-cores |
| `hw.perflevel0.physicalcpu_max` | int32 | 4 | Max physical P-cores |
| `hw.perflevel0.logicalcpu` | int32 | 4 | Logical P-cores |
| `hw.perflevel0.logicalcpu_max` | int32 | 4 | Max logical P-cores |
| `hw.perflevel0.l1icachesize` | int32 | 196608 | L1I cache per P-core (192 KB) |
| `hw.perflevel0.l1dcachesize` | int32 | 131072 | L1D cache per P-core (128 KB) |
| `hw.perflevel0.l2cachesize` | int32 | 16777216 | Shared L2 cache (16 MB) |
| `hw.perflevel0.cpusperl2` | int32 | 4 | CPUs sharing one L2 |

## perflevel1 — Efficiency Cores

| Key | Type | Example | Description |
|-----|------|---------|-------------|
| `hw.perflevel1.name` | string | "Efficiency" | Human-readable name |
| `hw.perflevel1.physicalcpu` | int32 | 6 | Physical E-cores |
| `hw.perflevel1.physicalcpu_max` | int32 | 6 | Max physical E-cores |
| `hw.perflevel1.logicalcpu` | int32 | 6 | Logical E-cores |
| `hw.perflevel1.logicalcpu_max` | int32 | 6 | Max logical E-cores |
| `hw.perflevel1.l1icachesize` | int32 | 131072 | L1I cache per E-core (128 KB) |
| `hw.perflevel1.l1dcachesize` | int32 | 65536 | L1D cache per E-core (64 KB) |
| `hw.perflevel1.l2cachesize` | int32 | 4194304 | Shared L2 cache (4 MB) |
| `hw.perflevel1.cpusperl2` | int32 | 6 | CPUs sharing one L2 |

## Notes

- `hw.nperflevels` tells you how many levels exist (2 on all current Apple Silicon).
- On chips with 3 cluster types (e.g., hypothetical future), `perflevel2` would appear.
- **P-core caches are larger** (192 KB L1I vs 128 KB, 16 MB L2 vs 4 MB).
- These are all 4 bytes (int32) unlike the top-level hw.* cache keys which are 8 bytes.
