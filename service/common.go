package service

import (
	"context"
	"fmt"

	"github.com/nick0323/K8sVision/model"
	v1 "k8s.io/api/core/v1"
)

const (
	BytesPerKiB = 1024
	BytesPerMiB = 1024 * 1024
	BytesPerGiB = 1024 * 1024 * 1024
)

func ParseCPU(cpuStr string) float64 {
	if cpuStr == "" {
		return 0
	}
	if n, ok := parseSuffix(cpuStr, "n"); ok {
		return n / 1e9
	}
	if n, ok := parseSuffix(cpuStr, "m"); ok {
		return n / 1000
	}
	n, _ := parseSuffix(cpuStr, "")
	return n
}

func ParseMemory(memStr string) float64 {
	if memStr == "" {
		return 0
	}
	if n, ok := parseSuffix(memStr, "Ki"); ok {
		return n * BytesPerKiB / BytesPerGiB
	}
	if n, ok := parseSuffix(memStr, "Mi"); ok {
		return n * BytesPerMiB / BytesPerGiB
	}
	if n, ok := parseSuffix(memStr, "Gi"); ok {
		return n
	}
	n, _ := parseSuffix(memStr, "")
	return n / BytesPerGiB
}

func parseSuffix(s string, suffix string) (float64, bool) {
	if suffix != "" && !hasSuffix(s, suffix) {
		return 0, false
	}
	if suffix != "" {
		s = s[:len(s)-len(suffix)]
	}
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err == nil
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func GetNodeAllocatableCPU(node v1.Node) float64 {
	if v, ok := node.Status.Allocatable[v1.ResourceName(model.ResourceCPU)]; ok {
		return float64(v.MilliValue()) / 1000
	}
	return 0
}

func GetNodeAllocatableMemory(node v1.Node) float64 {
	if v, ok := node.Status.Allocatable[v1.ResourceName(model.ResourceMemory)]; ok {
		return float64(v.Value()) / BytesPerGiB
	}
	return 0
}

func GetWorkloadStatus(ready, desired int32) string {
	if ready == desired && desired > 0 {
		return model.WorkloadAvailable
	} else if ready > 0 {
		return model.WorkloadPartial
	}
	return model.WorkloadUnavailable
}

func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func ListResourcesWithNamespace[T any](
	ctx context.Context,
	namespace string,
	listAll func() (T, error),
	listNS func(string) (T, error),
) (T, error) {
	if namespace == "" {
		return listAll()
	}
	return listNS(namespace)
}
