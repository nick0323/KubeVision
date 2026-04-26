package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/service"
	"go.uber.org/zap"
)

// RegisterRelatedRoutes 注册关联资源路由
func RegisterRelatedRoutes(r *gin.RouterGroup, logger *zap.Logger, getK8sClient K8sClientProvider) {
	r.GET("/:resourceType/:namespace/:name/related", getResourceRelated(logger, getK8sClient))
	r.GET("/:resourceType/_cluster_/:name/related", getResourceRelatedCluster(logger, getK8sClient))
}

// getResourceRelated 获取关联资源（命名空间级资源）
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
			middleware.ResponseError(c, logger, fmt.Errorf("invalid namespace format"), http.StatusBadRequest)
			return
		}
		if !isValidResourceName(name) {
			middleware.ResponseError(c, logger, fmt.Errorf("invalid resource name format"), http.StatusBadRequest)
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
		obj, err := service.GetResourceByName(ctx, clientset, resourceType, namespace, name)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}

		// 调用 Service 层
		related := service.FindRelatedResources(obj, resourceType, namespace, clientset, ctx, logger)
		middleware.ResponseSuccess(c, related, "Related resources retrieved successfully", nil)
	}
}

// getResourceRelatedCluster 获取关联资源（集群级资源，不带 namespace）
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
		obj, err := service.GetResourceByName(ctx, clientset, resourceType, namespace, name)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}

		// 调用 Service 层
		related := service.FindRelatedResources(obj, resourceType, namespace, clientset, ctx, logger)
		middleware.ResponseSuccess(c, related, "Related resources retrieved successfully", nil)
	}
}
