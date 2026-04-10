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

// writableTunables are metrics confirmed writable via `sysctl -aW` on Darwin
// ARM64 that represent admin-configurable tunables (not counters/gauges).
// These change only by explicit admin or OS action, not continuously.
var writableTunables = map[string]bool{
	// Kernel limits
	"kern.maxproc":          true,
	"kern.maxprocperuid":    true,
	"kern.maxfiles":         true,
	"kern.maxfilesperproc":  true,
	"kern.maxvnodes":        true,
	"kern.maxnbuf":          true,
	"kern.aiomax":           true,
	"kern.aioprocmax":       true,
	"kern.coredump":         true,

	// IPC limits
	"kern.ipc.maxsockbuf": true,
	"kern.ipc.somaxconn":  true,

	// Scheduling config
	"kern.cpu_checkin_interval":    true,
	"kern.wq_max_threads":         true,
	"kern.wq_max_constrained_threads": true,

	// Memory purge thresholds
	"kern.memorystatus_purge_on_warning":  true,
	"kern.memorystatus_purge_on_urgent":   true,
	"kern.memorystatus_purge_on_critical": true,

	// Misc writable tunables
	"kern.hostname":      true,
	"kern.hibernatemode": true,

	// VM wire limits
	"vm.global_user_wire_limit":    true,
	"vm.user_wire_limit":           true,
	"vm.global_no_user_wire_amount": true,

	// Network TCP config
	"net.inet.tcp.mssdflt":            true,
	"net.inet.tcp.v6mssdflt":          true,
	"net.inet.tcp.keepidle":           true,
	"net.inet.tcp.keepintvl":          true,
	"net.inet.tcp.keepcnt":            true,
	"net.inet.tcp.sendspace":          true,
	"net.inet.tcp.recvspace":          true,
	"net.inet.tcp.sack":              true,
	"net.inet.tcp.delayed_ack":        true,
	"net.inet.tcp.fastopen":           true,
	"net.inet.tcp.ecn_initiate_out":   true,
	"net.inet.tcp.ecn_negotiate_in":   true,
	"net.inet.tcp.blackhole":          true,
	"net.inet.tcp.always_keepalive":   true,
	"net.inet.tcp.sack_globalmaxholes": true,

	// Network UDP config
	"net.inet.udp.maxdgram":  true,
	"net.inet.udp.recvspace": true,
	"net.inet.udp.checksum":  true,

	// Network IP config
	"net.inet.ip.forwarding":     true,
	"net.inet.ip.ttl":            true,
	"net.inet.ip.maxfragpackets": true,

	// Debug
	"debug.lowpri_throttle_enabled": true,
	"debug.bpf_bufsize":             true,
	"debug.bpf_maxbufsize":          true,

	// IOGPU
	"iogpu.wired_limit_mb": true,
	"iogpu.wired_lwm_mb":   true,
	"iogpu.dynamic_lwm":    true,

	// KPC
	"kperf.debug_level": true,

	// machdep
	"machdep.user_idle_level": true,
}

