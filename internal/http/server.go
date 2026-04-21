package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/nick0323/K8sVision/api"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/cache"
	"github.com/nick0323/K8sVision/config"
	"github.com/nick0323/K8sVision/model"
	"github.com/nick0323/K8sVision/service"
	"go.uber.org/zap"
)

type Server struct {
	logger       *zap.Logger
	configMgr    *config.Manager
	lruCacheMgr  *cache.MemoryCache[interface{}]
	k8sClientMgr *service.ClientManager
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

	srv := &http.Server{
		Addr:    serverAddr,
		Handler: s.SetupRouter(),
	}

	go func() {
		s.logger.Info("HTTP server starting", zap.String("address", serverAddr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Server failed", zap.Error(err))
		}
	}()

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Gracefully shutting down server...")
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
	r.POST(model.LoginPath, api.LoginHandler(s.logger))

	r.GET(model.CacheStatsPath, func(c *gin.Context) {
		stats := s.lruCacheMgr.GetStats()
		c.JSON(200, stats)
	})

	r.GET(model.HealthCheckPath, s.healthCheckHandler)

	apiGroup := r.Group(model.APIPrefix)
	apiGroup.Use(middleware.JWTAuthMiddleware(s.logger, s.configMgr))
	s.registerAPIRoutes(apiGroup)
}

func (s *Server) healthCheckHandler(c *gin.Context) {
	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   model.Version,
	}

	if s.k8sClientMgr != nil {
		clientset, _, err := s.k8sClientMgr.GetDefaultClient()
		if err == nil {
			_, err = clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{Limit: 1})
			health["k8sConnected"] = (err == nil)
		} else {
			health["k8sConnected"] = false
		}
	}

	c.JSON(200, health)
}

func (s *Server) registerAPIRoutes(apiGroup *gin.RouterGroup) {
	clientset, _, _ := s.getK8sClient()
	overviewService := service.NewOverviewService(clientset)
	api.RegisterOverview(apiGroup, s.logger, func() (*model.OverviewStatus, error) {
		return overviewService.GetOverview(context.Background())
	})

	api.RegisterOperations(apiGroup, s.logger, s.getK8sClient)
	api.RegisterExecWS(apiGroup, s.logger, s.getK8sClient, s.configMgr)
	api.RegisterLogStream(apiGroup, s.logger, s.getK8sClient)
	api.RegisterK8sMetricsRoutes(apiGroup, s.logger, func() (*versioned.Clientset, error) {
		_, metricsClient, _ := s.k8sClientMgr.GetDefaultClient()
		return metricsClient, nil
	})
	api.RegisterRoutes(apiGroup, s.logger, s.getK8sClient)
	api.RegisterPasswordAdmin(apiGroup, s.logger)
}

func (s *Server) getK8sClient() (*kubernetes.Clientset, *versioned.Clientset, error) {
	if s.k8sClientMgr == nil {
		return nil, nil, fmt.Errorf("kubernetes client manager unavailable")
	}
	return s.k8sClientMgr.GetDefaultClient()
}
