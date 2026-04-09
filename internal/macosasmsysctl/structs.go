//go:build darwin

package macosasmsysctl

import (
	"encoding/binary"
	"fmt"
	"time"
)

// Timeval represents a kernel timeval64 (kern.boottime, kern.sleeptime, etc.).
type Timeval struct {
	Sec  uint64 // Seconds (epoch for boottime, duration for sleeptime)
	Usec uint64 // Microseconds
}

// Time converts to Go time.Time (only meaningful for epoch-based values like boottime).
func (t Timeval) Time() time.Time {
	return time.Unix(int64(t.Sec), int64(t.Usec)*1000)
}

// Duration converts to Go time.Duration (for duration-based values).
func (t Timeval) Duration() time.Duration {
	return time.Duration(t.Sec)*time.Second + time.Duration(t.Usec)*time.Microsecond
}

func (t Timeval) String() string {
	return fmt.Sprintf("{Sec: %d, Usec: %d}", t.Sec, t.Usec)
}

// Loadavg represents the vm.loadavg struct.
type Loadavg struct {
	Load1  float64 // 1-minute load average
	Load5  float64 // 5-minute load average
	Load15 float64 // 15-minute load average
}

func (l Loadavg) String() string {
	return fmt.Sprintf("%.2f %.2f %.2f", l.Load1, l.Load5, l.Load15)
}

// SwapUsage represents the vm.swapusage struct (xsw_usage).
type SwapUsage struct {
	Total uint64 // Total swap in bytes
	Avail uint64 // Available swap in bytes
	Used  uint64 // Used swap in bytes
	Flags uint64 // Swap flags
}

func (s SwapUsage) String() string {
	return fmt.Sprintf("total=%d avail=%d used=%d", s.Total, s.Avail, s.Used)
}

// Clockinfo represents the kern.clockrate struct.
type Clockinfo struct {
	Hz     int32 // Clock frequency (typically 100)
	Tick   int32 // Microseconds per tick
	Spare  int32 // (was tickadj)
	Profhz int32 // Profiling clock frequency
	Stathz int32 // Statistics clock frequency
}

func (c Clockinfo) String() string {
	return fmt.Sprintf("hz=%d tick=%d profhz=%d stathz=%d", c.Hz, c.Tick, c.Profhz, c.Stathz)
}

// GetTimeval reads a sysctl that returns a timeval struct (16 bytes).
func GetTimeval(name string) (Timeval, error) {
	raw, err := GetRaw(name)
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

// GetLoadavg reads vm.loadavg and parses the struct.
func GetLoadavg() (Loadavg, error) {
	raw, err := GetRaw("vm.loadavg")
	if err != nil {
		return Loadavg{}, err
	}
	if len(raw) != 24 {
		return Loadavg{}, fmt.Errorf("vm.loadavg: expected 24 bytes, got %d", len(raw))
	}
	// struct loadavg { fixpt_t ldavg[3]; long fscale; }
	// On ARM64: fixpt_t = uint32, long = int64.
	// Layout: [3]uint32 (12 bytes) + 4 bytes padding + int64 fscale
	l1 := binary.LittleEndian.Uint32(raw[0:4])
	l5 := binary.LittleEndian.Uint32(raw[4:8])
	l15 := binary.LittleEndian.Uint32(raw[8:12])
	fscale := binary.LittleEndian.Uint64(raw[16:24])
	if fscale == 0 {
		fscale = 2048 // fallback
	}
	return Loadavg{
		Load1:  float64(l1) / float64(fscale),
		Load5:  float64(l5) / float64(fscale),
		Load15: float64(l15) / float64(fscale),
	}, nil
}

// GetSwapUsage reads vm.swapusage and parses the xsw_usage struct.
func GetSwapUsage() (SwapUsage, error) {
	raw, err := GetRaw("vm.swapusage")
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

// GetClockinfo reads kern.clockrate and parses the clockinfo struct.
func GetClockinfo() (Clockinfo, error) {
	raw, err := GetRaw("kern.clockrate")
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
