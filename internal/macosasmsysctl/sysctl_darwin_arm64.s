#include "textflag.h"

// Assembly trampolines for macOS ARM64 sysctl calls.
// These jump directly into libSystem.B.dylib symbols imported via
// //go:cgo_import_dynamic in the Go source. This is the same pattern
// used by Go's runtime and x/sys/unix — no cgo required.

// func libc_sysctl_trampoline()
TEXT libc_sysctl_trampoline<>(SB),NOSPLIT,$0-0
	JMP	libc_sysctl(SB)
GLOBL	·libc_sysctl_trampoline_addr(SB), RODATA, $8
DATA	·libc_sysctl_trampoline_addr(SB)/8, $libc_sysctl_trampoline<>(SB)

// func libc_sysctlbyname_trampoline()
TEXT libc_sysctlbyname_trampoline<>(SB),NOSPLIT,$0-0
	JMP	libc_sysctlbyname(SB)
GLOBL	·libc_sysctlbyname_trampoline_addr(SB), RODATA, $8
DATA	·libc_sysctlbyname_trampoline_addr(SB)/8, $libc_sysctlbyname_trampoline<>(SB)
