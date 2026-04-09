//go:build darwin

package macosasmsysctl

import (
	"encoding/binary"
	"fmt"
	"sync"
)

// MIBCache caches sysctl name-to-MIB resolutions for faster repeated reads.
// Using pre-resolved MIBs avoids the name lookup on every read (~3x faster).
type MIBCache struct {
	mu    sync.RWMutex
	cache map[string][]uint32
}

// NewMIBCache returns a new empty MIB cache.
func NewMIBCache() *MIBCache {
	return &MIBCache{cache: make(map[string][]uint32)}
}

// Resolve returns the cached MIB for a name, resolving and caching it on first call.
func (c *MIBCache) Resolve(name string) ([]uint32, error) {
	c.mu.RLock()
	mib, ok := c.cache[name]
	c.mu.RUnlock()
	if ok {
		return mib, nil
	}

	mib, err := nametomib(name)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.cache[name] = mib
	c.mu.Unlock()
	return mib, nil
}

// Warm pre-resolves a list of names, caching all MIBs.
// Returns the count of successfully resolved names. Errors are silently skipped
// (the metric may not exist on this hardware).
func (c *MIBCache) Warm(names []string) int {
	resolved := 0
	for _, name := range names {
		if _, err := c.Resolve(name); err == nil {
			resolved++
		}
	}
	return resolved
}

// getRawMIB reads a sysctl value using a pre-resolved MIB array.
func getRawMIB(mib []uint32) ([]byte, error) {
	// First call: determine size.
	var n uintptr
	if err := rawSysctl(mib, nil, &n, nil, 0); err != nil {
		return nil, fmt.Errorf("sysctl size mib %v: %w", mib, err)
	}
	if n == 0 {
		return nil, nil
	}

	// Second call: read value.
	buf := make([]byte, n)
	if err := rawSysctl(mib, &buf[0], &n, nil, 0); err != nil {
		return nil, fmt.Errorf("sysctl read mib %v: %w", mib, err)
	}
	return buf[:n], nil
}

// GetRaw reads raw bytes using the MIB cache.
func (c *MIBCache) GetRaw(name string) ([]byte, error) {
	mib, err := c.Resolve(name)
	if err != nil {
		return nil, err
	}
	return getRawMIB(mib)
}

// GetString reads a string value using the MIB cache.
func (c *MIBCache) GetString(name string) (string, error) {
	raw, err := c.GetRaw(name)
	if err != nil {
		return "", err
	}
	if len(raw) > 0 && raw[len(raw)-1] == 0 {
		raw = raw[:len(raw)-1]
	}
	return string(raw), nil
}

// GetUint32 reads a uint32 value using the MIB cache.
func (c *MIBCache) GetUint32(name string) (uint32, error) {
	raw, err := c.GetRaw(name)
	if err != nil {
		return 0, err
	}
	if len(raw) != 4 {
		return 0, fmt.Errorf("sysctl %q: expected 4 bytes, got %d", name, len(raw))
	}
	return binary.LittleEndian.Uint32(raw), nil
}

// GetUint64 reads a uint64 value using the MIB cache.
func (c *MIBCache) GetUint64(name string) (uint64, error) {
	raw, err := c.GetRaw(name)
	if err != nil {
		return 0, err
	}
	if len(raw) != 8 {
		return 0, fmt.Errorf("sysctl %q: expected 8 bytes, got %d", name, len(raw))
	}
	return binary.LittleEndian.Uint64(raw), nil
}

// GetInt32 reads an int32 value using the MIB cache.
func (c *MIBCache) GetInt32(name string) (int32, error) {
	v, err := c.GetUint32(name)
	return int32(v), err
}

// GetInt64 reads an int64 value using the MIB cache.
func (c *MIBCache) GetInt64(name string) (int64, error) {
	v, err := c.GetUint64(name)
	return int64(v), err
}

// GetTimeval reads a timeval struct using the MIB cache.
func (c *MIBCache) GetTimeval(name string) (Timeval, error) {
	raw, err := c.GetRaw(name)
	if err != nil {
		return Timeval{}, err
	}
	if len(raw) != 16 {
		return Timeval{}, fmt.Errorf("sysctl %q: expected 16 bytes, got %d", name, len(raw))
	}
	return Timeval{
		Sec:  binary.LittleEndian.Uint64(raw[0:8]),
		Usec: binary.LittleEndian.Uint64(raw[8:16]),
	}, nil
}

// GetLoadavg reads vm.loadavg using the MIB cache.
func (c *MIBCache) GetLoadavg() (Loadavg, error) {
	raw, err := c.GetRaw("vm.loadavg")
	if err != nil {
		return Loadavg{}, err
	}
	if len(raw) != 24 {
		return Loadavg{}, fmt.Errorf("vm.loadavg: expected 24 bytes, got %d", len(raw))
	}
	l1 := binary.LittleEndian.Uint32(raw[0:4])
	l5 := binary.LittleEndian.Uint32(raw[4:8])
	l15 := binary.LittleEndian.Uint32(raw[8:12])
	fscale := binary.LittleEndian.Uint64(raw[16:24])
	if fscale == 0 {
		fscale = 2048
	}
	return Loadavg{
		Load1:  float64(l1) / float64(fscale),
		Load5:  float64(l5) / float64(fscale),
		Load15: float64(l15) / float64(fscale),
	}, nil
}

// GetSwapUsage reads vm.swapusage using the MIB cache.
func (c *MIBCache) GetSwapUsage() (SwapUsage, error) {
	raw, err := c.GetRaw("vm.swapusage")
	if err != nil {
		return SwapUsage{}, err
	}
	if len(raw) != 32 {
		return SwapUsage{}, fmt.Errorf("vm.swapusage: expected 32 bytes, got %d", len(raw))
	}
	return SwapUsage{
		Total: binary.LittleEndian.Uint64(raw[0:8]),
		Avail: binary.LittleEndian.Uint64(raw[8:16]),
		Used:  binary.LittleEndian.Uint64(raw[16:24]),
		Flags: binary.LittleEndian.Uint64(raw[24:32]),
	}, nil
}

// GetClockinfo reads kern.clockrate using the MIB cache.
func (c *MIBCache) GetClockinfo() (Clockinfo, error) {
	raw, err := c.GetRaw("kern.clockrate")
	if err != nil {
		return Clockinfo{}, err
	}
	if len(raw) != 20 {
		return Clockinfo{}, fmt.Errorf("kern.clockrate: expected 20 bytes, got %d", len(raw))
	}
	return Clockinfo{
		Hz:     int32(binary.LittleEndian.Uint32(raw[0:4])),
		Tick:   int32(binary.LittleEndian.Uint32(raw[4:8])),
		Spare:  int32(binary.LittleEndian.Uint32(raw[8:12])),
		Profhz: int32(binary.LittleEndian.Uint32(raw[12:16])),
		Stathz: int32(binary.LittleEndian.Uint32(raw[16:20])),
	}, nil
}
