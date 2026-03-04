package service

import (
	"context"
	"strconv"
	"strings"

	"github.com/nick0323/K8sVision/model"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

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

func ParseMemory(memStr string) float64 {
	if memStr == "" {
		return 0
	}
	if strings.HasSuffix(memStr, "Ki") {
		n, _ := strconv.ParseFloat(strings.TrimSuffix(memStr, "Ki"), 64)
		return n / 1048576
	}
	if strings.HasSuffix(memStr, "Mi") {
		n, _ := strconv.ParseFloat(strings.TrimSuffix(memStr, "Mi"), 64)
		return n / 1024
	}
	if strings.HasSuffix(memStr, "Gi") {
		n, _ := strconv.ParseFloat(strings.TrimSuffix(memStr, "Gi"), 64)
		return n
	}
	n, _ := strconv.ParseFloat(memStr, 64)
	return n
}

func GetNodeAllocatableCPU(node v1.Node) float64 {
	if v, ok := node.Status.Allocatable[v1.ResourceName(model.ResourceCPU)]; ok {
		return float64(v.MilliValue()) / 1000
	}
	return 0
}

func GetNodeAllocatableMemory(node v1.Node) float64 {
	if v, ok := node.Status.Allocatable[v1.ResourceName(model.ResourceMemory)]; ok {
		return float64(v.Value()) / 1024 / 1024 / 1024
	}
	return 0
}

func ListNodes(ctx context.Context, clientset *kubernetes.Clientset, pods *v1.PodList, nodeMetricsMap model.NodeMetricsMap) ([]model.NodeStatus, error) {
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	// 使用通用映射函数
	result := MapNodes(nodes.Items, pods, nodeMetricsMap)
	return result, nil
}
