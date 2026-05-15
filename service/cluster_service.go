package service

import (
	"context"
	"fmt"
	"time"

	"github.com/nick0323/K8sVision/config"
	"github.com/nick0323/K8sVision/model"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ClusterInfo struct {
	Name      string `json:"name"`
	APIServer string `json:"apiServer"`
	Version   string `json:"version"`
	Healthy   bool   `json:"healthy"`
	NodeCount int    `json:"nodeCount"`
	LastCheck int64  `json:"lastCheck"`
}

type ClusterService struct {
	clientMgr *ClientManager
	configMgr *config.Manager
	logger    *zap.Logger
}

func NewClusterService(clientMgr *ClientManager, configMgr *config.Manager, logger *zap.Logger) *ClusterService {
	return &ClusterService{
		clientMgr: clientMgr,
		configMgr: configMgr,
		logger:    logger,
	}
}

func (s *ClusterService) ListClusters(ctx context.Context) []ClusterInfo {
	healthList := s.clientMgr.GetClustersHealth(ctx)
	infos := make([]ClusterInfo, len(healthList))
	for i, h := range healthList {
		infos[i] = ClusterInfo{
			Name:      h.Name,
			APIServer: h.Host,
			Version:   h.Version,
			Healthy:   h.Healthy,
			NodeCount: h.NodeCount,
			LastCheck: h.LastCheck,
		}
	}
	return infos
}

func (s *ClusterService) AddCluster(ctx context.Context, name string, cfg *model.KubernetesConfig) error {
	if name == "" || name == "default" {
		return fmt.Errorf("cluster name cannot be empty or 'default'")
	}

	if err := s.clientMgr.AddCluster(name, cfg); err != nil {
		return fmt.Errorf("failed to add cluster: %w", err)
	}

	s.logger.Info("cluster added", zap.String("name", name), zap.String("apiServer", cfg.APIServer))
	return nil
}

func (s *ClusterService) RemoveCluster(ctx context.Context, name string) error {
	if name == "" || name == "default" {
		return fmt.Errorf("cannot remove default cluster")
	}

	if err := s.clientMgr.RemoveCluster(name); err != nil {
		return fmt.Errorf("failed to remove cluster: %w", err)
	}

	s.logger.Info("cluster removed", zap.String("name", name))
	return nil
}

func (s *ClusterService) TestConnection(ctx context.Context, cfg *model.KubernetesConfig) error {
	restConfig, err := buildK8sConfigFrom(cfg, false)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	restConfig.Timeout = 10 * time.Second

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	version, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	_, err = clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return fmt.Errorf("connected but failed to list nodes: %w", err)
	}

	s.logger.Info("cluster connection test succeeded",
		zap.String("apiServer", restConfig.Host),
		zap.String("version", version.GitVersion),
	)
	return nil
}

func (s *ClusterService) SaveToConfig(name string, clusterCfg *model.ClusterConfig) error {
	cfg := s.configMgr.GetConfig()

	existing := -1
	for i, c := range cfg.Clusters {
		if c.Name == name {
			existing = i
			break
		}
	}

	if existing >= 0 {
		cfg.Clusters[existing] = *clusterCfg
	} else {
		cfg.Clusters = append(cfg.Clusters, *clusterCfg)
	}

	s.logger.Info("cluster config saved", zap.String("name", name))
	return nil
}

func (s *ClusterService) RemoveFromConfig(name string) error {
	cfg := s.configMgr.GetConfig()

	idx := -1
	for i, c := range cfg.Clusters {
		if c.Name == name {
			idx = i
			break
		}
	}

	if idx < 0 {
		return fmt.Errorf("cluster %s not found in config", name)
	}

	cfg.Clusters = append(cfg.Clusters[:idx], cfg.Clusters[idx+1:]...)

	s.logger.Info("cluster config removed", zap.String("name", name))
	return nil
}
