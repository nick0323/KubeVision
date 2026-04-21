package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "X-Trace-ID"},
		AllowCredentials: false,
		MaxAge:           3600,
	}
}

func CORSMiddleware(config *CORSConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultCORSConfig()
	}

	if config.AllowCredentials && len(config.AllowOrigins) == 1 && config.AllowOrigins[0] == "*" {
		config.AllowOrigins = []string{"http://localhost:3000"}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		allowedOrigin := ""
		for _, o := range config.AllowOrigins {
			if o == "*" || o == origin {
				allowedOrigin = o
				break
			}
		}

		if allowedOrigin == "" && len(config.AllowOrigins) > 0 {
			if config.AllowOrigins[0] != "*" {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			allowedOrigin = "*"
		}

		c.Header("Access-Control-Allow-Origin", allowedOrigin)

		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if len(config.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
		}

		if config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
		}

		if len(config.AllowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
		}

		if len(config.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
