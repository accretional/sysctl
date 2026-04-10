//go:build darwin

package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"testing"

	"github.com/accretional/sysctl/internal/macosasmsysctl"
	"github.com/accretional/sysctl/internal/metrics"
)

type inventoryEntry struct {
	name       string
	mib        []uint32
	mibStr     string
	depth      int
	valueSize  int
	valueType  string // "int32", "int64", "string", "opaque", etc.
	topLevel   string // "kern", "hw", etc.
	registered bool
	resolveErr string
}

// TestFullSysctlInventory enumerates every sysctl on the system (via `sysctl -a`),
// resolves each to its MIB, and produces a full inventory CSV + summary.
func TestFullSysctlInventory(t *testing.T) {
	// Gather our registered metric names for cross-reference.
	registered := make(map[string]bool, len(metrics.Known))
	for _, info := range metrics.Known {
		registered[info.Name] = true
	}

	// Get all sysctl names from the system.
	allNames := getAllSysctlNames(t)
	t.Logf("system has %d sysctl entries, we register %d (including %d computed)",
		len(allNames), len(metrics.Known), countComputed())

	cache := macosasmsysctl.NewMIBCache()

	var entries []inventoryEntry
	resolved := 0
	failed := 0

	for _, name := range allNames {
		e := inventoryEntry{name: name, registered: registered[name]}

		mib, err := cache.Resolve(name)
		if err != nil {
			e.resolveErr = err.Error()
			failed++
			entries = append(entries, e)
			continue
		}
		resolved++

		e.mib = mib
		e.mibStr = formatMIB(mib)
		e.depth = len(mib)

		// Determine top-level
		switch mib[0] {
		case 1:
			e.topLevel = "kern"
		case 2:
			e.topLevel = "vm"
		case 3:
			e.topLevel = "vfs"
		case 4:
			e.topLevel = "net"
		case 5:
			e.topLevel = "debug"
		case 6:
			e.topLevel = "hw"
		case 7:
			e.topLevel = "machdep"
		case 8:
			e.topLevel = "user"
		default:
			e.topLevel = fmt.Sprintf("oid_%d", mib[0])
		}

		// Read value to get size
		raw, err := macosasmsysctl.GetRaw(name)
		if err != nil {
			e.valueType = "unreadable"
			e.valueSize = -1
		} else {
			e.valueSize = len(raw)
			e.valueType = classifyValueType(raw)
		}

		entries = append(entries, e)
	}

	t.Logf("resolved: %d, failed: %d", resolved, failed)

	// Write full inventory CSV
	writeInventoryCSV(t, entries)

	// Summary by top-level namespace
	t.Log("")
	t.Log("=== Coverage by top-level namespace ===")
	type nsSummary struct {
		total, ours, missing int
	}
	nsMap := make(map[string]*nsSummary)
	for _, e := range entries {
		ns := e.topLevel
		if e.resolveErr != "" {
			ns = strings.SplitN(e.name, ".", 2)[0]
		}
		if nsMap[ns] == nil {
			nsMap[ns] = &nsSummary{}
		}
		nsMap[ns].total++
		if e.registered {
			nsMap[ns].ours++
		} else {
			nsMap[ns].missing++
		}
	}
	nsKeys := make([]string, 0, len(nsMap))
	for k := range nsMap {
		nsKeys = append(nsKeys, k)
	}
	sort.Strings(nsKeys)
	for _, ns := range nsKeys {
		s := nsMap[ns]
		pct := float64(s.ours) / float64(s.total) * 100
		t.Logf("  %-12s  total=%4d  ours=%4d  missing=%4d  coverage=%.0f%%", ns, s.total, s.ours, s.missing, pct)
	}

	// Summary by second-level namespace (for the big ones)
	t.Log("")
	t.Log("=== Largest missing namespaces (second-level) ===")
	ns2Map := make(map[string]int)
	for _, e := range entries {
		if e.registered {
			continue
		}
		parts := strings.SplitN(e.name, ".", 3)
		if len(parts) >= 2 {
			ns2Map[parts[0]+"."+parts[1]]++
		}
	}
	type ns2Entry struct {
		name  string
		count int
	}
	var ns2List []ns2Entry
	for k, v := range ns2Map {
		ns2List = append(ns2List, ns2Entry{k, v})
	}
	sort.Slice(ns2List, func(i, j int) bool { return ns2List[i].count > ns2List[j].count })
	for i, e := range ns2List {
		if i >= 30 {
			break
		}
		t.Logf("  %-30s  %4d missing", e.name, e.count)
	}

	// Summary by value size
	t.Log("")
	t.Log("=== Value size distribution (all sysctls) ===")
	sizeMap := make(map[int]int)
	for _, e := range entries {
		if e.valueSize >= 0 {
			sizeMap[e.valueSize]++
		}
	}
	sizes := make([]int, 0, len(sizeMap))
	for s := range sizeMap {
		sizes = append(sizes, s)
	}
	sort.Ints(sizes)
	for _, s := range sizes {
		marker := ""
		if s <= 8 {
			marker = "  <-- fits in fixed64"
		}
		t.Logf("  %4d bytes: %4d metrics%s", s, sizeMap[s], marker)
	}

	// MIB depth distribution for ALL sysctls
	t.Log("")
	t.Log("=== MIB depth distribution (all sysctls) ===")
	depthMap := make(map[int]int)
	for _, e := range entries {
		if e.depth > 0 {
			depthMap[e.depth]++
		}
	}
	depths := make([]int, 0, len(depthMap))
	for d := range depthMap {
		depths = append(depths, d)
	}
	sort.Ints(depths)
	for _, d := range depths {
		t.Logf("  depth %d: %4d metrics", d, depthMap[d])
	}

	// MIB[0] distribution for ALL sysctls
	t.Log("")
	t.Log("=== Top-level MIB[0] values (all sysctls) ===")
	topMap := make(map[uint32]int)
	for _, e := range entries {
		if len(e.mib) > 0 {
			topMap[e.mib[0]]++
		}
	}
	topIDs := make([]uint32, 0, len(topMap))
	for k := range topMap {
		topIDs = append(topIDs, k)
	}
	sort.Slice(topIDs, func(i, j int) bool { return topIDs[i] < topIDs[j] })
	for _, id := range topIDs {
		label := fmt.Sprintf("unknown(%d)", id)
		switch id {
		case 1:
			label = "CTL_KERN"
		case 2:
			label = "CTL_VM"
		case 3:
			label = "CTL_VFS"
		case 4:
			label = "CTL_NET"
		case 5:
			label = "CTL_DEBUG"
		case 6:
			label = "CTL_HW"
		case 7:
			label = "CTL_MACHDEP"
		case 8:
			label = "CTL_USER"
		case 100:
			label = "ktrace"
		case 101:
			label = "kperf"
		case 102:
			label = "kpc"
		case 103:
			label = "security"
		case 104:
			label = "iogpu"
		}
		t.Logf("  MIB[0]=%3d  %-14s  %4d metrics", id, label, topMap[id])
	}

	// Check for MIB uniqueness across ALL sysctls
	t.Log("")
	t.Log("=== MIB uniqueness (all sysctls) ===")
	mibSeen := make(map[string][]string)
	for _, e := range entries {
		if e.mibStr != "" {
			mibSeen[e.mibStr] = append(mibSeen[e.mibStr], e.name)
		}
	}
	collisions := 0
	for mibStr, names := range mibSeen {
		if len(names) > 1 {
			t.Logf("  COLLISION: MIB %s -> %v", mibStr, names)
			collisions++
		}
	}
	if collisions == 0 {
		t.Logf("  no MIB collisions across all %d resolved sysctls", resolved)
	} else {
		t.Logf("  %d MIB collisions found", collisions)
	}

	// Max MIB values per level (all sysctls)
	t.Log("")
	t.Log("=== Max MIB value per level (all sysctls) ===")
	maxPerLevel := make([]uint32, 8)
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
			t.Logf("  level %d: max value %d (%d bits)", i, max, bits)
		}
	}
}

