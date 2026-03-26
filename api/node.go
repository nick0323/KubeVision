package api

import (
	"context"

	"github.com/nick0323/K8sVision/model"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RegisterNode(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listPodsWithRaw func(context.Context, *kubernetes.Clientset, model.PodMetricsMap, string) ([]model.PodStatus, *v1.PodList, error),
	listNodes func(context.Context, *kubernetes.Clientset, *v1.PodList, model.NodeMetricsMap) ([]model.NodeStatus, error),
) {
	r.GET("/nodes", getNodeList(logger, getK8sClient, listPodsWithRaw, listNodes))
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
