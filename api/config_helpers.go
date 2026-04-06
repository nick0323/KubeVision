package api

import (
	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/config"
	"github.com/nick0323/K8sVision/model"
)

// configManager 全局配置管理器（包内使用）
var configManager *config.Manager

// SetConfigManager 设置配置管理器（应用启动时调用）
func SetConfigManager(cm *config.Manager) {
	configManager = cm
}

// GetAuthConfig 获取认证配置
// 返回：认证配置，如果未初始化则返回 nil
func GetAuthConfig() *model.AuthConfig {
	if configManager == nil {
		return nil
	}
	return configManager.GetAuthConfig()
}

// GetUsernameFromContext 从 gin 上下文获取用户名
// 参数：c - gin 上下文
// 返回：用户名字符串，如果不存在则返回空字符串
func GetUsernameFromContext(c *gin.Context) string {
	username, exists := c.Get("username")
	if !exists {
		return ""
	}

	userStr, ok := username.(string)
	if !ok {
		return ""
	}

	return userStr
}
