package api

import (
	"net/http"

	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RegisterOverview 注册概览接口
func RegisterOverview(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getOverview func() (*model.OverviewStatus, error),
) {
	r.GET("/overview", func(c *gin.Context) {
		overview, err := getOverview()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}
		middleware.ResponseSuccess(c, overview, "Overview retrieved successfully", nil)
	})
}
