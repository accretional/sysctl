# vm — Pageout Activity

Counters for page-out daemon activity. All 8 bytes / `int64` (accumulated counters).

## Pageout counters

| Key | Description |
|-----|-------------|
| `vm.pageout_inactive_clean` | Inactive clean pages freed |
| `vm.pageout_inactive_used` | Inactive used pages (reactivated) |
| `vm.pageout_inactive_dirty_internal` | Inactive dirty internal pages written out |
| `vm.pageout_inactive_dirty_external` | Inactive dirty external pages written out |
| `vm.pageout_speculative_clean` | Speculative pages freed |
| `vm.pageout_freed_external` | External pages freed |
| `vm.pageout_freed_speculative` | Speculative pages freed (alt counter) |
| `vm.pageout_freed_cleaned` | Cleaned pages freed |
| `vm.pageout_protect_realtime` | Realtime pages protected from pageout |
| `vm.pageout_protected_realtime` | Currently protected realtime pages |
| `vm.pageout_forcereclaimed_realtime` | Realtime pages force-reclaimed |
| `vm.pageout_protected_sharedcache` | Shared cache pages protected |
| `vm.pageout_forcereclaimed_sharedcache` | Shared cache pages force-reclaimed |

## Reusable counters (int64)

| Key | Description |
|-----|-------------|
| `vm.reusable_success` | Successful reuse operations |
| `vm.reusable_failure` | Failed reuse attempts |
| `vm.reusable_pages_shared` | Pages shared via reuse |
| `vm.reusable_nonwritable` | Non-writable reusable pages |
| `vm.reusable_shared` | Shared reusable count |
| `vm.reusable_reclaimed` | Reusable pages reclaimed |
| `vm.reuse_success` | Reuse successes |
| `vm.reuse_failure` | Reuse failures |

## VM pageout consideration (int64)

| Key | Description |
|-----|-------------|
| `vm.vm_pageout_considered_bq_internal` | Internal pages considered for pageout |
| `vm.vm_pageout_considered_bq_external` | External pages considered |
| `vm.vm_pageout_rejected_bq_internal` | Internal pages rejected (not paged out) |
| `vm.vm_pageout_rejected_bq_external` | External pages rejected |

## Notes

- These are monotonically increasing counters since boot. Delta between reads = activity.
- High `pageout_inactive_dirty_internal` rate = heavy anonymous memory writeback (pressure).
- `pageout_inactive_used` = pages that were on the inactive list but got accessed (rescued).
- Compare `considered` vs `rejected` ratios to see pageout efficiency.
