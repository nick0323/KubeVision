package monitor

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

// 辅助函数：创建测试 logger
func createTestLogger(t *testing.T) *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	return logger
}

// ==================== Metrics 测试 ====================

// TestNewMetrics 测试创建 Metrics
func TestNewMetrics(t *testing.T) {
	logger := createTestLogger(t)
	metrics := NewMetrics(logger)

	if metrics == nil {
		t.Fatal("Expected metrics to be created")
	}
	if metrics.StartTime.IsZero() {
		t.Error("Expected StartTime to be set")
	}
	if metrics.ResourceCounts == nil {
		t.Error("Expected ResourceCounts to be initialized")
	}
}

// TestMetrics_RecordRequest 测试记录请求
func TestMetrics_RecordRequest(t *testing.T) {
	logger := createTestLogger(t)
	metrics := NewMetrics(logger)

	// 记录成功请求
	metrics.RecordRequest(true, 100*time.Millisecond)

	if metrics.TotalRequests != 1 {
		t.Errorf("Expected TotalRequests to be 1, got %d", metrics.TotalRequests)
	}
	if metrics.SuccessfulRequests != 1 {
		t.Errorf("Expected SuccessfulRequests to be 1, got %d", metrics.SuccessfulRequests)
	}
	if metrics.MinResponseTime != 100*time.Millisecond {
		t.Errorf("Expected MinResponseTime to be 100ms, got %v", metrics.MinResponseTime)
	}
	if metrics.MaxResponseTime != 100*time.Millisecond {
		t.Errorf("Expected MaxResponseTime to be 100ms, got %v", metrics.MaxResponseTime)
	}

	// 记录失败请求
	metrics.RecordRequest(false, 200*time.Millisecond)

	if metrics.TotalRequests != 2 {
		t.Errorf("Expected TotalRequests to be 2, got %d", metrics.TotalRequests)
	}
	if metrics.FailedRequests != 1 {
		t.Errorf("Expected FailedRequests to be 1, got %d", metrics.FailedRequests)
	}
	if metrics.MaxResponseTime != 200*time.Millisecond {
		t.Errorf("Expected MaxResponseTime to be 200ms, got %v", metrics.MaxResponseTime)
	}
}

// TestMetrics_RecordCacheHit 测试记录缓存命中
func TestMetrics_RecordCacheHit(t *testing.T) {
	logger := createTestLogger(t)
	metrics := NewMetrics(logger)

	metrics.RecordCacheHit()
	if metrics.CacheHits != 1 {
		t.Errorf("Expected CacheHits to be 1, got %d", metrics.CacheHits)
	}

	metrics.RecordCacheHit()
	if metrics.CacheHits != 2 {
		t.Errorf("Expected CacheHits to be 2, got %d", metrics.CacheHits)
	}
}

// TestMetrics_RecordCacheMiss 测试记录缓存未命中
func TestMetrics_RecordCacheMiss(t *testing.T) {
	logger := createTestLogger(t)
	metrics := NewMetrics(logger)

	metrics.RecordCacheMiss()
	if metrics.CacheMisses != 1 {
		t.Errorf("Expected CacheMisses to be 1, got %d", metrics.CacheMisses)
	}
}

// TestMetrics_RecordConnection 测试记录连接
func TestMetrics_RecordConnection(t *testing.T) {
	logger := createTestLogger(t)
	metrics := NewMetrics(logger)

	metrics.RecordConnection()
	if metrics.CurrentConnections != 1 {
		t.Errorf("Expected CurrentConnections to be 1, got %d", metrics.CurrentConnections)
	}
	if metrics.MaxConnections != 1 {
		t.Errorf("Expected MaxConnections to be 1, got %d", metrics.MaxConnections)
	}

	metrics.RecordConnection()
	if metrics.CurrentConnections != 2 {
		t.Errorf("Expected CurrentConnections to be 2, got %d", metrics.CurrentConnections)
	}
	if metrics.MaxConnections != 2 {
		t.Errorf("Expected MaxConnections to be 2, got %d", metrics.MaxConnections)
	}
}

// TestMetrics_RecordDisconnection 测试记录断开连接
func TestMetrics_RecordDisconnection(t *testing.T) {
	logger := createTestLogger(t)
	metrics := NewMetrics(logger)

	metrics.RecordConnection()
	metrics.RecordConnection()
	metrics.RecordDisconnection()

	if metrics.CurrentConnections != 1 {
		t.Errorf("Expected CurrentConnections to be 1, got %d", metrics.CurrentConnections)
	}
}

