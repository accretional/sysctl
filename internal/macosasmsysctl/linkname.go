//go:build darwin

package macosasmsysctl

import _ "unsafe"

// Link to the runtime's syscall dispatch on darwin.
// This is the same mechanism used by golang.org/x/sys/unix.

//go:linkname syscall_syscall6 syscall.syscall6
