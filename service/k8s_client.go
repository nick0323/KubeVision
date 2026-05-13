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
)

const (
	HealthCheckInterval = 30 * time.Second
	HealthCheckTimeout  = 5 * time.Second
	HealthCheckMaxAge   = 2 * time.Minute
)

type ClientHolder struct {
	clientset       *kubernetes.Clientset
	config          *rest.Config
	mu              sync.RWMutex
	closeCh         chan struct{}
	logger          *zap.Logger
	lastHealthCheck time.Time
	healthy         bool
}

type ClientManager struct {
	configMgr        *config.Manager
	logger           *zap.Logger
	defaultClient    *ClientHolder
	clientPool       sync.Map
	healthInterval   time.Duration
	stopCh           chan struct{}
	argocdManager    *ArgoCDManager
	crdManager       *CRDManager
	crdManagerPool   sync.Map
	argocdManagerPool sync.Map
}

func NewClientHolder(cfg *rest.Config, logger *zap.Logger) (*ClientHolder, error) {
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	holder := &ClientHolder{
		clientset: clientset,
		config:    cfg,
		closeCh:   make(chan struct{}),
		logger:    logger,
		healthy:   true,
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

func (h *ClientHolder) GetClientset() (*kubernetes.Clientset, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.clientset == nil {
		return nil, fmt.Errorf("kubernetes client is not initialized")
	}

	if !h.IsHealthy() && !h.lastHealthCheck.IsZero() {
		return nil, fmt.Errorf("kubernetes client is unhealthy")
	}
	return h.clientset, nil
}

func (h *ClientHolder) Close() {
	close(h.closeCh)
}

func NewClientManager(configMgr *config.Manager, logger *zap.Logger) (*ClientManager, error) {
	defaultHolder, err := createDefaultClientHolder(configMgr, logger)
	if err != nil {
		return nil, err
	}

	// 初始化 ArgoCD Manager
	var argocdMgr *ArgoCDManager
	if defaultHolder.config != nil {
		argocdMgr, err = NewArgoCDManager(defaultHolder.config, logger)
		if err != nil {
			logger.Warn("Failed to initialize ArgoCD manager", zap.Error(err))
		}
	}

	// 初始化 CRD Manager
	var crdMgr *CRDManager
	if defaultHolder.config != nil {
		crdMgr, err = NewCRDManager(defaultHolder.config, logger)
		if err != nil {
			logger.Warn("Failed to initialize CRD manager", zap.Error(err))
		}
	}

	manager := &ClientManager{
		configMgr:      configMgr,
		logger:         logger,
		defaultClient:  defaultHolder,
		healthInterval: HealthCheckInterval,
		stopCh:         make(chan struct{}),
		argocdManager:  argocdMgr,
		crdManager:     crdMgr,
	}

	// 加载多集群配置
	if cfg := configMgr.GetConfig(); cfg != nil {
		manager.loadClustersFromConfig(cfg)
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
	return buildK8sConfigFrom(k8sConfig, true)
}

func buildK8sConfigFrom(k8sConfig *model.KubernetesConfig, useEnv bool) (*rest.Config, error) {
	kubeconfig := ""
	if useEnv {
		kubeconfig = os.Getenv("KUBECONFIG")
	}
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

func (m *ClientManager) GetDefaultClient() (*kubernetes.Clientset, error) {
	return m.defaultClient.GetClientset()
}

func (m *ClientManager) GetDefaultRESTConfig() *rest.Config {
	return m.defaultClient.config
}

func (m *ClientManager) GetArgoCDManager() *ArgoCDManager {
	return m.argocdManager
}

func (m *ClientManager) GetCRDManager() *CRDManager {
	return m.crdManager
}

func (m *ClientManager) GetCRDManagerForCluster(cluster string) (*CRDManager, error) {
	if cluster == "" || cluster == "default" {
		return m.crdManager, nil
	}
	if cached, ok := m.crdManagerPool.Load(cluster); ok {
		return cached.(*CRDManager), nil
	}
	restConfig := m.GetClientRESTConfig(cluster)
	if restConfig == nil {
		return nil, fmt.Errorf("no config for cluster: %s", cluster)
	}
	mgr, err := NewCRDManager(restConfig, m.logger.With(zap.String("cluster", cluster)))
	if err != nil {
		return nil, err
	}
	m.crdManagerPool.Store(cluster, mgr)
	return mgr, nil
}

func (m *ClientManager) GetArgoCDManagerForCluster(cluster string) (*ArgoCDManager, error) {
	if cluster == "" || cluster == "default" {
		return m.argocdManager, nil
	}
	if cached, ok := m.argocdManagerPool.Load(cluster); ok {
		return cached.(*ArgoCDManager), nil
	}
	restConfig := m.GetClientRESTConfig(cluster)
	if restConfig == nil {
		return nil, fmt.Errorf("no config for cluster: %s", cluster)
	}
	mgr, err := NewArgoCDManager(restConfig, m.logger.With(zap.String("cluster", cluster)))
	if err != nil {
		return nil, err
	}
	m.argocdManagerPool.Store(cluster, mgr)
	return mgr, nil
}

func (m *ClientManager) GetClient(clusterName string) (*kubernetes.Clientset, error) {
	if clusterName == "" || clusterName == "default" {
		return m.GetDefaultClient()
	}

	if client, ok := m.clientPool.Load(clusterName); ok {
		holder := client.(*ClientHolder)
		return holder.GetClientset()
	}

	m.logger.Warn("cluster not found, falling back to default", zap.String("cluster", clusterName))
	return m.GetDefaultClient()
}

func (m *ClientManager) GetClientRESTConfig(clusterName string) *rest.Config {
	if clusterName == "" || clusterName == "default" {
		return m.GetDefaultRESTConfig()
	}
	if client, ok := m.clientPool.Load(clusterName); ok {
		holder := client.(*ClientHolder)
		return holder.config
	}
	return m.GetDefaultRESTConfig()
}

func (m *ClientManager) AddCluster(name string, k8sConfig *model.KubernetesConfig) error {
	restConfig, err := buildK8sConfigFrom(k8sConfig, false)
	if err != nil {
		return fmt.Errorf("failed to build config for cluster %s: %w", name, err)
	}

	if old, ok := m.clientPool.Load(name); ok {
		old.(*ClientHolder).Close()
	}

	holder, err := NewClientHolder(restConfig, m.logger.With(zap.String("cluster", name)))
	if err != nil {
		return fmt.Errorf("failed to create client for cluster %s: %w", name, err)
	}

	m.clientPool.Store(name, holder)
	m.logger.Info("added cluster", zap.String("name", name), zap.String("host", restConfig.Host))
	return nil
}

func (m *ClientManager) GetClusterNames() []string {
	var names []string
	names = append(names, "default")
	m.clientPool.Range(func(key, _ interface{}) bool {
		names = append(names, key.(string))
		return true
	})
	return names
}

func (m *ClientManager) loadClustersFromConfig(cfg *model.Config) {
	for _, cluster := range cfg.Clusters {
		k8sConfig := &model.KubernetesConfig{
			Kubeconfig: cluster.Kubeconfig,
			APIServer:  cluster.APIServer,
			Token:      cluster.Token,
			Insecure:   cluster.Insecure,
			CAFile:     cluster.CAFile,
		}
		if err := m.AddCluster(cluster.Name, k8sConfig); err != nil {
			m.logger.Warn("failed to load cluster", zap.String("name", cluster.Name), zap.Error(err))
		}
	}
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
	m.argocdManager = nil
	m.crdManager = nil
	m.crdManagerPool.Range(func(key, value interface{}) bool {
		m.crdManagerPool.Delete(key)
		return true
	})
	m.argocdManagerPool.Range(func(key, value interface{}) bool {
		m.argocdManagerPool.Delete(key)
		return true
	})
}
