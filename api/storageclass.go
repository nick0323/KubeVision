package api

import (
	"context"

	"github.com/nick0323/K8sVision/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

func RegisterStorageClass(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listStorageClasses func(context.Context, *kubernetes.Clientset) ([]model.StorageClassStatus, error),
) {
	r.GET("/storageclasses", getStorageClassList(logger, getK8sClient, listStorageClasses))
}

func getStorageClassList(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listStorageClasses func(context.Context, *kubernetes.Clientset) ([]model.StorageClassStatus, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		HandleListWithPagination(c, logger, func(ctx context.Context, params PaginationParams) ([]model.StorageClassStatus, error) {
			clientset, _, err := getK8sClient()
			if err != nil {
				return nil, err
			}
			return listStorageClasses(ctx, clientset)
		}, ListSuccessMessage)
	}
}
