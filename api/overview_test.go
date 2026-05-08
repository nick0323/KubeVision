package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRegisterOverview(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	r := gin.New()
	group := r.Group("/api/v1")

	getOverview := func() (*model.OverviewStatus, error) {
		return &model.OverviewStatus{NodeCount: 3, PodCount: 42}, nil
	}

	RegisterOverview(group, logger, getOverview)

	routes := r.Routes()
	assert.GreaterOrEqual(t, len(routes), 1)

	found := false
	for _, route := range routes {
		if route.Method == "GET" && route.Path == "/api/v1/overview" {
			found = true
			break
		}
	}
	assert.True(t, found, "GET /api/v1/overview should be registered")
}

func TestHandleOverview_HappyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	r := gin.New()

	getOverview := func() (*model.OverviewStatus, error) {
		return &model.OverviewStatus{NodeCount: 3, PodCount: 100}, nil
	}

	RegisterOverview(r.Group("/api/v1"), logger, getOverview)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/overview", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "100")
	assert.Contains(t, w.Body.String(), "nodeCount")
}

func TestHandleOverview_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	r := gin.New()

	getOverview := func() (*model.OverviewStatus, error) {
		return nil, errors.New("cluster unavailable")
	}

	RegisterOverview(r.Group("/api/v1"), logger, getOverview)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/overview", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
