package service

import (
	"fmt"
	"os"
	"time"

	"github.com/nick0323/K8sVision/cache"
	"github.com/nick0323/K8sVision/config"
	"github.com/nick0323/K8sVision/model"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

var (
	configManager *config.Manager
	clientsCache  *cache.MemoryCache
)

func SetConfigManager(cm *config.Manager) {
	configManager = cm
}

func SetCacheManager(cm *cache.Manager) {
	clientsCache = cache.NewMemoryCache(&model.CacheConfig{
		Enabled:         true,
		Type:            "memory",
		TTL:             30 * time.Minute,
		MaxSize:         10,
		CleanupInterval: 5 * time.Minute,
	}, cm.GetLogger())
}

func GetK8sConfig() (*rest.Config, error) {
	var k8sConfig *model.KubernetesConfig

	if configManager == nil {
		return nil, fmt.Errorf("配置管理器未初始化")
	}

	cfg := configManager.GetConfig()
	k8sConfig = &cfg.Kubernetes

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = k8sConfig.Kubeconfig
	}

	if kubeconfig != "" {
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}

		applyK8sConfig(config, k8sConfig)
		return config, nil
	}

	if k8sConfig.APIServer != "" && k8sConfig.Token != "" {
		config := &rest.Config{
			Host:        k8sConfig.APIServer,
			BearerToken: k8sConfig.Token,
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: k8sConfig.Insecure,
			},
		}

		applyK8sConfig(config, k8sConfig)
		return config, nil
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	applyK8sConfig(config, k8sConfig)
	return config, nil
}

func applyK8sConfig(config *rest.Config, k8sConfig *model.KubernetesConfig) {
	if k8sConfig.Timeout > 0 {
		config.Timeout = k8sConfig.Timeout
	}

	if k8sConfig.QPS > 0 {
		config.QPS = k8sConfig.QPS
	}
	if k8sConfig.Burst > 0 {
		config.Burst = k8sConfig.Burst
	}

	if k8sConfig.Insecure {
		config.Insecure = true
		config.TLSClientConfig.CAFile = ""
		config.TLSClientConfig.CAData = nil
	}

	if k8sConfig.CertFile != "" {
		config.TLSClientConfig.CertFile = k8sConfig.CertFile
	}
	if k8sConfig.KeyFile != "" {
		config.TLSClientConfig.KeyFile = k8sConfig.KeyFile
	}
	if k8sConfig.CAFile != "" {
		config.TLSClientConfig.CAFile = k8sConfig.CAFile
	}
}

func GetK8sClient() (*kubernetes.Clientset, *metrics.Clientset, error) {
	config, err := GetK8sConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("获取K8s配置失败: %w", err)
	}

	cacheKey := generateK8sClientCacheKey(config)

	if clientsCache != nil {
		if cached, exists := clientsCache.Get(cacheKey); exists {
			if clients, ok := cached.([]interface{}); ok && len(clients) == 2 {
				if k8sClient, ok := clients[0].(*kubernetes.Clientset); ok {
					if metricsClient, ok := clients[1].(*metrics.Clientset); ok {
						return k8sClient, metricsClient, nil
					}
				}
			}
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("创建K8s客户端失败: %w", err)
	}

	metricsClient, err := metrics.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("创建Metrics客户端失败: %w", err)
	}

	if clientsCache != nil {
		clientsCache.Set(cacheKey, []interface{}{clientset, metricsClient})
	}

	return clientset, metricsClient, nil
}

func generateK8sClientCacheKey(config *rest.Config) string {
	tokenSuffix := ""
	if len(config.BearerToken) > 0 {
		n := len(config.BearerToken)
		if n > 10 {
			n = 10
		}
		tokenSuffix = config.BearerToken[:n]
	}

	key := fmt.Sprintf("%s%s_%s_%v_%v",
		model.CacheKeyPrefixK8sClient,
		config.Host,
		tokenSuffix,
		config.Insecure,
		config.QPS,
	)
	return key
}
