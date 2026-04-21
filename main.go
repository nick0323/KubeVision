package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nick0323/K8sVision/cache"
	"github.com/nick0323/K8sVision/config"
	"github.com/nick0323/K8sVision/internal/bootstrap"
	"github.com/nick0323/K8sVision/internal/http"
	"github.com/nick0323/K8sVision/model"
	"github.com/nick0323/K8sVision/service"
	"go.uber.org/zap"
)

func main() {
	configFile := flag.String("config", "", "Path to config file")
	flag.Parse()

	configMgr := config.NewManager(zap.NewNop())
	if err := configMgr.Load(*configFile); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger, lruCacheMgr, err := bootstrap.NewInitializer(zap.NewNop(), configMgr).InitBaseComponents(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize base components: %v\n", err)
		os.Exit(1)
	}

	k8sClientMgr, err := bootstrap.NewInitializer(logger, configMgr).InitK8sComponents(logger)
	if err != nil {
		logger.Fatal("Failed to initialize K8s components", zap.Error(err))
	}

	bootstrap.InitServices(k8sClientMgr, logger)
	bootstrap.InitAPI(configMgr, k8sClientMgr, logger)

	logger.Info("Application initialization completed",
		zap.String("version", model.Version),
		zap.Bool("k8sEnabled", k8sClientMgr != nil),
	)

	app := NewApplication(logger, configMgr, lruCacheMgr, k8sClientMgr)
	defer app.Close()

	if err := app.Run(); err != nil {
		logger.Fatal("Application run failed", zap.Error(err))
	}
}

type Application struct {
	logger       *zap.Logger
	configMgr    *config.Manager
	lruCacheMgr  *cache.MemoryCache[interface{}]
	k8sClientMgr *service.ClientManager
}

func NewApplication(
	logger *zap.Logger,
	configMgr *config.Manager,
	lruCacheMgr *cache.MemoryCache[interface{}],
	k8sClientMgr *service.ClientManager,
) *Application {
	return &Application{
		logger:       logger,
		configMgr:    configMgr,
		lruCacheMgr:  lruCacheMgr,
		k8sClientMgr: k8sClientMgr,
	}
}

func (app *Application) Run() error {
	httpServer := http.NewServer(app.logger, app.configMgr, app.lruCacheMgr, app.k8sClientMgr)
	if err := httpServer.Run(); err != nil {
		return err
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.logger.Info("Exit signal received, gracefully shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		app.logger.Error("Server shutdown failed", zap.Error(err))
	}

	return nil
}

func (app *Application) Close() {
	app.logger.Info("Shutting down application...")

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
