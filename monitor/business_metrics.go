package monitor

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// 指标类型常量
const (
	MetricTypeCounter   = "counter"
	MetricTypeGauge     = "gauge"
	MetricTypeHistogram = "histogram"
	MetricTypeSummary   = "summary"
)

type BusinessMetric struct {
	Name       string
	Value      float64
	Timestamp  time.Time
	Labels     map[string]string
	MetricType string // counter, gauge, histogram, summary
}

type BusinessMetricsCollector struct {
	logger     *zap.Logger
	metrics    map[string]*BusinessMetric
	mutex      sync.RWMutex
	collectors []MetricCollector
}

type MetricCollector interface {
	Collect() []BusinessMetric
	GetName() string
}

// NewBusinessMetricsCollector 创建业务指标收集器
func NewBusinessMetricsCollector(logger *zap.Logger) *BusinessMetricsCollector {
	return &BusinessMetricsCollector{
		logger:     logger,
		metrics:    make(map[string]*BusinessMetric),
		collectors: make([]MetricCollector, 0),
	}
}

// RegisterCollector 注册指标收集器
func (bmc *BusinessMetricsCollector) RegisterCollector(collector MetricCollector) {
	bmc.mutex.Lock()
	defer bmc.mutex.Unlock()

	bmc.collectors = append(bmc.collectors, collector)
	bmc.logger.Info("Registered metric collector", zap.String("name", collector.GetName()))
}

// RecordMetric 记录业务指标
func (bmc *BusinessMetricsCollector) RecordMetric(metric BusinessMetric) {
	bmc.mutex.Lock()
	defer bmc.mutex.Unlock()

	key := bmc.generateKey(metric)
	bmc.metrics[key] = &metric

}

// GetMetric 获取指标
func (bmc *BusinessMetricsCollector) GetMetric(name string, labels map[string]string) *BusinessMetric {
	bmc.mutex.RLock()
	defer bmc.mutex.RUnlock()

	key := bmc.generateKey(BusinessMetric{Name: name, Labels: labels})
	return bmc.metrics[key]
}

// GetAllMetrics 获取所有指标
func (bmc *BusinessMetricsCollector) GetAllMetrics() map[string]*BusinessMetric {
	bmc.mutex.RLock()
	defer bmc.mutex.RUnlock()

	result := make(map[string]*BusinessMetric)
	for k, v := range bmc.metrics {
		result[k] = v
	}
	return result
}

// CollectMetrics 收集所有指标
func (bmc *BusinessMetricsCollector) CollectMetrics() []BusinessMetric {
	bmc.mutex.RLock()
	defer bmc.mutex.RUnlock()

	var allMetrics []BusinessMetric

	// 收集已记录的指标
	for _, metric := range bmc.metrics {
		allMetrics = append(allMetrics, *metric)
	}

	// 从收集器收集指标
	for _, collector := range bmc.collectors {
		metrics := collector.Collect()
		allMetrics = append(allMetrics, metrics...)
	}

	return allMetrics
}

// generateKey 生成指标键
func (bmc *BusinessMetricsCollector) generateKey(metric BusinessMetric) string {
	key := metric.Name
	for k, v := range metric.Labels {
		key += ":" + k + "=" + v
	}
	return key
}

// 具体的业务指标收集器

// K8sResourceCollector K8s资源指标收集器
type K8sResourceCollector struct {
	name string
}

func NewK8sResourceCollector() *K8sResourceCollector {
	return &K8sResourceCollector{name: "k8s_resources"}
}

func (krc *K8sResourceCollector) GetName() string {
	return krc.name
}

func (krc *K8sResourceCollector) Collect() []BusinessMetric {
	// 这里应该从K8s API收集实际的资源指标
	// 目前返回模拟数据
	return []BusinessMetric{
		{
			Name:       "k8s_pods_total",
			Value:      100,
			Timestamp:  time.Now(),
			Labels:     map[string]string{"namespace": "default"},
			MetricType: "gauge",
		},
		{
			Name:       "k8s_deployments_total",
			Value:      20,
			Timestamp:  time.Now(),
			Labels:     map[string]string{"namespace": "default"},
			MetricType: "gauge",
		},
	}
}

// APIMetricsCollector API指标收集器
type APIMetricsCollector struct {
	name string
}

func NewAPIMetricsCollector() *APIMetricsCollector {
	return &APIMetricsCollector{name: "api_metrics"}
}

func (amc *APIMetricsCollector) GetName() string {
	return amc.name
}

func (amc *APIMetricsCollector) Collect() []BusinessMetric {
	// 这里应该从API中间件收集实际的指标
	// 目前返回模拟数据
	return []BusinessMetric{
		{
			Name:       "api_requests_total",
			Value:      1000,
			Timestamp:  time.Now(),
			Labels:     map[string]string{"method": "GET", "endpoint": "/api/pods"},
			MetricType: "counter",
		},
		{
			Name:       "api_request_duration_seconds",
			Value:      0.1,
			Timestamp:  time.Now(),
			Labels:     map[string]string{"method": "GET", "endpoint": "/api/pods"},
			MetricType: "histogram",
		},
	}
}

// SystemMetricsCollector 系统指标收集器
type SystemMetricsCollector struct {
	name string
}

func NewSystemMetricsCollector() *SystemMetricsCollector {
	return &SystemMetricsCollector{name: "system_metrics"}
}

func (smc *SystemMetricsCollector) GetName() string {
	return smc.name
}

func (smc *SystemMetricsCollector) Collect() []BusinessMetric {
	// 这里应该收集实际的系统指标
	// 目前返回模拟数据
	return []BusinessMetric{
		{
			Name:       "memory_usage_bytes",
			Value:      1024 * 1024 * 100, // 100MB
			Timestamp:  time.Now(),
			Labels:     map[string]string{"type": "heap"},
			MetricType: "gauge",
		},
		{
			Name:       "cpu_usage_percent",
			Value:      25.5,
			Timestamp:  time.Now(),
			Labels:     map[string]string{"type": "user"},
			MetricType: "gauge",
		},
	}
}

// 全局业务指标收集器实例
var globalBusinessCollector *BusinessMetricsCollector

// InitBusinessMetrics 初始化业务指标收集器
func InitBusinessMetrics(logger *zap.Logger) {
	globalBusinessCollector = NewBusinessMetricsCollector(logger)

	// 注册默认收集器
	globalBusinessCollector.RegisterCollector(NewK8sResourceCollector())
	globalBusinessCollector.RegisterCollector(NewAPIMetricsCollector())
	globalBusinessCollector.RegisterCollector(NewSystemMetricsCollector())
}

// GetBusinessMetricsCollector 获取全局业务指标收集器
func GetBusinessMetricsCollector() *BusinessMetricsCollector {
	return globalBusinessCollector
}
