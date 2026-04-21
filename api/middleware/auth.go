package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/nick0323/K8sVision/model"
	"go.uber.org/zap"
)

const (
	JWTIssuer             = "k8svision"
	JWTAudience           = "k8svision-client"
	webSocketAuthProtocol = "k8svision.auth"
)

type ConfigProvider interface {
	GetJWTSecret() []byte
}

var (
	jwtSecret     []byte
	jwtSecretOnce sync.Once
)

func SetJWTSecret(secret []byte) {
	jwtSecretOnce.Do(func() {
		jwtSecret = secret
	})
}

func GetJWTSecretFromConfig() ([]byte, error) {
	if jwtSecret == nil {
		return nil, fmt.Errorf("JWT secret not initialized")
	}
	return jwtSecret, nil
}

func JWTAuthMiddleware(logger *zap.Logger, configProvider ConfigProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		traceId := c.GetString("traceId")
		tokenStr := getTokenFromRequest(c)

		if tokenStr == "" {
			logger.Warn("missing authorization",
				zap.String("traceId", traceId),
				zap.String("clientIP", c.ClientIP()),
				zap.String("path", c.Request.URL.Path),
			)
			ResponseError(c, logger, &model.APIError{
				Code:    http.StatusUnauthorized,
				Message: "Unauthorized access",
				Details: "Missing Authorization header or token parameter",
			}, http.StatusUnauthorized)
			c.Abort()
			return
		}

		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

		if tokenStr == "" {
			logger.Warn("empty token", zap.String("traceId", traceId))
			ResponseError(c, logger, &model.APIError{
				Code:    http.StatusUnauthorized,
				Message: "Token cannot be empty",
			}, http.StatusUnauthorized)
			c.Abort()
			return
		}

		claims, err := verifyToken(tokenStr, configProvider)
		if err != nil {
			logger.Warn("token verification failed",
				zap.String("traceId", traceId),
				zap.Error(err),
			)
			ResponseError(c, logger, &model.APIError{
				Code:    http.StatusUnauthorized,
				Message: "Token verification failed",
			}, http.StatusUnauthorized)
			c.Abort()
			return
		}

		username := claims["username"].(string)
		c.Set("username", username)

		if jti, ok := claims["jti"].(string); ok && jti != "" {
			c.Set("jti", jti)
		}

		logger.Info("authentication successful",
			zap.String("traceId", traceId),
			zap.String("username", username),
		)

		c.Next()
	}
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
	if len(parts) >= 2 && parts[0] == webSocketAuthProtocol && parts[1] != "" {
		return parts[1]
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
		return nil, fmt.Errorf("config provider not initialized")
	}
	secret := provider.GetJWTSecret()
	if len(secret) == 0 {
		return nil, fmt.Errorf("JWT secret is empty")
	}
	return secret, nil
}

func validateClaims(claims jwt.MapClaims) error {
	if username, ok := claims["username"].(string); !ok || username == "" {
		return fmt.Errorf("token missing username")
	}

	if iss, ok := claims["iss"].(string); ok && iss != "" && iss != JWTIssuer {
		return fmt.Errorf("invalid token issuer")
	}

	if aud, ok := claims["aud"].(string); ok && aud != "" && aud != JWTAudience {
		return fmt.Errorf("invalid token audience")
	}

	return nil
}

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
	if !ok || claims == nil {
		return nil, jwt.ErrInvalidKey
	}

	return claims, nil
}
