package api

import (
	"context"

	"github.com/nick0323/K8sVision/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

func RegisterCronJob(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listCronJobs func(context.Context, *kubernetes.Clientset, string) ([]model.CronJobStatus, error),
) {
	r.GET("/cronjobs", getCronJobList(logger, getK8sClient, listCronJobs))
}

func getCronJobList(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listCronJobs func(context.Context, *kubernetes.Clientset, string) ([]model.CronJobStatus, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		HandleListWithPagination(c, logger, func(ctx context.Context, params PaginationParams) ([]model.CronJobStatus, error) {
			clientset, _, err := getK8sClient()
			if err != nil {
				return nil, err
			}
			return listCronJobs(ctx, clientset, params.Namespace)
		}, ListSuccessMessage)
	}
}
