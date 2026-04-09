//go:build darwin

// Package metrics defines the known sysctl metrics and their types.
package metrics

// ValueType describes how to interpret a sysctl value.
type ValueType string

const (
	TypeString ValueType = "string"
	TypeUint32 ValueType = "uint32"
	TypeUint64 ValueType = "uint64"
	TypeInt32  ValueType = "int32"
	TypeInt64  ValueType = "int64"
	TypeRaw    ValueType = "raw"
)

// Info describes a known sysctl metric.
type Info struct {
	Name        string
	Description string
	Type        ValueType
}

// Known is the registry of sysctl metrics we support.
var Known = []Info{
	// Kernel
	{Name: "kern.ostype", Description: "Operating system type", Type: TypeString},
	{Name: "kern.osrelease", Description: "OS release version string", Type: TypeString},
	{Name: "kern.version", Description: "Kernel version string", Type: TypeString},
	{Name: "kern.hostname", Description: "System hostname", Type: TypeString},
	{Name: "kern.osrevision", Description: "OS revision", Type: TypeInt32},
	{Name: "kern.maxproc", Description: "Maximum number of processes", Type: TypeInt32},
	{Name: "kern.maxfiles", Description: "Maximum number of open files", Type: TypeInt32},
	{Name: "kern.boottime", Description: "System boot time", Type: TypeRaw},

	// Hardware
	{Name: "hw.ncpu", Description: "Number of CPUs", Type: TypeInt32},
	{Name: "hw.activecpu", Description: "Number of active CPUs", Type: TypeUint32},
	{Name: "hw.memsize", Description: "Physical memory size in bytes", Type: TypeUint64},
	{Name: "hw.pagesize", Description: "System page size", Type: TypeInt64},
	{Name: "hw.cachelinesize", Description: "CPU cache line size", Type: TypeInt64},
	{Name: "hw.l1icachesize", Description: "L1 instruction cache size", Type: TypeInt64},
	{Name: "hw.l1dcachesize", Description: "L1 data cache size", Type: TypeInt64},
	{Name: "hw.l2cachesize", Description: "L2 cache size", Type: TypeInt64},

	// macOS-specific hardware
	{Name: "machdep.cpu.brand_string", Description: "CPU brand string", Type: TypeString},
	{Name: "machdep.cpu.core_count", Description: "Physical CPU core count", Type: TypeInt32},
	{Name: "machdep.cpu.thread_count", Description: "Logical CPU thread count", Type: TypeInt32},

	// Performance / VM
	{Name: "vm.loadavg", Description: "System load averages", Type: TypeRaw},
	{Name: "vm.swapusage", Description: "Swap usage statistics", Type: TypeRaw},

	// Network
	{Name: "net.inet.tcp.mssdflt", Description: "Default TCP MSS", Type: TypeInt32},
	{Name: "net.inet.tcp.keepidle", Description: "TCP keep-alive idle time", Type: TypeInt32},
}

// ByName returns the Info for a known metric, or nil if unknown.
func ByName(name string) *Info {
	for i := range Known {
		if Known[i].Name == name {
			return &Known[i]
		}
	}
	return nil
}
