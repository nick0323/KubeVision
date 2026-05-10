package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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
	if s.httpServer != nil {
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
	s.registerMiddlewares(r, cfg)
	s.registerRoutes(r, cfg)
	return r
}

func (s *Server) registerMiddlewares(r *gin.Engine, cfg *model.Config) {
	r.Use(middleware.Recovery(s.logger))
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.LoggingMiddleware(s.logger))

	if cfg.IsDevelopment() {
		r.Use(middleware.CORSMiddleware(nil))
	}
}

func (s *Server) registerRoutes(r *gin.Engine, cfg *model.Config) {
	authManager, _ := api.InitAuthManager(s.logger, s.configMgr)
	loginHandler := api.NewLoginHandler(authManager, s.configMgr, s.logger)
	r.POST(model.LoginPath, loginHandler.Handle())

	r.GET(model.HealthCheckPath, s.healthCheckHandler)

	apiGroup := r.Group(model.APIPrefix)
	s.jwtMiddleware = middleware.NewJWTMiddleware(s.configMgr.GetJWTSecret(), s.logger)
	apiGroup.Use(s.jwtMiddleware.AuthMiddleware(s.configMgr))

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

	clientset, _, err := s.getK8sClient()
	if err != nil {
		s.logger.Warn("Failed to get K8s client for overview service", zap.Error(err))
		clientset = nil
	}
	overviewService := service.NewOverviewService(clientset)
	api.RegisterOverview(apiGroup, s.logger, func() (*model.OverviewStatus, error) {
		return overviewService.GetOverview(context.Background())
	})

	api.RegisterOperations(apiGroup, s.logger, s.getK8sClient)
	api.RegisterExecWS(apiGroup, s.logger, &serverClientProvider{mgr: s.k8sClientMgr}, s.configMgr)
	api.RegisterLogStream(apiGroup, s.logger, s.getK8sClient, s.configMgr)
	api.RegisterRoutes(apiGroup, s.logger, s.getK8sClient, s.lruCacheMgr)
	api.RegisterPasswordAdmin(apiGroup, s.configMgr, s.logger)
	api.RegisterArgoCDRoutes(apiGroup, s.logger, s.k8sClientMgr)
	api.RegisterCRDRoutes(apiGroup, s.logger, s.k8sClientMgr)
}

type serverClientProvider struct {
	mgr *service.ClientManager
}

func (s *serverClientProvider) GetClientset() (*kubernetes.Clientset, error) {
	return s.mgr.GetDefaultClient()
}

func (s *serverClientProvider) GetRESTConfig() (*rest.Config, error) {
	return s.mgr.GetDefaultRESTConfig(), nil
}

func (s *Server) getK8sClient() (*kubernetes.Clientset, *versioned.Clientset, error) {
	if s.k8sClientMgr == nil {
		return nil, nil, fmt.Errorf("kubernetes client manager unavailable")
	}
	clientset, err := s.k8sClientMgr.GetDefaultClient()
	if err != nil {
		return nil, nil, err
	}

	// 创建 metrics 客户端
	restConfig := s.k8sClientMgr.GetDefaultRESTConfig()
	if restConfig == nil {
		return clientset, nil, nil
	}

	metricsClient, err := versioned.NewForConfig(restConfig)
	if err != nil {
		// metrics 客户端创建失败不影响主功能
		return clientset, nil, nil
	}

	return clientset, metricsClient, nil
}
