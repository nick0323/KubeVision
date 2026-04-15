package monitor

import (
	"context"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ==================== 业务指标常量 ====================

const (
	// 指标上报间隔
	MetricsReportInterval = 10 * time.Second

	// 百分位计算采样大小
	PercentileSampleSize = 1000
)

// ==================== 业务指标类型 ====================

// BusinessMetrics 业务指标集合
type BusinessMetrics struct {
	// API 指标
	API APIMetrics `json:"api"`

	// 认证指标
	Auth AuthMetrics `json:"auth"`

	// Kubernetes 资源指标
	K8s K8sResourceMetrics `json:"k8s"`

	// 缓存指标
	Cache CacheMetrics `json:"cache"`

	// 系统指标
	System SystemMetrics `json:"system"`

	// 自定义业务指标
	Custom map[string]interface{} `json:"custom,omitempty"`
}

// APIMetrics API 相关指标
type APIMetrics struct {
	// 请求统计
	TotalRequests     int64   `json:"totalRequests"`     // 总请求数
	RequestsPerSecond float64 `json:"requestsPerSecond"` // 每秒请求数

	// 按方法统计
	GETRequests    int64 `json:"getRequests"`
	POSTRequests   int64 `json:"postRequests"`
	PUTRequests    int64 `json:"putRequests"`
	DELETERequests int64 `json:"deleteRequests"`

	// 按状态码统计
	Success2xx  int64 `json:"success2xx"`  // 2xx 成功
	Redirect3xx int64 `json:"redirect3xx"` // 3xx 重定向
	Client4xx   int64 `json:"client4xx"`   // 4xx 客户端错误
	Server5xx   int64 `json:"server5xx"`   // 5xx 服务器错误

	// 响应时间（毫秒）
	ResponseTimeP50 float64 `json:"responseTimeP50"` // P50 延迟
	ResponseTimeP90 float64 `json:"responseTimeP90"` // P90 延迟
	ResponseTimeP99 float64 `json:"responseTimeP99"` // P99 延迟
	AvgResponseTime float64 `json:"avgResponseTime"` // 平均延迟

	// 慢请求统计
	SlowRequests int64 `json:"slowRequests"` // 超过阈值的请求数
}

// AuthMetrics 认证相关指标
type AuthMetrics struct {
	// 登录统计
	LoginAttempts    int64   `json:"loginAttempts"`    // 登录尝试次数
	LoginSuccess     int64   `json:"loginSuccess"`     // 登录成功次数
	LoginFailed      int64   `json:"loginFailed"`      // 登录失败次数
	LoginSuccessRate float64 `json:"loginSuccessRate"` // 登录成功率

	// 账户锁定
	AccountLocked    int64 `json:"accountLocked"`    // 被锁定的账户数
	AccountUnlockeds int64 `json:"accountUnlockeds"` // 已解锁的账户数

	// Token 统计
	TokenIssued    int64 `json:"tokenIssued"`    // 签发的 Token 数
	TokenRefreshed int64 `json:"tokenRefreshed"` // 刷新的 Token 数
	TokenExpired   int64 `json:"tokenExpired"`   // 过期的 Token 数
	TokenInvalid   int64 `json:"tokenInvalid"`   // 无效的 Token 数

	// 会话统计
	ActiveSessions int64 `json:"activeSessions"` // 活跃会话数
	MaxSessions    int64 `json:"maxSessions"`    // 最大会话数
}

// K8sResourceMetrics Kubernetes 资源指标
type K8sResourceMetrics struct {
	// 资源数量
	Namespaces  int64 `json:"namespaces"`  // 命名空间数
	Nodes       int64 `json:"nodes"`       // 节点数
	Pods        int64 `json:"pods"`        // Pod 总数
	RunningPods int64 `json:"runningPods"` // 运行中的 Pod
	PendingPods int64 `json:"pendingPods"` // 等待中的 Pod
	FailedPods  int64 `json:"failedPods"`  // 失败的 Pod

	// 工作负载
	Deployments  int64 `json:"deployments"`  // Deployment 数
	StatefulSets int64 `json:"statefulSets"` // StatefulSet 数
	DaemonSets   int64 `json:"daemonSets"`   // DaemonSet 数
	Jobs         int64 `json:"jobs"`         // Job 数
	CronJobs     int64 `json:"cronJobs"`     // CronJob 数

	// 网络
	Services  int64 `json:"services"`  // Service 数
	Ingresses int64 `json:"ingresses"` // Ingress 数
	Endpoints int64 `json:"endpoints"` // Endpoints 数

	// 存储
	PVCs           int64 `json:"pvcs"`           // PVC 数
	PVs            int64 `json:"pvs"`            // PV 数
	StorageClasses int64 `json:"storageClasses"` // StorageClass 数

	// 配置
	ConfigMaps int64 `json:"configMaps"` // ConfigMap 数
	Secrets    int64 `json:"secrets"`    // Secret 数

	// API 调用
	APICallsTotal   int64            `json:"apiCallsTotal"`   // 总 API 调用数
	APICallsSuccess int64            `json:"apiCallsSuccess"` // 成功的 API 调用
	APICallsFailed  int64            `json:"apiCallsFailed"`  // 失败的 API 调用
	APILatencyP99   float64          `json:"apiLatencyP99"`   // API 调用 P99 延迟（毫秒）
	apiLatencyState *apiLatencyState // 延迟采样状态（使用指针避免拷贝问题）
}

// apiLatencyState K8s API 延迟采样状态
type apiLatencyState struct {
	mu      sync.Mutex
	samples []float64
	index   int
	full    bool
}

// CacheMetrics 缓存相关指标
type CacheMetrics struct {
	// 命中统计
	Hits    int64   `json:"hits"`    // 命中次数
	Misses  int64   `json:"misses"`  // 未命中次数
	HitRate float64 `json:"hitRate"` // 命中率

	// 容量统计
	Size        int64   `json:"size"`        // 当前缓存项数
	MaxSize     int64   `json:"maxSize"`     // 最大缓存项数
	Utilization float64 `json:"utilization"` // 使用率

	// 淘汰统计
	Evictions int64 `json:"evictions"` // 淘汰的项数
	Expired   int64 `json:"expired"`   // 过期的项数

	// 内存使用
	MemoryUsage int64 `json:"memoryUsage"` // 内存使用（字节）
}

// SystemMetrics 系统相关指标
type SystemMetrics struct {
	// 运行时间
	StartTime     time.Time `json:"startTime"`     // 启动时间
	UptimeSeconds int64     `json:"uptimeSeconds"` // 运行时长（秒）

	// 内存（字节）
	MemoryAlloc int64 `json:"memoryAlloc"` // 已分配内存
	MemoryTotal int64 `json:"memoryTotal"` // 总内存
	MemorySys   int64 `json:"memorySys"`   // 系统内存

	// Goroutine
	Goroutines int64 `json:"goroutines"` // Goroutine 数量

	// CPU
	CPUUsagePercent float64 `json:"cpuUsagePercent"` // CPU 使用率

	// GC
	GCPausesMs float64 `json:"gcPausesMs"` // GC 暂停时间（毫秒）
	GCCount    uint32  `json:"gcCount"`    // GC 次数
}

// ==================== 指标收集器 ====================

// MetricsCollector 指标收集器
type MetricsCollector struct {
	metrics BusinessMetrics
	mutex   sync.RWMutex
	logger  *zap.Logger
	ctx     context.Context
	cancel  context.CancelFunc

	// 响应时间采样（用于百分位计算）
	responseTimes []float64
	rtMutex       sync.Mutex
	rtIndex       int
	rtFull        bool

	// 速率计算
	lastTotalRequests int64
	lastCheckTime     time.Time
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector(logger *zap.Logger) *MetricsCollector {
	ctx, cancel := context.WithCancel(context.Background())

	collector := &MetricsCollector{
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
		responseTimes: make([]float64, PercentileSampleSize),
		lastCheckTime: time.Now(),
		metrics: BusinessMetrics{
			System: SystemMetrics{
				StartTime: time.Now(),
			},
			Custom: make(map[string]interface{}),
		},
	}

	return collector
}

// ==================== API 指标记录 ====================

// RecordAPIRequest 记录 API 请求
func (c *MetricsCollector) RecordAPIRequest(method string, statusCode int, durationMs float64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 总请求数
	c.metrics.API.TotalRequests++

	// 按方法统计
	switch method {
	case "GET":
		c.metrics.API.GETRequests++
	case "POST":
		c.metrics.API.POSTRequests++
	case "PUT":
		c.metrics.API.PUTRequests++
	case "DELETE":
		c.metrics.API.DELETERequests++
	}

	// 按状态码统计
	switch {
	case statusCode >= 200 && statusCode < 300:
		c.metrics.API.Success2xx++
	case statusCode >= 300 && statusCode < 400:
		c.metrics.API.Redirect3xx++
	case statusCode >= 400 && statusCode < 500:
		c.metrics.API.Client4xx++
	case statusCode >= 500:
		c.metrics.API.Server5xx++
	}

	// 慢请求检测（超过 1 秒）
	if durationMs > 1000 {
		c.metrics.API.SlowRequests++
	}

	// 记录响应时间（用于百分位计算）
	c.recordResponseTime(durationMs)
}

// recordResponseTime 记录响应时间（环形缓冲区）
func (c *MetricsCollector) recordResponseTime(durationMs float64) {
	c.rtMutex.Lock()
	defer c.rtMutex.Unlock()

	c.responseTimes[c.rtIndex] = durationMs
	c.rtIndex = (c.rtIndex + 1) % PercentileSampleSize
	if c.rtIndex == 0 {
		c.rtFull = true
	}
}

// calculatePercentile 计算百分位（通用函数）
func calculatePercentile(samples []float64, p float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	sort.Float64s(samples)
	index := int(float64(len(samples)-1) * p)
	if index >= len(samples) {
		index = len(samples) - 1
	}
	return samples[index]
}

// calculateResponseTimePercentile 计算 API 响应时间百分位
func (c *MetricsCollector) calculateResponseTimePercentile(p float64) float64 {
	c.rtMutex.Lock()
	defer c.rtMutex.Unlock()

	var samples []float64
	if c.rtFull {
		samples = make([]float64, len(c.responseTimes))
		copy(samples, c.responseTimes)
	} else {
		samples = c.responseTimes[:c.rtIndex]
	}

	return calculatePercentile(samples, p)
}

// UpdateAPIRates 更新 API 速率指标
func (c *MetricsCollector) UpdateAPIRates() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(c.lastCheckTime).Seconds()

	if elapsed > 0 {
		requestsDelta := c.metrics.API.TotalRequests - c.lastTotalRequests
		c.metrics.API.RequestsPerSecond = float64(requestsDelta) / elapsed

		c.lastTotalRequests = c.metrics.API.TotalRequests
		c.lastCheckTime = now
	}

	// 更新响应时间百分位
	c.metrics.API.ResponseTimeP50 = c.calculateResponseTimePercentile(0.5)
	c.metrics.API.ResponseTimeP90 = c.calculateResponseTimePercentile(0.9)
	c.metrics.API.ResponseTimeP99 = c.calculateResponseTimePercentile(0.99)

	// 计算平均响应时间
	total := c.metrics.API.TotalRequests
	if total > 0 {
		sum := 0.0
		for _, rt := range c.responseTimes {
			sum += rt
		}
		c.metrics.API.AvgResponseTime = sum / float64(len(c.responseTimes))
	}
}

// ==================== 认证指标记录 ====================

// RecordLoginAttempt 记录登录尝试
func (c *MetricsCollector) RecordLoginAttempt(success bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.metrics.Auth.LoginAttempts++
	if success {
		c.metrics.Auth.LoginSuccess++
	} else {
		c.metrics.Auth.LoginFailed++
	}

	// 计算成功率
	if c.metrics.Auth.LoginAttempts > 0 {
		c.metrics.Auth.LoginSuccessRate = float64(c.metrics.Auth.LoginSuccess) / float64(c.metrics.Auth.LoginAttempts) * 100
	}
}

// RecordAccountLocked 记录账户锁定
func (c *MetricsCollector) RecordAccountLocked() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.metrics.Auth.AccountLocked++
}

