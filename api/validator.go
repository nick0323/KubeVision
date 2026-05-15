package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/pkg/k8s"
	"go.uber.org/zap"
)

// isValidDNSName 检查 DNS 名称格式
func isValidDNSName(name string) bool {
	if len(name) == 0 || len(name) > 253 {
		return false
	}
	for i, r := range name {
		if r == '-' {
			if i == 0 || i == len(name)-1 {
				return false
			}
			continue
		}
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		return false
	}
	return true
}

// isValidResourceName 检查资源名称是否包含危险字符
func isValidResourceName(name string) bool {
	if !isValidDNSName(name) {
		return false
	}
	if strings.ContainsAny(name, "../\\") {
		return false
	}
	return true
}

// validateResourceParams 验证资源类型和 namespace 参数
func validateResourceParams(resourceType, namespace string) error {
	rt := k8s.ResourceType(strings.ToLower(resourceType)).Normalize()

	if rt.IsClusterScoped() {
		if namespace != "" {
			return fmt.Errorf("resource type %s is cluster-scoped, namespace should not be specified", resourceType)
		}
	} else {
		if namespace == "" {
			return fmt.Errorf("resource type %s is namespace-scoped, namespace must be specified", resourceType)
		}
	}
	return nil
}

// validateResourceIdentity 验证资源身份
func validateResourceIdentity(resourceType, namespace, name string, objData any) error {
	var payload struct {
		Metadata struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		} `json:"metadata"`
	}

	raw, err := json.Marshal(objData)
	if err != nil {
		return fmt.Errorf("failed to marshal resource identity: %w", err)
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return fmt.Errorf("failed to parse resource identity: %w", err)
	}

	if payload.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if payload.Metadata.Name != name {
		return fmt.Errorf("metadata.name does not match request path")
	}

	if k8s.ResourceType(strings.ToLower(resourceType)).Normalize().IsClusterScoped() {
		if payload.Metadata.Namespace != "" {
			return fmt.Errorf("cluster-scoped resource must not include metadata.namespace")
		}
		return nil
	}

	if payload.Metadata.Namespace == "" {
		return fmt.Errorf("metadata.namespace is required")
	}
	if payload.Metadata.Namespace != namespace {
		return fmt.Errorf("metadata.namespace does not match request path")
	}

	return nil
}

// validatePodLogParams 验证 Pod 日志参数
func validatePodLogParams(namespace, podName, container string) error {
	if !isValidResourceName(namespace) {
		return fmt.Errorf("invalid namespace format")
	}
	if !isValidResourceName(podName) {
		return fmt.Errorf("invalid pod name format")
	}
	if container != "" && !isValidResourceName(container) {
		return fmt.Errorf("invalid container name format")
	}
	if namespace == "" || podName == "" {
		return fmt.Errorf("namespace and pod parameters are required")
	}
	return nil
}

// validateWebSocketToken 验证 WebSocket token
func validateWebSocketToken(c *gin.Context, logger *zap.Logger, configProvider middleware.ConfigProvider) error {
	tokenStr := ExtractTokenFromRequest(c)
	if tokenStr == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return fmt.Errorf("missing token")
	}

	jwtSecret := configProvider.GetJWTSecret()
	if len(jwtSecret) == 0 {
		c.AbortWithStatus(http.StatusInternalServerError)
		return fmt.Errorf("JWT secret not configured")
	}

	_, err := middleware.VerifyToken(tokenStr, jwtSecret)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return err
	}
	return nil
}
