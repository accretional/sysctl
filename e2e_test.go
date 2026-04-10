//go:build darwin

package sysctl_test

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/accretional/sysctl/proto/sysctlpb"
	"github.com/accretional/sysctl/internal/server"
)

const bufSize = 1024 * 1024

// TestE2E_FullFlow tests the complete path: client -> gRPC -> server -> asm sysctl -> back.
func TestE2E_FullFlow(t *testing.T) {
	// Start server on bufconn.
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterSysctlServiceServer(s, server.New("24.6.0", 0))
	go func() { s.Serve(lis) }()
	t.Cleanup(func() { s.Stop() })

	// Connect client.
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewSysctlServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. List known metrics.
	listResp, err := client.ListKnownMetrics(ctx, &pb.ListKnownMetricsRequest{})
	if err != nil {
		t.Fatalf("ListKnownMetrics: %v", err)
	}
	reg := listResp.Registry
	if reg == nil || len(reg.Metrics) == 0 {
		t.Fatal("no known metrics")
	}
	t.Logf("server reports %d known metrics (%s %s)", len(reg.Metrics), reg.OsRegistry, reg.OsVersion)

	// 2. Fetch a single metric.
	getResp, err := client.GetMetric(ctx, &pb.GetMetricRequest{Name: "kern.ostype"})
	if err != nil {
		t.Fatalf("GetMetric(kern.ostype): %v", err)
	}
	if getResp.Metric.Error != "" {
		t.Fatalf("kern.ostype error: %s", getResp.Metric.Error)
	}
	sv, ok := getResp.Metric.Value.(*pb.Metric_StringValue)
	if !ok || sv.StringValue != "Darwin" {
		t.Fatalf("kern.ostype = %v, want Darwin", getResp.Metric.Value)
	}
	t.Log("kern.ostype = Darwin ✓")

	// 3. Fetch multiple metrics.
	names := make([]string, len(reg.Metrics))
	for i, m := range reg.Metrics {
		names[i] = m.Name
	}
	multiResp, err := client.GetMetrics(ctx, &pb.GetMetricsRequest{Names: names})
	if err != nil {
		t.Fatalf("GetMetrics: %v", err)
	}
	successCount := 0
	for _, m := range multiResp.Metrics {
		if m.Error == "" {
			successCount++
		} else {
			t.Logf("  %s: %s (may not be available on this hardware)", m.Name, m.Error)
		}
	}
	t.Logf("fetched %d/%d metrics successfully", successCount, len(names))

	if successCount == 0 {
		t.Fatal("no metrics could be read — something is fundamentally wrong")
	}
	// At minimum, the core metrics should work.
	if successCount < 5 {
		t.Errorf("only %d metrics succeeded, expected at least 5 core metrics", successCount)
	}
}
