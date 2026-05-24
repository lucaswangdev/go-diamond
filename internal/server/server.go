package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/go-diamond/internal/config"
	"github.com/example/go-diamond/internal/handler"
	"github.com/example/go-diamond/internal/middleware"
	"github.com/example/go-diamond/internal/notifier"
	"github.com/example/go-diamond/internal/service"
	"github.com/example/go-diamond/internal/store"
	"github.com/example/go-diamond/internal/watcher"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Server struct {
	cfg         *config.Config
	router      *gin.Engine
	httpServer  *http.Server
	logger      *zap.Logger
	configStore *store.ConfigStore
	hub         *watcher.WatcherHub
	notifier    *notifier.DBNotifier
}

func New(cfg *config.Config, logger *zap.Logger) (*Server, error) {
	db, err := sqlx.Connect("mysql", cfg.Database.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)

	configStore := store.NewConfigStore(db)
	historyStore := store.NewHistoryStore(db)
	configSvc := service.NewConfigService(configStore, historyStore)
	hub := watcher.NewWatcherHub()

	interval := time.Duration(cfg.Notifier.PollInterval) * time.Second
	dbNotifier := notifier.NewDBNotifier(configStore, hub, interval)

	configHandler := handler.NewConfigHandler(configSvc)
	watchHandler := handler.NewWatchHandler(configSvc, hub)
	healthHandler := handler.NewHealthHandler()

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.LoggerMiddleware(logger))
	router.Use(middleware.CorsMiddleware())

	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/api/v1/configs/:namespace/:group/:dataId", configHandler.GetConfig)
	router.GET("/api/v1/watch/:namespace/:group/:dataId", watchHandler.Watch)
	router.POST("/api/v1/watch/batch", watchHandler.BatchWatch)

	api := router.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(cfg.Server.AdminToken))
	{
		api.POST("/configs", configHandler.CreateConfig)
		api.PUT("/configs/:namespace/:group/:dataId", configHandler.UpdateConfig)
		api.DELETE("/configs/:namespace/:group/:dataId", configHandler.DeleteConfig)
		api.GET("/configs", configHandler.ListConfigs)
		api.GET("/configs/:namespace/:group/:dataId/histories", configHandler.GetHistories)
		api.POST("/configs/:namespace/:group/:dataId/rollback", configHandler.Rollback)
	}

	return &Server{
		cfg:         cfg,
		router:      router,
		logger:      logger,
		configStore: configStore,
		hub:         hub,
		notifier:    dbNotifier,
	}, nil
}

func (s *Server) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.notifier.Start(ctx)

	s.httpServer = &http.Server{
		Addr:         s.cfg.Server.Addr(),
		Handler:      s.router,
		ReadTimeout:  time.Duration(s.cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.cfg.Server.WriteTimeout) * time.Second,
	}

	go func() {
		s.logger.Info("server starting", zap.String("addr", s.cfg.Server.Addr()))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("server shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	s.notifier.Stop()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return err
	}

	s.logger.Info("server stopped")
	return nil
}

func (s *Server) Addr() string {
	return s.cfg.Server.Addr()
}