//go:build darwin

package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/accretional/sysctl/proto/sysctlpb"
)

func main() {
	addr := flag.String("addr", "localhost:50051", "server address")
	listFlag := flag.Bool("list", false, "list known metrics")
	allFlag := flag.Bool("all", false, "fetch all known metrics")
	flag.Parse()

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewSysctlServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if *listFlag {
		resp, err := client.ListKnownMetrics(ctx, &pb.ListKnownMetricsRequest{})
		if err != nil {
			log.Fatalf("ListKnownMetrics: %v", err)
		}
		for _, m := range resp.Metrics {
			fmt.Printf("%-40s %-8s %s\n", m.Name, m.ValueType, m.Description)
		}
		return
	}

	if *allFlag {
		// Get the list first, then fetch all.
		listResp, err := client.ListKnownMetrics(ctx, &pb.ListKnownMetricsRequest{})
		if err != nil {
			log.Fatalf("ListKnownMetrics: %v", err)
		}
		names := make([]string, len(listResp.Metrics))
		for i, m := range listResp.Metrics {
			names[i] = m.Name
		}
		resp, err := client.GetMetrics(ctx, &pb.GetMetricsRequest{Names: names})
		if err != nil {
			log.Fatalf("GetMetrics: %v", err)
		}
		for _, m := range resp.Metrics {
			printMetric(m)
		}
		return
	}

	// Fetch specific metrics from args.
	names := flag.Args()
	if len(names) == 0 {
		fmt.Fprintln(os.Stderr, "usage: client [-list] [-all] [-addr host:port] [metric-name ...]")
		os.Exit(1)
	}

	if len(names) == 1 {
		resp, err := client.GetMetric(ctx, &pb.GetMetricRequest{Name: names[0]})
		if err != nil {
			log.Fatalf("GetMetric: %v", err)
		}
		printMetric(resp.Metric)
	} else {
		resp, err := client.GetMetrics(ctx, &pb.GetMetricsRequest{Names: names})
		if err != nil {
			log.Fatalf("GetMetrics: %v", err)
		}
		for _, m := range resp.Metrics {
			printMetric(m)
		}
	}
}

func printMetric(m *pb.Metric) {
	if m.Error != "" {
		fmt.Printf("%-40s ERROR: %s\n", m.Name, m.Error)
		return
	}
	switch v := m.Value.(type) {
	case *pb.Metric_StringValue:
		fmt.Printf("%-40s %s\n", m.Name, v.StringValue)
	case *pb.Metric_Uint64Value:
		fmt.Printf("%-40s %d\n", m.Name, v.Uint64Value)
	case *pb.Metric_Int64Value:
		fmt.Printf("%-40s %d\n", m.Name, v.Int64Value)
	case *pb.Metric_Uint32Value:
		fmt.Printf("%-40s %d\n", m.Name, v.Uint32Value)
	case *pb.Metric_Int32Value:
		fmt.Printf("%-40s %d\n", m.Name, v.Int32Value)
	case *pb.Metric_RawValue:
		fmt.Printf("%-40s [%d bytes] %s\n", m.Name, len(v.RawValue), hex.EncodeToString(v.RawValue))
	default:
		fmt.Printf("%-40s <no value>\n", m.Name)
	}
}
