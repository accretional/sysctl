# vm — Memory Compressor

macOS compresses inactive memory before swapping. These metrics track compressor activity.

## Compressor state

| Key | Bytes | Type | Description |
|-----|-------|------|-------------|
| `vm.compressor_mode` | 4 | int32 | Compressor mode (7 = compress + swap) |
| `vm.compressor_is_active` | 4 | int32 | Compressor currently active |
| `vm.compressor_available` | 4 | int32 | Compressor available |

## Byte counters (uint64)

| Key | Bytes | Type | Description |
|-----|-------|------|-------------|
| `vm.compressor_input_bytes` | 8 | uint64 | Total bytes submitted to compressor |
| `vm.compressor_compressed_bytes` | 8 | uint64 | Bytes after compression |
| `vm.compressor_bytes_used` | 8 | uint64 | Current compressor memory usage |
| `vm.compressor_pool_size` | 8 | uint64 | Compressor pool size |

## Segment stats (int32)

| Key | Description |
|-----|-------------|
| `vm.compressor_segment_pages_compressed` | Pages in compressed segments |
| `vm.compressor_segment_limit` | Max compressed segments |
| `vm.compressor_segment_pages_compressed_limit` | Max compressed pages |
| `vm.compressor_segment_buffer_size` | Segment buffer size |
| `vm.compressor_segment_alloc_size` | Segment allocation size |
| `vm.compressor_min_csegs_per_major_compaction` | Min segments for major compaction |
| `vm.compressor_segment_svp_in_hash` | SVP entries in hash |
| `vm.compressor_segment_svp_hash_succeeded` | SVP hash hits |
| `vm.compressor_segment_svp_hash_failed` | SVP hash misses |

## Swapout stats (int64)

| Key | Bytes | Description |
|-----|-------|-------------|
| `vm.compressor_swapouts_under_30s` | 8 | Swapouts within 30s of compression |
| `vm.compressor_swapouts_under_60s` | 8 | Swapouts within 60s |
| `vm.compressor_swapouts_under_300s` | 8 | Swapouts within 300s |
| `vm.compressor_swapper_reclaim_swapins` | 8 | Reclaim swap-ins |
| `vm.compressor_swapper_defrag_swapins` | 8 | Defrag swap-ins |
| `vm.compressor_swapper_swapout_threshold_exceeded` | 8 | Swapout threshold exceeded |
| `vm.compressor_swapper_swapout_fragmentation_detected` | 8 | Fragmentation events |
| `vm.compressor_swapper_swapout_free_count_low` | 8 | Low free count swapouts |
| `vm.compressor_swapper_swapout_thrashing_detected` | 8 | Thrashing events |
| `vm.compressor_swapper_swapout_fileq_throttled` | 8 | File queue throttles |

## Compression algorithms — WK (WKdm) and LZ4

| Key | Bytes | Type | Description |
|-----|-------|------|-------------|
| `vm.wk_compressions` | 8 | int64 | WKdm compressions |
| `vm.wk_compressed_bytes_total` | 8 | int64 | WKdm total compressed bytes |
| `vm.wk_compressed_bytes_exclusive` | 8 | int64 | WKdm exclusive compressed bytes |
| `vm.wk_decompressions` | 8 | int64 | WKdm decompressions |
| `vm.wk_decompressed_bytes` | 8 | int64 | WKdm decompressed bytes |
| `vm.wk_sv_compressions` | 8 | int64 | WKdm single-value compressions |
| `vm.wk_sv_decompressions` | 8 | int64 | WKdm single-value decompressions |
| `vm.wk_mzv_compressions` | 8 | int64 | WKdm mostly-zero-value compressions |
| `vm.lz4_compressions` | 8 | int64 | LZ4 compressions |
| `vm.lz4_compressed_bytes` | 8 | int64 | LZ4 compressed bytes |
| `vm.lz4_decompressed_bytes` | 8 | int64 | LZ4 decompressed bytes |
| `vm.lz4_decompressions` | 8 | int64 | LZ4 decompressions |
| `vm.lz4_compression_failures` | 8 | int64 | LZ4 compression failures |

## Configuration

| Key | Type | Description |
|-----|------|-------------|
| `vm.compressor_swapout_target_age` | int32 | Target age before swapout (seconds) |
| `vm.compressor_eval_period_in_msecs` | int32 | Evaluation period |
| `vm.compressor_sample_min_in_msecs` | int32 | Min sample period |
| `vm.compressor_sample_max_in_msecs` | int32 | Max sample period |
| `vm.compressor_thrashing_threshold_per_10msecs` | int32 | Thrashing threshold |
| `vm.compressor_thrashing_min_per_10msecs` | int32 | Min thrashing count |
| `vm.compressor_timing_enabled` | int32 | Timing instrumentation |
| `vm.compressor_pool_multiplier` | int32 | Pool size multiplier |

## Notes

- **Compression ratio** = `compressor_input_bytes / compressor_compressed_bytes`
- macOS tries WKdm first (fast, good for memory patterns), falls back to LZ4.
- `compressor_bytes_used` = current RSS of the compressor itself.
- Watch `swapouts_under_30s` — high values mean memory is being compressed then immediately swapped (bad).
- `compressor_swapper_thrashing_detected` > 0 means the system is in a swap storm.
