package api

import (
	"context"

	"github.com/nick0323/K8sVision/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

func RegisterStatefulSet(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listStatefulSets func(context.Context, *kubernetes.Clientset, string) ([]model.StatefulSetStatus, error),
) {
	r.GET("/statefulsets", getStatefulSetList(logger, getK8sClient, listStatefulSets))
}

func getStatefulSetList(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listStatefulSets func(context.Context, *kubernetes.Clientset, string) ([]model.StatefulSetStatus, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		HandleListWithPagination(c, logger, func(ctx context.Context, params PaginationParams) ([]model.StatefulSetStatus, error) {
			clientset, _, err := getK8sClient()
			if err != nil {
				return nil, err
			}
			return listStatefulSets(ctx, clientset, params.Namespace)
		}, ListSuccessMessage)
	}
}
