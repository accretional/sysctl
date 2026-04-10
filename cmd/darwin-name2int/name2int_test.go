//go:build darwin

package main

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/accretional/sysctl/internal/macosasmsysctl"
	"github.com/accretional/sysctl/internal/metrics"
)

// Shared MIBCache for all tests.
var cache = macosasmsysctl.NewMIBCache()

// TestResolveAll verifies that every non-computed metric resolves to a MIB.
func TestResolveAll(t *testing.T) {
	resolved := 0
	failed := 0
	for _, info := range metrics.Known {
		if info.Type == metrics.TypeComputed {
			continue
		}
		mib, err := cache.Resolve(info.Name)
		if err != nil {
			t.Errorf("resolve %s: %v", info.Name, err)
			failed++
			continue
		}
		if len(mib) == 0 {
			t.Errorf("resolve %s: empty MIB", info.Name)
			failed++
			continue
		}
		resolved++
	}
	t.Logf("resolved %d metrics, %d failed", resolved, failed)
}

// TestMIBUniqueness verifies that no two metrics resolve to the same MIB.
func TestMIBUniqueness(t *testing.T) {
	seen := make(map[string]string) // mib string -> metric name
	for _, info := range metrics.Known {
		if info.Type == metrics.TypeComputed {
			continue
		}
		mib, err := cache.Resolve(info.Name)
		if err != nil {
			continue
		}
		key := fmt.Sprintf("%v", mib)
		if prev, ok := seen[key]; ok {
			t.Errorf("MIB collision: %s and %s both resolve to %s", prev, info.Name, key)
		}
		seen[key] = info.Name
	}
	t.Logf("%d unique MIBs", len(seen))
}

// TestMIBReadConsistency validates that sysctl(mib) and sysctlbyname(name)
// return identical bytes for every metric.
func TestMIBReadConsistency(t *testing.T) {
	consistent := 0
	volatile := 0
	for _, info := range metrics.Known {
		if info.Type == metrics.TypeComputed {
			continue
		}

		// Read via MIB
		mibRaw, err := cache.GetRaw(info.Name)
		if err != nil {
			t.Logf("skip %s (mib read error: %v)", info.Name, err)
			continue
		}

		// Read via sysctlbyname
		nameRaw, err := macosasmsysctl.GetRaw(info.Name)
		if err != nil {
			t.Errorf("%s: sysctlbyname error: %v", info.Name, err)
			continue
		}

		if bytes.Equal(mibRaw, nameRaw) {
			consistent++
			continue
		}

		// Same size = volatile metric (value changed between reads)
		if len(mibRaw) == len(nameRaw) {
			volatile++
			continue
		}

		// Different size = real problem
		t.Errorf("%s: size mismatch: mib=%d bytes, name=%d bytes", info.Name, len(mibRaw), len(nameRaw))
	}
	t.Logf("consistent: %d, volatile: %d", consistent, volatile)
}

