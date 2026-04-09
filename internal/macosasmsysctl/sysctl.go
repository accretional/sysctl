//go:build darwin

// Package macosasmsysctl provides direct access to macOS sysctl via
// ARM64 assembly trampolines into libSystem. No cgo, no external deps.
package macosasmsysctl

import (
	"encoding/binary"
	"fmt"
	"unsafe"
)

// Errno is a syscall error number.
type Errno uintptr

func (e Errno) Error() string {
	return fmt.Sprintf("errno %d", int(e))
}

// Low-level declarations — implemented in assembly + linked via cgo_import_dynamic.
func syscall_syscall6(fn, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err Errno)

// Trampoline address variables — populated by the linker from assembly.
var libc_sysctl_trampoline_addr uintptr
var libc_sysctlbyname_trampoline_addr uintptr

// rawSysctl calls the sysctl(3) function via our assembly trampoline.
// mib is the Management Information Base path, e.g. {CTL_HW, HW_MEMSIZE}.
func rawSysctl(mib []uint32, old *byte, oldlen *uintptr, newp *byte, newlen uintptr) error {
	var mibp unsafe.Pointer
	if len(mib) > 0 {
		mibp = unsafe.Pointer(&mib[0])
	}
	_, _, e := syscall_syscall6(
		libc_sysctl_trampoline_addr,
		uintptr(mibp),
		uintptr(len(mib)),
		uintptr(unsafe.Pointer(old)),
		uintptr(unsafe.Pointer(oldlen)),
		uintptr(unsafe.Pointer(newp)),
		uintptr(newlen),
	)
	if e != 0 {
		return e
	}
	return nil
}

// rawSysctlbyname calls sysctlbyname(3) via our assembly trampoline.
func rawSysctlbyname(name *byte, old *byte, oldlen *uintptr, newp *byte, newlen uintptr) error {
	_, _, e := syscall_syscall6(
		libc_sysctlbyname_trampoline_addr,
		uintptr(unsafe.Pointer(name)),
		uintptr(unsafe.Pointer(old)),
		uintptr(unsafe.Pointer(oldlen)),
		uintptr(unsafe.Pointer(newp)),
		uintptr(newlen),
		0,
	)
	if e != 0 {
		return e
	}
	return nil
}

// nametomib converts a sysctl name like "kern.hostname" to a MIB array
// using the magic sysctl {0, 3}.
func nametomib(name string) ([]uint32, error) {
	const maxName = 12 // CTL_MAXNAME
	var buf [maxName + 2]uint32
	n := uintptr(maxName) * unsafe.Sizeof(buf[0])

	nameBytes := append([]byte(name), 0)
	mib := []uint32{0, 3}
	if err := rawSysctl(mib, (*byte)(unsafe.Pointer(&buf[0])), &n, &nameBytes[0], uintptr(len(name))); err != nil {
		return nil, fmt.Errorf("nametomib %q: %w", name, err)
	}
	return buf[0 : n/unsafe.Sizeof(buf[0])], nil
}

// GetRaw returns the raw bytes for a sysctl by name.
func GetRaw(name string) ([]byte, error) {
	nameBytes := append([]byte(name), 0)

	// First call: determine size.
	var n uintptr
	if err := rawSysctlbyname(&nameBytes[0], nil, &n, nil, 0); err != nil {
		return nil, fmt.Errorf("sysctl size %q: %w", name, err)
	}
	if n == 0 {
		return nil, nil
	}

	// Second call: read value.
	buf := make([]byte, n)
	if err := rawSysctlbyname(&nameBytes[0], &buf[0], &n, nil, 0); err != nil {
		return nil, fmt.Errorf("sysctl read %q: %w", name, err)
	}
	return buf[:n], nil
}

// GetString returns a sysctl string value.
func GetString(name string) (string, error) {
	raw, err := GetRaw(name)
	if err != nil {
		return "", err
	}
	// Strip trailing NUL.
	if len(raw) > 0 && raw[len(raw)-1] == 0 {
		raw = raw[:len(raw)-1]
	}
	return string(raw), nil
}

// GetUint32 returns a sysctl uint32 value.
func GetUint32(name string) (uint32, error) {
	raw, err := GetRaw(name)
	if err != nil {
		return 0, err
	}
	if len(raw) != 4 {
		return 0, fmt.Errorf("sysctl %q: expected 4 bytes, got %d", name, len(raw))
	}
	return binary.LittleEndian.Uint32(raw), nil
}

// GetUint64 returns a sysctl uint64 value.
func GetUint64(name string) (uint64, error) {
	raw, err := GetRaw(name)
	if err != nil {
		return 0, err
	}
	if len(raw) != 8 {
		return 0, fmt.Errorf("sysctl %q: expected 8 bytes, got %d", name, len(raw))
	}
	return binary.LittleEndian.Uint64(raw), nil
}

// GetInt32 returns a sysctl int32 value.
func GetInt32(name string) (int32, error) {
	v, err := GetUint32(name)
	return int32(v), err
}

// GetInt64 returns a sysctl int64 value.
func GetInt64(name string) (int64, error) {
	v, err := GetUint64(name)
	return int64(v), err
}
