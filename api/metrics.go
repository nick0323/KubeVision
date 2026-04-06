package api

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/model"
	"github.com/nick0323/K8sVision/monitor"
	"go.uber.org/zap"
)

// ==================== 响应结构体 ====================

// MetricsResponse 指标响应
type MetricsResponse struct {
	Timestamp time.Time              `json:"timestamp"`
	System    SystemMetrics          `json:"system"`
	Business  map[string]interface{} `json:"business"`
	Summary   MetricsSummary         `json:"summary"`
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	CPU         CPUUsage        `json:"cpu"`
	Memory      MemoryUsage     `json:"memory"`
	Network     NetworkIO       `json:"network"`
	Connections ConnectionStats `json:"connections"`
	CollectedAt time.Time       `json:"collected_at"`
}

// CPUUsage CPU 使用率
type CPUUsage struct {
	UsagePercent float64 `json:"usage_percent"`
	Cores        int     `json:"cores"`
}

// MemoryUsage 内存使用
type MemoryUsage struct {
	UsedMB       float64 `json:"used_mb"`
	TotalMB      float64 `json:"total_mb"`
	UsagePercent float64 `json:"usage_percent"`
}

// NetworkIO 网络 IO
type NetworkIO struct {
	BytesIn  int64 `json:"bytes_in"`
	BytesOut int64 `json:"bytes_out"`
}

// ConnectionStats 连接统计
type ConnectionStats struct {
	Active int `json:"active"`
	Idle   int `json:"idle"`
}

// MetricsSummary 指标摘要
type MetricsSummary struct {
	TotalCount    int `json:"total_count"`
	SystemCount   int `json:"system_count"`
	BusinessCount int `json:"business_count"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    HealthStatus `json:"status"`
	Score     float64      `json:"score"`
	Timestamp time.Time    `json:"timestamp"`
}

// HealthStatus 健康状态
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// RegisterMetrics 注册指标相关路由
func RegisterMetrics(r *gin.RouterGroup, logger *zap.Logger) {
	r.GET("/metrics", getMetrics(logger))
	r.GET("/metrics/business", getBusinessMetrics(logger))
	r.GET("/metrics/system", getSystemMetrics(logger))
	r.GET("/metrics/health", getHealthMetrics(logger))
}

// getMetrics 获取所有指标
func getMetrics(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		metricsManager := monitor.GetMetricsManager()
		if metricsManager == nil {
			logger.Warn("metrics manager is nil")
			middleware.ResponseError(c, logger, nil, http.StatusInternalServerError)
			return
		}

		// 获取系统指标
		rawMetrics := metricsManager.GetAllMetrics()
		systemMetrics := parseSystemMetrics(rawMetrics, logger)

		// 获取业务指标
		businessCollector := monitor.GetBusinessMetricsCollector()
		if businessCollector == nil {
			logger.Warn("business metrics collector is nil")
			middleware.ResponseError(c, logger, nil, http.StatusInternalServerError)
			return
		}

		businessMetrics := businessCollector.GetMetricsJSON()

		// 构建响应
		response := MetricsResponse{
			Timestamp: time.Now(),
			System:    systemMetrics,
			Business:  businessMetrics,
			Summary: MetricsSummary{
				TotalCount:    1,
				SystemCount:   1,
				BusinessCount: len(businessMetrics),
			},
		}

		middleware.ResponseSuccess(c, response, "指标获取成功", nil)
	}
}

// getBusinessMetrics 获取业务指标
func getBusinessMetrics(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		collector := monitor.GetBusinessMetricsCollector()
		if collector == nil {
			logger.Warn("business metrics collector is nil")
			middleware.ResponseError(c, logger, nil, http.StatusInternalServerError)
			return
		}

		businessMetrics := collector.GetMetricsJSON()
		middleware.ResponseSuccess(c, businessMetrics, "业务指标获取成功", nil)
	}
}

// getSystemMetrics 获取系统指标
func getSystemMetrics(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		metricsManager := monitor.GetMetricsManager()
		if metricsManager == nil {
			logger.Warn("metrics manager is nil")
			middleware.ResponseError(c, logger, nil, http.StatusInternalServerError)
			return
		}

		rawMetrics := metricsManager.GetAllMetrics()
		systemMetrics := parseSystemMetrics(rawMetrics, logger)

		response := map[string]interface{}{
			"metrics": systemMetrics,
		}

		middleware.ResponseSuccess(c, response, "系统指标获取成功", nil)
	}
}

// getHealthMetrics 获取健康指标
func getHealthMetrics(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查各组件状态
		status, score := calculateHealth(logger)

		response := HealthResponse{
			Status:    status,
			Score:     score,
			Timestamp: time.Now(),
		}

		// 根据健康状态设置 HTTP 状态码
		httpStatus := http.StatusOK
		if status == HealthStatusUnhealthy {
			httpStatus = http.StatusServiceUnavailable
		}

		c.JSON(httpStatus, gin.H{
			"code":    model.CodeSuccess,
			"message": "健康指标获取成功",
			"data":    response,
		})
	}
}

// calculateHealth 计算健康状态
func calculateHealth(logger *zap.Logger) (HealthStatus, float64) {
	score := 100.0

	// 检查系统指标
	metricsManager := monitor.GetMetricsManager()
	if metricsManager == nil || len(metricsManager.GetAllMetrics()) == 0 {
		score -= 30
		logger.Warn("health check: metrics unavailable")
	}

	// 检查业务指标收集器
	businessCollector := monitor.GetBusinessMetricsCollector()
	if businessCollector == nil {
		score -= 20
		logger.Warn("health check: business metrics degraded")
	}

	// 确定状态
	var status HealthStatus
	if score >= 90 {
		status = HealthStatusHealthy
	} else if score >= 70 {
		status = HealthStatusDegraded
	} else {
		status = HealthStatusUnhealthy
	}

	logger.Info("health check completed",
		zap.String("status", string(status)),
		zap.Float64("score", score),
	)

	return status, score
}

// parseSystemMetrics 解析系统指标
func parseSystemMetrics(rawMetrics map[string]interface{}, logger *zap.Logger) SystemMetrics {
	systemMetrics := SystemMetrics{
		CollectedAt: time.Now(),
	}

	// 从 runtime 获取系统信息
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 设置内存指标
	systemMetrics.Memory.UsedMB = float64(memStats.Alloc) / 1024 / 1024
	systemMetrics.Memory.TotalMB = float64(memStats.Sys) / 1024 / 1024
	if memStats.Sys > 0 {
		systemMetrics.Memory.UsagePercent = float64(memStats.Alloc) / float64(memStats.Sys) * 100
	}

	// 设置 CPU 核心数
	systemMetrics.CPU.Cores = runtime.NumCPU()

	// 从 rawMetrics 中提取其他指标
	if connections, exists := rawMetrics["currentConnections"]; exists {
		if active, ok := connections.(float64); ok {
			systemMetrics.Connections.Active = int(active)
		}
	}

	logger.Debug("system metrics parsed",
		zap.Float64("memory_used_mb", systemMetrics.Memory.UsedMB),
		zap.Int("cpu_cores", systemMetrics.CPU.Cores),
	)

	return systemMetrics
}
