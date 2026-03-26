package api

import (
	"context"

	"github.com/nick0323/K8sVision/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

func RegisterIngress(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listIngresses func(context.Context, *kubernetes.Clientset, string) ([]model.IngressStatus, error),
) {
	r.GET("/ingress", getIngressList(logger, getK8sClient, listIngresses))
}

func getIngressList(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listIngresses func(context.Context, *kubernetes.Clientset, string) ([]model.IngressStatus, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		HandleListWithPagination(c, logger, func(ctx context.Context, params PaginationParams) ([]model.IngressStatus, error) {
			clientset, _, err := getK8sClient()
			if err != nil {
				return nil, err
			}
			return listIngresses(ctx, clientset, params.Namespace)
		}, ListSuccessMessage)
	}
}
