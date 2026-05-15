package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiter *rate.Limiter
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(r, b),
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    http.StatusTooManyRequests,
				"message": "too many requests, please try again later",
			})
			return
		}
		c.Next()
	}
}

type IPRateLimiter struct {
	limiters      sync.Map
	rate          rate.Limit
	burst         int
	cleanupTicker *time.Ticker
	stopCh        chan struct{}
	closeOnce     sync.Once
	wg            sync.WaitGroup
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	rl := &IPRateLimiter{
		rate:          r,
		burst:         b,
		cleanupTicker: time.NewTicker(10 * time.Minute),
		stopCh:        make(chan struct{}),
	}
	rl.wg.Add(1)
	go rl.cleanupWorker()
	return rl
}

func (i *IPRateLimiter) Close() {
	i.closeOnce.Do(func() {
		close(i.stopCh)
	})
	i.wg.Wait()
}

func (i *IPRateLimiter) cleanupWorker() {
	defer i.wg.Done()
	for {
		select {
		case <-i.cleanupTicker.C:
			i.limiters.Range(func(key, value any) bool {
				limiter := value.(*rate.Limiter)
				if limiter.Tokens() >= float64(i.burst) {
					i.limiters.Delete(key)
				}
				return true
			})
		case <-i.stopCh:
			i.cleanupTicker.Stop()
			return
		}
	}
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	actual, _ := i.limiters.LoadOrStore(ip, rate.NewLimiter(i.rate, i.burst))
	limiter, _ := actual.(*rate.Limiter)
	return limiter
}

func (i *IPRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		limiter := i.GetLimiter(c.ClientIP())
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    http.StatusTooManyRequests,
				"message": "too many requests, please try again later",
			})
			return
		}
		c.Next()
	}
}
