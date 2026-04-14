// Package api 提供 Kubernetes 资源管理的 HTTP API 接口
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
	"github.com/gorilla/websocket"
	"github.com/nick0323/K8sVision/model"
	"k8s.io/client-go/kubernetes"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

// WebSocket upgrader 统一配置
// 注意：默认拒绝所有源，必须通过 InitWebSocketUpgrader 显式配置允许的源
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return false // 默认拒绝所有源（生产环境安全配置）
	},
}

// InitWebSocketUpgrader 初始化 WebSocket upgrader，配置允许的源
func InitWebSocketUpgrader(allowedOrigins []string) {
	// 如果没有配置允许的源，使用默认策略（拒绝所有）
	if len(allowedOrigins) == 0 {
		return
	}

	upgrader.CheckOrigin = func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // 允许同源请求
		}
		for _, allowed := range allowedOrigins {
			if allowed == "*" || allowed == origin {
				return true
			}
		}
		return false
	}
}

// PaginationParams 分页参数
type PaginationParams struct {
	Limit     int
	Offset    int
	Search    string
	Namespace string
	SortBy    string
	SortOrder string
}

// K8sClientProvider K8s 客户端提供者函数类型
type K8sClientProvider func() (*kubernetes.Clientset, *versioned.Clientset, error)

// SearchableItem 可搜索接口
type SearchableItem interface {
	GetSearchableFields() map[string]string
}

// ParsePaginationParams 解析分页参数
func ParsePaginationParams(c *gin.Context) PaginationParams {
	limitStr := c.DefaultQuery("limit", strconv.Itoa(model.DefaultPageSize))
	offsetStr := c.DefaultQuery("offset", strconv.Itoa(model.DefaultPageOffset))

	limit := model.DefaultPageSize
	if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
		limit = parsedLimit
	}

	offset := model.DefaultPageOffset
	if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
		offset = parsedOffset
	}

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

// GetRequestContext 获取请求上下文
func GetRequestContext(c *gin.Context) context.Context {
	if ctx := c.Request.Context(); ctx != nil {
		return ctx
	}
	return context.Background()
}

// GetTraceID 获取追踪 ID
func GetTraceID(c *gin.Context) string {
	tid := c.GetHeader("X-Trace-ID")
	if tid == "" {
		tid = c.GetString("traceId")
	}
	return tid
}

// Paginate 分页
func Paginate[T any](list []T, offset, limit int) []T {
	if offset < 0 || limit <= 0 || offset >= len(list) {
		return []T{}
	}

	end := offset + limit
	if end > len(list) {
		end = len(list)
	}

	return list[offset:end]
}

// GenericSearchFilter 通用搜索过滤
func GenericSearchFilter[T SearchableItem](items []T, search string) []T {
	if search == "" {
		return items
	}

	searchLower := strings.ToLower(search)
	var filtered []T

	for _, item := range items {
		if matchesSearch(item, searchLower) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

func matchesSearch(item SearchableItem, searchLower string) bool {
	fields := item.GetSearchableFields()
	for _, value := range fields {
		if strings.Contains(strings.ToLower(value), searchLower) {
			return true
		}
	}
	return false
}

// SortItems 排序
func SortItems[T any](items []T, sortBy string, sortOrder string) []T {
	sorted := make([]T, len(items))
	copy(sorted, items)

	sort.Slice(sorted, func(i, j int) bool {
		valI := getFieldValue(sorted[i], sortBy)
		valJ := getFieldValue(sorted[j], sortBy)
		compare := compareValues(valI, valJ, sortBy)
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

	// 处理 map 类型
	if m, ok := item.(map[string]interface{}); ok {
		if val, exists := m[fieldName]; exists {
			return fmt.Sprintf("%v", val)
		}
	}

	// 处理结构体类型（使用反射）
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return ""
	}

	// 尝试多种字段名变体（兼容不同命名风格）
	fieldNames := []string{
		fieldName, // 原始名称
		strings.ToUpper(fieldName[:1]) + fieldName[1:], // 首字母大写
		strings.ToLower(fieldName[:1]) + fieldName[1:], // 首字母小写
	}

	for _, name := range fieldNames {
		field := v.FieldByName(name)
		if field.IsValid() && field.CanInterface() {
			return fmt.Sprintf("%v", field.Interface())
		}
	}

	return ""
}

// compareValues 比较两个字段的值
func compareValues(valA, valB, sortBy string) int {
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

	return strings.Compare(valA, valB)
}

// parseAgeToSeconds 将年龄字符串转换为秒数
func parseAgeToSeconds(ageStr string) int64 {
	if ageStr == "" || ageStr == "-" {
		return 0
	}

	i := 0
	for i < len(ageStr) && (ageStr[i] >= '0' && ageStr[i] <= '9') {
		i++
	}

	if i == 0 {
		return 0
	}

	num, _ := strconv.ParseInt(ageStr[:i], 10, 64)
	unit := ageStr[i:]

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
		return num
	}
}

// ExtractTokenFromRequest 从请求中提取 JWT token
// 优先级：Sec-WebSocket-Protocol header > Authorization header > query parameter
func ExtractTokenFromRequest(c *gin.Context) string {
	// 优先从 Sec-WebSocket-Protocol header 获取（安全方式）
	tokenStr := c.GetHeader("Sec-WebSocket-Protocol")
	if tokenStr == "" {
		// Fallback: 尝试从 Authorization header 获取
		tokenStr = c.GetHeader("Authorization")
		if tokenStr != "" {
			tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
		}
	}
	// 最后尝试从 URL 参数获取（兼容性考虑）
	if tokenStr == "" {
		tokenStr = c.Query("token")
	}
	return tokenStr
}
