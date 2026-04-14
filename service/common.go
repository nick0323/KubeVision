package service

import (
	"context"
	"strconv"
	"strings"

	"github.com/nick0323/K8sVision/model"
	v1 "k8s.io/api/core/v1"
)

// 资源格式化常量
const (
	BytesPerKiB = 1024
	BytesPerMiB = 1024 * 1024
	BytesPerGiB = 1024 * 1024 * 1024
)

// ParseCPU 解析 CPU 字符串
func ParseCPU(cpuStr string) float64 {
	if cpuStr == "" {
		return 0
	}
	if strings.HasSuffix(cpuStr, "n") {
		n, _ := strconv.ParseFloat(strings.TrimSuffix(cpuStr, "n"), 64)
		return n / 1e9
	}
	if strings.HasSuffix(cpuStr, "m") {
		n, _ := strconv.ParseFloat(strings.TrimSuffix(cpuStr, "m"), 64)
		return n / 1000
	}
	n, _ := strconv.ParseFloat(cpuStr, 64)
	return n
}

// ParseMemory 解析内存字符串
func ParseMemory(memStr string) float64 {
	if memStr == "" {
		return 0
	}
	if strings.HasSuffix(memStr, "Ki") {
		n, _ := strconv.ParseFloat(strings.TrimSuffix(memStr, "Ki"), 64)
		return n * BytesPerKiB / BytesPerGiB
	}
	if strings.HasSuffix(memStr, "Mi") {
		n, _ := strconv.ParseFloat(strings.TrimSuffix(memStr, "Mi"), 64)
		return n * BytesPerMiB / BytesPerGiB
	}
	if strings.HasSuffix(memStr, "Gi") {
		n, _ := strconv.ParseFloat(strings.TrimSuffix(memStr, "Gi"), 64)
		return n
	}
	n, _ := strconv.ParseFloat(memStr, 64)
	return n / BytesPerGiB
}

// GetNodeAllocatableCPU 获取节点可分配 CPU
func GetNodeAllocatableCPU(node v1.Node) float64 {
	if v, ok := node.Status.Allocatable[v1.ResourceName(model.ResourceCPU)]; ok {
		return float64(v.MilliValue()) / 1000
	}
	return 0
}

// GetNodeAllocatableMemory 获取节点可分配内存
func GetNodeAllocatableMemory(node v1.Node) float64 {
	if v, ok := node.Status.Allocatable[v1.ResourceName(model.ResourceMemory)]; ok {
		return float64(v.Value()) / BytesPerGiB
	}
	return 0
}

// ListResourcesWithNamespace 通用资源列表函数（支持命名空间）
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

func GetWorkloadStatus(ready, desired int32) string {
	if ready == desired && desired > 0 {
		return model.WorkloadAvailable
	} else if ready > 0 {
		return model.WorkloadPartial
	}
	return model.WorkloadUnavailable
}

func GetJobStatus(succeeded, failed, active int32) string {
	if succeeded > 0 {
		return model.PodSucceeded
	} else if failed > 0 {
		return model.PodFailed
	} else if active > 0 {
		return model.PodRunning
	}
	return model.PodPending
}

func GetCronJobStatus(activeCount int, lastSuccessfulTime interface{}) string {
	if activeCount > 0 {
		return model.PodRunning
	} else if lastSuccessfulTime != nil {
		return model.PodSucceeded
	}
	return model.PodPending
}

func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
