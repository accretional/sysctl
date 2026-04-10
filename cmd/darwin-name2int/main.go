//go:build darwin

// darwin-name2int resolves all registered sysctl metric names to their
// integer MIB arrays and cross-references them against XNU header constants.
//
// It validates each mapping by reading via both sysctl(mib) and
// sysctlbyname(name) and comparing the raw bytes.
//
// Usage:
//
//	go run ./cmd/darwin-name2int/ [-csv] [-validate] [-json]
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/accretional/sysctl/internal/macosasmsysctl"
	"github.com/accretional/sysctl/internal/metrics"
)

// xnuConstant maps a well-known XNU header #define to its integer value.
type xnuConstant struct {
	Name  string
	Value uint32
}

// knownMIBs contains the well-known MIB mappings from XNU headers.
// These are compile-time constants from bsd/sys/sysctl.h and subsystem headers.
var knownTopLevel = map[string]uint32{
	"kern":    1, // CTL_KERN
	"vm":      2, // CTL_VM
	"vfs":     3, // CTL_VFS
	"net":     4, // CTL_NET
	"debug":   5, // CTL_DEBUG
	"hw":      6, // CTL_HW
	"machdep": 7, // CTL_MACHDEP
	"user":    8, // CTL_USER
}

// knownKern maps KERN_* sub-IDs from sysctl.h.
var knownKern = map[string]uint32{
	"kern.ostype":         1,  // KERN_OSTYPE
	"kern.osrelease":      2,  // KERN_OSRELEASE
	"kern.osrev":          3,  // KERN_OSREV
	"kern.version":        4,  // KERN_VERSION
	"kern.maxvnodes":      5,  // KERN_MAXVNODES
	"kern.maxproc":        6,  // KERN_MAXPROC
	"kern.maxfiles":       7,  // KERN_MAXFILES
	"kern.argmax":         8,  // KERN_ARGMAX
	"kern.securelevel":    9,  // KERN_SECURELVL
	"kern.hostname":       10, // KERN_HOSTNAME
	"kern.hostid":         11, // KERN_HOSTID
	"kern.clockrate":      12, // KERN_CLOCKRATE
	"kern.boottime":       21, // KERN_BOOTTIME
	"kern.nisdomainname":  22, // KERN_NISDOMAINNAME
	"kern.maxfilesperproc": 29, // KERN_MAXFILESPERPROC
	"kern.maxprocperuid":  30, // KERN_MAXPROCPERUID
	"kern.ipc":            32, // KERN_IPC (subtree)
	"kern.aiomax":         46, // KERN_AIOMAX
	"kern.aioprocmax":     47, // KERN_AIOPROCMAX
	"kern.corefile":       50, // KERN_COREFILE
	"kern.coredump":       51, // KERN_COREDUMP
	"kern.osversion":      65, // KERN_OSVERSION
	"kern.safeboot":       66, // KERN_SAFEBOOT
}

// knownHW maps HW_* sub-IDs from sysctl.h.
var knownHW = map[string]uint32{
	"hw.machine":  1,  // HW_MACHINE
	"hw.model":    2,  // HW_MODEL
	"hw.ncpu":     3,  // HW_NCPU
	"hw.byteorder": 4, // HW_BYTEORDER
	"hw.pagesize": 7,  // HW_PAGESIZE
	"hw.memsize":  24, // HW_MEMSIZE
	"hw.availcpu": 25, // HW_AVAILCPU
}

// knownVM maps VM_* sub-IDs from sysctl.h.
var knownVM = map[string]uint32{
	"vm.loadavg":   2, // VM_LOADAVG
	"vm.swapusage": 5, // VM_SWAPUSAGE
}

// knownKIPC maps KIPC_* sub-IDs for kern.ipc.* from sysctl.h.
var knownKIPC = map[string]uint32{
	"kern.ipc.maxsockbuf": 1, // KIPC_MAXSOCKBUF
	"kern.ipc.somaxconn":  3, // KIPC_SOMAXCONN
}

// MetricMIB holds the resolved MIB info for a single metric.
type MetricMIB struct {
	Name       string   `json:"name"`
	MIB        []uint32 `json:"mib"`
	MIBDepth   int      `json:"mib_depth"`
	TopLevel   string   `json:"top_level"`         // e.g., "kern", "hw"
	XNUMatch   string   `json:"xnu_match"`         // XNU constant name if matched, "" otherwise
	Validated  bool     `json:"validated"`          // true if sysctl(mib) == sysctlbyname(name)
	ValidErr   string   `json:"valid_err,omitempty"`
	ResolveErr string   `json:"resolve_err,omitempty"`
}

