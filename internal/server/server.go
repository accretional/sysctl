//go:build darwin

// Package server implements the gRPC SysctlService.
package server

import (
	"context"
	"fmt"
	"time"

	"github.com/accretional/sysctl/internal/macosasmsysctl"
	"github.com/accretional/sysctl/internal/metrics"
	pb "github.com/accretional/sysctl/proto/sysctlpb"
)

// SysctlServer implements the SysctlService gRPC service.
type SysctlServer struct {
	pb.UnimplementedSysctlServiceServer
	cache          *macosasmsysctl.MIBCache
	kernelRegistry *pb.KernelMetricRegistry // raw from textproto
	fullRegistry   *pb.KernelMetricRegistry // merged with MetricInfo fields
}

// New returns a new SysctlServer with a warmed MIB cache.
// If osVersion is non-empty, the matching kernel registry is loaded.
func New(osVersion string) *SysctlServer {
	s := &SysctlServer{
		cache: macosasmsysctl.NewMIBCache(),
	}

	// Load kernel registry if available.
	if osVersion != "" {
		reg, err := metrics.LoadKernelRegistry(osVersion)
		if err == nil {
			s.kernelRegistry = reg
		}
	}

	// Build the full registry by merging Known (identity) with kernel registry (access patterns).
	s.fullRegistry = s.buildFullRegistry()

	// Pre-resolve MIBs for all known non-computed metrics.
	names := make([]string, 0, len(metrics.Known))
	for _, info := range metrics.Known {
		if info.Type != metrics.TypeComputed {
			names = append(names, info.Name)
		}
	}
	s.cache.Warm(names)

	return s
}

// buildFullRegistry merges the Known metric registry (descriptions, types, categories)
// with the kernel registry (access patterns) into a single KernelMetricRegistry.
func (s *SysctlServer) buildFullRegistry() *pb.KernelMetricRegistry {
	reg := &pb.KernelMetricRegistry{}
	if s.kernelRegistry != nil {
		reg.OsRegistry = s.kernelRegistry.OsRegistry
		reg.OsVersion = s.kernelRegistry.OsVersion
	}

	// Index kernel metrics by name for lookup.
	var kernelByName map[string]*pb.KernelMetric
	if s.kernelRegistry != nil {
		kernelByName = metrics.KernelRegistryByName(s.kernelRegistry)
	}

	// Default access config for metrics not in the kernel registry.
	defaultAccess := &pb.AccessConfig{Pattern: pb.AccessPattern_DYNAMIC}

	for _, info := range metrics.Known {
		km := &pb.KernelMetric{
			Name:        info.Name,
			Description: info.Description,
			ValueType:   string(info.Type),
			Category:    string(info.Category),
		}

		if km2, ok := kernelByName[info.Name]; ok {
			km.KernelAccessPattern = km2.KernelAccessPattern
			km.RecommendedAccessPattern = km2.RecommendedAccessPattern
			km.Notes = km2.Notes
		} else {
			km.KernelAccessPattern = defaultAccess
			km.RecommendedAccessPattern = defaultAccess
		}

		reg.Metrics = append(reg.Metrics, km)
	}

	return reg
}

func (s *SysctlServer) GetMetric(_ context.Context, req *pb.GetMetricRequest) (*pb.GetMetricResponse, error) {
	m := s.readMetric(req.Name)
	return &pb.GetMetricResponse{Metric: m}, nil
}

func (s *SysctlServer) GetMetrics(_ context.Context, req *pb.GetMetricsRequest) (*pb.GetMetricsResponse, error) {
	resp := &pb.GetMetricsResponse{
		Metrics: make([]*pb.Metric, len(req.Names)),
	}
	for i, name := range req.Names {
		resp.Metrics[i] = s.readMetric(name)
	}
	return resp, nil
}

func (s *SysctlServer) GetMetricsByCategory(_ context.Context, req *pb.GetMetricsByCategoryRequest) (*pb.GetMetricsResponse, error) {
	cat := metrics.Category(req.Category)
	infos := metrics.ByCategory(cat)
	resp := &pb.GetMetricsResponse{
		Metrics: make([]*pb.Metric, len(infos)),
	}
	for i, info := range infos {
		resp.Metrics[i] = s.readMetric(info.Name)
	}
	return resp, nil
}