func getAllSysctlNames(t *testing.T) []string {
	t.Helper()
	out, err := exec.Command("sysctl", "-a").Output()
	if err != nil {
		t.Fatalf("sysctl -a: %v", err)
	}
	lines := strings.Split(string(out), "\n")
	var names []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		name := strings.TrimSpace(line[:idx])
		if name != "" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func classifyValueType(raw []byte) string {
	switch len(raw) {
	case 0:
		return "empty"
	case 4:
		return "int32"
	case 8:
		return "int64"
	case 16:
		return "struct16" // timeval
	case 20:
		return "struct20" // clockinfo
	case 24:
		return "struct24" // loadavg
	case 32:
		return "struct32" // xsw_usage
	default:
		// Check if it's a printable string
		if len(raw) > 0 && raw[len(raw)-1] == 0 {
			printable := true
			for _, b := range raw[:len(raw)-1] {
				if b < 0x20 || b > 0x7e {
					printable = false
					break
				}
			}
			if printable {
				return fmt.Sprintf("string(%d)", len(raw))
			}
		}
		return fmt.Sprintf("opaque(%d)", len(raw))
	}
}

func countComputed() int {
	n := 0
	for _, info := range metrics.Known {
		if info.Type == metrics.TypeComputed {
			n++
		}
	}
	return n
}

func writeInventoryCSV(t *testing.T, entries []inventoryEntry) {
	t.Helper()
	f, err := os.Create("full_inventory.csv")
	if err != nil {
		t.Fatalf("create CSV: %v", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	w.Write([]string{"name", "mib", "depth", "top_level", "value_size", "value_type", "registered", "resolve_error"})

	for _, e := range entries {
		reg := "no"
		if e.registered {
			reg = "yes"
		}
		w.Write([]string{
			e.name,
			e.mibStr,
			fmt.Sprintf("%d", e.depth),
			e.topLevel,
			fmt.Sprintf("%d", e.valueSize),
			e.valueType,
			reg,
			e.resolveErr,
		})
	}
	t.Logf("wrote full_inventory.csv (%d entries)", len(entries))
}
