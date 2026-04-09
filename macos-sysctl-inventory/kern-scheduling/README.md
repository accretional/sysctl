# kern — Scheduling

Scheduler configuration and workqueue parameters.

| Key | Bytes | Type | Example | Description |
|-----|-------|------|---------|-------------|
| `kern.sched` | string | string | "edge" | Scheduler name |
| `kern.sched_recommended_cores` | 8 | int64 | | Bitmask of recommended cores |
| `kern.sched_allow_NO_SMT_threads` | 4 | int32 | 1 | Allow non-SMT thread placement |
| `kern.sched_rt_avoid_cpu0` | 4 | int32 | 0 | RT threads avoid CPU 0 |
| `kern.cpu_checkin_interval` | 4 | int32 | 4000 | CPU checkin interval (ms) |
| `kern.wq_max_threads` | 4 | int32 | 512 | Max workqueue threads |
| `kern.wq_max_constrained_threads` | 4 | int32 | 64 | Max constrained workqueue threads |
| `kern.wq_max_timer_interval_usecs` | 4 | int32 | | Max workqueue timer interval |
| `kern.wq_reduce_pool_window_usecs` | 4 | int32 | | Pool reduction window |
| `kern.wq_stalled_window_usecs` | 4 | int32 | | Stall detection window |
| `kern.thread_groups_supported` | 4 | int32 | 1 | Thread group scheduling support |
| `kern.direct_handoff` | 4 | int32 | 0 | Direct handoff scheduling |

## Notes

- macOS uses the "edge" scheduler on Apple Silicon.
- `kern.wq_max_threads` limits GCD/libdispatch workqueue thread count.
- `kern.thread_groups_supported` = 1 enables Apple's thread group QoS scheduling (P/E core assignment).
- `kern.sched_recommended_cores` bitmask indicates which cores the scheduler prefers for new work.
