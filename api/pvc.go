package api

import (
	"context"

	"github.com/nick0323/K8sVision/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

func RegisterPVC(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listPVCs func(context.Context, *kubernetes.Clientset, string) ([]model.PVCStatus, error),
) {
	r.GET("/pvcs", getPVCList(logger, getK8sClient, listPVCs))
}

func getPVCList(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listPVCs func(context.Context, *kubernetes.Clientset, string) ([]model.PVCStatus, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		HandleListWithPagination(c, logger, func(ctx context.Context, params PaginationParams) ([]model.PVCStatus, error) {
			clientset, _, err := getK8sClient()
			if err != nil {
				return nil, err
			}
			return listPVCs(ctx, clientset, params.Namespace)
		}, ListSuccessMessage)
	}
}
