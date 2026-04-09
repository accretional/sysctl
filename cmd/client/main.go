//go:build darwin

package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/accretional/sysctl/proto/sysctlpb"
)

func main() {
	addr := flag.String("addr", "localhost:50051", "server address")
	listFlag := flag.Bool("list", false, "list known metrics")
	allFlag := flag.Bool("all", false, "fetch all known metrics")
	catFlag := flag.String("cat", "", "fetch/list metrics in a category")
	catsFlag := flag.Bool("cats", false, "list all categories")
	flag.Parse()

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewSysctlServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if *catsFlag {
		resp, err := client.ListCategories(ctx, &pb.ListCategoriesRequest{})
		if err != nil {
			log.Fatalf("ListCategories: %v", err)
		}
		for _, c := range resp.Categories {
			fmt.Printf("%-20s %d metrics\n", c.Name, c.MetricCount)
		}
		return
	}

	if *listFlag {
		resp, err := client.ListKnownMetrics(ctx, &pb.ListKnownMetricsRequest{Category: *catFlag})
		if err != nil {
			log.Fatalf("ListKnownMetrics: %v", err)
		}
		reg := resp.Registry
		for _, m := range reg.Metrics {
			recPattern := ""
			if m.RecommendedAccessPattern != nil {
				recPattern = m.RecommendedAccessPattern.Pattern.String()
				if m.RecommendedAccessPattern.Ttl != nil {
					recPattern += fmt.Sprintf("@%ds", m.RecommendedAccessPattern.Ttl.Seconds)
				}
			}
			fmt.Printf("%-50s %-8s %-12s %s\n", m.Name, m.ValueType, recPattern, m.Description)
		}
		fmt.Fprintf(os.Stderr, "\n%d metrics (%s %s)\n", len(reg.Metrics), reg.OsRegistry, reg.OsVersion)
		return
	}

	if *catFlag != "" {
		resp, err := client.GetMetricsByCategory(ctx, &pb.GetMetricsByCategoryRequest{Category: *catFlag})
		if err != nil {
			log.Fatalf("GetMetricsByCategory: %v", err)
		}
		for _, m := range resp.Metrics {
			printMetric(m)
		}
		return
	}

	if *allFlag {
		listResp, err := client.ListKnownMetrics(ctx, &pb.ListKnownMetricsRequest{})
		if err != nil {
			log.Fatalf("ListKnownMetrics: %v", err)
		}
		names := make([]string, len(listResp.Registry.Metrics))
		for i, m := range listResp.Registry.Metrics {
			names[i] = m.Name
		}
		resp, err := client.GetMetrics(ctx, &pb.GetMetricsRequest{Names: names})
		if err != nil {
			log.Fatalf("GetMetrics: %v", err)
		}
		lastCat := ""
		for _, m := range resp.Metrics {
			if m.Category != lastCat {
				if lastCat != "" {
					fmt.Println()
				}
				fmt.Printf("=== %s ===\n", m.Category)
				lastCat = m.Category
			}
			printMetric(m)
		}
		return
	}

	// Fetch specific metrics from args.
	names := flag.Args()
	if len(names) == 0 {
		fmt.Fprintln(os.Stderr, "usage: client [-list] [-all] [-cats] [-cat category] [-addr host:port] [metric-name ...]")
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
		fmt.Printf("%-50s ERROR: %s\n", m.Name, m.Error)
		return
	}
	switch v := m.Value.(type) {
	case *pb.Metric_StringValue:
		fmt.Printf("%-50s %s\n", m.Name, v.StringValue)
	case *pb.Metric_Uint64Value:
		fmt.Printf("%-50s %d\n", m.Name, v.Uint64Value)
	case *pb.Metric_Int64Value:
		fmt.Printf("%-50s %d\n", m.Name, v.Int64Value)
	case *pb.Metric_Uint32Value:
		fmt.Printf("%-50s %d\n", m.Name, v.Uint32Value)
	case *pb.Metric_Int32Value:
		fmt.Printf("%-50s %d\n", m.Name, v.Int32Value)
	case *pb.Metric_RawValue:
		fmt.Printf("%-50s [%d bytes] %s\n", m.Name, len(v.RawValue), hex.EncodeToString(v.RawValue))
	case *pb.Metric_StructValue:
		// Sort keys for stable output.
		keys := make([]string, 0, len(v.StructValue.Fields))
		for k := range v.StructValue.Fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s=%s", k, v.StructValue.Fields[k]))
		}
		fmt.Printf("%-50s {%s} %s\n", m.Name, v.StructValue.TypeName, strings.Join(parts, " "))
	default:
		fmt.Printf("%-50s <no value>\n", m.Name)
	}
}
