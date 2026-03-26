package api

import (
	"context"

	"github.com/nick0323/K8sVision/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

func RegisterDaemonSet(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listDaemonSets func(context.Context, *kubernetes.Clientset, string) ([]model.DaemonSetStatus, error),
) {
	r.GET("/daemonsets", getDaemonSetList(logger, getK8sClient, listDaemonSets))
}

func getDaemonSetList(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listDaemonSets func(context.Context, *kubernetes.Clientset, string) ([]model.DaemonSetStatus, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		HandleListWithPagination(c, logger, func(ctx context.Context, params PaginationParams) ([]model.DaemonSetStatus, error) {
			clientset, _, err := getK8sClient()
			if err != nil {
				return nil, err
			}
			return listDaemonSets(ctx, clientset, params.Namespace)
		}, ListSuccessMessage)
	}
}
