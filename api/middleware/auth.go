package middleware

import (
	"fmt"
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

// safeStringClaim 安全地从JWT claims中获取字符串值
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
				Code:    model.CodeUnauthorized,
				Message: model.GetErrorMessage(model.CodeUnauthorized),
				Details: "缺少Authorization头",
			}, 401)
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
				Code:    model.CodeUnauthorized,
				Message: model.GetErrorMessage(model.CodeUnauthorized),
				Details: "Authorization格式错误，应为Bearer token",
			}, 401)
			c.Abort()
			return
		}

		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

		// 添加基本格式验证
		if tokenStr == "" {
			logger.Warn("empty token after prefix removal",
				zap.String("traceId", traceId),
				zap.String("clientIP", c.ClientIP()),
			)
			ResponseError(c, logger, &model.APIError{
				Code:    model.CodeAuthError,
				Message: model.GetErrorMessage(model.CodeAuthError),
				Details: "Token不能为空",
			}, 401)
			c.Abort()
			return
		}

		segments := strings.Split(tokenStr, ".")
		if len(segments) != 3 {
			logger.Warn("invalid token format - wrong number of segments",
				zap.String("traceId", traceId),
				zap.String("clientIP", c.ClientIP()),
				zap.Int("segments", len(segments)),
				zap.String("token", maskToken(tokenStr)), // 遮蔽敏感信息
			)
			ResponseError(c, logger, &model.APIError{
				Code:    model.CodeAuthError,
				Message: model.GetErrorMessage(model.CodeAuthError),
				Details: fmt.Sprintf("Token格式错误：期望3个段，实际%d个段", len(segments)),
			}, 401)
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
				zap.String("maskedToken", maskToken(tokenStr)), // 遮蔽敏感信息
				zap.Error(err),
			)
			ResponseError(c, logger, &model.APIError{
				Code:    model.CodeAuthError,
				Message: model.GetErrorMessage(model.CodeAuthError),
				Details: "Token解析失败: " + err.Error(),
			}, 401)
			c.Abort()
			return
		}

		if !token.Valid {
			logger.Warn("invalid jwt token",
				zap.String("traceId", traceId),
				zap.String("clientIP", c.ClientIP()),
				zap.String("maskedToken", maskToken(tokenStr)), // 遮蔽敏感信息
			)
			ResponseError(c, logger, &model.APIError{
				Code:    model.CodeAuthError,
				Message: model.GetErrorMessage(model.CodeAuthError),
				Details: "Token无效或已过期",
			}, 401)
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && claims != nil {
			// 添加额外的安全检查
			defer func() {
				if r := recover(); r != nil {
					logger.Error("JWT claims processing panic recovered",
						zap.String("traceId", traceId),
						zap.String("clientIP", c.ClientIP()),
					)
					ResponseError(c, logger, &model.APIError{
						Code:    model.CodeAuthError,
						Message: model.GetErrorMessage(model.CodeAuthError),
						Details: "Token处理失败",
					}, 401)
					c.Abort()
				}
			}()
			
			// 验证必要字段
			username, usernameExists := safeStringClaim(claims, "username")

			if !usernameExists || username == "" {
				logger.Warn("JWT token missing username",
					zap.String("traceId", traceId),
					zap.String("clientIP", c.ClientIP()),
				)
				ResponseError(c, logger, &model.APIError{
					Code:    model.CodeAuthError,
					Message: model.GetErrorMessage(model.CodeAuthError),
					Details: "Token缺少用户名信息",
				}, 401)
				c.Abort()
				return
			}

			// 安全地获取可选字段
			iss, issExists := safeStringClaim(claims, "iss")
			aud, audExists := safeStringClaim(claims, "aud")
			jti, jtiExists := safeStringClaim(claims, "jti")

			// 验证签发者和受众（如果存在）
			if issExists && iss != "k8svision" {
				logger.Warn("JWT token invalid issuer",
					zap.String("traceId", traceId),
					zap.String("clientIP", c.ClientIP()),
					zap.String("issuer", iss),
				)
				ResponseError(c, logger, &model.APIError{
					Code:    model.CodeAuthError,
					Message: model.GetErrorMessage(model.CodeAuthError),
					Details: "Token签发者无效",
				}, 401)
				c.Abort()
				return
			}

			if audExists && aud != "k8svision-client" {
				logger.Warn("JWT token invalid audience",
					zap.String("traceId", traceId),
					zap.String("clientIP", c.ClientIP()),
					zap.String("audience", aud),
				)
				ResponseError(c, logger, &model.APIError{
					Code:    model.CodeAuthError,
					Message: model.GetErrorMessage(model.CodeAuthError),
					Details: "Token受众无效",
				}, 401)
				c.Abort()
				return
			}

			// 验证时间相关声明
			if exp, ok := claims["exp"]; ok {
				if expFloat, ok := exp.(float64); ok {
					if int64(expFloat) < time.Now().Unix() {
						logger.Warn("JWT token expired",
							zap.String("traceId", traceId),
							zap.String("clientIP", c.ClientIP()),
						)
						ResponseError(c, logger, &model.APIError{
							Code:    model.CodeAuthError,
							Message: model.GetErrorMessage(model.CodeAuthError),
							Details: "Token已过期",
						}, 401)
						c.Abort()
						return
					}
				}
			}

			c.Set("username", username)

			// 安全地设置JTI（如果存在）
			if jtiExists && jti != "" {
				c.Set("jti", jti) // 保存JTI用于撤销
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

// maskToken 遮蔽敏感的token信息，只保留部分字符
func maskToken(token string) string {
	if len(token) <= 10 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
