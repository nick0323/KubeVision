// Package api 提供 Kubernetes 资源操作接口
package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nick0323/K8sVision/api/middleware"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// WebSocket 连接管理
var (
	wsConnectionCount atomic.Int32
	maxWSConnections  = int32(100)
)

// ClusterScopeResources 集群级资源列表（不需要 namespace）
var ClusterScopeResources = map[string]bool{
	"persistentvolume": true,
	"pv":               true,
	"storageclass":     true,
	"namespace":        true,
	"node":             true,
}

// 输入验证：检查 DNS 名称格式（namespace 和资源名称）
func isValidDNSName(name string) bool {
	if len(name) == 0 || len(name) > 253 {
		return false
	}
	// DNS-1123 标签格式：小写字母、数字、连字符，必须以字母或数字开头和结尾
	for i, r := range name {
		if r == '-' {
			if i == 0 || i == len(name)-1 {
				return false
			}
			continue
		}
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		return false
	}
	return true
}

// 输入验证：检查资源名称是否包含危险字符
func isValidResourceName(name string) bool {
	if !isValidDNSName(name) {
		return false
	}
	// 防止路径遍历攻击
	if strings.ContainsAny(name, "../\\") {
		return false
	}
	return true
}

// RegisterOperations 注册资源操作接口（YAML、关联资源等）
func RegisterOperations(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) {
	// 获取资源 YAML
	r.GET("/:resourceType/:namespace/:name/yaml", getResourceYAML(logger, getK8sClient))
	// 更新资源 YAML
	r.PUT("/:resourceType/:namespace/:name/yaml", updateResourceYAML(logger, getK8sClient))
	// 获取关联资源（命名空间级资源）
	r.GET("/:resourceType/:namespace/:name/related", getResourceRelated(logger, getK8sClient))
	// 获取关联资源（集群级资源，不带 namespace）
	r.GET("/:resourceType/_cluster_/:name/related", getResourceRelatedCluster(logger, getK8sClient))
}

// RegisterLogStream 注册日志流接口（WebSocket）
func RegisterLogStream(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) {
	// 日志实时流 WebSocket
	r.GET("/ws/stream", func(c *gin.Context) {
		logger.Info("收到日志 WebSocket 请求",
			zap.String("path", c.Request.URL.Path),
			zap.String("query", sanitizeRawQuery(c.Request.URL.RawQuery)),
		)
		streamPodLog(logger, getK8sClient)(c)
		// 停止 Gin 继续执行中间件
		c.Abort()
	})
}

// getResourceYAML 获取资源 YAML
// 不支持 Event 资源
func getResourceYAML(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceType := c.Param("resourceType")
		namespace := c.Param("namespace")
		name := c.Param("name")

		// Event 不支持 YAML 接口
		if resourceType == "event" || resourceType == "events" {
			middleware.ResponseError(c, logger, fmt.Errorf("Event 资源不支持 YAML 格式"), http.StatusBadRequest)
			return
		}

		if err := validateResourceParams(resourceType, namespace); err != nil {
			middleware.ResponseError(c, logger, err, http.StatusBadRequest)
			return
		}

		clientset, _, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		ctx := GetRequestContext(c)
		obj, err := getResourceByName(ctx, clientset, resourceType, namespace, name)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}

		yamlBytes, err := yaml.Marshal(obj)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, string(yamlBytes), "YAML 获取成功", nil)
	}
}

// updateResourceYAML 更新资源 YAML
func updateResourceYAML(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceType := c.Param("resourceType")
		namespace := c.Param("namespace")
		name := c.Param("name")

		// Event 不支持 YAML 接口
		if resourceType == "event" || resourceType == "events" {
			middleware.ResponseError(c, logger, fmt.Errorf("Event 资源不支持 YAML 更新"), http.StatusBadRequest)
			return
		}

		if err := validateResourceParams(resourceType, namespace); err != nil {
			middleware.ResponseError(c, logger, err, http.StatusBadRequest)
			return
		}

		// 解析请求体 - 支持两种格式：{yaml: {...}} 或直接 {...}
		var reqBody map[string]interface{}
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			middleware.ResponseError(c, logger, fmt.Errorf("无效的请求体：%v", err), http.StatusBadRequest)
			return
		}

		// 如果是 {yaml: {...}} 格式，提取 yaml 字段
		var objData interface{} = reqBody
		if yamlData, ok := reqBody["yaml"]; ok {
			objData = yamlData
		}

		if err := validateResourceIdentity(resourceType, namespace, name, objData); err != nil {
			middleware.ResponseError(c, logger, err, http.StatusBadRequest)
			return
		}

		// 将 map 转换为 JSON 字节
		jsonBytes, err := json.Marshal(objData)
		if err != nil {
			middleware.ResponseError(c, logger, fmt.Errorf("JSON 序列化失败：%v", err), http.StatusBadRequest)
			return
		}

		clientset, _, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		ctx := GetRequestContext(c)

		// 根据资源类型调用不同的更新方法
		err = updateResourceByType(ctx, clientset, resourceType, namespace, name, jsonBytes)
		if err != nil {
			logger.Error("更新资源失败", zap.Error(err))
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, nil, "资源更新成功", nil)
	}
}

