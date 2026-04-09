# macOS Sysctl Inventory — ARM64 (Apple Silicon)

Complete inventory of macOS sysctl keys accessible via `sysctlbyname(3)`, organized by
category. Documented on macOS 15.6.1 / Darwin 24.6.0 / Apple M4.

**Total keys observed: ~1,695** (598 net, 425 kern, 256 vm, 124 hw, 113 vfs, 56 debug, 18 machdep, plus kpc/kperf/iogpu/security)

## Categories

| Directory | Category | Keys | Description |
|-----------|----------|------|-------------|
| [hw-cpu-topology/](hw-cpu-topology/) | `hw.*` | 14 | CPU counts, types, families |
| [hw-memory/](hw-memory/) | `hw.*` | 4 | Physical memory, page size |
| [hw-cache/](hw-cache/) | `hw.*` | 7 | Cache line, L1/L2 sizes, cache topology |
| [hw-perflevel/](hw-perflevel/) | `hw.perflevel*` | 18 | Apple Silicon P-core/E-core details |
| [hw-arm-features/](hw-arm-features/) | `hw.optional.arm.*` | 60 | ARM feature flags (FEAT_*, SME, AdvSIMD) |
| [hw-misc/](hw-misc/) | `hw.*` | 21 | Target type, features, ephemeral storage |
| [kern-identity/](kern-identity/) | `kern.*` | 18 | OS type/version/release, hostname, UUID |
| [kern-process-thread/](kern-process-thread/) | `kern.*` | 10 | Task/thread/file/vnode counts |
| [kern-memory-status/](kern-memory-status/) | `kern.*` | 5 | Memory pressure level, purge thresholds |
| [kern-scheduling/](kern-scheduling/) | `kern.*` | 12 | Scheduler config, workqueue, CPU checkin |
| [kern-time-boot/](kern-time-boot/) | `kern.*` | 12 | Boot/sleep/wake times, monotonic clock |
| [kern-limits/](kern-limits/) | `kern.*` | 14 | maxproc, maxfiles, maxvnodes, AIO limits |
| [kern-ipc/](kern-ipc/) | `kern.ipc.*` | 25 | IPC shared memory, semaphores, message queues |
| [kern-misc/](kern-misc/) | `kern.*` | ~315 | Coredump, crypto, dtrace, skywalk, etc. |
| [vm-pressure/](vm-pressure/) | `vm.*` | 8 | Memory pressure, darkwake mode |
| [vm-pages/](vm-pages/) | `vm.*` | 20 | Page free/speculative/purgeable/reusable counts |
| [vm-compressor/](vm-compressor/) | `vm.*` | 30 | Memory compressor stats, WK/LZ4 compression |
| [vm-pageout/](vm-pageout/) | `vm.*` | 25 | Page-out activity counters |
| [vm-swap/](vm-swap/) | `vm.*` | 5 | Swap usage, prefix, enabled |
| [vm-wire-limits/](vm-wire-limits/) | `vm.*` | 6 | Wired memory limits, kernel address range |
| [vm-misc/](vm-misc/) | `vm.*` | ~160 | Code signing, reusable, fault stats |
| [machdep/](machdep/) | `machdep.*` | 18 | CPU brand, core count, ptrauth, timing |
| [net-tcp/](net-tcp/) | `net.inet.tcp.*` | ~180 | TCP config, stats, SACK, congestion |
| [net-udp/](net-udp/) | `net.inet.udp.*` | ~50 | UDP config, stats, logging |
| [net-ip/](net-ip/) | `net.inet.ip.*` | ~60 | IP config, forwarding, fragmentation |
| [net-misc/](net-misc/) | `net.*` | ~300 | inet6, link, local, route, necp, etc. |
| [vfs/](vfs/) | `vfs.*` | 113 | Filesystem stats, APFS, vnode cache |
| [debug/](debug/) | `debug.*` | 56 | BPF, I/O throttling, scheduler debug |
| [kpc-kperf/](kpc-kperf/) | `kpc.*/kperf.*` | 10 | Performance counters, sampling limits |
| [iogpu/](iogpu/) | `iogpu.*` | 5 | GPU wired memory limits |
| [security/](security/) | `security.*` | 6 | Code signing, MAC framework |

## Type system

All values returned by `sysctlbyname` are raw bytes. Types are determined by byte count:

| Bytes | Go Type | Notes |
|-------|---------|-------|
| 4 | `int32` | Most common (~1,200 keys). Includes booleans (0/1). |
| 8 | `int64` or `uint64` | Counters, timestamps, large values. `uint64` for addresses and memory sizes. |
| Variable | `string` | Null-terminated C strings |
| 16 | `Timeval` | `struct { Sec, Usec uint64 }` — boot/sleep/wake times |
| 20 | `Clockinfo` | `struct { Hz, Tick, Tickadj, Profhz, Stathz int32 }` |
| 24 | `Loadavg` | `struct { Ldavg [3]uint32; _ uint32; Fscale int64 }` |
| 32 | `SwapUsage` | `struct { Total, Avail, Used, Flags uint64 }` |
| 80 | `[10]uint64` | `hw.cacheconfig`, `hw.cachesize` arrays |

## ARM64 type quirks

Several keys that are `int32` on x86_64 return 8 bytes on ARM64:
- `hw.pagesize`, `hw.cachelinesize`, `hw.l1icachesize`, `hw.l1dcachesize`, `hw.l2cachesize`
- `hw.tbfrequency`

Keys that don't exist on Apple Silicon (ENOENT):
- `hw.cpufrequency`, `hw.busfrequency` — Apple doesn't expose clock speeds via sysctl

`user.*` keys (20 total) are NOT accessible via `sysctlbyname` — they require MIB integer access only.
