package middleware

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMaskSensitiveQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{"empty query", "", ""},
		{"no sensitive params", "name=foo&page=1", "name=foo&page=1"},
		{"token param masked", "token=abc123", "token=%2A%2A%2A"},
		{"password param masked", "password=secret&user=admin", "password=%2A%2A%2A&user=admin"},
		{"auth param masked", "auth=bearer", "auth=%2A%2A%2A"},
		{"multiple sensitive params", "token=x&password=y&name=test", "name=test&password=%2A%2A%2A&token=%2A%2A%2A"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskSensitiveQuery(tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateTraceID(t *testing.T) {
	id := generateTraceID()

	assert.Len(t, id, 31)

	parts := strings.Split(id, "-")
	assert.Len(t, parts, 2)

	matched, _ := regexp.MatchString(`^\d{14}$`, parts[0])
	assert.True(t, matched, "timestamp part should be 14 digits")

	matched, _ = regexp.MatchString(`^[0-9a-f]{16}$`, parts[1])
	assert.True(t, matched, "random part should be 16 hex chars")
}

func TestTraceMiddleware_WithHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("X-Trace-ID", "custom-trace-123")

	handler := TraceMiddleware()
	handler(c)

	traceID, exists := c.Get("traceId")
	assert.True(t, exists)
	assert.Equal(t, "custom-trace-123", traceID)
	assert.Equal(t, "custom-trace-123", w.Header().Get("X-Trace-ID"))
}

func TestTraceMiddleware_WithoutHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	handler := TraceMiddleware()
	handler(c)

	traceID, exists := c.Get("traceId")
	assert.True(t, exists)
	assert.NotEmpty(t, traceID)
	assert.NotEqual(t, "custom-trace-123", traceID)
	assert.Equal(t, traceID, w.Header().Get("X-Trace-ID"))
}

func TestLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Set("traceId", "test-trace")

	handler := LoggingMiddleware(logger)
	handler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLoggingMiddleware_WithSensitiveQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/login?password=secret", nil)
	c.Set("traceId", "test-trace")

	handler := LoggingMiddleware(logger)
	handler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}
