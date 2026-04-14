package service

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/nick0323/K8sVision/config"
	"github.com/nick0323/K8sVision/model"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

// K8s 客户端配置常量
const (
	HealthCheckInterval = 30 * time.Second
	HealthCheckTimeout  = 5 * time.Second
	HealthCheckMaxAge   = 2 * time.Minute
)

// ClientHolder K8s 客户端持有者
type ClientHolder struct {
	clientset       *kubernetes.Clientset
	metricsClient   *metrics.Clientset
	config          *rest.Config
	mu              sync.RWMutex
	healthCh        chan struct{}
	closeCh         chan struct{}
	logger          *zap.Logger
	lastHealthCheck time.Time
	healthy         bool
}

// ClientManager K8s 客户端管理器 - 支持多集群和连接池
type ClientManager struct {
	configMgr      *config.Manager
	logger         *zap.Logger
	defaultClient  *ClientHolder
	clientPool     sync.Map // map[string]*ClientHolder
	healthInterval time.Duration
	stopCh         chan struct{}
}

// NewClientHolder 创建新的客户端持有者
func NewClientHolder(cfg *rest.Config, logger *zap.Logger) (*ClientHolder, error) {
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	metricsClient, err := metrics.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %w", err)
	}

	holder := &ClientHolder{
		clientset:     clientset,
		metricsClient: metricsClient,
		config:        cfg,
		healthCh:      make(chan struct{}, 1),
		closeCh:       make(chan struct{}),
		logger:        logger,
		healthy:       true,
	}

	// 启动健康检查
	go holder.startHealthCheck()

	return holder, nil
}

// startHealthCheck 启动健康检查
func (h *ClientHolder) startHealthCheck() {
	// 立即执行首次健康检查
	h.checkHealth()

	ticker := time.NewTicker(HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.checkHealth()
		case <-h.closeCh:
			return
		}
	}
}

// checkHealth 检查客户端健康状态
func (h *ClientHolder) checkHealth() {
	ctx, cancel := context.WithTimeout(context.Background(), HealthCheckTimeout)
	defer cancel()

	_, err := h.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})

	h.mu.Lock()
	defer h.mu.Unlock()

	h.lastHealthCheck = time.Now()
	if err != nil {
		h.healthy = false
		h.logger.Warn("kubernetes client health check failed",
			zap.Error(err),
			zap.String("host", h.config.Host),
		)
	} else {
		h.healthy = true
	}
}

// IsHealthy 检查客户端是否健康
func (h *ClientHolder) IsHealthy() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// 首次启动时，如果还未进行检查，认为客户端是健康的（允许宽限期）
	if h.lastHealthCheck.IsZero() {
		return true
	}

	return h.healthy && time.Since(h.lastHealthCheck) < HealthCheckMaxAge
}

// GetClientset 获取客户端集
func (h *ClientHolder) GetClientset() (*kubernetes.Clientset, *metrics.Clientset, error) {
	// 首次调用时，如果不健康但还未进行检查，允许使用客户端
	if !h.IsHealthy() && !h.lastHealthCheck.IsZero() {
		return nil, nil, fmt.Errorf("kubernetes client is unhealthy")
	}
	return h.clientset, h.metricsClient, nil
}

// Close 关闭客户端
func (h *ClientHolder) Close() {
	close(h.closeCh)
}

// NewClientManager 创建客户端管理器
func NewClientManager(configMgr *config.Manager, logger *zap.Logger) (*ClientManager, error) {
	defaultHolder, err := createDefaultClientHolder(configMgr, logger)
	if err != nil {
		return nil, err
	}

	manager := &ClientManager{
		configMgr:      configMgr,
		logger:         logger,
		defaultClient:  defaultHolder,
		healthInterval: HealthCheckInterval,
		stopCh:         make(chan struct{}),
	}

	// 启动定期健康检查
	go manager.startHealthMonitor()

	return manager, nil
}

func createDefaultClientHolder(configMgr *config.Manager, logger *zap.Logger) (*ClientHolder, error) {
	cfg := configMgr.GetConfig()
	k8sConfig := &cfg.Kubernetes

	restConfig, err := buildK8sConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build k8s config: %w", err)
	}

	return NewClientHolder(restConfig, logger)
}

// buildK8sConfig 构建 K8s REST 配置
func buildK8sConfig(k8sConfig *model.KubernetesConfig) (*rest.Config, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = k8sConfig.Kubeconfig
	}

	var config *rest.Config
	var err error

	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
		}
	} else if k8sConfig.APIServer != "" && k8sConfig.Token != "" {
		config = &rest.Config{
			Host:        k8sConfig.APIServer,
			BearerToken: k8sConfig.Token,
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: k8sConfig.Insecure,
			},
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
		}
	}

	// 应用配置覆盖
	applyK8sConfigOptimized(config, k8sConfig)
	return config, nil
}

// applyK8sConfigOptimized 应用 K8s 配置优化
func applyK8sConfigOptimized(cfg *rest.Config, k8sConfig *model.KubernetesConfig) {
	if k8sConfig.Timeout > 0 {
		cfg.Timeout = k8sConfig.Timeout
	}
	if k8sConfig.QPS > 0 {
		cfg.QPS = k8sConfig.QPS
	}
	if k8sConfig.Burst > 0 {
		cfg.Burst = k8sConfig.Burst
	}

	// 优化 TLS 配置
	cfg.TLSClientConfig.Insecure = k8sConfig.Insecure
	if k8sConfig.CertFile != "" {
		cfg.TLSClientConfig.CertFile = k8sConfig.CertFile
	}
	if k8sConfig.KeyFile != "" {
		cfg.TLSClientConfig.KeyFile = k8sConfig.KeyFile
	}
	if k8sConfig.CAFile != "" {
		cfg.TLSClientConfig.CAFile = k8sConfig.CAFile
	}

	// 优化用户代理
	cfg.UserAgent = "KubeVision/1.0"
}

// GetDefaultClient 获取默认客户端
func (m *ClientManager) GetDefaultClient() (*kubernetes.Clientset, *metrics.Clientset, error) {
	return m.defaultClient.GetClientset()
}

// GetDefaultRESTConfig 获取默认客户端的 REST 配置（用于 WebSocket exec 等场景）
func (m *ClientManager) GetDefaultRESTConfig() *rest.Config {
	return m.defaultClient.config
}

// GetClient 获取指定名称的客户端（支持多集群）
func (m *ClientManager) GetClient(clusterName string) (*kubernetes.Clientset, *metrics.Clientset, error) {
	if clusterName == "" || clusterName == "default" {
		return m.GetDefaultClient()
	}

	if client, ok := m.clientPool.Load(clusterName); ok {
		holder := client.(*ClientHolder)
		return holder.GetClientset()
	}

	// TODO: 支持动态加载多集群配置
	return m.GetDefaultClient()
}

// startHealthMonitor 启动健康监控
func (m *ClientManager) startHealthMonitor() {
	ticker := time.NewTicker(m.healthInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkAllClients()
		case <-m.stopCh:
			return
		}
	}
}

// checkAllClients 检查所有客户端健康状态
func (m *ClientManager) checkAllClients() {
	m.defaultClient.checkHealth()

	m.clientPool.Range(func(key, value interface{}) bool {
		holder := value.(*ClientHolder)
		holder.checkHealth()
		return true
	})
}

// Close 关闭客户端管理器
func (m *ClientManager) Close() {
	close(m.stopCh)
	m.defaultClient.Close()

	m.clientPool.Range(func(key, value interface{}) bool {
		holder := value.(*ClientHolder)
		holder.Close()
		return true
	})
}
