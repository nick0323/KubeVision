package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/model"
	"go.uber.org/zap"
)

func RegisterOverview(r *gin.RouterGroup, logger *zap.Logger, getOverview func() (*model.OverviewStatus, error)) {
	r.GET("/overview", func(c *gin.Context) {
		overview, err := getOverview()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}
		middleware.ResponseSuccess(c, overview, "Overview retrieved successfully", nil)
	})
}
