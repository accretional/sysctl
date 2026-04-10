//go:build darwin

package main

import (
	"testing"

	"github.com/accretional/sysctl/internal/macosasmsysctl"
	"github.com/accretional/sysctl/internal/metrics"
)

// TestMetricsLargerThanUint64 lists all metrics whose raw sysctl value
// exceeds 8 bytes (the size of a uint64/int64).
func TestMetricsLargerThanUint64(t *testing.T) {
	cache := macosasmsysctl.NewMIBCache()
	count := 0
	for _, info := range metrics.Known {
		if info.Type == metrics.TypeComputed {
			continue
		}
		raw, err := cache.GetRaw(info.Name)
		if err != nil {
			continue
		}
		if len(raw) > 8 {
			count++
			t.Logf("%-45s  type=%-8s  size=%3d bytes", info.Name, info.Type, len(raw))
		}
	}
	t.Logf("total: %d metrics > 8 bytes", count)
}

// TestHWPagesizeDivergence investigates why hw.pagesize resolves to {6,115}
// instead of the XNU header constant HW_PAGESIZE=7.
func TestHWPagesizeDivergence(t *testing.T) {
	cache := macosasmsysctl.NewMIBCache()

	// Resolve all hw.pagesize* variants
	pagesizeNames := []string{"hw.pagesize", "hw.pagesize32"}
	for _, name := range pagesizeNames {
		mib, err := cache.Resolve(name)
		if err != nil {
			t.Logf("%s: resolve error: %v", name, err)
			continue
		}
		raw, _ := cache.GetRaw(name)
		t.Logf("%s -> MIB %v, value=%d (%d bytes)", name, mib, btoi(raw), len(raw))
	}

	// Try reading via the legacy MIB {6, 7} (HW_PAGESIZE from XNU header)
	legacyMIB := []uint32{6, 7}
	legacyRaw, err := macosasmsysctl.GetRawMIB(legacyMIB)
	if err != nil {
		t.Logf("legacy MIB {6,7} (HW_PAGESIZE): error: %v", err)
	} else {
		t.Logf("legacy MIB {6,7} (HW_PAGESIZE): value=%d (%d bytes)", btoi(legacyRaw), len(legacyRaw))
	}

	// Try reading via the resolved MIB {6, 115}
	resolvedMIB, _ := cache.Resolve("hw.pagesize")
	resolvedRaw, err := cache.GetRaw("hw.pagesize")
	if err != nil {
		t.Logf("resolved MIB %v: error: %v", resolvedMIB, err)
	} else {
		t.Logf("resolved MIB %v: value=%d (%d bytes)", resolvedMIB, btoi(resolvedRaw), len(resolvedRaw))
	}

	// Check nearby OIDs in the hw namespace to understand the numbering
	t.Log("")
	t.Log("--- hw.* OID neighborhood around 115 ---")
	nearby := []string{
		"hw.ncpu",         // HW_NCPU=3
		"hw.byteorder",    // HW_BYTEORDER=4
		"hw.memsize",      // HW_MEMSIZE=24
		"hw.activecpu",    // HW_AVAILCPU=25
		"hw.physicalcpu",  // no legacy constant
		"hw.logicalcpu",   // no legacy constant
		"hw.cachelinesize", // no legacy constant
		"hw.pagesize",     // HW_PAGESIZE=7 legacy, but resolves to 115
		"hw.pagesize32",   // no legacy constant
	}
	for _, name := range nearby {
		mib, err := cache.Resolve(name)
		if err != nil {
			continue
		}
		t.Logf("  %-25s -> MIB %v", name, mib)
	}

	// Check if legacy {6,7} and new {6,115} return the same value
	if legacyRaw != nil && resolvedRaw != nil {
		if btoi(legacyRaw) == btoi(resolvedRaw) {
			t.Log("")
			t.Log("CONCLUSION: legacy {6,7} and resolved {6,115} return the SAME value.")
			t.Log("The kernel registers hw.pagesize as a new-style OID (115) but")
			t.Log("the legacy HW_PAGESIZE=7 still works and returns the same data.")
			t.Logf("Legacy returns %d bytes, resolved returns %d bytes", len(legacyRaw), len(resolvedRaw))
		} else {
			t.Logf("WARNING: legacy {6,7} returns %d, resolved {6,115} returns %d — different values!",
				btoi(legacyRaw), btoi(resolvedRaw))
		}
	}
}

func btoi(raw []byte) int64 {
	if len(raw) == 4 {
		return int64(int32(raw[0]) | int32(raw[1])<<8 | int32(raw[2])<<16 | int32(raw[3])<<24)
	}
	if len(raw) == 8 {
		var v uint64
		for i := 0; i < 8; i++ {
			v |= uint64(raw[i]) << (uint(i) * 8)
		}
		return int64(v)
	}
	return 0
}
