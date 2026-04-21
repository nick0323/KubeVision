package api

import (
	"crypto/rand"
	"encoding/base64"
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

var (
	authManager   *AuthManager
	configManager *config.Manager
)

func SetConfigManager(mgr *config.Manager) {
	configManager = mgr
}

func GetAuthConfig() model.AuthConfig {
	if configManager == nil {
		return model.AuthConfig{}
	}
	return configManager.GetConfig().Auth
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

func generateJTI() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

func InitAuthManager(logger *zap.Logger) {
	if configManager == nil {
		logger.Fatal("Config manager not initialized")
	}
	authManager = NewAuthManager(logger, configManager)
}

func LoginHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.ResponseError(c, logger, &model.APIError{
				Code:    http.StatusBadRequest,
				Message: "Invalid request parameter format",
				Details: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		req.Username = strings.TrimSpace(req.Username)
		req.Password = strings.TrimSpace(req.Password)

		if err := validateLoginRequest(req); err != nil {
			middleware.ResponseError(c, logger, err, http.StatusBadRequest)
			return
		}

		authConfig := GetAuthConfig()
		if authConfig.Username == "" {
			middleware.ResponseError(c, logger, &model.APIError{
				Code:    http.StatusInternalServerError,
				Message: "System configuration not initialized",
			}, http.StatusInternalServerError)
			return
		}

		clientIP := c.ClientIP()
		username := req.Username

		if authManager != nil && authManager.IsLocked(username, clientIP) {
			sendLockResponse(c, logger, authManager, username, clientIP, authConfig)
			return
		}

		usernameMatch := req.Username == authConfig.Username
		passwordMatch := verifyPassword(req.Password, authConfig.Password)

		if usernameMatch && passwordMatch {
			handleLoginSuccess(c, logger, username, req.Username, clientIP, authConfig)
			return
		}

		handleLoginFailure(c, logger, username, clientIP, authConfig)
	}
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
			Message: fmt.Sprintf("Username length cannot exceed %d characters", UsernameMaxLen),
		}
	}
	if len(req.Password) > PasswordMaxLen {
		return &model.APIError{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Password length cannot exceed %d characters", PasswordMaxLen),
		}
	}
	return nil
}

func sendLockResponse(c *gin.Context, logger *zap.Logger, authManager *AuthManager, username, clientIP string, authConfig model.AuthConfig) {
	middleware.ResponseError(c, logger, &model.APIError{
		Code:    http.StatusTooManyRequests,
		Message: "Too many login failures, account locked",
		Details: map[string]interface{}{
			"remainingAttempts": authManager.GetRemainingAttempts(username, clientIP),
			"maxFailCount":      authConfig.MaxLoginFail,
			"lockDuration":      authConfig.LockDuration.String(),
			"lockTime":          authManager.GetLockTime(username, clientIP).String(),
		},
	}, http.StatusTooManyRequests)
}

func verifyPassword(reqPassword, configPassword string) bool {
	if isHashedPassword(configPassword) {
		pm := NewPasswordManager()
		return pm.VerifyPassword(reqPassword, configPassword)
	}
	return reqPassword == configPassword
}

func handleLoginSuccess(c *gin.Context, logger *zap.Logger, username, reqUsername, clientIP string, authConfig model.AuthConfig) {
	logger.Info("User login successful",
		zap.String("username", reqUsername),
		zap.String("clientIP", clientIP),
	)

	tokenString, err := generateToken(reqUsername, authConfig)
	if err != nil {
		logger.Error("Token generation failed", zap.String("username", reqUsername), zap.Error(err))
		middleware.ResponseError(c, logger, &model.APIError{
			Code:    http.StatusInternalServerError,
			Message: "Token generation failed",
		}, http.StatusInternalServerError)
		return
	}

	authManager.RecordSuccess(username, clientIP)
	middleware.ResponseSuccess(c, map[string]string{"token": tokenString}, "Login successful", nil)
}

func handleLoginFailure(c *gin.Context, logger *zap.Logger, username, clientIP string, authConfig model.AuthConfig) {
	logger.Warn("User login failed",
		zap.String("username", username),
		zap.String("clientIP", clientIP),
	)

	if authManager != nil {
		authManager.RecordFailure(username, clientIP)
		middleware.ResponseError(c, logger, &model.APIError{
			Code:    http.StatusUnauthorized,
			Message: "Invalid username or password",
			Details: map[string]interface{}{
				"remainingAttempts": authManager.GetRemainingAttempts(username, clientIP),
				"maxFailCount":      authConfig.MaxLoginFail,
			},
		}, http.StatusUnauthorized)
		return
	}

	middleware.ResponseError(c, logger, &model.APIError{
		Code:    http.StatusUnauthorized,
		Message: "Invalid username or password",
		Details: map[string]interface{}{
			"maxFailCount": authConfig.MaxLoginFail,
		},
	}, http.StatusUnauthorized)
}

func generateToken(username string, authConfig model.AuthConfig) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(authConfig.SessionTimeout).Unix(),
		"iss":      middleware.JWTIssuer,
		"aud":      middleware.JWTAudience,
		"jti":      generateJTI(),
	})

	return token.SignedString(configManager.GetJWTSecret())
}
