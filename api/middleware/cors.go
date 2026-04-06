package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig CORS 配置
type CORSConfig struct {
	AllowOrigins     []string // 允许的来源列表
	AllowMethods     []string // 允许的方法
	AllowHeaders     []string // 允许的请求头
	ExposeHeaders    []string // 暴露的响应头
	AllowCredentials bool     // 是否允许携带凭证
	MaxAge           int      // 预检请求缓存时间（秒）
}

// DefaultCORSConfig 开发环境默认配置
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "X-Trace-ID"},
		AllowCredentials: false, // 开发环境默认不允许凭证
		MaxAge:           3600,  // 缓存 1 小时
	}
}

// CORSMiddleware CORS 中间件
// 参数：
//   - config: CORS 配置，如果为 nil 则使用默认配置
//
// 返回：gin.HandlerFunc
func CORSMiddleware(config *CORSConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultCORSConfig()
	}

	// 验证配置
	if config.AllowCredentials && len(config.AllowOrigins) == 1 && config.AllowOrigins[0] == "*" {
		// 如果允许凭证，不能使用 *
		config.AllowOrigins = []string{"http://localhost:3000"}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// 检查来源是否允许
		allowedOrigin := ""
		for _, o := range config.AllowOrigins {
			if o == "*" || o == origin {
				allowedOrigin = o
				break
			}
		}

		if allowedOrigin == "" && len(config.AllowOrigins) > 0 {
			// 如果没有匹配的来源，使用第一个（或拒绝）
			if config.AllowOrigins[0] != "*" {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			allowedOrigin = "*"
		}

		// 设置 CORS 头
		c.Header("Access-Control-Allow-Origin", allowedOrigin)

		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 暴露的响应头
		if len(config.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
		}

		// 预检请求缓存
		if config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
		}

		// 允许的方法和请求头
		if len(config.AllowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
		}

		if len(config.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
		}

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
