# vm — Memory Pressure

Primary memory health indicators.

| Key | Bytes | Type | Example | Description |
|-----|-------|------|---------|-------------|
| `vm.memory_pressure` | 4 | int32 | 1 | Memory pressure level (1=normal, 2=warning, 4=critical) |
| `vm.page_free_wanted` | 4 | int32 | 0 | Pages wanted by free list (>0 = pressure) |
| `vm.vm_page_free_target` | 4 | int32 | 4000 | Target free pages |
| `vm.vm_page_filecache_min` | 4 | int32 | | Min file cache pages |
| `vm.vm_page_xpmapped_min` | 4 | int32 | | Min XP-mapped pages |
| `vm.vm_ripe_target_age_in_secs` | 4 | int32 | | Age target for ripe pages |
| `vm.darkwake_mode` | 4 | int32 | 0 | Memory behavior during dark wake |
| `vm.panic_ws_crash` | 4 | int32 | 0 | Panic on working set crash |

## How to interpret memory pressure

1. **`vm.memory_pressure`** — the top-level signal:
   - `1` = Normal (no pressure)
   - `2` = Warning (starting to reclaim)
   - `4` = Critical (aggressive reclamation, Jetsam killing apps)

2. **`vm.page_free_wanted`** — when > 0, the system is actively starved for free pages

3. **`kern.memorystatus_level`** — 0-100% scale (see [kern-memory-status/](../kern-memory-status/))

4. **`vm.page_free_count`** vs `vm.vm_page_free_target` — if free < target, pressure is building

## Notes

- On Apple Silicon with 16 KB pages, page counts × 16384 = bytes.
- Memory pressure drives compressor activation, swap, and Jetsam.
