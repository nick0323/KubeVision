package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nick0323/K8sVision/api"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/cache"
	"github.com/nick0323/K8sVision/config"
	"github.com/nick0323/K8sVision/model"
	"github.com/nick0323/K8sVision/monitor"
	"github.com/nick0323/K8sVision/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

// 常量定义
const (
	// 版本信息
	Version = "2.0.0-optimized"

	// 监控配置
	MonitorInterval = 5 * time.Minute // 监控数据采集间隔

	// 路由路径
	HealthCheckPath = "/health"
	CacheStatsPath  = "/cache/stats"
	APIPrefix       = "/api"
	LoginPath       = "/api/login"

	// 缓存配置
	cacheSyncTimeout = 2 * time.Minute // 缓存同步超时时间
)

// Application 应用结构
type Application struct {
	configFile   string
	logger       *zap.Logger
	configMgr    *config.Manager
	lruCacheMgr  *cache.MemoryCache[interface{}] // LRU 缓存管理器（用于统计）
	monitorMgr   *monitor.Monitor
	k8sClientMgr *service.ClientManager
}

var configFile = flag.String("config", "", "配置文件路径")

// initLogger 初始化日志
func initLogger(cfg *model.Config) (*zap.Logger, error) {
	var zapConfig zap.Config

	if cfg.IsDevelopment() {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	// 设置日志级别
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

	// 设置日志格式
	if cfg.Log.Format == "console" {
		zapConfig.Encoding = "console"
	} else {
		zapConfig.Encoding = "json"
	}

	// 配置时间戳格式
	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")

	return zapConfig.Build()
}

// NewApplication 创建应用实例
func NewApplication(configFile string) *Application {
	return &Application{
		configFile: configFile,
	}
}

// Initialize 初始化应用（分三个阶段）
func (app *Application) Initialize() error {
	// ========== 阶段 1：基础组件初始化 ==========
	if err := app.initBaseComponents(); err != nil {
		return err
	}

	// ========== 阶段 2：K8s 组件初始化 ==========
	if err := app.initK8sComponents(); err != nil {
		return err
	}

	// ========== 阶段 3：业务组件初始化 ==========
	app.initBusinessComponents()

	app.logger.Info("应用初始化完成",
		zap.String("version", Version),
		zap.Bool("k8sEnabled", app.k8sClientMgr != nil),
	)

	return nil
}

// initBaseComponents 阶段 1：初始化基础组件（配置、日志、缓存）
func (app *Application) initBaseComponents() error {
	// 1.1 初始化配置管理器（使用临时 logger）
	tempLogger, _ := zap.NewProduction()
	app.configMgr = config.NewManager(tempLogger)

	if err := app.configMgr.Load(app.configFile); err != nil {
		tempLogger.Fatal("加载配置失败", zap.Error(err))
	}

	// 1.2 初始化正式 logger
	cfg := app.configMgr.GetConfig()
	var err error
	app.logger, err = initLogger(cfg)
	if err != nil {
		return fmt.Errorf("初始化日志失败：%w", err)
	}
	app.configMgr.UpdateLogger(app.logger)
	tempLogger.Sync()

	// 1.3 初始化 LRU 缓存
	app.initLRUCache(cfg)

	// 1.4 初始化监控器
	app.initMonitor()

	return nil
}

// initK8sComponents 阶段 2：初始化 K8s 组件（客户端、缓存）
func (app *Application) initK8sComponents() error {
	// 2.1 初始化 K8s 客户端
	if err := app.initK8sClient(); err != nil {
		return err
	}

	// 2.2 初始化 K8s 缓存（Informer 机制）
	app.initK8sCache()

	return nil
}

// initBusinessComponents 阶段 3：初始化业务组件（服务、API、监控）
func (app *Application) initBusinessComponents() {
	// 3.1 初始化服务层
	app.initServices()

	// 3.2 初始化 API 层
	app.initAPI()

	// 3.3 初始化监控系统
	app.initMonitoring()

	// 3.4 启动配置监听（可选）
	if err := app.configMgr.Watch(); err != nil {
		app.logger.Warn("启动配置监听失败", zap.Error(err))
	}
}

// initLRUCache 初始化 LRU 缓存（用于统计）
func (app *Application) initLRUCache(cfg *model.Config) {
	app.lruCacheMgr = cache.NewMemoryCache(&cfg.Cache, app.logger)
}

// initMonitor 初始化监控器
func (app *Application) initMonitor() {
	app.monitorMgr = monitor.NewMonitor(app.logger)
}

// initK8sClient 初始化 K8s 客户端
func (app *Application) initK8sClient() error {
	var err error
	app.k8sClientMgr, err = service.NewClientManager(app.configMgr, app.logger)
	if err != nil {
		return fmt.Errorf("初始化 K8s 客户端失败：%w", err)
	}
	return nil
}

// initK8sCache 初始化 K8s 缓存（Informer 机制）
// 注：已移除未使用的缓存管理器，只保留 PodInformer
func (app *Application) initK8sCache() {
	clientset, _, err := app.k8sClientMgr.GetDefaultClient()
	if err != nil || clientset == nil {
		app.logger.Warn("K8s 客户端不可用，跳过缓存初始化")
		return
	}

	// 初始化 PodInformer 用于计算 Restarts
	podInformer := service.NewPodInformer(clientset, "")
	service.SetPodInformer(podInformer)
	go podInformer.Start(context.Background())
	app.logger.Info("Pod Informer 已启动")
}

// initServices 初始化服务层
func (app *Application) initServices() {
	clientset, _, _ := app.k8sClientMgr.GetDefaultClient()
	if clientset == nil {
		return
	}

	// PodInformer 已在 initK8sCache 中初始化
}

// initAPI 初始化 API 层
func (app *Application) initAPI() {
	cfg := app.configMgr.GetConfig()

	// 设置配置管理器
	api.SetConfigManager(app.configMgr)

	// 初始化认证管理器
	api.InitAuthManager(app.logger)

	// 设置全局 JWT secret（用于 WebSocket token 验证）
	middleware.SetJWTSecret([]byte(cfg.JWT.Secret))

	// 设置全局 K8s ClientManager（用于 WebSocket exec 等场景）
	api.SetGlobalClientManager(app.k8sClientMgr)

	// 初始化 WebSocket upgrader（配置允许的源）
	api.InitWebSocketUpgrader(cfg.Server.AllowedOrigin)
}

// initMonitoring 初始化监控系统
func (app *Application) initMonitoring() {
	clientset, metricsClient, err := app.k8sClientMgr.GetDefaultClient()
	if err != nil || clientset == nil {
		return
	}

	monitor.InitTracing(app.logger)
	monitor.InitBusinessMetrics(app.logger)

	app.monitorMgr.SetK8sClients(clientset, metricsClient)
	app.logger.Info("K8s 监控已启用")
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

	// 开发环境启用 CORS
	if cfg.IsDevelopment() {
		r.Use(middleware.CORSMiddleware(nil)) // 使用默认配置
	}

	if cfg.Auth.EnableRateLimit {
		r.Use(middleware.ConcurrencyMiddleware(app.logger, cfg.Auth.RateLimit))
	}
}

// registerRoutes 注册路由
func (app *Application) registerRoutes(r *gin.Engine, cfg *model.Config) {
	r.POST(LoginPath, api.LoginHandler(app.logger))

	r.GET(CacheStatsPath, app.handleCacheStats)
	r.GET(HealthCheckPath, app.handleHealthCheck)

	apiGroup := r.Group(APIPrefix)
	apiGroup.Use(middleware.JWTAuthMiddleware(app.logger, app.configMgr))

	app.registerAPIRoutes(apiGroup)
}

// registerAPIRoutes 注册 API 路由
func (app *Application) registerAPIRoutes(apiGroup *gin.RouterGroup) {
	// 注册概览
	clientset, _, _ := app.getK8sClient()
	overviewService := service.NewOverviewService(clientset)
	api.RegisterOverview(apiGroup, app.logger, func() (*model.OverviewStatus, error) {
		return overviewService.GetOverview(context.Background())
	})

	// 注册资源操作接口（YAML、关联资源等）
	api.RegisterOperations(apiGroup, app.logger, app.getK8sClient)

	// 注册 Exec WebSocket
	api.RegisterExecWS(apiGroup, app.logger, app.getK8sClient, app.configMgr)

	// 注册日志流 WebSocket
	api.RegisterLogStream(apiGroup, app.logger, app.getK8sClient)

	// 注册 K8s Metrics 接口
	api.RegisterK8sMetricsRoutes(apiGroup, app.logger, func() (*versioned.Clientset, error) {
		_, metricsClient, _ := app.k8sClientMgr.GetDefaultClient()
		return metricsClient, nil
	})

	// 注册通用资源接口（动态路由 /api/:resourceType）
	// 支持所有 Kubernetes 资源：
	// - 工作负载：pods, deployments, statefulsets, daemonsets, jobs, cronjobs
	// - 服务网络：services, ingresses, endpoints
	// - 配置：configmaps, secrets
	// - 存储：pvcs, pvs, storageclasses
	// - 集群：namespaces, nodes, events
	api.RegisterRoutes(apiGroup, app.logger, app.getK8sClient)

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

// handleCacheStats 处理缓存统计
func (app *Application) handleCacheStats(c *gin.Context) {
	stats := app.lruCacheMgr.GetStats()
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

// Run 运行应用（带优雅关闭）
func (app *Application) Run() error {
	cfg := app.configMgr.GetConfig()
	serverAddr := cfg.GetServerAddress()

	app.logger.Info("服务器启动",
		zap.String("address", serverAddr),
		zap.Bool("cacheEnabled", cfg.Cache.Enabled),
		zap.Bool("rateLimitEnabled", cfg.Auth.EnableRateLimit),
	)

	// 启动定期监控
	app.monitorMgr.StartPeriodicLogging(MonitorInterval)

	router := app.SetupRouter()

	// 创建 http.Server 实例（支持优雅关闭）
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	// 启动服务器（在 goroutine 中）
	go func() {
		app.logger.Info("HTTP 服务器开始监听", zap.String("address", serverAddr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.logger.Error("服务器运行失败", zap.Error(err))
		}
	}()

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.logger.Info("收到退出信号，正在优雅关闭服务器...")

	// 优雅关闭：等待活跃请求完成（最多等待 30 秒）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		app.logger.Error("服务器优雅关闭失败", zap.Error(err))
		return err
	}

	app.logger.Info("HTTP 服务器已优雅关闭")
	return nil
}

// Close 关闭应用
func (app *Application) Close() {
	app.logger.Info("正在关闭应用...")

	// 按顺序关闭各个组件
	if app.monitorMgr != nil {
		app.monitorMgr.Close()
	}
	if app.lruCacheMgr != nil {
		app.lruCacheMgr.Close()
	}
	if app.k8sClientMgr != nil {
		app.k8sClientMgr.Close()
	}
	if app.configMgr != nil {
		app.configMgr.Close()
	}
	if app.logger != nil {
		app.logger.Sync()
	}

	app.logger.Info("应用已关闭")
}

func main() {
	flag.Parse()

	app := NewApplication(*configFile)

	// 初始化应用
	if err := app.Initialize(); err != nil {
		if app.logger != nil {
			app.logger.Fatal("应用初始化失败", zap.Error(err))
		}
		fmt.Fprintf(os.Stderr, "应用初始化失败：%v\n", err)
		os.Exit(1)
	}

	// 确保资源清理
	defer app.Close()

	// 运行应用
	if err := app.Run(); err != nil {
		app.logger.Fatal("应用运行失败", zap.Error(err))
	}
}
