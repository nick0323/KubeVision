package api

import (
	"context"
	"net/http"

	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

func RegisterService(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listServices func(context.Context, *kubernetes.Clientset, string) ([]model.ServiceStatus, error),
) {
	r.GET("/services", getServiceList(logger, getK8sClient, listServices))
	r.GET("/services/:namespace/:name", getServiceDetail(logger, getK8sClient))
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

func getServiceDetail(
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
		namespace := c.Param("namespace")
		name := c.Param("name")
		
		svc, err := clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}

		// 转换为 Unstructured 对象（原始 map 格式）
		objMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(svc)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, objMap, DetailSuccessMessage, nil)
	}
}
