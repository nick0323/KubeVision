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
	"github.com/nick0323/K8sVision/model"
	"go.uber.org/zap"
)

// 登录配置常量
const (
	UsernameMaxLen = 50
	PasswordMaxLen = 128
)

var (
	authManager *AuthManager
)

// generateJTI 生成 JWT ID
func generateJTI() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

func InitAuthManager(logger *zap.Logger) {
	if configManager == nil {
		logger.Fatal("配置管理器未初始化")
		return
	}
	authManager = NewAuthManager(logger, configManager)
}

// LoginHandler 登录接口
func LoginHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.ResponseError(c, logger, &model.APIError{
				Code:    http.StatusBadRequest,
				Message: "请求参数格式错误",
				Details: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		req.Username = strings.TrimSpace(req.Username)
		req.Password = strings.TrimSpace(req.Password)

		if req.Username == "" || req.Password == "" {
			middleware.ResponseError(c, logger, &model.APIError{
				Code:    http.StatusBadRequest,
				Message: "用户名和密码不能为空",
				Details: "请提供用户名和密码",
			}, http.StatusBadRequest)
			return
		}

		if len(req.Username) > UsernameMaxLen {
			middleware.ResponseError(c, logger, &model.APIError{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("用户名长度不能超过 %d 个字符", UsernameMaxLen),
				Details: "请使用较短的用户名",
			}, http.StatusBadRequest)
			return
		}

		if len(req.Password) > PasswordMaxLen {
			middleware.ResponseError(c, logger, &model.APIError{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("密码长度不能超过 %d 个字符", PasswordMaxLen),
				Details: "请使用较短的密码",
			}, http.StatusBadRequest)
			return
		}

		authConfig := GetAuthConfig()
		if authConfig == nil {
			middleware.ResponseError(c, logger, &model.APIError{
				Code:    http.StatusInternalServerError,
				Message: "系统配置未初始化",
			}, http.StatusInternalServerError)
			return
		}

		clientIP := c.ClientIP()
		username := req.Username

		if authManager != nil && authManager.IsLocked(username, clientIP) {
			remainingAttempts := authManager.GetRemainingAttempts(username, clientIP)
			lockTime := authManager.GetLockTime(username, clientIP)
			middleware.ResponseError(c, logger, &model.APIError{
				Code:    http.StatusTooManyRequests,
				Message: "登录失败次数过多，账户已锁定",
				Details: map[string]interface{}{
					"remainingAttempts": remainingAttempts,
					"maxFailCount":      authConfig.MaxLoginFail,
					"lockDuration":      authConfig.LockDuration.String(),
					"lockTime":          lockTime.String(),
				},
			}, http.StatusTooManyRequests)
			return
		}

		usernameMatch := req.Username == authConfig.Username
		passwordMatch := false

		if isHashedPassword(authConfig.Password) {
			pm := NewPasswordManager()
			passwordMatch = pm.VerifyPassword(req.Password, authConfig.Password)
		} else {
			passwordMatch = req.Password == authConfig.Password
		}

		if usernameMatch && passwordMatch {
			logger.Info("用户登录成功",
				zap.String("username", req.Username),
				zap.String("clientIP", c.ClientIP()),
				zap.String("userAgent", c.GetHeader("User-Agent")),
				zap.String("event", "login_success"),
			)

			secret := configManager.GetJWTSecret()
			authConfig := GetAuthConfig()

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"username": req.Username,
				"iat":      time.Now().Unix(),
				"exp":      time.Now().Add(authConfig.SessionTimeout).Unix(),
				"iss":      middleware.JWTIssuer,
				"aud":      middleware.JWTAudience,
				"jti":      generateJTI(),
			})
			tokenString, err := token.SignedString(secret)
			if err != nil {
				logger.Error("Token 生成失败",
					zap.String("username", req.Username),
					zap.Error(err),
				)
				middleware.ResponseError(c, logger, &model.APIError{
					Code:    http.StatusInternalServerError,
					Message: "Token 生成失败",
					Details: "请稍后重试",
				}, http.StatusInternalServerError)
				return
			}

			logger.Info("JWT token 生成成功",
				zap.String("username", req.Username),
			)

			if authManager != nil {
				authManager.RecordSuccess(username, clientIP)
			}

			middleware.ResponseSuccess(c, map[string]string{
				"token": tokenString,
			}, "登录成功", nil)
			return
		}

		if authManager != nil {
			authManager.RecordFailure(username, clientIP)
			remainingAttempts := authManager.GetRemainingAttempts(username, clientIP)

			logger.Warn("用户登录失败",
				zap.String("username", req.Username),
				zap.String("clientIP", c.ClientIP()),
				zap.String("userAgent", c.GetHeader("User-Agent")),
				zap.String("event", "login_failed"),
				zap.Int("remainingAttempts", remainingAttempts),
				zap.Bool("usernameMatch", usernameMatch),
			)

			middleware.ResponseError(c, logger, &model.APIError{
				Code:    http.StatusUnauthorized,
				Message: "用户名或密码错误",
				Details: map[string]interface{}{
					"remainingAttempts": remainingAttempts,
					"maxFailCount":      authConfig.MaxLoginFail,
				},
			}, http.StatusUnauthorized)
			return
		}

		logger.Warn("用户登录失败",
			zap.String("username", req.Username),
			zap.String("clientIP", c.ClientIP()),
			zap.String("userAgent", c.GetHeader("User-Agent")),
			zap.String("event", "login_failed"),
			zap.Bool("usernameMatch", usernameMatch),
		)

		middleware.ResponseError(c, logger, &model.APIError{
			Code:    http.StatusUnauthorized,
			Message: "用户名或密码错误",
			Details: map[string]interface{}{
				"maxFailCount": authConfig.MaxLoginFail,
			},
		}, http.StatusUnauthorized)
	}
}
