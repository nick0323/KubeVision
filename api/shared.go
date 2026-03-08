package api

import (
	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/config"
	"github.com/nick0323/K8sVision/model"
	"github.com/nick0323/K8sVision/service"
)

// 共享的配置管理器
var (
	configManager *config.Manager
	configStore   *service.ConfigMapStore
)

// SetConfigManager 设置配置管理器
func SetConfigManager(cm *config.Manager) {
	configManager = cm
}

// GetConfigManager 获取配置管理器
func GetConfigManager() *config.Manager {
	return configManager
}

// SetConfigStore 设置 ConfigMap 配置存储
func SetConfigStore(cs *service.ConfigMapStore) {
	configStore = cs
}

// GetConfigStore 获取 ConfigMap 配置存储
func GetConfigStore() *service.ConfigMapStore {
	return configStore
}

// GetAuthConfig 获取认证配置（优先从 ConfigMap 存储获取）
func GetAuthConfig() *model.AuthConfig {
	if configStore != nil {
		if storedConfig := configStore.GetAuthConfig(); storedConfig != nil {
			return storedConfig
		}
	}
	return configManager.GetAuthConfig()
}

// GetUsernameFromContext 从上下文获取用户名
func GetUsernameFromContext(c *gin.Context) string {
	if username, exists := c.Get("username"); exists {
		if userStr, ok := username.(string); ok {
			return userStr
		}
	}
	return ""
}
