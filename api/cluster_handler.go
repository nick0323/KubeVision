package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/config"
	"github.com/nick0323/K8sVision/model"
	"github.com/nick0323/K8sVision/service"
	"go.uber.org/zap"
)

type AddClusterRequest struct {
	Name       string `json:"name" binding:"required"`
	APIServer  string `json:"apiServer"`
	Token      string `json:"token"`
	Kubeconfig string `json:"kubeconfig"`
	CAFile     string `json:"caFile"`
	Insecure   bool   `json:"insecure"`
}

type TestConnectionRequest struct {
	APIServer  string `json:"apiServer"`
	Token      string `json:"token"`
	Kubeconfig string `json:"kubeconfig"`
	CAFile     string `json:"caFile"`
	Insecure   bool   `json:"insecure"`
}

func RegisterClusterRoutes(r *gin.RouterGroup, logger *zap.Logger, clientMgr *service.ClientManager, configMgr *config.Manager) {
	clusterSvc := service.NewClusterService(clientMgr, configMgr, logger)

	clusterGroup := r.Group("/clusters")
	{
		clusterGroup.GET("", listClusters(logger, clusterSvc))
		clusterGroup.GET("/health", clusterHealth(logger, clientMgr))
		clusterGroup.POST("", addCluster(logger, clusterSvc))
		clusterGroup.PUT("/:name", updateCluster(logger, clusterSvc))
		clusterGroup.DELETE("/:name", deleteCluster(logger, clusterSvc))
		clusterGroup.POST("/test", testClusterConnection(logger, clusterSvc))
	}
}

func clusterHealth(logger *zap.Logger, clientMgr *service.ClientManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()
		health := clientMgr.GetClustersHealth(ctx)
		middleware.ResponseSuccess(c, health, "ok", nil)
	}
}

func listClusters(logger *zap.Logger, svc *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusters := svc.ListClusters(c.Request.Context())
		middleware.ResponseSuccess(c, clusters, "ok", nil)
	}
}

func addCluster(logger *zap.Logger, svc *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AddClusterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.ResponseError(c, logger, err, http.StatusBadRequest)
			return
		}

		k8sCfg := &model.KubernetesConfig{
			APIServer:  req.APIServer,
			Token:      req.Token,
			Kubeconfig: req.Kubeconfig,
			CAFile:     req.CAFile,
			Insecure:   req.Insecure,
		}

		if err := svc.AddCluster(c.Request.Context(), req.Name, k8sCfg); err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		clusterCfg := &model.ClusterConfig{
			Name:       req.Name,
			APIServer:  req.APIServer,
			Token:      req.Token,
			Kubeconfig: req.Kubeconfig,
			CAFile:     req.CAFile,
			Insecure:   req.Insecure,
		}
		svc.SaveToConfig(req.Name, clusterCfg)

		middleware.ResponseSuccess(c, gin.H{"name": req.Name}, "cluster added", nil)
	}
}

func updateCluster(logger *zap.Logger, svc *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		var req AddClusterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.ResponseError(c, logger, err, http.StatusBadRequest)
			return
		}

		svc.RemoveFromConfig(name)
		svc.RemoveCluster(c.Request.Context(), name)

		k8sCfg := &model.KubernetesConfig{
			APIServer:  req.APIServer,
			Token:      req.Token,
			Kubeconfig: req.Kubeconfig,
			CAFile:     req.CAFile,
			Insecure:   req.Insecure,
		}

		if err := svc.AddCluster(c.Request.Context(), req.Name, k8sCfg); err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		clusterCfg := &model.ClusterConfig{
			Name:       req.Name,
			APIServer:  req.APIServer,
			Token:      req.Token,
			Kubeconfig: req.Kubeconfig,
			CAFile:     req.CAFile,
			Insecure:   req.Insecure,
		}
		svc.SaveToConfig(req.Name, clusterCfg)

		middleware.ResponseSuccess(c, gin.H{"name": req.Name}, "cluster updated", nil)
	}
}

func deleteCluster(logger *zap.Logger, svc *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		if err := svc.RemoveCluster(c.Request.Context(), name); err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		svc.RemoveFromConfig(name)

		middleware.ResponseSuccess(c, nil, "cluster deleted", nil)
	}
}

func testClusterConnection(logger *zap.Logger, svc *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req TestConnectionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.ResponseError(c, logger, err, http.StatusBadRequest)
			return
		}

		k8sCfg := &model.KubernetesConfig{
			APIServer:  req.APIServer,
			Token:      req.Token,
			Kubeconfig: req.Kubeconfig,
			CAFile:     req.CAFile,
			Insecure:   req.Insecure,
		}

		if err := svc.TestConnection(c.Request.Context(), k8sCfg); err != nil {
			middleware.ResponseError(c, logger, err, http.StatusBadRequest)
			return
		}

		middleware.ResponseSuccess(c, nil, "connection successful", nil)
	}
}
