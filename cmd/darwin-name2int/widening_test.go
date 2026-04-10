//go:build darwin

package main

import (
	"testing"

	"github.com/accretional/sysctl/internal/macosasmsysctl"
	"github.com/accretional/sysctl/internal/metrics"
)

// TestTypeWidening checks which hw.* metrics use OIDs beyond the legacy
// HW_* range (>25), indicating they were re-registered as new-style OIDs.
// Cross-references with the CLAUDE.md note about ARM64 type widening.
func TestTypeWidening(t *testing.T) {
	cache := macosasmsysctl.NewMIBCache()

	// Legacy HW_* constants max out at HW_PRODUCT=27
	const legacyHWMax = 27

	t.Log("=== hw.* metrics with OID > legacy range (MIB[1] > 27) ===")
	widened := 0
	for _, info := range metrics.Known {
		if info.Type == metrics.TypeComputed || info.Category == "" {
			continue
		}
		mib, err := cache.Resolve(info.Name)
		if err != nil || len(mib) < 2 {
			continue
		}
		// Only check hw.* namespace
		if mib[0] != 6 {
			continue
		}
		if mib[1] > legacyHWMax {
			raw, _ := cache.GetRaw(info.Name)
			t.Logf("  %-40s MIB {6,%3d}  %d bytes  type=%s", info.Name, mib[1], len(raw), info.Type)
			widened++
		}
	}
	t.Logf("%d hw.* metrics use new-style OIDs (MIB[1] > %d)", widened, legacyHWMax)

	// Now check which specific hw.* metrics share the legacy MIB IDs (1-27)
	t.Log("")
	t.Log("=== hw.* metrics in legacy OID range (MIB[1] <= 27) ===")
	legacy := 0
	for _, info := range metrics.Known {
		if info.Type == metrics.TypeComputed {
			continue
		}
		mib, err := cache.Resolve(info.Name)
		if err != nil || len(mib) < 2 {
			continue
		}
		if mib[0] != 6 {
			continue
		}
		if mib[1] <= legacyHWMax {
			raw, _ := cache.GetRaw(info.Name)
			t.Logf("  %-40s MIB {6,%3d}  %d bytes  type=%s", info.Name, mib[1], len(raw), info.Type)
			legacy++
		}
	}
	t.Logf("%d hw.* metrics use legacy OIDs (MIB[1] <= %d)", legacy, legacyHWMax)
}
