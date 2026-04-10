//go:build ignore

// This program generates the darwin-arm64 KernelMetricRegistry textproto
// from the metrics registry. Run with: go run generate_textproto.go
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/accretional/sysctl/internal/metrics"
)

// classifyKernel returns the kernel_access_pattern for a metric.
// This encodes whether the Darwin kernel CAN change the value at runtime.
// STATIC = immutable after boot (hardware properties, CPU features, compiled-in constants).
// DYNAMIC = value can change at runtime (counters, tunables, state).
func classifyKernel(name string, cat metrics.Category) string {
	// Hardware properties — truly immutable, set by silicon/firmware
	switch cat {
	case metrics.CatHWCPU, metrics.CatHWMemory, metrics.CatHWCache,
		metrics.CatHWPerflevel, metrics.CatHWARM, metrics.CatHWMisc:
		return "STATIC"
	}

	// Kernel identity — OS version strings, UUIDs (immutable for this boot)
	if cat == metrics.CatKernIdentity {
		// hostname is writable at runtime
		if name == "kern.hostname" {
			return "DYNAMIC"
		}
		return "STATIC"
	}

	// machdep — CPU brand/topology is immutable, timing/idle is dynamic
	if cat == metrics.CatMachdep {
		switch name {
		case "machdep.cpu.brand_string", "machdep.cpu.core_count",
			"machdep.cpu.thread_count", "machdep.cpu.cores_per_package",
			"machdep.cpu.logical_per_package", "machdep.ptrauth_enabled",
			"machdep.virtual_address_size":
			return "STATIC"
		default:
			return "DYNAMIC"
		}
	}

	// Kernel limits — these are writable tunables (sysctl -w)
	if cat == metrics.CatKernLimits {
		return "DYNAMIC"
	}

	// IPC limits — writable tunables
	if cat == metrics.CatKernIPC {
		return "DYNAMIC"
	}

	// Scheduling config — writable tunables
	if cat == metrics.CatKernSched {
		return "DYNAMIC"
	}

	// Kernel misc — mix of boot-time flags and writable tunables
	if cat == metrics.CatKernMisc {
		// These are genuinely immutable boot-time properties
		switch name {
		case "kern.hv_support", "kern.secure_kernel", "kern.safeboot",
			"kern.slide", "kern.stack_size", "kern.stack_depth_max":
			return "STATIC"
		default:
			return "DYNAMIC"
		}
	}

	// Kernel time — boottime is set once, everything else changes
	if cat == metrics.CatKernTime {
		if name == "kern.boottime" || name == "kern.clockrate" {
			return "STATIC"
		}
		return "DYNAMIC"
	}

	// Memory status — all can change at runtime
	if cat == metrics.CatKernMemory {
		return "DYNAMIC"
	}

	// Process/thread counters
	if cat == metrics.CatKernProcess {
		return "DYNAMIC"
	}

	// All VM categories — dynamic kernel counters and tunables
	switch cat {
	case metrics.CatVMPressure, metrics.CatVMPages, metrics.CatVMPageout,
		metrics.CatVMCompressor, metrics.CatVMSwap, metrics.CatVMWire,
		metrics.CatVMMisc:
		return "DYNAMIC"
	}

	// Network — TCP/UDP counters are dynamic, config tunables are also writable
	switch cat {
	case metrics.CatNetTCP, metrics.CatNetUDP, metrics.CatNetIP, metrics.CatNetMisc:
		return "DYNAMIC"
	}

	// VFS — all dynamic
	if cat == metrics.CatVFS {
		return "DYNAMIC"
	}

	// Debug — writable tunables
	if cat == metrics.CatDebug {
		return "DYNAMIC"
	}

	// KPC/kperf — runtime config
	if cat == metrics.CatKPC {
		return "DYNAMIC"
	}

	// IOGPU — runtime-configurable limits
	if cat == metrics.CatIOGPU {
		return "DYNAMIC"
	}

	// Security counters — dynamic
	if cat == metrics.CatSecurity {
		return "DYNAMIC"
	}

	// Computed
	if cat == metrics.CatComputed {
		return "DYNAMIC"
	}

	return "DYNAMIC"
}