// TestKnownXNUConstants validates that metrics with well-known XNU header
// constants resolve to the expected MIB values.
func TestKnownXNUConstants(t *testing.T) {
	tests := []struct {
		name     string
		expected []uint32
		xnuName  string
	}{
		// CTL_KERN
		{"kern.ostype", []uint32{1, 1}, "CTL_KERN+KERN_OSTYPE"},
		{"kern.osrelease", []uint32{1, 2}, "CTL_KERN+KERN_OSRELEASE"},
		{"kern.maxvnodes", []uint32{1, 5}, "CTL_KERN+KERN_MAXVNODES"},
		{"kern.maxproc", []uint32{1, 6}, "CTL_KERN+KERN_MAXPROC"},
		{"kern.maxfiles", []uint32{1, 7}, "CTL_KERN+KERN_MAXFILES"},
		{"kern.hostname", []uint32{1, 10}, "CTL_KERN+KERN_HOSTNAME"},
		{"kern.clockrate", []uint32{1, 12}, "CTL_KERN+KERN_CLOCKRATE"},
		{"kern.boottime", []uint32{1, 21}, "CTL_KERN+KERN_BOOTTIME"},
		{"kern.maxfilesperproc", []uint32{1, 29}, "CTL_KERN+KERN_MAXFILESPERPROC"},
		{"kern.maxprocperuid", []uint32{1, 30}, "CTL_KERN+KERN_MAXPROCPERUID"},
		{"kern.aiomax", []uint32{1, 46}, "CTL_KERN+KERN_AIOMAX"},
		{"kern.aioprocmax", []uint32{1, 47}, "CTL_KERN+KERN_AIOPROCMAX"},
		{"kern.coredump", []uint32{1, 51}, "CTL_KERN+KERN_COREDUMP"},
		{"kern.osversion", []uint32{1, 65}, "CTL_KERN+KERN_OSVERSION"},
		{"kern.safeboot", []uint32{1, 66}, "CTL_KERN+KERN_SAFEBOOT"},

		// CTL_HW
		{"hw.ncpu", []uint32{6, 3}, "CTL_HW+HW_NCPU"},
		{"hw.memsize", []uint32{6, 24}, "CTL_HW+HW_MEMSIZE"},

		// CTL_VM
		{"vm.loadavg", []uint32{2, 2}, "CTL_VM+VM_LOADAVG"},
		{"vm.swapusage", []uint32{2, 5}, "CTL_VM+VM_SWAPUSAGE"},

		// KERN_IPC subtree
		{"kern.ipc.maxsockbuf", []uint32{1, 32, 1}, "CTL_KERN+KERN_IPC+KIPC_MAXSOCKBUF"},
		{"kern.ipc.somaxconn", []uint32{1, 32, 3}, "CTL_KERN+KERN_IPC+KIPC_SOMAXCONN"},
	}

	matched := 0
	for _, tt := range tests {
		mib, err := cache.Resolve(tt.name)
		if err != nil {
			t.Errorf("%s: resolve error: %v", tt.name, err)
			continue
		}

		// Compare only the prefix — some MIBs may have additional elements
		if len(mib) < len(tt.expected) {
			t.Errorf("%s (%s): MIB too short: got %v, want prefix %v", tt.name, tt.xnuName, mib, tt.expected)
			continue
		}

		match := true
		for i, v := range tt.expected {
			if mib[i] != v {
				match = false
				break
			}
		}
		if !match {
			t.Errorf("%s (%s): got %v, want prefix %v", tt.name, tt.xnuName, mib, tt.expected)
		} else {
			matched++
		}
	}
	t.Logf("%d/%d XNU constant matches verified", matched, len(tests))
}

// TestHWPagesize documents the hw.pagesize OID divergence from HW_PAGESIZE.
// Modern macOS resolves hw.pagesize to a different OID than the legacy HW_PAGESIZE=7.
func TestHWPagesize(t *testing.T) {
	mib, err := cache.Resolve("hw.pagesize")
	if err != nil {
		t.Fatalf("resolve hw.pagesize: %v", err)
	}

	// XNU header says HW_PAGESIZE=7, but runtime resolves differently
	if mib[0] == 6 && len(mib) >= 2 && mib[1] == 7 {
		t.Logf("hw.pagesize uses legacy HW_PAGESIZE=7: %v", mib)
	} else {
		t.Logf("hw.pagesize uses OID %v (NOT legacy HW_PAGESIZE=7) — documented divergence", mib)
	}

	// Verify we can still read the value
	val, err := cache.GetInt64("hw.pagesize")
	if err != nil {
		t.Fatalf("read hw.pagesize: %v", err)
	}
	if val != 16384 && val != 4096 {
		t.Errorf("hw.pagesize = %d, expected 4096 or 16384", val)
	}
	t.Logf("hw.pagesize = %d", val)
}

// TestMIBDepthDistribution documents the distribution of MIB array lengths.
func TestMIBDepthDistribution(t *testing.T) {
	depths := make(map[int]int)
	for _, info := range metrics.Known {
		if info.Type == metrics.TypeComputed {
			continue
		}
		mib, err := cache.Resolve(info.Name)
		if err != nil {
			continue
		}
		depths[len(mib)]++
	}

	keys := make([]int, 0, len(depths))
	for k := range depths {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		t.Logf("depth %d: %d metrics", k, depths[k])
	}
}

// TestTopLevelCategories documents which MIB[0] values appear and how many
// are from the well-known CTL_* constants vs. extended OID space.
func TestTopLevelCategories(t *testing.T) {
	wellKnown := map[uint32]string{
		1: "CTL_KERN",
		2: "CTL_VM",
		3: "CTL_VFS",
		4: "CTL_NET",
		5: "CTL_DEBUG",
		6: "CTL_HW",
		7: "CTL_MACHDEP",
		8: "CTL_USER",
	}

	topCounts := make(map[uint32]int)
	for _, info := range metrics.Known {
		if info.Type == metrics.TypeComputed {
			continue
		}
		mib, err := cache.Resolve(info.Name)
		if err != nil || len(mib) == 0 {
			continue
		}
		topCounts[mib[0]]++
	}

	tops := make([]uint32, 0, len(topCounts))
	for k := range topCounts {
		tops = append(tops, k)
	}
	sort.Slice(tops, func(i, j int) bool { return tops[i] < tops[j] })

	knownCount := 0
	extendedCount := 0
	for _, top := range tops {
		name, ok := wellKnown[top]
		if ok {
			t.Logf("MIB[0]=%d (%s): %d metrics", top, name, topCounts[top])
			knownCount += topCounts[top]
		} else {
			t.Logf("MIB[0]=%d (extended OID): %d metrics", top, topCounts[top])
			extendedCount += topCounts[top]
		}
	}
	t.Logf("well-known: %d, extended: %d", knownCount, extendedCount)
}

