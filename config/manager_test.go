package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func newTestManager() *Manager {
	logger, _ := zap.NewDevelopment()
	mgr := NewManager(logger)
	mgr.config.JWT.Secret = "test-jwt-secret-that-is-long-enough-for-testing"
	mgr.config.Auth.Password = "hashed-password-here"
	mgr.config.Auth.Username = "admin"
	return mgr
}

func TestNewManager(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mgr := NewManager(logger)
	assert.NotNil(t, mgr)
	assert.NotNil(t, mgr.config)
	assert.Equal(t, "8080", mgr.config.Server.Port)
}

func TestManagerGetConfig(t *testing.T) {
	mgr := newTestManager()
	cfg := mgr.GetConfig()
	assert.NotNil(t, cfg)
	assert.Equal(t, "admin", cfg.Auth.Username)
}

func TestManagerGetJWTSecret(t *testing.T) {
	mgr := newTestManager()
	secret := mgr.GetJWTSecret()
	assert.Equal(t, "test-jwt-secret-that-is-long-enough-for-testing", string(secret))
}

func TestManagerGetAuthConfig(t *testing.T) {
	mgr := newTestManager()
	auth := mgr.GetAuthConfig()
	assert.NotNil(t, auth)
	assert.Equal(t, "admin", auth.Username)
	assert.Equal(t, "hashed-password-here", auth.Password)
}

func TestManagerClose(t *testing.T) {
	mgr := newTestManager()
	err := mgr.Close()
	assert.NoError(t, err)
}

func TestManagerSet(t *testing.T) {
	mgr := newTestManager()

	t.Run("set auth password", func(t *testing.T) {
		mgr.Set("auth.password", "new-hashed-pwd")
		assert.Equal(t, "new-hashed-pwd", mgr.config.Auth.Password)
	})

	t.Run("set auth username", func(t *testing.T) {
		mgr.Set("auth.username", "newadmin")
		assert.Equal(t, "newadmin", mgr.config.Auth.Username)
	})

	t.Run("set jwt secret", func(t *testing.T) {
		mgr.Set("jwt.secret", "new-secret-key-here-1234567890123456")
		assert.Equal(t, "new-secret-key-here-1234567890123456", mgr.config.JWT.Secret)
	})

	t.Run("set non-2-part key does nothing", func(t *testing.T) {
		mgr.Set("unknown.key.path", "value")
	})

	t.Run("set with wrong type does nothing", func(t *testing.T) {
		mgr.Set("auth.password", 12345)
		assert.NotEqual(t, "12345", mgr.config.Auth.Password)
	})
}

func TestManagerUpdateLogger(t *testing.T) {
	mgr := newTestManager()
	newLogger, _ := zap.NewDevelopment()
	mgr.UpdateLogger(newLogger)
}

func TestSecurityChecker(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("valid config passes", func(t *testing.T) {
		mgr := newTestManager()
		checker := NewSecurityChecker(mgr, logger)
		assert.NoError(t, checker.CheckAndValidate())
	})

	t.Run("empty JWT secret fails", func(t *testing.T) {
		mgr := newTestManager()
		mgr.config.JWT.Secret = ""
		checker := NewSecurityChecker(mgr, logger)
		err := checker.CheckAndValidate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT Secret is not configured")
	})

	t.Run("short JWT secret fails", func(t *testing.T) {
		mgr := newTestManager()
		mgr.config.JWT.Secret = "short"
		checker := NewSecurityChecker(mgr, logger)
		err := checker.CheckAndValidate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT Secret length must be at least 32 characters")
	})

	t.Run("empty auth password fails", func(t *testing.T) {
		mgr := newTestManager()
		mgr.config.Auth.Password = ""
		checker := NewSecurityChecker(mgr, logger)
		err := checker.CheckAndValidate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Admin password is not configured")
	})
}
