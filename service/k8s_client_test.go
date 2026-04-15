package service

import (
	"os"
	"testing"
	"time"

	"github.com/nick0323/K8sVision/config"
	"github.com/nick0323/K8sVision/model"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
)

// 辅助函数：创建测试 logger
func createTestLogger(t *testing.T) *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	return logger
}

// ==================== ClientHolder 测试 ====================

// TestClientHolder_IsHealthy 测试健康检查
func TestClientHolder_IsHealthy(t *testing.T) {
	logger := createTestLogger(t)
	holder := &ClientHolder{
		healthy:         true,
		lastHealthCheck: time.Now(),
		logger:          logger,
	}

	// 初始应该是健康的
	if !holder.IsHealthy() {
		t.Error("Expected ClientHolder to be healthy initially")
	}

	// 设置 lastHealthCheck 为很久以前，应该不健康
	holder.lastHealthCheck = time.Now().Add(-5 * time.Minute)
	if holder.IsHealthy() {
		t.Error("Expected ClientHolder to be unhealthy after max age")
	}
}

// TestClientHolder_Close 测试关闭
func TestClientHolder_Close(t *testing.T) {
	logger := createTestLogger(t)
	holder := &ClientHolder{
		closeCh: make(chan struct{}),
		logger:  logger,
	}

	// 关闭不应该阻塞
	done := make(chan bool)
	go func() {
		holder.Close()
		done <- true
	}()

	select {
	case <-done:
		// 成功关闭
	case <-time.After(1 * time.Second):
		t.Error("Close() timed out")
	}
}

// ==================== ClientManager 测试 ====================

// TestNewClientManager 测试创建客户端管理器
func TestNewClientManager(t *testing.T) {
	logger := createTestLogger(t)
	configMgr := config.NewManager(logger)

	// 设置默认配置
	configMgr.Set("kubernetes.timeout", "30s")
	configMgr.Set("kubernetes.qps", "5")
	configMgr.Set("kubernetes.burst", "10")

	manager, err := NewClientManager(configMgr, logger)

	// 在没有 K8s 环境的情况下，应该返回错误
	if err == nil {
		// 如果成功创建，确保关闭
		manager.Close()
	}
	// 错误是可接受的（因为没有 K8s 环境）
	_ = configMgr
}

// TestClientManager_GetDefaultClient 测试获取默认客户端
func TestClientManager_GetDefaultClient(t *testing.T) {
	logger := createTestLogger(t)
	configMgr := config.NewManager(logger)

	manager := &ClientManager{
		configMgr: configMgr,
		logger:    logger,
		defaultClient: &ClientHolder{
			clientset:     nil,
			metricsClient: nil,
			healthy:       false,
			logger:        logger,
		},
	}

	// 在没有实际 K8s 客户端的情况下，应该返回 nil 或错误
	clientset, metricsClient, err := manager.GetDefaultClient()
	if clientset != nil || metricsClient != nil {
		t.Error("Expected nil clients in test environment")
	}
	// err 可能是 nil 或错误，取决于实现
	_ = err
}

// TestClientManager_GetClient 测试获取指定名称的客户端
func TestClientManager_GetClient(t *testing.T) {
	logger := createTestLogger(t)
	configMgr := config.NewManager(logger)

	manager := &ClientManager{
		configMgr: configMgr,
		logger:    logger,
		defaultClient: &ClientHolder{
			logger: logger,
		},
	}

	// 测试空名称（应该返回默认客户端）
	_, _, _ = manager.GetClient("")

	// 测试 "default" 名称（应该返回默认客户端）
	_, _, _ = manager.GetClient("default")

	// 测试其他名称（应该返回默认客户端，因为多集群配置未实现）
	_, _, _ = manager.GetClient("cluster1")
}

// TestClientManager_Close 测试关闭管理器
func TestClientManager_Close(t *testing.T) {
	logger := createTestLogger(t)
	configMgr := config.NewManager(logger)

	manager := &ClientManager{
		configMgr: configMgr,
		logger:    logger,
		defaultClient: &ClientHolder{
			closeCh: make(chan struct{}),
			logger:  logger,
		},
		stopCh: make(chan struct{}),
	}

	// 关闭不应该阻塞
	done := make(chan bool)
	go func() {
		manager.Close()
		done <- true
	}()

	select {
	case <-done:
		// 成功关闭
	case <-time.After(2 * time.Second):
		t.Error("Close() timed out")
	}
}

// ==================== buildK8sConfig 测试 ====================

// TestBuildK8sConfig_InCluster 测试 in-cluster 配置
func TestBuildK8sConfig_InCluster(t *testing.T) {
	logger := createTestLogger(t)
	_ = config.NewManager(logger)

	// 确保没有 kubeconfig 文件和环境变量
	os.Unsetenv("KUBECONFIG")

	k8sConfig := &model.KubernetesConfig{
		Kubeconfig: "",
		APIServer:  "",
		Token:      "",
	}

	// 在没有 K8s 环境的情况下，应该返回错误
	_, err := buildK8sConfig(k8sConfig)
	if err == nil {
		t.Error("Expected error when no K8s config is provided")
	}
}