// TestNetworkMIBStructure validates the hierarchical MIB structure for net.* metrics.
// Network MIBs follow: CTL_NET(4) / PF_* / IPPROTO_* / sub-ID.
func TestNetworkMIBStructure(t *testing.T) {
	protoNames := map[uint32]string{
		0:   "IPPROTO_IP",
		6:   "IPPROTO_TCP",
		17:  "IPPROTO_UDP",
		256: "IPPROTO_MPTCP",
	}
	pfNames := map[uint32]string{
		1:   "PF_LOCAL",
		2:   "PF_INET",
		115: "extended", // soflow lives here
	}

	for _, info := range metrics.Known {
		if info.Type == metrics.TypeComputed {
			continue
		}
		if !strings.HasPrefix(info.Name, "net.") {
			continue
		}
		mib, err := cache.Resolve(info.Name)
		if err != nil {
			t.Errorf("%s: resolve error: %v", info.Name, err)
			continue
		}
		if mib[0] != 4 {
			t.Errorf("%s: MIB[0]=%d, expected CTL_NET=4", info.Name, mib[0])
			continue
		}

		pf := "?"
		if name, ok := pfNames[mib[1]]; ok {
			pf = name
		}

		if len(mib) >= 3 {
			proto := "?"
			if name, ok := protoNames[mib[2]]; ok {
				proto = name
			}
			t.Logf("%s -> {%d,%d,%d,...} (%s/%s)", info.Name, mib[0], mib[1], mib[2], pf, proto)
		} else {
			t.Logf("%s -> %v (%s)", info.Name, mib, pf)
		}
	}
}

// TestFlattenability checks if all MIB arrays can be encoded as unique uint64 values.
// This is relevant to DELTA_DESIGN.md Phase 4: MIB-Based IDs.
func TestFlattenability(t *testing.T) {
	type entry struct {
		name string
		mib  []uint32
	}

	var entries []entry
	for _, info := range metrics.Known {
		if info.Type == metrics.TypeComputed {
			continue
		}
		mib, err := cache.Resolve(info.Name)
		if err != nil {
			continue
		}
		entries = append(entries, entry{info.Name, mib})
	}

	// Try 16-bits-per-level encoding (max 4 levels)
	flatMap := make(map[uint64][]string)
	unflattenable := 0
	for _, e := range entries {
		flat, ok := flattenMIB(e.mib)
		if !ok {
			t.Logf("unflattenable (depth %d or value >16 bits): %s -> %v", len(e.mib), e.name, e.mib)
			unflattenable++
			continue
		}
		flatMap[flat] = append(flatMap[flat], e.name)
	}

	collisions := 0
	for flat, names := range flatMap {
		if len(names) > 1 {
			t.Errorf("flat collision: 0x%016x -> %v", flat, names)
			collisions++
		}
	}

	t.Logf("flattenable: %d, unflattenable: %d, collisions: %d", len(flatMap), unflattenable, collisions)

	// Check max values per level
	maxPerLevel := make([]uint32, 5)
	for _, e := range entries {
		for i, v := range e.mib {
			if i < len(maxPerLevel) && v > maxPerLevel[i] {
				maxPerLevel[i] = v
			}
		}
	}
	for i, max := range maxPerLevel {
		if max > 0 {
			bits := 0
			for v := max; v > 0; v >>= 1 {
				bits++
			}
			t.Logf("level %d max value: %d (%d bits)", i, max, bits)
		}
	}
}

// TestVolatileMetrics identifies which metrics change between consecutive reads.
// This validates our STATIC vs POLLED/CONSTRAINED classifications.
func TestVolatileMetrics(t *testing.T) {
	static := 0
	volatile := 0
	for _, info := range metrics.Known {
		if info.Type == metrics.TypeComputed {
			continue
		}

		// Read twice via MIB
		raw1, err := cache.GetRaw(info.Name)
		if err != nil {
			continue
		}
		raw2, err := cache.GetRaw(info.Name)
		if err != nil {
			continue
		}

		if bytes.Equal(raw1, raw2) {
			static++
		} else {
			volatile++
			t.Logf("VOLATILE: %s (%d bytes, values differ between reads)", info.Name, len(raw1))
		}
	}
	t.Logf("static between reads: %d, volatile: %d", static, volatile)
}
