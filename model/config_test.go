package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "json", cfg.Log.Format)
	assert.True(t, cfg.Cache.Enabled)
	assert.Equal(t, 1000, cfg.Cache.MaxSize)
	assert.Equal(t, 5*time.Minute, cfg.Cache.TTL)
	assert.Equal(t, 24*time.Hour, cfg.JWT.Expiration)
	assert.Equal(t, "admin", cfg.Auth.Username)
	assert.Equal(t, 5, cfg.Auth.MaxLoginFail)
}

func TestConfigValidate(t *testing.T) {
	t.Run("valid default config", func(t *testing.T) {
		cfg := DefaultConfig()
		assert.NoError(t, cfg.Validate())
	})

	t.Run("empty server port", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Server.Port = ""
		assert.Error(t, cfg.Validate())
	})

	t.Run("empty server host", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Server.Host = ""
		assert.Error(t, cfg.Validate())
	})

	t.Run("invalid log level", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Log.Level = "trace"
		assert.Error(t, cfg.Validate())
	})

	t.Run("invalid log format", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Log.Format = "xml"
		assert.Error(t, cfg.Validate())
	})

	t.Run("zero cache TTL when enabled", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Cache.TTL = 0
		assert.Error(t, cfg.Validate())
	})

	t.Run("zero cache max size when enabled", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Cache.MaxSize = 0
		assert.Error(t, cfg.Validate())
	})

	t.Run("disabled cache skips validation", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Cache.Enabled = false
		cfg.Cache.TTL = 0
		cfg.Cache.MaxSize = 0
		assert.NoError(t, cfg.Validate())
	})

	t.Run("kubernetes QPS must be positive", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Kubernetes.QPS = 0
		assert.Error(t, cfg.Validate())
	})

	t.Run("kubernetes Burst must be positive", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Kubernetes.Burst = 0
		assert.Error(t, cfg.Validate())
	})

	t.Run("kubernetes Timeout must be positive", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Kubernetes.Timeout = 0
		assert.Error(t, cfg.Validate())
	})

	t.Run("multiple validation errors", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Server.Port = ""
		cfg.Log.Level = "trace"
		err := cfg.Validate()
		assert.Error(t, err)
		assert.ErrorContains(t, err, "config validation failed: 2 errors")
	})
}

func TestConfigGetServerAddress(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, "0.0.0.0:8080", cfg.GetServerAddress())
}

func TestConfigIsDevelopment(t *testing.T) {
	t.Run("debug level returns true", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Log.Level = "debug"
		assert.True(t, cfg.IsDevelopment())
	})

	t.Run("info level returns false", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Log.Level = "info"
		assert.False(t, cfg.IsDevelopment())
	})
}
