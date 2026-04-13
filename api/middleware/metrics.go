package middleware

import (
	"net/http"
	"sync"
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

// RateLimiter 简单的速率限制器（滑动窗口实现）
// 比令牌桶更简单，性能更好
type RateLimiter struct {
	limit  int           // 窗口内最大请求数
	window time.Duration // 时间窗口
	mu     sync.Mutex
	reqs   []time.Time // 请求时间戳
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:  limit,
		window: window,
		reqs:   make([]time.Time, 0, limit),
	}
}

// Allow 检查是否允许请求
// 使用原地压缩优化，避免每次分配新切片
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// 原地压缩：移除窗口外的请求
	writeIdx := 0
	for readIdx := 0; readIdx < len(rl.reqs); readIdx++ {
		if rl.reqs[readIdx].After(windowStart) {
			rl.reqs[writeIdx] = rl.reqs[readIdx]
			writeIdx++
		}
	}
	rl.reqs = rl.reqs[:writeIdx]

	// 检查是否超过限制
	if len(rl.reqs) >= rl.limit {
		return false
	}

	// 添加当前请求
	rl.reqs = append(rl.reqs, now)
	return true
}

// RateLimitMiddleware 速率限制中间件
func RateLimitMiddleware(logger *zap.Logger, limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow() {
			if logger != nil {
				logger.Warn("rate limit exceeded",
					zap.String("path", c.Request.URL.Path),
					zap.String("clientIP", c.ClientIP()),
				)
			}

			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    http.StatusTooManyRequests,
				"message": "请求过于频繁，请稍后重试",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
