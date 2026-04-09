//go:build darwin

package macosasmsysctl

import (
	"testing"
)

func TestMIBCache_Resolve(t *testing.T) {
	c := NewMIBCache()

	mib, err := c.Resolve("kern.ostype")
	if err != nil {
		t.Fatalf("Resolve(kern.ostype): %v", err)
	}
	if len(mib) == 0 {
		t.Fatal("MIB is empty")
	}
	t.Logf("kern.ostype MIB = %v", mib)

	// Second resolve should return cached value.
	mib2, err := c.Resolve("kern.ostype")
	if err != nil {
		t.Fatalf("second Resolve: %v", err)
	}
	if len(mib) != len(mib2) {
		t.Fatalf("MIB changed: %v vs %v", mib, mib2)
	}
	for i := range mib {
		if mib[i] != mib2[i] {
			t.Fatalf("MIB changed at %d: %v vs %v", i, mib, mib2)
		}
	}
}

func TestMIBCache_GetString(t *testing.T) {
	c := NewMIBCache()

	v, err := c.GetString("kern.ostype")
	if err != nil {
		t.Fatalf("GetString: %v", err)
	}
	if v != "Darwin" {
		t.Errorf("kern.ostype = %q, want Darwin", v)
	}
}

func TestMIBCache_GetUint64(t *testing.T) {
	c := NewMIBCache()

	v, err := c.GetUint64("hw.memsize")
	if err != nil {
		t.Fatalf("GetUint64: %v", err)
	}
	if v < 1<<30 {
		t.Errorf("hw.memsize = %d, want >= 1 GB", v)
	}
	t.Logf("hw.memsize = %d bytes (%.1f GB)", v, float64(v)/(1<<30))
}

func TestMIBCache_Warm(t *testing.T) {
	c := NewMIBCache()

	names := []string{
		"kern.ostype",
		"hw.memsize",
		"vm.page_free_count",
		"bogus.nonexistent",
	}
	resolved := c.Warm(names)
	if resolved != 3 {
		t.Errorf("Warm resolved %d, want 3", resolved)
	}
	t.Logf("warmed %d/%d names", resolved, len(names))
}

func TestMIBCache_MatchesByname(t *testing.T) {
	c := NewMIBCache()

	// Verify MIB-cached reads match direct sysctlbyname reads.
	tests := []string{"kern.ostype", "hw.memsize", "hw.ncpu", "kern.hostname"}
	for _, name := range tests {
		direct, err := GetRaw(name)
		if err != nil {
			t.Fatalf("GetRaw(%s): %v", name, err)
		}
		cached, err := c.GetRaw(name)
		if err != nil {
			t.Fatalf("cache.GetRaw(%s): %v", name, err)
		}
		if len(direct) != len(cached) {
			t.Errorf("%s: direct len=%d, cached len=%d", name, len(direct), len(cached))
			continue
		}
		for i := range direct {
			if direct[i] != cached[i] {
				t.Errorf("%s: byte %d differs: direct=%d, cached=%d", name, i, direct[i], cached[i])
				break
			}
		}
	}
}
