package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/nick0323/K8sVision/model"
	"go.uber.org/zap"
)

type ConfigProvider interface {
	GetJWTSecret() []byte
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

		if tokenStr == "" {
			logger.Warn("missing authorization header",
				zap.String("traceId", traceId),
				zap.String("clientIP", c.ClientIP()),
				zap.String("path", c.Request.URL.Path),
			)
			ResponseError(c, logger, &model.APIError{
				Code:    http.StatusUnauthorized,
				Message: "未授权访问",
				Details: "缺少 Authorization 头",
			}, http.StatusUnauthorized)
			c.Abort()
			return
		}

		if !strings.HasPrefix(tokenStr, "Bearer ") {
			logger.Warn("invalid authorization format",
				zap.String("traceId", traceId),
				zap.String("clientIP", c.ClientIP()),
				zap.String("header", tokenStr),
			)
			ResponseError(c, logger, &model.APIError{
				Code:    http.StatusUnauthorized,
				Message: "Authorization 格式错误",
				Details: "应为 Bearer token",
			}, http.StatusUnauthorized)
			c.Abort()
			return
		}

		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

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

		segments := strings.Split(tokenStr, ".")
		if len(segments) != 3 {
			logger.Warn("invalid token format",
				zap.String("traceId", traceId),
				zap.String("clientIP", c.ClientIP()),
				zap.Int("segments", len(segments)),
			)
			ResponseError(c, logger, &model.APIError{
				Code:    http.StatusUnauthorized,
				Message: "Token 格式错误",
				Details: fmt.Sprintf("期望 3 个段，实际%d个段", len(segments)),
			}, http.StatusUnauthorized)
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
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
				Details: err.Error(),
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

		if claims, ok := token.Claims.(jwt.MapClaims); ok && claims != nil {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("JWT claims processing panic recovered",
						zap.String("traceId", traceId),
						zap.String("clientIP", c.ClientIP()),
					)
					ResponseError(c, logger, &model.APIError{
						Code:    http.StatusUnauthorized,
						Message: "Token 处理失败",
					}, http.StatusUnauthorized)
					c.Abort()
				}
			}()

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

			iss, issExists := safeStringClaim(claims, "iss")
			aud, audExists := safeStringClaim(claims, "aud")
			jti, jtiExists := safeStringClaim(claims, "jti")

			if issExists && iss != "k8svision" {
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

			if audExists && aud != "k8svision-client" {
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

			if exp, ok := claims["exp"]; ok {
				if expFloat, ok := exp.(float64); ok {
					if int64(expFloat) < time.Now().Unix() {
						logger.Warn("JWT token expired",
							zap.String("traceId", traceId),
							zap.String("clientIP", c.ClientIP()),
						)
						ResponseError(c, logger, &model.APIError{
							Code:    http.StatusUnauthorized,
							Message: "Token 已过期",
						}, http.StatusUnauthorized)
						c.Abort()
						return
					}
				}
			}

			c.Set("username", username)
			if jtiExists && jti != "" {
				c.Set("jti", jti)
			}
		}

		logger.Info("authentication successful",
			zap.String("traceId", traceId),
			zap.String("clientIP", c.ClientIP()),
			zap.String("path", c.Request.URL.Path),
		)

		c.Next()
	}
}

func maskToken(token string) string {
	if len(token) <= 10 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
