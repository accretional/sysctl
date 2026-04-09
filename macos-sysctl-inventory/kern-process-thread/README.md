# kern — Process & Thread Counts

Dynamic counts of tasks, threads, files, and vnodes. Key performance indicators for system load.

| Key | Bytes | Type | Example | Description |
|-----|-------|------|---------|-------------|
| `kern.num_tasks` | 4 | int32 | ~400 | Current number of tasks (processes) |
| `kern.num_threads` | 4 | int32 | ~2000 | Current number of threads |
| `kern.num_taskthreads` | 4 | int32 | ~1800 | Threads belonging to tasks |
| `kern.num_files` | 4 | int32 | ~8000 | Open file descriptors |
| `kern.num_vnodes` | 4 | int32 | ~50000 | Active vnodes |
| `kern.num_recycledvnodes` | 8 | int64 | 0 | Recycled vnode count |
| `kern.num_static_scalable_counters` | 8 | int64 | | Static scalable counter count |
| `kern.free_vnodes` | 4 | int32 | | Free vnodes available |
| `kern.procname` | string | | Current process name |
| `kern.threadname` | raw(63) | | Current thread name |

## Notes

- These are **live counters** — they change on every read.
- `kern.num_tasks` ≈ number of processes. Compare against `kern.maxproc` for saturation.
- `kern.num_files` ≈ open FDs system-wide. Compare against `kern.maxfiles`.
- `kern.num_vnodes` / `kern.maxvnodes` indicates vnode cache pressure.
- `kern.procname` and `kern.threadname` return the name of the calling process/thread (the sysctl reader itself).
