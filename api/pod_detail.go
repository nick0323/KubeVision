package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

// RegisterPodDetail 注册 Pod 详情相关路由
func RegisterPodDetail(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) {
	r.GET("/pods/:namespace/:name", getPodDetail(logger, getK8sClient))
	r.PUT("/pods/:namespace/:name", updatePod(logger, getK8sClient))
	r.DELETE("/pods/:namespace/:name", deletePod(logger, getK8sClient))
	r.GET("/pods/:namespace/:name/yaml", getPodYAML(logger, getK8sClient))
	r.PUT("/pods/:namespace/:name/yaml", updatePodYAML(logger, getK8sClient))
	r.GET("/pods/:namespace/:name/related", getPodRelated(logger, getK8sClient))
}

// getPodDetail 获取 Pod 详情
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

		ctx := c.Request.Context()
		namespace := c.Param("namespace")
		name := c.Param("name")

		pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}

		middleware.ResponseSuccess(c, pod, "Pod 详情获取成功", nil)
	}
}

// updatePod 更新 Pod
func updatePod(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientset, _, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		ctx := c.Request.Context()
		namespace := c.Param("namespace")
		_ = c.Param("name")

		var pod v1.Pod
		if err := c.ShouldBindJSON(&pod); err != nil {
			middleware.ResponseError(c, logger, err, http.StatusBadRequest)
			return
		}

		updated, err := clientset.CoreV1().Pods(namespace).Update(ctx, &pod, metav1.UpdateOptions{})
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, updated, "Pod 更新成功", nil)
	}
}

// deletePod 删除 Pod
func deletePod(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientset, _, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		ctx := c.Request.Context()
		namespace := c.Param("namespace")
		name := c.Param("name")

		err = clientset.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, nil, "Pod 删除成功", nil)
	}
}

// getPodYAML 获取 Pod YAML
func getPodYAML(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientset, _, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		ctx := c.Request.Context()
		namespace := c.Param("namespace")
		name := c.Param("name")

		pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}

		// 转换为 YAML
		pod.TypeMeta = metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		}

		yamlData, err := yaml.Marshal(pod)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, string(yamlData), "Pod YAML 获取成功", nil)
	}
}

// updatePodYAML 更新 Pod YAML
func updatePodYAML(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientset, _, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		ctx := c.Request.Context()
		namespace := c.Param("namespace")
		_ = c.Param("name")

		var req struct {
			YAML string `json:"yaml"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.ResponseError(c, logger, err, http.StatusBadRequest)
			return
		}

		// 解析 YAML
		var pod v1.Pod
		if err := yaml.Unmarshal([]byte(req.YAML), &pod); err != nil {
			middleware.ResponseError(c, logger, err, http.StatusBadRequest)
			return
		}

		updated, err := clientset.CoreV1().Pods(namespace).Update(ctx, &pod, metav1.UpdateOptions{})
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, updated, "Pod YAML 更新成功", nil)
	}
}

// getPodRelated 获取 Pod 关联资源
func getPodRelated(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientset, _, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		ctx := c.Request.Context()
		namespace := c.Param("namespace")
		name := c.Param("name")

		// 获取 Pod
		pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}

		related := []map[string]interface{}{}

		// 查找 OwnerReferences
		for _, ownerRef := range pod.OwnerReferences {
			related = append(related, map[string]interface{}{
				"kind":       ownerRef.Kind,
				"name":       ownerRef.Name,
				"apiVersion": ownerRef.APIVersion,
				"namespace":  namespace,
			})
		}

		// 查找关联的 Service（通过 label selector）
		if len(pod.Labels) > 0 {
			services, err := clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
			if err == nil {
				for _, svc := range services.Items {
					if matchesLabels(svc.Spec.Selector, pod.Labels) {
						related = append(related, map[string]interface{}{
							"kind":       "Service",
							"name":       svc.Name,
							"apiVersion": "v1",
							"namespace":  namespace,
						})
					}
				}
			}
		}

		// 查找关联的 PVC
		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil {
				related = append(related, map[string]interface{}{
					"kind":       "PersistentVolumeClaim",
					"name":       volume.PersistentVolumeClaim.ClaimName,
					"apiVersion": "v1",
					"namespace":  namespace,
				})
			}
			if volume.ConfigMap != nil {
				related = append(related, map[string]interface{}{
					"kind":       "ConfigMap",
					"name":       volume.ConfigMap.Name,
					"apiVersion": "v1",
					"namespace":  namespace,
				})
			}
			if volume.Secret != nil {
				related = append(related, map[string]interface{}{
					"kind":       "Secret",
					"name":       volume.Secret.SecretName,
					"apiVersion": "v1",
					"namespace":  namespace,
				})
			}
		}

		middleware.ResponseSuccess(c, related, "关联资源获取成功", nil)
	}
}

// matchesLabels 检查 selector 是否匹配 labels
func matchesLabels(selector, labels map[string]string) bool {
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

// convertRuntimeObject 转换运行时对象为 YAML
func convertRuntimeObject(obj runtime.Object) (string, error) {
	data, err := yaml.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
