package bootstrap

import (
	"context"
	"fmt"

	"github.com/nick0323/K8sVision/api"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/cache"
	"github.com/nick0323/K8sVision/config"
	"github.com/nick0323/K8sVision/internal/security"
	"github.com/nick0323/K8sVision/model"
	"github.com/nick0323/K8sVision/service"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Initializer struct {
	logger    *zap.Logger
	configMgr *config.Manager
}

func NewInitializer(logger *zap.Logger, configMgr *config.Manager) *Initializer {
	return &Initializer{logger: logger, configMgr: configMgr}
}

func (i *Initializer) InitBaseComponents(configFile string) (*zap.Logger, *cache.MemoryCache[interface{}], error) {
	cfg := i.configMgr.GetConfig()
	logger, err := InitLogger(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	i.configMgr.UpdateLogger(logger)

	if err := security.NewSecurityConfig(i.configMgr, logger).CheckAndValidate(); err != nil {
		return nil, nil, err
	}

	cfg = i.configMgr.GetConfig()
	if err := cfg.Validate(); err != nil {
		return nil, nil, fmt.Errorf("config validation failed: %w", err)
	}

	return logger, InitLRUCache(i.configMgr, logger), nil
}

func (i *Initializer) InitK8sComponents(logger *zap.Logger) (*service.ClientManager, error) {
	k8sClientMgr, err := InitK8sClient(i.configMgr, logger)
	if err != nil {
		return nil, err
	}

	InitK8sCache(k8sClientMgr, logger)
	return k8sClientMgr, nil
}

func InitLogger(cfg *model.Config) (*zap.Logger, error) {
	var zapConfig zap.Config

	if cfg.IsDevelopment() {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	logLevel := zap.InfoLevel
	switch cfg.Log.Level {
	case "debug":
		logLevel = zap.DebugLevel
	case "info":
		logLevel = zap.InfoLevel
	case "warn":
		logLevel = zap.WarnLevel
	case "error":
		logLevel = zap.ErrorLevel
	}
	zapConfig.Level = zap.NewAtomicLevelAt(logLevel)

	if cfg.Log.Format == "console" {
		zapConfig.Encoding = "console"
	} else {
		zapConfig.Encoding = "json"
	}

	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")

	return zapConfig.Build()
}

func InitLRUCache(configMgr *config.Manager, logger *zap.Logger) *cache.MemoryCache[interface{}] {
	return cache.NewMemoryCache(&configMgr.GetConfig().Cache, logger)
}

func InitK8sClient(configMgr *config.Manager, logger *zap.Logger) (*service.ClientManager, error) {
	k8sClientMgr, err := service.NewClientManager(configMgr, logger)
	if err != nil {
		logger.Warn("K8s client initialization failed, K8s features will be unavailable",
			zap.Error(err),
			zap.String("hint", "Can be enabled by configuring kubeconfig or using in-cluster mode"),
		)
		return nil, nil
	}
	return k8sClientMgr, nil
}

func InitK8sCache(k8sClientMgr *service.ClientManager, logger *zap.Logger) {
	if k8sClientMgr == nil {
		logger.Debug("K8s client manager not initialized, skipping cache initialization")
		return
	}

	clientset, _, err := k8sClientMgr.GetDefaultClient()
	if err != nil || clientset == nil {
		logger.Warn("K8s client unavailable, skipping cache initialization")
		return
	}

	podInformer := service.NewPodInformer(clientset, "")
	service.SetPodInformer(podInformer)
	go podInformer.Start(context.Background())
	logger.Info("Pod Informer started")
}

func InitServices(k8sClientMgr *service.ClientManager, logger *zap.Logger) {
	if k8sClientMgr == nil {
		return
	}
	clientset, _, _ := k8sClientMgr.GetDefaultClient()
	if clientset == nil {
		return
	}
}

func InitAPI(configMgr *config.Manager, k8sClientMgr *service.ClientManager, logger *zap.Logger) {
	cfg := configMgr.GetConfig()

	api.SetConfigManager(configMgr)
	api.InitAuthManager(logger)
	middleware.SetJWTSecret([]byte(cfg.JWT.Secret))
	api.SetGlobalClientManager(k8sClientMgr)
	api.InitWebSocketUpgrader(cfg.Server.AllowedOrigin)
}
