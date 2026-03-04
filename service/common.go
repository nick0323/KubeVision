package service

import (
	"context"
	"fmt"

	"github.com/nick0323/K8sVision/model"
	"k8s.io/client-go/kubernetes"
)

type K8sResourceLister[T any] interface {
	List(ctx context.Context, namespace string) (T, error)
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

// ResourceLister 通用资源列表器
type ResourceLister[T any] struct {
	Clientset *kubernetes.Clientset
}

// NewResourceLister 创建新的通用资源列表器
func NewResourceLister[T any](clientset *kubernetes.Clientset) *ResourceLister[T] {
	return &ResourceLister[T]{Clientset: clientset}
}

func FormatResourceUsage(cpu, mem int64) (cpuStr, memStr string) {
	if cpu > 0 {
		cpuStr = fmt.Sprintf("%.2f mCPU", float64(cpu))
	} else {
		cpuStr = "-"
	}
	if mem > 0 {
		memStr = fmt.Sprintf("%.2f MiB", float64(mem)/(1024*1024))
	} else {
		memStr = "-"
	}
	return cpuStr, memStr
}

func FormatPodResourceUsage(podMetricsMap model.PodMetricsMap, namespace, name string) (cpuStr, memStr string) {
	cpuStr, memStr = "-", "-"
	if m, ok := podMetricsMap[namespace+"/"+name]; ok {
		cpuStr = fmt.Sprintf("%.2f mCPU", float64(m.CPU))
		memStr = fmt.Sprintf("%.2f MiB", float64(m.Mem)/(1024*1024))
	}
	return cpuStr, memStr
}

func GetResourceStatus(ready, desired int32) string {
	if ready == desired && desired > 0 {
		return "Ready"
	} else if ready > 0 {
		return "PartialAvailable"
	} else if desired == 0 {
		return "Scaled to zero"
	} else {
		return "Not Ready"
	}
}

func GetWorkloadStatus(ready, desired int32) string {
	if ready == desired && desired > 0 {
		return "Healthy"
	} else if ready > 0 {
		return "PartialAvailable"
	} else {
		return "Abnormal"
	}
}

func GetJobStatus(succeeded, failed, active int32) string {
	if succeeded > 0 {
		return "Succeeded"
	} else if failed > 0 {
		return "Failed"
	} else if active > 0 {
		return "Running"
	} else {
		return "Pending"
	}
}

func GetCronJobStatus(activeCount int, lastSuccessfulTime interface{}) string {
	if activeCount > 0 {
		return "Running"
	} else if lastSuccessfulTime != nil {
		return "Succeeded"
	} else {
		return "Pending"
	}
}

func ExtractKeys[T any](data map[string]T) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	return keys
}

func SafeInt32Ptr(ptr *int32, defaultValue int32) int32 {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

func SafeBoolPtr(ptr *bool, defaultValue bool) bool {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

func SafeStringPtr(ptr *string, defaultValue string) string {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

func SafeInt64Ptr(ptr *int64, defaultValue int64) int64 {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

func SafeFloat64Ptr(ptr *float64, defaultValue float64) float64 {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

func IsEmptyString(s string) bool {
	return len(s) == 0
}

func IsNotEmptyString(s string) bool {
	return len(s) > 0
}

func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func ContainsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func RemoveEmptyStrings(slice []string) []string {
	var result []string
	for _, s := range slice {
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

func MergeStringMaps(m1, m2 map[string]string) map[string]string {
	result := make(map[string]string)

	for k, v := range m1 {
		result[k] = v
	}

	for k, v := range m2 {
		result[k] = v
	}

	return result
}

func GetMapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func GetMapValues[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}
