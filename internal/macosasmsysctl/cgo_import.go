//go:build darwin

package macosasmsysctl

// Dynamic imports from libSystem.B.dylib — resolved by the linker.
// These provide the symbols that our assembly trampolines jump to.

//go:cgo_import_dynamic libc_sysctl sysctl "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_sysctlbyname sysctlbyname "/usr/lib/libSystem.B.dylib"
