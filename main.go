package main

import (
	"context"
	"crypto/rand"
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
	"golang.org/x/crypto/bcrypt"

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

var configFile = flag.String("config", "", "Path to config file")

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

	app.logger.Info("Application initialization completed",
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
		tempLogger.Fatal("Failed to load config", zap.Error(err))
	}

	// 1.2 初始化正式 logger
	cfg := app.configMgr.GetConfig()
	var err error
	app.logger, err = initLogger(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	app.configMgr.UpdateLogger(app.logger)
	if syncErr := tempLogger.Sync(); syncErr != nil {
		// 忽略临时 logger 的 Sync 错误
		_ = syncErr
	}

	// 1.3 安全配置检查和自动生成（必须在验证配置之前）
	app.checkAndGenerateSecurityConfig()

	// 1.4 验证配置（必须在安全配置生成之后，重新获取最新配置）
	cfg = app.configMgr.GetConfig()
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// 1.5 初始化 LRU 缓存
	app.initLRUCache(cfg)

	// 1.6 初始化监控器
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
		app.logger.Warn("Failed to start config watcher", zap.Error(err))
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

// initK8sClient 初始化 K8s 客户端（可选）
func (app *Application) initK8sClient() error {
	var err error
	app.k8sClientMgr, err = service.NewClientManager(app.configMgr, app.logger)
	if err != nil {
		// K8s 客户端初始化失败时记录警告，但不阻止启动
		app.logger.Warn("K8s client initialization failed, K8s features will be unavailable",
			zap.Error(err),
			zap.String("hint", "Can be enabled by configuring kubeconfig or using in-cluster mode"),
		)
		// 不返回错误，允许服务继续启动
		return nil
	}
	return nil
}

// initK8sCache 初始化 K8s 缓存（Informer 机制）
// 注：已移除未使用的缓存管理器，只保留 PodInformer
func (app *Application) initK8sCache() {
	if app.k8sClientMgr == nil {
		app.logger.Debug("K8s client manager not initialized, skipping cache initialization")
		return
	}

	clientset, _, err := app.k8sClientMgr.GetDefaultClient()
	if err != nil || clientset == nil {
		app.logger.Warn("K8s client unavailable, skipping cache initialization")
		return
	}

	// 初始化 PodInformer 用于计算 Restarts
	podInformer := service.NewPodInformer(clientset, "")
	service.SetPodInformer(podInformer)
	go podInformer.Start(context.Background())
	app.logger.Info("Pod Informer started")
}

// initServices 初始化服务层
func (app *Application) initServices() {
	if app.k8sClientMgr == nil {
		return
	}
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
	if app.k8sClientMgr == nil {
		app.logger.Debug("K8s client manager not initialized, skipping monitoring initialization")
		return
	}

	clientset, metricsClient, err := app.k8sClientMgr.GetDefaultClient()
	if err != nil || clientset == nil {
		app.logger.Debug("K8s client unavailable, skipping monitoring initialization")
		return
	}

	monitor.InitBusinessMetrics(app.logger)

	app.monitorMgr.SetK8sClients(clientset, metricsClient)
	app.logger.Info("K8s monitoring enabled")
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
		return nil, nil, fmt.Errorf("kubernetes client manager unavailable")
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

	app.logger.Info("Server starting",
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
		app.logger.Info("HTTP server starting", zap.String("address", serverAddr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.logger.Error("Server failed", zap.Error(err))
		}
	}()

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.logger.Info("Exit signal received, gracefully shutting down server...")

	// 优雅关闭：等待活跃请求完成（最多等待 30 秒）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		app.logger.Error("Server shutdown failed", zap.Error(err))
		return err
	}

	app.logger.Info("HTTP server gracefully shutdown")
	return nil
}

// Close 关闭应用
func (app *Application) Close() {
	app.logger.Info("Shutting down application...")

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
		if syncErr := app.logger.Sync(); syncErr != nil {
			app.logger.Error("Failed to sync logger", zap.Error(syncErr))
		}
	}

	app.logger.Info("Application closed")
}

// checkAndGenerateSecurityConfig 检查并生成安全配置
func (app *Application) checkAndGenerateSecurityConfig() {
	cfg := app.configMgr.GetConfig()

	app.logger.Info("Starting security config check",
		zap.Bool("jwt_secret_empty", cfg.JWT.Secret == ""),
		zap.Bool("auth_password_empty", cfg.Auth.Password == ""),
	)

	needsSave := false
	shouldPersist := app.configMgr.GetConfigFile() != ""

	// 1. 检查 JWT Secret
	if cfg.JWT.Secret == "" {
		// 生成随机 JWT Secret
		secret := generateRandomString(64)
		cfg.JWT.Secret = secret

		// 关键：将 JWT Secret 更新到配置管理器（同时更新 viper 和 config 对象）
		app.configMgr.Set("jwt.secret", secret)

		app.logger.Info("JWT Secret auto-generated",
			zap.Int("length", len(secret)),
			zap.String("hint", "It is recommended to save the JWT Secret to the config file or set it via the K8SVISION_JWT_SECRET environment variable"),
		)
		needsSave = true
	} else if len(cfg.JWT.Secret) < 32 {
		app.logger.Fatal("JWT Secret length is less than 32 characters, please set it via the K8SVISION_JWT_SECRET environment variable")
	}

	// 2. 检查密码
	if cfg.Auth.Password == "" {
		// 生成随机密码并哈希
		generatedPassword := generateRandomString(16)

		// 使用 bcrypt 哈希密码
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(generatedPassword), bcrypt.DefaultCost)
		if err != nil {
			app.logger.Fatal("Password hashing failed", zap.Error(err))
		}

		// 将哈希后的密码转换为字符串
		hashedPasswordStr := string(hashedPassword)

		// 关键：将哈希后的密码更新到配置管理器（同时更新 viper 和 config 对象）
		app.configMgr.Set("auth.password", hashedPasswordStr)

		app.logger.Warn("Admin password not configured",
			zap.String("username", cfg.Auth.Username),
			zap.String("hint", "A random password hash has been generated and is ready for persistence. Please reset the password via the admin interface or set it explicitly in the config file."),
		)
		needsSave = true
	}

	// 如果需要保存配置
	if needsSave {
		if shouldPersist {
			if err := app.configMgr.WriteConfigWithBackup(); err != nil {
				app.logger.Fatal("Security config persistence failed", zap.Error(err))
			}
			app.logger.Info("Security config updated and persisted")
		} else {
			app.logger.Warn("Security config only in memory, please provide config.yaml or environment variables for persistence")
		}
	}

	app.logger.Info("Security config check completed",
		zap.Int("jwt_secret_final_length", len(app.configMgr.GetConfig().JWT.Secret)),
		zap.Int("auth_password_final_length", len(app.configMgr.GetConfig().Auth.Password)),
	)
}

// generateRandomString 生成随机字符串（使用 crypto/rand）
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	bytes := make([]byte, length)
	for i := range bytes {
		// 使用 crypto/rand 生成随机索引
		randByte := make([]byte, 1)
		if _, err := rand.Read(randByte); err == nil {
			bytes[i] = charset[int(randByte[0])%len(charset)]
		}
	}
	return string(bytes)
}

func main() {
	flag.Parse()

	app := NewApplication(*configFile)

	// 初始化应用
	if err := app.Initialize(); err != nil {
		if app.logger != nil {
			app.logger.Fatal("Application initialization failed", zap.Error(err))
		}
		fmt.Fprintf(os.Stderr, "Application initialization failed: %v\n", err)
		os.Exit(1)
	}

	// 确保资源清理
	defer app.Close()

	// 运行应用
	if err := app.Run(); err != nil {
		app.logger.Fatal("Application run failed", zap.Error(err))
	}
}