func (s *SysctlServer) ListKnownMetrics(_ context.Context, req *pb.ListKnownMetricsRequest) (*pb.ListKnownMetricsResponse, error) {
	if req.Category == "" {
		return &pb.ListKnownMetricsResponse{Registry: s.fullRegistry}, nil
	}

	// Filter by category.
	filtered := &pb.KernelMetricRegistry{
		OsRegistry: s.fullRegistry.OsRegistry,
		OsVersion:  s.fullRegistry.OsVersion,
	}
	for _, km := range s.fullRegistry.Metrics {
		if km.Category == req.Category {
			filtered.Metrics = append(filtered.Metrics, km)
		}
	}
	return &pb.ListKnownMetricsResponse{Registry: filtered}, nil
}

func (s *SysctlServer) ListCategories(_ context.Context, _ *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	cats := metrics.Categories()
	resp := &pb.ListCategoriesResponse{}
	for _, cat := range cats {
		infos := metrics.ByCategory(cat)
		resp.Categories = append(resp.Categories, &pb.CategoryInfo{
			Name:        string(cat),
			MetricCount: int32(len(infos)),
		})
	}
	return resp, nil
}

func (s *SysctlServer) GetKernelRegistry(_ context.Context, _ *pb.GetKernelRegistryRequest) (*pb.GetKernelRegistryResponse, error) {
	return &pb.GetKernelRegistryResponse{Registry: s.fullRegistry}, nil
}

func (s *SysctlServer) readMetric(name string) *pb.Metric {
	info := metrics.ByName(name)

	m := &pb.Metric{Name: name}
	if info != nil {
		m.Category = string(info.Category)
	}

	if info == nil {
		// Unknown metric — try as raw bytes.
		raw, err := s.cache.GetRaw(name)
		if err != nil {
			m.Error = err.Error()
			return m
		}
		m.Value = &pb.Metric_RawValue{RawValue: raw}
		return m
	}

	if info.Type == metrics.TypeComputed {
		return s.readComputed(name, m)
	}

	switch info.Type {
	case metrics.TypeString:
		v, err := s.cache.GetString(name)
		if err != nil {
			m.Error = err.Error()
			return m
		}
		m.Value = &pb.Metric_StringValue{StringValue: v}

	case metrics.TypeUint64:
		v, err := s.cache.GetUint64(name)
		if err != nil {
			m.Error = err.Error()
			return m
		}
		m.Value = &pb.Metric_Uint64Value{Uint64Value: v}

	case metrics.TypeUint32:
		v, err := s.cache.GetUint32(name)
		if err != nil {
			m.Error = err.Error()
			return m
		}
		m.Value = &pb.Metric_Uint32Value{Uint32Value: v}

	case metrics.TypeInt32:
		v, err := s.cache.GetInt32(name)
		if err != nil {
			m.Error = err.Error()
			return m
		}
		m.Value = &pb.Metric_Int32Value{Int32Value: v}

	case metrics.TypeInt64:
		v, err := s.cache.GetInt64(name)
		if err != nil {
			m.Error = err.Error()
			return m
		}
		m.Value = &pb.Metric_Int64Value{Int64Value: v}

	case metrics.TypeTimeval:
		tv, err := s.cache.GetTimeval(name)
		if err != nil {
			m.Error = err.Error()
			return m
		}
		m.Value = &pb.Metric_StructValue{StructValue: &pb.StructValue{
			TypeName: "timeval",
			Fields: map[string]string{
				"sec":  fmt.Sprintf("%d", tv.Sec),
				"usec": fmt.Sprintf("%d", tv.Usec),
				"time": tv.Time().Format("2006-01-02T15:04:05Z07:00"),
			},
		}}

	case metrics.TypeLoadavg:
		la, err := s.cache.GetLoadavg()
		if err != nil {
			m.Error = err.Error()
			return m
		}
		m.Value = &pb.Metric_StructValue{StructValue: &pb.StructValue{
			TypeName: "loadavg",
			Fields: map[string]string{
				"load1":  fmt.Sprintf("%.2f", la.Load1),
				"load5":  fmt.Sprintf("%.2f", la.Load5),
				"load15": fmt.Sprintf("%.2f", la.Load15),
			},
		}}

	case metrics.TypeSwap:
		su, err := s.cache.GetSwapUsage()
		if err != nil {
			m.Error = err.Error()
			return m
		}
		m.Value = &pb.Metric_StructValue{StructValue: &pb.StructValue{
			TypeName: "swap",
			Fields: map[string]string{
				"total": fmt.Sprintf("%d", su.Total),
				"avail": fmt.Sprintf("%d", su.Avail),
				"used":  fmt.Sprintf("%d", su.Used),
			},
		}}

	case metrics.TypeClock:
		ci, err := s.cache.GetClockinfo()
		if err != nil {
			m.Error = err.Error()
			return m
		}
		m.Value = &pb.Metric_StructValue{StructValue: &pb.StructValue{
			TypeName: "clock",
			Fields: map[string]string{
				"hz":     fmt.Sprintf("%d", ci.Hz),
				"tick":   fmt.Sprintf("%d", ci.Tick),
				"profhz": fmt.Sprintf("%d", ci.Profhz),
				"stathz": fmt.Sprintf("%d", ci.Stathz),
			},
		}}

	case metrics.TypeRaw:
		raw, err := s.cache.GetRaw(name)
		if err != nil {
			m.Error = err.Error()
			return m
		}
		m.Value = &pb.Metric_RawValue{RawValue: raw}

	default:
		m.Error = fmt.Sprintf("unsupported type %q", info.Type)
	}
	return m
}