// updateResourceByType 根据资源类型更新资源
// 注意：K8s Update 操作需要 resourceVersion 字段
func updateResourceByType(ctx context.Context, clientset *kubernetes.Clientset, resourceType, namespace, name string, jsonBytes []byte) error {
	resourceType = strings.ToLower(resourceType)

	switch resourceType {
	case "pod":
		pod := &v1.Pod{}
		if err := json.Unmarshal(jsonBytes, pod); err != nil {
			return fmt.Errorf("无效的 Pod 对象：%v", err)
		}
		// 检查必需字段
		if pod.ResourceVersion == "" {
			return fmt.Errorf("缺少必需字段：resourceVersion")
		}
		_, err := clientset.CoreV1().Pods(namespace).Update(ctx, pod, metav1.UpdateOptions{})
		return err
	case "deployment":
		dep := &appsv1.Deployment{}
		if err := json.Unmarshal(jsonBytes, dep); err != nil {
			return fmt.Errorf("无效的 Deployment 对象：%v", err)
		}
		if dep.ResourceVersion == "" {
			return fmt.Errorf("缺少必需字段：resourceVersion")
		}
		_, err := clientset.AppsV1().Deployments(namespace).Update(ctx, dep, metav1.UpdateOptions{})
		return err
	case "statefulset":
		sts := &appsv1.StatefulSet{}
		if err := json.Unmarshal(jsonBytes, sts); err != nil {
			return fmt.Errorf("无效的 StatefulSet 对象：%v", err)
		}
		if sts.ResourceVersion == "" {
			return fmt.Errorf("缺少必需字段：resourceVersion")
		}
		_, err := clientset.AppsV1().StatefulSets(namespace).Update(ctx, sts, metav1.UpdateOptions{})
		return err
	case "daemonset":
		ds := &appsv1.DaemonSet{}
		if err := json.Unmarshal(jsonBytes, ds); err != nil {
			return fmt.Errorf("无效的 DaemonSet 对象：%v", err)
		}
		if ds.ResourceVersion == "" {
			return fmt.Errorf("缺少必需字段：resourceVersion")
		}
		_, err := clientset.AppsV1().DaemonSets(namespace).Update(ctx, ds, metav1.UpdateOptions{})
		return err
	case "service":
		svc := &v1.Service{}
		if err := json.Unmarshal(jsonBytes, svc); err != nil {
			return fmt.Errorf("无效的 Service 对象：%v", err)
		}
		if svc.ResourceVersion == "" {
			return fmt.Errorf("缺少必需字段：resourceVersion")
		}
		_, err := clientset.CoreV1().Services(namespace).Update(ctx, svc, metav1.UpdateOptions{})
		return err
	case "configmap":
		cm := &v1.ConfigMap{}
		if err := json.Unmarshal(jsonBytes, cm); err != nil {
			return fmt.Errorf("无效的 ConfigMap 对象：%v", err)
		}
		if cm.ResourceVersion == "" {
			return fmt.Errorf("缺少必需字段：resourceVersion")
		}
		_, err := clientset.CoreV1().ConfigMaps(namespace).Update(ctx, cm, metav1.UpdateOptions{})
		return err
	case "secret":
		secret := &v1.Secret{}
		if err := json.Unmarshal(jsonBytes, secret); err != nil {
			return fmt.Errorf("无效的 Secret 对象：%v", err)
		}
		if secret.ResourceVersion == "" {
			return fmt.Errorf("缺少必需字段：resourceVersion")
		}
		_, err := clientset.CoreV1().Secrets(namespace).Update(ctx, secret, metav1.UpdateOptions{})
		return err
	case "ingress":
		ing := &networkingv1.Ingress{}
		if err := json.Unmarshal(jsonBytes, ing); err != nil {
			return fmt.Errorf("无效的 Ingress 对象：%v", err)
		}
		if ing.ResourceVersion == "" {
			return fmt.Errorf("缺少必需字段：resourceVersion")
		}
		_, err := clientset.NetworkingV1().Ingresses(namespace).Update(ctx, ing, metav1.UpdateOptions{})
		return err
	case "job":
		job := &batchv1.Job{}
		if err := json.Unmarshal(jsonBytes, job); err != nil {
			return fmt.Errorf("无效的 Job 对象：%v", err)
		}
		if job.ResourceVersion == "" {
			return fmt.Errorf("缺少必需字段：resourceVersion")
		}
		_, err := clientset.BatchV1().Jobs(namespace).Update(ctx, job, metav1.UpdateOptions{})
		return err
	case "cronjob":
		cj := &batchv1.CronJob{}
		if err := json.Unmarshal(jsonBytes, cj); err != nil {
			return fmt.Errorf("无效的 CronJob 对象：%v", err)
		}
		if cj.ResourceVersion == "" {
			return fmt.Errorf("缺少必需字段：resourceVersion")
		}
		_, err := clientset.BatchV1().CronJobs(namespace).Update(ctx, cj, metav1.UpdateOptions{})
		return err
	case "persistentvolumeclaim", "pvc":
		pvc := &v1.PersistentVolumeClaim{}
		if err := json.Unmarshal(jsonBytes, pvc); err != nil {
			return fmt.Errorf("无效的 PVC 对象：%v", err)
		}
		if pvc.ResourceVersion == "" {
			return fmt.Errorf("缺少必需字段：resourceVersion")
		}
		_, err := clientset.CoreV1().PersistentVolumeClaims(namespace).Update(ctx, pvc, metav1.UpdateOptions{})
		return err
	case "persistentvolume", "pv":
		pv := &v1.PersistentVolume{}
		if err := json.Unmarshal(jsonBytes, pv); err != nil {
			return fmt.Errorf("无效的 PV 对象：%v", err)
		}
		if pv.ResourceVersion == "" {
			return fmt.Errorf("缺少必需字段：resourceVersion")
		}
		_, err := clientset.CoreV1().PersistentVolumes().Update(ctx, pv, metav1.UpdateOptions{})
		return err
	case "storageclass":
		sc := &storagev1.StorageClass{}
		if err := json.Unmarshal(jsonBytes, sc); err != nil {
			return fmt.Errorf("无效的 StorageClass 对象：%v", err)
		}
		if sc.ResourceVersion == "" {
			return fmt.Errorf("缺少必需字段：resourceVersion")
		}
		_, err := clientset.StorageV1().StorageClasses().Update(ctx, sc, metav1.UpdateOptions{})
		return err
	case "namespace":
		ns := &v1.Namespace{}
		if err := json.Unmarshal(jsonBytes, ns); err != nil {
			return fmt.Errorf("无效的 Namespace 对象：%v", err)
		}
		if ns.ResourceVersion == "" {
			return fmt.Errorf("缺少必需字段：resourceVersion")
		}
		_, err := clientset.CoreV1().Namespaces().Update(ctx, ns, metav1.UpdateOptions{})
		return err
	case "node":
		node := &v1.Node{}
		if err := json.Unmarshal(jsonBytes, node); err != nil {
			return fmt.Errorf("无效的 Node 对象：%v", err)
		}
		if node.ResourceVersion == "" {
			return fmt.Errorf("缺少必需字段：resourceVersion")
		}
		_, err := clientset.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
		return err
	default:
		return fmt.Errorf("不支持的资源类型：%s", resourceType)
	}
}

func getResourceRelated(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceType := c.Param("resourceType")
		namespace := c.Param("namespace")
		name := c.Param("name")

		// 输入验证
		if !isValidResourceName(namespace) {
			middleware.ResponseError(c, logger, fmt.Errorf("无效的 namespace 格式"), http.StatusBadRequest)
			return
		}
		if !isValidResourceName(name) {
			middleware.ResponseError(c, logger, fmt.Errorf("无效的资源名称格式"), http.StatusBadRequest)
			return
		}

		if err := validateResourceParams(resourceType, namespace); err != nil {
			middleware.ResponseError(c, logger, err, http.StatusBadRequest)
			return
		}

		clientset, _, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		ctx := GetRequestContext(c)
		obj, err := getResourceByName(ctx, clientset, resourceType, namespace, name)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}

		related := findRelatedResources(obj, resourceType, namespace, clientset, ctx, logger)
		middleware.ResponseSuccess(c, related, "关联资源获取成功", nil)
	}
}

