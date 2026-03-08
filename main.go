// Package main KubeVision - Kubernetes 集群可视化管理平台
package main

import (
	"context"
	"flag"
	"time"

	"github.com/nick0323/K8sVision/api"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/cache"
	"github.com/nick0323/K8sVision/config"
	"github.com/nick0323/K8sVision/monitor"
	"github.com/nick0323/K8sVision/model"
	"github.com/nick0323/K8sVision/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

// 路由常量
const (
	HealthCheckPath = "/health"
	CacheStatsPath  = "/cache/stats"
	APIPrefix       = "/api"
	LoginPath       = "/api/login"

	// K8s 配置
	DefaultConfigMapNamespace = "k8svision-system"
)

// Application 应用结构
type Application struct {
	configFile      string
	configMapNS     string
	logger          *zap.Logger
	configMgr       *config.Manager
	cacheMgr        *cache.CacheManager  // K8s Informer 缓存管理器
	lruCacheMgr     *cache.Manager       // LRU 缓存管理器
	monitorMgr      *monitor.Monitor
	k8sClientMgr    *service.ClientManager
	configStore     *service.ConfigMapStore
}

var configFile = flag.String("config", "", "配置文件路径")
var configMapNS = flag.String("configmap-ns", DefaultConfigMapNamespace, "ConfigMap 命名空间")

// initLogger 初始化日志
func initLogger(cfg *model.Config) (*zap.Logger, error) {
	var zapConfig zap.Config

	if cfg.IsDevelopment() {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	// 设置日志级别
	switch cfg.Log.Level {
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	// 设置日志格式
	if cfg.Log.Format == "console" {
		zapConfig.Encoding = "console"
	} else {
		zapConfig.Encoding = "json"
	}

	// 配置时间戳格式为可读格式
	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")

	return zapConfig.Build()
}

// NewApplication 创建应用实例
func NewApplication(configFile string, configMapNS string) *Application {
	return &Application{
		configFile:  configFile,
		configMapNS: configMapNS,
	}
}

// Initialize 初始化应用
func (app *Application) Initialize() error {
	tempLogger, _ := zap.NewProduction()
	app.configMgr = config.NewManager(tempLogger)

	if err := app.configMgr.Load(app.configFile); err != nil {
		tempLogger.Fatal("加载配置失败", zap.Error(err))
	}

	cfg := app.configMgr.GetConfig()

	var err error
	app.logger, err = initLogger(cfg)
	if err != nil {
		tempLogger.Fatal("初始化日志失败", zap.Error(err))
	}

	app.configMgr.UpdateLogger(app.logger)

	// 初始化 LRU 缓存管理器
	app.lruCacheMgr = cache.NewManager(&cfg.Cache, app.logger)

	// 初始化监控器
	app.monitorMgr = monitor.NewMonitor(app.logger)

	// 【优化】初始化 K8s 客户端管理器
	app.k8sClientMgr, err = service.NewClientManager(app.configMgr, app.logger)
	if err != nil {
		app.logger.Fatal("初始化 K8s 客户端失败", zap.Error(err))
	}

	// 【优化】初始化 ConfigMap 配置存储
	clientset, _, err := app.k8sClientMgr.GetDefaultClient()
	if err != nil {
		app.logger.Warn("获取 K8s 客户端失败，ConfigMap 存储将不可用", zap.Error(err))
	} else {
		app.configStore = service.NewConfigMapStore(clientset, app.logger, app.configMapNS)
		if err := app.configStore.Start(); err != nil {
			app.logger.Warn("启动 ConfigMap 存储失败", zap.Error(err))
		}
	}

	// 【优化】初始化 K8s 缓存管理器（Informer 机制）
	if clientset != nil {
		cacheMgr := cache.NewCacheManager(app.logger)
		
		// 注册 Pods 缓存
		podCache, err := cache.NewInformerCache(context.Background(), clientset, cache.ResourcePods, "", app.logger)
		if err != nil {
			app.logger.Warn("创建 Pods 缓存失败", zap.Error(err))
		} else {
			cacheMgr.RegisterCache(cache.ResourcePods, podCache)
		}
		
		// 注册 Nodes 缓存
		nodeCache, err := cache.NewInformerCache(context.Background(), clientset, cache.ResourceNodes, "", app.logger)
		if err != nil {
			app.logger.Warn("创建 Nodes 缓存失败", zap.Error(err))
		} else {
			cacheMgr.RegisterCache(cache.ResourceNodes, nodeCache)
		}
		
		// 启动所有缓存
		cacheMgr.StartAll()
		
		// 等待缓存同步（最多 2 分钟）
		if !cacheMgr.WaitForCacheSync(2 * time.Minute) {
			app.logger.Warn("缓存同步超时")
		} else {
			app.logger.Info("K8s 缓存已同步", zap.Any("stats", cacheMgr.GetCacheStats()))
		}
		
		// 保存到 Application
		app.cacheMgr = cacheMgr
		
		// 设置全局缓存管理器
		service.SetCacheManager(cacheMgr)
		
		// 【优化】初始化缓存排序服务
		sortService := service.NewCachedSortService(cacheMgr, app.logger)
		service.SetCachedSortService(sortService)
		app.logger.Info("缓存排序服务已启用")
	}

	// 设置全局配置管理器
	service.SetConfigManager(app.configMgr)
	api.SetConfigManager(app.configMgr)

	// 【优化】使用 ConfigMap 存储（如果可用）
	if app.configStore != nil {
		api.SetConfigStore(app.configStore)
		app.logger.Info("ConfigMap 配置存储已启用", zap.String("namespace", app.configMapNS))
	}

	// 初始化认证管理器
	api.InitAuthManager(app.logger)

	// 初始化监控和追踪
	monitor.InitTracing(app.logger)
	monitor.InitBusinessMetrics(app.logger)

	// 【优化】设置 K8s 客户端到监控器
	if clientset != nil {
		_, metricsClient, _ := app.k8sClientMgr.GetDefaultClient()
		app.monitorMgr.SetK8sClients(clientset, metricsClient)
		app.logger.Info("K8s 监控已启用")
	}

	if err := app.configMgr.Watch(); err != nil {
		app.logger.Warn("启动配置监听失败", zap.Error(err))
	}

	app.logger.Info("应用初始化完成",
		zap.String("version", "2.0.0-optimized"),
		zap.Bool("configStoreEnabled", app.configStore != nil),
		zap.Bool("k8sMonitoringEnabled", clientset != nil),
	)

	return nil
}

// SetupRouter 设置路由
func (app *Application) SetupRouter() *gin.Engine {
	cfg := app.configMgr.GetConfig()

	if cfg.IsDevelopment() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	app.registerMiddlewares(r, cfg)
	app.registerRoutes(r, cfg)
	return r
}

// registerMiddlewares 注册中间件
func (app *Application) registerMiddlewares(r *gin.Engine, cfg *model.Config) {
	r.Use(middleware.Recovery(app.logger))
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.LoggingMiddleware(app.logger))
	r.Use(middleware.MetricsMiddleware(app.monitorMgr.GetMetrics()))

	if cfg.Auth.EnableRateLimit {
		r.Use(middleware.ConcurrencyMiddleware(cfg.Auth.RateLimit))
	}
}

// registerRoutes 注册路由
func (app *Application) registerRoutes(r *gin.Engine, cfg *model.Config) {
	r.POST(LoginPath, api.LoginHandler(app.logger))

	r.GET(CacheStatsPath, app.handleCacheStats)
	r.GET(HealthCheckPath, app.handleHealthCheck)

	apiGroup := r.Group(APIPrefix)
	apiGroup.Use(middleware.JWTAuthMiddleware(app.logger, app.configMgr))

	if cfg.Cache.Enabled {
		apiGroup.Use(middleware.CacheMiddleware(app.cacheMgr, cfg.Cache.TTL))
	}

	app.registerAPIRoutes(apiGroup)
}

// registerAPIRoutes 注册 API 路由
func (app *Application) registerAPIRoutes(apiGroup *gin.RouterGroup) {
	// 注册概览
	api.RegisterOverview(apiGroup, app.logger, app.getOverviewHandler())

	// 注册资源类型 - 【优化】使用新的 K8s 客户端管理器
	api.RegisterPod(apiGroup, app.logger, app.getK8sClient, service.ListPodsWithRaw)
	api.RegisterDeployment(apiGroup, app.logger, app.getK8sClient, service.ListDeployments)
	api.RegisterService(apiGroup, app.logger, app.getK8sClient, service.ListServices)
	api.RegisterNode(apiGroup, app.logger, app.getK8sClient, service.ListPodsWithRaw, service.ListNodes)
	api.RegisterNamespace(apiGroup, app.logger, app.getK8sClient, service.ListNamespaces)
	api.RegisterEvent(apiGroup, app.logger, app.getK8sClient, service.ListEvents)

	// 注册其他资源类型
	api.RegisterStatefulSet(apiGroup, app.logger, app.getK8sClient, service.ListStatefulSets)
	api.RegisterDaemonSet(apiGroup, app.logger, app.getK8sClient, service.ListDaemonSets)
	api.RegisterIngress(apiGroup, app.logger, app.getK8sClient, service.ListIngresses)
	api.RegisterCronJob(apiGroup, app.logger, app.getK8sClient, service.ListCronJobs)
	api.RegisterJob(apiGroup, app.logger, app.getK8sClient, service.ListJobs)
	api.RegisterPVC(apiGroup, app.logger, app.getK8sClient, service.ListPVCs)
	api.RegisterPV(apiGroup, app.logger, app.getK8sClient, service.ListPVs)
	api.RegisterStorageClass(apiGroup, app.logger, app.getK8sClient, service.ListStorageClasses)
	api.RegisterConfigMap(apiGroup, app.logger, app.getK8sClient, service.ListConfigMaps)
	api.RegisterSecret(apiGroup, app.logger, app.getK8sClient, service.ListSecrets)

	// 注册管理 API
	api.RegisterPasswordAdmin(apiGroup, app.logger)
	api.RegisterMetrics(apiGroup, app.logger)
}

// getK8sClient 【优化】获取 K8s 客户端（使用客户端管理器）
func (app *Application) getK8sClient() (*kubernetes.Clientset, *versioned.Clientset, error) {
	if app.k8sClientMgr == nil {
		return nil, nil, nil
	}
	return app.k8sClientMgr.GetDefaultClient()
}

// getOverviewHandler 获取概览处理器
func (app *Application) getOverviewHandler() func(limit, offset int) (*model.OverviewStatus, string, error) {
	return func(limit, offset int) (*model.OverviewStatus, string, error) {
		clientset, _, err := app.getK8sClient()
		if err != nil {
			return nil, "k8s client error", err
		}
		overview, err := service.GetOverviewStatus(clientset)
		if err != nil {
			return nil, "overview error", err
		}
		return overview, "", nil
	}
}

// handleCacheStats 处理缓存统计
func (app *Application) handleCacheStats(c *gin.Context) {
	stats := app.lruCacheMgr.GetAllStats()
	c.JSON(200, stats)
}

// handleHealthCheck 处理健康检查
func (app *Application) handleHealthCheck(c *gin.Context) {
	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "2.0.0-optimized",
	}

	// 添加 K8s 连接状态
	if app.k8sClientMgr != nil {
		clientset, _, err := app.k8sClientMgr.GetDefaultClient()
		if err == nil {
			_, err = clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{Limit: 1})
			health["k8sConnected"] = (err == nil)
		} else {
			health["k8sConnected"] = false
		}
	}

	c.JSON(200, health)
}

