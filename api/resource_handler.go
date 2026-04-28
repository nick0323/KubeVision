package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/model"
	"github.com/nick0323/K8sVision/pkg/util"
	"github.com/nick0323/K8sVision/service"
	"go.uber.org/zap"
)

func isClusterResource(resourceType string) bool {
	clusterResources := map[string]bool{
		"node": true, "pv": true, "persistentvolume": true,
		"storageclass": true, "namespace": true,
	}
	return clusterResources[strings.ToLower(resourceType)]
}

func RegisterRoutes(r *gin.RouterGroup, logger *zap.Logger, getK8sClient K8sClientProvider) {
	r.GET("/:resourceType", getResourceList(logger, getK8sClient))
	r.GET("/:resourceType/:namespace/:name", getResourceDetail(logger, getK8sClient))
	// 注意: DELETE 操作是危险操作，生产环境建议添加额外的权限验证
	r.DELETE("/:resourceType/:namespace/:name", deleteResource(logger, getK8sClient))
}

// getResourceList 获取资源列表
func getResourceList(logger *zap.Logger, getK8sClient K8sClientProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceType := c.Param("resourceType")
		if resourceType == "" {
			middleware.ResponseError(c, logger, fmt.Errorf("resource type is required"), http.StatusBadRequest)
			return
		}

		clientset, metricsClient, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}
		if clientset == nil {
			middleware.ResponseError(c, logger, fmt.Errorf("kubernetes client unavailable"), http.StatusServiceUnavailable)
			return
		}

		ctx := GetRequestContext(c)
		namespace := c.Query("namespace")
		labelSelector := c.Query("labelSelector")
		if labelSelector == "" {
			labelSelector = c.Query("selector")
		}
		fieldSelector := c.Query("fieldSelector")
		involvedObject := c.Query("involvedObject")
		since := c.Query("since")

		// 特殊处理 Node 类型，传入 metricsClient
		if strings.ToLower(resourceType) == "node" {
			result, err := service.ListNodes(ctx, clientset, metricsClient, nil, labelSelector, fieldSelector)
			if err != nil {
				middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
				return
			}
			
			// 转换为 []model.SearchableItem
			items := make([]model.SearchableItem, len(result))
			for i := range result {
				items[i] = &result[i]
			}
			
			params := ParsePaginationParams(c)
			filteredItems := util.GenericSearchFilter(items, params.Search)
			if params.SortBy != "" && params.SortOrder != "" {
				filteredItems = util.SortItems(filteredItems, params.SortBy, params.SortOrder)
			}
			paged := util.Paginate(filteredItems, params.Offset, params.Limit)
			
			middleware.ResponseSuccess(c, paged, "List retrieved successfully", &model.PageMeta{
				Total:  len(filteredItems),
				Limit:  params.Limit,
				Offset: params.Offset,
			})
			return
		}

		// 其他资源类型使用通用接口
		result, err := service.ListResourcesByType(ctx, clientset, resourceType, namespace, labelSelector, fieldSelector, involvedObject, since)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		params := ParsePaginationParams(c)
		filteredItems := util.GenericSearchFilter(result, params.Search)
		if params.SortBy != "" && params.SortOrder != "" {
			filteredItems = util.SortItems(filteredItems, params.SortBy, params.SortOrder)
		}
		paged := util.Paginate(filteredItems, params.Offset, params.Limit)

		middleware.ResponseSuccess(c, paged, "List retrieved successfully", &model.PageMeta{
			Total:  len(filteredItems),
			Limit:  params.Limit,
			Offset: params.Offset,
		})
	}
}

// getResourceDetail 获取资源详情
func getResourceDetail(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceType := c.Param("resourceType")
		namespace := c.Param("namespace")
		name := c.Param("name")

		logger.Info("Get resource details",
			zap.String("resourceType", resourceType),
			zap.String("namespace", namespace),
			zap.String("name", name),
		)

		if resourceType == "" {
			middleware.ResponseError(c, logger, fmt.Errorf("resource type is required"), http.StatusBadRequest)
			return
		}
		if name == "" {
			middleware.ResponseError(c, logger, fmt.Errorf("name is required"), http.StatusBadRequest)
			return
		}

		clientset, _, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}
		if clientset == nil {
			middleware.ResponseError(c, logger, fmt.Errorf("kubernetes client unavailable"), http.StatusServiceUnavailable)
			return
		}

		ctx := GetRequestContext(c)

		// 判断是否为集群资源（不需要 namespace）
		ns := namespace
		if isClusterResource(resourceType) {
			ns = ""
		}

		// 调用 Service 层
		obj, err := service.GetResourceByName(ctx, clientset, resourceType, ns, name)
		if err != nil {
			logger.Error("Failed to get resource", zap.Error(err))
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}

		middleware.ResponseSuccess(c, obj, "Resource details retrieved successfully", nil)
	}
}

// deleteResource 删除资源
func deleteResource(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceType := c.Param("resourceType")
		namespace := c.Param("namespace")
		name := c.Param("name")

		if resourceType == "" {
			middleware.ResponseError(c, logger, fmt.Errorf("resource type is required"), http.StatusBadRequest)
			return
		}
		if name == "" {
			middleware.ResponseError(c, logger, fmt.Errorf("name is required"), http.StatusBadRequest)
			return
		}

		clientset, _, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}
		if clientset == nil {
			middleware.ResponseError(c, logger, fmt.Errorf("kubernetes client unavailable"), http.StatusServiceUnavailable)
			return
		}

		ctx := GetRequestContext(c)

		// 判断是否为集群资源（不需要 namespace）
		ns := namespace
		if isClusterResource(resourceType) {
			ns = ""
		}

		// 调用 Service 层
		err = service.DeleteResourceByType(ctx, clientset, resourceType, ns, name)
		if err != nil {
			logger.Error("Failed to delete resource", zap.Error(err))
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, nil, "Resource deleted successfully", nil)
	}
}
