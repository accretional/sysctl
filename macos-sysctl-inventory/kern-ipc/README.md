# kern — IPC (Inter-Process Communication)

System V IPC configuration under `kern.ipc.*` and `kern.sysv.*`.

## kern.ipc.* — Socket & Pipe Buffers

| Key | Type | Example | Description |
|-----|------|---------|-------------|
| `kern.ipc.maxsockbuf` | 4 | int32 | Max socket buffer size |
| `kern.ipc.sockbuf_waste_factor` | 4 | int32 | Socket buffer waste factor |
| `kern.ipc.maxsockets` | 4 | int32 | Max sockets |
| `kern.ipc.somaxconn` | 4 | int32 | Max listen backlog |
| `kern.ipc.nmbclusters` | 4 | int32 | Network memory buffer clusters |
| `kern.ipc.soqlimitcompat` | 4 | int32 | Socket queue limit compat |
| `kern.ipc.njcl` | 4 | int32 | Jumbo clusters |
| `kern.ipc.njclbytes` | 4 | int32 | Jumbo cluster bytes |
| `kern.ipc.sbmb_cnt` | 4 | int32 | Socket buffer mbuf count |
| `kern.ipc.sbmb_cnt_peak` | 4 | int32 | Peak socket buffer mbuf count |
| `kern.ipc.sbmb_cnt_floor` | 4 | int32 | Floor socket buffer mbuf count |
| `kern.ipc.mb_stat` | raw | | Mbuf statistics |
| `kern.ipc.io_policy` | raw | | I/O policy config |

## kern.sysv.* — System V IPC

| Key | Type | Example | Description |
|-----|------|---------|-------------|
| `kern.sysv.shmmax` | 8 | int64 | Max shared memory segment size |
| `kern.sysv.shmmin` | 8 | int64 | Min shared memory segment size |
| `kern.sysv.shmmni` | 8 | int64 | Max shared memory identifiers |
| `kern.sysv.shmseg` | 8 | int64 | Max shared memory segments per process |
| `kern.sysv.shmall` | 8 | int64 | Max shared memory pages |
| `kern.sysv.semmni` | 4 | int32 | Max semaphore identifiers |
| `kern.sysv.semmns` | 4 | int32 | Max semaphores system-wide |
| `kern.sysv.semmnu` | 4 | int32 | Max semaphore undo entries |
| `kern.sysv.semmsl` | 4 | int32 | Max semaphores per set |
| `kern.sysv.semume` | 4 | int32 | Max undo entries per process |

## Notes

- `kern.ipc.somaxconn` controls TCP listen backlog — critical for high-connection servers.
- `kern.ipc.maxsockbuf` limits per-socket buffer allocation.
- System V IPC is mostly legacy but still used by some database engines (PostgreSQL shared buffers).
