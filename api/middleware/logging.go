package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var sensitiveParams = []string{
	"token", "password", "passwd", "pwd",
	"secret", "key", "api_key", "apikey",
	"access_token", "refresh_token", "auth",
}

func MaskSensitiveQuery(query string) string {
	if query == "" {
		return ""
	}

	values, err := url.ParseQuery(query)
	if err != nil {
		return "***parse_error***"
	}

	for _, param := range sensitiveParams {
		if values.Has(param) {
			values.Set(param, "***")
		}
	}

	return values.Encode()
}

func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		method := c.Request.Method
		traceId := c.GetString("traceId")

		maskedQuery := MaskSensitiveQuery(raw)

		logger.Debug("request started",
			zap.String("traceId", traceId),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", maskedQuery),
			zap.String("clientIP", c.ClientIP()),
		)

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		logLevel := zap.InfoLevel
		if statusCode >= 500 {
			logLevel = zap.ErrorLevel
		} else if statusCode >= 400 {
			logLevel = zap.WarnLevel
		}

		logger.Log(logLevel, "request completed",
			zap.String("traceId", traceId),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("statusCode", statusCode),
			zap.Duration("latency", latency),
			zap.Int("bodySize", bodySize),
			zap.Strings("errors", c.Errors.Errors()),
		)
	}
}

func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceId := c.GetHeader("X-Trace-ID")
		if traceId == "" {
			traceId = generateTraceID()
		}

		c.Set("traceId", traceId)
		c.Header("X-Trace-ID", traceId)

		c.Next()
	}
}

func generateTraceID() string {
	timestamp := time.Now().Format("20060102150405")

	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return timestamp + "-00000000"
	}

	return timestamp + "-" + hex.EncodeToString(randomBytes)
}
