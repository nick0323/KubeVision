package service

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/nick0323/K8sVision/model"
	v1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ResourceConverter 定义资源转换接口
type ResourceConverter[T, U any] interface {
	ConvertToModel(resource T) U
}

// GenericListFunction 通用列表函数类型
type GenericListFunction[T any] func(context.Context, *kubernetes.Clientset, string) (T, error)


// ListResources 通用资源列表函数
func ListResources[T any, U any](
	ctx context.Context,
	clientset *kubernetes.Clientset,
	namespace string,
	listFunc func(context.Context, string) (T, error),
	convertFunc func(T) []U,
) ([]U, error) {
	resources, err := listFunc(ctx, namespace)
	if err != nil {
		return nil, err
	}

	return convertFunc(resources), nil
}

// ListResourceItems 通用资源项列表函数
func ListResourceItems[T any, U any](
	ctx context.Context,
	clientset *kubernetes.Clientset,
	namespace string,
	listFunc func(context.Context, string) ([]T, error),
	convertFunc func(T) U,
) ([]U, error) {
	items, err := listFunc(ctx, namespace)
	if err != nil {
		return nil, err
	}

	result := make([]U, len(items))
	for i, item := range items {
		result[i] = convertFunc(item)
	}

	return result, nil
}

// ListResourceWithMetrics 通用带指标的资源列表函数
func ListResourceWithMetrics[T any, U any, M ~map[string]any](
	ctx context.Context,
	clientset *kubernetes.Clientset,
	metricsMap M,
	namespace string,
	listFunc func(context.Context, *kubernetes.Clientset, string) (T, error),
	convertFunc func(T, M) []U,
) ([]U, error) {
	resources, err := listFunc(ctx, clientset, namespace)
	if err != nil {
		return nil, err
	}

	return convertFunc(resources, metricsMap), nil
}

// GetResourceByNamespace 通用资源获取函数（按命名空间）
func GetResourceByNamespace[T any](
	ctx context.Context,
	clientset *kubernetes.Clientset,
	resourceGetter func(string) T,
	namespace string,
) T {
	if namespace == "" {
		// 如果没有指定命名空间，获取所有命名空间的资源
		return resourceGetter(metav1.NamespaceAll)
	}
	return resourceGetter(namespace)
}

