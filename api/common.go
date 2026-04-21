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
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/model"
	"k8s.io/client-go/kubernetes"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return false },
}

const webSocketAuthProtocol = "k8svision.auth"

func InitWebSocketUpgrader(allowedOrigins []string) {
	if len(allowedOrigins) == 0 {
		return
	}

	upgrader.CheckOrigin = func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		for _, allowed := range allowedOrigins {
			if allowed == "*" || allowed == origin {
				return true
			}
		}
		return false
	}
}

type PaginationParams struct {
	Limit     int
	Offset    int
	Search    string
	Namespace string
	SortBy    string
	SortOrder string
}

type K8sClientProvider func() (*kubernetes.Clientset, *versioned.Clientset, error)

func ParsePaginationParams(c *gin.Context) PaginationParams {
	limit := model.DefaultPageSize
	if l, err := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(limit))); err == nil && l > 0 {
		limit = min(l, model.MaxPageSize)
	}

	offset := 0
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o >= 0 {
		offset = o
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

func GetRequestContext(c *gin.Context) context.Context {
	if ctx := c.Request.Context(); ctx != nil {
		return ctx
	}
	return context.Background()
}

func Paginate[T any](list []T, offset, limit int) []T {
	if offset < 0 || limit <= 0 || offset >= len(list) {
		return []T{}
	}
	if end := offset + limit; end < len(list) {
		return list[offset:end]
	}
	return list[offset:]
}

func GenericSearchFilter[T model.SearchableItem](items []T, search string) []T {
	if search == "" {
		return items
	}
	searchLower := strings.ToLower(search)
	filtered := make([]T, 0, len(items))
	for _, item := range items {
		if matchesSearch(item, searchLower) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func matchesSearch(item model.SearchableItem, searchLower string) bool {
	for _, value := range item.GetSearchableFields() {
		if strings.Contains(strings.ToLower(value), searchLower) {
			return true
		}
	}
	return false
}

func SortItems[T any](items []T, sortBy, sortOrder string) []T {
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

func getFieldValue(item interface{}, fieldName string) string {
	if item == nil {
		return ""
	}

	if m, ok := item.(map[string]interface{}); ok {
		if val, exists := m[fieldName]; exists {
			return fmt.Sprintf("%v", val)
		}
	}

	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return ""
	}

	variations := []string{
		fieldName,
		strings.ToUpper(fieldName[:1]) + fieldName[1:],
		strings.ToLower(fieldName[:1]) + fieldName[1:],
	}

	for _, name := range variations {
		if field := v.FieldByName(name); field.IsValid() && field.CanInterface() {
			return fmt.Sprintf("%v", field.Interface())
		}
	}

	return ""
}

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

func parseAgeToSeconds(ageStr string) int64 {
	if ageStr == "" || ageStr == "-" {
		return 0
	}

	i := 0
	for i < len(ageStr) && ageStr[i] >= '0' && ageStr[i] <= '9' {
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

func ExtractTokenFromRequest(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}

	if token := extractTokenFromWebSocketProtocol(c.GetHeader("Sec-WebSocket-Protocol")); token != "" {
		return token
	}

	if token := c.GetHeader("Authorization"); token != "" {
		return strings.TrimPrefix(token, "Bearer ")
	}

	return c.Query("token")
}

func sanitizeRawQuery(raw string) string {
	return middleware.MaskSensitiveQuery(raw)
}

func extractTokenFromWebSocketProtocol(headerValue string) string {
	if headerValue == "" {
		return ""
	}

	parts := strings.Split(headerValue, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	if len(parts) >= 2 && parts[0] == webSocketAuthProtocol && parts[1] != "" {
		return parts[1]
	}

	return strings.TrimSpace(headerValue)
}

func buildWebSocketUpgradeHeaders(c *gin.Context) http.Header {
	headerValue := c.GetHeader("Sec-WebSocket-Protocol")
	if headerValue == "" {
		return nil
	}

	parts := strings.Split(headerValue, ",")
	for _, part := range parts {
		if strings.TrimSpace(part) == webSocketAuthProtocol {
			return http.Header{"Sec-WebSocket-Protocol": []string{webSocketAuthProtocol}}
		}
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