// RecordAccountUnlocked 记录账户解锁
func (c *MetricsCollector) RecordAccountUnlocked() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.metrics.Auth.AccountUnlockeds++
}

// RecordTokenIssued 记录 Token 签发
func (c *MetricsCollector) RecordTokenIssued() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.metrics.Auth.TokenIssued++
}

// RecordTokenRefreshed 记录 Token 刷新
func (c *MetricsCollector) RecordTokenRefreshed() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.metrics.Auth.TokenRefreshed++
}

// RecordTokenInvalid 记录无效 Token
func (c *MetricsCollector) RecordTokenInvalid() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.metrics.Auth.TokenInvalid++
}

// UpdateActiveSessions 更新活跃会话数
func (c *MetricsCollector) UpdateActiveSessions(count int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.metrics.Auth.ActiveSessions = count
	if count > c.metrics.Auth.MaxSessions {
		c.metrics.Auth.MaxSessions = count
	}
}

// ==================== K8s 资源指标记录 ====================

// UpdateK8sResourceCounts 更新 K8s 资源数量
func (c *MetricsCollector) UpdateK8sResourceCounts(counts K8sResourceMetrics) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.metrics.K8s.Namespaces = counts.Namespaces
	c.metrics.K8s.Nodes = counts.Nodes
	c.metrics.K8s.Pods = counts.Pods
	c.metrics.K8s.RunningPods = counts.RunningPods
	c.metrics.K8s.PendingPods = counts.PendingPods
	c.metrics.K8s.FailedPods = counts.FailedPods
	c.metrics.K8s.Deployments = counts.Deployments
	c.metrics.K8s.StatefulSets = counts.StatefulSets
	c.metrics.K8s.DaemonSets = counts.DaemonSets
	c.metrics.K8s.Jobs = counts.Jobs
	c.metrics.K8s.CronJobs = counts.CronJobs
	c.metrics.K8s.Services = counts.Services
	c.metrics.K8s.Ingresses = counts.Ingresses
	c.metrics.K8s.PVCs = counts.PVCs
	c.metrics.K8s.PVs = counts.PVs
	c.metrics.K8s.ConfigMaps = counts.ConfigMaps
	c.metrics.K8s.Secrets = counts.Secrets
}

