package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func RegisterNamespace(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listNamespaces func(context.Context, *kubernetes.Clientset) ([]model.NamespaceDetail, error),
) {
	r.GET("/namespaces", getNamespaceList(logger, getK8sClient, listNamespaces))
	r.GET("/namespaces/:name", getNamespaceDetail(logger, getK8sClient))
}

func getNamespaceList(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listNamespaces func(context.Context, *kubernetes.Clientset) ([]model.NamespaceDetail, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientset, _, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}
		ctx := GetRequestContext(c)

		// 获取所有 namespaces
		nsList, err := listNamespaces(ctx, clientset)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		// 搜索过滤（如果提供了 search 参数）
		search := c.Query("search")
		if search != "" {
			filtered := make([]model.NamespaceDetail, 0)
			for _, ns := range nsList {
				if strings.Contains(strings.ToLower(ns.Name), strings.ToLower(search)) {
					filtered = append(filtered, ns)
				}
			}
			nsList = filtered
		}

		// 排序（如果提供了 sortBy 和 sortOrder 参数）
		sortBy := c.Query("sortBy")
		sortOrder := c.Query("sortOrder")
		if sortBy != "" && sortOrder != "" {
			nsList = SortItems(nsList, sortBy, sortOrder)
		}

		// 返回所有数据，不分页
		middleware.ResponseSuccess(c, nsList, ListSuccessMessage, &model.PageMeta{
			Total:  len(nsList),
			Limit:  len(nsList),
			Offset: 0,
		})
	}
}

func getNamespaceDetail(
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
		ns, err := clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}

		namespaceDetail := model.NamespaceDetail{
			Name:   ns.Name,
			Status: string(ns.Status.Phase),
			Labels: ns.Labels,
		}
		middleware.ResponseSuccess(c, namespaceDetail, DetailSuccessMessage, nil)
	}
}
