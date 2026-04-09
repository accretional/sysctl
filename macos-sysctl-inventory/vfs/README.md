# vfs — Filesystem

Virtual filesystem stats and configuration. 113 keys.

## Vnode stats (vfs.vnstats.*)

| Key | Bytes | Type | Description |
|-----|-------|------|-------------|
| `vfs.vnstats.num_vnodes` | 8 | int64 | Total vnodes |
| `vfs.vnstats.num_free_vnodes` | 8 | int64 | Free vnodes |
| `vfs.vnstats.num_recycledvnodes` | 8 | int64 | Recycled vnodes |
| `vfs.vnstats.num_newvnode_calls` | 8 | int64 | New vnode allocations |
| `vfs.vnstats.num_rapid_aging_vnodes` | 4 | int32 | Rapidly aging vnodes |

## Namespace cache (vfs.ncstats.*)

| Key | Type | Description |
|-----|------|-------------|
| `vfs.ncstats.ncs_goodhits` | 4 | Good cache hits |
| `vfs.ncstats.ncs_neghits` | 4 | Negative cache hits |
| `vfs.ncstats.ncs_badhits` | 4 | Bad cache hits |
| `vfs.ncstats.ncs_miss` | 4 | Cache misses |
| `vfs.ncstats.ncs_long` | 4 | Long lookups |
| `vfs.ncstats.ncs_pass2` | 4 | Pass-2 lookups |
| `vfs.ncstats.ncs_2passes` | 4 | Two-pass lookups |

## APFS (vfs.generic.apfs.*)

| Key | Bytes | Type | Description |
|-----|-------|------|-------------|
| `vfs.generic.apfs.allocated` | 8 | int64 | APFS allocated bytes |
| `vfs.generic.apfs.unwritten_freeze_threshold` | 8 | uint64 | Unwritten freeze threshold |

## LIFS (vfs.generic.lifs.*)

| Key | Bytes | Type | Description |
|-----|-------|------|-------------|
| `vfs.generic.lifs.read_meta_cache_hit` | 8 | int64 | Read metadata cache hits |
| `vfs.generic.lifs.write_meta_cache_hit` | 8 | int64 | Write metadata cache hits |

## General

| Key | Type | Description |
|-----|------|-------------|
| `vfs.nummntops` | int32 | Mount operations count |
| `vfs.purge_vm_pagers` | int32 | VM pager purges |

## Notes

- Name cache hit rate = `ncs_goodhits / (ncs_goodhits + ncs_miss)` — should be > 90%.
- `vfs.vnstats.num_vnodes` vs `kern.maxvnodes` shows vnode cache utilization.
- APFS keys are specific to Apple File System volumes.
- LIFS (LiveFiles) is Apple's userspace filesystem framework.
