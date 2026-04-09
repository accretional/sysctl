# vm — Page Counts

Page-level memory counters. All 4 bytes / `int32`. Multiply by page size (16384 on ARM64) for bytes.

## Current state

| Key | Description |
|-----|-------------|
| `vm.pages` | Total physical pages |
| `vm.page_free_count` | Free pages available |
| `vm.page_speculative_count` | Speculative (prefetched) pages |
| `vm.page_cleaned_count` | Cleaned (written-back) pages |
| `vm.page_pageable_internal_count` | Pageable internal (anonymous) pages |
| `vm.page_pageable_external_count` | Pageable external (file-backed) pages |
| `vm.page_purgeable_count` | Purgeable pages |
| `vm.page_purgeable_wired_count` | Purgeable wired pages |
| `vm.page_reusable_count` | Reusable pages |
| `vm.page_realtime_count` | Realtime pages |
| `vm.vm_page_external_count` | External (file-cache) pages |
| `vm.vm_page_background_count` | Background pages |
| `vm.vm_page_background_internal_count` | Background internal pages |
| `vm.vm_page_background_external_count` | Background external pages |
| `vm.vm_page_background_target` | Background page target |

## Configuration

| Key | Description |
|-----|-------------|
| `vm.vm_page_free_target` | Target free page count |
| `vm.vm_page_background_mode` | Background page mode |
| `vm.vm_page_donate_mode` | Page donation mode |
| `vm.vm_page_donate_target_high` | High donation target |
| `vm.vm_page_donate_target_low` | Low donation target |

## Accumulated counters (int64)

| Key | Bytes | Description |
|-----|-------|-------------|
| `vm.pages_grabbed` | 8 | Total pages grabbed from free list |

## Notes

- **Memory breakdown**: free + speculative + pageable_internal + pageable_external + purgeable + wired ≈ total
- `vm.page_free_count` < `vm.vm_page_free_target` → memory pressure is active
- `vm.page_speculative_count` = prefetched file data, first to be reclaimed
- `vm.page_purgeable_count` = can be instantly discarded (NSPurgeableData, etc.)
- All page counts are in 16 KB pages on Apple Silicon.
