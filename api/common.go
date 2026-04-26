package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
		return origin == "" || allowedOrigin(allowedOrigins, origin)
	}
}

func allowedOrigin(allowedOrigins []string, origin string) bool {
	for _, o := range allowedOrigins {
		if o == "*" || o == origin {
			return true
		}
	}
	return false
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

func ExtractTokenFromRequest(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	if token := extractTokenFromWebSocket(c.GetHeader("Sec-WebSocket-Protocol")); token != "" {
		return token
	}
	if token := c.GetHeader("Authorization"); token != "" {
		return strings.TrimPrefix(token, "Bearer ")
	}
	return c.Query("token")
}

func extractTokenFromWebSocket(headerValue string) string {
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
