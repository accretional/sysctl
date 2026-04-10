//go:build darwin

package metrics

import (
	"testing"

	pb "github.com/accretional/sysctl/proto/sysctlpb"
)

func TestLoadKernelRegistry(t *testing.T) {
	reg, err := LoadKernelRegistry("24.6.0")
	if err != nil {
		t.Fatalf("LoadKernelRegistry: %v", err)
	}
	if reg.OsRegistry != "darwin-arm64" {
		t.Errorf("os_registry = %q, want darwin-arm64", reg.OsRegistry)
	}
	if len(reg.Metrics) == 0 {
		t.Fatal("no metrics in kernel registry")
	}
	t.Logf("loaded %d kernel metrics for %s (%s)", len(reg.Metrics), reg.OsRegistry, reg.OsVersion)
}

func TestKernelRegistryMatchesKnown(t *testing.T) {
	reg, err := LoadKernelRegistry("24.6.0")
	if err != nil {
		t.Fatalf("LoadKernelRegistry: %v", err)
	}

	if len(reg.Metrics) != len(Known) {
		t.Errorf("kernel registry has %d metrics, Known has %d", len(reg.Metrics), len(Known))
	}

	issues := ValidateRegistry(reg)
	for _, issue := range issues {
		t.Error(issue)
	}
	if len(issues) == 0 {
		t.Logf("kernel registry and Known registry are fully consistent (%d metrics)", len(Known))
	}
}

func TestKernelRegistryAccessPatterns(t *testing.T) {
	reg, err := LoadKernelRegistry("24.6.0")
	if err != nil {
		t.Fatalf("LoadKernelRegistry: %v", err)
	}

	byName := KernelRegistryByName(reg)

	// Spot-check: hw.memsize should be STATIC/STATIC
	if km, ok := byName["hw.memsize"]; ok {
		if km.KernelAccessPattern.Pattern != pb.AccessPattern_STATIC {
			t.Errorf("hw.memsize kernel pattern = %v, want STATIC", km.KernelAccessPattern.Pattern)
		}
		if km.RecommendedAccessPattern.Pattern != pb.AccessPattern_STATIC {
			t.Errorf("hw.memsize recommended pattern = %v, want STATIC", km.RecommendedAccessPattern.Pattern)
		}
	} else {
		t.Error("hw.memsize not in kernel registry")
	}

	// Spot-check: vm.page_free_count should be DYNAMIC/POLLED
	if km, ok := byName["vm.page_free_count"]; ok {
		if km.KernelAccessPattern.Pattern != pb.AccessPattern_DYNAMIC {
			t.Errorf("vm.page_free_count kernel pattern = %v, want DYNAMIC", km.KernelAccessPattern.Pattern)
		}
		if km.RecommendedAccessPattern.Pattern != pb.AccessPattern_POLLED {
			t.Errorf("vm.page_free_count recommended pattern = %v, want POLLED", km.RecommendedAccessPattern.Pattern)
		}
		if km.RecommendedAccessPattern.Ttl == nil {
			t.Error("vm.page_free_count recommended TTL is nil")
		} else if km.RecommendedAccessPattern.Ttl.Seconds != 10 {
			t.Errorf("vm.page_free_count recommended TTL = %ds, want 10s", km.RecommendedAccessPattern.Ttl.Seconds)
		}
	} else {
		t.Error("vm.page_free_count not in kernel registry")
	}

	// Count patterns
	counts := make(map[pb.AccessPattern]int)
	for _, km := range reg.Metrics {
		if km.RecommendedAccessPattern != nil {
			counts[km.RecommendedAccessPattern.Pattern]++
		}
	}
	t.Logf("recommended patterns: STATIC=%d POLLED=%d CONSTRAINED=%d DYNAMIC=%d DISABLED=%d",
		counts[pb.AccessPattern_STATIC],
		counts[pb.AccessPattern_POLLED],
		counts[pb.AccessPattern_CONSTRAINED],
		counts[pb.AccessPattern_DYNAMIC],
		counts[pb.AccessPattern_DISABLED])
}
