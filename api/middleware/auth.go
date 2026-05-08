package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/nick0323/K8sVision/model"
	"go.uber.org/zap"
)

const (
	JWTIssuer   = "k8svision"
	JWTAudience = "k8svision-client"
)

type ConfigProvider interface {
	GetJWTSecret() []byte
}

type JWTMiddleware struct {
	secret    []byte
	logger    *zap.Logger
	blacklist *TokenBlacklist
}

func NewJWTMiddleware(secret []byte, logger *zap.Logger) *JWTMiddleware {
	return &JWTMiddleware{
		secret:    secret,
		logger:    logger,
		blacklist: NewTokenBlacklist(10000), // 最多存储 10000 个黑名单 token
	}
}

// GetBlacklist 返回 token 黑名单实例（用于 logout 接口）
func (m *JWTMiddleware) GetBlacklist() *TokenBlacklist {
	return m.blacklist
}

// Close 关闭中间件，释放后台资源（黑名单清理 goroutine）
func (m *JWTMiddleware) Close() {
	if m.blacklist != nil {
		m.blacklist.Close()
	}
}

func (m *JWTMiddleware) AuthMiddleware(configProvider ConfigProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		traceId := c.GetString("traceId")
		tokenStr := getTokenFromRequest(c)

		if tokenStr == "" {
			m.logger.Warn("missing authorization",
				zap.String("traceId", traceId),
				zap.String("clientIP", c.ClientIP()),
			)
			ResponseError(c, m.logger, &model.APIError{
				Code:    http.StatusUnauthorized,
				Message: "Unauthorized access",
			}, http.StatusUnauthorized)
			c.Abort()
			return
		}

		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
		if tokenStr == "" {
			m.logger.Warn("empty token", zap.String("traceId", traceId))
			ResponseError(c, m.logger, &model.APIError{
				Code:    http.StatusUnauthorized,
				Message: "Token cannot be empty",
			}, http.StatusUnauthorized)
			c.Abort()
			return
		}

		username, jti, err := m.verifyAndSetClaims(c, tokenStr, configProvider)
		if err != nil {
			m.logger.Warn("token verification failed",
				zap.String("traceId", traceId),
				zap.Error(err),
			)
			ResponseError(c, m.logger, &model.APIError{
				Code:    http.StatusUnauthorized,
				Message: "Token verification failed",
			}, http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 检查 token 是否在黑名单中
		if jti != "" && m.blacklist.IsBlacklisted(jti) {
			m.logger.Warn("token is blacklisted",
				zap.String("traceId", traceId),
				zap.String("jti", jti),
			)
			ResponseError(c, m.logger, &model.APIError{
				Code:    http.StatusUnauthorized,
				Message: "Token has been revoked",
			}, http.StatusUnauthorized)
			c.Abort()
			return
		}

		m.logger.Info("authentication successful",
			zap.String("traceId", traceId),
			zap.String("username", username),
		)
		c.Next()
	}
}

func (m *JWTMiddleware) verifyAndSetClaims(c *gin.Context, tokenStr string, configProvider ConfigProvider) (string, string, error) {
	claims, err := verifyToken(tokenStr, configProvider)
	if err != nil {
		return "", "", err
	}

	username, ok := claims["username"].(string)
	if !ok || username == "" {
		return "", "", errors.New("token missing username")
	}

	c.Set("username", username)

	var jti string
	if jtiClaim, ok := claims["jti"].(string); ok && jtiClaim != "" {
		jti = jtiClaim
		c.Set("jti", jti)
	}

	return username, jti, nil
}

func getTokenFromRequest(c *gin.Context) string {
	if token := c.GetHeader("Authorization"); token != "" {
		return token
	}
	if token := extractTokenFromWebSocket(c.GetHeader("Sec-WebSocket-Protocol")); token != "" {
		return token
	}
	return c.Query("token")
}

func extractTokenFromWebSocket(headerValue string) string {
	if headerValue == "" {
		return ""
	}
	parts := strings.Split(headerValue, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	if parts[0] == "k8svision.auth" {
		if len(parts) >= 2 && parts[1] != "" {
			return parts[1]
		}
		return ""
	}
	return strings.TrimSpace(headerValue)
}

func verifyToken(tokenStr string, configProvider ConfigProvider) (jwt.MapClaims, error) {
	secret, err := getJWTSecret(configProvider)
	if err != nil {
		return nil, err
	}

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
	if !ok || claims == nil {
		return nil, jwt.ErrInvalidKey
	}

	if err := validateClaims(claims); err != nil {
		return nil, err
	}

	return claims, nil
}

func getJWTSecret(provider ConfigProvider) ([]byte, error) {
	if provider == nil {
		return nil, errors.New("config provider not initialized")
	}
	secret := provider.GetJWTSecret()
	if len(secret) == 0 {
		return nil, errors.New("JWT secret not configured")
	}
	return secret, nil
}

func validateClaims(claims jwt.MapClaims) error {
	username, ok := claims["username"].(string)
	if !ok || username == "" {
		return errors.New("token missing username")
	}

	iss, ok := claims["iss"].(string)
	if !ok || iss == "" {
		return errors.New("token missing issuer")
	}
	if iss != JWTIssuer {
		return errors.New("invalid token issuer")
	}

	aud, ok := claims["aud"].(string)
	if !ok || aud == "" {
		return errors.New("token missing audience")
	}
	if aud != JWTAudience {
		return errors.New("invalid token audience")
	}

	return nil
}

func VerifyToken(tokenStr string, secret []byte) (jwt.MapClaims, error) {
	if len(secret) == 0 {
		return nil, errors.New("JWT secret not configured")
	}

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
	if !ok || claims == nil {
		return nil, jwt.ErrInvalidKey
	}

	return claims, nil
}
