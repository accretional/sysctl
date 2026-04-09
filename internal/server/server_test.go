//go:build darwin

package server

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/accretional/sysctl/proto/sysctlpb"
)

const bufSize = 1024 * 1024

func startTestServer(t *testing.T) pb.SysctlServiceClient {
	t.Helper()
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterSysctlServiceServer(s, New())

	go func() {
		if err := s.Serve(lis); err != nil {
			// Server stopped.
		}
	}()
	t.Cleanup(func() { s.Stop() })

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial bufconn: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	return pb.NewSysctlServiceClient(conn)
}

func TestGetMetric_KernOstype(t *testing.T) {
	client := startTestServer(t)
	resp, err := client.GetMetric(context.Background(), &pb.GetMetricRequest{Name: "kern.ostype"})
	if err != nil {
		t.Fatalf("GetMetric: %v", err)
	}
	m := resp.Metric
	if m.Error != "" {
		t.Fatalf("metric error: %s", m.Error)
	}
	sv, ok := m.Value.(*pb.Metric_StringValue)
	if !ok {
		t.Fatalf("expected string value, got %T", m.Value)
	}
	if sv.StringValue != "Darwin" {
		t.Errorf("kern.ostype = %q, want Darwin", sv.StringValue)
	}
}

func TestGetMetric_HwMemsize(t *testing.T) {
	client := startTestServer(t)
	resp, err := client.GetMetric(context.Background(), &pb.GetMetricRequest{Name: "hw.memsize"})
	if err != nil {
		t.Fatalf("GetMetric: %v", err)
	}
	m := resp.Metric
	if m.Error != "" {
		t.Fatalf("metric error: %s", m.Error)
	}
	uv, ok := m.Value.(*pb.Metric_Uint64Value)
	if !ok {
		t.Fatalf("expected uint64 value, got %T", m.Value)
	}
	if uv.Uint64Value < 1<<30 {
		t.Errorf("hw.memsize = %d, expected >= 1 GB", uv.Uint64Value)
	}
}

func TestGetMetrics_Multiple(t *testing.T) {
	client := startTestServer(t)
	names := []string{"kern.ostype", "hw.ncpu", "hw.memsize"}
	resp, err := client.GetMetrics(context.Background(), &pb.GetMetricsRequest{Names: names})
	if err != nil {
		t.Fatalf("GetMetrics: %v", err)
	}
	if len(resp.Metrics) != 3 {
		t.Fatalf("got %d metrics, want 3", len(resp.Metrics))
	}
	for _, m := range resp.Metrics {
		if m.Error != "" {
			t.Errorf("metric %s error: %s", m.Name, m.Error)
		}
	}
}

func TestGetMetric_InvalidName(t *testing.T) {
	client := startTestServer(t)
	resp, err := client.GetMetric(context.Background(), &pb.GetMetricRequest{Name: "bogus.nonexistent"})
	if err != nil {
		t.Fatalf("GetMetric RPC error: %v", err)
	}
	if resp.Metric.Error == "" {
		t.Error("expected error for bogus metric name")
	}
}

func TestListKnownMetrics(t *testing.T) {
	client := startTestServer(t)
	resp, err := client.ListKnownMetrics(context.Background(), &pb.ListKnownMetricsRequest{})
	if err != nil {
		t.Fatalf("ListKnownMetrics: %v", err)
	}
	if len(resp.Metrics) == 0 {
		t.Fatal("no known metrics returned")
	}
	for _, m := range resp.Metrics {
		if m.Name == "" || m.ValueType == "" {
			t.Errorf("metric missing name or type: %v", m)
		}
	}
	t.Logf("listed %d known metrics", len(resp.Metrics))
}
