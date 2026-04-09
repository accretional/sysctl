# kern — Miscellaneous

Remaining `kern.*` keys (~315) covering a wide range of kernel subsystems. Grouped by area.

## Hypervisor (kern.hv.*)

| Key | Type | Example | Description |
|-----|------|---------|-------------|
| `kern.hv_support` | int32 | 1 | Hypervisor framework supported |
| `kern.hv.supported` | int32 | 1 | HV API available |
| `kern.hv_vmm_present` | int32 | 0 | VMM currently active |
| `kern.hv_disable` | int32 | 0 | HV disabled |
| `kern.hv.ipa_size_16k` | uint64 | 4398046511104 | IPA size for 16K pages |
| `kern.hv.ipa_size_4k` | uint64 | 1099511627776 | IPA size for 4K pages |
| `kern.hv.max_address_spaces` | int32 | 128 | Max address spaces |

## Crypto (kern.crypto.*)

Random number generator and entropy subsystem config.

## DTrace (kern.dtrace.*)

DTrace configuration and buffer sizes.

## Skywalk (kern.skywalk.*)

Apple's networking stack configuration (Skywalk/flowswitch/channel).

## Timer Coalescing (kern.timer_coalesce_*)

Power management timer coalescing parameters for background, foreground, and kernel timers.
Each tier has `_ns_max` (max coalescing window) and `_scale` (scaling factor).

## CPC (kern.cpc.*)

CPU performance counter configuration.

## Monotonic (kern.monotonic.*)

| Key | Type | Description |
|-----|------|-------------|
| `kern.monotonic.supported` | int32 | Monotonic counters supported |
| `kern.monotonic.task_thread_counting` | int32 | Per-task/thread counting enabled |
| `kern.monotonic.pmis` | int64 | PMI count (2×uint32 packed) |
| `kern.monotonic.retrograde_updates` | int64 | Counter retrograde events |

## Pervasive Energy (kern.pervasive_energy.*)

Energy attribution subsystem.

## POSIX (kern.posix.*)

POSIX compliance settings (saved IDs, set group, job control).

## Notes

- Most keys in this category are configuration knobs or debug counters.
- Many are read-only; some are tunable with root.
- The hypervisor keys are relevant for VM workloads (Docker, UTM, etc.).
