# kern — Memory Status

Jetsam / memory pressure signals from the kernel memory status subsystem.

| Key | Bytes | Type | Example | Description |
|-----|-------|------|---------|-------------|
| `kern.memorystatus_level` | 4 | int32 | 80 | Current memory pressure level (0-100, %) |
| `kern.memorystatus_purge_on_warning` | 4 | int32 | 2 | Purge behavior on memory warning |
| `kern.memorystatus_purge_on_urgent` | 4 | int32 | 5 | Purge behavior on urgent memory |
| `kern.memorystatus_purge_on_critical` | 4 | int32 | 8 | Purge behavior on critical memory |
| `kern.vm_pressure_level_transition_threshold` | 4 | int32 | | Threshold for pressure level transitions |

## Notes

- `kern.memorystatus_level` is the primary memory health indicator. Values near 0 = severe pressure, 100 = abundant.
- This drives macOS Jetsam (process killing) decisions and app memory warnings.
- The purge values configure how aggressively purgeable memory is reclaimed at each pressure tier.
- Monitor `kern.memorystatus_level` alongside `vm.page_free_count` for a complete memory picture.
