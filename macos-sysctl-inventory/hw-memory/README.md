# hw — Memory

Physical memory and page size. Note: these return **8 bytes** on ARM64 (not 4).

| Key | Bytes | Type | Example | Description |
|-----|-------|------|---------|-------------|
| `hw.memsize` | 8 | uint64 | 17179869184 | Total physical memory in bytes |
| `hw.memsize_usable` | 8 | uint64 | 16632807424 | Usable physical memory (minus firmware/kernel reserved) |
| `hw.pagesize` | 8 | int64 | 16384 | System page size (16 KB on Apple Silicon) |
| `hw.pagesize32` | 8 | int64 | 16384 | 32-bit page size (same on ARM64) |

## Notes

- Apple Silicon uses **16 KB pages** (vs 4 KB on Intel). This affects page count metrics in `vm.*`.
- `hw.memsize_usable` is typically ~500 MB less than `hw.memsize` due to reserved regions.
- These are all 8 bytes on ARM64 despite being 4 bytes on x86_64. Use `int64`/`uint64` types.