// classifyKernel returns the kernel_access_pattern for a metric.
// STATIC = immutable after boot (hardware properties, CPU features, compiled-in constants).
// CONSTRAINED = mutable via admin/OS configuration (writable tunables).
// DYNAMIC = value changes at runtime (counters, gauges, state).
func classifyKernel(name string, cat metrics.Category) string {
	// Hardware properties — truly immutable, set by silicon/firmware
	switch cat {
	case metrics.CatHWCPU, metrics.CatHWMemory, metrics.CatHWCache,
		metrics.CatHWPerflevel, metrics.CatHWARM, metrics.CatHWMisc:
		return "STATIC"
	}

	// Kernel identity — OS version strings, UUIDs (immutable for this boot)
	if cat == metrics.CatKernIdentity {
		if writableTunables[name] {
			return "CONSTRAINED"
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
			if writableTunables[name] {
				return "CONSTRAINED"
			}
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

	// Kernel misc — boot-time flags are static
	if cat == metrics.CatKernMisc {
		switch name {
		case "kern.hv_support", "kern.secure_kernel", "kern.safeboot",
			"kern.slide", "kern.stack_size", "kern.stack_depth_max":
			return "STATIC"
		}
	}

	// Check writable tunables map for all remaining metrics
	if writableTunables[name] {
		return "CONSTRAINED"
	}

	// Anything not STATIC or CONSTRAINED is DYNAMIC
	return "DYNAMIC"
}

// classifyRecommended returns the recommended_access_pattern and TTL.
// This is independent of kernel_access_pattern — it encodes what monitoring
// clients would reasonably want in terms of freshness.
//
// STATIC: never changes in practice.
// POLLED: dynamic counters/gauges that change continuously.
// CONSTRAINED: writable tunables that change only by admin action.
func classifyRecommended(name string, cat metrics.Category) (string, string) {
	// --- STATIC: values that don't change in practice ---

	// Hardware — never changes
	switch cat {
	case metrics.CatHWCPU, metrics.CatHWMemory, metrics.CatHWCache,
		metrics.CatHWPerflevel, metrics.CatHWARM, metrics.CatHWMisc:
		return "STATIC", ""
	}

	// Kernel identity — doesn't change within a boot (hostname is writable
	// but rare enough to be CONSTRAINED, handled below via writableTunables)
	if cat == metrics.CatKernIdentity {
		if writableTunables[name] {
			return "CONSTRAINED", `ttl { seconds: 60 }`
		}
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

	// --- CONSTRAINED: writable tunables (change by admin action only) ---
	// Check early so tunables don't fall through to POLLED counters.
	if writableTunables[name] {
		return "CONSTRAINED", `ttl { seconds: 60 }`
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
		return "POLLED", `ttl { seconds: 10 }`
	}

	// Memory status level — 10s
	if name == "kern.memorystatus_level" {
		return "POLLED", `ttl { seconds: 10 }`
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

	// Swap config — 60s (counters, not tunables)
	if cat == metrics.CatVMSwap {
		return "POLLED", `ttl { seconds: 60 }`
	}

	// Wire violation counters (limits are CONSTRAINED above) — 60s
	if cat == metrics.CatVMWire {
		return "POLLED", `ttl { seconds: 60 }`
	}

	// Timing counters — 30s
	if cat == metrics.CatKernTime {
		return "POLLED", `ttl { seconds: 30 }`
	}

	// machdep timing/idle (non-tunable remainder) — 30s
	if cat == metrics.CatMachdep {
		return "POLLED", `ttl { seconds: 30 }`
	}

	// --- POLLED@60s: slow-changing non-tunable values ---

	// Kernel limits (non-tunable remainder, if any)
	if cat == metrics.CatKernLimits {
		return "POLLED", `ttl { seconds: 60 }`
	}

	// IPC (non-tunable remainder)
	if cat == metrics.CatKernIPC {
		return "POLLED", `ttl { seconds: 60 }`
	}

	// Scheduling (non-tunable remainder)
	if cat == metrics.CatKernSched {
		return "POLLED", `ttl { seconds: 60 }`
	}

	// Kernel misc — boot flags and non-tunable state
	if cat == metrics.CatKernMisc {
		return "POLLED", `ttl { seconds: 60 }`
	}

	// Memory (non-tunable remainder)
	if cat == metrics.CatKernMemory {
		return "POLLED", `ttl { seconds: 60 }`
	}

	// Network (non-tunable remainder — connection counts handled above)
	if cat == metrics.CatNetTCP || cat == metrics.CatNetUDP ||
		cat == metrics.CatNetIP || cat == metrics.CatNetMisc {
		return "POLLED", `ttl { seconds: 60 }`
	}

	// Debug/KPC/IOGPU (non-tunable remainder)
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