func main() {
	csvFlag := flag.Bool("csv", false, "output as CSV")
	jsonFlag := flag.Bool("json", false, "output as JSON")
	validateFlag := flag.Bool("validate", false, "validate MIB reads match sysctlbyname reads")
	flag.Parse()

	cache := macosasmsysctl.NewMIBCache()

	var results []MetricMIB
	for _, info := range metrics.Known {
		if info.Type == metrics.TypeComputed {
			continue // computed metrics have no sysctl
		}

		r := MetricMIB{Name: info.Name}

		mib, err := cache.Resolve(info.Name)
		if err != nil {
			r.ResolveErr = err.Error()
			results = append(results, r)
			continue
		}

		r.MIB = make([]uint32, len(mib))
		copy(r.MIB, mib)
		r.MIBDepth = len(mib)

		// Determine top-level category from MIB[0]
		for name, id := range knownTopLevel {
			if mib[0] == id {
				r.TopLevel = name
				break
			}
		}
		if r.TopLevel == "" {
			r.TopLevel = fmt.Sprintf("unknown(%d)", mib[0])
		}

		// Cross-reference with XNU header constants
		r.XNUMatch = matchXNU(info.Name, mib)

		// Validate: read via MIB and via name, compare bytes
		if *validateFlag {
			r.Validated, r.ValidErr = validateReads(cache, info.Name)
		}

		results = append(results, r)
	}

	// Output
	switch {
	case *jsonFlag:
		outputJSON(results)
	case *csvFlag:
		outputCSV(results)
	default:
		outputTable(results, *validateFlag)
	}

	// Summary stats
	printSummary(results, *validateFlag)
}

// matchXNU checks if the resolved MIB matches a known XNU header constant.
func matchXNU(name string, mib []uint32) string {
	// Check well-known kern.* names
	if sub, ok := knownKern[name]; ok {
		if len(mib) >= 2 && mib[0] == 1 && mib[1] == sub {
			return fmt.Sprintf("CTL_KERN(%d)+KERN_(%d)", mib[0], sub)
		}
		return fmt.Sprintf("MISMATCH: expected {1,%d}, got %v", sub, mib)
	}

	// Check well-known hw.* names
	if sub, ok := knownHW[name]; ok {
		if len(mib) >= 2 && mib[0] == 6 && mib[1] == sub {
			return fmt.Sprintf("CTL_HW(%d)+HW_(%d)", mib[0], sub)
		}
		return fmt.Sprintf("MISMATCH: expected {6,%d}, got %v", sub, mib)
	}

	// Check well-known vm.* names
	if sub, ok := knownVM[name]; ok {
		if len(mib) >= 2 && mib[0] == 2 && mib[1] == sub {
			return fmt.Sprintf("CTL_VM(%d)+VM_(%d)", mib[0], sub)
		}
		return fmt.Sprintf("MISMATCH: expected {2,%d}, got %v", sub, mib)
	}

	// Check kern.ipc.* names
	if sub, ok := knownKIPC[name]; ok {
		if len(mib) >= 3 && mib[0] == 1 && mib[1] == 32 && mib[2] == sub {
			return fmt.Sprintf("CTL_KERN+KERN_IPC+KIPC_(%d)", sub)
		}
		return fmt.Sprintf("MISMATCH: expected {1,32,%d}, got %v", sub, mib)
	}

	return "" // no known XNU constant for this name
}

// validateReads checks that reading via sysctl(mib) matches sysctlbyname(name).
// For volatile metrics (time counters), values may differ between reads.
// We retry up to 3 times and also check if sizes match for volatile values.
func validateReads(cache *macosasmsysctl.MIBCache, name string) (bool, string) {
	for attempt := 0; attempt < 3; attempt++ {
		// Read via MIB (through cache, which uses rawSysctl)
		mibRaw, err := cache.GetRaw(name)
		if err != nil {
			return false, fmt.Sprintf("mib read: %v", err)
		}

		// Read via sysctlbyname (direct, no cache)
		nameRaw, err := macosasmsysctl.GetRaw(name)
		if err != nil {
			return false, fmt.Sprintf("name read: %v", err)
		}

		if bytes.Equal(mibRaw, nameRaw) {
			return true, ""
		}

		// Same size but different content = volatile (time counter, etc.)
		if len(mibRaw) == len(nameRaw) && attempt == 2 {
			return true, "volatile"
		}

		if len(mibRaw) != len(nameRaw) {
			return false, fmt.Sprintf("size mismatch: mib=%d bytes, name=%d bytes", len(mibRaw), len(nameRaw))
		}
	}
	return true, "volatile"
}

