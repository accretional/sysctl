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

func TestPoller_ServesFromCache(t *testing.T) {
	// Start server with polling enabled (100ms tick for fast test).
	srv := New("24.6.0", 100*time.Millisecond)
	t.Cleanup(func() { srv.Stop() })

	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	pb.RegisterSysctlServiceServer(s, srv)
	go s.Serve(lis)
	t.Cleanup(func() { s.Stop() })

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
	ctx := context.Background()

	// Wait for at least one poll tick.
	time.Sleep(200 * time.Millisecond)

	// STATIC metric should be in cache.
	resp, err := client.GetMetrics(ctx, &pb.GetMetricsRequest{Names: []string{"hw.memsize"}})
	if err != nil {
		t.Fatalf("GetMetrics(hw.memsize): %v", err)
	}
	if resp.Metrics[0].Error != "" {
		t.Fatalf("hw.memsize error: %s", resp.Metrics[0].Error)
	}
	uv, ok := resp.Metrics[0].Value.(*pb.Metric_Uint64Value)
	if !ok {
		t.Fatalf("expected uint64, got %T", resp.Metrics[0].Value)
	}
	if uv.Uint64Value < 1<<30 {
		t.Errorf("hw.memsize = %d, want >= 1 GB", uv.Uint64Value)
	}

	// POLLED metric should be in cache.
	resp, err = client.GetMetrics(ctx, &pb.GetMetricsRequest{Names: []string{"vm.page_free_count"}})
	if err != nil {
		t.Fatalf("GetMetrics(vm.page_free_count): %v", err)
	}
	if resp.Metrics[0].Error != "" {
		t.Fatalf("vm.page_free_count error: %s", resp.Metrics[0].Error)
	}
	_, ok = resp.Metrics[0].Value.(*pb.Metric_Int32Value)
	if !ok {
		t.Fatalf("expected int32, got %T", resp.Metrics[0].Value)
	}
}

func TestPoller_PolledMetricsRefresh(t *testing.T) {
	srv := New("24.6.0", 50*time.Millisecond)
	t.Cleanup(func() { srv.Stop() })

	// Wait for initial gather + a couple of ticks.
	time.Sleep(200 * time.Millisecond)

	// Verify a polled metric is in the store.
	m, ok := srv.poller.store.get("vm.page_free_count")
	if !ok {
		t.Fatal("vm.page_free_count not in poller store after ticks")
	}
	if m.Error != "" {
		t.Fatalf("vm.page_free_count error: %s", m.Error)
	}
	t.Logf("vm.page_free_count from store: %v", m.Value)

	// Verify a STATIC metric is also in the store.
	m, ok = srv.poller.store.get("hw.memsize")
	if !ok {
		t.Fatal("hw.memsize not in poller store")
	}
	if m.Error != "" {
		t.Fatalf("hw.memsize error: %s", m.Error)
	}
}

func TestPoller_MetricCounts(t *testing.T) {
	srv := New("24.6.0", 100*time.Millisecond)
	t.Cleanup(func() { srv.Stop() })

	if srv.poller == nil {
		t.Fatal("poller is nil")
	}

	staticCount := 0
	constrainedCount := 0
	for _, km := range srv.fullRegistry.Metrics {
		if km.RecommendedAccessPattern == nil {
			continue
		}
		switch km.RecommendedAccessPattern.Pattern {
		case pb.AccessPattern_STATIC:
			staticCount++
		case pb.AccessPattern_CONSTRAINED:
			constrainedCount++
		}
	}

	polledCount := len(srv.poller.polledMetrics) // includes both POLLED and CONSTRAINED
	t.Logf("poller: %d static (pre-loaded), %d polled+constrained (scheduled, %d constrained)", staticCount, polledCount, constrainedCount)

	if staticCount == 0 {
		t.Error("expected some static metrics")
	}
	if polledCount == 0 {
		t.Error("expected some polled/constrained metrics")
	}
	if constrainedCount == 0 {
		t.Error("expected some constrained metrics")
	}
	if staticCount+polledCount != len(srv.fullRegistry.Metrics) {
		t.Logf("note: %d metrics are neither STATIC nor POLLED/CONSTRAINED (DYNAMIC passthrough)",
			len(srv.fullRegistry.Metrics)-staticCount-polledCount)
	}
}