// TestMetrics_RecordError 测试记录错误
func TestMetrics_RecordError(t *testing.T) {
	logger := createTestLogger(t)
	metrics := NewMetrics(logger)

	metrics.RecordError("test error")

	if metrics.ErrorCount != 1 {
		t.Errorf("Expected ErrorCount to be 1, got %d", metrics.ErrorCount)
	}
	if metrics.LastError != "test error" {
		t.Errorf("Expected LastError to be 'test error', got %s", metrics.LastError)
	}
	if metrics.LastErrorTime.IsZero() {
		t.Error("Expected LastErrorTime to be set")
	}
}

// TestMetrics_RecordK8sAPICall 测试记录 K8s API 调用
func TestMetrics_RecordK8sAPICall(t *testing.T) {
	logger := createTestLogger(t)
	metrics := NewMetrics(logger)

	// 成功调用
	metrics.RecordK8sAPICall(100*time.Millisecond, true)
	if metrics.K8sAPICalls != 1 {
		t.Errorf("Expected K8sAPICalls to be 1, got %d", metrics.K8sAPICalls)
	}
	if metrics.K8sAPIErrors != 0 {
		t.Errorf("Expected K8sAPIErrors to be 0, got %d", metrics.K8sAPIErrors)
	}

	// 失败调用
	metrics.RecordK8sAPICall(200*time.Millisecond, false)
	if metrics.K8sAPICalls != 2 {
		t.Errorf("Expected K8sAPICalls to be 2, got %d", metrics.K8sAPICalls)
	}
	if metrics.K8sAPIErrors != 1 {
		t.Errorf("Expected K8sAPIErrors to be 1, got %d", metrics.K8sAPIErrors)
	}
}

// TestMetrics_RecordResourceCount 测试记录资源数量
func TestMetrics_RecordResourceCount(t *testing.T) {
	logger := createTestLogger(t)
	metrics := NewMetrics(logger)

	metrics.RecordResourceCount("pods", 10)
	if count, ok := metrics.ResourceCounts["pods"]; !ok || count != 10 {
		t.Errorf("Expected pods count to be 10, got %v", count)
	}

	metrics.RecordResourceCount("deployments", 5)
	if count, ok := metrics.ResourceCounts["deployments"]; !ok || count != 5 {
		t.Errorf("Expected deployments count to be 5, got %v", count)
	}
}

// TestMetrics_RecordMemoryUsage 测试记录内存使用
func TestMetrics_RecordMemoryUsage(t *testing.T) {
	logger := createTestLogger(t)
	metrics := NewMetrics(logger)

	metrics.RecordMemoryUsage(1024, 2048)
	if metrics.MemoryUsage != 1024 {
		t.Errorf("Expected MemoryUsage to be 1024, got %d", metrics.MemoryUsage)
	}
	if metrics.MemoryAllocated != 2048 {
		t.Errorf("Expected MemoryAllocated to be 2048, got %d", metrics.MemoryAllocated)
	}
}

// TestMetrics_RecordGoroutineCount 测试记录 Goroutine 数量
func TestMetrics_RecordGoroutineCount(t *testing.T) {
	logger := createTestLogger(t)
	metrics := NewMetrics(logger)

	metrics.RecordGoroutineCount(50)
	if metrics.GoroutineCount != 50 {
		t.Errorf("Expected GoroutineCount to be 50, got %d", metrics.GoroutineCount)
	}
}

// TestMetrics_UpdateSystemMetrics 测试更新系统指标
func TestMetrics_UpdateSystemMetrics(t *testing.T) {
	logger := createTestLogger(t)
	metrics := NewMetrics(logger)

	metrics.UpdateSystemMetrics()

	if metrics.MemoryUsage <= 0 {
		t.Error("Expected MemoryUsage to be greater than 0")
	}
	if metrics.GoroutineCount <= 0 {
		t.Error("Expected GoroutineCount to be greater than 0")
	}
}

