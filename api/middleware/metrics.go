package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MetricsRecorder 性能指标记录接口
type MetricsRecorder interface {
	RecordRequest(success bool, responseTime time.Duration)
	RecordError(err string)
}

// MetricsMiddleware 性能监控中间件
// 记录请求的响应时间和成功率
func MetricsMiddleware(recorder MetricsRecorder) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		// 计算响应时间
		responseTime := time.Since(start)

		// 判断请求是否成功
		success := c.Writer.Status() < 400

		// 记录请求统计
		recorder.RecordRequest(success, responseTime)

		// 记录错误
		if !success && len(c.Errors) > 0 {
			recorder.RecordError(c.Errors.String())
		}
	}
}

// ConcurrencyMiddleware 并发控制中间件
// 限制同时处理的请求数量，防止系统过载
func ConcurrencyMiddleware(logger *zap.Logger, maxConcurrency int) gin.HandlerFunc {
	semaphore := make(chan struct{}, maxConcurrency)

	return func(c *gin.Context) {
		select {
		case semaphore <- struct{}{}:
			defer func() { <-semaphore }()
			c.Next()
		default:
			// 并发数超限，记录日志
			if logger != nil {
				logger.Warn("concurrency limit exceeded",
					zap.String("path", c.Request.URL.Path),
					zap.String("clientIP", c.ClientIP()),
				)
			}

			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    http.StatusServiceUnavailable,
				"message": "服务繁忙，请稍后重试",
			})
			c.Abort()
		}
	}
}