// readComputed handles server-side computed aggregate metrics.
func (s *SysctlServer) readComputed(name string, m *pb.Metric) *pb.Metric {
	switch name {
	case "computed.memory_utilization_pct":
		return s.computeMemoryUtil(m)
	case "computed.compression_ratio":
		return s.computeCompressionRatio(m)
	case "computed.swap_utilization_pct":
		return s.computeSwapUtil(m)
	case "computed.compressor_pressure_pct":
		return s.computeCompressorPressure(m)
	case "computed.total_connections":
		return s.computeTotalConnections(m)
	case "computed.uptime_seconds":
		return s.computeUptime(m)
	case "computed.vfs_reclamation_pct":
		return s.computeVFSReclamation(m)
	default:
		m.Error = fmt.Sprintf("unknown computed metric %q", name)
		return m
	}
}

// computeMemoryUtil: (total_pages - free_pages) / total_pages * 100
func (s *SysctlServer) computeMemoryUtil(m *pb.Metric) *pb.Metric {
	total, err := s.cache.GetInt32("vm.pages")
	if err != nil {
		m.Error = fmt.Sprintf("vm.pages: %v", err)
		return m
	}
	free, err := s.cache.GetInt32("vm.page_free_count")
	if err != nil {
		m.Error = fmt.Sprintf("vm.page_free_count: %v", err)
		return m
	}
	if total == 0 {
		m.Error = "vm.pages is 0"
		return m
	}
	pct := float64(total-free) / float64(total) * 100
	m.Value = &pb.Metric_StructValue{StructValue: &pb.StructValue{
		TypeName: "computed",
		Fields: map[string]string{
			"value":       fmt.Sprintf("%.2f", pct),
			"unit":        "percent",
			"total_pages": fmt.Sprintf("%d", total),
			"free_pages":  fmt.Sprintf("%d", free),
		},
	}}
	return m
}

// computeCompressionRatio: input_bytes / compressed_bytes
func (s *SysctlServer) computeCompressionRatio(m *pb.Metric) *pb.Metric {
	input, err := s.cache.GetUint64("vm.compressor_input_bytes")
	if err != nil {
		m.Error = fmt.Sprintf("vm.compressor_input_bytes: %v", err)
		return m
	}
	compressed, err := s.cache.GetUint64("vm.compressor_compressed_bytes")
	if err != nil {
		m.Error = fmt.Sprintf("vm.compressor_compressed_bytes: %v", err)
		return m
	}
	ratio := 0.0
	if compressed > 0 {
		ratio = float64(input) / float64(compressed)
	}
	m.Value = &pb.Metric_StructValue{StructValue: &pb.StructValue{
		TypeName: "computed",
		Fields: map[string]string{
			"value":            fmt.Sprintf("%.2f", ratio),
			"unit":             "ratio",
			"input_bytes":      fmt.Sprintf("%d", input),
			"compressed_bytes": fmt.Sprintf("%d", compressed),
		},
	}}
	return m
}

// computeSwapUtil: used / total * 100
func (s *SysctlServer) computeSwapUtil(m *pb.Metric) *pb.Metric {
	su, err := s.cache.GetSwapUsage()
	if err != nil {
		m.Error = fmt.Sprintf("vm.swapusage: %v", err)
		return m
	}
	pct := 0.0
	if su.Total > 0 {
		pct = float64(su.Used) / float64(su.Total) * 100
	}
	m.Value = &pb.Metric_StructValue{StructValue: &pb.StructValue{
		TypeName: "computed",
		Fields: map[string]string{
			"value":      fmt.Sprintf("%.2f", pct),
			"unit":       "percent",
			"swap_used":  fmt.Sprintf("%d", su.Used),
			"swap_total": fmt.Sprintf("%d", su.Total),
		},
	}}
	return m
}

