//go:build darwin

package darwin_test

import (
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/accretional/sysctl/internal/macosasmsysctl"
	"github.com/accretional/sysctl/internal/metrics"
	pb "github.com/accretional/sysctl/proto/sysctlpb"
)

// ---------------------------------------------------------------------------
// Shared infrastructure
// ---------------------------------------------------------------------------

var (
	cache     *macosasmsysctl.MIBCache
	regByName map[string]*pb.KernelMetric
)

func TestMain(m *testing.M) {
	cache = macosasmsysctl.NewMIBCache()

	reg, err := metrics.LoadKernelRegistry("24.6.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "load kernel registry: %v\n", err)
		os.Exit(1)
	}
	regByName = metrics.KernelRegistryByName(reg)

	// Pre-warm MIB cache for all non-computed metrics.
	var names []string
	for _, info := range metrics.Known {
		if info.Type != metrics.TypeComputed {
			names = append(names, info.Name)
		}
	}
	cache.Warm(names)

	code := m.Run()
	eval.writeCSV()
	os.Exit(code)
}

// ---------------------------------------------------------------------------
// Eval CSV logger — mutex-locked, writes on deferred flush after all tests
// ---------------------------------------------------------------------------

type evalRecord struct {
	TestType    string // "normal" or "provoked"
	Category    string
	Name        string
	ValueType   string
	KernelPat   string
	RecPat      string
	TTLSec      string
	Read1       string
	Read2       string
	Changed     string
	Expected    string
	Notes       string
}

type evalLogger struct {
	mu      sync.Mutex
	records []evalRecord
}

var eval = &evalLogger{}

func (l *evalLogger) add(r evalRecord) {
	l.mu.Lock()
	l.records = append(l.records, r)
	l.mu.Unlock()
}

func (l *evalLogger) writeCSV() {
	f, err := os.Create("24.6.0.eval.csv")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create eval csv: %v\n", err)
		return
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()
	w.Write([]string{
		"test_type", "category", "name", "value_type",
		"kernel_pattern", "recommended_pattern", "ttl_seconds",
		"read1", "read2", "changed", "expected", "notes",
	})
	for _, r := range l.records {
		w.Write([]string{
			r.TestType, r.Category, r.Name, r.ValueType,
			r.KernelPat, r.RecPat, r.TTLSec,
			r.Read1, r.Read2, r.Changed, r.Expected, r.Notes,
		})
	}
}

// ---------------------------------------------------------------------------
// Reading helpers
// ---------------------------------------------------------------------------

func readValue(info *metrics.Info) (string, error) {
	switch info.Type {
	case metrics.TypeString:
		v, err := cache.GetString(info.Name)
		return v, err
	case metrics.TypeUint32:
		v, err := cache.GetUint32(info.Name)
		return fmt.Sprintf("%d", v), err
	case metrics.TypeUint64:
		v, err := cache.GetUint64(info.Name)
		return fmt.Sprintf("%d", v), err
	case metrics.TypeInt32:
		v, err := cache.GetInt32(info.Name)
		return fmt.Sprintf("%d", v), err
	case metrics.TypeInt64:
		v, err := cache.GetInt64(info.Name)
		return fmt.Sprintf("%d", v), err
	case metrics.TypeTimeval:
		tv, err := cache.GetTimeval(info.Name)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d.%06d", tv.Sec, tv.Usec), nil
	case metrics.TypeLoadavg:
		la, err := cache.GetLoadavg()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%.4f|%.4f|%.4f", la.Load1, la.Load5, la.Load15), nil
	case metrics.TypeSwap:
		su, err := cache.GetSwapUsage()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d|%d|%d", su.Total, su.Avail, su.Used), nil
	case metrics.TypeClock:
		ci, err := cache.GetClockinfo()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d|%d|%d|%d", ci.Hz, ci.Tick, ci.Profhz, ci.Stathz), nil
	case metrics.TypeRaw:
		raw, err := cache.GetRaw(info.Name)
		if err != nil {
			return "", err
		}
		return hex.EncodeToString(raw), nil
	case metrics.TypeComputed:
		return "", fmt.Errorf("computed: derived from other metrics")
	}
	return "", fmt.Errorf("unknown type %s", info.Type)
}

