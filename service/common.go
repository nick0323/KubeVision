package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/nick0323/K8sVision/model"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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
		memStr = fmt.Sprintf("%.2f MiB", float64(mem)/BytesPerMiB)
	} else {
		memStr = "-"
	}
	return cpuStr, memStr
}

func FormatPodResourceUsage(podMetricsMap model.PodMetricsMap, namespace, name string) (cpuStr, memStr string) {
	cpuStr, memStr = "-", "-"
	if m, ok := podMetricsMap[namespace+"/"+name]; ok {
		if m.CPU != "" {
			cpuStr = m.CPU
		}
		if m.Mem != "" {
			memStr = m.Mem
		}
	}
	return cpuStr, memStr
}

func GetResourceStatus(ready, desired int32) string {
	if ready == desired && desired > 0 {
		return model.WorkloadHealthy
	} else if ready > 0 {
		return model.WorkloadPartial
	} else if desired == 0 {
		return model.WorkloadScaledToZero
	}
	return model.WorkloadUnavailable
}

func GetWorkloadStatus(ready, desired int32) string {
	if ready == desired && desired > 0 {
		return model.WorkloadHealthy
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

// CalculateDuration 计算任务运行时长
func CalculateDuration(startTime, completionTime *metav1.Time) string {
	if startTime == nil {
		return "-"
	}

	var endTime metav1.Time
	if completionTime != nil {
		endTime = *completionTime
	} else {
		endTime = metav1.Now()
	}

	duration := endTime.Time.Sub(startTime.Time)

	// 格式化时长
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	} else {
		return fmt.Sprintf("%ds", seconds)
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