// computeCompressorPressure: bytes_used / pool_size * 100
func (s *SysctlServer) computeCompressorPressure(m *pb.Metric) *pb.Metric {
	used, err := s.cache.GetUint64("vm.compressor_bytes_used")
	if err != nil {
		m.Error = fmt.Sprintf("vm.compressor_bytes_used: %v", err)
		return m
	}
	pool, err := s.cache.GetUint64("vm.compressor_pool_size")
	if err != nil {
		m.Error = fmt.Sprintf("vm.compressor_pool_size: %v", err)
		return m
	}
	pct := 0.0
	if pool > 0 {
		pct = float64(used) / float64(pool) * 100
	}
	m.Value = &pb.Metric_StructValue{StructValue: &pb.StructValue{
		TypeName: "computed",
		Fields: map[string]string{
			"value":      fmt.Sprintf("%.2f", pct),
			"unit":       "percent",
			"bytes_used": fmt.Sprintf("%d", used),
			"pool_size":  fmt.Sprintf("%d", pool),
		},
	}}
	return m
}

// computeTotalConnections: tcp + udp + unix
func (s *SysctlServer) computeTotalConnections(m *pb.Metric) *pb.Metric {
	tcp, err := s.cache.GetInt32("net.inet.tcp.pcbcount")
	if err != nil {
		m.Error = fmt.Sprintf("net.inet.tcp.pcbcount: %v", err)
		return m
	}
	udp, err := s.cache.GetInt32("net.inet.udp.pcbcount")
	if err != nil {
		m.Error = fmt.Sprintf("net.inet.udp.pcbcount: %v", err)
		return m
	}
	unix, err := s.cache.GetInt32("net.local.pcbcount")
	if err != nil {
		m.Error = fmt.Sprintf("net.local.pcbcount: %v", err)
		return m
	}
	total := int64(tcp) + int64(udp) + int64(unix)
	m.Value = &pb.Metric_StructValue{StructValue: &pb.StructValue{
		TypeName: "computed",
		Fields: map[string]string{
			"value": fmt.Sprintf("%d", total),
			"unit":  "connections",
			"tcp":   fmt.Sprintf("%d", tcp),
			"udp":   fmt.Sprintf("%d", udp),
			"unix":  fmt.Sprintf("%d", unix),
		},
	}}
	return m
}

// computeUptime: now - boottime
func (s *SysctlServer) computeUptime(m *pb.Metric) *pb.Metric {
	tv, err := s.cache.GetTimeval("kern.boottime")
	if err != nil {
		m.Error = fmt.Sprintf("kern.boottime: %v", err)
		return m
	}
	uptime := time.Since(tv.Time())
	secs := int64(uptime.Seconds())
	m.Value = &pb.Metric_StructValue{StructValue: &pb.StructValue{
		TypeName: "computed",
		Fields: map[string]string{
			"value":    fmt.Sprintf("%d", secs),
			"unit":     "seconds",
			"human":    uptime.Truncate(time.Second).String(),
			"boottime": tv.Time().Format("2006-01-02T15:04:05Z07:00"),
		},
	}}
	return m
}

// computeVFSReclamation: recycled / total
func (s *SysctlServer) computeVFSReclamation(m *pb.Metric) *pb.Metric {
	total, err := s.cache.GetInt64("vfs.vnstats.num_vnodes")
	if err != nil {
		m.Error = fmt.Sprintf("vfs.vnstats.num_vnodes: %v", err)
		return m
	}
	recycled, err := s.cache.GetInt64("vfs.vnstats.num_recycledvnodes")
	if err != nil {
		m.Error = fmt.Sprintf("vfs.vnstats.num_recycledvnodes: %v", err)
		return m
	}
	pct := 0.0
	if total > 0 {
		pct = float64(recycled) / float64(total) * 100
	}
	m.Value = &pb.Metric_StructValue{StructValue: &pb.StructValue{
		TypeName: "computed",
		Fields: map[string]string{
			"value":    fmt.Sprintf("%.2f", pct),
			"unit":     "percent",
			"recycled": fmt.Sprintf("%d", recycled),
			"total":    fmt.Sprintf("%d", total),
		},
	}}
	return m
}
