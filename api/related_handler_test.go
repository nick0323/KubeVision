package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

func mockK8sClientSuccess() (*kubernetes.Clientset, *versioned.Clientset, error) {
	return nil, nil, fmt.Errorf("no kubeconfig")
}

func TestRegisterRelatedRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	r := gin.New()
	group := r.Group("/api/v1")
	RegisterRelatedRoutes(group, logger, mockK8sClientError)

	routes := r.Routes()
	assert.GreaterOrEqual(t, len(routes), 2)

	paths := make(map[string]string)
	for _, route := range routes {
		paths[route.Method+" "+route.Path] = route.Handler
	}

	basePath := "/api/v1"
	assert.Contains(t, paths, "GET "+basePath+"/:resourceType/:namespace/:name/related")
	assert.Contains(t, paths, "GET "+basePath+"/:resourceType/_cluster_/:name/related")
}

func TestGetResourceRelated_MissingParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	tests := []struct {
		name       string
		params     []gin.Param
		statusCode int
	}{
		{
			name:       "empty resource type",
			params:     []gin.Param{},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "invalid namespace",
			params: []gin.Param{
				{Key: "resourceType", Value: "pods"},
				{Key: "namespace", Value: "InvalidNS"},
				{Key: "name", Value: "my-pod"},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "invalid name",
			params: []gin.Param{
				{Key: "resourceType", Value: "pods"},
				{Key: "namespace", Value: "default"},
				{Key: "name", Value: "-invalid"},
			},
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			c.Params = tt.params

			handler := getResourceRelated(logger, mockK8sClientError)
			handler(c)
			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

func TestGetResourceRelated_ClientError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = []gin.Param{
		{Key: "resourceType", Value: "pods"},
		{Key: "namespace", Value: "default"},
		{Key: "name", Value: "my-pod"},
	}

	handler := getResourceRelated(logger, mockK8sClientError)
	handler(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetResourceRelatedCluster_EmptyType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = []gin.Param{}

	handler := getResourceRelatedCluster(logger, mockK8sClientError)
	handler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetResourceRelatedCluster_ClientError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = []gin.Param{
		{Key: "resourceType", Value: "node"},
		{Key: "name", Value: "worker-1"},
	}

	handler := getResourceRelatedCluster(logger, mockK8sClientError)
	handler(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