// TestMetrics_GetStats 测试获取统计信息
func TestMetrics_GetStats(t *testing.T) {
	logger := createTestLogger(t)
	metrics := NewMetrics(logger)

	// 记录一些数据
	metrics.RecordRequest(true, 100*time.Millisecond)
	metrics.RecordCacheHit()
	metrics.RecordConnection()

	stats := metrics.GetStats()

	if stats["totalRequests"] != int64(1) {
		t.Errorf("Expected totalRequests to be 1, got %v", stats["totalRequests"])
	}
	if stats["cacheHits"] != int64(1) {
		t.Errorf("Expected cacheHits to be 1, got %v", stats["cacheHits"])
	}
	if stats["currentConnections"] != int64(1) {
		t.Errorf("Expected currentConnections to be 1, got %v", stats["currentConnections"])
	}
	if stats["successRate"] != 100.0 {
		t.Errorf("Expected successRate to be 100.0, got %v", stats["successRate"])
	}
}

// TestMetrics_Reset 测试重置统计
func TestMetrics_Reset(t *testing.T) {
	logger := createTestLogger(t)
	metrics := NewMetrics(logger)

	// 记录一些数据
	metrics.RecordRequest(true, 100*time.Millisecond)
	metrics.RecordCacheHit()
	metrics.RecordError("test error")

	// 重置
	metrics.Reset()

	if metrics.TotalRequests != 0 {
		t.Errorf("Expected TotalRequests to be 0 after reset, got %d", metrics.TotalRequests)
	}
	if metrics.CacheHits != 0 {
		t.Errorf("Expected CacheHits to be 0 after reset, got %d", metrics.CacheHits)
	}
	if metrics.ErrorCount != 0 {
		t.Errorf("Expected ErrorCount to be 0 after reset, got %d", metrics.ErrorCount)
	}
	if metrics.LastError != "" {
		t.Errorf("Expected LastError to be empty after reset, got %s", metrics.LastError)
	}
}

// ==================== Monitor 测试 ====================

// TestNewMonitor 测试创建 Monitor
func TestNewMonitor(t *testing.T) {
	logger := createTestLogger(t)
	monitor := NewMonitor(logger)

	if monitor == nil {
		t.Fatal("Expected monitor to be created")
	}
	if monitor.metrics == nil {
		t.Error("Expected metrics to be initialized")
	}
	if monitor.logger == nil {
		t.Error("Expected logger to be set")
	}
}

// TestMonitor_GetMetrics 测试获取指标
func TestMonitor_GetMetrics(t *testing.T) {
	logger := createTestLogger(t)
	monitor := NewMonitor(logger)

	metrics := monitor.GetMetrics()
	if metrics == nil {
		t.Error("Expected metrics to be returned")
	}
}

// TestMonitor_GetAllMetrics 测试获取所有指标
func TestMonitor_GetAllMetrics(t *testing.T) {
	logger := createTestLogger(t)
	monitor := NewMonitor(logger)

	allMetrics := monitor.GetAllMetrics()
	if allMetrics == nil {
		t.Error("Expected allMetrics to be returned")
	}
	if _, ok := allMetrics["total_requests"]; !ok {
		t.Error("Expected total_requests to be in allMetrics")
	}
}

// TestMonitor_StartPeriodicLogging 测试定期日志记录
func TestMonitor_StartPeriodicLogging(t *testing.T) {
	logger := createTestLogger(t)
	monitor := NewMonitor(logger)

	// 启动定期日志（很短的间隔用于测试）
	monitor.StartPeriodicLogging(50 * time.Millisecond)

	// 等待一段时间让日志记录执行
	time.Sleep(100 * time.Millisecond)

	// 测试主要是确保不崩溃
}

// TestMonitor_Close 测试关闭监控器
func TestMonitor_Close(t *testing.T) {
	logger := createTestLogger(t)
	monitor := NewMonitor(logger)

	monitor.Close()
	// 测试主要是确保不崩溃
}

// TestGetMetricsManager 测试获取全局监控管理器
func TestGetMetricsManager(t *testing.T) {
	logger := createTestLogger(t)
	monitor := NewMonitor(logger)

	globalMonitor := GetMetricsManager()
	if globalMonitor != monitor {
		t.Error("Expected GetMetricsManager to return the created monitor")
	}
}

// TestMetrics_ConcurrentAccess 测试并发访问安全性
func TestMetrics_ConcurrentAccess(t *testing.T) {
	logger := createTestLogger(t)
	metrics := NewMetrics(logger)

	done := make(chan bool)

	// 启动多个 goroutine 并发访问
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			metrics.RecordRequest(true, 10*time.Millisecond)
			metrics.RecordCacheHit()
			metrics.RecordConnection()
			metrics.GetStats()
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
}
