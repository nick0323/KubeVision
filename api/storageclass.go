package api

import (
	"context"
	"net/http"

	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func RegisterStorageClass(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listStorageClasses func(context.Context, *kubernetes.Clientset) ([]model.StorageClassStatus, error),
) {
	r.GET("/storageclasses", getStorageClassList(logger, getK8sClient, listStorageClasses))
	r.GET("/storageclasses/:name", getStorageClassDetail(logger, getK8sClient))
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

func getStorageClassDetail(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientset, _, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}
		ctx := GetRequestContext(c)
		name := c.Param("name")
		storageClass, err := clientset.StorageV1().StorageClasses().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}

		isDefault := false
		if storageClass.Annotations != nil {
			if _, ok := storageClass.Annotations[model.AnnotationStorageClassDefault]; ok {
				isDefault = true
			}
		}

		storageClassDetail := model.StorageClassDetail{
			Name:              storageClass.Name,
			Provisioner:       storageClass.Provisioner,
			ReclaimPolicy:     string(*storageClass.ReclaimPolicy),
			VolumeBindingMode: string(*storageClass.VolumeBindingMode),
			IsDefault:         isDefault,
			BaseMetadata: model.BaseMetadata{
				Labels:      storageClass.Labels,
				Annotations: storageClass.Annotations,
			},
			Parameters: storageClass.Parameters,
		}
		middleware.ResponseSuccess(c, storageClassDetail, DetailSuccessMessage, nil)
	}
}
