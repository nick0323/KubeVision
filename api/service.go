package api

import (
	"context"

	"github.com/nick0323/K8sVision/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

func RegisterService(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listServices func(context.Context, *kubernetes.Clientset, string) ([]model.ServiceStatus, error),
) {
	r.GET("/services", getServiceList(logger, getK8sClient, listServices))
}

func getServiceList(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listServices func(context.Context, *kubernetes.Clientset, string) ([]model.ServiceStatus, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		HandleListWithPagination(c, logger, func(ctx context.Context, params PaginationParams) ([]model.ServiceStatus, error) {
			clientset, _, err := getK8sClient()
			if err != nil {
				return nil, err
			}
			return listServices(ctx, clientset, params.Namespace)
		}, ListSuccessMessage)
	}
}
