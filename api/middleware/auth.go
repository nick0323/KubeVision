package middleware

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/nick0323/K8sVision/model"
	"go.uber.org/zap"
)

// JWT 配置常量
const (
	JWTIssuer   = "k8svision"
	JWTAudience = "k8svision-client"
)

// ConfigProvider 配置提供者接口
type ConfigProvider interface {
	GetJWTSecret() []byte
}

// 全局 JWT secret（用于 WebSocket 等无法使用中间件场景）
var (
	jwtSecret     []byte
	jwtSecretOnce sync.Once
)

// SetJWTSecret 设置全局 JWT secret（在应用启动时调用）
func SetJWTSecret(secret []byte) {
	jwtSecretOnce.Do(func() {
		jwtSecret = secret
	})
}

// GetJWTSecretFromConfig 从全局配置获取 JWT secret
func GetJWTSecretFromConfig() []byte {
	if jwtSecret == nil {
		// 返回默认密钥（仅用于开发环境）
		return []byte("k8svision-default-secret-change-in-production")
	}
	return jwtSecret
}

func getJWTSecret(provider ConfigProvider) []byte {
	if provider == nil {
		panic("配置提供者未初始化")
	}
	return provider.GetJWTSecret()
}

func safeStringClaim(claims jwt.MapClaims, key string) (string, bool) {
	if claims == nil {
		return "", false
	}
	value, exists := claims[key]
	if !exists {
		return "", false
	}
	str, ok := value.(string)
	return str, ok
}

func JWTAuthMiddleware(logger *zap.Logger, configProvider ConfigProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		traceId := c.GetString("traceId")
		tokenStr := c.GetHeader("Authorization")

		// 1. 检查 Authorization 头是否存在，如果不存在则尝试从 query 参数获取（WebSocket 支持）
		if tokenStr == "" {
			tokenStr = c.Query("token")
		}

		if tokenStr == "" {
			logger.Warn("missing authorization",
				zap.String("traceId", traceId),
				zap.String("clientIP", c.ClientIP()),
				zap.String("path", c.Request.URL.Path),
			)
			ResponseError(c, logger, &model.APIError{
				Code:    http.StatusUnauthorized,
				Message: "未授权访问",
				Details: "缺少 Authorization 头或 token 参数",
			}, http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 2. 检查 Bearer 格式（如果是 header）
		if strings.HasPrefix(tokenStr, "Bearer ") {
			tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
		}

		if tokenStr == "" {
			logger.Warn("empty token",
				zap.String("traceId", traceId),
				zap.String("clientIP", c.ClientIP()),
			)
			ResponseError(c, logger, &model.APIError{
				Code:    http.StatusUnauthorized,
				Message: "Token 不能为空",
			}, http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 3. 解析和验证 JWT
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			// 验证签名算法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return getJWTSecret(configProvider), nil
		})

		if err != nil {
			logger.Warn("jwt parse error",
				zap.String("traceId", traceId),
				zap.String("clientIP", c.ClientIP()),
				zap.Error(err),
			)
			ResponseError(c, logger, &model.APIError{
				Code:    http.StatusUnauthorized,
				Message: "Token 解析失败",
			}, http.StatusUnauthorized)
			c.Abort()
			return
		}

		if !token.Valid {
			logger.Warn("invalid jwt token",
				zap.String("traceId", traceId),
				zap.String("clientIP", c.ClientIP()),
			)
			ResponseError(c, logger, &model.APIError{
				Code:    http.StatusUnauthorized,
				Message: "Token 无效或已过期",
			}, http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 4. 验证 claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims == nil {
			logger.Warn("invalid jwt claims",
				zap.String("traceId", traceId),
				zap.String("clientIP", c.ClientIP()),
			)
			ResponseError(c, logger, &model.APIError{
				Code:    http.StatusUnauthorized,
				Message: "Token 声明无效",
			}, http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 验证 username
		username, usernameExists := safeStringClaim(claims, "username")
		if !usernameExists || username == "" {
			logger.Warn("JWT token missing username",
				zap.String("traceId", traceId),
				zap.String("clientIP", c.ClientIP()),
			)
			ResponseError(c, logger, &model.APIError{
				Code:    http.StatusUnauthorized,
				Message: "Token 缺少用户名",
			}, http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 验证 issuer
		iss, issExists := safeStringClaim(claims, "iss")
		if issExists && iss != JWTIssuer {
			logger.Warn("JWT token invalid issuer",
				zap.String("traceId", traceId),
				zap.String("clientIP", c.ClientIP()),
			)
			ResponseError(c, logger, &model.APIError{
				Code:    http.StatusUnauthorized,
				Message: "Token 签发者无效",
			}, http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 验证 audience
		aud, audExists := safeStringClaim(claims, "aud")
		if audExists && aud != JWTAudience {
			logger.Warn("JWT token invalid audience",
				zap.String("traceId", traceId),
				zap.String("clientIP", c.ClientIP()),
			)
			ResponseError(c, logger, &model.APIError{
				Code:    http.StatusUnauthorized,
				Message: "Token 受众无效",
			}, http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 5. 设置上下文
		c.Set("username", username)

		jti, jtiExists := safeStringClaim(claims, "jti")
		if jtiExists && jti != "" {
			c.Set("jti", jti)
		}

		logger.Info("authentication successful",
			zap.String("traceId", traceId),
			zap.String("clientIP", c.ClientIP()),
			zap.String("path", c.Request.URL.Path),
		)

		c.Next()
	}
}

// VerifyToken 验证 JWT token 并返回 claims
func VerifyToken(tokenStr string, secret []byte) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrInvalidKey
	}
	if claims == nil {
		return nil, jwt.ErrInvalidKey
	}

	return claims, nil
}
