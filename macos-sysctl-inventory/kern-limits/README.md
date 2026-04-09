# kern — System Limits

Maximum resource limits configured in the kernel.

| Key | Bytes | Type | Example | Description |
|-----|-------|------|---------|-------------|
| `kern.maxproc` | 4 | int32 | 4000 | Max processes system-wide |
| `kern.maxprocperuid` | 4 | int32 | 2666 | Max processes per user |
| `kern.maxfiles` | 4 | int32 | 122880 | Max open files system-wide |
| `kern.maxfilesperproc` | 4 | int32 | 49152 | Max open files per process |
| `kern.maxvnodes` | 4 | int32 | 263168 | Max vnodes |
| `kern.maxnbuf` | 4 | int32 | 16384 | Max buffer cache entries |
| `kern.nbuf` | 4 | int32 | 16384 | Current buffer cache entries |
| `kern.argmax` | 4 | int32 | 1048576 | Max argument list size (bytes) |
| `kern.posix1version` | 4 | int32 | 200112 | POSIX.1 version |
| `kern.ngroups` | 4 | int32 | 16 | Max supplementary groups |
| `kern.aiomax` | 4 | int32 | 90 | Max AIO operations |
| `kern.aioprocmax` | 4 | int32 | 16 | Max AIO per process |
| `kern.aiothreads` | 4 | int32 | 4 | AIO thread count |
| `kern.coredump` | 4 | int32 | 1 | Core dumps enabled |

## Notes

- Compare `kern.num_tasks` vs `kern.maxproc` and `kern.num_files` vs `kern.maxfiles` for saturation detection.
- `kern.argmax` = 1 MB is relevant for exec() argument passing.
- `kern.maxvnodes` affects filesystem caching capacity.
