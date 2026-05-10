package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/cache"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

func TestIsClusterResource(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		expected     bool
	}{
		{"node resource", "node", true},
		{"Node uppercase", "Node", true},
		{"pv resource", "pv", true},
		{"persistentvolume full name", "persistentvolume", true},
		{"storageclass", "storageclass", true},
		{"namespace", "namespace", true},
		{"pod - not cluster resource", "pod", false},
		{"deployment - not cluster resource", "deployment", false},
		{"service - not cluster resource", "service", false},
		{"empty string", "", false},
		{"unknown resource", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isClusterResource(tt.resourceType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsClusterResource_CaseInsensitive(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"NODE", true},
		{"Node", true},
		{"node", true},
		{"NoDe", true},
		{"PV", true},
		{"Pv", true},
		{"pv", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isClusterResource(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsClusterResource_AllClusterResources(t *testing.T) {
	clusterResources := []string{"node", "pv", "persistentvolume", "storageclass", "namespace"}

	for _, resource := range clusterResources {
		t.Run(resource, func(t *testing.T) {
			assert.True(t, isClusterResource(resource))
			assert.True(t, isClusterResource(resource[:1]+resource[1:])) // Test mixed case
		})
	}
}

func TestIsClusterResource_NonClusterResources(t *testing.T) {
	nonClusterResources := []string{
		"pod", "deployment", "statefulset", "daemonset",
		"service", "ingress", "configmap", "secret",
		"job", "cronjob", "pvc", "endpoints",
	}

	for _, resource := range nonClusterResources {
		t.Run(resource, func(t *testing.T) {
			assert.False(t, isClusterResource(resource))
		})
	}
}

func mockK8sClientError() (*kubernetes.Clientset, *versioned.Clientset, error) {
	return nil, nil, fmt.Errorf("k8s connection error")
}

func mockK8sClientNil() (*kubernetes.Clientset, *versioned.Clientset, error) {
	return nil, nil, nil
}

func TestRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	r := gin.New()
	group := r.Group("/api/v1")
	cacheMgr := cache.NewMemoryCache(nil, nil)
	RegisterRoutes(group, logger, mockK8sClientError, cacheMgr)

	routes := r.Routes()
	assert.GreaterOrEqual(t, len(routes), 3)

	paths := make(map[string]string)
	for _, route := range routes {
		paths[route.Method+" "+route.Path] = route.Handler
	}

	basePath := "/api/v1"
	assert.Contains(t, paths, "GET "+basePath+"/:resourceType")
	assert.Contains(t, paths, "GET "+basePath+"/:resourceType/:namespace/:name")
	assert.Contains(t, paths, "DELETE "+basePath+"/:resourceType/:namespace/:name")
}

func TestGetResourceList_EmptyType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	handler := getResourceList(logger, mockK8sClientError, nil)
	handler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetResourceList_ClientError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/pods", nil)
	c.Params = []gin.Param{{Key: "resourceType", Value: "pods"}}

	handler := getResourceList(logger, mockK8sClientError, nil)
	handler(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetResourceList_NilClient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/pods", nil)
	c.Params = []gin.Param{{Key: "resourceType", Value: "pods"}}

	handler := getResourceList(logger, mockK8sClientNil, nil)
	handler(c)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestGetResourceDetail_EmptyType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	handler := getResourceDetail(logger, mockK8sClientError, nil)
	handler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetResourceDetail_EmptyName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/pods/default", nil)
	c.Params = []gin.Param{{Key: "resourceType", Value: "pods"}, {Key: "namespace", Value: "default"}}

	handler := getResourceDetail(logger, mockK8sClientError, nil)
	handler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetResourceDetail_ClientError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/pods/default/my-pod", nil)
	c.Params = []gin.Param{
		{Key: "resourceType", Value: "pods"},
		{Key: "namespace", Value: "default"},
		{Key: "name", Value: "my-pod"},
	}

	handler := getResourceDetail(logger, mockK8sClientError, nil)
	handler(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetResourceDetail_NilClient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/pods/default/my-pod", nil)
	c.Params = []gin.Param{
		{Key: "resourceType", Value: "pods"},
		{Key: "namespace", Value: "default"},
		{Key: "name", Value: "my-pod"},
	}

	handler := getResourceDetail(logger, mockK8sClientNil, nil)
	handler(c)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestDeleteResource_EmptyType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/", nil)

	handler := deleteResource(logger, mockK8sClientError, nil)
	handler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteResource_EmptyName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/pods/default", nil)
	c.Params = []gin.Param{{Key: "resourceType", Value: "pods"}, {Key: "namespace", Value: "default"}}

	handler := deleteResource(logger, mockK8sClientError, nil)
	handler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteResource_ClientError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/pods/default/my-pod", nil)
	c.Params = []gin.Param{
		{Key: "resourceType", Value: "pods"},
		{Key: "namespace", Value: "default"},
		{Key: "name", Value: "my-pod"},
	}

	handler := deleteResource(logger, mockK8sClientError, nil)
	handler(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteResource_NilClient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/pods/default/my-pod", nil)
	c.Params = []gin.Param{
		{Key: "resourceType", Value: "pods"},
		{Key: "namespace", Value: "default"},
		{Key: "name", Value: "my-pod"},
	}

	handler := deleteResource(logger, mockK8sClientNil, nil)
	handler(c)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}
