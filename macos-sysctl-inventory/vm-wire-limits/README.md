# vm — Wire Limits & Kernel Address Space

Wired memory limits and kernel virtual address range.

| Key | Bytes | Type | Example | Description |
|-----|-------|------|---------|-------------|
| `vm.global_user_wire_limit` | 8 | uint64 | | Max user-wired memory (bytes) |
| `vm.user_wire_limit` | 8 | uint64 | | Per-task user-wire limit |
| `vm.global_no_user_wire_amount` | 8 | uint64 | | Reserved non-wirable memory |
| `vm.add_wire_count_over_global_limit` | 8 | int64 | | Wire attempts exceeding global limit |
| `vm.add_wire_count_over_user_limit` | 8 | int64 | | Wire attempts exceeding user limit |
| `vm.vm_min_kernel_address` | 8 | uint64 | | Kernel address space start |
| `vm.vm_max_kernel_address` | 8 | uint64 | | Kernel address space end |

## Notes

- Wired memory cannot be paged out or compressed — it stays in physical RAM.
- `add_wire_count_over_*_limit` > 0 means applications are hitting wired memory limits.
- Kernel address range is relevant for KASLR and virtual memory layout analysis.
