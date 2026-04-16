package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/model"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

// metricsClientProvider Metrics client 提供者
type metricsClientProvider func() (*versioned.Clientset, error)

// RegisterK8sMetricsRoutes 注册 K8s metrics 相关路由
func RegisterK8sMetricsRoutes(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getMetricsClient metricsClientProvider,
) {
	// 获取节点 metrics
	r.GET("/metrics/nodes", GetNodeMetricsHandler(logger, getMetricsClient))
}

// GetNodeMetricsHandler 获取节点 metrics Handler 工厂
func GetNodeMetricsHandler(
	logger *zap.Logger,
	getMetricsClient metricsClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 获取 metrics client
		metricsClient, err := getMetricsClient()
		if err != nil {
			logger.Warn("Metrics client unavailable", zap.Error(err))
			// metrics-server 未安装，返回空数据
			middleware.ResponseSuccess(c, []model.NodeMetrics{}, "metrics-server not installed", nil)
			return
		}

		// 获取节点 metrics
		nodeMetricsList, err := metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Error("Failed to get node metrics", zap.Error(err))
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		// 转换为前端格式
		result := make([]model.NodeMetrics, 0, len(nodeMetricsList.Items))
		for _, nm := range nodeMetricsList.Items {
			result = append(result, model.NodeMetrics{
				CPU:    nm.Usage.Cpu().String(),    // 如 "1500m" (1.5 核)
				Memory: nm.Usage.Memory().String(), // 如 "5368709120" (5GB)
			})
		}

		middleware.ResponseSuccess(c, result, "Node metrics retrieved successfully", nil)
	}
}

// GetNodeMetrics 获取节点 metrics（供 service 层调用）
func GetNodeMetrics(ctx context.Context, metricsClient *versioned.Clientset) (map[string]model.NodeMetrics, error) {
	if metricsClient == nil {
		return nil, nil // metrics-server 未安装
	}

	nodeMetricsList, err := metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make(map[string]model.NodeMetrics)
	for _, nm := range nodeMetricsList.Items {
		result[nm.Name] = model.NodeMetrics{
			CPU:    nm.Usage.Cpu().String(),    // 如 "1500m" (1.5 核)
			Memory: nm.Usage.Memory().String(), // 如 "5368709120" (5GB)
		}
	}
	return result, nil
}