// RecordK8sAPICall 记录 K8s API 调用
func (c *MetricsCollector) RecordK8sAPICall(durationMs float64, success bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.metrics.K8s.APICallsTotal++
	if success {
		c.metrics.K8s.APICallsSuccess++
	} else {
		c.metrics.K8s.APICallsFailed++
	}

	// 初始化延迟状态
	if c.metrics.K8s.apiLatencyState == nil {
		c.metrics.K8s.apiLatencyState = &apiLatencyState{
			samples: make([]float64, PercentileSampleSize),
		}
	}

	state := c.metrics.K8s.apiLatencyState

	// 使用滑动窗口记录延迟采样
	state.mu.Lock()
	defer state.mu.Unlock()

	// 环形缓冲区写入
	state.samples[state.index] = durationMs
	state.index = (state.index + 1) % PercentileSampleSize
	if state.index == 0 {
		state.full = true
	}

	// 计算真实 P99
	c.metrics.K8s.APILatencyP99 = c.calculateK8sAPILatencyPercentile(0.99)
}

// calculateK8sAPILatencyPercentile 计算 K8s API 延迟的百分位
func (c *MetricsCollector) calculateK8sAPILatencyPercentile(p float64) float64 {
	if c.metrics.K8s.apiLatencyState == nil {
		return 0
	}

	state := c.metrics.K8s.apiLatencyState
	state.mu.Lock()
	defer state.mu.Unlock()

	var samples []float64
	if state.full {
		samples = make([]float64, len(state.samples))
		copy(samples, state.samples)
	} else {
		samples = state.samples[:state.index]
	}

	return calculatePercentile(samples, p)
}

