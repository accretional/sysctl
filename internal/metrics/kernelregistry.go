//go:build darwin

package metrics

import (
	"embed"
	"fmt"
	"runtime"

	pb "github.com/accretional/sysctl/proto/sysctlpb"
	"google.golang.org/protobuf/encoding/prototext"
)

//go:embed darwin/*.textproto
var darwinConfigs embed.FS

// LoadKernelRegistry loads the KernelMetricRegistry for the current platform
// and OS version. It returns an error if no matching textproto is found.
func LoadKernelRegistry(osVersion string) (*pb.KernelMetricRegistry, error) {
	if runtime.GOARCH != "arm64" || runtime.GOOS != "darwin" {
		return nil, fmt.Errorf("no kernel registry for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	path := fmt.Sprintf("darwin/%s.textproto", osVersion)
	data, err := darwinConfigs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load kernel registry %q: %w", path, err)
	}

	reg := &pb.KernelMetricRegistry{}
	if err := prototext.Unmarshal(data, reg); err != nil {
		return nil, fmt.Errorf("parse kernel registry %q: %w", path, err)
	}

	return reg, nil
}

// ValidateRegistry checks that the KernelMetricRegistry is consistent with
// the Known metrics registry. It returns a list of issues found.
func ValidateRegistry(reg *pb.KernelMetricRegistry) []string {
	var issues []string

	// Build lookup from kernel registry.
	kernelMetrics := make(map[string]*pb.KernelMetric, len(reg.Metrics))
	for _, km := range reg.Metrics {
		if _, dup := kernelMetrics[km.Name]; dup {
			issues = append(issues, fmt.Sprintf("duplicate kernel metric: %s", km.Name))
		}
		kernelMetrics[km.Name] = km
	}

	// Check every Known metric has a kernel registry entry.
	for _, info := range Known {
		km, ok := kernelMetrics[info.Name]
		if !ok {
			issues = append(issues, fmt.Sprintf("metric %q in registry but missing from kernel registry", info.Name))
			continue
		}

		// Validate access configs are present.
		if km.KernelAccessPattern == nil {
			issues = append(issues, fmt.Sprintf("metric %q: missing kernel_access_pattern", info.Name))
		}
		if km.RecommendedAccessPattern == nil {
			issues = append(issues, fmt.Sprintf("metric %q: missing recommended_access_pattern", info.Name))
		}

		// Validate POLLED/CACHED have TTLs.
		if km.RecommendedAccessPattern != nil {
			p := km.RecommendedAccessPattern.Pattern
			if (p == pb.AccessPattern_POLLED || p == pb.AccessPattern_CACHED) && km.RecommendedAccessPattern.Ttl == nil {
				issues = append(issues, fmt.Sprintf("metric %q: %s recommended but no TTL set", info.Name, p))
			}
		}

		delete(kernelMetrics, info.Name)
	}

	// Check for kernel registry entries with no Known metric.
	for name := range kernelMetrics {
		issues = append(issues, fmt.Sprintf("kernel metric %q not in Known registry", name))
	}

	return issues
}

// KernelRegistryByName builds a lookup map from metric name to KernelMetric.
func KernelRegistryByName(reg *pb.KernelMetricRegistry) map[string]*pb.KernelMetric {
	m := make(map[string]*pb.KernelMetric, len(reg.Metrics))
	for _, km := range reg.Metrics {
		m[km.Name] = km
	}
	return m
}
