# iogpu — GPU Memory

GPU wired memory management. 5 keys, all `int32`.

| Key | Example | Description |
|-----|---------|-------------|
| `iogpu.wired_limit_mb` | 11316 | GPU wired memory limit (MB) |
| `iogpu.wired_lwm_mb` | 10184 | GPU wired low water mark (MB) |
| `iogpu.dynamic_lwm` | 1 | Dynamic low water mark enabled |
| `iogpu.debug_flags` | 0 | GPU debug flags |
| `iogpu.disable_wired_collector` | 0 | Disable wired memory collector |

## Notes

- On unified memory (Apple Silicon), GPU shares physical RAM with CPU.
- `wired_limit_mb` is typically ~66% of physical RAM.
- The wired collector reclaims GPU memory when approaching limits.
- `dynamic_lwm` allows the low water mark to adjust based on load.
