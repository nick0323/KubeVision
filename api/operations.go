package api

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RegisterOperations 注册所有资源操作接口（YAML、关联资源等）
// 注意：此函数仅作为路由聚合器，具体逻辑已拆分至 yaml_handler.go 和 related_handler.go
func RegisterOperations(r *gin.RouterGroup, logger *zap.Logger, getK8sClient K8sClientProvider) {
	// 1. 注册 YAML 路由
	RegisterYAMLRoutes(r, logger, getK8sClient)

	// 2. 注册关联资源路由
	RegisterRelatedRoutes(r, logger, getK8sClient)
}