// defaultExpect returns the expected change behavior for a metric
// during 1s of normal (non-provoked) operation.
func defaultExpect(name string, cat metrics.Category, km *pb.KernelMetric) string {
	if km != nil && km.KernelAccessPattern != nil &&
		km.KernelAccessPattern.Pattern == pb.AccessPattern_STATIC {
		return "no"
	}

	// Monotonic timers — always increment.
	switch name {
	case "kern.monotonicclock_usecs", "machdep.time_since_reset":
		return "likely"
	}

	// Security hooks fire on every process exec / dylib load.
	if cat == metrics.CatSecurity {
		return "likely"
	}

	// VFS allocation counters accumulate constantly.
	switch name {
	case "vfs.vnstats.num_newvnode_calls", "vfs.vnstats.num_recycledvnodes":
		return "likely"
	}

	// VM page counters fluctuate under normal activity.
	switch name {
	case "vm.page_free_count", "vm.pages_grabbed":
		return "likely"
	}

	// Tunables that only change via explicit sysctl -w or admin action.
	switch cat {
	case metrics.CatKernLimits, metrics.CatKernIPC, metrics.CatKernSched,
		metrics.CatDebug, metrics.CatKPC, metrics.CatIOGPU, metrics.CatNetIP:
		return "unlikely"
	case metrics.CatKernMisc:
		return "unlikely"
	}

	// Specific metrics that rarely change.
	switch name {
	case "vm.compressor_mode", "vm.compressor_segment_limit",
		"vm.swap_enabled", "vm.swapfileprefix",
		"kern.hostname":
		return "unlikely"
	}

	// Wire limits are fixed until admin changes them.
	if cat == metrics.CatVMWire {
		switch name {
		case "vm.global_user_wire_limit", "vm.user_wire_limit",
			"vm.global_no_user_wire_amount":
			return "unlikely"
		}
	}

	// Network config tunables (not counters).
	if cat == metrics.CatNetTCP {
		switch name {
		case "net.inet.tcp.pcbcount", "net.inet.tcp.sack_globalholes",
			"net.inet.tcp.cubic_sockets":
			return "possible"
		default:
			return "unlikely"
		}
	}
	if cat == metrics.CatNetUDP {
		if name == "net.inet.udp.pcbcount" {
			return "possible"
		}
		return "unlikely"
	}
	if cat == metrics.CatNetMisc {
		return "possible"
	}

	return "possible"
}

func ttlSeconds(km *pb.KernelMetric) string {
	if km == nil || km.RecommendedAccessPattern == nil || km.RecommendedAccessPattern.Ttl == nil {
		return ""
	}
	return fmt.Sprintf("%d", km.RecommendedAccessPattern.Ttl.Seconds)
}

func patternStr(ac *pb.AccessConfig) string {
	if ac == nil {
		return "UNSPECIFIED"
	}
	return ac.Pattern.String()
}

// ---------------------------------------------------------------------------
// Core test helper
// ---------------------------------------------------------------------------