// getResourceRelatedCluster 获取集群级资源的关联资源（不带 namespace）
func getResourceRelatedCluster(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceType := c.Param("resourceType")
		name := c.Param("name")
		// 集群级资源，namespace 为空
		namespace := ""

		if err := validateResourceParams(resourceType, namespace); err != nil {
			middleware.ResponseError(c, logger, err, http.StatusBadRequest)
			return
		}

		clientset, _, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		ctx := GetRequestContext(c)
		obj, err := getResourceByName(ctx, clientset, resourceType, namespace, name)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}

		related := findRelatedResources(obj, resourceType, namespace, clientset, ctx, logger)
		middleware.ResponseSuccess(c, related, "关联资源获取成功", nil)
	}
}

// streamPodLog 流式获取 Pod 日志（WebSocket 实时推送）
// 使用 k8s.io/client-go 的 Follow 模式
func streamPodLog(
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

		// 获取查询参数
		namespace := c.Query("namespace")
		podName := c.Query("pod")
		container := c.Query("container")
		tailLines := c.Query("tailLines")
		timestamps := c.Query("timestamps")
		previous := c.Query("previous")

		// 输入验证
		if err := validatePodLogParams(namespace, podName, container); err != nil {
			logger.Warn("参数验证失败", zap.Error(err))
			middleware.ResponseError(c, logger, err, http.StatusBadRequest)
			return
		}

		// WebSocket token 验证
		if err := validateWebSocketToken(c, logger); err != nil {
			return
		}

		// 检查连接数限制
		if err := checkConnectionLimit(logger); err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"type":    "error",
				"message": err.Error(),
			})
			return
		}

		// 构建日志选项
		opts := buildPodLogOptions(container, timestamps, previous, tailLines)

		// WebSocket 升级
		ws, err := upgradeWebSocket(c, logger)
		if err != nil {
			wsConnectionCount.Add(-1) // 升级失败时递减连接计数
			return
		}
		defer ws.Close()

		// 设置 WebSocket 关闭处理器（负责递减连接计数）
		wsCloseOnce := setupWebSocketCloseHandler(ws, logger, podName)

		logger.Info("日志 WebSocket 连接成功",
			zap.String("namespace", namespace),
			zap.String("pod", podName),
			zap.String("container", container),
			zap.Int32("activeConnections", wsConnectionCount.Load()),
		)

		// 获取日志流
		podLogs, err := getPodLogStream(ctx, clientset, namespace, podName, opts, logger)
		if err != nil {
			logger.Error("获取日志流失败", zap.Error(err))
			if writeErr := ws.WriteJSON(gin.H{"type": "error", "message": fmt.Sprintf("获取日志流失败：%v", err)}); writeErr != nil {
				logger.Error("发送错误消息失败", zap.Error(writeErr))
			}
			return
		}
		defer podLogs.Close()

		// 发送连接成功消息
		if writeErr := ws.WriteJSON(gin.H{
			"type":    "connected",
			"message": fmt.Sprintf("已连接到 %s/%s (%s)", namespace, podName, container),
		}); writeErr != nil {
			logger.Error("发送连接消息失败", zap.Error(writeErr))
			return
		}

		// 启动日志读取 goroutine
		logChan := make(chan string, 200)
		errorChan := make(chan error, 1)
		doneChan := make(chan struct{})

		go readPodLogs(ctx, podLogs, logChan, errorChan, doneChan, podName, logger)

		// 主循环：处理日志、心跳、超时
		runWebSocketLoop(ctx, ws, logChan, errorChan, doneChan, podName, logger, wsCloseOnce)
	}
}

// validatePodLogParams 验证 Pod 日志参数
func validatePodLogParams(namespace, podName, container string) error {
	if !isValidResourceName(namespace) {
		return fmt.Errorf("无效的 namespace 格式")
	}
	if !isValidResourceName(podName) {
		return fmt.Errorf("无效的 pod 名称格式")
	}
	if container != "" && !isValidResourceName(container) {
		return fmt.Errorf("无效的 container 名称格式")
	}
	if namespace == "" || podName == "" {
		return fmt.Errorf("namespace 和 pod 参数为必填")
	}
	return nil
}

// validateWebSocketToken 验证 WebSocket token
func validateWebSocketToken(c *gin.Context, logger *zap.Logger) error {
	tokenStr := ExtractTokenFromRequest(c)
	if tokenStr == "" {
		logger.Warn("WebSocket 缺少 token 参数")
		c.AbortWithStatus(http.StatusUnauthorized)
		return fmt.Errorf("缺少 token")
	}

	jwtSecret, err := middleware.GetJWTSecretFromConfig()
	if err != nil {
		logger.Error("JWT secret not configured", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return err
	}

	_, err = middleware.VerifyToken(tokenStr, jwtSecret)
	if err != nil {
		logger.Warn("WebSocket token 验证失败", zap.Error(err))
		c.AbortWithStatus(http.StatusUnauthorized)
		return err
	}
	return nil
}

// checkConnectionLimit 检查连接数限制
func checkConnectionLimit(logger *zap.Logger) error {
	currentConnections := wsConnectionCount.Load()
	if currentConnections >= maxWSConnections {
		logger.Warn("WebSocket 连接数过多", zap.Int32("current", currentConnections))
		return fmt.Errorf("连接数过多：当前 %d，最大 %d", currentConnections, maxWSConnections)
	}
	wsConnectionCount.Add(1)
	return nil
}

// buildPodLogOptions 构建 Pod 日志选项
func buildPodLogOptions(container, timestamps, previous, tailLines string) *v1.PodLogOptions {
	opts := &v1.PodLogOptions{
		Follow:    true,
		TailLines: nil,
	}
	if container != "" {
		opts.Container = container
	}
	if timestamps == "true" {
		opts.Timestamps = true
	}
	if previous == "true" {
		opts.Previous = true
	}
	if tailLines != "" && tailLines != "0" {
		var lines int64
		if _, err := fmt.Sscanf(tailLines, "%d", &lines); err == nil && lines > 0 {
			opts.TailLines = &lines
		}
	}
	return opts
}

// upgradeWebSocket 升级 WebSocket 连接
func upgradeWebSocket(c *gin.Context, logger *zap.Logger) (*websocket.Conn, error) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, buildWebSocketUpgradeHeaders(c))
	if err != nil {
		logger.Error("WebSocket 升级失败", zap.Error(err))
		c.Abort()
		return nil, err
	}
	return ws, nil
}

