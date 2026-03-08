// Package api 提供Kubernetes资源管理的HTTP API接口
package api

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/model"
	"github.com/nick0323/K8sVision/service"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

type PaginationParams struct {
	Limit     int
	Offset    int
	Search    string
	Namespace string
	SortBy    string
	SortOrder string
}

// 成功消息常量
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

func ParsePaginationParams(c *gin.Context) PaginationParams {
	limitStr := c.DefaultQuery("limit", strconv.Itoa(model.DefaultPageSize))
	offsetStr := c.DefaultQuery("offset", strconv.Itoa(model.DefaultPageOffset))

	// 解析分页参数
	limit := model.DefaultPageSize
	if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
		limit = parsedLimit
	}

	offset := model.DefaultPageOffset
	if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
		offset = parsedOffset
	}

	// 限制最大分页大小
	if limit > model.MaxPageSize {
		limit = model.MaxPageSize
	}

	return PaginationParams{
		Limit:     limit,
		Offset:    offset,
		Search:    strings.TrimSpace(c.DefaultQuery("search", "")),
		Namespace: strings.TrimSpace(c.DefaultQuery("namespace", "")),
		SortBy:    strings.TrimSpace(c.DefaultQuery("sortBy", "name")),
		SortOrder: strings.TrimSpace(c.DefaultQuery("sortOrder", "asc")),
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

	// 尝试使用缓存排序服务
	sortService := service.GetCachedSortService()
	if sortService != nil {
		// 从缓存排序服务获取数据
		// 注意：这里需要泛型转换，暂时使用原有逻辑
		items, err := operation(ctx, params)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		// 搜索过滤
		filteredItems := GenericSearchFilter(items, params.Search)

		// 排序（在分页前）
		if params.SortBy != "" && params.SortOrder != "" {
			filteredItems = SortItems(filteredItems, params.SortBy, params.SortOrder)
		}

		// 分页
		paged := Paginate(filteredItems, params.Offset, params.Limit)
		middleware.ResponseSuccess(c, paged, successMessage, &model.PageMeta{
			Total:  len(filteredItems),
			Limit:  params.Limit,
			Offset: params.Offset,
		})
		return
	}

	// 降级到原有逻辑
	items, err := operation(ctx, params)
	if err != nil {
		middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
		return
	}

	// 搜索过滤
	filteredItems := GenericSearchFilter(items, params.Search)

	// 排序（在分页前）
	if params.SortBy != "" && params.SortOrder != "" {
		filteredItems = SortItems(filteredItems, params.SortBy, params.SortOrder)
	}

	// 分页
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

// SortItems 对 items 进行排序
func SortItems[T any](items []T, sortBy string, sortOrder string) []T {
	// 创建副本，避免修改原数组
	sorted := make([]T, len(items))
	copy(sorted, items)

	// 使用 Go 标准库排序（快速排序）
	sort.Slice(sorted, func(i, j int) bool {
		// 获取字段的字符串值进行比较
		valI := getFieldValue(sorted[i], sortBy)
		valJ := getFieldValue(sorted[j], sortBy)

		// 使用 compareValues 比较（支持 Age 字段特殊处理）
		compare := compareValues(valI, valJ, sortBy)

		// 如果是降序，反转比较结果
		if sortOrder == "desc" {
			compare = -compare
		}

		return compare < 0
	})

	return sorted
}

// getFieldValue 获取结构体字段的字符串值
func getFieldValue(item interface{}, fieldName string) string {
	if item == nil {
		return ""
	}

	// 尝试作为 map 处理（如果 API 返回的是 map）
	if m, ok := item.(map[string]interface{}); ok {
		if val, exists := m[fieldName]; exists {
			return fmt.Sprintf("%v", val)
		}
	}

	// 使用反射获取结构体字段值
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		// 1. 尝试原始字段名（如 "Name"）
		field := v.FieldByName(fieldName)
		if field.IsValid() && field.CanInterface() {
			return fmt.Sprintf("%v", field.Interface())
		}
		
		// 2. 尝试首字母大写（如 "name" → "Name"）
		fieldNameCap := strings.ToUpper(fieldName[:1]) + fieldName[1:]
		field = v.FieldByName(fieldNameCap)
		if field.IsValid() && field.CanInterface() {
			return fmt.Sprintf("%v", field.Interface())
		}
		
		// 3. 尝试首字母小写（如 "Name" → "name"）
		fieldNameLower := strings.ToLower(fieldName[:1]) + fieldName[1:]
		field = v.FieldByName(fieldNameLower)
		if field.IsValid() && field.CanInterface() {
			return fmt.Sprintf("%v", field.Interface())
		}
	}

	return ""
}

// compareValues 比较两个字段的值
func compareValues(valA, valB, sortBy string) int {
	// 特殊处理 Age 字段（时间格式：15s, 2m, 3h, 4d）
	if sortBy == "age" || sortBy == "Age" {
		timeA := parseAgeToSeconds(valA)
		timeB := parseAgeToSeconds(valB)
		if timeA < timeB {
			return -1
		} else if timeA > timeB {
			return 1
		}
		return 0
	}
	
	// 默认字符串比较
	return strings.Compare(valA, valB)
}

// parseAgeToSeconds 将年龄字符串转换为秒数
// 支持格式：15s, 2m, 3h, 4d
func parseAgeToSeconds(ageStr string) int64 {
	if ageStr == "" || ageStr == "-" {
		return 0
	}
	
	// 提取数字和单位
	var num int64
	var unit string
	
	// 解析数字部分
	i := 0
	for i < len(ageStr) && (ageStr[i] >= '0' && ageStr[i] <= '9') {
		i++
	}
	
	if i == 0 {
		return 0
	}
	
	num, _ = strconv.ParseInt(ageStr[:i], 10, 64)
	unit = ageStr[i:]
	
	// 根据单位转换为秒
	switch unit {
	case "s":
		return num
	case "m":
		return num * 60
	case "h":
		return num * 3600
	case "d":
		return num * 86400
	default:
		return num // 默认为秒
	}
}