func testCategory(t *testing.T, cat metrics.Category, testType string, provoke func(), overrides map[string]string) {
	t.Helper()
	infos := metrics.ByCategory(cat)
	if len(infos) == 0 {
		t.Skipf("no metrics in category %s", cat)
	}

	// --- Read 1 ---
	values1 := make(map[string]string, len(infos))
	errors1 := make(map[string]string, len(infos))
	for i := range infos {
		v, err := readValue(&infos[i])
		if err != nil {
			errors1[infos[i].Name] = err.Error()
		} else {
			values1[infos[i].Name] = v
		}
	}

	// --- Provoke changes (if any) ---
	if provoke != nil {
		provoke()
	}

	// --- Wait ---
	time.Sleep(1 * time.Second)

	// --- Read 2 + compare ---
	for i := range infos {
		info := &infos[i]
		km := regByName[info.Name]

		// Determine expected change.
		expected := defaultExpect(info.Name, cat, km)
		if ov, ok := overrides[info.Name]; ok {
			expected = ov
		}

		// Handle first-read errors.
		if errMsg, hadErr := errors1[info.Name]; hadErr {
			eval.add(evalRecord{
				TestType: testType, Category: string(cat), Name: info.Name,
				ValueType: string(info.Type), KernelPat: patternStr(km.GetKernelAccessPattern()),
				RecPat: patternStr(km.GetRecommendedAccessPattern()), TTLSec: ttlSeconds(km),
				Read1: "ERROR", Read2: "", Changed: "N/A", Expected: expected,
				Notes: errMsg,
			})
			continue
		}

		v2, err := readValue(info)
		if err != nil {
			eval.add(evalRecord{
				TestType: testType, Category: string(cat), Name: info.Name,
				ValueType: string(info.Type), KernelPat: patternStr(km.GetKernelAccessPattern()),
				RecPat: patternStr(km.GetRecommendedAccessPattern()), TTLSec: ttlSeconds(km),
				Read1: values1[info.Name], Read2: "ERROR", Changed: "N/A", Expected: expected,
				Notes: err.Error(),
			})
			continue
		}

		v1 := values1[info.Name]
		changed := v1 != v2

		// STATIC metrics must not change.
		if km != nil && km.KernelAccessPattern != nil &&
			km.KernelAccessPattern.Pattern == pb.AccessPattern_STATIC && changed {
			t.Errorf("STATIC metric %s changed: %q -> %q", info.Name, v1, v2)
		}

		notes := ""
		if changed && expected == "no" {
			notes = "UNEXPECTED CHANGE"
		}

		eval.add(evalRecord{
			TestType: testType, Category: string(cat), Name: info.Name,
			ValueType: string(info.Type), KernelPat: patternStr(km.GetKernelAccessPattern()),
			RecPat: patternStr(km.GetRecommendedAccessPattern()), TTLSec: ttlSeconds(km),
			Read1: v1, Read2: v2, Changed: fmt.Sprintf("%t", changed),
			Expected: expected, Notes: notes,
		})

		t.Logf("%-45s %-8s changed=%-5t expected=%s", info.Name, string(info.Type), changed, expected)
	}
}

// ---------------------------------------------------------------------------
// Normal category tests — each category gets its own test function
// ---------------------------------------------------------------------------

func TestCategory_HWCPU(t *testing.T) {
	testCategory(t, metrics.CatHWCPU, "normal", nil, nil)
}

func TestCategory_HWMemory(t *testing.T) {
	testCategory(t, metrics.CatHWMemory, "normal", nil, nil)
}

func TestCategory_HWCache(t *testing.T) {
	testCategory(t, metrics.CatHWCache, "normal", nil, nil)
}

func TestCategory_HWPerflevel(t *testing.T) {
	testCategory(t, metrics.CatHWPerflevel, "normal", nil, nil)
}

func TestCategory_HWARM(t *testing.T) {
	testCategory(t, metrics.CatHWARM, "normal", nil, nil)
}

func TestCategory_KernIdentity(t *testing.T) {
	testCategory(t, metrics.CatKernIdentity, "normal", nil, nil)
}

func TestCategory_KernProcess(t *testing.T) {
	testCategory(t, metrics.CatKernProcess, "normal", nil, nil)
}

func TestCategory_KernMemory(t *testing.T) {
	testCategory(t, metrics.CatKernMemory, "normal", nil, nil)
}

