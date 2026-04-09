# kern — Time & Boot

System timestamps, boot time, and monotonic clock.

## Struct types (16 bytes = timeval: sec + usec as uint64)

| Key | Bytes | Type | Description |
|-----|-------|------|-------------|
| `kern.boottime` | 16 | Timeval | System boot time (epoch seconds + microseconds) |
| `kern.sleeptime` | 16 | Timeval | Time spent sleeping |
| `kern.waketime` | 16 | Timeval | Last wake time |

## Absolute times (uint64 — mach_absolute_time units)

| Key | Bytes | Type | Description |
|-----|-------|------|-------------|
| `kern.wake_abs_time` | 8 | uint64 | Absolute time of last wake |
| `kern.sleep_abs_time` | 8 | uint64 | Absolute time when sleep began |
| `kern.useractive_abs_time` | 8 | uint64 | Total user-active absolute time |
| `kern.userinactive_abs_time` | 8 | uint64 | Total user-inactive absolute time |

## Monotonic clock

| Key | Bytes | Type | Description |
|-----|-------|------|-------------|
| `kern.monotonicclock` | 4 | int32 | Monotonic clock enabled |
| `kern.monotonicclock_usecs` | 8 | uint64 | Monotonic clock in microseconds |
| `kern.monotonicclock_rate_usecs` | 8 | int64 | Rate adjustment |
| `kern.monotoniclock_offset_usecs` | 8 | uint64 | Clock offset |

## Struct layout: Timeval

```
struct timeval64 {
    Sec  uint64  // seconds since epoch (for boottime) or duration
    Usec uint64  // microseconds
}
```

## Notes

- Convert `kern.boottime.Sec` to time.Unix() for human-readable boot time.
- Absolute times use `mach_absolute_time` units. Divide by `hw.tbfrequency` (24 MHz) for seconds.
- `kern.monotonicclock_usecs` is the wall-clock-monotonic time in microseconds.
