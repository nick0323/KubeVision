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

// RegisterStatefulSet 注册 StatefulSet 相关路由
func RegisterStatefulSet(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listStatefulSets func(context.Context, *kubernetes.Clientset, string) ([]model.StatefulSetStatus, error),
) {
	r.GET("/statefulsets", getStatefulSetList(logger, getK8sClient, listStatefulSets))
	r.GET("/statefulsets/:namespace/:name", getStatefulSetDetail(logger, getK8sClient))
}

// getStatefulSetList 获取 StatefulSet 列表的处理函数
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

// getStatefulSetDetail 获取 StatefulSet 详情的处理函数
func getStatefulSetDetail(
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

		sts, err := clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}

		// 转换为 Unstructured 对象（原始 map 格式）
		objMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(sts)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, objMap, DetailSuccessMessage, nil)
	}
}
