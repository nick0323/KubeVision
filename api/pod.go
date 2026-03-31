package api

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/model"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Pod 资源常量
const (
	PodDescribeSuccessMessage = "Pod describe retrieved successfully"
)

// 格式化常量
const (
	BytesPerKiB = 1024
	BytesPerMiB = 1024 * 1024
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
	if bytes < BytesPerMiB {
		return fmt.Sprintf("%dKi", bytes/BytesPerKiB)
	}
	return fmt.Sprintf("%dMi", bytes/BytesPerMiB)
}

func RegisterPod(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listPodsWithRaw func(context.Context, *kubernetes.Clientset, model.PodMetricsMap, string) ([]model.PodStatus, *v1.PodList, error),
) {
	r.GET("/pods", getPodList(logger, getK8sClient, listPodsWithRaw))
	r.GET("/pods/:namespace/:name/logs", getPodLogs(logger, getK8sClient))
	// streamPodLogs 直接在 main.go 中注册，绕过认证
}

// StreamPodLogs 导出给 main.go 使用
func StreamPodLogs(logger *zap.Logger, getK8sClient K8sClientProvider) gin.HandlerFunc {
	return streamPodLogs(logger, getK8sClient)
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
		previous := c.Query("previous")

		logger.Debug("获取日志参数",
			zap.String("container", container),
			zap.String("tailLines", tailLines),
			zap.String("previous", previous),
			zap.String("timestamps", timestamps),
		)

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
		if previous == "true" {
			opts.Previous = true
		}
		// 默认获取最后 1000 行日志
		var tailLinesDefault int64 = 1000
		opts.TailLines = &tailLinesDefault
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
			// 处理 Previous 日志不存在的情况
			if previous == "true" {
				// 返回空日志而不是 500 错误
				logger.Debug("Previous 日志不存在或不可用", zap.Error(err))
				middleware.ResponseSuccess(c, "", "Previous 日志不存在", nil)
				return
			}
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

// streamPodLogs 流式获取 Pod 日志（WebSocket 实时推送）
func streamPodLogs(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientset, _, err := getK8sClient()
		if err != nil {
			logger.Error("获取 K8s 客户端失败", zap.Error(err))
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
		// since 参数暂不支持

		// 升级 WebSocket 连接（从 query 参数获取 token）
		tokenStr := c.Query("token")
		if tokenStr == "" {
			logger.Warn("WebSocket 缺少 token 参数")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Error("WebSocket 升级失败", zap.Error(err))
			c.Abort()
			return
		}
		defer ws.Close()

		// TODO: 验证 token 有效性
		// 目前只是检查 token 是否存在

		logger.Info("日志 WebSocket 连接成功",
			zap.String("namespace", namespace),
			zap.String("pod", name),
			zap.String("container", container),
		)

		// 获取日志流
		req := clientset.CoreV1().Pods(namespace).GetLogs(name, opts)
		podLogs, err := req.Stream(ctx)
		if err != nil {
			logger.Error("获取日志流失败", zap.Error(err))
			ws.WriteJSON(gin.H{"type": "error", "message": fmt.Sprintf("Failed to get log stream: %v", err)})
			return
		}
		defer podLogs.Close()

		logger.Info("日志流打开成功",
			zap.String("namespace", namespace),
			zap.String("pod", name),
			zap.String("container", container),
		)

		// 发送连接成功消息
		ws.WriteJSON(gin.H{
			"type":    "connected",
			"message": fmt.Sprintf("Connected to %s/%s (%s)", namespace, name, container),
		})

		// 使用 bufio.Reader 按行读取日志
		reader := bufio.NewReader(podLogs)

		for {
			line, err := reader.ReadBytes('\n')
			if len(line) > 0 {
				// 发送日志行
				err = ws.WriteJSON(gin.H{
					"type":    "log",
					"content": string(line),
				})
				if err != nil {
					logger.Debug("发送日志失败", zap.Error(err))
					break
				}
			}
			if err != nil {
				if err == io.EOF {
					// 日志流结束，等待 5 秒后重新连接
					select {
					case <-ctx.Done():
						break
					case <-time.After(5 * time.Second):
						continue
					}
				} else if strings.Contains(err.Error(), "Client.Timeout") || strings.Contains(err.Error(), "context cancellation") {
					// 超时或取消，等待 5 秒后重新连接
					logger.Debug("连接超时，重新连接", zap.String("pod", name))
					select {
					case <-ctx.Done():
						break
					case <-time.After(5 * time.Second):
						continue
					}
				} else {
					logger.Error("读取日志流错误", zap.Error(err))
				}
				break
			}
		}
	}
}
