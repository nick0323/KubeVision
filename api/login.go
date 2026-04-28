package api

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/config"
	"github.com/nick0323/K8sVision/model"
	"go.uber.org/zap"
)

const (
	UsernameMaxLen = 50
	PasswordMaxLen = 128
)

type LoginHandler struct {
	authManager   *AuthManager
	configManager *config.Manager
	passwordMgr   *PasswordManager
	logger        *zap.Logger
}

func NewLoginHandler(authManager *AuthManager, configManager *config.Manager, logger *zap.Logger) *LoginHandler {
	return &LoginHandler{
		authManager:   authManager,
		configManager: configManager,
		passwordMgr:   NewPasswordManager(),
		logger:        logger,
	}
}

func InitAuthManager(logger *zap.Logger, configMgr *config.Manager) (*AuthManager, error) {
	if configMgr == nil {
		return nil, errors.New("config manager not initialized")
	}
	return NewAuthManager(logger, configMgr), nil
}

func (h *LoginHandler) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.ResponseError(c, h.logger, &model.APIError{
				Code:    http.StatusBadRequest,
				Message: "Invalid request parameter format",
				Details: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		req.Username = strings.TrimSpace(req.Username)
		req.Password = strings.TrimSpace(req.Password)

		if err := validateLoginRequest(req); err != nil {
			middleware.ResponseError(c, h.logger, err, http.StatusBadRequest)
			return
		}

		authConfig := h.getAuthConfig()
		if authConfig.Username == "" {
			middleware.ResponseError(c, h.logger, &model.APIError{
				Code:    http.StatusInternalServerError,
				Message: "System configuration not initialized",
			}, http.StatusInternalServerError)
			return
		}

		clientIP := c.ClientIP()
		username := req.Username

		if h.authManager != nil && h.authManager.IsLocked(username, clientIP) {
			h.sendLockResponse(c, username, clientIP, authConfig)
			return
		}

		if h.authenticate(c, username, req.Password, clientIP, authConfig) {
			h.handleLoginSuccess(c, username, clientIP, authConfig)
		} else {
			h.handleLoginFailure(c, username, clientIP, authConfig)
		}
	}
}

func (h *LoginHandler) getAuthConfig() model.AuthConfig {
	if h.configManager == nil {
		return model.AuthConfig{}
	}
	return h.configManager.GetConfig().Auth
}

func (h *LoginHandler) sendLockResponse(c *gin.Context, username, clientIP string, authConfig model.AuthConfig) {
	middleware.ResponseError(c, h.logger, &model.APIError{
		Code:    http.StatusTooManyRequests,
		Message: "Too many login failures, account locked",
		Details: map[string]interface{}{
			"remainingAttempts": h.authManager.GetRemainingAttempts(username, clientIP),
			"maxFailCount":      authConfig.MaxLoginFail,
			"lockDuration":      authConfig.LockDuration.String(),
			"lockTime":          h.authManager.GetLockTime(username, clientIP).String(),
		},
	}, http.StatusTooManyRequests)
}

func (h *LoginHandler) authenticate(c *gin.Context, username, password, clientIP string, authConfig model.AuthConfig) bool {
	usernameMatch := username == authConfig.Username
	passwordMatch := verifyPassword(password, authConfig.Password, h.passwordMgr)

	return usernameMatch && passwordMatch
}

func verifyPassword(reqPassword, configPassword string, pm *PasswordManager) bool {
	if configPassword == "" {
		return false
	}
	hashed := isHashedPassword(configPassword)
	if hashed {
		return pm.VerifyPassword(reqPassword, configPassword)
	}
	if reqPassword == configPassword {
		return true
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (h *LoginHandler) handleLoginSuccess(c *gin.Context, username, clientIP string, authConfig model.AuthConfig) {
	h.logger.Info("User login successful",
		zap.String("username", username),
		zap.String("clientIP", clientIP),
	)

	tokenString, err := h.generateToken(username, authConfig)
	if err != nil {
		h.logger.Error("Token generation failed", zap.String("username", username), zap.Error(err))
		middleware.ResponseError(c, h.logger, &model.APIError{
			Code:    http.StatusInternalServerError,
			Message: "Token generation failed",
		}, http.StatusInternalServerError)
		return
	}

	if h.authManager != nil {
		h.authManager.RecordSuccess(username, clientIP)
	}

	middleware.ResponseSuccess(c, map[string]string{"token": tokenString}, "Login successful", nil)
}

func (h *LoginHandler) handleLoginFailure(c *gin.Context, username, clientIP string, authConfig model.AuthConfig) {
	h.logger.Warn("User login failed",
		zap.String("username", username),
		zap.String("clientIP", clientIP),
	)

	details := map[string]interface{}{
		"maxFailCount": authConfig.MaxLoginFail,
	}

	if h.authManager != nil {
		h.authManager.RecordFailure(username, clientIP)
		details["remainingAttempts"] = h.authManager.GetRemainingAttempts(username, clientIP)
	}

	middleware.ResponseError(c, h.logger, &model.APIError{
		Code:    http.StatusUnauthorized,
		Message: "Invalid username or password",
		Details: details,
	}, http.StatusUnauthorized)
}

func (h *LoginHandler) generateToken(username string, authConfig model.AuthConfig) (string, error) {
	jti := generateJTI()
	if jti == "" {
		h.logger.Warn("failed to generate JTI, using timestamp")
		jti = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(authConfig.SessionTimeout).Unix(),
		"iss":      middleware.JWTIssuer,
		"aud":      middleware.JWTAudience,
		"jti":      jti,
	})

	secret := h.configManager.GetJWTSecret()
	if len(secret) == 0 {
		return "", errors.New("JWT secret not configured")
	}

	return token.SignedString(secret)
}

func validateLoginRequest(req model.LoginRequest) error {
	if req.Username == "" || req.Password == "" {
		return &model.APIError{
			Code:    http.StatusBadRequest,
			Message: "Username and password cannot be empty",
		}
	}
	if len(req.Username) > UsernameMaxLen {
		return &model.APIError{
			Code:    http.StatusBadRequest,
			Message: "Username length cannot exceed 50 characters",
		}
	}
	if len(req.Password) > PasswordMaxLen {
		return &model.APIError{
			Code:    http.StatusBadRequest,
			Message: "Password length cannot exceed 128 characters",
		}
	}
	return nil
}

func generateJTI() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

func GetUsernameFromContext(c *gin.Context) string {
	username, exists := c.Get("username")
	if !exists {
		return ""
	}
	if usernameStr, ok := username.(string); ok {
		return usernameStr
	}
	return ""
}