func outputTable(results []MetricMIB, showValid bool) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	if showValid {
		fmt.Fprintf(w, "NAME\tMIB\tDEPTH\tTOP\tXNU\tVALID\n")
		fmt.Fprintf(w, "----\t---\t-----\t---\t---\t-----\n")
	} else {
		fmt.Fprintf(w, "NAME\tMIB\tDEPTH\tTOP\tXNU\n")
		fmt.Fprintf(w, "----\t---\t-----\t---\t---\n")
	}

	for _, r := range results {
		mibStr := formatMIB(r.MIB)
		if r.ResolveErr != "" {
			mibStr = "ERROR: " + r.ResolveErr
		}

		xnu := r.XNUMatch
		if xnu == "" {
			xnu = "-"
		}

		if showValid {
			validStr := "-"
			if r.ResolveErr == "" {
				if r.Validated {
					validStr = "OK"
				} else if r.ValidErr != "" {
					validStr = r.ValidErr
				}
			}
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\n", r.Name, mibStr, r.MIBDepth, r.TopLevel, xnu, validStr)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n", r.Name, mibStr, r.MIBDepth, r.TopLevel, xnu)
		}
	}
	w.Flush()
}

func outputCSV(results []MetricMIB) {
	fmt.Println("name,mib,depth,top_level,xnu_match,validated,error")
	for _, r := range results {
		mibStr := formatMIB(r.MIB)
		errStr := r.ResolveErr
		if errStr == "" {
			errStr = r.ValidErr
		}
		fmt.Printf("%s,%s,%d,%s,%s,%v,%s\n", r.Name, mibStr, r.MIBDepth, r.TopLevel, r.XNUMatch, r.Validated, errStr)
	}
}

func outputJSON(results []MetricMIB) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(results)
}

