//go:build darwin

package server

import (
	"context"
	"net"
	"testing"
	"time"

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
	pb.RegisterSysctlServiceServer(s, New("24.6.0", 0))

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

func TestGetMetrics_KernOstype(t *testing.T) {
	client := startTestServer(t)
	resp, err := client.GetMetrics(context.Background(), &pb.GetMetricsRequest{Names: []string{"kern.ostype"}})
	if err != nil {
		t.Fatalf("GetMetrics: %v", err)
	}
	m := resp.Metrics[0]
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

func TestGetMetrics_HwMemsize(t *testing.T) {
	client := startTestServer(t)
	resp, err := client.GetMetrics(context.Background(), &pb.GetMetricsRequest{Names: []string{"hw.memsize"}})
	if err != nil {
		t.Fatalf("GetMetrics: %v", err)
	}
	m := resp.Metrics[0]
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

func TestGetMetrics_InvalidName(t *testing.T) {
	client := startTestServer(t)
	resp, err := client.GetMetrics(context.Background(), &pb.GetMetricsRequest{Names: []string{"bogus.nonexistent"}})
	if err != nil {
		t.Fatalf("GetMetrics RPC error: %v", err)
	}
	if resp.Metrics[0].Error == "" {
		t.Error("expected error for bogus metric name")
	}
}

func TestListKnownMetrics(t *testing.T) {
	client := startTestServer(t)
	resp, err := client.ListKnownMetrics(context.Background(), &pb.ListKnownMetricsRequest{})
	if err != nil {
		t.Fatalf("ListKnownMetrics: %v", err)
	}
	reg := resp.Registry
	if reg == nil || len(reg.Metrics) == 0 {
		t.Fatal("no known metrics returned")
	}
	for _, m := range reg.Metrics {
		if m.Name == "" || m.ValueType == "" {
			t.Errorf("metric missing name or type: %v", m)
		}
		if m.KernelAccessPattern == nil {
			t.Errorf("metric %s missing kernel_access_pattern", m.Name)
		}
		if m.RecommendedAccessPattern == nil {
			t.Errorf("metric %s missing recommended_access_pattern", m.Name)
		}
	}
	t.Logf("listed %d known metrics (%s %s)", len(reg.Metrics), reg.OsRegistry, reg.OsVersion)
}

func TestListKnownMetrics_CategoryFilter(t *testing.T) {
	client := startTestServer(t)
	resp, err := client.ListKnownMetrics(context.Background(), &pb.ListKnownMetricsRequest{Category: "hw.cpu"})
	if err != nil {
		t.Fatalf("ListKnownMetrics(hw.cpu): %v", err)
	}
	reg := resp.Registry
	if reg == nil || len(reg.Metrics) == 0 {
		t.Fatal("no hw.cpu metrics returned")
	}
	for _, m := range reg.Metrics {
		if m.Category != "hw.cpu" {
			t.Errorf("got category %q, want hw.cpu", m.Category)
		}
	}
	t.Logf("hw.cpu has %d metrics", len(reg.Metrics))
}

func TestSubscribe(t *testing.T) {
	client := startTestServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.Subscribe(ctx, &pb.SubscribeRequest{
		Names:      []string{"kern.ostype", "hw.memsize"},
		IntervalNs: int64(100 * time.Millisecond),
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	// First response: full snapshot with timestamp + errors.
	resp, err := stream.Recv()
	if err != nil {
		t.Fatalf("Recv 0: %v", err)
	}
	if resp.TimestampNs == 0 {
		t.Error("first response missing timestamp_ns")
	}
	if len(resp.Deltas) != 2 {
		t.Fatalf("first response: got %d deltas, want 2", len(resp.Deltas))
	}
	for _, d := range resp.Deltas {
		if len(d.Value) == 0 {
			t.Errorf("delta %s has empty value", d.Name)
		}
		t.Logf("delta: %s = %d bytes", d.Name, len(d.Value))
	}

	// Second response should have a later timestamp.
	resp2, err := stream.Recv()
	if err != nil {
		t.Fatalf("Recv 1: %v", err)
	}
	if resp2.TimestampNs <= resp.TimestampNs {
		t.Errorf("timestamp did not advance: %d -> %d", resp.TimestampNs, resp2.TimestampNs)
	}
	t.Logf("second response: %d deltas, timestamp advanced by %dns",
		len(resp2.Deltas), resp2.TimestampNs-resp.TimestampNs)
	cancel()
}

func TestSubscribe_DeltaOnly(t *testing.T) {
	client := startTestServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Subscribe to two STATIC metrics — after the first full snapshot, nothing changes.
	stream, err := client.Subscribe(ctx, &pb.SubscribeRequest{
		Names:      []string{"hw.memsize", "kern.ostype"},
		IntervalNs: int64(100 * time.Millisecond),
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	// First: full snapshot with both metrics.
	resp, err := stream.Recv()
	if err != nil {
		t.Fatalf("Recv 0: %v", err)
	}
	if len(resp.Deltas) != 2 {
		t.Fatalf("first: got %d deltas, want 2", len(resp.Deltas))
	}

	// Subsequent responses should have zero deltas (STATIC values don't change).
	for i := 0; i < 3; i++ {
		resp, err = stream.Recv()
		if err != nil {
			t.Fatalf("Recv %d: %v", i+1, err)
		}
		if len(resp.Deltas) > 0 {
			for _, d := range resp.Deltas {
				t.Errorf("unexpected delta for STATIC metric %s after first snapshot", d.Name)
			}
		}
	}
	t.Log("STATIC metrics correctly excluded from subsequent deltas")
	cancel()
}

func TestSubscribe_Errors(t *testing.T) {
	client := startTestServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Include a computed metric and an unknown metric — errors on first response only.
	stream, err := client.Subscribe(ctx, &pb.SubscribeRequest{
		Names:      []string{"hw.memsize", "computed.uptime_seconds", "bogus.nonexistent"},
		IntervalNs: int64(100 * time.Millisecond),
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	resp, err := stream.Recv()
	if err != nil {
		t.Fatalf("Recv: %v", err)
	}
	if len(resp.Errors) != 2 {
		t.Fatalf("got %d errors, want 2 (computed + unknown)", len(resp.Errors))
	}
	for _, e := range resp.Errors {
		t.Logf("error: %s", e)
	}
	// Only hw.memsize should be in deltas.
	if len(resp.Deltas) != 1 || resp.Deltas[0].Name != "hw.memsize" {
		t.Errorf("expected 1 delta for hw.memsize, got %d", len(resp.Deltas))
	}

	// Second response should have no errors.
	resp2, err := stream.Recv()
	if err != nil {
		t.Fatalf("Recv 2: %v", err)
	}
	if len(resp2.Errors) != 0 {
		t.Errorf("second response should have no errors, got %d", len(resp2.Errors))
	}
	cancel()
}

func TestSubscribe_RejectTooFast(t *testing.T) {
	client := startTestServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Request 1ns interval — below server minimum, should be rejected.
	stream, err := client.Subscribe(ctx, &pb.SubscribeRequest{
		Names:      []string{"hw.memsize"},
		IntervalNs: 1,
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	_, err = stream.Recv()
	if err == nil {
		t.Fatal("expected error for too-fast interval, got nil")
	}
	t.Logf("correctly rejected: %v", err)
}

func TestGetKernelRegistry_MinInterval(t *testing.T) {
	client := startTestServer(t)
	resp, err := client.GetKernelRegistry(context.Background(), &pb.GetKernelRegistryRequest{})
	if err != nil {
		t.Fatalf("GetKernelRegistry: %v", err)
	}
	if resp.MinIntervalNs <= 0 {
		t.Fatalf("min_interval_ns = %d, want > 0", resp.MinIntervalNs)
	}
	t.Logf("min_interval_ns = %d (%v)", resp.MinIntervalNs, time.Duration(resp.MinIntervalNs))
}

func TestGetKernelRegistry(t *testing.T) {
	client := startTestServer(t)
	resp, err := client.GetKernelRegistry(context.Background(), &pb.GetKernelRegistryRequest{})
	if err != nil {
		t.Fatalf("GetKernelRegistry: %v", err)
	}
	reg := resp.Registry
	if reg == nil || len(reg.Metrics) == 0 {
		t.Fatal("no metrics in kernel registry")
	}
	if reg.OsRegistry != "darwin-arm64" {
		t.Errorf("os_registry = %q, want darwin-arm64", reg.OsRegistry)
	}
	t.Logf("kernel registry: %d metrics (%s %s)", len(reg.Metrics), reg.OsRegistry, reg.OsVersion)
}