// setupWebSocketCloseHandler 设置 WebSocket 关闭处理器
func setupWebSocketCloseHandler(ws *websocket.Conn, logger *zap.Logger, podName string) *sync.Once {
	var wsCloseOnce sync.Once
	ws.SetCloseHandler(func(code int, text string) error {
		logger.Info("客户端主动断开 WebSocket 连接",
			zap.Int("code", code),
			zap.String("text", text),
			zap.String("pod", podName),
		)
		wsCloseOnce.Do(func() {
			wsConnectionCount.Add(-1)
		})
		return nil
	})
	return &wsCloseOnce
}

// getPodLogStream 获取 Pod 日志流
func getPodLogStream(ctx context.Context, clientset *kubernetes.Clientset, namespace, podName string, opts *v1.PodLogOptions, logger *zap.Logger) (io.ReadCloser, error) {
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, opts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return nil, err
	}

	logger.Info("日志流打开成功",
		zap.String("namespace", namespace),
		zap.String("pod", podName),
	)

	return podLogs, nil
}

// readPodLogs 读取 Pod 日志
func readPodLogs(ctx context.Context, podLogs io.ReadCloser, logChan chan<- string, errorChan chan<- error, doneChan chan struct{}, podName string, logger *zap.Logger) {
	defer close(doneChan)

	reader := bufio.NewReader(podLogs)
	logCount := int64(0)
	lastReadTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			logger.Debug("日志读取 goroutine 收到取消信号", zap.String("pod", podName))
			return
		default:
			line, err := reader.ReadString('\n')
			readDuration := time.Since(lastReadTime)

			if len(line) > 0 {
				count := atomic.AddInt64(&logCount, 1)
				logger.Debug("读取到日志行",
					zap.String("pod", podName),
					zap.Int64("lineNumber", count),
					zap.Int("length", len(line)),
					zap.Duration("sinceLastRead", readDuration),
				)
				select {
				case logChan <- line:
				default:
					logger.Warn("logChan 已满，丢弃日志",
						zap.String("pod", podName),
						zap.Int64("lineNumber", count),
					)
				}
				lastReadTime = time.Now()
			}

			if err != nil {
				if err == io.EOF {
					logger.Debug("日志流 EOF，等待新数据", zap.String("pod", podName))
					time.Sleep(100 * time.Millisecond)
					continue
				} else if isTimeoutError(err) {
					logger.Debug("连接超时，等待重试", zap.String("pod", podName), zap.Error(err))
					time.Sleep(100 * time.Millisecond)
					continue
				} else {
					logger.Error("读取日志流错误", zap.String("pod", podName), zap.Error(err))
					errorChan <- err
					return
				}
			}
		}
	}
}

// runWebSocketLoop 运行 WebSocket 主循环
func runWebSocketLoop(ctx context.Context, ws *websocket.Conn, logChan <-chan string, errorChan <-chan error, doneChan <-chan struct{}, podName string, logger *zap.Logger, wsCloseOnce *sync.Once) {
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	maxDuration := 10 * time.Minute
	timeoutTimer := time.NewTimer(maxDuration)
	defer timeoutTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("上下文取消，关闭日志流", zap.String("pod", podName))
			wsCloseOnce.Do(func() {
				wsConnectionCount.Add(-1)
			})
			return
		case <-timeoutTimer.C:
			logger.Info("连接超时，关闭日志流", zap.String("pod", podName), zap.Duration("maxDuration", maxDuration))
			if writeErr := ws.WriteJSON(gin.H{"type": "info", "message": "连接超时，请重新连接"}); writeErr != nil {
				logger.Error("发送超时消息失败", zap.Error(writeErr))
			}
			wsCloseOnce.Do(func() {
				wsConnectionCount.Add(-1)
			})
			return
		case <-heartbeatTicker.C:
			if err := ws.WriteJSON(gin.H{"type": "heartbeat"}); err != nil {
				// 客户端可能已断开连接
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Info("WebSocket 客户端断开连接（心跳检测）",
						zap.String("pod", podName),
						zap.Error(err),
					)
				} else {
					logger.Debug("发送心跳失败", zap.String("pod", podName), zap.Error(err))
				}
				wsCloseOnce.Do(func() {
					wsConnectionCount.Add(-1)
				})
				return
			}
			logger.Debug("心跳发送成功", zap.String("pod", podName))
		case logLine := <-logChan:
			if err := ws.WriteJSON(gin.H{
				"type":    "log",
				"content": logLine,
			}); err != nil {
				// 客户端可能已断开连接
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Info("WebSocket 客户端断开连接",
						zap.String("pod", podName),
						zap.Error(err),
					)
				} else {
					logger.Error("发送日志到 WebSocket 失败", zap.String("pod", podName), zap.Error(err))
				}
				wsCloseOnce.Do(func() {
					wsConnectionCount.Add(-1)
				})
				return
			}
		case err := <-errorChan:
			logger.Error("日志读取错误", zap.String("pod", podName), zap.Error(err))
			if writeErr := ws.WriteJSON(gin.H{"type": "error", "message": fmt.Sprintf("读取日志失败：%v", err)}); writeErr != nil {
				logger.Error("发送错误消息失败", zap.Error(writeErr))
			}
			wsCloseOnce.Do(func() {
				wsConnectionCount.Add(-1)
			})
			return
		case <-doneChan:
			logger.Info("日志读取 goroutine 退出", zap.String("pod", podName))
			wsCloseOnce.Do(func() {
				wsConnectionCount.Add(-1)
			})
			return
		}
	}
}

// isTimeoutError 判断是否为超时错误
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "Client.Timeout") ||
		strings.Contains(errStr, "context cancellation") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "i/o timeout")
}

// validateResourceParams 验证资源类型和 namespace 参数
func validateResourceParams(resourceType, namespace string) error {
	normalizedType := strings.ToLower(strings.TrimSpace(resourceType))
	if strings.HasSuffix(normalizedType, "ses") {
		normalizedType = resourceType
	} else if strings.HasSuffix(normalizedType, "s") && !strings.HasSuffix(normalizedType, "ss") {
		normalizedType = normalizedType[:len(normalizedType)-1]
	}

	if ClusterScopeResources[normalizedType] {
		if namespace != "" {
			return fmt.Errorf("资源类型 %s 为集群级资源，不应指定 namespace", resourceType)
		}
	} else {
		if namespace == "" {
			return fmt.Errorf("资源类型 %s 为命名空间级资源，必须指定 namespace", resourceType)
		}
	}
	return nil
}

