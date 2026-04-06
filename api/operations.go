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
	"time"

	"github.com/gin-gonic/gin"
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

// ClusterScopeResources 集群级资源列表（不需要 namespace）
var ClusterScopeResources = map[string]bool{
	"persistentvolume": true,
	"pv":               true,
	"storageclass":     true,
	"namespace":        true,
	"node":             true,
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
	// 获取关联资源
	r.GET("/:resourceType/:namespace/:name/related", getResourceRelated(logger, getK8sClient))
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
			zap.String("query", c.Request.URL.RawQuery),
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
func updateResourceByType(ctx context.Context, clientset *kubernetes.Clientset, resourceType, namespace, name string, jsonBytes []byte) error {
	resourceType = strings.ToLower(resourceType)

	switch resourceType {
	case "pod", "pods":
		pod := &v1.Pod{}
		if err := json.Unmarshal(jsonBytes, pod); err != nil {
			return fmt.Errorf("无效的 Pod 对象：%v", err)
		}
		_, err := clientset.CoreV1().Pods(namespace).Update(ctx, pod, metav1.UpdateOptions{})
		return err
	case "deployment", "deployments":
		dep := &appsv1.Deployment{}
		if err := json.Unmarshal(jsonBytes, dep); err != nil {
			return fmt.Errorf("无效的 Deployment 对象：%v", err)
		}
		_, err := clientset.AppsV1().Deployments(namespace).Update(ctx, dep, metav1.UpdateOptions{})
		return err
	case "statefulset", "statefulsets":
		sts := &appsv1.StatefulSet{}
		if err := json.Unmarshal(jsonBytes, sts); err != nil {
			return fmt.Errorf("无效的 StatefulSet 对象：%v", err)
		}
		_, err := clientset.AppsV1().StatefulSets(namespace).Update(ctx, sts, metav1.UpdateOptions{})
		return err
	case "daemonset", "daemonsets":
		ds := &appsv1.DaemonSet{}
		if err := json.Unmarshal(jsonBytes, ds); err != nil {
			return fmt.Errorf("无效的 DaemonSet 对象：%v", err)
		}
		_, err := clientset.AppsV1().DaemonSets(namespace).Update(ctx, ds, metav1.UpdateOptions{})
		return err
	case "service", "services":
		svc := &v1.Service{}
		if err := json.Unmarshal(jsonBytes, svc); err != nil {
			return fmt.Errorf("无效的 Service 对象：%v", err)
		}
		_, err := clientset.CoreV1().Services(namespace).Update(ctx, svc, metav1.UpdateOptions{})
		return err
	case "configmap", "configmaps":
		cm := &v1.ConfigMap{}
		if err := json.Unmarshal(jsonBytes, cm); err != nil {
			return fmt.Errorf("无效的 ConfigMap 对象：%v", err)
		}
		_, err := clientset.CoreV1().ConfigMaps(namespace).Update(ctx, cm, metav1.UpdateOptions{})
		return err
	case "secret", "secrets":
		secret := &v1.Secret{}
		if err := json.Unmarshal(jsonBytes, secret); err != nil {
			return fmt.Errorf("无效的 Secret 对象：%v", err)
		}
		_, err := clientset.CoreV1().Secrets(namespace).Update(ctx, secret, metav1.UpdateOptions{})
		return err
	case "ingress", "ingresses":
		ing := &networkingv1.Ingress{}
		if err := json.Unmarshal(jsonBytes, ing); err != nil {
			return fmt.Errorf("无效的 Ingress 对象：%v", err)
		}
		_, err := clientset.NetworkingV1().Ingresses(namespace).Update(ctx, ing, metav1.UpdateOptions{})
		return err
	case "job", "jobs":
		job := &batchv1.Job{}
		if err := json.Unmarshal(jsonBytes, job); err != nil {
			return fmt.Errorf("无效的 Job 对象：%v", err)
		}
		_, err := clientset.BatchV1().Jobs(namespace).Update(ctx, job, metav1.UpdateOptions{})
		return err
	case "cronjob", "cronjobs":
		cj := &batchv1.CronJob{}
		if err := json.Unmarshal(jsonBytes, cj); err != nil {
			return fmt.Errorf("无效的 CronJob 对象：%v", err)
		}
		_, err := clientset.BatchV1().CronJobs(namespace).Update(ctx, cj, metav1.UpdateOptions{})
		return err
	case "persistentvolumeclaim", "persistentvolumeclaims", "pvc", "pvcs":
		pvc := &v1.PersistentVolumeClaim{}
		if err := json.Unmarshal(jsonBytes, pvc); err != nil {
			return fmt.Errorf("无效的 PVC 对象：%v", err)
		}
		_, err := clientset.CoreV1().PersistentVolumeClaims(namespace).Update(ctx, pvc, metav1.UpdateOptions{})
		return err
	case "persistentvolume", "persistentvolumes", "pv", "pvs":
		pv := &v1.PersistentVolume{}
		if err := json.Unmarshal(jsonBytes, pv); err != nil {
			return fmt.Errorf("无效的 PV 对象：%v", err)
		}
		_, err := clientset.CoreV1().PersistentVolumes().Update(ctx, pv, metav1.UpdateOptions{})
		return err
	case "storageclass", "storageclasses":
		sc := &storagev1.StorageClass{}
		if err := json.Unmarshal(jsonBytes, sc); err != nil {
			return fmt.Errorf("无效的 StorageClass 对象：%v", err)
		}
		_, err := clientset.StorageV1().StorageClasses().Update(ctx, sc, metav1.UpdateOptions{})
		return err
	case "namespace", "namespaces":
		ns := &v1.Namespace{}
		if err := json.Unmarshal(jsonBytes, ns); err != nil {
			return fmt.Errorf("无效的 Namespace 对象：%v", err)
		}
		_, err := clientset.CoreV1().Namespaces().Update(ctx, ns, metav1.UpdateOptions{})
		return err
	case "node", "nodes":
		node := &v1.Node{}
		if err := json.Unmarshal(jsonBytes, node); err != nil {
			return fmt.Errorf("无效的 Node 对象：%v", err)
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

		related := findRelatedResources(obj)
		middleware.ResponseSuccess(c, related, "关联资源获取成功", nil)
	}
}

// streamPodLog 流式获取 Pod 日志（WebSocket 实时推送）
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

		if namespace == "" || podName == "" {
			logger.Warn("缺少必要参数：namespace 或 pod")
			middleware.ResponseError(c, logger, fmt.Errorf("namespace 和 pod 参数为必填"), http.StatusBadRequest)
			return
		}

		// 验证并解析 WebSocket token
		tokenStr := c.Query("token")
		if tokenStr == "" {
			logger.Warn("WebSocket 缺少 token 参数")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 验证 token 有效性
		_, err = middleware.VerifyToken(tokenStr, middleware.GetJWTSecretFromConfig())
		if err != nil {
			logger.Warn("WebSocket token 验证失败", zap.Error(err))
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

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
		if previous == "true" {
			opts.Previous = true
		}
		if tailLines != "" && tailLines != "0" {
			var lines int64
			if _, err := fmt.Sscanf(tailLines, "%d", &lines); err == nil && lines > 0 {
				opts.TailLines = &lines
			}
		}

		// 升级 WebSocket 连接
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Error("WebSocket 升级失败", zap.Error(err))
			c.Abort()
			return
		}
		defer ws.Close()

		logger.Info("日志 WebSocket 连接成功",
			zap.String("namespace", namespace),
			zap.String("pod", podName),
			zap.String("container", container),
		)

		// 获取日志流
		req := clientset.CoreV1().Pods(namespace).GetLogs(podName, opts)
		podLogs, err := req.Stream(ctx)
		if err != nil {
			logger.Error("获取日志流失败", zap.Error(err))
			ws.WriteJSON(gin.H{"type": "error", "message": fmt.Sprintf("获取日志流失败：%v", err)})
			return
		}
		defer podLogs.Close()

		logger.Info("日志流打开成功",
			zap.String("namespace", namespace),
			zap.String("pod", podName),
			zap.String("container", container),
		)

		// 发送连接成功消息
		ws.WriteJSON(gin.H{
			"type":    "connected",
			"message": fmt.Sprintf("已连接到 %s/%s (%s)", namespace, podName, container),
		})

		// 设置心跳定时器
		heartbeatTicker := time.NewTicker(30 * time.Second)
		defer heartbeatTicker.Stop()

		// 设置最大连接时长（10 分钟）
		maxDuration := 10 * time.Minute
		timeoutTimer := time.NewTimer(maxDuration)
		defer timeoutTimer.Stop()

		// 使用 bufio.Reader 按行读取日志
		reader := bufio.NewReader(podLogs)
		reconnectDelay := 5 * time.Second

		for {
			select {
			case <-ctx.Done():
				logger.Info("上下文取消，关闭日志流", zap.String("pod", podName))
				return
			case <-timeoutTimer.C:
				logger.Info("连接超时，关闭日志流", zap.String("pod", podName), zap.Duration("maxDuration", maxDuration))
				ws.WriteJSON(gin.H{"type": "info", "message": "连接超时，请重新连接"})
				return
			case <-heartbeatTicker.C:
				// 发送心跳
				if err := ws.WriteJSON(gin.H{"type": "heartbeat"}); err != nil {
					logger.Debug("发送心跳失败", zap.Error(err))
					return
				}
			default:
				// 读取日志
				line, err := reader.ReadBytes('\n')
				if len(line) > 0 {
					if err := ws.WriteJSON(gin.H{
						"type":    "log",
						"content": string(line),
					}); err != nil {
						logger.Debug("发送日志失败", zap.Error(err))
						return
					}
				}
				if err != nil {
					if err == io.EOF {
						logger.Debug("日志流 EOF，等待新数据", zap.String("pod", podName))
						select {
						case <-ctx.Done():
							return
						case <-time.After(reconnectDelay):
							continue
						}
					} else if isTimeoutError(err) {
						logger.Debug("连接超时，等待重试", zap.String("pod", podName), zap.Error(err))
						select {
						case <-ctx.Done():
							return
						case <-time.After(reconnectDelay):
							continue
						}
					} else {
						logger.Error("读取日志流错误", zap.Error(err))
						ws.WriteJSON(gin.H{"type": "error", "message": fmt.Sprintf("读取日志失败：%v", err)})
					}
					break
				}
			}
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

// findRelatedResources 查找关联资源
func findRelatedResources(obj interface{}) []interface{} {
	result := make([]interface{}, 0)

	switch o := obj.(type) {
	case *v1.Pod:
		for _, ownerRef := range o.OwnerReferences {
			result = append(result, map[string]string{
				"kind":     ownerRef.Kind,
				"name":     ownerRef.Name,
				"relation": "owner",
			})
		}
		for _, volume := range o.Spec.Volumes {
			if volume.ConfigMap != nil {
				result = append(result, map[string]string{
					"kind":     "ConfigMap",
					"name":     volume.ConfigMap.Name,
					"relation": "volume",
				})
			}
			if volume.Secret != nil {
				result = append(result, map[string]string{
					"kind":     "Secret",
					"name":     volume.Secret.SecretName,
					"relation": "volume",
				})
			}
		}

	case *v1.Service:
		if o.Spec.Selector != nil {
			result = append(result, map[string]interface{}{
				"kind":     "Pod",
				"selector": o.Spec.Selector,
				"relation": "selects",
			})
		}
	}

	return result
}
