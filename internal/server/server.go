//go:build darwin

// Package server implements the gRPC SysctlService.
package server

import (
	"context"
	"fmt"

	"github.com/accretional/sysctl/internal/macosasmsysctl"
	"github.com/accretional/sysctl/internal/metrics"
	pb "github.com/accretional/sysctl/proto/sysctlpb"
)

// SysctlServer implements the SysctlService gRPC service.
type SysctlServer struct {
	pb.UnimplementedSysctlServiceServer
}

// New returns a new SysctlServer.
func New() *SysctlServer {
	return &SysctlServer{}
}

func (s *SysctlServer) GetMetric(_ context.Context, req *pb.GetMetricRequest) (*pb.GetMetricResponse, error) {
	m := readMetric(req.Name)
	return &pb.GetMetricResponse{Metric: m}, nil
}

func (s *SysctlServer) GetMetrics(_ context.Context, req *pb.GetMetricsRequest) (*pb.GetMetricsResponse, error) {
	resp := &pb.GetMetricsResponse{
		Metrics: make([]*pb.Metric, len(req.Names)),
	}
	for i, name := range req.Names {
		resp.Metrics[i] = readMetric(name)
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
		resp.Metrics[i] = readMetric(info.Name)
	}
	return resp, nil
}

func (s *SysctlServer) ListKnownMetrics(_ context.Context, req *pb.ListKnownMetricsRequest) (*pb.ListKnownMetricsResponse, error) {
	resp := &pb.ListKnownMetricsResponse{}
	for _, info := range metrics.Known {
		if req.Category != "" && string(info.Category) != req.Category {
			continue
		}
		resp.Metrics = append(resp.Metrics, &pb.MetricInfo{
			Name:        info.Name,
			Description: info.Description,
			ValueType:   string(info.Type),
			Category:    string(info.Category),
		})
	}
	return resp, nil
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

func readMetric(name string) *pb.Metric {
	info := metrics.ByName(name)

	m := &pb.Metric{Name: name}
	if info != nil {
		m.Category = string(info.Category)
	}

	if info == nil {
		// Unknown metric — try as raw bytes.
		raw, err := macosasmsysctl.GetRaw(name)
		if err != nil {
			m.Error = err.Error()
			return m
		}
		m.Value = &pb.Metric_RawValue{RawValue: raw}
		return m
	}

	switch info.Type {
	case metrics.TypeString:
		v, err := macosasmsysctl.GetString(name)
		if err != nil {
			m.Error = err.Error()
			return m
		}
		m.Value = &pb.Metric_StringValue{StringValue: v}

	case metrics.TypeUint64:
		v, err := macosasmsysctl.GetUint64(name)
		if err != nil {
			m.Error = err.Error()
			return m
		}
		m.Value = &pb.Metric_Uint64Value{Uint64Value: v}

	case metrics.TypeUint32:
		v, err := macosasmsysctl.GetUint32(name)
		if err != nil {
			m.Error = err.Error()
			return m
		}
		m.Value = &pb.Metric_Uint32Value{Uint32Value: v}

	case metrics.TypeInt32:
		v, err := macosasmsysctl.GetInt32(name)
		if err != nil {
			m.Error = err.Error()
			return m
		}
		m.Value = &pb.Metric_Int32Value{Int32Value: v}

	case metrics.TypeInt64:
		v, err := macosasmsysctl.GetInt64(name)
		if err != nil {
			m.Error = err.Error()
			return m
		}
		m.Value = &pb.Metric_Int64Value{Int64Value: v}

	case metrics.TypeTimeval:
		tv, err := macosasmsysctl.GetTimeval(name)
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
		la, err := macosasmsysctl.GetLoadavg()
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
		su, err := macosasmsysctl.GetSwapUsage()
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
		ci, err := macosasmsysctl.GetClockinfo()
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
		raw, err := macosasmsysctl.GetRaw(name)
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