func validateResourceIdentity(resourceType, namespace, name string, objData interface{}) error {
	var payload struct {
		Metadata struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		} `json:"metadata"`
	}

	raw, err := json.Marshal(objData)
	if err != nil {
		return fmt.Errorf("failed to marshal resource identity: %w", err)
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return fmt.Errorf("failed to parse resource identity: %w", err)
	}

	if payload.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if payload.Metadata.Name != name {
		return fmt.Errorf("metadata.name does not match request path")
	}

	normalizedType := strings.ToLower(strings.TrimSpace(resourceType))
	if strings.HasSuffix(normalizedType, "ses") {
		normalizedType = resourceType
	} else if strings.HasSuffix(normalizedType, "s") && !strings.HasSuffix(normalizedType, "ss") {
		normalizedType = normalizedType[:len(normalizedType)-1]
	}

	if ClusterScopeResources[normalizedType] {
		if payload.Metadata.Namespace != "" {
			return fmt.Errorf("cluster-scoped resource must not include metadata.namespace")
		}
		return nil
	}

	if payload.Metadata.Namespace == "" {
		return fmt.Errorf("metadata.namespace is required")
	}
	if payload.Metadata.Namespace != namespace {
		return fmt.Errorf("metadata.namespace does not match request path")
	}

	return nil
}

// 关联资源查询的最大数量限制
const maxRelatedResources = 100

