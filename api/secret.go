package api

import (
	"context"

	"github.com/nick0323/K8sVision/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

func RegisterSecret(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listSecrets func(context.Context, *kubernetes.Clientset, string) ([]model.SecretStatus, error),
) {
	r.GET("/secrets", getSecretList(logger, getK8sClient, listSecrets))
}

func getSecretList(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listSecrets func(context.Context, *kubernetes.Clientset, string) ([]model.SecretStatus, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		HandleListWithPagination(c, logger, func(ctx context.Context, params PaginationParams) ([]model.SecretStatus, error) {
			clientset, _, err := getK8sClient()
			if err != nil {
				return nil, err
			}
			return listSecrets(ctx, clientset, params.Namespace)
		}, ListSuccessMessage)
	}
}
