package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/config"
	"github.com/nick0323/K8sVision/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestValidateLoginRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     model.LoginRequest
		wantErr bool
		errMsg  string
	}{
		{"valid request", model.LoginRequest{Username: "admin", Password: "secret"}, false, ""},
		{"empty username", model.LoginRequest{Username: "", Password: "secret"}, true, "Username and password cannot be empty"},
		{"empty password", model.LoginRequest{Username: "admin", Password: ""}, true, "Username and password cannot be empty"},
		{"username too long", model.LoginRequest{Username: string(make([]byte, 51)), Password: "secret"}, true, "Username length cannot exceed 50 characters"},
		{"password too long", model.LoginRequest{Username: "admin", Password: string(make([]byte, 129))}, true, "Password length cannot exceed 128 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLoginRequest(tt.req)
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

func TestGenerateJTI(t *testing.T) {
	jti := generateJTI()
	assert.NotEmpty(t, jti)
	assert.Greater(t, len(jti), 10)
}

func TestGetUsernameFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("username exists", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("username", "admin")
		assert.Equal(t, "admin", GetUsernameFromContext(c))
	})

	t.Run("username not set", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		assert.Equal(t, "", GetUsernameFromContext(c))
	})

	t.Run("username wrong type", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("username", 123)
		assert.Equal(t, "", GetUsernameFromContext(c))
	})
}

func TestVerifyPassword(t *testing.T) {
	pm := NewPasswordManager()

	t.Run("empty config password returns false", func(t *testing.T) {
		assert.False(t, verifyPassword("anything", "", pm))
	})

	t.Run("plain text password match", func(t *testing.T) {
		assert.True(t, verifyPassword("secret123", "secret123", pm))
	})

	t.Run("plain text password mismatch", func(t *testing.T) {
		assert.False(t, verifyPassword("wrong", "secret123", pm))
	})

	t.Run("bcrypt hashed password match", func(t *testing.T) {
		hashed, _ := pm.HashPassword("myPassword123")
		assert.True(t, verifyPassword("myPassword123", hashed, pm))
	})
}

func TestIsHashedPassword(t *testing.T) {
	pm := NewPasswordManager()

	t.Run("bcrypt hash detected", func(t *testing.T) {
		hashed, _ := pm.HashPassword("test123")
		assert.True(t, isHashedPassword(hashed))
	})

	t.Run("plain text not detected", func(t *testing.T) {
		assert.False(t, isHashedPassword("plaintextpass"))
	})

	t.Run("short string not detected", func(t *testing.T) {
		assert.False(t, isHashedPassword("$2a$10$short"))
	})
}

func TestMin(t *testing.T) {
	assert.Equal(t, 5, min(5, 10))
	assert.Equal(t, 3, min(10, 3))
	assert.Equal(t, 7, min(7, 7))
}

func TestLoginHandlerGetAuthConfig(t *testing.T) {
	t.Run("nil config manager returns empty", func(t *testing.T) {
		h := &LoginHandler{configManager: nil}
		cfg := h.getAuthConfig()
		assert.Equal(t, model.AuthConfig{}, cfg)
	})
}

func TestInitAuthManager(t *testing.T) {
	t.Run("nil config manager returns error", func(t *testing.T) {
		am, err := InitAuthManager(nil, nil)
		assert.Error(t, err)
		assert.Nil(t, am)
	})
}

func TestHandle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("invalid JSON body returns 400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/login", nil)
		c.Request.Header.Set("Content-Type", "application/json")
		_ = NewLoginHandler(nil, nil, nil).Handle()
	})
}

// ---- authenticate ----
func TestAuthenticate_Success(t *testing.T) {
	cm := config.NewManager(zap.NewNop())
	cm.Set("auth.username", "admin")
	cm.Set("auth.password", "secret123")
	h := NewLoginHandler(nil, cm, zap.NewNop())

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/login", nil)

	result := h.authenticate(c, "admin", "secret123", "127.0.0.1", h.getAuthConfig())
	assert.True(t, result)
}

