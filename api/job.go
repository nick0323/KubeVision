package api

import (
	"context"

	"github.com/nick0323/K8sVision/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

func RegisterJob(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listJobs func(context.Context, *kubernetes.Clientset, string) ([]model.JobStatus, error),
) {
	r.GET("/jobs", getJobList(logger, getK8sClient, listJobs))
}

func getJobList(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listJobs func(context.Context, *kubernetes.Clientset, string) ([]model.JobStatus, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		HandleListWithPagination(c, logger, func(ctx context.Context, params PaginationParams) ([]model.JobStatus, error) {
			clientset, _, err := getK8sClient()
			if err != nil {
				return nil, err
			}
			return listJobs(ctx, clientset, params.Namespace)
		}, ListSuccessMessage)
	}
}
