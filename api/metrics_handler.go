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

type metricsClientProvider func() (*versioned.Clientset, error)

func RegisterK8sMetricsRoutes(r *gin.RouterGroup, logger *zap.Logger, getMetricsClient metricsClientProvider) {
	r.GET("/metrics/nodes", GetNodeMetricsHandler(logger, getMetricsClient))
}

func GetNodeMetricsHandler(logger *zap.Logger, getMetricsClient metricsClientProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		metricsClient, err := getMetricsClient()
		if err != nil {
			logger.Warn("Metrics client unavailable", zap.Error(err))
			middleware.ResponseSuccess(c, []model.NodeMetrics{}, "metrics-server not installed", nil)
			return
		}

		nodeMetricsList, err := metricsClient.MetricsV1beta1().NodeMetricses().List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			logger.Error("Failed to get node metrics", zap.Error(err))
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		result := make([]model.NodeMetrics, 0, len(nodeMetricsList.Items))
		for _, nm := range nodeMetricsList.Items {
			result = append(result, model.NodeMetrics{
				CPU:    nm.Usage.Cpu().String(),
				Memory: nm.Usage.Memory().String(),
			})
		}

		middleware.ResponseSuccess(c, result, "Node metrics retrieved successfully", nil)
	}
}

func GetNodeMetrics(ctx context.Context, metricsClient *versioned.Clientset) (map[string]model.NodeMetrics, error) {
	if metricsClient == nil {
		return nil, nil
	}

	nodeMetricsList, err := metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make(map[string]model.NodeMetrics)
	for _, nm := range nodeMetricsList.Items {
		result[nm.Name] = model.NodeMetrics{
			CPU:    nm.Usage.Cpu().String(),
			Memory: nm.Usage.Memory().String(),
		}
	}
	return result, nil
}
