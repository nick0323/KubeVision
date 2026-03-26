package api

import (
	"context"

	"github.com/nick0323/K8sVision/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

func RegisterDeployment(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listDeployments func(context.Context, *kubernetes.Clientset, string) ([]model.DeploymentStatus, error),
) {
	r.GET("/deployments", getDeploymentList(logger, getK8sClient, listDeployments))
}

func getDeploymentList(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listDeployments func(context.Context, *kubernetes.Clientset, string) ([]model.DeploymentStatus, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		HandleListWithPagination(c, logger, func(ctx context.Context, params PaginationParams) ([]model.DeploymentStatus, error) {
			clientset, _, err := getK8sClient()
			if err != nil {
				return nil, err
			}
			return listDeployments(ctx, clientset, params.Namespace)
		}, ListSuccessMessage)
	}
}
