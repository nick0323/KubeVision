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

const (
	HealthCheckInterval = 30 * time.Second
	HealthCheckTimeout  = 5 * time.Second
	HealthCheckMaxAge   = 2 * time.Minute
)

type ClientHolder struct {
	clientset       *kubernetes.Clientset
	metricsClient   *metrics.Clientset
	config          *rest.Config
	mu              sync.RWMutex
	closeCh         chan struct{}
	logger          *zap.Logger
	lastHealthCheck time.Time
	healthy         bool
}

type ClientManager struct {
	configMgr      *config.Manager
	logger         *zap.Logger
	defaultClient  *ClientHolder
	clientPool     sync.Map
	healthInterval time.Duration
	stopCh         chan struct{}
}

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
		closeCh:       make(chan struct{}),
		logger:        logger,
		healthy:       true,
	}

	go holder.startHealthCheck()
	return holder, nil
}

func (h *ClientHolder) startHealthCheck() {
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

func (h *ClientHolder) checkHealth() {
	ctx, cancel := context.WithTimeout(context.Background(), HealthCheckTimeout)
	defer cancel()

	_, err := h.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})

	h.mu.Lock()
	defer h.mu.Unlock()

	h.lastHealthCheck = time.Now()
	if err != nil {
		h.healthy = false
		h.logger.Warn("kubernetes client health check failed", zap.Error(err), zap.String("host", h.config.Host))
	} else {
		h.healthy = true
	}
}

func (h *ClientHolder) IsHealthy() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.lastHealthCheck.IsZero() {
		return true
	}
	return h.healthy && time.Since(h.lastHealthCheck) < HealthCheckMaxAge
}

func (h *ClientHolder) GetClientset() (*kubernetes.Clientset, *metrics.Clientset, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.clientset == nil {
		return nil, nil, fmt.Errorf("kubernetes client is not initialized")
	}

	if !h.IsHealthy() && !h.lastHealthCheck.IsZero() {
		return nil, nil, fmt.Errorf("kubernetes client is unhealthy")
	}
	return h.clientset, h.metricsClient, nil
}

func (h *ClientHolder) Close() {
	close(h.closeCh)
}

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

	go manager.startHealthMonitor()
	return manager, nil
}

func createDefaultClientHolder(configMgr *config.Manager, logger *zap.Logger) (*ClientHolder, error) {
	cfg := configMgr.GetConfig()
	restConfig, err := buildK8sConfig(&cfg.Kubernetes)
	if err != nil {
		return nil, fmt.Errorf("failed to build k8s config: %w", err)
	}
	return NewClientHolder(restConfig, logger)
}

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

	applyK8sConfig(config, k8sConfig)
	return config, nil
}

func applyK8sConfig(cfg *rest.Config, k8sConfig *model.KubernetesConfig) {
	if k8sConfig.Timeout > 0 {
		cfg.Timeout = k8sConfig.Timeout
	}
	if k8sConfig.QPS > 0 {
		cfg.QPS = k8sConfig.QPS
	}
	if k8sConfig.Burst > 0 {
		cfg.Burst = k8sConfig.Burst
	}

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

	cfg.UserAgent = "KubeVision/1.0"
}

func (m *ClientManager) GetDefaultClient() (*kubernetes.Clientset, *metrics.Clientset, error) {
	return m.defaultClient.GetClientset()
}

func (m *ClientManager) GetDefaultRESTConfig() *rest.Config {
	return m.defaultClient.config
}

func (m *ClientManager) GetClient(clusterName string) (*kubernetes.Clientset, *metrics.Clientset, error) {
	if clusterName == "" || clusterName == "default" {
		return m.GetDefaultClient()
	}

	if client, ok := m.clientPool.Load(clusterName); ok {
		holder := client.(*ClientHolder)
		return holder.GetClientset()
	}

	return m.GetDefaultClient()
}

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

func (m *ClientManager) checkAllClients() {
	m.defaultClient.checkHealth()
	m.clientPool.Range(func(key, value interface{}) bool {
		holder := value.(*ClientHolder)
		holder.checkHealth()
		return true
	})
}

func (m *ClientManager) Close() {
	close(m.stopCh)
	m.defaultClient.Close()
	m.clientPool.Range(func(key, value interface{}) bool {
		holder := value.(*ClientHolder)
		holder.Close()
		return true
	})
}
