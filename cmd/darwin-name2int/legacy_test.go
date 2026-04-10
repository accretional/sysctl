//go:build darwin

package main

import (
	"bytes"
	"testing"

	"github.com/accretional/sysctl/internal/macosasmsysctl"
)

// TestLegacyMIBs reads each XNU header-defined MIB directly and compares
// against the sysctlbyname path to find divergences in value or size.
func TestLegacyMIBs(t *testing.T) {
	type legacyEntry struct {
		name      string
		legacyMIB []uint32 // from XNU header
	}

	entries := []legacyEntry{
		// CTL_KERN
		{"kern.ostype", []uint32{1, 1}},
		{"kern.osrelease", []uint32{1, 2}},
		{"kern.version", []uint32{1, 4}},
		{"kern.maxvnodes", []uint32{1, 5}},
		{"kern.maxproc", []uint32{1, 6}},
		{"kern.maxfiles", []uint32{1, 7}},
		{"kern.argmax", []uint32{1, 8}},
		{"kern.hostname", []uint32{1, 10}},
		{"kern.clockrate", []uint32{1, 12}},
		{"kern.boottime", []uint32{1, 21}},
		{"kern.maxfilesperproc", []uint32{1, 29}},
		{"kern.maxprocperuid", []uint32{1, 30}},
		{"kern.aiomax", []uint32{1, 46}},
		{"kern.aioprocmax", []uint32{1, 47}},
		{"kern.coredump", []uint32{1, 51}},
		{"kern.osversion", []uint32{1, 65}},
		{"kern.safeboot", []uint32{1, 66}},

		// CTL_HW
		{"hw.ncpu", []uint32{6, 3}},
		{"hw.byteorder", []uint32{6, 4}},
		{"hw.pagesize", []uint32{6, 7}},
		{"hw.memsize", []uint32{6, 24}},

		// CTL_VM
		{"vm.loadavg", []uint32{2, 2}},
		{"vm.swapusage", []uint32{2, 5}},

		// KERN_IPC
		{"kern.ipc.maxsockbuf", []uint32{1, 32, 1}},
		{"kern.ipc.somaxconn", []uint32{1, 32, 3}},
	}

	cache := macosasmsysctl.NewMIBCache()
	diverged := 0

	for _, e := range entries {
		// Read via legacy MIB
		legacyRaw, legacyErr := macosasmsysctl.GetRawMIB(e.legacyMIB)

		// Read via sysctlbyname
		nameRaw, nameErr := macosasmsysctl.GetRaw(e.name)

		// Read via resolved MIB
		resolvedMIB, _ := cache.Resolve(e.name)
		resolvedRaw, _ := cache.GetRaw(e.name)

		if legacyErr != nil {
			t.Logf("%-25s  legacy MIB %v: ERROR %v", e.name, e.legacyMIB, legacyErr)
			continue
		}
		if nameErr != nil {
			t.Logf("%-25s  sysctlbyname: ERROR %v", e.name, nameErr)
			continue
		}

		sameValue := bytes.Equal(legacyRaw, nameRaw)
		sameSize := len(legacyRaw) == len(nameRaw)
		sameMIB := mibEqual(e.legacyMIB, resolvedMIB)

		if sameValue && sameMIB {
			t.Logf("%-25s  ✓ legacy %v == resolved %v, %d bytes, identical", e.name, e.legacyMIB, resolvedMIB, len(legacyRaw))
		} else {
			diverged++
			if !sameMIB && sameSize && !sameValue {
				t.Logf("%-25s  ✗ MIB diverged: legacy %v → resolved %v, same size (%d), different bytes", e.name, e.legacyMIB, resolvedMIB, len(legacyRaw))
			} else if !sameMIB && !sameSize {
				t.Logf("%-25s  ✗ MIB diverged: legacy %v (%d bytes) → resolved %v (%d bytes), TYPE WIDENED", e.name, e.legacyMIB, len(legacyRaw), resolvedMIB, len(resolvedRaw))
			} else if !sameMIB && sameValue {
				t.Logf("%-25s  ✗ MIB diverged: legacy %v → resolved %v, but values identical (%d bytes)", e.name, e.legacyMIB, resolvedMIB, len(legacyRaw))
			} else {
				t.Logf("%-25s  ? legacy %v vs resolved %v, legacy=%d bytes, resolved=%d bytes", e.name, e.legacyMIB, resolvedMIB, len(legacyRaw), len(resolvedRaw))
			}
		}
	}
	t.Logf("diverged: %d / %d", diverged, len(entries))
}

func mibEqual(a, b []uint32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
