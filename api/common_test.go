package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAllowedOrigin(t *testing.T) {
	tests := []struct {
		name          string
		allowedOrigins []string
		origin        string
		expected      bool
	}{
		{"exact match", []string{"http://localhost:8080"}, "http://localhost:8080", true},
		{"no match", []string{"http://localhost:8080"}, "http://evil.com", false},
		{"wildcard *", []string{"*"}, "http://anything.com", true},
		{"empty allowed", []string{}, "http://localhost:8080", false}, // 空列表默认拒绝
		{"multiple origins", []string{"http://a.com", "http://b.com"}, "http://b.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := allowedOrigin(tt.allowedOrigins, tt.origin)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParsePaginationParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		query    map[string]string
		expected PaginationParams
	}{
	{"default values", map[string]string{}, PaginationParams{Offset: 0, Limit: 15, SortBy: "name", SortOrder: "asc"}},
		{"custom values", map[string]string{"offset": "5", "limit": "20"}, PaginationParams{Offset: 5, Limit: 20, SortBy: "name", SortOrder: "asc"}},
		{"invalid offset", map[string]string{"offset": "abc", "limit": "10"}, PaginationParams{Offset: 0, Limit: 10, SortBy: "name", SortOrder: "asc"}},
		{"invalid limit", map[string]string{"offset": "0", "limit": "abc"}, PaginationParams{Offset: 0, Limit: 15, SortBy: "name", SortOrder: "asc"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = &http.Request{}
			c.Request.URL = &url.URL{Path: "/", RawQuery: toQuery(tt.query)}

			result := ParsePaginationParams(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{}
	c.Request.Header = make(http.Header)
	ctx := GetRequestContext(c)
	assert.NotNil(t, ctx)
}

func TestExtractTokenFromRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("from Authorization header", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{Header: make(http.Header)}
		c.Request.Header.Set("Authorization", "Bearer my-token-123")
		assert.Equal(t, "my-token-123", ExtractTokenFromRequest(c))
	})

	t.Run("from query parameter", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{}
		c.Request.URL = &url.URL{RawQuery: "token=query-token"}
		assert.Equal(t, "query-token", ExtractTokenFromRequest(c))
	})

	t.Run("from WebSocket protocol header", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{Header: make(http.Header)}
		c.Request.Header.Set("Sec-WebSocket-Protocol", "k8svision.auth, ws-token-456")
		assert.Equal(t, "ws-token-456", ExtractTokenFromRequest(c))
	})

	t.Run("no token available", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{Header: make(http.Header)}
		c.Request.URL = &url.URL{Path: "/"}
		assert.Equal(t, "", ExtractTokenFromRequest(c))
	})

	t.Run("nil request", func(t *testing.T) {
		assert.Equal(t, "", ExtractTokenFromRequest(nil))
	})
}

func TestExtractTokenFromWebSocket(t *testing.T) {
	t.Run("empty header", func(t *testing.T) {
		assert.Equal(t, "", extractTokenFromWebSocket(""))
	})

	t.Run("valid protocol + token", func(t *testing.T) {
		assert.Equal(t, "my-token", extractTokenFromWebSocket("k8svision.auth, my-token"))
	})

	t.Run("only protocol", func(t *testing.T) {
		assert.Equal(t, "k8svision.auth", extractTokenFromWebSocket("k8svision.auth"))
	})

	t.Run("multiple protocols without auth", func(t *testing.T) {
		assert.Equal(t, "proto1, proto2", extractTokenFromWebSocket("proto1, proto2"))
	})
}

func TestBuildWebSocketUpgradeHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("with auth protocol", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{Header: make(http.Header)}
		c.Request.Header.Set("Sec-WebSocket-Protocol", "k8svision.auth, my-token")
		headers := buildWebSocketUpgradeHeaders(c)
		assert.NotNil(t, headers)
	})

	t.Run("without auth protocol", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{Header: make(http.Header)}
		c.Request.Header.Set("Sec-WebSocket-Protocol", "other-proto")
		assert.Nil(t, buildWebSocketUpgradeHeaders(c))
	})

	t.Run("empty header", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{Header: make(http.Header)}
		assert.Nil(t, buildWebSocketUpgradeHeaders(c))
	})
}

// toQuery helper
func toQuery(m map[string]string) string {
	query := ""
	for k, v := range m {
		if query != "" {
			query += "&"
		}
		query += k + "=" + v
	}
	return query
}