// ==================== 缓存指标记录 ====================

// UpdateCacheMetrics 更新缓存指标
func (c *MetricsCollector) UpdateCacheMetrics(hits, misses, size, maxSize, memoryUsage int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.metrics.Cache.Hits = hits
	c.metrics.Cache.Misses = misses
	c.metrics.Cache.Size = size
	c.metrics.Cache.MaxSize = maxSize
	c.metrics.Cache.MemoryUsage = memoryUsage

	// 计算命中率
	total := hits + misses
	if total > 0 {
		c.metrics.Cache.HitRate = float64(hits) / float64(total) * 100
	}

	// 计算使用率
	if maxSize > 0 {
		c.metrics.Cache.Utilization = float64(size) / float64(maxSize) * 100
	}
}

// RecordCacheEviction 记录缓存淘汰
func (c *MetricsCollector) RecordCacheEviction() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.metrics.Cache.Evictions++
}

// RecordCacheExpired 记录缓存过期
func (c *MetricsCollector) RecordCacheExpired(count int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.metrics.Cache.Expired += count
}

// ==================== 系统指标记录 ====================

// UpdateSystemMetrics 更新系统指标
func (c *MetricsCollector) UpdateSystemMetrics(memAlloc, memTotal, memSys, goroutines int64, cpuPercent float64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.metrics.System.MemoryAlloc = memAlloc
	c.metrics.System.MemoryTotal = memTotal
	c.metrics.System.MemorySys = memSys
	c.metrics.System.Goroutines = goroutines
	c.metrics.System.CPUUsagePercent = cpuPercent
	c.metrics.System.UptimeSeconds = int64(time.Since(c.metrics.System.StartTime).Seconds())
}

