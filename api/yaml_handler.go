package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/service"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// RegisterYAMLRoutes 注册 YAML 相关路由
func RegisterYAMLRoutes(r *gin.RouterGroup, logger *zap.Logger, getK8sClient K8sClientProvider) {
	r.GET("/:resourceType/:namespace/:name/yaml", getResourceYAML(logger, getK8sClient))
	r.PUT("/:resourceType/:namespace/:name/yaml", updateResourceYAML(logger, getK8sClient))
}

// getResourceYAML 获取资源 YAML
// 不支持 Event 资源
func getResourceYAML(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceType := c.Param("resourceType")
		namespace := c.Param("namespace")
		name := c.Param("name")

		// Event 不支持 YAML 接口
		if resourceType == "event" || resourceType == "events" {
			middleware.ResponseError(c, logger, fmt.Errorf("Events resource does not support YAML format"), http.StatusBadRequest)
			return
		}

		if err := validateResourceParams(resourceType, namespace); err != nil {
			middleware.ResponseError(c, logger, err, http.StatusBadRequest)
			return
		}

		clientset, _, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		ctx := GetRequestContext(c)
		obj, err := service.GetResourceByName(ctx, clientset, resourceType, namespace, name)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}

		yamlBytes, err := yaml.Marshal(obj)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, string(yamlBytes), "YAML retrieved successfully", nil)
	}
}

// updateResourceYAML 更新资源 YAML
func updateResourceYAML(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceType := c.Param("resourceType")
		namespace := c.Param("namespace")
		name := c.Param("name")

		// Event 不支持 YAML 接口
		if resourceType == "event" || resourceType == "events" {
			middleware.ResponseError(c, logger, fmt.Errorf("Events resource does not support YAML update"), http.StatusBadRequest)
			return
		}

		if err := validateResourceParams(resourceType, namespace); err != nil {
			middleware.ResponseError(c, logger, err, http.StatusBadRequest)
			return
		}

		// 解析请求体 - 支持两种格式：{yaml: {...}} 或直接 {...}
		var reqBody map[string]interface{}
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			middleware.ResponseError(c, logger, fmt.Errorf("invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		// 如果是 {yaml: {...}} 格式，提取 yaml 字段
		var objData interface{} = reqBody
		if yamlData, ok := reqBody["yaml"]; ok {
			objData = yamlData
		}

		if err := validateResourceIdentity(resourceType, namespace, name, objData); err != nil {
			middleware.ResponseError(c, logger, err, http.StatusBadRequest)
			return
		}

		// 将 map 转换为 JSON 字节
		jsonBytes, err := json.Marshal(objData)
		if err != nil {
			middleware.ResponseError(c, logger, fmt.Errorf("JSON serialization failed: %v", err), http.StatusBadRequest)
			return
		}

		clientset, _, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		ctx := GetRequestContext(c)

		// 根据资源类型调用不同的更新方法
		err = service.UpdateResourceByType(ctx, clientset, resourceType, namespace, name, jsonBytes)
		if err != nil {
			logger.Error("Failed to update resource", zap.Error(err))
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, nil, "Resource updated successfully", nil)
	}
}
