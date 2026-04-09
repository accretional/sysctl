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

func (s *SysctlServer) ListKnownMetrics(_ context.Context, _ *pb.ListKnownMetricsRequest) (*pb.ListKnownMetricsResponse, error) {
	resp := &pb.ListKnownMetricsResponse{}
	for _, info := range metrics.Known {
		resp.Metrics = append(resp.Metrics, &pb.MetricInfo{
			Name:        info.Name,
			Description: info.Description,
			ValueType:   string(info.Type),
		})
	}
	return resp, nil
}

func readMetric(name string) *pb.Metric {
	m := &pb.Metric{Name: name}

	info := metrics.ByName(name)
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
