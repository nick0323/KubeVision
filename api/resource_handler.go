package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/cache"
	"github.com/nick0323/K8sVision/model"
	"github.com/nick0323/K8sVision/pkg/k8s"
	"github.com/nick0323/K8sVision/pkg/util"
	"github.com/nick0323/K8sVision/service"
	"go.uber.org/zap"
)

const cacheVersion = "v3"

func isClusterResource(resourceType string) bool {
	return k8s.ResourceType(strings.ToLower(resourceType)).Normalize().IsClusterScoped()
}

func RegisterRoutes(r *gin.RouterGroup, logger *zap.Logger, getK8sClient K8sClientProvider, cacheMgr *cache.MemoryCache[interface{}]) {
	r.GET("/:resourceType", getResourceList(logger, getK8sClient, cacheMgr))
	r.GET("/:resourceType/:namespace/:name", getResourceDetail(logger, getK8sClient, cacheMgr))
	// 注意: DELETE 操作是危险操作，生产环境建议添加额外的权限验证
	r.DELETE("/:resourceType/:namespace/:name", deleteResource(logger, getK8sClient, cacheMgr))
}

// getResourceList 获取资源列表
func getResourceList(logger *zap.Logger, getK8sClient K8sClientProvider, cacheMgr *cache.MemoryCache[interface{}]) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceType := c.Param("resourceType")
		if resourceType == "" {
			middleware.ResponseError(c, logger, fmt.Errorf("resource type is required"), http.StatusBadRequest)
			return
		}

		cluster := c.Query("cluster")
		clientset, metricsClient, err := getK8sClient(cluster)
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
		ownerUid := c.Query("ownerUid")
		forceRefresh := c.Query("force") == "true"

		// 尝试从缓存获取（除非强制刷新，ownerUid 查询不缓存）
		if cacheMgr != nil && !forceRefresh && ownerUid == "" {
			cacheKey := fmt.Sprintf("list:%s:%s:%s:%s:%s:%s:%s:%s", cacheVersion, cluster, resourceType, namespace, labelSelector, fieldSelector, involvedObject, since)
			if cached, ok := cacheMgr.Get(cacheKey); ok {
				if result, ok := cached.([]model.SearchableItem); ok {
					params := ParsePaginationParams(c)
					filteredItems := util.GenericSearchFilter(result, params.Search)
					if params.SortBy != "" && params.SortOrder != "" {
						filteredItems = util.SortItems(filteredItems, params.SortBy, params.SortOrder)
					}
					paged := util.Paginate(filteredItems, params.Offset, params.Limit)

					middleware.ResponseSuccess(c, paged, "List retrieved successfully (cached)", &model.PageMeta{
						Total:  len(filteredItems),
						Limit:  params.Limit,
						Offset: params.Offset,
					})
					return
				}
			}
		}

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

			// 写入缓存（ownerUid 查询不缓存）
			if cacheMgr != nil && ownerUid == "" {
				cacheKey := fmt.Sprintf("list:%s:%s:%s:%s:%s:%s:%s:%s", cacheVersion, cluster, resourceType, namespace, labelSelector, fieldSelector, involvedObject, since)
				cacheMgr.Set(cacheKey, items)
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

		// 写入缓存（ownerUid 查询不缓存）
		if cacheMgr != nil && ownerUid == "" {
			cacheKey := fmt.Sprintf("list:%s:%s:%s:%s:%s:%s:%s:%s", cacheVersion, cluster, resourceType, namespace, labelSelector, fieldSelector, involvedObject, since)
			cacheMgr.Set(cacheKey, result)
		}

		// 按 owner UID 过滤（用于 Pod 列表精确过滤，避免不同资源相同 label 导致误匹配）
		if ownerUid != "" && strings.ToLower(resourceType) == "pod" {
			pods := make([]model.Pod, 0, len(result))
			for _, item := range result {
				if pod, ok := item.(*model.Pod); ok {
					pods = append(pods, *pod)
				}
			}
			filteredPods := service.FilterPodsByOwner(ctx, clientset, logger, pods, ownerUid, namespace)
			result = make([]model.SearchableItem, len(filteredPods))
			for i := range filteredPods {
				result[i] = &filteredPods[i]
			}
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
	cacheMgr *cache.MemoryCache[interface{}],
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

		cluster := c.Query("cluster")
		clientset, _, err := getK8sClient(cluster)
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

		forceRefresh := c.Query("force") == "true"

		// 尝试从缓存获取（除非强制刷新）
		if cacheMgr != nil && !forceRefresh {
			cacheKey := fmt.Sprintf("detail:%s:%s:%s:%s:%s", cacheVersion, cluster, resourceType, ns, name)
			if cached, ok := cacheMgr.Get(cacheKey); ok {
				middleware.ResponseSuccess(c, cached, "Resource details retrieved successfully (cached)", nil)
				return
			}
		}

		// 调用 Service 层
		obj, err := service.GetResourceByName(ctx, clientset, resourceType, ns, name)
		if err != nil {
			logger.Error("Failed to get resource", zap.Error(err))
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}

		// 写入缓存（2分钟 TTL）
		if cacheMgr != nil {
			cacheKey := fmt.Sprintf("detail:%s:%s:%s:%s:%s", cacheVersion, cluster, resourceType, ns, name)
			cacheMgr.SetWithTTL(cacheKey, obj, 2*time.Minute)
		}

		middleware.ResponseSuccess(c, obj, "Resource details retrieved successfully", nil)
	}
}

// deleteResource 删除资源
func deleteResource(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	cacheMgr *cache.MemoryCache[interface{}],
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

		cluster := c.Query("cluster")
		clientset, _, err := getK8sClient(cluster)
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

		// 清除相关缓存
		if cacheMgr != nil {
			cacheMgr.Delete(fmt.Sprintf("detail:%s:%s:%s:%s:%s", cacheVersion, cluster, resourceType, ns, name))
			cacheMgr.DeleteByPrefix(fmt.Sprintf("list:%s:%s:%s:", cacheVersion, cluster, resourceType))
		}

		middleware.ResponseSuccess(c, nil, "Resource deleted successfully", nil)
	}
}
