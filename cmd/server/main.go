package main

import (
	"flag"
	"os"

	"github.com/example/go-diamond/internal/config"
	"github.com/example/go-diamond/internal/server"
	"go.uber.org/zap"
)

var (
	configPath = flag.String("config", "deploy/config.yaml", "path to config file")
)

func main() {
	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		os.Exit(1)
	}
	defer logger.Sync()

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("failed to load config", zap.Error(err))
		os.Exit(1)
	}

	srv, err := server.New(cfg, logger)
	if err != nil {
		logger.Error("failed to create server", zap.Error(err))
		os.Exit(1)
	}

	if err := srv.Start(); err != nil {
		logger.Error("server error", zap.Error(err))
		os.Exit(1)
	}
}