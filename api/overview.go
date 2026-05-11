package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/service"
	"go.uber.org/zap"
)

func RegisterOverview(r *gin.RouterGroup, logger *zap.Logger, getK8sClient K8sClientProvider) {
	r.GET("/overview", func(c *gin.Context) {
		cluster := c.Query("cluster")
		clientset, _, err := getK8sClient(cluster)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}
		if clientset == nil {
			middleware.ResponseError(c, logger, fmt.Errorf("kubernetes client unavailable"), http.StatusServiceUnavailable)
			return
		}
		overviewService := service.NewOverviewService(clientset)
		overview, err := overviewService.GetOverview(GetRequestContext(c))
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}
		middleware.ResponseSuccess(c, overview, "Overview retrieved successfully", nil)
	})
}
