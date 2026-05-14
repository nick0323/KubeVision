package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/nick0323/K8sVision/api"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/cache"
	"github.com/nick0323/K8sVision/config"
	"github.com/nick0323/K8sVision/model"
	"github.com/nick0323/K8sVision/service"
	"go.uber.org/zap"
)

type Server struct {
	logger        *zap.Logger
	configMgr     *config.Manager
	lruCacheMgr   *cache.MemoryCache[interface{}]
	k8sClientMgr  *service.ClientManager
	httpServer    *http.Server
	jwtMiddleware *middleware.JWTMiddleware
}

func NewServer(
	logger *zap.Logger,
	configMgr *config.Manager,
	lruCacheMgr *cache.MemoryCache[interface{}],
	k8sClientMgr *service.ClientManager,
) *Server {
	return &Server{
		logger:       logger,
		configMgr:    configMgr,
		lruCacheMgr:  lruCacheMgr,
		k8sClientMgr: k8sClientMgr,
	}
}

func (s *Server) Run() error {
	cfg := s.configMgr.GetConfig()
	serverAddr := cfg.GetServerAddress()

	s.logger.Info("Server starting",
		zap.String("address", serverAddr),
		zap.Bool("cacheEnabled", cfg.Cache.Enabled),
	)

	s.httpServer = &http.Server{
		Addr:         serverAddr,
		Handler:      s.SetupRouter(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		s.logger.Info("HTTP server starting", zap.String("address", serverAddr))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Server failed", zap.Error(err))
		}
	}()

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Gracefully shutting down server...")

	if s.jwtMiddleware != nil {
		s.jwtMiddleware.Close()
	}

	if wsMgr := api.GetWebSocketManager(); wsMgr != nil {
		if ctx == nil {
			ctx = context.Background()
		}
		wsCtx, wsCancel := context.WithTimeout(ctx, 10*time.Second)
		defer wsCancel()
		if err := wsMgr.Shutdown(wsCtx); err != nil {
			s.logger.Warn("WebSocket shutdown timed out, proceeding with server shutdown", zap.Error(err))
		} else {
			s.logger.Info("All WebSocket connections closed gracefully")
		}
	}

	if s.httpServer != nil {
		if ctx == nil {
			ctx = context.Background()
		}
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

func (s *Server) SetupRouter() *gin.Engine {
	cfg := s.configMgr.GetConfig()

	if cfg.IsDevelopment() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	// 全局请求体大小限制：10MB
	r.MaxMultipartMemory = 10 << 20
	s.registerMiddlewares(r, cfg)
	s.registerRoutes(r, cfg)
	return r
}

func (s *Server) registerMiddlewares(r *gin.Engine, cfg *model.Config) {
	r.Use(middleware.Recovery(s.logger))
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.LoggingMiddleware(s.logger))
	r.Use(middleware.MetricsMiddleware())

	if cfg.IsDevelopment() {
		r.Use(middleware.CORSMiddleware(nil))
	}

	// Content-Security-Policy 保护
	r.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:; connect-src 'self' ws: wss:")
	})
}

func (s *Server) registerRoutes(r *gin.Engine, cfg *model.Config) {
	authManager, _ := api.InitAuthManager(s.logger, s.configMgr)
	loginHandler := api.NewLoginHandler(authManager, s.configMgr, s.logger)
	// 登录端点添加 IP 速率限制：每秒最多 3 次尝试
	loginRateLimiter := middleware.NewIPRateLimiter(3, 6)
	r.POST(model.LoginPath, loginRateLimiter.Middleware(), loginHandler.Handle())

	r.GET(model.HealthCheckPath, s.healthCheckHandler)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	apiGroup := r.Group(model.APIPrefix)
	s.jwtMiddleware = middleware.NewJWTMiddleware(s.configMgr.GetJWTSecret(), s.logger)
	apiGroup.Use(s.jwtMiddleware.AuthMiddleware(s.configMgr))

	// 刷新令牌端点（需要在 jwtMiddleware 初始化之后，以获取黑名单实例）
	refreshRateLimiter := middleware.NewIPRateLimiter(5, 10)
	r.POST(model.RefreshPath, refreshRateLimiter.Middleware(), loginHandler.Refresh(s.jwtMiddleware.GetBlacklist()))

	apiGroup.POST("/logout", loginHandler.Logout(s.jwtMiddleware.GetBlacklist()))

	s.registerAPIRoutes(apiGroup)
}

func (s *Server) healthCheckHandler(c *gin.Context) {
	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   model.Version,
	}

	if s.k8sClientMgr != nil {
		clientset, err := s.k8sClientMgr.GetDefaultClient()
		if err == nil {
			// 添加 5 秒超时控制，防止 K8s API Server 无响应时阻塞
			ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
			defer cancel()
			_, err = clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
			health["k8sConnected"] = (err == nil)
		} else {
			health["k8sConnected"] = false
		}
	}

	c.JSON(200, health)
}

func (s *Server) registerAPIRoutes(apiGroup *gin.RouterGroup) {
	cfg := s.configMgr.GetConfig()
	api.InitWebSocketManager(cfg.Server.MaxWsConnections)

	api.RegisterOverview(apiGroup, s.logger, s.getK8sClient)

	api.RegisterOperations(apiGroup, s.logger, s.getK8sClient)
	api.RegisterExecWS(apiGroup, s.logger, &serverClientProvider{mgr: s.k8sClientMgr}, s.configMgr)
	api.RegisterLogStream(apiGroup, s.logger, s.getK8sClient, s.configMgr)
	api.RegisterRoutes(apiGroup, s.logger, s.getK8sClient, s.lruCacheMgr)
	api.RegisterPasswordAdmin(apiGroup, s.configMgr, s.logger)
	api.RegisterArgoCDRoutes(apiGroup, s.logger, s.k8sClientMgr)
	api.RegisterCRDRoutes(apiGroup, s.logger, s.k8sClientMgr, s.lruCacheMgr)
	apiGroup.GET("/clusters", func(c *gin.Context) {
		names := s.k8sClientMgr.GetClusterNames()
		c.JSON(200, gin.H{"data": names})
	})
	apiGroup.GET("/clusters/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()
		health := s.k8sClientMgr.GetClustersHealth(ctx)
		c.JSON(200, gin.H{"data": health})
	})
}

type serverClientProvider struct {
	mgr *service.ClientManager
}

func (s *serverClientProvider) GetClientset(cluster string) (*kubernetes.Clientset, error) {
	return s.mgr.GetClient(cluster)
}

func (s *serverClientProvider) GetRESTConfig(cluster string) (*rest.Config, error) {
	cfg := s.mgr.GetClientRESTConfig(cluster)
	if cfg == nil {
		return nil, fmt.Errorf("rest config not available for cluster %s", cluster)
	}
	return cfg, nil
}

func (s *Server) getK8sClient(cluster string) (kubernetes.Interface, interface{}, error) {
	if s.k8sClientMgr == nil {
		return nil, nil, fmt.Errorf("kubernetes client manager unavailable")
	}
	clientset, err := s.k8sClientMgr.GetClient(cluster)
	if err != nil {
		return nil, nil, err
	}

	// 创建 metrics 客户端
	restConfig := s.k8sClientMgr.GetClientRESTConfig(cluster)
	if restConfig == nil {
		return clientset, nil, nil
	}

	metricsClient, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return clientset, nil, nil
	}

	return clientset, metricsClient, nil
}
