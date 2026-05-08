package api

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRegisterOperations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	r := gin.New()
	group := r.Group("/api/v1")

	RegisterOperations(group, logger, mockK8sClientError)

	routes := r.Routes()
	assert.GreaterOrEqual(t, len(routes), 5)

	paths := make(map[string]string)
	for _, route := range routes {
		paths[route.Method+" "+route.Path] = route.Handler
	}

	basePath := "/api/v1"
	assert.Contains(t, paths, "GET "+basePath+"/:resourceType/:namespace/:name/yaml")
	assert.Contains(t, paths, "PUT "+basePath+"/:resourceType/:namespace/:name/yaml")
	assert.Contains(t, paths, "POST "+basePath+"/:resourceType/yaml")
	assert.Contains(t, paths, "GET "+basePath+"/:resourceType/:namespace/:name/related")
	assert.Contains(t, paths, "GET "+basePath+"/:resourceType/_cluster_/:name/related")
}