// findRelatedResources 查找关联资源（支持多种资源类型）
// 返回格式：[]map[string]string{kind, name, relation}
func findRelatedResources(
	obj interface{},
	resourceType string,
	namespace string,
	clientset *kubernetes.Clientset,
	ctx context.Context,
	logger *zap.Logger,
) []interface{} {
	result := make([]interface{}, 0, 50) // 预分配容量

	switch o := obj.(type) {
	// ==================== Pod ====================
	case *v1.Pod:
		// 1. 父资源 (OwnerReferences) - ReplicaSet, Deployment, Job 等
		for _, ownerRef := range o.OwnerReferences {
			if len(result) >= maxRelatedResources {
				break
			}
			result = append(result, map[string]string{
				"kind":     ownerRef.Kind,
				"name":     ownerRef.Name,
				"relation": "owner",
			})
		}
		// 2. 关联的 Service (通过 label 匹配)
		if o.Labels != nil && len(result) < maxRelatedResources {
			svcList, err := clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				logger.Warn("查询 Pod 关联的 Service 失败", zap.Error(err))
			} else {
				for _, svc := range svcList.Items {
					if len(result) >= maxRelatedResources {
						break
					}
					if svc.Spec.Selector != nil && matchesSelector(o.Labels, svc.Spec.Selector) {
						result = append(result, map[string]string{
							"kind":     "Service",
							"name":     svc.Name,
							"relation": "selectedBy",
						})
					}
				}
			}
		}
		// 3. Volume 相关 - ConfigMap, Secret, PVC
		for _, vol := range o.Spec.Volumes {
			if vol.ConfigMap != nil {
				result = append(result, map[string]string{
					"kind":     "ConfigMap",
					"name":     vol.ConfigMap.Name,
					"relation": "volume",
				})
			}
			if vol.Secret != nil {
				result = append(result, map[string]string{
					"kind":     "Secret",
					"name":     vol.Secret.SecretName,
					"relation": "volume",
				})
			}
			if vol.PersistentVolumeClaim != nil {
				result = append(result, map[string]string{
					"kind":     "PersistentVolumeClaim",
					"name":     vol.PersistentVolumeClaim.ClaimName,
					"relation": "volume",
				})
			}
		}
		// 4. Node
		if o.Spec.NodeName != "" {
			result = append(result, map[string]string{
				"kind":     "Node",
				"name":     o.Spec.NodeName,
				"relation": "scheduledOn",
			})
		}

	// ==================== Deployment ====================
	case *appsv1.Deployment:
		// 1. 子资源 - ReplicaSet
		rsList, err := clientset.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("查询 Deployment 的 ReplicaSet 失败", zap.Error(err))
		} else {
			for _, rs := range rsList.Items {
				for _, ownerRef := range rs.OwnerReferences {
					if ownerRef.Kind == "Deployment" && ownerRef.Name == o.Name && ownerRef.UID == o.UID {
						result = append(result, map[string]string{
							"kind":     "ReplicaSet",
							"name":     rs.Name,
							"relation": "child",
						})
					}
				}
			}
		}
		// 2. 关联的 Service (通过 selector 匹配)
		if o.Spec.Selector != nil {
			svcList, err := clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				logger.Warn("查询 Deployment 关联的 Service 失败", zap.Error(err))
			} else {
				for _, svc := range svcList.Items {
					if svc.Spec.Selector != nil && matchesSelector(o.Spec.Selector.MatchLabels, svc.Spec.Selector) {
						result = append(result, map[string]string{
							"kind":     "Service",
							"name":     svc.Name,
							"relation": "exposedBy",
						})
					}
				}
			}
		}
		// 3. HPA (HorizontalPodAutoscaler)
		hpaList, err := clientset.AutoscalingV1().HorizontalPodAutoscalers(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("查询 Deployment 的 HPA 失败", zap.Error(err))
		} else {
			for _, hpa := range hpaList.Items {
				if hpa.Spec.ScaleTargetRef.Kind == "Deployment" && hpa.Spec.ScaleTargetRef.Name == o.Name {
					result = append(result, map[string]string{
						"kind":     "HorizontalPodAutoscaler",
						"name":     hpa.Name,
						"relation": "autoscaled",
					})
				}
			}
		}
		// 4. PDB (PodDisruptionBudget)
		pdbList, err := clientset.PolicyV1().PodDisruptionBudgets(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("查询 Deployment 的 PDB 失败", zap.Error(err))
		} else {
			for _, pdb := range pdbList.Items {
				if pdb.Spec.Selector != nil && matchesSelector(o.Spec.Selector.MatchLabels, pdb.Spec.Selector.MatchLabels) {
					result = append(result, map[string]string{
						"kind":     "PodDisruptionBudget",
						"name":     pdb.Name,
						"relation": "protected",
					})
				}
			}
		}
		// 5. Ingress (如果 Service 关联了 Ingress)
		ingressList, err := clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("查询 Deployment 关联的 Ingress 失败", zap.Error(err))
		} else {
			// 先找到关联的 Service
			svcNames := make(map[string]bool)
			svcList, _ := clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
			for _, svc := range svcList.Items {
				if svc.Spec.Selector != nil && matchesSelector(o.Spec.Selector.MatchLabels, svc.Spec.Selector) {
					svcNames[svc.Name] = true
				}
			}
			// 再找关联这些 Service 的 Ingress
			for _, ing := range ingressList.Items {
				for _, rule := range ing.Spec.Rules {
					if rule.HTTP != nil {
						for _, path := range rule.HTTP.Paths {
							if svcNames[path.Backend.Service.Name] {
								result = append(result, map[string]string{
									"kind":     "Ingress",
									"name":     ing.Name,
									"relation": "routedBy",
								})
							}
						}
					}
				}
			}
		}

	// ==================== StatefulSet ====================
	case *appsv1.StatefulSet:
		// 1. 子资源 - Pod
		podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("查询 StatefulSet 的 Pod 失败", zap.Error(err))
		} else {
			for _, pod := range podList.Items {
				for _, ownerRef := range pod.OwnerReferences {
					if ownerRef.Kind == "StatefulSet" && ownerRef.Name == o.Name && ownerRef.UID == o.UID {
						result = append(result, map[string]string{
							"kind":     "Pod",
							"name":     pod.Name,
							"relation": "child",
						})
					}
				}
			}
		}
		// 2. 关联的 Service
		if o.Spec.Selector != nil {
			svcList, err := clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				logger.Warn("查询 StatefulSet 关联的 Service 失败", zap.Error(err))
			} else {
				for _, svc := range svcList.Items {
					if svc.Spec.Selector != nil && matchesSelector(o.Spec.Selector.MatchLabels, svc.Spec.Selector) {
						result = append(result, map[string]string{
							"kind":     "Service",
							"name":     svc.Name,
							"relation": "exposedBy",
						})
					}
				}
			}
		}
		// 3. Headless Service (spec.serviceName)
		if o.Spec.ServiceName != "" {
			result = append(result, map[string]string{
				"kind":     "Service",
				"name":     o.Spec.ServiceName,
				"relation": "headlessService",
			})
		}
		// 4. HPA
		hpaList, err := clientset.AutoscalingV1().HorizontalPodAutoscalers(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("查询 StatefulSet 的 HPA 失败", zap.Error(err))
		} else {
			for _, hpa := range hpaList.Items {
				if hpa.Spec.ScaleTargetRef.Kind == "StatefulSet" && hpa.Spec.ScaleTargetRef.Name == o.Name {
					result = append(result, map[string]string{
						"kind":     "HorizontalPodAutoscaler",
						"name":     hpa.Name,
						"relation": "autoscaled",
					})
				}
			}
		}
		// 5. PVC (volumeClaimTemplates)
		for _, pvc := range o.Spec.VolumeClaimTemplates {
			result = append(result, map[string]string{
				"kind":     "PersistentVolumeClaim",
				"name":     pvc.Name,
				"relation": "volumeClaim",
			})
		}

	// ==================== DaemonSet ====================
	case *appsv1.DaemonSet:
		// 1. 子资源 - Pod
		podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("查询 DaemonSet 的 Pod 失败", zap.Error(err))
		} else {
			for _, pod := range podList.Items {
				for _, ownerRef := range pod.OwnerReferences {
					if ownerRef.Kind == "DaemonSet" && ownerRef.Name == o.Name && ownerRef.UID == o.UID {
						result = append(result, map[string]string{
							"kind":     "Pod",
							"name":     pod.Name,
							"relation": "child",
						})
					}
				}
			}
		}
		// 2. 关联的 Service
		if o.Spec.Selector != nil {
			svcList, err := clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				logger.Warn("查询 DaemonSet 关联的 Service 失败", zap.Error(err))
			} else {
				for _, svc := range svcList.Items {
					if svc.Spec.Selector != nil && matchesSelector(o.Spec.Selector.MatchLabels, svc.Spec.Selector) {
						result = append(result, map[string]string{
							"kind":     "Service",
							"name":     svc.Name,
							"relation": "exposedBy",
						})
					}
				}
			}
		}

	// ==================== Job ====================
	case *batchv1.Job:
		// 1. 子资源 - Pod
		podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("查询 Job 的 Pod 失败", zap.Error(err))
		} else {
			for _, pod := range podList.Items {
				for _, ownerRef := range pod.OwnerReferences {
					if ownerRef.Kind == "Job" && ownerRef.Name == o.Name && ownerRef.UID == o.UID {
						result = append(result, map[string]string{
							"kind":     "Pod",
							"name":     pod.Name,
							"relation": "child",
						})
					}
				}
			}
		}
		// 2. 父资源 - CronJob
		for _, ownerRef := range o.OwnerReferences {
			if ownerRef.Kind == "CronJob" {
				result = append(result, map[string]string{
					"kind":     "CronJob",
					"name":     ownerRef.Name,
					"relation": "owner",
				})
			}
		}

	// ==================== CronJob ====================
	case *batchv1.CronJob:
		// 1. 子资源 - Job
		jobList, err := clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("查询 CronJob 的 Job 失败", zap.Error(err))
		} else {
			for _, job := range jobList.Items {
				for _, ownerRef := range job.OwnerReferences {
					if ownerRef.Kind == "CronJob" && ownerRef.Name == o.Name && ownerRef.UID == o.UID {
						result = append(result, map[string]string{
							"kind":     "Job",
							"name":     job.Name,
							"relation": "child",
						})
					}
				}
			}
		}

	// ==================== Service ====================
	case *v1.Service:
		// 1. 关联的 Pod (通过 selector 匹配)
		if o.Spec.Selector != nil {
			podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				logger.Warn("查询 Service 关联的 Pod 失败", zap.Error(err))
			} else {
				for _, pod := range podList.Items {
					if pod.Labels != nil && matchesSelector(pod.Labels, o.Spec.Selector) {
						result = append(result, map[string]string{
							"kind":     "Pod",
							"name":     pod.Name,
							"relation": "selects",
						})
					}
				}
			}
		}
		// 2. Endpoint
		epList, err := clientset.CoreV1().Endpoints(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("查询 Service 的 Endpoint 失败", zap.Error(err))
		} else {
			for _, ep := range epList.Items {
				if ep.Name == o.Name {
					result = append(result, map[string]string{
						"kind":     "Endpoints",
						"name":     ep.Name,
						"relation": "endpoints",
					})
				}
			}
		}
		// 3. Ingress
		ingressList, err := clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("查询 Service 关联的 Ingress 失败", zap.Error(err))
		} else {
			for _, ing := range ingressList.Items {
				for _, rule := range ing.Spec.Rules {
					if rule.HTTP != nil {
						for _, path := range rule.HTTP.Paths {
							if path.Backend.Service.Name == o.Name {
								result = append(result, map[string]string{
									"kind":     "Ingress",
									"name":     ing.Name,
									"relation": "routedBy",
								})
							}
						}
					}
				}
			}
		}

	// ==================== ConfigMap ====================
	case *v1.ConfigMap:
		// 使用 map 去重
		addedPods := make(map[string]bool)

		// 1. 查询通过 Volume 引用此 ConfigMap 的 Pod
		podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("查询 ConfigMap 的引用 Pod 失败", zap.Error(err))
		} else {
			for _, pod := range podList.Items {
				for _, vol := range pod.Spec.Volumes {
					if vol.ConfigMap != nil && vol.ConfigMap.Name == o.Name {
						if !addedPods[pod.Name] {
							result = append(result, map[string]string{
								"kind":     "Pod",
								"name":     pod.Name,
								"relation": "usedBy",
							})
							addedPods[pod.Name] = true
						}
						break
					}
				}
			}
		}
		// 2. 查询通过 envFrom.configMapRef 引用此 ConfigMap 的 Pod（containers 和 initContainers）
		for _, pod := range podList.Items {
			if addedPods[pod.Name] {
				continue
			}
			// 检查 containers
			for _, container := range pod.Spec.Containers {
				for _, envFrom := range container.EnvFrom {
					if envFrom.ConfigMapRef != nil && envFrom.ConfigMapRef.Name == o.Name {
						result = append(result, map[string]string{
							"kind":     "Pod",
							"name":     pod.Name,
							"relation": "usedBy",
						})
						addedPods[pod.Name] = true
						break
					}
				}
				if addedPods[pod.Name] {
					break
				}
			}
			// 检查 initContainers
			if !addedPods[pod.Name] {
				for _, container := range pod.Spec.InitContainers {
					for _, envFrom := range container.EnvFrom {
						if envFrom.ConfigMapRef != nil && envFrom.ConfigMapRef.Name == o.Name {
							result = append(result, map[string]string{
								"kind":     "Pod",
								"name":     pod.Name,
								"relation": "usedBy",
							})
							addedPods[pod.Name] = true
							break
						}
					}
					if addedPods[pod.Name] {
						break
					}
				}
			}
		}
		// 3. 查询通过 env.valueFrom.configMapKeyRef 引用此 ConfigMap 的 Pod
		for _, pod := range podList.Items {
			if addedPods[pod.Name] {
				continue
			}
			for _, container := range pod.Spec.Containers {
				for _, env := range container.Env {
					if env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil {
						if env.ValueFrom.ConfigMapKeyRef.Name == o.Name {
							result = append(result, map[string]string{
								"kind":     "Pod",
								"name":     pod.Name,
								"relation": "usedBy",
							})
							addedPods[pod.Name] = true
							break
						}
					}
				}
				if addedPods[pod.Name] {
					break
				}
			}
		}
		// 4. 查询通过 projection 引用的 Pod（高级用法）
		for _, pod := range podList.Items {
			if addedPods[pod.Name] {
				continue
			}
			for _, vol := range pod.Spec.Volumes {
				if vol.Projected != nil {
					for _, source := range vol.Projected.Sources {
						if source.ConfigMap != nil && source.ConfigMap.Name == o.Name {
							result = append(result, map[string]string{
								"kind":     "Pod",
								"name":     pod.Name,
								"relation": "usedBy",
							})
							addedPods[pod.Name] = true
							break
						}
					}
				}
				if addedPods[pod.Name] {
					break
				}
			}
		}

	// ==================== Secret ====================
	case *v1.Secret:
		// 使用 map 去重
		addedPods := make(map[string]bool)
		addedSA := make(map[string]bool)

		// 1. 查询通过 Volume 引用此 Secret 的 Pod
		podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("查询 Secret 的引用 Pod 失败", zap.Error(err))
		} else {
			for _, pod := range podList.Items {
				for _, vol := range pod.Spec.Volumes {
					if vol.Secret != nil && vol.Secret.SecretName == o.Name {
						if !addedPods[pod.Name] {
							result = append(result, map[string]string{
								"kind":     "Pod",
								"name":     pod.Name,
								"relation": "usedBy",
							})
							addedPods[pod.Name] = true
						}
						break
					}
				}
			}
		}
		// 2. 查询通过 imagePullSecrets 引用此 Secret 的 Pod
		for _, pod := range podList.Items {
			for _, ips := range pod.Spec.ImagePullSecrets {
				if ips.Name == o.Name {
					if !addedPods[pod.Name] {
						result = append(result, map[string]string{
							"kind":     "Pod",
							"name":     pod.Name,
							"relation": "usedBy",
						})
						addedPods[pod.Name] = true
					}
					break
				}
			}
		}
		// 3. 查询通过 envFrom.secretRef 引用此 Secret 的 Pod（containers 和 initContainers）
		for _, pod := range podList.Items {
			if addedPods[pod.Name] {
				continue
			}
			// 检查 containers
			for _, container := range pod.Spec.Containers {
				for _, envFrom := range container.EnvFrom {
					if envFrom.SecretRef != nil && envFrom.SecretRef.Name == o.Name {
						result = append(result, map[string]string{
							"kind":     "Pod",
							"name":     pod.Name,
							"relation": "usedBy",
						})
						addedPods[pod.Name] = true
						break
					}
				}
				if addedPods[pod.Name] {
					break
				}
			}
			// 检查 initContainers
			if !addedPods[pod.Name] {
				for _, container := range pod.Spec.InitContainers {
					for _, envFrom := range container.EnvFrom {
						if envFrom.SecretRef != nil && envFrom.SecretRef.Name == o.Name {
							result = append(result, map[string]string{
								"kind":     "Pod",
								"name":     pod.Name,
								"relation": "usedBy",
							})
							addedPods[pod.Name] = true
							break
						}
					}
					if addedPods[pod.Name] {
						break
					}
				}
			}
		}
		// 4. 查询通过 env.valueFrom.secretKeyRef 引用此 Secret 的 Pod
		for _, pod := range podList.Items {
			if addedPods[pod.Name] {
				continue
			}
			for _, container := range pod.Spec.Containers {
				for _, env := range container.Env {
					if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
						if env.ValueFrom.SecretKeyRef.Name == o.Name {
							result = append(result, map[string]string{
								"kind":     "Pod",
								"name":     pod.Name,
								"relation": "usedBy",
							})
							addedPods[pod.Name] = true
							break
						}
					}
				}
				if addedPods[pod.Name] {
					break
				}
			}
		}
		// 5. 查询引用此 Secret 的 ServiceAccount
		saList, err := clientset.CoreV1().ServiceAccounts(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("查询 Secret 的引用 ServiceAccount 失败", zap.Error(err))
		} else {
			for _, sa := range saList.Items {
				// 检查 imagePullSecrets
				for _, ips := range sa.ImagePullSecrets {
					if ips.Name == o.Name && !addedSA[sa.Name] {
						result = append(result, map[string]string{
							"kind":     "ServiceAccount",
							"name":     sa.Name,
							"relation": "usedBy",
						})
						addedSA[sa.Name] = true
						break
					}
				}
				// 检查 secrets
				for _, secret := range sa.Secrets {
					if secret.Name == o.Name && !addedSA[sa.Name] {
						result = append(result, map[string]string{
							"kind":     "ServiceAccount",
							"name":     sa.Name,
							"relation": "usedBy",
						})
						addedSA[sa.Name] = true
						break
					}
				}
			}
		}

	// ==================== Ingress ====================
	case *networkingv1.Ingress:
		// 1. 关联的 Service
		for _, rule := range o.Spec.Rules {
			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					result = append(result, map[string]string{
						"kind":     "Service",
						"name":     path.Backend.Service.Name,
						"relation": "routesTo",
					})
				}
			}
		}
		// 2. TLS Secret
		for _, tls := range o.Spec.TLS {
			if tls.SecretName != "" {
				result = append(result, map[string]string{
					"kind":     "Secret",
					"name":     tls.SecretName,
					"relation": "tlsSecret",
				})
			}
		}

	// ==================== PVC ====================
	case *v1.PersistentVolumeClaim:
		// 1. 关联的 PV
		if o.Spec.VolumeName != "" {
			result = append(result, map[string]string{
				"kind":     "PersistentVolume",
				"name":     o.Spec.VolumeName,
				"relation": "boundPV",
			})
		}
		// 2. 关联的 Pod
		podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("查询 PVC 的引用 Pod 失败", zap.Error(err))
		} else {
			for _, pod := range podList.Items {
				for _, vol := range pod.Spec.Volumes {
					if vol.PersistentVolumeClaim != nil && vol.PersistentVolumeClaim.ClaimName == o.Name {
						result = append(result, map[string]string{
							"kind":     "Pod",
							"name":     pod.Name,
							"relation": "usedBy",
						})
					}
				}
			}
		}

	// ==================== PV ====================
	case *v1.PersistentVolume:
		// 1. 关联的 PVC
		if o.Spec.ClaimRef != nil {
			result = append(result, map[string]string{
				"kind":     "PersistentVolumeClaim",
				"name":     o.Spec.ClaimRef.Name,
				"relation": "boundPVC",
			})
			// 2. 查询使用此 PVC 的 Pod
			pvcNamespace := o.Spec.ClaimRef.Namespace
			if pvcNamespace != "" {
				podList, err := clientset.CoreV1().Pods(pvcNamespace).List(ctx, metav1.ListOptions{})
				if err != nil {
					logger.Warn("查询 PV 的引用 Pod 失败", zap.Error(err))
				} else {
					for _, pod := range podList.Items {
						for _, vol := range pod.Spec.Volumes {
							if vol.PersistentVolumeClaim != nil && vol.PersistentVolumeClaim.ClaimName == o.Spec.ClaimRef.Name {
								result = append(result, map[string]string{
									"kind":     "Pod",
									"name":     pod.Name,
									"relation": "usedBy",
								})
							}
						}
					}
				}
			}
		}
		// 3. StorageClass
		if o.Spec.StorageClassName != "" {
			result = append(result, map[string]string{
				"kind":     "StorageClass",
				"name":     o.Spec.StorageClassName,
				"relation": "storageClass",
			})
		}

	// ==================== Node ====================
	case *v1.Node:
		// 查询运行在此 Node 上的 Pod
		podList, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
			FieldSelector: "spec.nodeName=" + o.Name,
		})
		if err != nil {
			logger.Warn("查询 Node 上的 Pod 失败", zap.Error(err))
		} else {
			for _, pod := range podList.Items {
				result = append(result, map[string]string{
					"kind":     "Pod",
					"name":     pod.Name,
					"relation": "scheduled",
				})
			}
		}

	// ==================== Namespace ====================
	case *v1.Namespace:
		// 查询 Namespace 中的主要资源数量
		result = append(result, map[string]string{
			"kind":     "ResourceQuota",
			"name":     "ResourceQuota",
			"relation": "quota",
		})
		// 查询此 Namespace 中的 Pod
		podList, err := clientset.CoreV1().Pods(o.Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("查询 Namespace 中的 Pod 失败", zap.Error(err))
		} else {
			for _, pod := range podList.Items {
				result = append(result, map[string]string{
					"kind":     "Pod",
					"name":     pod.Name,
					"relation": "contains",
				})
			}
		}

	// ==================== StorageClass ====================
	case *storagev1.StorageClass:
		// 查询使用此 StorageClass 的 PVC
		pvcList, err := clientset.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("查询 StorageClass 的 PVC 失败", zap.Error(err))
		} else {
			for _, pvc := range pvcList.Items {
				if pvc.Spec.StorageClassName != nil && *pvc.Spec.StorageClassName == o.Name {
					result = append(result, map[string]string{
						"kind":     "PersistentVolumeClaim",
						"name":     pvc.Name,
						"relation": "provisionedPVC",
					})
				}
			}
		}
		// 查询使用此 StorageClass 的 PV
		pvList, err := clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("查询 StorageClass 的 PV 失败", zap.Error(err))
		} else {
			for _, pv := range pvList.Items {
				if pv.Spec.StorageClassName == o.Name {
					result = append(result, map[string]string{
						"kind":     "PersistentVolume",
						"name":     pv.Name,
						"relation": "provisionedPV",
					})
				}
			}
		}
	}

	return result
}

// matchesSelector 检查 labels 是否匹配 selector
// 用于判断 Pod 是否被 Service 选中，或 Deployment 是否被 Service 暴露
func matchesSelector(labels, selector map[string]string) bool {
	if len(selector) == 0 {
		return false
	}
	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}
	return true
}