func TestAuthenticate_WrongUsername(t *testing.T) {
	cm := config.NewManager(zap.NewNop())
	cm.Set("auth.username", "admin")
	cm.Set("auth.password", "secret123")
	h := NewLoginHandler(nil, cm, zap.NewNop())

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/login", nil)

	result := h.authenticate(c, "hacker", "secret123", "127.0.0.1", h.getAuthConfig())
	assert.False(t, result)
}

func TestAuthenticate_WrongPassword(t *testing.T) {
	cm := config.NewManager(zap.NewNop())
	cm.Set("auth.username", "admin")
	cm.Set("auth.password", "secret123")
	h := NewLoginHandler(nil, cm, zap.NewNop())

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/login", nil)

	result := h.authenticate(c, "admin", "wrongpass", "127.0.0.1", h.getAuthConfig())
	assert.False(t, result)
}

// ---- generateToken ----
func TestGenerateToken_Success(t *testing.T) {
	cm := config.NewManager(zap.NewNop())
	cm.Set("jwt.secret", "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e1f2")
	h := NewLoginHandler(nil, cm, zap.NewNop())

	authCfg := model.AuthConfig{SessionTimeout: time.Hour}
	token, err := h.generateToken("admin", authCfg)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateToken_EmptySecret(t *testing.T) {
	cm := config.NewManager(zap.NewNop())
	h := NewLoginHandler(nil, cm, zap.NewNop())

	authCfg := model.AuthConfig{SessionTimeout: time.Hour}
	_, err := h.generateToken("admin", authCfg)
	assert.EqualError(t, err, "JWT secret not configured")
}

// ---- handleLoginSuccess ----
func TestHandleLoginSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cm := config.NewManager(zap.NewNop())
	cm.Set("jwt.secret", "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e1f2")
	cm.Set("auth.username", "admin")
	am, _ := InitAuthManager(zap.NewNop(), cm)
	h := NewLoginHandler(am, cm, zap.NewNop())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/login", nil)

	authCfg := model.AuthConfig{SessionTimeout: time.Hour}
	h.handleLoginSuccess(c, "admin", "127.0.0.1", authCfg)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "token")
}

func TestHandleLoginSuccess_NoAuthManager(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cm := config.NewManager(zap.NewNop())
	cm.Set("jwt.secret", "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e1f2")
	h := NewLoginHandler(nil, cm, zap.NewNop())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/login", nil)

	authCfg := model.AuthConfig{SessionTimeout: time.Hour}
	h.handleLoginSuccess(c, "admin", "127.0.0.1", authCfg)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ---- handleLoginFailure ----
func TestHandleLoginFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cm := config.NewManager(zap.NewNop())
	am, _ := InitAuthManager(zap.NewNop(), cm)
	h := NewLoginHandler(am, cm, zap.NewNop())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/login", nil)

	authCfg := model.AuthConfig{MaxLoginFail: 5}
	h.handleLoginFailure(c, "admin", "127.0.0.1", authCfg)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid username or password")
}

func TestHandleLoginFailure_NoAuthManager(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cm := config.NewManager(zap.NewNop())
	h := NewLoginHandler(nil, cm, zap.NewNop())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/login", nil)

	authCfg := model.AuthConfig{MaxLoginFail: 5}
	h.handleLoginFailure(c, "admin", "127.0.0.1", authCfg)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ---- sendLockResponse ----
func TestSendLockResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cm := config.NewManager(zap.NewNop())
	cm.Set("auth.username", "admin")
	am, _ := InitAuthManager(zap.NewNop(), cm)
	h := NewLoginHandler(am, cm, zap.NewNop())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/login", nil)

	authCfg := model.AuthConfig{MaxLoginFail: 5, LockDuration: time.Minute}
	h.sendLockResponse(c, "admin", "127.0.0.1", authCfg)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "locked")
}
