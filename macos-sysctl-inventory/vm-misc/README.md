# vm — Miscellaneous

Remaining `vm.*` keys (~160) covering code signing, faults, copy-on-write, and debug counters.

## Code signing (vm.cs_*)

| Key | Type | Description |
|-----|------|-------------|
| `vm.cs_blob_count` | int32 | Active code signature blobs |
| `vm.cs_blob_count_peak` | int32 | Peak blob count |
| `vm.cs_blob_size` | int32 | Current blob memory usage |
| `vm.cs_blob_size_peak` | int32 | Peak blob memory |
| `vm.cs_blob_size_max` | int32 | Max single blob size |
| `vm.cs_all_vnodes` | int32 | CS enforcement on all vnodes |
| `vm.cs_debug` | int32 | CS debug mode |
| `vm.cs_enforcement_panic` | int32 | Panic on CS violation |
| `vm.cs_force_hard` | int32 | Force hard CS validation |
| `vm.cs_force_kill` | int32 | Kill on CS failure |

## Copy-on-write & faults

| Key | Bytes | Type | Description |
|-----|-------|------|-------------|
| `vm.copied_on_read` | 8 | int64 | Copy-on-read events |
| `vm.vm_should_cow_but_wired` | 4 | int32 | COW skipped (wired) |
| `vm.vm_create_upl_extra_cow` | 8 | int64 | Extra COW from UPL creation |
| `vm.vm_create_upl_extra_cow_pages` | 8 | int64 | Extra COW pages |
| `vm.vm_copy_src_large` | 4 | int32 | Large source copies |
| `vm.vm_copy_src_not_internal` | 4 | int32 | External source copies |

## Fault counters

| Key | Type | Description |
|-----|------|-------------|
| `vm.fault_resilient_media_initiate` | int64 | Resilient media fault initiations |
| `vm.fault_resilient_media_retry` | int64 | Resilient media retries |
| `vm.fault_resilient_media_proceed` | int64 | Resilient media proceeds |
| `vm.fault_resilient_media_release` | int64 | Resilient media releases |
| `vm.fault_resilient_media_abort1` | int64 | Resilient media abort type 1 |
| `vm.fault_resilient_media_abort2` | int64 | Resilient media abort type 2 |

## Shared regions

| Key | Type | Description |
|-----|------|-------------|
| `vm.shared_region_count` | int32 | Active shared regions |
| `vm.shared_region_peak` | int32 | Peak shared regions |
| `vm.shared_region_pager_copied` | int32 | Shared region pages copied |
| `vm.shared_region_pager_slid` | int32 | Shared region pages slid (ASLR) |
| `vm.shared_region_pager_reclaimed` | int32 | Shared region pages reclaimed |

## Prefault

| Key | Type | Description |
|-----|------|-------------|
| `vm.prefault_nb_pages` | int32 | Prefault page count |
| `vm.prefault_nb_bailout` | int32 | Prefault bailouts |

## Notes

- Code signing stats show memory overhead of Gatekeeper/codesign validation.
- High `cs_blob_count` correlates with many running executables/libraries.
- Copy-on-write counters are useful for understanding fork() and mmap() behavior.
