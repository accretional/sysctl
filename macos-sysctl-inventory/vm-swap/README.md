# vm — Swap

Swap file configuration and usage.

| Key | Bytes | Type | Example | Description |
|-----|-------|------|---------|-------------|
| `vm.swapusage` | 32 | SwapUsage | | Swap total/available/used |
| `vm.swap_enabled` | 4 | int32 | 1 | Swap enabled (usually 1) |
| `vm.swapfileprefix` | varies | string | "/System/Volumes/VM/swapfile" | Swap file path prefix |
| `vm.compressor_available` | 4 | int32 | 1 | Memory compressor available |
| `vm.compressor_mode` | 4 | int32 | 7 | Mode (7 = compress + swap) |

## SwapUsage struct layout (32 bytes)

```
struct xsw_usage {
    Total uint64  // Total swap space in bytes
    Avail uint64  // Available swap space
    Used  uint64  // Used swap space
    Flags uint64  // Swap flags
}
```

## Notes

- macOS creates swap files dynamically at `/System/Volumes/VM/swapfile*`.
- The compressor (see [vm-compressor/](../vm-compressor/)) runs before swap — most memory pressure is handled by compression.
- `vm.compressor_mode` = 7 means both compression and swap are enabled (normal state).
- On Apple Silicon, aggressive compression means swap usage is typically low unless under extreme pressure.
