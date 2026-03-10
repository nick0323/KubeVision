package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

// formatCPU 将 CPU 毫核值转换为字符串格式
func formatCPU(milli int64) string {
	if milli == 0 {
		return "-"
	}
	return fmt.Sprintf("%dm", milli)
}

// formatMemory 将内存字节值转换为字符串格式
func formatMemory(bytes int64) string {
	if bytes == 0 {
		return "-"
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%dKi", bytes/1024)
	}
	return fmt.Sprintf("%dMi", bytes/(1024*1024))
}

func RegisterPod(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listPodsWithRaw func(context.Context, *kubernetes.Clientset, model.PodMetricsMap, string) ([]model.PodStatus, *v1.PodList, error),
) {
	r.GET("/pods", getPodList(logger, getK8sClient, listPodsWithRaw))
	r.GET("/pods/:namespace/:name", getPodDetail(logger, getK8sClient))
	r.GET("/pods/:namespace/:name/logs", getPodLogs(logger, getK8sClient))
	r.GET("/pods/:namespace/:name/logs/stream", streamPodLogs(logger, getK8sClient))
}

func getPodList(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listPodsWithRaw func(context.Context, *kubernetes.Clientset, model.PodMetricsMap, string) ([]model.PodStatus, *v1.PodList, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		HandleListWithPagination(c, logger, func(ctx context.Context, params PaginationParams) ([]model.PodStatus, error) {
			clientset, metricsClient, err := getK8sClient()
			if err != nil {
				return nil, err
			}

			metricsList, _ := metricsClient.MetricsV1beta1().PodMetricses(params.Namespace).List(ctx, metav1.ListOptions{})
			podMetricsMap := make(model.PodMetricsMap)
			if metricsList != nil {
				for _, m := range metricsList.Items {
					var cpuSum, memSum int64
					for _, ctn := range m.Containers {
						cpuSum += ctn.Usage.Cpu().MilliValue()
						memSum += ctn.Usage.Memory().Value()
					}
					podMetricsMap[m.Namespace+"/"+m.Name] = model.PodMetrics{
						CPU: formatCPU(cpuSum),
						Mem: formatMemory(memSum),
					}
				}
			}
			podStatuses, _, err := listPodsWithRaw(ctx, clientset, podMetricsMap, params.Namespace)
			return podStatuses, err
		}, ListSuccessMessage)
	}
}

func getPodDetail(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientset, _, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}
		ctx := GetRequestContext(c)
		namespace := c.Param("namespace")
		name := c.Param("name")

		pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}

		// 转换为 Unstructured 对象（原始 map 格式）
		objMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, objMap, DetailSuccessMessage, nil)
	}
}

// getPodLogs 获取 Pod 日志
func getPodLogs(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientset, _, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		ctx := GetRequestContext(c)
		namespace := c.Param("namespace")
		name := c.Param("name")

		// 获取查询参数
		container := c.Query("container")
		tailLines := c.Query("tailLines")
		timestamps := c.Query("timestamps")
		follow := c.Query("follow")

		// 构建日志选项
		opts := &v1.PodLogOptions{}
		if container != "" {
			opts.Container = container
		}
		if timestamps == "true" {
			opts.Timestamps = true
		}
		if follow == "true" {
			opts.Follow = true
		}
		if tailLines != "" && tailLines != "0" {
			var lines int64
			if _, err := fmt.Sscanf(tailLines, "%d", &lines); err == nil && lines > 0 {
				opts.TailLines = &lines
			}
		}

		// 获取日志流
		req := clientset.CoreV1().Pods(namespace).GetLogs(name, opts)
		podLogs, err := req.Stream(ctx)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}
		defer podLogs.Close()

		// 读取日志内容
		buf := make([]byte, 2000)
		logStr := ""
		for {
			n, err := podLogs.Read(buf)
			if n > 0 {
				logStr += string(buf[:n])
			}
			if err != nil {
				break
			}
		}

		middleware.ResponseSuccess(c, logStr, "日志获取成功", nil)
	}
}

// streamPodLogs 流式获取 Pod 日志（用于实时更新）
func streamPodLogs(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientset, _, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		ctx := GetRequestContext(c)
		namespace := c.Param("namespace")
		name := c.Param("name")

		// 获取查询参数
		container := c.Query("container")
		tailLines := c.Query("tailLines")
		timestamps := c.Query("timestamps")

		// 构建日志选项
		opts := &v1.PodLogOptions{
			Follow: true,
		}
		if container != "" {
			opts.Container = container
		}
		if timestamps == "true" {
			opts.Timestamps = true
		}
		if tailLines != "" && tailLines != "0" {
			var lines int64
			if _, err := fmt.Sscanf(tailLines, "%d", &lines); err == nil && lines > 0 {
				opts.TailLines = &lines
			}
		}

		// 设置 SSE 头
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")

		// 获取日志流
		req := clientset.CoreV1().Pods(namespace).GetLogs(name, opts)
		podLogs, err := req.Stream(ctx)
		if err != nil {
			logger.Error("获取日志流失败", zap.Error(err))
			return
		}
		defer podLogs.Close()

		// 流式读取并发送
		buf := make([]byte, 1024)
		for {
			n, err := podLogs.Read(buf)
			if n > 0 {
				// 发送 SSE 消息
				c.SSEvent("message", string(buf[:n]))
				c.Writer.Flush()
			}
			if err != nil {
				break
			}
		}
	}
}
