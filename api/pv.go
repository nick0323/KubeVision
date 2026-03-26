package api

import (
	"context"

	"github.com/nick0323/K8sVision/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

func RegisterPV(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listPVs func(context.Context, *kubernetes.Clientset) ([]model.PVStatus, error),
) {
	r.GET("/pvs", getPVList(logger, getK8sClient, listPVs))
}

func getPVList(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listPVs func(context.Context, *kubernetes.Clientset) ([]model.PVStatus, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		HandleListWithPagination(c, logger, func(ctx context.Context, params PaginationParams) ([]model.PVStatus, error) {
			clientset, _, err := getK8sClient()
			if err != nil {
				return nil, err
			}
			return listPVs(ctx, clientset)
		}, ListSuccessMessage)
	}
}