// Run 运行应用
func (app *Application) Run() error {
	cfg := app.configMgr.GetConfig()
	serverAddr := cfg.GetServerAddress()

	app.logger.Info("服务器启动",
		zap.String("address", serverAddr),
		zap.Bool("cacheEnabled", cfg.Cache.Enabled),
		zap.Bool("rateLimitEnabled", cfg.Auth.EnableRateLimit),
		zap.Bool("configStoreEnabled", app.configStore != nil),
	)

	// 启动定期监控（包含 K8s 集群指标）
	app.monitorMgr.StartPeriodicLogging(5 * time.Minute)

	router := app.SetupRouter()
	return router.Run(serverAddr)
}

// Close 关闭应用
func (app *Application) Close() {
	app.logger.Info("正在关闭应用...")

	if app.logger != nil {
		app.logger.Sync()
	}
	if app.cacheMgr != nil {
		app.cacheMgr.Close()
	}
	if app.monitorMgr != nil {
		app.monitorMgr.Close()
	}
	if app.configStore != nil {
		app.configStore.Stop()
	}
	if app.k8sClientMgr != nil {
		app.k8sClientMgr.Close()
	}
	if app.configMgr != nil {
		app.configMgr.Close()
	}

	app.logger.Info("应用已关闭")
}

func main() {
	flag.Parse()

	app := NewApplication(*configFile, *configMapNS)

	if err := app.Initialize(); err != nil {
		// 初始化失败时 logger 可能未完全初始化
		if app.logger != nil {
			app.logger.Fatal("应用初始化失败", zap.Error(err))
		} else {
			tempLogger, _ := zap.NewProduction()
			tempLogger.Fatal("应用初始化失败", zap.Error(err))
		}
	}

	defer app.Close()

	if err := app.Run(); err != nil {
		app.logger.Fatal("应用运行失败", zap.Error(err))
	}
}
