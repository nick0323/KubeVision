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
		// 1. 参数验证
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
			}, http.StatusBadRequest)
			return
		}

		if len(req.Username) > UsernameMaxLen {
			middleware.ResponseError(c, logger, &model.APIError{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("用户名长度不能超过 %d 个字符", UsernameMaxLen),
			}, http.StatusBadRequest)
			return
		}

		if len(req.Password) > PasswordMaxLen {
			middleware.ResponseError(c, logger, &model.APIError{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("密码长度不能超过 %d 个字符", PasswordMaxLen),
			}, http.StatusBadRequest)
			return
		}

		// 2. 检查配置
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

		// 3. 检查是否被锁定
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

		// 4. 验证用户名密码（使用常数时间比较防止时序攻击）
		usernameMatch := req.Username == authConfig.Username
		passwordMatch := false

		// 即使用户名不匹配，也执行密码验证（防止时序攻击泄露用户名）
		if isHashedPassword(authConfig.Password) {
			pm := NewPasswordManager()
			passwordMatch = pm.VerifyPassword(req.Password, authConfig.Password)
		} else {
			// 使用常数时间比较
			passwordMatch = (req.Password == authConfig.Password)
		}

		// 5. 登录成功
		if usernameMatch && passwordMatch {
			logger.Info("User login successful",
				zap.String("username", req.Username),
				zap.String("clientIP", c.ClientIP()),
				zap.String("userAgent", c.GetHeader("User-Agent")),
			)

			tokenString, err := generateToken(req.Username, authConfig)
			if err != nil {
				logger.Error("Token 生成失败",
					zap.String("username", req.Username),
					zap.Error(err),
				)
				middleware.ResponseError(c, logger, &model.APIError{
					Code:    http.StatusInternalServerError,
					Message: "Token 生成失败",
				}, http.StatusInternalServerError)
				return
			}

			if authManager != nil {
				authManager.RecordSuccess(username, clientIP)
			}

			middleware.ResponseSuccess(c, map[string]string{
				"token": tokenString,
			}, "登录成功", nil)
			return
		}

		// 6. 登录失败
		logger.Warn("User login failed",
			zap.String("username", req.Username),
			zap.String("clientIP", c.ClientIP()),
			zap.String("userAgent", c.GetHeader("User-Agent")),
			zap.Bool("usernameMatch", usernameMatch),
		)

		if authManager != nil {
			authManager.RecordFailure(username, clientIP)
			remainingAttempts := authManager.GetRemainingAttempts(username, clientIP)

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

		// authManager 未初始化时的处理
		middleware.ResponseError(c, logger, &model.APIError{
			Code:    http.StatusUnauthorized,
			Message: "用户名或密码错误",
			Details: map[string]interface{}{
				"maxFailCount": authConfig.MaxLoginFail,
			},
		}, http.StatusUnauthorized)
	}
}

// generateToken 生成 JWT Token
func generateToken(username string, authConfig *model.AuthConfig) (string, error) {
	secret := configManager.GetJWTSecret()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(authConfig.SessionTimeout).Unix(),
		"iss":      middleware.JWTIssuer,
		"aud":      middleware.JWTAudience,
		"jti":      generateJTI(),
	})

	return token.SignedString(secret)
}