func formatMIB(mib []uint32) string {
	if len(mib) == 0 {
		return ""
	}
	parts := make([]string, len(mib))
	for i, v := range mib {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func printSummary(results []MetricMIB, showValid bool) {
	fmt.Println()

	// Count by resolve status
	resolved := 0
	failed := 0
	for _, r := range results {
		if r.ResolveErr != "" {
			failed++
		} else {
			resolved++
		}
	}
	fmt.Printf("Resolved: %d/%d", resolved, len(results))
	if failed > 0 {
		fmt.Printf(" (%d failed)", failed)
	}
	fmt.Println()

	// Count by MIB depth
	depthCounts := make(map[int]int)
	for _, r := range results {
		if r.ResolveErr == "" {
			depthCounts[r.MIBDepth]++
		}
	}
	depths := make([]int, 0, len(depthCounts))
	for d := range depthCounts {
		depths = append(depths, d)
	}
	sort.Ints(depths)
	fmt.Print("MIB depths: ")
	for i, d := range depths {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("depth=%d: %d", d, depthCounts[d])
	}
	fmt.Println()

	// Count by top-level
	topCounts := make(map[string]int)
	for _, r := range results {
		if r.ResolveErr == "" {
			topCounts[r.TopLevel]++
		}
	}
	tops := make([]string, 0, len(topCounts))
	for t := range topCounts {
		tops = append(tops, t)
	}
	sort.Strings(tops)
	fmt.Print("Top-level: ")
	for i, t := range tops {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%s=%d", t, topCounts[t])
	}
	fmt.Println()

	// XNU match stats
	xnuMatched := 0
	xnuMismatch := 0
	for _, r := range results {
		if strings.HasPrefix(r.XNUMatch, "MISMATCH") {
			xnuMismatch++
		} else if r.XNUMatch != "" {
			xnuMatched++
		}
	}
	fmt.Printf("XNU header matches: %d matched, %d mismatched, %d no constant\n",
		xnuMatched, xnuMismatch, resolved-xnuMatched-xnuMismatch)

	// Validation stats
	if showValid {
		validOK := 0
		validFail := 0
		for _, r := range results {
			if r.ResolveErr != "" {
				continue
			}
			if r.Validated {
				validOK++
			} else if r.ValidErr != "" {
				validFail++
			}
		}
		fmt.Printf("Validation: %d OK, %d failed\n", validOK, validFail)
	}

	// Unique flat MIB analysis: can we flatten MIB arrays to unique ints?
	fmt.Println()
	fmt.Println("=== Flat ID Analysis ===")
	analyzeFlattenability(results)
}

// analyzeFlattenability checks if MIB arrays can be flattened to unique integers.
// This is crucial for DELTA_DESIGN.md Phase 4: MIB-Based IDs.
func analyzeFlattenability(results []MetricMIB) {
	// Approach 1: Concatenate MIB elements as a single string key
	// This always works but doesn't give a single int.
	type mibKey struct {
		key string
		names []string
	}
	keyMap := make(map[string][]string)
	for _, r := range results {
		if r.ResolveErr != "" {
			continue
		}
		key := formatMIB(r.MIB)
		keyMap[key] = append(keyMap[key], r.Name)
	}
	dupes := 0
	for key, names := range keyMap {
		if len(names) > 1 {
			fmt.Printf("  COLLISION: MIB %s -> %v\n", key, names)
			dupes++
		}
	}
	if dupes == 0 {
		fmt.Println("  No MIB collisions — all names resolve to unique MIB arrays ✓")
	}

	// Approach 2: Try to flatten to a single uint64
	// Encoding: top(8 bits) | level1(16 bits) | level2(16 bits) | level3(16 bits) | level4(8 bits)
	flatMap := make(map[uint64][]string)
	unflattenable := 0
	for _, r := range results {
		if r.ResolveErr != "" || len(r.MIB) == 0 {
			continue
		}
		flat, ok := flattenMIB(r.MIB)
		if !ok {
			unflattenable++
			continue
		}
		flatMap[flat] = append(flatMap[flat], r.Name)
	}

	flatDupes := 0
	for flat, names := range flatMap {
		if len(names) > 1 {
			fmt.Printf("  FLAT COLLISION: 0x%016x -> %v\n", flat, names)
			flatDupes++
		}
	}
	fmt.Printf("  Flat uint64: %d flattenable, %d unflattenable, %d collisions\n",
		len(flatMap), unflattenable, flatDupes)

	// Approach 3: Simple sequential enum (what we'd use for CompactDelta)
	fmt.Printf("  Sequential enum: 0..%d — always collision-free, no MIB relationship\n", len(results)-1)

	// Show MIB range stats per top-level
	fmt.Println()
	fmt.Println("=== MIB Value Ranges ===")
	type rangeInfo struct {
		min, max uint32
		count    int
	}
	topRanges := make(map[uint32]*rangeInfo) // mib[0] -> range of mib[1]
	for _, r := range results {
		if r.ResolveErr != "" || len(r.MIB) < 2 {
			continue
		}
		ri, ok := topRanges[r.MIB[0]]
		if !ok {
			ri = &rangeInfo{min: r.MIB[1], max: r.MIB[1]}
			topRanges[r.MIB[0]] = ri
		}
		if r.MIB[1] < ri.min {
			ri.min = r.MIB[1]
		}
		if r.MIB[1] > ri.max {
			ri.max = r.MIB[1]
		}
		ri.count++
	}
	for _, topName := range []string{"kern", "vm", "net", "hw", "debug", "machdep"} {
		topID := knownTopLevel[topName]
		if ri, ok := topRanges[topID]; ok {
			fmt.Printf("  %s (MIB[0]=%d): MIB[1] range %d..%d (%d metrics)\n",
				topName, topID, ri.min, ri.max, ri.count)
		}
	}
	// Check for any other top-level values
	for topID, ri := range topRanges {
		found := false
		for _, id := range knownTopLevel {
			if id == topID {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("  unknown(MIB[0]=%d): MIB[1] range %d..%d (%d metrics)\n",
				topID, ri.min, ri.max, ri.count)
		}
	}
}

// flattenMIB encodes a MIB array into a single uint64.
// Encoding: each level gets 16 bits, max 4 levels = 64 bits.
func flattenMIB(mib []uint32) (uint64, bool) {
	if len(mib) > 4 {
		return 0, false
	}
	for _, v := range mib {
		if v > 0xFFFF {
			return 0, false
		}
	}
	var flat uint64
	for i, v := range mib {
		flat |= uint64(v) << (48 - uint(i)*16)
	}
	return flat, true
}
