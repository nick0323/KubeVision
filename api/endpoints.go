package api

import (
	"net/http"

	"github.com/nick0323/K8sVision/api/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func RegisterEndpoints(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) {
	r.GET("/endpoints", getEndpointsList(logger, getK8sClient))
	r.GET("/endpoints/:namespace/:name", getEndpointsDetail(logger, getK8sClient))
}

func getEndpointsList(
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
		namespace := c.Query("namespace")

		endpointsList, err := clientset.CoreV1().Endpoints(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		result := make([]interface{}, 0, len(endpointsList.Items))
		for _, ep := range endpointsList.Items {
			objMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&ep)
			if err != nil {
				continue
			}
			result = append(result, objMap)
		}

		middleware.ResponseSuccess(c, result, ListSuccessMessage, nil)
	}
}

func getEndpointsDetail(
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

		endpoints, err := clientset.CoreV1().Endpoints(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}

		// 转换为 Unstructured 对象（原始 map 格式）
		objMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(endpoints)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, objMap, DetailSuccessMessage, nil)
	}
}
