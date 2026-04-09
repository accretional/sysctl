//go:build darwin

// Package metrics defines the known sysctl metrics and their types.
package metrics

// ValueType describes how to interpret a sysctl value.
type ValueType string

const (
	TypeString  ValueType = "string"
	TypeUint32  ValueType = "uint32"
	TypeUint64  ValueType = "uint64"
	TypeInt32   ValueType = "int32"
	TypeInt64   ValueType = "int64"
	TypeRaw     ValueType = "raw"
	TypeTimeval ValueType = "timeval" // 16 bytes: {Sec, Usec uint64}
	TypeLoadavg ValueType = "loadavg" // 24 bytes: {Ldavg [3]uint32, pad uint32, Fscale int64}
	TypeSwap    ValueType = "swap"    // 32 bytes: {Total, Avail, Used, Flags uint64}
	TypeClock   ValueType = "clock"   // 20 bytes: {Hz, Tick, Tickadj, Profhz, Stathz int32}
)

// Category groups related metrics.
type Category string

const (
	CatHWCPU         Category = "hw.cpu"
	CatHWMemory      Category = "hw.memory"
	CatHWCache       Category = "hw.cache"
	CatHWPerflevel   Category = "hw.perflevel"
	CatHWARM         Category = "hw.arm"
	CatHWMisc        Category = "hw.misc"
	CatKernIdentity  Category = "kern.identity"
	CatKernProcess   Category = "kern.process"
	CatKernMemory    Category = "kern.memory"
	CatKernSched     Category = "kern.sched"
	CatKernTime      Category = "kern.time"
	CatKernLimits    Category = "kern.limits"
	CatKernIPC       Category = "kern.ipc"
	CatKernMisc      Category = "kern.misc"
	CatVMPressure    Category = "vm.pressure"
	CatVMPages       Category = "vm.pages"
	CatVMCompressor  Category = "vm.compressor"
	CatVMPageout     Category = "vm.pageout"
	CatVMSwap        Category = "vm.swap"
	CatVMWire        Category = "vm.wire"
	CatVMMisc        Category = "vm.misc"
	CatMachdep       Category = "machdep"
	CatNetTCP        Category = "net.tcp"
	CatNetUDP        Category = "net.udp"
	CatNetIP         Category = "net.ip"
	CatNetMisc       Category = "net.misc"
	CatVFS           Category = "vfs"
	CatDebug         Category = "debug"
	CatKPC           Category = "kpc"
	CatIOGPU         Category = "iogpu"
	CatSecurity      Category = "security"
)

// Info describes a known sysctl metric.
type Info struct {
	Name        string
	Description string
	Type        ValueType
	Category    Category
}

