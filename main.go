// Package main K8sVision 主程序

package main

import (
	"flag"
	"time"

	"github.com/nick0323/K8sVision/api"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/cache"
	"github.com/nick0323/K8sVision/config"
	"github.com/nick0323/K8sVision/monitor"
	"github.com/nick0323/K8sVision/service"

	"github.com/nick0323/K8sVision/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	DefaultConfigFile = ""
	HealthCheckPath   = "/health"
	CacheStatsPath    = "/cache/stats"
	APIPrefix         = "/api"
	LoginPath         = "/api/login"
	ConsoleFormat     = "console"
	JSONFormat        = "json"
)

type Application struct {
	configFile string
	logger     *zap.Logger
	configMgr  *config.Manager
	cacheMgr   *cache.Manager
	monitorMgr *monitor.Monitor
}

var (
	configFile  = flag.String("config", DefaultConfigFile, "配置文件路径")
	logLevelMap = map[string]zap.AtomicLevel{
		"debug": zap.NewAtomicLevelAt(zap.DebugLevel),
		"info":  zap.NewAtomicLevelAt(zap.InfoLevel),
		"warn":  zap.NewAtomicLevelAt(zap.WarnLevel),
		"error": zap.NewAtomicLevelAt(zap.ErrorLevel),
	}
)

func initLogger(cfg *model.Config) (*zap.Logger, error) {
	var zapConfig zap.Config

	if cfg.IsDevelopment() {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	if level, exists := logLevelMap[cfg.Log.Level]; exists {
		zapConfig.Level = level
	} else {
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	if cfg.Log.Format == ConsoleFormat {
		zapConfig.Encoding = ConsoleFormat
	} else {
		zapConfig.Encoding = JSONFormat
	}

	// 配置时间戳格式为可读格式
	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")

	return zapConfig.Build()
}

func NewApplication(configFile string) *Application {
	return &Application{
		configFile: configFile,
	}
}

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
	app.cacheMgr = cache.NewManager(&cfg.Cache, app.logger)
	app.monitorMgr = monitor.NewMonitor(app.logger)

	service.SetConfigManager(app.configMgr)
	service.SetCacheManager(app.cacheMgr)
	api.SetConfigManager(app.configMgr)
	api.InitAuthManager(app.logger)

	monitor.InitTracing(app.logger)
	monitor.InitBusinessMetrics(app.logger)

	if err := app.configMgr.Watch(); err != nil {
		app.logger.Warn("启动配置监听失败", zap.Error(err))
	}

	return nil
}

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

func (app *Application) registerMiddlewares(r *gin.Engine, cfg *model.Config) {
	r.Use(middleware.Recovery(app.logger))
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.LoggingMiddleware(app.logger))
	r.Use(middleware.MetricsMiddleware(app.monitorMgr.GetMetrics()))

	if cfg.Auth.EnableRateLimit {
		r.Use(middleware.ConcurrencyMiddleware(cfg.Auth.RateLimit))
	}
}

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

func (app *Application) registerAPIRoutes(apiGroup *gin.RouterGroup) {
	// 注册概览
	api.RegisterOverview(apiGroup, app.logger, app.getOverviewHandler())

	// 注册资源类型（使用原有的注册函数）
	api.RegisterPod(apiGroup, app.logger, service.GetK8sClient, service.ListPodsWithRaw)
	api.RegisterDeployment(apiGroup, app.logger, service.GetK8sClient, service.ListDeployments)
	api.RegisterService(apiGroup, app.logger, service.GetK8sClient, service.ListServices)
	api.RegisterNode(apiGroup, app.logger, service.GetK8sClient, service.ListPodsWithRaw, service.ListNodes)
	api.RegisterNamespace(apiGroup, app.logger, service.GetK8sClient, service.ListNamespaces)
	api.RegisterEvent(apiGroup, app.logger, service.GetK8sClient, service.ListEvents)

	// 注册其他资源类型
	api.RegisterStatefulSet(apiGroup, app.logger, service.GetK8sClient, service.ListStatefulSets)
	api.RegisterDaemonSet(apiGroup, app.logger, service.GetK8sClient, service.ListDaemonSets)
	api.RegisterIngress(apiGroup, app.logger, service.GetK8sClient, service.ListIngresses)
	api.RegisterCronJob(apiGroup, app.logger, service.GetK8sClient, service.ListCronJobs)
	api.RegisterJob(apiGroup, app.logger, service.GetK8sClient, service.ListJobs)
	api.RegisterPVC(apiGroup, app.logger, service.GetK8sClient, service.ListPVCs)
	api.RegisterPV(apiGroup, app.logger, service.GetK8sClient, service.ListPVs)
	api.RegisterStorageClass(apiGroup, app.logger, service.GetK8sClient, service.ListStorageClasses)
	api.RegisterConfigMap(apiGroup, app.logger, service.GetK8sClient, service.ListConfigMaps)
	api.RegisterSecret(apiGroup, app.logger, service.GetK8sClient, service.ListSecrets)

	// 注册管理API
	api.RegisterPasswordAdmin(apiGroup, app.logger)
	api.RegisterMetrics(apiGroup, app.logger)
}

func (app *Application) getOverviewHandler() func(limit, offset int) (*model.OverviewStatus, string, error) {
	return func(limit, offset int) (*model.OverviewStatus, string, error) {
		clientset, _, err := service.GetK8sClient()
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

func (app *Application) handleCacheStats(c *gin.Context) {
	stats := app.cacheMgr.GetAllStats()
	c.JSON(200, stats)
}

func (app *Application) handleHealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	})
}

func (app *Application) Run() error {
	cfg := app.configMgr.GetConfig()
	serverAddr := cfg.GetServerAddress()

	app.logger.Info("服务器启动",
		zap.String("address", serverAddr),
		zap.Bool("cacheEnabled", cfg.Cache.Enabled),
		zap.Bool("rateLimitEnabled", cfg.Auth.EnableRateLimit),
	)

	app.monitorMgr.StartPeriodicLogging(5 * time.Minute)
	router := app.SetupRouter()
	return router.Run(serverAddr)
}

func (app *Application) Close() {
	if app.logger != nil {
		app.logger.Sync()
	}
	if app.cacheMgr != nil {
		app.cacheMgr.Close()
	}
	if app.monitorMgr != nil {
		app.monitorMgr.Close()
	}
	if app.configMgr != nil {
		app.configMgr.Close()
	}
}

func main() {
	flag.Parse()

	app := NewApplication(*configFile)

	if err := app.Initialize(); err != nil {
		app.logger.Fatal("应用初始化失败", zap.Error(err))
	}

	defer app.Close()

	if err := app.Run(); err != nil {
		app.logger.Fatal("应用运行失败", zap.Error(err))
	}
}
