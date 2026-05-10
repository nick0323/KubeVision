package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/service"
	"go.uber.org/zap"
)

func RegisterCRDRoutes(r *gin.RouterGroup, logger *zap.Logger, k8sClientMgr *service.ClientManager) {
	crdGroup := r.Group("/crds")
	{
		crdGroup.GET("", listCRDs(logger, k8sClientMgr))
		crdGroup.GET("/:group/:version/:plural", listCRDInstances(logger, k8sClientMgr))
		crdGroup.GET("/:group/:version/:plural/:namespace/:name", getCRDInstance(logger, k8sClientMgr))
	}
}

func listCRDs(logger *zap.Logger, k8sClientMgr *service.ClientManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		crdMgr := k8sClientMgr.GetCRDManager()
		if crdMgr == nil {
			middleware.ResponseError(c, logger, fmt.Errorf("CRD client not available"), http.StatusServiceUnavailable)
			return
		}

		crds, err := crdMgr.ListCRDs(c.Request.Context())
		if err != nil {
			logger.Error("Failed to list CRDs", zap.Error(err))
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, crds, "CRDs retrieved successfully", nil)
	}
}

func listCRDInstances(logger *zap.Logger, k8sClientMgr *service.ClientManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		crdMgr := k8sClientMgr.GetCRDManager()
		if crdMgr == nil {
			middleware.ResponseError(c, logger, fmt.Errorf("CRD client not available"), http.StatusServiceUnavailable)
			return
		}

		group := c.Param("group")
		version := c.Param("version")
		plural := c.Param("plural")
		namespace := c.Query("namespace")

		instances, err := crdMgr.ListCRDInstances(c.Request.Context(), group, version, plural, namespace)
		if err != nil {
			logger.Error("Failed to list CRD instances", zap.Error(err))
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, instances.Items, "CRD instances retrieved successfully", nil)
	}
}

func getCRDInstance(logger *zap.Logger, k8sClientMgr *service.ClientManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		crdMgr := k8sClientMgr.GetCRDManager()
		if crdMgr == nil {
			middleware.ResponseError(c, logger, fmt.Errorf("CRD client not available"), http.StatusServiceUnavailable)
			return
		}

		group := c.Param("group")
		version := c.Param("version")
		plural := c.Param("plural")
		namespace := c.Param("namespace")
		if namespace == "_" {
			namespace = ""
		}
		name := c.Param("name")

		instance, err := crdMgr.GetCRDInstance(c.Request.Context(), group, version, plural, namespace, name)
		if err != nil {
			logger.Error("Failed to get CRD instance", zap.Error(err))
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, instance, "CRD instance retrieved successfully", nil)
	}
}
