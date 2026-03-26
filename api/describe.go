package api

import (
	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/service"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

// GetK8sClientFunc 获取 K8s 客户端的函数类型
type GetK8sClientFunc func() (*kubernetes.Clientset, *versioned.Clientset, error)

// RegisterDescribe 注册统一的 describe 接口
func RegisterDescribe(r *gin.RouterGroup, logger *zap.Logger, getClient GetK8sClientFunc) {
	// 有命名空间的资源
	r.GET("/describe/:resourceType/:namespace/:name", getDescribe(logger, getClient))
}

// getDescribe 获取资源的 describe 输出
func getDescribe(
	logger *zap.Logger,
	getClient GetK8sClientFunc,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		resourceType := c.Param("resourceType")
		namespace := c.Param("namespace")
		name := c.Param("name")

		// 处理 _all 作为空命名空间（用于 Node、PV、StorageClass 等资源）
		if namespace == "_all" {
			namespace = ""
		}

		clientset, _, err := getClient()
		if err != nil {
			logger.Error("获取 K8s 客户端失败", zap.Error(err))
			middleware.ResponseError(c, logger, err, 500)
			return
		}

		ctx := c.Request.Context()

		// 根据资源类型调用对应的 describe 函数
		var result interface{}
		switch resourceType {
		case "pods", "pod":
			result, err = service.DescribePod(ctx, clientset, namespace, name)
		case "deployments", "deployment":
			result, err = service.DescribeDeployment(ctx, clientset, namespace, name)
		case "services", "service":
			result, err = service.DescribeService(ctx, clientset, namespace, name)
		case "nodes", "node":
			result, err = service.DescribeNode(ctx, clientset, name)
		case "configmaps", "configmap":
			result, err = service.DescribeConfigMap(ctx, clientset, namespace, name)
		case "secrets", "secret":
			result, err = service.DescribeSecret(ctx, clientset, namespace, name)
		case "statefulsets", "statefulset":
			result, err = service.DescribeStatefulSet(ctx, clientset, namespace, name)
		case "daemonsets", "daemonset":
			result, err = service.DescribeDaemonSet(ctx, clientset, namespace, name)
		case "ingresses", "ingress":
			result, err = service.DescribeIngress(ctx, clientset, namespace, name)
		case "cronjobs", "cronjob":
			result, err = service.DescribeCronJob(ctx, clientset, namespace, name)
		case "jobs", "job":
			result, err = service.DescribeJob(ctx, clientset, namespace, name)
		case "pvcs", "pvc":
			result, err = service.DescribePVC(ctx, clientset, namespace, name)
		case "pvs", "pv":
			result, err = service.DescribePV(ctx, clientset, name)
		case "storageclasses", "storageclass":
			result, err = service.DescribeStorageClass(ctx, clientset, name)
		case "namespaces", "namespace":
			result, err = service.DescribeNamespace(ctx, clientset, name)
		default:
			middleware.ResponseError(c, logger, nil, 400)
			return
		}

		if err != nil {
			logger.Error("Describe 失败",
				zap.String("resourceType", resourceType),
				zap.String("namespace", namespace),
				zap.String("name", name),
				zap.Error(err))
			middleware.ResponseError(c, logger, err, 500)
			return
		}

		middleware.ResponseSuccess(c, result, "Describe 成功", nil)
	}
}
