//go:build darwin

package macosasmsysctl

import (
	"testing"
)

func TestGetString_KernOstype(t *testing.T) {
	val, err := GetString("kern.ostype")
	if err != nil {
		t.Fatalf("GetString(kern.ostype): %v", err)
	}
	if val != "Darwin" {
		t.Errorf("kern.ostype = %q, want %q", val, "Darwin")
	}
}

func TestGetString_KernVersion(t *testing.T) {
	val, err := GetString("kern.version")
	if err != nil {
		t.Fatalf("GetString(kern.version): %v", err)
	}
	if val == "" {
		t.Error("kern.version returned empty string")
	}
	t.Logf("kern.version = %q", val)
}

func TestGetUint64_HwMemsize(t *testing.T) {
	val, err := GetUint64("hw.memsize")
	if err != nil {
		t.Fatalf("GetUint64(hw.memsize): %v", err)
	}
	// Sanity check: at least 1 GB.
	if val < 1<<30 {
		t.Errorf("hw.memsize = %d, expected at least 1 GB", val)
	}
	t.Logf("hw.memsize = %d bytes (%.1f GB)", val, float64(val)/(1<<30))
}

func TestGetInt32_HwNcpu(t *testing.T) {
	val, err := GetInt32("hw.ncpu")
	if err != nil {
		t.Fatalf("GetInt32(hw.ncpu): %v", err)
	}
	if val < 1 {
		t.Errorf("hw.ncpu = %d, expected at least 1", val)
	}
	t.Logf("hw.ncpu = %d", val)
}

func TestGetUint32_HwActivecpu(t *testing.T) {
	val, err := GetUint32("hw.activecpu")
	if err != nil {
		t.Fatalf("GetUint32(hw.activecpu): %v", err)
	}
	if val < 1 {
		t.Errorf("hw.activecpu = %d, expected at least 1", val)
	}
	t.Logf("hw.activecpu = %d", val)
}

func TestGetRaw_InvalidName(t *testing.T) {
	_, err := GetRaw("bogus.nonexistent.key.12345")
	if err == nil {
		t.Error("expected error for bogus sysctl name, got nil")
	}
}

func TestGetString_KernHostname(t *testing.T) {
	val, err := GetString("kern.hostname")
	if err != nil {
		t.Fatalf("GetString(kern.hostname): %v", err)
	}
	if val == "" {
		t.Error("kern.hostname returned empty")
	}
	t.Logf("kern.hostname = %q", val)
}

func TestGetUint64_HwPagesize(t *testing.T) {
	// hw.pagesize is often int32 on macOS, but let's try raw and check size.
	raw, err := GetRaw("hw.pagesize")
	if err != nil {
		t.Fatalf("GetRaw(hw.pagesize): %v", err)
	}
	t.Logf("hw.pagesize raw len = %d, bytes = %x", len(raw), raw)
}

func TestNametomib(t *testing.T) {
	mib, err := nametomib("kern.ostype")
	if err != nil {
		t.Fatalf("nametomib(kern.ostype): %v", err)
	}
	if len(mib) < 2 {
		t.Errorf("mib too short: %v", mib)
	}
	// kern = 1, ostype = 1
	if mib[0] != 1 || mib[1] != 1 {
		t.Errorf("mib = %v, want [1, 1]", mib)
	}
}
