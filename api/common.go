// Package api 提供Kubernetes资源管理的HTTP API接口
package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/model"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

type PaginationParams struct {
	Limit     int
	Offset    int
	Search    string
	Namespace string
}

const (
	SuccessMessage       = "操作成功"
	ListSuccessMessage   = "获取列表成功"
	DetailSuccessMessage = "获取详情成功"
	CreateSuccessMessage = "创建成功"
	UpdateSuccessMessage = "更新成功"
	DeleteSuccessMessage = "删除成功"
	LoginSuccessMessage  = "登录成功"
	LogoutSuccessMessage = "登出成功"
)

var (
	commonLimitValues = map[string]int{
		"20":  20,
		"50":  50,
		"100": 100,
	}
	commonOffsetValues = map[string]int{
		"0":   0,
		"20":  20,
		"40":  40,
		"50":  50,
		"100": 100,
	}
)

func ParsePaginationParams(c *gin.Context) PaginationParams {
	limitStr := c.DefaultQuery("limit", strconv.Itoa(model.DefaultPageSize))
	offsetStr := c.DefaultQuery("offset", strconv.Itoa(model.DefaultPageOffset))

	limit := model.DefaultPageSize
	if cachedLimit, ok := commonLimitValues[limitStr]; ok {
		limit = cachedLimit
	} else if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
		limit = parsedLimit
	}

	offset := model.DefaultPageOffset
	if cachedOffset, ok := commonOffsetValues[offsetStr]; ok {
		offset = cachedOffset
	} else if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
		offset = parsedOffset
	}

	if limit > 1000 {
		limit = 1000
	}

	return PaginationParams{
		Limit:     limit,
		Offset:    offset,
		Search:    strings.TrimSpace(c.DefaultQuery("search", "")),
		Namespace: strings.TrimSpace(c.DefaultQuery("namespace", "")),
	}
}

func HandleWithErrorWrapper[T any](
	c *gin.Context,
	logger *zap.Logger,
	operation func() (T, error),
	successMessage string,
) {
	result, err := operation()
	if err != nil {
		middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
		return
	}
	middleware.ResponseSuccess(c, result, successMessage, nil)
}

func HandleListWithPagination[T SearchableItem](
	c *gin.Context,
	logger *zap.Logger,
	operation func(ctx context.Context, params PaginationParams) ([]T, error),
	successMessage string,
) {
	ctx := GetRequestContext(c)
	params := ParsePaginationParams(c)

	items, err := operation(ctx, params)
	if err != nil {
		middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
		return
	}

	filteredItems := GenericSearchFilter(items, params.Search)
	paged := Paginate(filteredItems, params.Offset, params.Limit)
	middleware.ResponseSuccess(c, paged, successMessage, &model.PageMeta{
		Total:  len(filteredItems),
		Limit:  params.Limit,
		Offset: params.Offset,
	})
}

type K8sClientProvider func() (*kubernetes.Clientset, *versioned.Clientset, error)

// HandleDetailWithK8s 处理K8s资源详情请求
func HandleDetailWithK8s[T any](
	c *gin.Context,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	operation func(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (T, error),
	successMessage string,
) {
	clientset, _, err := getK8sClient()
	if err != nil {
		middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
		return
	}

	ctx := GetRequestContext(c)
	namespace := c.Param("namespace")
	name := c.Param("name")

	result, err := operation(ctx, clientset, namespace, name)
	if err != nil {
		middleware.ResponseError(c, logger, err, http.StatusNotFound)
		return
	}

	middleware.ResponseSuccess(c, result, successMessage, nil)
}

// HandleListWithK8s 处理K8s资源列表请求
func HandleListWithK8s[T SearchableItem](
	c *gin.Context,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	operation func(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]T, error),
	successMessage string,
) {
	ctx := GetRequestContext(c)
	params := ParsePaginationParams(c)

	clientset, _, err := getK8sClient()
	if err != nil {
		middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
		return
	}

	items, err := operation(ctx, clientset, params.Namespace)
	if err != nil {
		middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
		return
	}

	filteredItems := GenericSearchFilter(items, params.Search)
	paged := Paginate(filteredItems, params.Offset, params.Limit)
	middleware.ResponseSuccess(c, paged, successMessage, &model.PageMeta{
		Total:  len(filteredItems),
		Limit:  params.Limit,
		Offset: params.Offset,
	})
}

func GetRequestContext(c *gin.Context) context.Context {
	if ctx := c.Request.Context(); ctx != nil {
		return ctx
	}
	return context.Background()
}

func GetTraceID(c *gin.Context) string {
	tid := c.GetHeader("X-Trace-ID")
	if tid == "" {
		tid = c.GetString("traceId")
		if tid == "" {
			return ""
		}
	}
	return tid
}

func Paginate[T any](list []T, offset, limit int) []T {
	if offset >= len(list) || offset < 0 {
		return []T{}
	}
	if limit <= 0 {
		return []T{}
	}
	
	end := offset + limit
	if end > len(list) {
		end = len(list)
	}
	
	if offset >= end {
		return []T{}
	}
	
	return list[offset:end]
}

type SearchableItem interface {
	GetSearchableFields() map[string]string
}

func GenericSearchFilter[T any](items []T, search string) []T {
	if search == "" {
		return items
	}

	searchLower := strings.ToLower(search)
	var filtered []T

	for _, item := range items {
		if matchesSearchOptimized(item, searchLower) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

func matchesSearchOptimized[T any](item T, searchLower string) bool {
	if searchable, ok := any(item).(SearchableItem); ok {
		fields := searchable.GetSearchableFields()
		for _, value := range fields {
			if strings.Contains(strings.ToLower(value), searchLower) {
				return true
			}
		}
	}
	return false
}
