package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func RegisterNode(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listPodsWithRaw func(context.Context, *kubernetes.Clientset, model.PodMetricsMap, string) ([]model.PodStatus, *v1.PodList, error),
	listNodes func(context.Context, *kubernetes.Clientset, *v1.PodList, model.NodeMetricsMap) ([]model.NodeStatus, error),
) {
	r.GET("/nodes", getNodeList(logger, getK8sClient, listPodsWithRaw, listNodes))
	r.GET("/nodes/:name", getNodeDetail(logger, getK8sClient))
}

func getNodeList(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listPodsWithRaw func(context.Context, *kubernetes.Clientset, model.PodMetricsMap, string) ([]model.PodStatus, *v1.PodList, error),
	listNodes func(context.Context, *kubernetes.Clientset, *v1.PodList, model.NodeMetricsMap) ([]model.NodeStatus, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		HandleListWithPagination(c, logger, func(ctx context.Context, params PaginationParams) ([]model.NodeStatus, error) {
			clientset, metricsClient, err := getK8sClient()
			if err != nil {
				return nil, err
			}

			podMetricsList, _ := metricsClient.MetricsV1beta1().PodMetricses("").List(ctx, metav1.ListOptions{})
			podMetricsMap := make(model.PodMetricsMap)
			if podMetricsList != nil {
				// 这里应根据实际类型断言和处理
			}
			_, podList, _ := listPodsWithRaw(ctx, clientset, podMetricsMap, "")
			metricsList, _ := metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
			nodeMetricsMap := make(model.NodeMetricsMap)
			if metricsList != nil {
				for _, m := range metricsList.Items {
					cpu := m.Usage.Cpu().String()
					mem := m.Usage.Memory().String()
					nodeMetricsMap[m.Name] = model.NodeMetrics{CPU: cpu, Mem: mem}
				}
			}
			return listNodes(ctx, clientset, podList, nodeMetricsMap)
		}, ListSuccessMessage)
	}
}

func getNodeDetail(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientset, metricsClient, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}
		ctx := GetRequestContext(c)
		name := c.Param("name")
		node, err := clientset.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}
		pods, _ := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
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
		ip := ""
		for _, addr := range node.Status.Addresses {
			if addr.Type == "InternalIP" {
				ip = addr.Address
				break
			}
		}
		roles := make([]string, 0)
		for key := range node.Labels {
			if strings.HasPrefix(key, model.LabelNodeRolePrefix) {
				role := strings.TrimPrefix(key, model.LabelNodeRolePrefix)
				if role == "" {
					role = "worker"
				}
				roles = append(roles, role)
			}
		}
		if len(roles) == 0 {
			roles = append(roles, "worker")
		}
		metricsList, _ := metricsClient.MetricsV1beta1().NodeMetricses().Get(ctx, name, metav1.GetOptions{})
		var cpuPercent, memPercent float64
		if metricsList != nil {
			cpuUsed := metricsList.Usage.Cpu().MilliValue()
			memUsed := metricsList.Usage.Memory().Value()
			cpuTotal := node.Status.Allocatable.Cpu().MilliValue()
			memTotal := node.Status.Allocatable.Memory().Value()
			if cpuTotal > 0 {
				cpuPercent = float64(cpuUsed) / float64(cpuTotal) * 100
			}
			if memTotal > 0 {
				memPercent = float64(memUsed) / float64(memTotal) * 100
			}
		}
		nodeDetail := model.NodeDetail{
			CommonResourceFields: model.CommonResourceFields{
				Namespace: "", // Node没有namespace
				Name:      node.Name,
				Status:    string(node.Status.Conditions[len(node.Status.Conditions)-1].Type),
				BaseMetadata: model.BaseMetadata{
					Labels:      node.Labels,
					Annotations: node.Annotations,
				},
			},
			IP:           ip,
			CPUUsage:     cpuPercent,
			MemoryUsage:  memPercent,
			Role:         roles,
			PodsUsed:     podsUsed,
			PodsCapacity: podsCapacity,
		}
		middleware.ResponseSuccess(c, nodeDetail, DetailSuccessMessage, nil)
	}
}
