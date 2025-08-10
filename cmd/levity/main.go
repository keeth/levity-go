package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/keeth/levity/config"
	"github.com/keeth/levity/core"
	"github.com/keeth/levity/monitoring"
	"github.com/keeth/levity/server"
)

func main() {
	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logging
	logger := initLogger(cfg)
	logger.Info("Starting MCPP Central System...")

	// Initialize core components
	coreSystem, err := core.NewSystem(cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize core system", slog.Any("error", err))
		os.Exit(1)
	}

	// Initialize monitoring
	metrics := monitoring.NewMetrics()

	// Initialize server
	srv := server.NewServer(cfg, coreSystem, metrics, logger)

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server", slog.String("addr", cfg.Server.Address))
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", slog.Any("error", err))
	}

	logger.Info("Server exited")
}

func initLogger(cfg *config.Config) *slog.Logger {
	logger, _ := config.ConfigureLogger(&cfg.Log)
	return logger
}
