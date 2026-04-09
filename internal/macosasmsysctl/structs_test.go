//go:build darwin

package macosasmsysctl

import (
	"testing"
	"time"
)

func TestGetTimeval_Boottime(t *testing.T) {
	tv, err := GetTimeval("kern.boottime")
	if err != nil {
		t.Fatalf("GetTimeval(kern.boottime): %v", err)
	}
	if tv.Sec == 0 {
		t.Error("boottime.Sec = 0, expected non-zero")
	}
	bt := tv.Time()
	if bt.Before(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("boottime = %v, expected after 2020", bt)
	}
	t.Logf("boottime = %v", bt)
}

func TestGetTimeval_Sleeptime(t *testing.T) {
	tv, err := GetTimeval("kern.sleeptime")
	if err != nil {
		t.Fatalf("GetTimeval(kern.sleeptime): %v", err)
	}
	t.Logf("sleeptime = %s (duration: %v)", tv, tv.Duration())
}

func TestGetLoadavg(t *testing.T) {
	la, err := GetLoadavg()
	if err != nil {
		t.Fatalf("GetLoadavg: %v", err)
	}
	if la.Load1 < 0 || la.Load1 > 1000 {
		t.Errorf("load1 = %.2f, out of sane range", la.Load1)
	}
	t.Logf("loadavg = %s", la)
}

func TestGetSwapUsage(t *testing.T) {
	su, err := GetSwapUsage()
	if err != nil {
		t.Fatalf("GetSwapUsage: %v", err)
	}
	if su.Total == 0 {
		t.Log("swap total = 0 (swap may be disabled)")
	}
	if su.Used > su.Total {
		t.Errorf("swap used (%d) > total (%d)", su.Used, su.Total)
	}
	t.Logf("swap = %s", su)
}

func TestGetClockinfo(t *testing.T) {
	ci, err := GetClockinfo()
	if err != nil {
		t.Fatalf("GetClockinfo: %v", err)
	}
	if ci.Hz <= 0 {
		t.Errorf("clockinfo.Hz = %d, expected > 0", ci.Hz)
	}
	t.Logf("clockinfo = %s", ci)
}
