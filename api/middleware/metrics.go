package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kubevision_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kubevision_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	cacheSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "kubevision_cache_size",
			Help: "Current cache size",
		},
	)

	cacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "kubevision_cache_hits_total",
			Help: "Total number of cache hits",
		},
	)

	cacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "kubevision_cache_misses_total",
			Help: "Total number of cache misses",
		},
	)
)

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(time.Since(start).Seconds())
	}
}

func SetCacheMetrics(size int, hits, misses int64) {
	cacheSize.Set(float64(size))
	cacheHits.Add(float64(hits))
	cacheMisses.Add(float64(misses))
}
