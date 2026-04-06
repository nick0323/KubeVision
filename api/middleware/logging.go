package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// sensitiveParams 需要脱敏的敏感参数名
var sensitiveParams = []string{
	"token", "password", "passwd", "pwd",
	"secret", "key", "api_key", "apikey",
	"access_token", "refresh_token", "auth",
}

// maskSensitiveQuery 脱敏敏感查询参数
// 支持脱敏：token, password, secret, key 等敏感字段
func maskSensitiveQuery(query string) string {
	if query == "" {
		return ""
	}

	// 解析查询字符串
	values, err := url.ParseQuery(query)
	if err != nil {
		// 解析失败，返回原始查询（可能包含敏感信息）
		return "***parse_error***"
	}

	// 脱敏敏感参数
	for _, param := range sensitiveParams {
		if values.Has(param) {
			values.Set(param, "***")
		}
	}

	return values.Encode()
}

// LoggingMiddleware 请求日志记录中间件
// 记录请求的完整生命周期，包括开始和结束
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		method := c.Request.Method
		traceId := c.GetString("traceId")

		// 脱敏敏感信息
		maskedQuery := maskSensitiveQuery(raw)

		// 记录请求开始日志（仅 DEBUG 级别）
		logger.Debug("request started",
			zap.String("traceId", traceId),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", maskedQuery),
			zap.String("clientIP", c.ClientIP()),
		)

		// 处理请求
		c.Next()

		// 计算处理时间
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		// 根据状态码选择日志级别
		logLevel := zap.InfoLevel
		if statusCode >= 500 {
			logLevel = zap.ErrorLevel
		} else if statusCode >= 400 {
			logLevel = zap.WarnLevel
		}

		// 记录请求完成日志
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

// TraceMiddleware 请求追踪中间件
// 从请求头获取或生成 X-Trace-ID，用于分布式追踪
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取 traceId，如果没有则生成新的
		traceId := c.GetHeader("X-Trace-ID")
		if traceId == "" {
			traceId = generateTraceID()
		}

		// 设置 traceId 到 context 中
		c.Set("traceId", traceId)

		// 在响应头中返回 traceId
		c.Header("X-Trace-ID", traceId)

		c.Next()
	}
}

// generateTraceID 生成追踪 ID
// 格式：YYYYMMDDHHMMSS-XXXXXXXX (时间戳 -8 位随机十六进制)
// 使用 8 字节随机数（64 位）降低碰撞概率
func generateTraceID() string {
	timestamp := time.Now().Format("20060102150405")

	randomBytes := make([]byte, 8) // 8 字节 = 64 位随机数
	if _, err := rand.Read(randomBytes); err != nil {
		// 随机数生成失败，使用时间戳 + 固定值
		return timestamp + "-00000000"
	}

	randomHex := hex.EncodeToString(randomBytes)
	return timestamp + "-" + randomHex
}
