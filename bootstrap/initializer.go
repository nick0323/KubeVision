package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/nick0323/K8sVision/api"
	"github.com/nick0323/K8sVision/cache"
	"github.com/nick0323/K8sVision/config"
	"github.com/nick0323/K8sVision/model"
	"github.com/nick0323/K8sVision/service"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ErrK8sUnavailable 表示 K8s 客户端未初始化（非致命错误）
var ErrK8sUnavailable = errors.New("kubernetes client unavailable")

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

	if err := config.NewSecurityChecker(i.configMgr, logger).CheckAndValidate(); err != nil {
		return nil, nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, nil, fmt.Errorf("config validation failed: %w", err)
	}

	return logger, InitLRUCache(i.configMgr, logger), nil
}

func (i *Initializer) InitK8sComponents(ctx context.Context, logger *zap.Logger) (*service.ClientManager, error) {
	k8sClientMgr, err := InitK8sClient(i.configMgr, logger)
	if err != nil {
		return nil, err
	}

	InitK8sCache(ctx, k8sClientMgr, logger)
	return k8sClientMgr, nil
}

func InitLogger(cfg *model.Config) (*zap.Logger, error) {
	// 1. 设置日志级别
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

	// 2. 配置编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if cfg.Log.Format == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	if cfg.Log.Format != "console" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 3. 配置输出目标 (分级输出)
	// 开发环境：全部输出到控制台
	// 生产环境：Info+ 到 stdout，Error+ 到 stderr
	var writeSyncers []zapcore.WriteSyncer
	writeSyncers = append(writeSyncers, zapcore.AddSync(os.Stdout))

	if !cfg.IsDevelopment() {
		writeSyncers = append(writeSyncers, zapcore.AddSync(os.Stderr))
	}

	// 4. 构建 Core
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), logLevel),
	)

	if !cfg.IsDevelopment() {
		// 生产环境额外将 Error 级别输出到 stderr
		core = zapcore.NewTee(
			zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), logLevel),
			zapcore.NewCore(encoder, zapcore.AddSync(os.Stderr), zapcore.ErrorLevel),
		)
	}

	// 5. 构建 Logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	return logger, nil
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
		return nil, ErrK8sUnavailable
	}
	return k8sClientMgr, nil
}

func InitK8sCache(ctx context.Context, k8sClientMgr *service.ClientManager, logger *zap.Logger) {
	if k8sClientMgr == nil {
		logger.Debug("K8s client manager not initialized, skipping cache initialization")
		return
	}

	clientset, err := k8sClientMgr.GetDefaultClient()
	if err != nil || clientset == nil {
		logger.Warn("K8s client unavailable, skipping cache initialization")
		return
	}

	podInformer := service.NewPodInformer(clientset, "")
	service.SetPodInformer(podInformer)
	go podInformer.Start(ctx)
	logger.Info("Pod Informer started")
}

// InitServices 预留服务初始化扩展点
// TODO: 未来可在此处初始化非 K8s 依赖的后台服务（如定时任务、消息队列消费者等）
func InitServices(k8sClientMgr *service.ClientManager, logger *zap.Logger) {
	// 目前为空，保留用于后续扩展
}

func InitAPI(configMgr *config.Manager, k8sClientMgr *service.ClientManager, logger *zap.Logger) {
	cfg := configMgr.GetConfig()

	api.InitExecClientManager(k8sClientMgr, configMgr)
	api.InitWebSocketUpgrader(cfg.Server.AllowedOrigin)
}
