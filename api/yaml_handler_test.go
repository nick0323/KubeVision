package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestValidateResourceType(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		objData      interface{}
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "matching kind",
			resourceType: "deployment",
			objData:      map[string]interface{}{"kind": "Deployment"},
			wantErr:      false,
		},
		{
			name:         "mismatched kind",
			resourceType: "deployment",
			objData:      map[string]interface{}{"kind": "Service"},
			wantErr:      true,
			errMsg:       "resource kind mismatch",
		},
		{
			name:         "empty kind skips validation",
			resourceType: "deployment",
			objData:      map[string]interface{}{"kind": ""},
			wantErr:      false,
		},
		{
			name:         "unknown resource type",
			resourceType: "unknown",
			objData:      map[string]interface{}{"kind": "Unknown"},
			wantErr:      false,
		},
		{
			name:         "invalid data format",
			resourceType: "pod",
			objData:      "not a map",
			wantErr:      true,
			errMsg:       "invalid resource data format",
		},
		{
			name:         "no kind field",
			resourceType: "service",
			objData:      map[string]interface{}{"metadata": map[string]interface{}{"name": "test"}},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateResourceType(tt.resourceType, tt.objData)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateResourceIdentity(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		namespace    string
		nameParam    string
		objData      interface{}
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "valid identity",
			resourceType: "pod",
			namespace:    "default",
			nameParam:    "my-pod",
			objData:      map[string]interface{}{"metadata": map[string]interface{}{"name": "my-pod", "namespace": "default"}},
			wantErr:      false,
		},
		{
			name:         "name mismatch",
			resourceType: "pod",
			namespace:    "default",
			nameParam:    "my-pod",
			objData:      map[string]interface{}{"metadata": map[string]interface{}{"name": "other-pod", "namespace": "default"}},
			wantErr:      true,
			errMsg:       "name",
		},
		{
			name:         "namespace mismatch",
			resourceType: "pod",
			namespace:    "default",
			nameParam:    "my-pod",
			objData:      map[string]interface{}{"metadata": map[string]interface{}{"name": "my-pod", "namespace": "other"}},
			wantErr:      true,
			errMsg:       "namespace",
		},
		{
			name:         "invalid data format",
			resourceType: "pod",
			namespace:    "default",
			nameParam:    "my-pod",
			objData:      "not a map",
			wantErr:      true,
			errMsg:       "cannot unmarshal string",
		},
		{
			name:         "no metadata means no name",
			resourceType: "pod",
			namespace:    "default",
			nameParam:    "my-pod",
			objData:      map[string]interface{}{},
			wantErr:      true,
			errMsg:       "metadata.name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateResourceIdentity(tt.resourceType, tt.namespace, tt.nameParam, tt.objData)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisterYAMLRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	r := gin.New()
	group := r.Group("/api/v1")
	RegisterYAMLRoutes(group, logger, mockK8sClientError)

	routes := r.Routes()
	paths := make(map[string]string)
	for _, route := range routes {
		paths[route.Method+" "+route.Path] = route.Handler
	}

	basePath := "/api/v1"
	assert.Contains(t, paths, "GET "+basePath+"/:resourceType/:namespace/:name/yaml")
	assert.Contains(t, paths, "PUT "+basePath+"/:resourceType/:namespace/:name/yaml")
	assert.Contains(t, paths, "POST "+basePath+"/:resourceType/yaml")
}

func TestGetResourceYAML_EventRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = []gin.Param{{Key: "resourceType", Value: "event"}}

	handler := getResourceYAML(logger, mockK8sClientError)
	handler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetResourceYAML_EventsRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = []gin.Param{{Key: "resourceType", Value: "events"}}

	handler := getResourceYAML(logger, mockK8sClientError)
	handler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetResourceYAML_EmptyParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	handler := getResourceYAML(logger, mockK8sClientError)
	handler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetResourceYAML_ClientError(t *testing.T) {
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

	handler := getResourceYAML(logger, mockK8sClientError)
	handler(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateResourceYAML_EventRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/", nil)
	c.Params = []gin.Param{{Key: "resourceType", Value: "event"}}

	handler := updateResourceYAML(logger, mockK8sClientError)
	handler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateResourceYAML_EmptyParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/", nil)

	handler := updateResourceYAML(logger, mockK8sClientError)
	handler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateResourceYAML_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/", strings.NewReader("not json"))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = []gin.Param{
		{Key: "resourceType", Value: "pods"},
		{Key: "namespace", Value: "default"},
		{Key: "name", Value: "my-pod"},
	}

	handler := updateResourceYAML(logger, mockK8sClientError)
	handler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateResourceYAML_ClientError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"metadata":{"name":"my-pod","namespace":"default"}}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = []gin.Param{
		{Key: "resourceType", Value: "pods"},
		{Key: "namespace", Value: "default"},
		{Key: "name", Value: "my-pod"},
	}

	handler := updateResourceYAML(logger, mockK8sClientError)
	handler(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateResourceYAML_EventRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Params = []gin.Param{{Key: "resourceType", Value: "event"}}

	handler := createResourceYAML(logger, mockK8sClientError)
	handler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateResourceYAML_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not json"))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = []gin.Param{{Key: "resourceType", Value: "pods"}}

	handler := createResourceYAML(logger, mockK8sClientError)
	handler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateResourceYAML_KindMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"kind":"Deployment","metadata":{"name":"test"}}`
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = []gin.Param{{Key: "resourceType", Value: "pod"}}

	handler := createResourceYAML(logger, mockK8sClientError)
	handler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateResourceYAML_ClientError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"kind":"Pod","metadata":{"name":"test"}}`
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = []gin.Param{{Key: "resourceType", Value: "pods"}}

	handler := createResourceYAML(logger, mockK8sClientError)
	handler(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateResourceYAML_YamlWrapperFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"yaml":{"kind":"Pod","metadata":{"name":"test"}}}`
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = []gin.Param{{Key: "resourceType", Value: "pods"}}

	handler := createResourceYAML(logger, mockK8sClientError)
	handler(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
