package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/service"
	"go.uber.org/zap"
)

// RegisterArgoCDRoutes 注册 ArgoCD 相关路由
func RegisterArgoCDRoutes(r *gin.RouterGroup, logger *zap.Logger, k8sClientMgr *service.ClientManager) {
	argoCDGroup := r.Group("/argocd")
	{
		argoCDGroup.GET("/apps", listArgoCDApps(logger, k8sClientMgr))
		argoCDGroup.GET("/apps/:name", getArgoCDApp(logger, k8sClientMgr))
		argoCDGroup.POST("/apps/:name/sync", syncArgoCDApp(logger, k8sClientMgr))
		argoCDGroup.POST("/apps/:name/refresh", refreshArgoCDApp(logger, k8sClientMgr))
		argoCDGroup.DELETE("/apps/:name", deleteArgoCDApp(logger, k8sClientMgr))
	}
}

// listArgoCDApps 列出 ArgoCD 应用
func listArgoCDApps(logger *zap.Logger, k8sClientMgr *service.ClientManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		argoCDMgr := k8sClientMgr.GetArgoCDManager()
		if argoCDMgr == nil {
			middleware.ResponseError(c, logger, fmt.Errorf("ArgoCD client not available"), http.StatusServiceUnavailable)
			return
		}

		project := c.Query("project")
		apps, err := argoCDMgr.ListApplications(c.Request.Context(), project)
		if err != nil {
			logger.Error("Failed to list ArgoCD applications", zap.Error(err))
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, apps.Items, "Applications retrieved successfully", nil)
	}
}

// getArgoCDApp 获取单个应用详情
func getArgoCDApp(logger *zap.Logger, k8sClientMgr *service.ClientManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		argoCDMgr := k8sClientMgr.GetArgoCDManager()
		if argoCDMgr == nil {
			middleware.ResponseError(c, logger, fmt.Errorf("ArgoCD client not available"), http.StatusServiceUnavailable)
			return
		}

		name := c.Param("name")
		if name == "" {
			middleware.ResponseError(c, logger, fmt.Errorf("application name is required"), http.StatusBadRequest)
			return
		}

		app, err := argoCDMgr.GetApplicationByName(c.Request.Context(), name)
		if err != nil {
			logger.Error("Failed to get ArgoCD application", zap.String("name", name), zap.Error(err))
			middleware.ResponseError(c, logger, fmt.Errorf("application not found"), http.StatusNotFound)
			return
		}

		middleware.ResponseSuccess(c, app, "Application retrieved successfully", nil)
	}
}

// syncArgoCDApp 同步应用
func syncArgoCDApp(logger *zap.Logger, k8sClientMgr *service.ClientManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		argoCDMgr := k8sClientMgr.GetArgoCDManager()
		if argoCDMgr == nil {
			middleware.ResponseError(c, logger, fmt.Errorf("ArgoCD client not available"), http.StatusServiceUnavailable)
			return
		}

		name := c.Param("name")
		if name == "" {
			middleware.ResponseError(c, logger, fmt.Errorf("application name is required"), http.StatusBadRequest)
			return
		}

		err := argoCDMgr.SyncApplicationByName(c.Request.Context(), name)
		if err != nil {
			logger.Error("Failed to sync ArgoCD application", zap.String("name", name), zap.Error(err))
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, nil, "Application sync triggered", nil)
	}
}

// refreshArgoCDApp 刷新应用状态
func refreshArgoCDApp(logger *zap.Logger, k8sClientMgr *service.ClientManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		argoCDMgr := k8sClientMgr.GetArgoCDManager()
		if argoCDMgr == nil {
			middleware.ResponseError(c, logger, fmt.Errorf("ArgoCD client not available"), http.StatusServiceUnavailable)
			return
		}

		name := c.Param("name")
		if name == "" {
			middleware.ResponseError(c, logger, fmt.Errorf("application name is required"), http.StatusBadRequest)
			return
		}

		err := argoCDMgr.RefreshApplicationByName(c.Request.Context(), name)
		if err != nil {
			logger.Error("Failed to refresh ArgoCD application", zap.String("name", name), zap.Error(err))
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, nil, "Application refresh triggered", nil)
	}
}

// deleteArgoCDApp 删除应用
func deleteArgoCDApp(logger *zap.Logger, k8sClientMgr *service.ClientManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		argoCDMgr := k8sClientMgr.GetArgoCDManager()
		if argoCDMgr == nil {
			middleware.ResponseError(c, logger, fmt.Errorf("ArgoCD client not available"), http.StatusServiceUnavailable)
			return
		}

		name := c.Param("name")
		if name == "" {
			middleware.ResponseError(c, logger, fmt.Errorf("application name is required"), http.StatusBadRequest)
			return
		}

		err := argoCDMgr.DeleteApplicationByName(c.Request.Context(), name)
		if err != nil {
			logger.Error("Failed to delete ArgoCD application", zap.String("name", name), zap.Error(err))
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, nil, "Application deleted successfully", nil)
	}
}
