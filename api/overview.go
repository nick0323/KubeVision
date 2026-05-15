package api

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/service"
	"go.uber.org/zap"
)

func RegisterOverview(r *gin.RouterGroup, logger *zap.Logger, getK8sClient K8sClientProvider) {
	var (
		svcs sync.Map
	)

	r.GET("/overview", func(c *gin.Context) {
		cluster := c.Query("cluster")
		clientset, metricsClient, err := getK8sClient(cluster)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}
		if clientset == nil {
			middleware.ResponseError(c, logger, fmt.Errorf("kubernetes client unavailable"), http.StatusServiceUnavailable)
			return
		}

		key := cluster
		if key == "" {
			key = "default"
		}
		svcVal, _ := svcs.LoadOrStore(key, service.NewOverviewService(clientset, metricsClient))
		svc, _ := svcVal.(*service.OverviewService)

		overview, err := svc.GetOverview(GetRequestContext(c))
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}
		middleware.ResponseSuccess(c, overview, "Overview retrieved successfully", nil)
	})
}
