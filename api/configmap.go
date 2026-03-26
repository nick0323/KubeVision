package api

import (
	"context"

	"github.com/nick0323/K8sVision/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

func RegisterConfigMap(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listConfigMaps func(context.Context, *kubernetes.Clientset, string) ([]model.ConfigMapStatus, error),
) {
	r.GET("/configmaps", getConfigMapList(logger, getK8sClient, listConfigMaps))
}

func getConfigMapList(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listConfigMaps func(context.Context, *kubernetes.Clientset, string) ([]model.ConfigMapStatus, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		HandleListWithPagination(c, logger, func(ctx context.Context, params PaginationParams) ([]model.ConfigMapStatus, error) {
			clientset, _, err := getK8sClient()
			if err != nil {
				return nil, err
			}
			return listConfigMaps(ctx, clientset, params.Namespace)
		}, ListSuccessMessage)
	}
}