// Known is the registry of sysctl metrics we support.
// Organized by category, comprehensive coverage of performance-relevant metrics.
var Known = []Info{
	// =========================================================================
	// hw.cpu — CPU Topology
	// =========================================================================
	{Name: "hw.ncpu", Description: "Total number of CPUs (logical)", Type: TypeInt32, Category: CatHWCPU},
	{Name: "hw.activecpu", Description: "Currently active CPUs", Type: TypeInt32, Category: CatHWCPU},
	{Name: "hw.physicalcpu", Description: "Physical CPU cores", Type: TypeInt32, Category: CatHWCPU},
	{Name: "hw.physicalcpu_max", Description: "Max physical CPU cores", Type: TypeInt32, Category: CatHWCPU},
	{Name: "hw.logicalcpu", Description: "Logical CPU cores", Type: TypeInt32, Category: CatHWCPU},
	{Name: "hw.logicalcpu_max", Description: "Max logical CPU cores", Type: TypeInt32, Category: CatHWCPU},
	{Name: "hw.packages", Description: "CPU packages", Type: TypeInt32, Category: CatHWCPU},
	{Name: "hw.nperflevels", Description: "Performance levels (P+E)", Type: TypeInt32, Category: CatHWCPU},
	{Name: "hw.cpu64bit_capable", Description: "64-bit capable", Type: TypeInt32, Category: CatHWCPU},
	{Name: "hw.cputype", Description: "Mach-O CPU type", Type: TypeInt32, Category: CatHWCPU},
	{Name: "hw.cpusubtype", Description: "Mach-O CPU subtype", Type: TypeInt32, Category: CatHWCPU},
	{Name: "hw.cpufamily", Description: "CPU family identifier", Type: TypeInt32, Category: CatHWCPU},
	{Name: "hw.cpusubfamily", Description: "CPU subfamily", Type: TypeInt32, Category: CatHWCPU},
	{Name: "hw.byteorder", Description: "Byte order (1234=LE)", Type: TypeInt32, Category: CatHWCPU},

	// =========================================================================
	// hw.memory — Physical Memory
	// =========================================================================
	{Name: "hw.memsize", Description: "Physical memory in bytes", Type: TypeUint64, Category: CatHWMemory},
	{Name: "hw.memsize_usable", Description: "Usable physical memory", Type: TypeUint64, Category: CatHWMemory},
	{Name: "hw.pagesize", Description: "System page size", Type: TypeInt64, Category: CatHWMemory},
	{Name: "hw.pagesize32", Description: "32-bit page size", Type: TypeInt64, Category: CatHWMemory},

	// =========================================================================
	// hw.cache — Cache Hierarchy
	// =========================================================================
	{Name: "hw.cachelinesize", Description: "Cache line size in bytes", Type: TypeInt64, Category: CatHWCache},
	{Name: "hw.l1icachesize", Description: "L1 instruction cache size", Type: TypeInt64, Category: CatHWCache},
	{Name: "hw.l1dcachesize", Description: "L1 data cache size", Type: TypeInt64, Category: CatHWCache},
	{Name: "hw.l2cachesize", Description: "L2 cache size", Type: TypeInt64, Category: CatHWCache},
	{Name: "hw.tbfrequency", Description: "Timebase frequency (Hz)", Type: TypeInt64, Category: CatHWCache},
	{Name: "hw.cacheconfig", Description: "CPUs per cache level [10]uint64", Type: TypeRaw, Category: CatHWCache},
	{Name: "hw.cachesize", Description: "Cache size per level [10]uint64", Type: TypeRaw, Category: CatHWCache},

	// =========================================================================
	// hw.perflevel — Apple Silicon P/E Cores
	// =========================================================================
	{Name: "hw.perflevel0.name", Description: "P-core cluster name", Type: TypeString, Category: CatHWPerflevel},
	{Name: "hw.perflevel0.physicalcpu", Description: "P-core physical CPUs", Type: TypeInt32, Category: CatHWPerflevel},
	{Name: "hw.perflevel0.physicalcpu_max", Description: "P-core max physical CPUs", Type: TypeInt32, Category: CatHWPerflevel},
	{Name: "hw.perflevel0.logicalcpu", Description: "P-core logical CPUs", Type: TypeInt32, Category: CatHWPerflevel},
	{Name: "hw.perflevel0.logicalcpu_max", Description: "P-core max logical CPUs", Type: TypeInt32, Category: CatHWPerflevel},
	{Name: "hw.perflevel0.l1icachesize", Description: "P-core L1I cache", Type: TypeInt32, Category: CatHWPerflevel},
	{Name: "hw.perflevel0.l1dcachesize", Description: "P-core L1D cache", Type: TypeInt32, Category: CatHWPerflevel},
	{Name: "hw.perflevel0.l2cachesize", Description: "P-core shared L2 cache", Type: TypeInt32, Category: CatHWPerflevel},
	{Name: "hw.perflevel0.cpusperl2", Description: "P-cores per L2", Type: TypeInt32, Category: CatHWPerflevel},
	{Name: "hw.perflevel1.name", Description: "E-core cluster name", Type: TypeString, Category: CatHWPerflevel},
	{Name: "hw.perflevel1.physicalcpu", Description: "E-core physical CPUs", Type: TypeInt32, Category: CatHWPerflevel},
	{Name: "hw.perflevel1.physicalcpu_max", Description: "E-core max physical CPUs", Type: TypeInt32, Category: CatHWPerflevel},
	{Name: "hw.perflevel1.logicalcpu", Description: "E-core logical CPUs", Type: TypeInt32, Category: CatHWPerflevel},
	{Name: "hw.perflevel1.logicalcpu_max", Description: "E-core max logical CPUs", Type: TypeInt32, Category: CatHWPerflevel},
	{Name: "hw.perflevel1.l1icachesize", Description: "E-core L1I cache", Type: TypeInt32, Category: CatHWPerflevel},
	{Name: "hw.perflevel1.l1dcachesize", Description: "E-core L1D cache", Type: TypeInt32, Category: CatHWPerflevel},
	{Name: "hw.perflevel1.l2cachesize", Description: "E-core shared L2 cache", Type: TypeInt32, Category: CatHWPerflevel},
	{Name: "hw.perflevel1.cpusperl2", Description: "E-cores per L2", Type: TypeInt32, Category: CatHWPerflevel},

	// =========================================================================
	// hw.arm — ARM Feature Flags (selected)
	// =========================================================================
	{Name: "hw.optional.arm.FEAT_AES", Description: "AES instructions", Type: TypeInt32, Category: CatHWARM},
	{Name: "hw.optional.arm.FEAT_SHA256", Description: "SHA-256 instructions", Type: TypeInt32, Category: CatHWARM},
	{Name: "hw.optional.arm.FEAT_SHA512", Description: "SHA-512 instructions", Type: TypeInt32, Category: CatHWARM},
	{Name: "hw.optional.arm.FEAT_SHA3", Description: "SHA-3 instructions", Type: TypeInt32, Category: CatHWARM},
	{Name: "hw.optional.arm.FEAT_LSE", Description: "Large System Extensions (atomics)", Type: TypeInt32, Category: CatHWARM},
	{Name: "hw.optional.arm.FEAT_LSE2", Description: "LSE v2 (128-bit atomics)", Type: TypeInt32, Category: CatHWARM},
	{Name: "hw.optional.arm.FEAT_CRC32", Description: "CRC32 instructions", Type: TypeInt32, Category: CatHWARM},
	{Name: "hw.optional.arm.FEAT_BF16", Description: "BFloat16 instructions", Type: TypeInt32, Category: CatHWARM},
	{Name: "hw.optional.arm.FEAT_I8MM", Description: "Int8 matrix multiply", Type: TypeInt32, Category: CatHWARM},
	{Name: "hw.optional.arm.FEAT_DotProd", Description: "Dot product instructions", Type: TypeInt32, Category: CatHWARM},
	{Name: "hw.optional.arm.FEAT_SME", Description: "Scalable Matrix Extension", Type: TypeInt32, Category: CatHWARM},
	{Name: "hw.optional.arm.FEAT_SME2", Description: "SME version 2", Type: TypeInt32, Category: CatHWARM},
	{Name: "hw.optional.arm.FEAT_PAuth", Description: "Pointer authentication", Type: TypeInt32, Category: CatHWARM},
	{Name: "hw.optional.arm.FEAT_BTI", Description: "Branch target identification", Type: TypeInt32, Category: CatHWARM},
	{Name: "hw.optional.arm.AdvSIMD", Description: "Advanced SIMD (NEON)", Type: TypeInt32, Category: CatHWARM},
	{Name: "hw.optional.arm.FEAT_FP16", Description: "Half-precision FP", Type: TypeInt32, Category: CatHWARM},
	{Name: "hw.optional.floatingpoint", Description: "FP support", Type: TypeInt32, Category: CatHWARM},

	// =========================================================================
	// kern.identity — System Identification
	// =========================================================================
	{Name: "kern.ostype", Description: "OS type", Type: TypeString, Category: CatKernIdentity},
	{Name: "kern.osrelease", Description: "Darwin kernel release", Type: TypeString, Category: CatKernIdentity},
	{Name: "kern.osversion", Description: "macOS build version", Type: TypeString, Category: CatKernIdentity},
	{Name: "kern.osproductversion", Description: "macOS product version", Type: TypeString, Category: CatKernIdentity},
	{Name: "kern.version", Description: "Full kernel version string", Type: TypeString, Category: CatKernIdentity},
	{Name: "kern.hostname", Description: "System hostname", Type: TypeString, Category: CatKernIdentity},
	{Name: "kern.uuid", Description: "System UUID", Type: TypeString, Category: CatKernIdentity},
	{Name: "kern.bootuuid", Description: "Boot session UUID", Type: TypeString, Category: CatKernIdentity},
	{Name: "kern.osrevision", Description: "OS revision number", Type: TypeInt32, Category: CatKernIdentity},
	{Name: "kern.osreleasetype", Description: "Release type", Type: TypeString, Category: CatKernIdentity},
	{Name: "kern.iossupportversion", Description: "iOS compatibility version", Type: TypeString, Category: CatKernIdentity},

	// =========================================================================
	// kern.process — Process & Thread Counts
	// =========================================================================
	{Name: "kern.num_tasks", Description: "Current task (process) count", Type: TypeInt32, Category: CatKernProcess},
	{Name: "kern.num_threads", Description: "Current thread count", Type: TypeInt32, Category: CatKernProcess},
	{Name: "kern.num_taskthreads", Description: "Threads belonging to tasks", Type: TypeInt32, Category: CatKernProcess},
	{Name: "kern.num_files", Description: "Open file descriptors", Type: TypeInt32, Category: CatKernProcess},
	{Name: "kern.num_vnodes", Description: "Active vnodes", Type: TypeInt32, Category: CatKernProcess},
	{Name: "kern.num_recycledvnodes", Description: "Recycled vnode count", Type: TypeInt64, Category: CatKernProcess},
	{Name: "kern.free_vnodes", Description: "Free vnodes available", Type: TypeInt32, Category: CatKernProcess},

	// =========================================================================
	// kern.memory — Memory Status
	// =========================================================================
	{Name: "kern.memorystatus_level", Description: "Memory pressure level (0-100%)", Type: TypeInt32, Category: CatKernMemory},
	{Name: "kern.memorystatus_purge_on_warning", Description: "Purge on memory warning", Type: TypeInt32, Category: CatKernMemory},
	{Name: "kern.memorystatus_purge_on_urgent", Description: "Purge on urgent memory", Type: TypeInt32, Category: CatKernMemory},
	{Name: "kern.memorystatus_purge_on_critical", Description: "Purge on critical memory", Type: TypeInt32, Category: CatKernMemory},

	// =========================================================================
	// kern.sched — Scheduling
	// =========================================================================
	{Name: "kern.sched", Description: "Scheduler name", Type: TypeString, Category: CatKernSched},
	{Name: "kern.sched_allow_NO_SMT_threads", Description: "Allow non-SMT thread placement", Type: TypeInt32, Category: CatKernSched},
	{Name: "kern.cpu_checkin_interval", Description: "CPU checkin interval (ms)", Type: TypeInt32, Category: CatKernSched},
	{Name: "kern.wq_max_threads", Description: "Max workqueue threads", Type: TypeInt32, Category: CatKernSched},
	{Name: "kern.wq_max_constrained_threads", Description: "Max constrained WQ threads", Type: TypeInt32, Category: CatKernSched},
	{Name: "kern.thread_groups_supported", Description: "Thread group QoS scheduling", Type: TypeInt32, Category: CatKernSched},

	// =========================================================================
	// kern.time — Time & Boot
	// =========================================================================
	{Name: "kern.boottime", Description: "System boot time", Type: TypeTimeval, Category: CatKernTime},
	{Name: "kern.sleeptime", Description: "Time spent sleeping", Type: TypeTimeval, Category: CatKernTime},
	{Name: "kern.waketime", Description: "Last wake time", Type: TypeTimeval, Category: CatKernTime},
	{Name: "kern.wake_abs_time", Description: "Absolute time of last wake", Type: TypeUint64, Category: CatKernTime},
	{Name: "kern.sleep_abs_time", Description: "Absolute time of sleep", Type: TypeUint64, Category: CatKernTime},
	{Name: "kern.useractive_abs_time", Description: "User-active absolute time", Type: TypeUint64, Category: CatKernTime},
	{Name: "kern.userinactive_abs_time", Description: "User-inactive absolute time", Type: TypeUint64, Category: CatKernTime},
	{Name: "kern.monotonicclock_usecs", Description: "Monotonic clock (usecs)", Type: TypeUint64, Category: CatKernTime},
	{Name: "kern.clockrate", Description: "Clock rate info", Type: TypeClock, Category: CatKernTime},

	// =========================================================================
	// kern.limits — System Limits
	// =========================================================================
	{Name: "kern.maxproc", Description: "Max processes", Type: TypeInt32, Category: CatKernLimits},
	{Name: "kern.maxprocperuid", Description: "Max processes per user", Type: TypeInt32, Category: CatKernLimits},
	{Name: "kern.maxfiles", Description: "Max open files", Type: TypeInt32, Category: CatKernLimits},
	{Name: "kern.maxfilesperproc", Description: "Max open files per process", Type: TypeInt32, Category: CatKernLimits},
	{Name: "kern.maxvnodes", Description: "Max vnodes", Type: TypeInt32, Category: CatKernLimits},
	{Name: "kern.maxnbuf", Description: "Max buffer cache entries", Type: TypeInt32, Category: CatKernLimits},
	{Name: "kern.argmax", Description: "Max argument list size", Type: TypeInt32, Category: CatKernLimits},
	{Name: "kern.ngroups", Description: "Max supplementary groups", Type: TypeInt32, Category: CatKernLimits},
	{Name: "kern.aiomax", Description: "Max AIO operations", Type: TypeInt32, Category: CatKernLimits},
	{Name: "kern.aioprocmax", Description: "Max AIO per process", Type: TypeInt32, Category: CatKernLimits},
	{Name: "kern.coredump", Description: "Core dumps enabled", Type: TypeInt32, Category: CatKernLimits},

	// =========================================================================
	// kern.ipc — IPC
	// =========================================================================
	{Name: "kern.ipc.maxsockbuf", Description: "Max socket buffer size", Type: TypeInt32, Category: CatKernIPC},
	{Name: "kern.ipc.somaxconn", Description: "Max listen backlog", Type: TypeInt32, Category: CatKernIPC},
	{Name: "kern.ipc.nmbclusters", Description: "Network mbuf clusters", Type: TypeInt32, Category: CatKernIPC},

	// =========================================================================
	// kern.misc — Miscellaneous
	// =========================================================================
	{Name: "kern.hv_support", Description: "Hypervisor support", Type: TypeInt32, Category: CatKernMisc},
	{Name: "kern.hv_vmm_present", Description: "VMM currently active", Type: TypeInt32, Category: CatKernMisc},
	{Name: "kern.secure_kernel", Description: "Secure kernel mode", Type: TypeInt32, Category: CatKernMisc},
	{Name: "kern.safeboot", Description: "Safe boot mode", Type: TypeInt32, Category: CatKernMisc},
	{Name: "kern.hibernatemode", Description: "Hibernate mode", Type: TypeInt32, Category: CatKernMisc},
	{Name: "kern.hibernatecount", Description: "Hibernate count", Type: TypeInt32, Category: CatKernMisc},
	{Name: "kern.slide", Description: "Kernel ASLR slide", Type: TypeInt32, Category: CatKernMisc},
	{Name: "kern.stack_size", Description: "Default stack size", Type: TypeInt32, Category: CatKernMisc},
	{Name: "kern.stack_depth_max", Description: "Max stack depth", Type: TypeInt32, Category: CatKernMisc},

	// =========================================================================
	// vm.pressure — Memory Pressure
	// =========================================================================
	{Name: "vm.memory_pressure", Description: "Memory pressure (1=normal, 2=warn, 4=critical)", Type: TypeInt32, Category: CatVMPressure},
	{Name: "vm.page_free_wanted", Description: "Pages wanted by free list", Type: TypeInt32, Category: CatVMPressure},
	{Name: "vm.vm_page_free_target", Description: "Target free page count", Type: TypeInt32, Category: CatVMPressure},

	// =========================================================================
	// vm.pages — Page Counts
	// =========================================================================
	{Name: "vm.pages", Description: "Total physical pages", Type: TypeInt32, Category: CatVMPages},
	{Name: "vm.pagesize", Description: "VM page size", Type: TypeInt32, Category: CatVMPages},
	{Name: "vm.page_free_count", Description: "Free pages", Type: TypeInt32, Category: CatVMPages},
	{Name: "vm.page_speculative_count", Description: "Speculative (prefetched) pages", Type: TypeInt32, Category: CatVMPages},
	{Name: "vm.page_cleaned_count", Description: "Cleaned pages", Type: TypeInt32, Category: CatVMPages},
	{Name: "vm.page_pageable_internal_count", Description: "Pageable internal (anon) pages", Type: TypeInt32, Category: CatVMPages},
	{Name: "vm.page_pageable_external_count", Description: "Pageable external (file) pages", Type: TypeInt32, Category: CatVMPages},
	{Name: "vm.page_purgeable_count", Description: "Purgeable pages", Type: TypeInt32, Category: CatVMPages},
	{Name: "vm.page_purgeable_wired_count", Description: "Purgeable wired pages", Type: TypeInt32, Category: CatVMPages},
	{Name: "vm.page_reusable_count", Description: "Reusable pages", Type: TypeInt32, Category: CatVMPages},
	{Name: "vm.page_realtime_count", Description: "Realtime pages", Type: TypeInt32, Category: CatVMPages},
	{Name: "vm.vm_page_external_count", Description: "External (file-cache) pages", Type: TypeInt32, Category: CatVMPages},
	{Name: "vm.vm_page_background_count", Description: "Background pages", Type: TypeInt32, Category: CatVMPages},
	{Name: "vm.vm_page_background_internal_count", Description: "Background internal pages", Type: TypeInt32, Category: CatVMPages},
	{Name: "vm.vm_page_background_external_count", Description: "Background external pages", Type: TypeInt32, Category: CatVMPages},
	{Name: "vm.vm_page_background_target", Description: "Background page target", Type: TypeInt32, Category: CatVMPages},
	{Name: "vm.pages_grabbed", Description: "Total pages grabbed from free list", Type: TypeInt64, Category: CatVMPages},

	// =========================================================================
	// vm.compressor — Memory Compressor
	// =========================================================================
	{Name: "vm.compressor_mode", Description: "Compressor mode (7=compress+swap)", Type: TypeInt32, Category: CatVMCompressor},
	{Name: "vm.compressor_is_active", Description: "Compressor active", Type: TypeInt32, Category: CatVMCompressor},
	{Name: "vm.compressor_available", Description: "Compressor available", Type: TypeInt32, Category: CatVMCompressor},
	{Name: "vm.compressor_input_bytes", Description: "Bytes submitted to compressor", Type: TypeUint64, Category: CatVMCompressor},
	{Name: "vm.compressor_compressed_bytes", Description: "Bytes after compression", Type: TypeUint64, Category: CatVMCompressor},
	{Name: "vm.compressor_bytes_used", Description: "Current compressor memory usage", Type: TypeUint64, Category: CatVMCompressor},
	{Name: "vm.compressor_pool_size", Description: "Compressor pool size", Type: TypeUint64, Category: CatVMCompressor},
	{Name: "vm.compressor_segment_pages_compressed", Description: "Compressed segment pages", Type: TypeInt32, Category: CatVMCompressor},
	{Name: "vm.compressor_segment_limit", Description: "Max compressed segments", Type: TypeInt32, Category: CatVMCompressor},
	{Name: "vm.compressor_swapouts_under_30s", Description: "Swapouts within 30s of compress", Type: TypeInt64, Category: CatVMCompressor},
	{Name: "vm.compressor_swapouts_under_60s", Description: "Swapouts within 60s", Type: TypeInt64, Category: CatVMCompressor},
	{Name: "vm.compressor_swapouts_under_300s", Description: "Swapouts within 300s", Type: TypeInt64, Category: CatVMCompressor},
	{Name: "vm.compressor_swapper_reclaim_swapins", Description: "Reclaim swap-ins", Type: TypeInt64, Category: CatVMCompressor},
	{Name: "vm.compressor_swapper_defrag_swapins", Description: "Defrag swap-ins", Type: TypeInt64, Category: CatVMCompressor},
	{Name: "vm.compressor_swapper_swapout_thrashing_detected", Description: "Swap thrashing events", Type: TypeInt64, Category: CatVMCompressor},
	// WK (WKdm) compression
	{Name: "vm.wk_compressions", Description: "WKdm compressions", Type: TypeInt64, Category: CatVMCompressor},
	{Name: "vm.wk_compressed_bytes_total", Description: "WKdm total compressed bytes", Type: TypeInt64, Category: CatVMCompressor},
	{Name: "vm.wk_decompressions", Description: "WKdm decompressions", Type: TypeInt64, Category: CatVMCompressor},
	{Name: "vm.wk_decompressed_bytes", Description: "WKdm decompressed bytes", Type: TypeInt64, Category: CatVMCompressor},
	// LZ4 compression
	{Name: "vm.lz4_compressions", Description: "LZ4 compressions", Type: TypeInt64, Category: CatVMCompressor},
	{Name: "vm.lz4_compressed_bytes", Description: "LZ4 compressed bytes", Type: TypeInt64, Category: CatVMCompressor},
	{Name: "vm.lz4_decompressions", Description: "LZ4 decompressions", Type: TypeInt64, Category: CatVMCompressor},
	{Name: "vm.lz4_compression_failures", Description: "LZ4 compression failures", Type: TypeInt64, Category: CatVMCompressor},

	// =========================================================================
	// vm.pageout — Pageout Activity
	// =========================================================================
	{Name: "vm.pageout_inactive_clean", Description: "Inactive clean pages freed", Type: TypeInt32, Category: CatVMPageout},
	{Name: "vm.pageout_inactive_used", Description: "Inactive used pages reactivated", Type: TypeInt32, Category: CatVMPageout},
	{Name: "vm.pageout_inactive_dirty_internal", Description: "Inactive dirty internal written", Type: TypeInt64, Category: CatVMPageout},
	{Name: "vm.pageout_inactive_dirty_external", Description: "Inactive dirty external written", Type: TypeInt64, Category: CatVMPageout},
	{Name: "vm.pageout_speculative_clean", Description: "Speculative pages freed", Type: TypeInt64, Category: CatVMPageout},
	{Name: "vm.pageout_freed_external", Description: "External pages freed", Type: TypeInt64, Category: CatVMPageout},
	{Name: "vm.pageout_freed_speculative", Description: "Speculative pages freed (alt)", Type: TypeInt64, Category: CatVMPageout},
	{Name: "vm.pageout_freed_cleaned", Description: "Cleaned pages freed", Type: TypeInt64, Category: CatVMPageout},
	{Name: "vm.pageout_protect_realtime", Description: "Realtime pages protected", Type: TypeInt32, Category: CatVMPageout},
	{Name: "vm.vm_pageout_considered_bq_internal", Description: "Internal pages considered", Type: TypeInt64, Category: CatVMPageout},
	{Name: "vm.vm_pageout_considered_bq_external", Description: "External pages considered", Type: TypeInt64, Category: CatVMPageout},
	{Name: "vm.vm_pageout_rejected_bq_internal", Description: "Internal pages rejected", Type: TypeInt64, Category: CatVMPageout},
	{Name: "vm.vm_pageout_rejected_bq_external", Description: "External pages rejected", Type: TypeInt64, Category: CatVMPageout},

	// =========================================================================
	// vm.swap — Swap
	// =========================================================================
	{Name: "vm.swapusage", Description: "Swap usage stats", Type: TypeSwap, Category: CatVMSwap},
	{Name: "vm.swap_enabled", Description: "Swap enabled", Type: TypeInt32, Category: CatVMSwap},
	{Name: "vm.swapfileprefix", Description: "Swap file path prefix", Type: TypeString, Category: CatVMSwap},
	{Name: "vm.loadavg", Description: "System load averages", Type: TypeLoadavg, Category: CatVMSwap},

	// =========================================================================
	// vm.wire — Wire Limits
	// =========================================================================
	{Name: "vm.global_user_wire_limit", Description: "Max user-wired memory", Type: TypeUint64, Category: CatVMWire},
	{Name: "vm.user_wire_limit", Description: "Per-task user-wire limit", Type: TypeUint64, Category: CatVMWire},
	{Name: "vm.global_no_user_wire_amount", Description: "Reserved non-wirable memory", Type: TypeUint64, Category: CatVMWire},
	{Name: "vm.add_wire_count_over_global_limit", Description: "Wire attempts over global limit", Type: TypeInt64, Category: CatVMWire},
	{Name: "vm.add_wire_count_over_user_limit", Description: "Wire attempts over user limit", Type: TypeInt64, Category: CatVMWire},

	// =========================================================================
	// vm.misc — Miscellaneous VM
	// =========================================================================
	{Name: "vm.cs_blob_count", Description: "Code signature blobs", Type: TypeInt32, Category: CatVMMisc},
	{Name: "vm.cs_blob_count_peak", Description: "Peak CS blob count", Type: TypeInt32, Category: CatVMMisc},
	{Name: "vm.cs_blob_size", Description: "CS blob memory usage", Type: TypeInt64, Category: CatVMMisc},
	{Name: "vm.cs_blob_size_peak", Description: "Peak CS blob memory", Type: TypeInt64, Category: CatVMMisc},
	{Name: "vm.shared_region_count", Description: "Active shared regions", Type: TypeInt32, Category: CatVMMisc},
	{Name: "vm.shared_region_peak", Description: "Peak shared regions", Type: TypeInt32, Category: CatVMMisc},
	{Name: "vm.copied_on_read", Description: "Copy-on-read events", Type: TypeInt64, Category: CatVMMisc},

	// =========================================================================
	// machdep — Machine-Dependent
	// =========================================================================
	{Name: "machdep.cpu.brand_string", Description: "CPU brand string", Type: TypeString, Category: CatMachdep},
	{Name: "machdep.cpu.core_count", Description: "Physical core count", Type: TypeInt32, Category: CatMachdep},
	{Name: "machdep.cpu.thread_count", Description: "Thread count", Type: TypeInt32, Category: CatMachdep},
	{Name: "machdep.cpu.cores_per_package", Description: "Cores per package", Type: TypeInt32, Category: CatMachdep},
	{Name: "machdep.cpu.logical_per_package", Description: "Logical CPUs per package", Type: TypeInt32, Category: CatMachdep},
	{Name: "machdep.ptrauth_enabled", Description: "Pointer authentication enabled", Type: TypeInt32, Category: CatMachdep},
	{Name: "machdep.virtual_address_size", Description: "Virtual address bits", Type: TypeInt32, Category: CatMachdep},
	{Name: "machdep.time_since_reset", Description: "Time since CPU reset (abs)", Type: TypeUint64, Category: CatMachdep},
	{Name: "machdep.wake_abstime", Description: "Absolute time of last wake", Type: TypeUint64, Category: CatMachdep},
	{Name: "machdep.wake_conttime", Description: "Continuous time of last wake", Type: TypeUint64, Category: CatMachdep},
	{Name: "machdep.user_idle_level", Description: "User idle level", Type: TypeInt32, Category: CatMachdep},

	// =========================================================================
	// net.tcp — TCP
	// =========================================================================
	{Name: "net.inet.tcp.mssdflt", Description: "Default TCP MSS", Type: TypeInt32, Category: CatNetTCP},
	{Name: "net.inet.tcp.keepidle", Description: "Keepalive idle time (ms)", Type: TypeInt32, Category: CatNetTCP},
	{Name: "net.inet.tcp.keepintvl", Description: "Keepalive interval (ms)", Type: TypeInt32, Category: CatNetTCP},
	{Name: "net.inet.tcp.keepcnt", Description: "Keepalive probe count", Type: TypeInt32, Category: CatNetTCP},
	{Name: "net.inet.tcp.sendspace", Description: "Default send buffer", Type: TypeInt32, Category: CatNetTCP},
	{Name: "net.inet.tcp.recvspace", Description: "Default receive buffer", Type: TypeInt32, Category: CatNetTCP},
	{Name: "net.inet.tcp.pcbcount", Description: "Active TCP connections", Type: TypeInt32, Category: CatNetTCP},
	{Name: "net.inet.tcp.sack", Description: "SACK enabled", Type: TypeInt32, Category: CatNetTCP},
	{Name: "net.inet.tcp.sack_globalholes", Description: "Current global SACK holes", Type: TypeInt32, Category: CatNetTCP},
	{Name: "net.inet.tcp.sack_globalmaxholes", Description: "Max global SACK holes", Type: TypeInt32, Category: CatNetTCP},
	{Name: "net.inet.tcp.ecn_initiate_out", Description: "ECN for outgoing", Type: TypeInt32, Category: CatNetTCP},
	{Name: "net.inet.tcp.ecn_negotiate_in", Description: "ECN for incoming", Type: TypeInt32, Category: CatNetTCP},
	{Name: "net.inet.tcp.fastopen", Description: "TCP Fast Open", Type: TypeInt32, Category: CatNetTCP},
	{Name: "net.inet.tcp.delayed_ack", Description: "Delayed ACK", Type: TypeInt32, Category: CatNetTCP},
	{Name: "net.inet.tcp.blackhole", Description: "Blackhole mode", Type: TypeInt32, Category: CatNetTCP},
	{Name: "net.inet.tcp.cubic_sockets", Description: "Sockets using CUBIC", Type: TypeInt32, Category: CatNetTCP},
	{Name: "net.inet.tcp.v6mssdflt", Description: "Default IPv6 MSS", Type: TypeInt32, Category: CatNetTCP},
	{Name: "net.inet.tcp.always_keepalive", Description: "Force keepalive", Type: TypeInt32, Category: CatNetTCP},

	// =========================================================================
	// net.udp — UDP
	// =========================================================================
	{Name: "net.inet.udp.pcbcount", Description: "Active UDP sockets", Type: TypeInt32, Category: CatNetUDP},
	{Name: "net.inet.udp.maxdgram", Description: "Max datagram size", Type: TypeInt32, Category: CatNetUDP},
	{Name: "net.inet.udp.recvspace", Description: "Default receive buffer", Type: TypeInt32, Category: CatNetUDP},
	{Name: "net.inet.udp.checksum", Description: "Checksum enabled", Type: TypeInt32, Category: CatNetUDP},

	// =========================================================================
	// net.ip — IP
	// =========================================================================
	{Name: "net.inet.ip.forwarding", Description: "IP forwarding", Type: TypeInt32, Category: CatNetIP},
	{Name: "net.inet.ip.ttl", Description: "Default TTL", Type: TypeInt32, Category: CatNetIP},
	{Name: "net.inet.ip.maxfragpackets", Description: "Max fragmented packets", Type: TypeInt32, Category: CatNetIP},

	// =========================================================================
	// net.misc — Misc Network
	// =========================================================================
	{Name: "net.local.pcbcount", Description: "Unix domain socket count", Type: TypeInt32, Category: CatNetMisc},

	// =========================================================================
	// vfs — Filesystem
	// =========================================================================
	{Name: "vfs.vnstats.num_vnodes", Description: "Total vnodes", Type: TypeInt64, Category: CatVFS},
	{Name: "vfs.vnstats.num_free_vnodes", Description: "Free vnodes", Type: TypeInt64, Category: CatVFS},
	{Name: "vfs.vnstats.num_recycledvnodes", Description: "Recycled vnodes", Type: TypeInt64, Category: CatVFS},
	{Name: "vfs.vnstats.num_newvnode_calls", Description: "New vnode allocations", Type: TypeInt64, Category: CatVFS},
	{Name: "vfs.nummntops", Description: "Mount operations", Type: TypeInt32, Category: CatVFS},
	{Name: "vfs.generic.apfs.allocated", Description: "APFS allocated bytes", Type: TypeInt64, Category: CatVFS},

	// =========================================================================
	// debug — Debug / I/O Throttling
	// =========================================================================
	{Name: "debug.lowpri_throttle_enabled", Description: "Low-priority I/O throttling", Type: TypeInt32, Category: CatDebug},
	{Name: "debug.bpf_bufsize", Description: "BPF buffer size", Type: TypeInt32, Category: CatDebug},
	{Name: "debug.bpf_maxbufsize", Description: "Max BPF buffer size", Type: TypeInt32, Category: CatDebug},

	// =========================================================================
	// kpc — Performance Counters
	// =========================================================================
	{Name: "kpc.pc_capture_supported", Description: "PC capture supported", Type: TypeInt32, Category: CatKPC},
	{Name: "kperf.debug_level", Description: "kperf debug level", Type: TypeInt32, Category: CatKPC},

	// =========================================================================
	// iogpu — GPU Memory
	// =========================================================================
	{Name: "iogpu.wired_limit_mb", Description: "GPU wired memory limit (MB)", Type: TypeInt32, Category: CatIOGPU},
	{Name: "iogpu.wired_lwm_mb", Description: "GPU wired low water mark (MB)", Type: TypeInt32, Category: CatIOGPU},
	{Name: "iogpu.dynamic_lwm", Description: "Dynamic low water mark", Type: TypeInt32, Category: CatIOGPU},
}

// knownByName is a lookup index built lazily.
var knownByName map[string]*Info

// ByName returns the Info for a known metric, or nil if unknown.
func ByName(name string) *Info {
	if knownByName == nil {
		knownByName = make(map[string]*Info, len(Known))
		for i := range Known {
			knownByName[Known[i].Name] = &Known[i]
		}
	}
	return knownByName[name]
}

// ByCategory returns all metrics in a category.
func ByCategory(cat Category) []Info {
	var out []Info
	for _, m := range Known {
		if m.Category == cat {
			out = append(out, m)
		}
	}
	return out
}

// Categories returns all distinct categories in registry order.
func Categories() []Category {
	seen := make(map[Category]bool)
	var out []Category
	for _, m := range Known {
		if !seen[m.Category] {
			seen[m.Category] = true
			out = append(out, m.Category)
		}
	}
	return out
}