func TestCategory_KernSched(t *testing.T) {
	testCategory(t, metrics.CatKernSched, "normal", nil, nil)
}

func TestCategory_KernTime(t *testing.T) {
	testCategory(t, metrics.CatKernTime, "normal", nil, map[string]string{
		"kern.sleeptime":             "unlikely",
		"kern.waketime":              "unlikely",
		"kern.wake_abs_time":         "unlikely",
		"kern.sleep_abs_time":        "unlikely",
		"kern.useractive_abs_time":   "possible",
		"kern.userinactive_abs_time": "possible",
	})
}

func TestCategory_KernLimits(t *testing.T) {
	testCategory(t, metrics.CatKernLimits, "normal", nil, nil)
}

func TestCategory_KernIPC(t *testing.T) {
	testCategory(t, metrics.CatKernIPC, "normal", nil, nil)
}

func TestCategory_KernMisc(t *testing.T) {
	testCategory(t, metrics.CatKernMisc, "normal", nil, nil)
}

func TestCategory_VMPressure(t *testing.T) {
	testCategory(t, metrics.CatVMPressure, "normal", nil, nil)
}

func TestCategory_VMPages(t *testing.T) {
	testCategory(t, metrics.CatVMPages, "normal", nil, nil)
}

func TestCategory_VMCompressor(t *testing.T) {
	testCategory(t, metrics.CatVMCompressor, "normal", nil, nil)
}

func TestCategory_VMPageout(t *testing.T) {
	testCategory(t, metrics.CatVMPageout, "normal", nil, nil)
}

func TestCategory_VMSwap(t *testing.T) {
	testCategory(t, metrics.CatVMSwap, "normal", nil, nil)
}

func TestCategory_VMWire(t *testing.T) {
	testCategory(t, metrics.CatVMWire, "normal", nil, nil)
}

func TestCategory_VMMisc(t *testing.T) {
	testCategory(t, metrics.CatVMMisc, "normal", nil, nil)
}

func TestCategory_Machdep(t *testing.T) {
	testCategory(t, metrics.CatMachdep, "normal", nil, map[string]string{
		"machdep.wake_abstime":  "unlikely",
		"machdep.wake_conttime": "unlikely",
	})
}

func TestCategory_NetTCP(t *testing.T) {
	testCategory(t, metrics.CatNetTCP, "normal", nil, nil)
}

func TestCategory_NetUDP(t *testing.T) {
	testCategory(t, metrics.CatNetUDP, "normal", nil, nil)
}

func TestCategory_NetIP(t *testing.T) {
	testCategory(t, metrics.CatNetIP, "normal", nil, nil)
}

func TestCategory_NetMisc(t *testing.T) {
	testCategory(t, metrics.CatNetMisc, "normal", nil, nil)
}

func TestCategory_VFS(t *testing.T) {
	testCategory(t, metrics.CatVFS, "normal", nil, nil)
}

func TestCategory_Debug(t *testing.T) {
	testCategory(t, metrics.CatDebug, "normal", nil, nil)
}

func TestCategory_KPC(t *testing.T) {
	testCategory(t, metrics.CatKPC, "normal", nil, nil)
}

func TestCategory_IOGPU(t *testing.T) {
	testCategory(t, metrics.CatIOGPU, "normal", nil, nil)
}

func TestCategory_Security(t *testing.T) {
	testCategory(t, metrics.CatSecurity, "normal", nil, nil)
}

func TestCategory_Computed(t *testing.T) {
	// Computed metrics are server-side aggregates derived from other sysctls.
	// Their component metrics are already tested in their respective categories.
	// Log each as N/A in the CSV.
	infos := metrics.ByCategory(metrics.CatComputed)
	for _, info := range infos {
		km := regByName[info.Name]
		eval.add(evalRecord{
			TestType: "normal", Category: string(metrics.CatComputed), Name: info.Name,
			ValueType: string(info.Type), KernelPat: patternStr(km.GetKernelAccessPattern()),
			RecPat: patternStr(km.GetRecommendedAccessPattern()), TTLSec: ttlSeconds(km),
			Read1: "N/A", Read2: "N/A", Changed: "N/A", Expected: "N/A",
			Notes: "server-side aggregate; components tested in source categories",
		})
	}
	t.Logf("skipped %d computed metrics (tested via their source categories)", len(infos))
}

