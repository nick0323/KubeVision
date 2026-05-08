package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware_NilConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Origin", "http://example.com")

	handler := CORSMiddleware(nil)
	handler(c)

	assert.Equal(t, "*", c.Writer.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCORSMiddleware_CredentialsWithWildcard(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Origin", "http://localhost:3000")

	config := &CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowCredentials: true,
	}

	handler := CORSMiddleware(config)
	handler(c)

	assert.Equal(t, "http://localhost:3000", c.Writer.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", c.Writer.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORSMiddleware_SpecificOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Origin", "https://myapp.com")

	config := &CORSConfig{
		AllowOrigins:     []string{"https://myapp.com"},
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Authorization"},
		ExposeHeaders:    []string{"X-Custom"},
		MaxAge:           600,
	}

	handler := CORSMiddleware(config)
	handler(c)

	assert.Equal(t, "https://myapp.com", c.Writer.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", c.Writer.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "GET, POST", c.Writer.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Authorization", c.Writer.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "X-Custom", c.Writer.Header().Get("Access-Control-Expose-Headers"))
	assert.Equal(t, "600", c.Writer.Header().Get("Access-Control-Max-Age"))
}

func TestCORSMiddleware_UnallowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Origin", "https://evil.com")

	config := &CORSConfig{
		AllowOrigins: []string{"https://allowed.com"},
	}

	handler := CORSMiddleware(config)
	handler(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCORSMiddleware_OptionsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodOptions, "/", nil)
	c.Request.Header.Set("Origin", "http://example.com")

	handler := CORSMiddleware(nil)
	handler(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestCORSMiddleware_MatchingOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Origin", "https://mysite.com")

	config := &CORSConfig{
		AllowOrigins: []string{"https://mysite.com", "https://backup.com"},
	}

	handler := CORSMiddleware(config)
	handler(c)

	assert.Equal(t, "https://mysite.com", c.Writer.Header().Get("Access-Control-Allow-Origin"))
}
