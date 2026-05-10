package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/service"
	"go.uber.org/zap"
)

func RegisterScaleAndRestartRoutes(r *gin.RouterGroup, logger *zap.Logger, getK8sClient K8sClientProvider) {
	r.PUT("/:resourceType/:namespace/:name/scale", scaleResource(logger, getK8sClient))
	r.POST("/:resourceType/:namespace/:name/restart", restartResource(logger, getK8sClient))
}

func scaleResource(logger *zap.Logger, getK8sClient K8sClientProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceType := strings.ToLower(c.Param("resourceType"))
		namespace := c.Param("namespace")
		name := c.Param("name")

		if resourceType == "" || name == "" {
			middleware.ResponseError(c, logger, nil, http.StatusBadRequest)
			return
		}

		var req service.ScaleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.ResponseError(c, logger, err, http.StatusBadRequest)
			return
		}

		clientset, _, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		ctx := GetRequestContext(c)
		if err := service.ScaleResource(ctx, clientset, resourceType, namespace, name, req.Replicas); err != nil {
			logger.Error("Failed to scale resource", zap.Error(err))
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, gin.H{"replicas": req.Replicas}, "Resource scaled successfully", nil)
	}
}

func restartResource(logger *zap.Logger, getK8sClient K8sClientProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceType := strings.ToLower(c.Param("resourceType"))
		namespace := c.Param("namespace")
		name := c.Param("name")

		if resourceType == "" || name == "" {
			middleware.ResponseError(c, logger, nil, http.StatusBadRequest)
			return
		}

		clientset, _, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		ctx := GetRequestContext(c)
		if err := service.RestartResource(ctx, clientset, resourceType, namespace, name); err != nil {
			logger.Error("Failed to restart resource", zap.Error(err))
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, nil, "Resource restart initiated", nil)
	}
}