// ---------------------------------------------------------------------------
// Provocative tests — deliberately induce metric changes
// ---------------------------------------------------------------------------

func TestCategory_VMPages_Provoked(t *testing.T) {
	testCategory(t, metrics.CatVMPages, "provoked", func() {
		// Allocate and touch 100MB to force page table changes.
		buf := make([]byte, 100*1024*1024)
		for i := 0; i < len(buf); i += 4096 {
			buf[i] = byte(i)
		}
		runtime.KeepAlive(buf)
	}, map[string]string{
		"vm.page_free_count": "likely",
		"vm.pages_grabbed":   "likely",
	})
}

func TestCategory_KernProcess_Provoked(t *testing.T) {
	testCategory(t, metrics.CatKernProcess, "provoked", func() {
		// Open 100 temp files to increase kern.num_files.
		files := make([]*os.File, 100)
		for i := range files {
			f, err := os.CreateTemp("", "sysctl-eval-*")
			if err != nil {
				continue
			}
			files[i] = f
		}
		t.Cleanup(func() {
			for _, f := range files {
				if f != nil {
					f.Close()
					os.Remove(f.Name())
				}
			}
		})
	}, map[string]string{
		"kern.num_files": "likely",
	})
}

func TestCategory_VFS_Provoked(t *testing.T) {
	testCategory(t, metrics.CatVFS, "provoked", func() {
		// Create and stat temp files to drive vnode allocation.
		dir, err := os.MkdirTemp("", "sysctl-eval-vfs-*")
		if err != nil {
			return
		}
		t.Cleanup(func() { os.RemoveAll(dir) })
		for i := 0; i < 50; i++ {
			f, err := os.CreateTemp(dir, "vfs-*")
			if err != nil {
				continue
			}
			f.Write([]byte("eval"))
			f.Close()
		}
	}, map[string]string{
		"vfs.vnstats.num_newvnode_calls": "likely",
	})
}

func TestCategory_Security_Provoked(t *testing.T) {
	testCategory(t, metrics.CatSecurity, "provoked", func() {
		// Execute processes to trigger ASP exec/library hooks.
		for i := 0; i < 20; i++ {
			exec.Command("true").Run()
		}
	}, map[string]string{
		"security.mac.asp.stats.exec_hook_count":    "likely",
		"security.mac.asp.stats.library_hook_count":  "likely",
		"security.mac.asp.stats.exec_hook_work_time": "likely",
		"security.mac.asp.stats.library_hook_time":   "likely",
		"security.mac.vnode_label_count":              "possible",
	})
}

func TestCategory_NetTCP_Provoked(t *testing.T) {
	testCategory(t, metrics.CatNetTCP, "provoked", func() {
		// Create TCP connections to change pcbcount.
		var listeners []net.Listener
		var conns []net.Conn
		for i := 0; i < 10; i++ {
			l, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				continue
			}
			listeners = append(listeners, l)
			c, err := net.Dial("tcp", l.Addr().String())
			if err != nil {
				continue
			}
			conns = append(conns, c)
			// Accept the server side to complete the handshake.
			go func(ln net.Listener) {
				sc, _ := ln.Accept()
				if sc != nil {
					conns = append(conns, sc)
				}
			}(l)
		}
		t.Cleanup(func() {
			for _, c := range conns {
				c.Close()
			}
			for _, l := range listeners {
				l.Close()
			}
		})
	}, map[string]string{
		"net.inet.tcp.pcbcount": "likely",
	})
}
