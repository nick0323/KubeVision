package service

import (
	"fmt"
	"os"

	"github.com/nick0323/K8sVision/config"
	"github.com/nick0323/K8sVision/model"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

var configManager *config.Manager

func SetConfigManager(cm *config.Manager) {
	configManager = cm
}

func GetK8sConfig() (*rest.Config, error) {
	if configManager == nil {
		return nil, fmt.Errorf("配置管理器未初始化")
	}

	cfg := configManager.GetConfig()
	k8sConfig := &cfg.Kubernetes

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
		return nil, nil, fmt.Errorf("获取 K8s 配置失败：%w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("创建 K8s 客户端失败：%w", err)
	}

	metricsClient, err := metrics.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("创建 Metrics 客户端失败：%w", err)
	}

	return clientset, metricsClient, nil
}