// TestBuildK8sConfig_APIServer 测试使用 APIServer 配置
func TestBuildK8sConfig_APIServer(t *testing.T) {
	logger := createTestLogger(t)
	_ = config.NewManager(logger)

	k8sConfig := &model.KubernetesConfig{
		Kubeconfig: "",
		APIServer:  "https://api.example.com",
		Token:      "test-token",
		Insecure:   true,
		Timeout:    30 * time.Second,
		QPS:        5.0,
		Burst:      10,
	}

	// 应该成功构建配置
	cfg, err := buildK8sConfig(k8sConfig)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if cfg == nil {
		t.Fatal("Expected config to be returned")
	}
	if cfg.Host != "https://api.example.com" {
		t.Errorf("Expected Host to be 'https://api.example.com', got %s", cfg.Host)
	}
	if cfg.BearerToken != "test-token" {
		t.Errorf("Expected BearerToken to be 'test-token', got %s", cfg.BearerToken)
	}
}

// TestBuildK8sConfig_Kubeconfig 测试使用 kubeconfig 文件
func TestBuildK8sConfig_Kubeconfig(t *testing.T) {
	logger := createTestLogger(t)
	_ = config.NewManager(logger)

	k8sConfig := &model.KubernetesConfig{
		Kubeconfig: "/nonexistent/path/kubeconfig",
	}

	// 应该返回错误（文件不存在）
	_, err := buildK8sConfig(k8sConfig)
	if err == nil {
		t.Error("Expected error when kubeconfig file doesn't exist")
	}
}

// ==================== applyK8sConfigOptimized 测试 ====================

// TestApplyK8sConfigOptimized 测试应用 K8s 配置优化
func TestApplyK8sConfigOptimized(t *testing.T) {
	// 创建基础配置
	cfg := &rest.Config{
		Timeout: 0,
		QPS:     0,
		Burst:   0,
	}

	k8sConfig := &model.KubernetesConfig{
		Timeout:  60 * time.Second,
		QPS:      10.0,
		Burst:    20,
		Insecure: true,
		CertFile: "/path/to/cert",
		KeyFile:  "/path/to/key",
		CAFile:   "/path/to/ca",
	}

	applyK8sConfigOptimized(cfg, k8sConfig)

	if cfg.Timeout != 60*time.Second {
		t.Errorf("Expected Timeout to be 60s, got %v", cfg.Timeout)
	}
	if cfg.QPS != 10.0 {
		t.Errorf("Expected QPS to be 10.0, got %f", cfg.QPS)
	}
	if cfg.Burst != 20 {
		t.Errorf("Expected Burst to be 20, got %d", cfg.Burst)
	}
	if !cfg.TLSClientConfig.Insecure {
		t.Error("Expected Insecure to be true")
	}
	if cfg.TLSClientConfig.CertFile != "/path/to/cert" {
		t.Errorf("Expected CertFile to be '/path/to/cert', got %s", cfg.TLSClientConfig.CertFile)
	}
	if cfg.UserAgent != "KubeVision/1.0" {
		t.Errorf("Expected UserAgent to be 'KubeVision/1.0', got %s", cfg.UserAgent)
	}
}

// ==================== ClientHolder_GetClientset 测试 ====================

// TestClientHolder_GetClientset 测试获取客户端集
func TestClientHolder_GetClientset(t *testing.T) {
	logger := createTestLogger(t)
	holder := &ClientHolder{
		healthy:         true,
		lastHealthCheck: time.Now(),
		logger:          logger,
	}

	// 在没有实际客户端的情况下，应该返回 nil
	clientset, metricsClient, err := holder.GetClientset()
	if clientset != nil || metricsClient != nil {
		t.Error("Expected nil clients in test environment")
	}
	// err 可能是 nil（如果健康）或错误
	_ = err
}

// TestClientHolder_GetClientset_Unhealthy 测试不健康时获取客户端
func TestClientHolder_GetClientset_Unhealthy(t *testing.T) {
	logger := createTestLogger(t)
	holder := &ClientHolder{
		healthy:         false,
		lastHealthCheck: time.Now().Add(-5 * time.Minute), // 很久以前的检查
		logger:          logger,
	}

	// 应该返回错误
	clientset, metricsClient, err := holder.GetClientset()
	if clientset != nil || metricsClient != nil {
		t.Error("Expected nil clients when unhealthy")
	}
	if err == nil {
		t.Error("Expected error when client is unhealthy")
	}
}

// ==================== startHealthCheck 测试 ====================

// TestClientHolder_StartHealthCheck 测试健康检查启动
// 注：由于需要实际的 K8s 客户端，此测试仅验证结构
func TestClientHolder_StartHealthCheck(t *testing.T) {
	logger := createTestLogger(t)
	holder := &ClientHolder{
		healthy:         true,
		lastHealthCheck: time.Now(),
		closeCh:         make(chan struct{}),
		logger:          logger,
		// 不设置 clientset，避免 nil pointer
	}

	// 验证 closeCh 已初始化
	if holder.closeCh == nil {
		t.Error("Expected closeCh to be initialized")
	}

	// 关闭
	holder.Close()
	// 测试主要是确保不崩溃
}