// ApplyFilters 通用过滤应用函数
func ApplyFilters[T any](
	items []T,
	filterFunc func(T) bool,
) []T {
	var filtered []T
	for _, item := range items {
		if filterFunc(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// ApplySort 通用排序应用函数
func ApplySort[T any](
	items []T,
	sortFunc func(T, T) bool,
) []T {
	// 简单的冒泡排序实现，实际项目中应该使用更高效的算法
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if sortFunc(items[i], items[j]) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	return items
}

// FilterAndMapResource 结合过滤和映射的函数
func FilterAndMapResource[T, U any](
	items []T,
	filterFunc func(T) bool,
	mapFunc func(T) U,
) []U {
	var result []U
	for _, item := range items {
		if filterFunc(item) {
			result = append(result, mapFunc(item))
		}
	}
	return result
}

// MapDeployments 专门用于映射Deployment的函数
func MapDeployments(deployments []appsv1.Deployment) []model.DeploymentStatus {
	result := make([]model.DeploymentStatus, len(deployments))
	for i, d := range deployments {
		status := GetWorkloadStatus(d.Status.ReadyReplicas, d.Status.Replicas)
		result[i] = model.DeploymentStatus{
			Namespace: d.Namespace,
			Name:      d.Name,
			Available: d.Status.ReadyReplicas,
			Desired:   d.Status.Replicas,
			Status:    status,
		}
	}
	return result
}

// MapPods 专门用于映射Pod的函数
func MapPods(pods []v1.Pod, podMetricsMap model.PodMetricsMap) []model.PodStatus {
	result := make([]model.PodStatus, len(pods))
	for i, pod := range pods {
		cpuVal, memVal := FormatPodResourceUsage(podMetricsMap, pod.Namespace, pod.Name)
		result[i] = model.PodStatus{
			Namespace:   pod.Namespace,
			Name:        pod.Name,
			Status:      string(pod.Status.Phase),
			CPUUsage:    cpuVal,
			MemoryUsage: memVal,
			PodIP:       pod.Status.PodIP,
			NodeName:    pod.Spec.NodeName,
		}
	}
	return result
}

// MapServices 专门用于映射Service的函数
func MapServices(services []v1.Service) []model.ServiceStatus {
	result := make([]model.ServiceStatus, len(services))
	for i, svc := range services {
		ports := make([]string, len(svc.Spec.Ports))
		for j, port := range svc.Spec.Ports {
			ports[j] = fmt.Sprintf("%d/%s", port.Port, port.Protocol)
		}
		result[i] = model.ServiceStatus{
			Namespace: svc.Namespace,
			Name:      svc.Name,
			Type:      string(svc.Spec.Type),
			ClusterIP: svc.Spec.ClusterIP,
			Ports:     ports,
		}
	}
	return result
}

// MapNodes 专门用于映射Node的函数
func MapNodes(nodes []v1.Node, pods *v1.PodList, nodeMetricsMap model.NodeMetricsMap) []model.NodeStatus {
	result := make([]model.NodeStatus, len(nodes))
	for i, node := range nodes {
		status := "Unknown"
		ip := ""
		for _, addr := range node.Status.Addresses {
			if addr.Type == "InternalIP" {
				ip = addr.Address
				break
			}
		}
		for _, cond := range node.Status.Conditions {
			if cond.Type == "Ready" && cond.Status == "True" {
				status = "Active"
				break
			}
		}
		roles := []string{}
		for k := range node.Labels {
			if strings.HasPrefix(k, model.LabelNodeRolePrefix) {
				role := strings.TrimPrefix(k, model.LabelNodeRolePrefix)
				if role == "" {
					role = "worker"
				}
				roles = append(roles, role)
			}
		}
		if len(roles) == 0 {
			roles = append(roles, "worker")
		}
		podsUsed := 0
		for _, pod := range pods.Items {
			if pod.Spec.NodeName == node.Name {
				podsUsed++
			}
		}
		podsCapacity := 0
		if v, ok := node.Status.Allocatable[v1.ResourceName(model.ResourcePods)]; ok {
			podsCapacity = int(v.Value())
		}
		metric := nodeMetricsMap[node.Name]
		cpuUsed := ParseCPU(metric.CPU)
		memUsed := ParseMemory(metric.Mem)
		cpuTotal := GetNodeAllocatableCPU(node)
		memTotal := GetNodeAllocatableMemory(node)
		cpuPercent := 0.0
		memPercent := 0.0
		if cpuTotal > 0 {
			cpuPercent = cpuUsed / cpuTotal * 100
		}
		if memTotal > 0 {
			memPercent = memUsed / memTotal * 100
		}
		result[i] = model.NodeStatus{
			Name:         node.Name,
			IP:           ip,
			Status:       status,
			CPUUsage:     math.Round(cpuPercent*10) / 10,
			MemoryUsage:  math.Round(memPercent*10) / 10,
			Role:         roles,
			PodsUsed:     podsUsed,
			PodsCapacity: podsCapacity,
		}
	}
	return result
}

// MapNamespaces 专门用于映射Namespace的函数
func MapNamespaces(namespaces []v1.Namespace) []model.NamespaceDetail {
	result := make([]model.NamespaceDetail, len(namespaces))
	for i, ns := range namespaces {
		result[i] = model.NamespaceDetail{
			Name:   ns.Name,
			Status: string(ns.Status.Phase),
			BaseMetadata: model.BaseMetadata{
				Labels:      ns.Labels,
				Annotations: ns.Annotations,
			},
		}
	}
	return result
}