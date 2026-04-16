package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/model"
	"github.com/nick0323/K8sVision/service"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

// ResourceRegistry 全局资源注册表（由 main.go 初始化）
var globalResourceRegistry *ResourceRegistry

// SetResourceRegistry 设置全局资源注册表（在 main.go 初始化时调用）
func SetResourceRegistry(registry *ResourceRegistry) {
	globalResourceRegistry = registry
}

// getResourceRegistry 获取资源注册表
func getResourceRegistry() *ResourceRegistry {
	return globalResourceRegistry
}

// RegisterRoutes 注册通用资源接口
func RegisterRoutes(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) {
	// 初始化资源注册表
	registry := NewResourceRegistry(logger)
	SetResourceRegistry(registry)

	r.GET("/:resourceType", getResourceList(logger, getK8sClient))
	// 资源详情：/api/:resourceType/:namespace/:name
	// 对于集群资源（node, pv, storageclass, namespace），:namespace 参数会被忽略
	r.GET("/:resourceType/:namespace/:name", getResourceDetail(logger, getK8sClient))
	r.DELETE("/:resourceType/:namespace/:name", deleteResource(logger, getK8sClient))
}

// getResourceList 获取资源列表
func getResourceList(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceType := c.Param("resourceType")
		if resourceType == "" {
			middleware.ResponseError(c, logger, fmt.Errorf("resource type is required"), http.StatusBadRequest)
			return
		}

		// 从注册表获取处理器
		registry := getResourceRegistry()
		handler, exists := registry.GetHandler(resourceType)
		if !exists {
			middleware.ResponseError(c, logger, fmt.Errorf("unsupported resource type: %s. Supported types: %v",
				resourceType, registry.GetSupportedResourceTypes()), http.StatusBadRequest)
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

		// 集群资源忽略 namespace
		if handler.IsClusterScoped() {
			namespace = ""
		}

		// 特殊处理 Node 资源（需要 metrics）
		var result []model.SearchableItem
		if strings.ToLower(resourceType) == "node" && metricsClient != nil {
			result, err = getNodeListWithMetrics(ctx, clientset, metricsClient, labelSelector, fieldSelector)
		} else {
			result, err = handler.List(ctx, clientset, namespace, labelSelector, fieldSelector)
		}

		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		// 分页和搜索
		params := ParsePaginationParams(c)
		filteredItems := GenericSearchFilter(result, params.Search)
		if params.SortBy != "" && params.SortOrder != "" {
			filteredItems = SortItems(filteredItems, params.SortBy, params.SortOrder)
		}
		paged := Paginate(filteredItems, params.Offset, params.Limit)

		middleware.ResponseSuccess(c, paged, "List retrieved successfully", &model.PageMeta{
			Total:  len(filteredItems),
			Limit:  params.Limit,
			Offset: params.Offset,
		})
	}
}

// getNodeListWithMetrics 获取带 metrics 的 Node 列表
func getNodeListWithMetrics(ctx context.Context, clientset *kubernetes.Clientset, metricsClient *versioned.Clientset, labelSelector, fieldSelector string) ([]model.SearchableItem, error) {
	nodeMetricsMap, _ := GetNodeMetrics(ctx, metricsClient)
	nodes, err := service.ListNodes(ctx, clientset, nil, nodeMetricsMap, labelSelector, fieldSelector)
	if err != nil {
		return nil, err
	}
	result := make([]model.SearchableItem, len(nodes))
	for i := range nodes {
		result[i] = &nodes[i]
	}
	return result, nil
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

		// 从注册表获取处理器
		registry := getResourceRegistry()
		handler, exists := registry.GetHandler(resourceType)
		if !exists {
			middleware.ResponseError(c, logger, fmt.Errorf("unsupported resource type: %s. Supported types: %v",
				resourceType, registry.GetSupportedResourceTypes()), http.StatusBadRequest)
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

		// 集群资源忽略 namespace
		ns := namespace
		if handler.IsClusterScoped() {
			ns = ""
		}

		obj, err := handler.Get(ctx, clientset, ns, name)
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

		// 从注册表获取处理器
		registry := getResourceRegistry()
		handler, exists := registry.GetHandler(resourceType)
		if !exists {
			middleware.ResponseError(c, logger, fmt.Errorf("unsupported resource type: %s. Supported types: %v",
				resourceType, registry.GetSupportedResourceTypes()), http.StatusBadRequest)
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

		// 集群资源忽略 namespace
		ns := namespace
		if handler.IsClusterScoped() {
			ns = ""
		}

		err = handler.Delete(ctx, clientset, ns, name)
		if err != nil {
			logger.Error("Failed to delete resource", zap.Error(err))
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, nil, "Resource deleted successfully", nil)
	}
}