// UpdateGCMetrics 更新 GC 指标
func (c *MetricsCollector) UpdateGCMetrics(pausesMs float64, count uint32) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.metrics.System.GCPausesMs = pausesMs
	c.metrics.System.GCCount = count
}

// ==================== 自定义指标 ====================

// SetCustomMetric 设置自定义指标
func (c *MetricsCollector) SetCustomMetric(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.metrics.Custom[key] = value
}

// RemoveCustomMetric 移除自定义指标
func (c *MetricsCollector) RemoveCustomMetric(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.metrics.Custom, key)
}

// ==================== 获取指标 ====================

// GetMetrics 获取所有业务指标
func (c *MetricsCollector) GetMetrics() BusinessMetrics {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.metrics
}

// GetMetricsJSON 获取 JSON 格式的业务指标
func (c *MetricsCollector) GetMetricsJSON() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return map[string]interface{}{
		"api":    c.metrics.API,
		"auth":   c.metrics.Auth,
		"k8s":    c.metrics.K8s,
		"cache":  c.metrics.Cache,
		"system": c.metrics.System,
		"custom": c.metrics.Custom,
	}
}

// StartPeriodicUpdates 启动定期更新
func (c *MetricsCollector) StartPeriodicUpdates(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.UpdateAPIRates()
				c.logger.Debug("Business metrics updated")
			case <-c.ctx.Done():
				return
			}
		}
	}()
}

// Close 关闭指标收集器
func (c *MetricsCollector) Close() {
	c.cancel()
	c.logger.Info("Business metrics collector closed")
}

// ==================== 全局实例 ====================

var globalCollector *MetricsCollector

// InitBusinessMetrics 初始化全局业务指标收集器
func InitBusinessMetrics(logger *zap.Logger) *MetricsCollector {
	globalCollector = NewMetricsCollector(logger)
	globalCollector.StartPeriodicUpdates(MetricsReportInterval)
	return globalCollector
}

// GetBusinessMetricsCollector 获取全局业务指标收集器
func GetBusinessMetricsCollector() *MetricsCollector {
	return globalCollector
}
