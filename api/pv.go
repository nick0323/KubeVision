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

func RegisterPV(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listPVs func(context.Context, *kubernetes.Clientset) ([]model.PVStatus, error),
) {
	r.GET("/pvs", getPVList(logger, getK8sClient, listPVs))
	r.GET("/pvs/:name", getPVDetail(logger, getK8sClient))
}

func getPVList(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listPVs func(context.Context, *kubernetes.Clientset) ([]model.PVStatus, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		HandleListWithPagination(c, logger, func(ctx context.Context, params PaginationParams) ([]model.PVStatus, error) {
			clientset, _, err := getK8sClient()
			if err != nil {
				return nil, err
			}
			return listPVs(ctx, clientset)
		}, ListSuccessMessage)
	}
}

func getPVDetail(
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
		pv, err := clientset.CoreV1().PersistentVolumes().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}

		capacity := pv.Spec.Capacity.Storage().String()
		accessModes := []string{}
		for _, mode := range pv.Spec.AccessModes {
			accessModes = append(accessModes, string(mode))
		}

		storageClass := pv.Spec.StorageClassName

		claimRef := ""
		if pv.Spec.ClaimRef != nil {
			claimRef = pv.Spec.ClaimRef.Namespace + "/" + pv.Spec.ClaimRef.Name
		}

		pvDetail := model.PVDetail{
			Name:          pv.Name,
			Status:        string(pv.Status.Phase),
			Capacity:      capacity,
			AccessMode:    accessModes,
			StorageClass:  storageClass,
			ClaimRef:      claimRef,
			ReclaimPolicy: string(pv.Spec.PersistentVolumeReclaimPolicy),
			Labels:        pv.Labels,
		}
		middleware.ResponseSuccess(c, pvDetail, DetailSuccessMessage, nil)
	}
}
