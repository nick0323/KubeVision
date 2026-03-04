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

// RegisterDaemonSet 注册 DaemonSet 相关路由
func RegisterDaemonSet(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listDaemonSets func(context.Context, *kubernetes.Clientset, string) ([]model.DaemonSetStatus, error),
) {
	r.GET("/daemonsets", getDaemonSetList(logger, getK8sClient, listDaemonSets))
	r.GET("/daemonsets/:namespace/:name", getDaemonSetDetail(logger, getK8sClient))
}

// getDaemonSetList 获取 DaemonSet 列表的处理函数
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

// getDaemonSetDetail 获取 DaemonSet 详情的处理函数
func getDaemonSetDetail(
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
		
		ds, err := clientset.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}

		// 转换为 Unstructured 对象（原始 map 格式）
		objMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(ds)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, objMap, DetailSuccessMessage, nil)
	}
}
