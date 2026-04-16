package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidate_ErrorDetail(t *testing.T) {
	// 测试 JWT 配置错误
	config := &Config{
		Server: ServerConfig{
			Port: "8080",
			Host: "0.0.0.0",
		},
		JWT: JWTConfig{
			Secret: "", // 空 secret
		},
		Auth: AuthConfig{
			Username:     "admin",
			Password:     "password",
			MaxLoginFail: 5,
			LockDuration: 10,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
		Cache: CacheConfig{
			Enabled: true,
			TTL:     300,
			MaxSize: 1000,
		},
		Kubernetes: KubernetesConfig{
			Timeout: 30,
			QPS:     100,
			Burst:   200,
		},
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config validation failed")
	assert.Contains(t, err.Error(), "JWT config")
	assert.Contains(t, err.Error(), "1. ")
}

func TestConfigValidate_MultipleErrors(t *testing.T) {
	// 测试多个配置错误
	config := &Config{
		Server: ServerConfig{
			Port: "", // 错误：空端口
			Host: "", // 错误：空主机
		},
		JWT: JWTConfig{
			Secret: "", // 错误：空 secret
		},
		Auth: AuthConfig{
			Username:     "", // 错误：空用户名
			Password:     "", // 错误：空密码
			MaxLoginFail: 0,  // 错误：0 次失败
		},
		Log: LogConfig{
			Level:  "invalid", // 错误：无效级别
			Format: "invalid", // 错误：无效格式
		},
		Cache: CacheConfig{
			Enabled: true,
			TTL:     0, // 错误：0 TTL
		},
		Kubernetes: KubernetesConfig{
			Timeout: 0, // 错误：0 超时
			QPS:     0, // 错误：0 QPS
		},
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config validation failed")
	// 验证错误消息包含多个错误
	assert.Greater(t, len(err.Error()), 50)
}

func TestConfigValidate_Success(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Port: "8080",
			Host: "0.0.0.0",
		},
		JWT: JWTConfig{
			Secret:     "test-secret-key-123456",
			Expiration: 86400,
		},
		Auth: AuthConfig{
			Username:     "admin",
			Password:     "password",
			MaxLoginFail: 5,
			LockDuration: 600,
			BcryptCost:   12,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
		Cache: CacheConfig{
			Enabled: true,
			TTL:     300,
			MaxSize: 1000,
		},
		Kubernetes: KubernetesConfig{
			Timeout: 30,
			QPS:     100,
			Burst:   200,
		},
	}

	err := config.Validate()
	assert.NoError(t, err)
}