// classifyRecommended returns the recommended_access_pattern and TTL.
// This is independent of kernel_access_pattern — it encodes what monitoring
// clients would reasonably want in terms of freshness.
func classifyRecommended(name string, cat metrics.Category) (string, string) {
	// --- STATIC: values that don't change in practice ---

	// Hardware — never changes
	switch cat {
	case metrics.CatHWCPU, metrics.CatHWMemory, metrics.CatHWCache,
		metrics.CatHWPerflevel, metrics.CatHWARM, metrics.CatHWMisc:
		return "STATIC", ""
	}

	// Kernel identity — doesn't change within a boot
	if cat == metrics.CatKernIdentity {
		return "STATIC", ""
	}

	// machdep CPU identification — doesn't change
	if cat == metrics.CatMachdep {
		switch name {
		case "machdep.cpu.brand_string", "machdep.cpu.core_count",
			"machdep.cpu.thread_count", "machdep.cpu.cores_per_package",
			"machdep.cpu.logical_per_package", "machdep.ptrauth_enabled",
			"machdep.virtual_address_size":
			return "STATIC", ""
		}
	}

	// Strings that are effectively immutable at runtime.
	switch name {
	case "kern.sched", "vm.swapfileprefix":
		return "STATIC", ""
	}

	// Boot time, clockrate — set once
	if name == "kern.boottime" || name == "kern.clockrate" {
		return "STATIC", ""
	}

	// --- POLLED@10s: fast-changing counters ---

	// Computed metrics aggregate dynamic sources
	if cat == metrics.CatComputed {
		return "POLLED", `ttl { seconds: 10 }`
	}

	// VM page state, pageout, pressure
	fastPoll := []metrics.Category{
		metrics.CatVMPages, metrics.CatVMPageout, metrics.CatVMPressure,
		metrics.CatKernProcess, metrics.CatSecurity,
	}
	for _, fc := range fastPoll {
		if cat == fc {
			return "POLLED", `ttl { seconds: 10 }`
		}
	}

	// Compressor counters — 10s
	if cat == metrics.CatVMCompressor {
		// Mode and segment limit change rarely but are tunables
		if name == "vm.compressor_mode" || name == "vm.compressor_segment_limit" {
			return "POLLED", `ttl { seconds: 60 }`
		}
		return "POLLED", `ttl { seconds: 10 }`
	}

	// Memory status level — 10s
	if name == "kern.memorystatus_level" {
		return "POLLED", `ttl { seconds: 10 }`
	}

	// Memory purge thresholds — tunables, rarely changed
	if cat == metrics.CatKernMemory {
		return "POLLED", `ttl { seconds: 60 }`
	}

	// Network connection counts — 10s
	if name == "net.inet.tcp.pcbcount" || name == "net.inet.udp.pcbcount" ||
		name == "net.local.pcbcount" || name == "net.soflow.count" ||
		name == "net.inet.mptcp.pcbcount" ||
		name == "net.inet.tcp.sack_globalholes" ||
		name == "net.inet.tcp.cubic_sockets" {
		return "POLLED", `ttl { seconds: 10 }`
	}

	// VFS stats — 10s
	if cat == metrics.CatVFS {
		return "POLLED", `ttl { seconds: 10 }`
	}

	// VM misc (cs_blob, shared_region, loadavg) — 10s
	if cat == metrics.CatVMMisc {
		return "POLLED", `ttl { seconds: 10 }`
	}

	// --- POLLED@30s: moderately changing values ---

	// Swap usage — 30s
	if name == "vm.swapusage" {
		return "POLLED", `ttl { seconds: 30 }`
	}

	// Swap config — rarely changes
	if cat == metrics.CatVMSwap {
		return "POLLED", `ttl { seconds: 60 }`
	}

	// Wire limits and violation counters — 60s
	if cat == metrics.CatVMWire {
		return "POLLED", `ttl { seconds: 60 }`
	}

	// Timing counters — 30s
	if cat == metrics.CatKernTime {
		return "POLLED", `ttl { seconds: 30 }`
	}

	// machdep timing/idle — 30s
	if cat == metrics.CatMachdep {
		return "POLLED", `ttl { seconds: 30 }`
	}

	// --- POLLED@60s: slow-changing config/tunables ---

	// Kernel limits — rarely changed tunables
	if cat == metrics.CatKernLimits {
		return "POLLED", `ttl { seconds: 60 }`
	}

	// IPC limits — rarely changed
	if cat == metrics.CatKernIPC {
		return "POLLED", `ttl { seconds: 60 }`
	}

	// Scheduling config — rarely changed
	if cat == metrics.CatKernSched {
		return "POLLED", `ttl { seconds: 60 }`
	}

	// Kernel misc — boot flags and tunables
	if cat == metrics.CatKernMisc {
		return "POLLED", `ttl { seconds: 60 }`
	}

	// Network config tunables — rarely changed
	if cat == metrics.CatNetTCP || cat == metrics.CatNetUDP ||
		cat == metrics.CatNetIP || cat == metrics.CatNetMisc {
		return "POLLED", `ttl { seconds: 60 }`
	}

	// Debug/KPC/IOGPU — config tunables
	if cat == metrics.CatDebug || cat == metrics.CatKPC || cat == metrics.CatIOGPU {
		return "POLLED", `ttl { seconds: 60 }`
	}

	return "DYNAMIC", ""
}

func main() {
	var b strings.Builder
	b.WriteString(`# KernelMetricRegistry for Apple Silicon macOS
# Generated from internal/metrics/registry.go
# Darwin 24.6.0 / xnu-11417 / ARM64
#
# kernel_access_pattern: whether the kernel CAN change this value at runtime
# recommended_access_pattern: what monitoring clients should use for freshness

os_registry: "darwin-arm64"
os_version: "Darwin 24.6.0 / xnu-11417"

`)

	lastCat := metrics.Category("")
	for _, info := range metrics.Known {
		if info.Category != lastCat {
			if lastCat != "" {
				b.WriteString("\n")
			}
			b.WriteString(fmt.Sprintf("# --- %s ---\n", info.Category))
			lastCat = info.Category
		}

		kernelPattern := classifyKernel(info.Name, info.Category)
		recPattern, recTTL := classifyRecommended(info.Name, info.Category)

		b.WriteString("metrics {\n")
		b.WriteString(fmt.Sprintf("  name: %q\n", info.Name))

		// kernel_access_pattern
		b.WriteString(fmt.Sprintf("  kernel_access_pattern { pattern: %s }\n", kernelPattern))

		// recommended_access_pattern
		if recTTL != "" {
			b.WriteString(fmt.Sprintf("  recommended_access_pattern { pattern: %s %s }\n", recPattern, recTTL))
		} else {
			b.WriteString(fmt.Sprintf("  recommended_access_pattern { pattern: %s }\n", recPattern))
		}

		b.WriteString("}\n")
	}

	if err := os.WriteFile("24.6.0.textproto", []byte(b.String()), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("wrote 24.6.0.textproto (%d metrics)\n", len(metrics.Known))
}
