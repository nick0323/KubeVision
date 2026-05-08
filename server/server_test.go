package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/cache"
	"github.com/nick0323/K8sVision/config"
	"github.com/nick0323/K8sVision/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func newTestServer() *Server {
	logger, _ := zap.NewDevelopment()
	mgr := config.NewManager(logger)
	mgr.GetConfig().JWT.Secret = "test-jwt-secret-that-is-long-enough-for-unit-tests"
	mgr.GetConfig().Auth.Password = "$2a$10$hashedpassword"
	cacheMgr := 	cache.NewMemoryCache(nil, nil)
	return &Server{
		logger:      logger,
		configMgr:   mgr,
		lruCacheMgr: cacheMgr,
	}
}

func TestNewServer(t *testing.T) {
	s := newTestServer()
	assert.NotNil(t, s)
}

func TestServerSetupRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := newTestServer()
	s.configMgr.GetConfig().Log.Level = "debug"
	router := s.SetupRouter()
	assert.NotNil(t, router)
}

func TestServerSetupRouterReleaseMode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := newTestServer()
	s.configMgr.GetConfig().Log.Level = "info"
	router := s.SetupRouter()
	assert.NotNil(t, router)
}

func TestHealthCheckHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := newTestServer()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

	s.healthCheckHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, model.Version, response["version"])
	assert.NotNil(t, response["timestamp"])
}

func TestServerShutdown(t *testing.T) {
	s := newTestServer()
	err := s.Shutdown(nil)
	assert.NoError(t, err)
}

func TestServerGetServerAddress(t *testing.T) {
	s := newTestServer()
	s.configMgr.GetConfig().Server.Port = "9090"
	s.configMgr.GetConfig().Server.Host = "127.0.0.1"
	assert.Equal(t, "127.0.0.1:9090", s.configMgr.GetConfig().GetServerAddress())
}

func TestServerRun(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := newTestServer()
	s.configMgr.GetConfig().Server.Port = "0"
	err := s.Run()
	assert.NoError(t, err)
	if s.httpServer != nil {
		s.httpServer.Close()
	}
}

func TestServerRegisterMiddlewares(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := newTestServer()
	r := gin.New()
	cfg := model.DefaultConfig()
	s.registerMiddlewares(r, cfg)
	assert.Len(t, r.Routes(), 0)
}
