package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/keeth/levity/config"
	"github.com/keeth/levity/core"
	"github.com/keeth/levity/db"
	"github.com/keeth/levity/monitoring"
	"github.com/keeth/levity/server"
)

func main() {
	// Parse command line flags
	var (
		migrateUp     = flag.Bool("migrate-up", false, "Run database migrations up")
		migrateDown   = flag.Int("migrate-down", 0, "Roll back N migration steps")
		migrateForce  = flag.Int("migrate-force", -1, "Force migration to specific version")
		migrateStatus = flag.Bool("migrate-status", false, "Show current migration status")
	)
	flag.Parse()

	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logging
	logger := initLogger(cfg)

	// Handle migration commands
	if *migrateUp || *migrateDown > 0 || *migrateForce >= 0 || *migrateStatus {
		handleMigrationCommands(cfg, logger, *migrateUp, *migrateDown, *migrateForce, *migrateStatus)
		return
	}

	logger.Info("Starting OCPP Central System...")

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

// handleMigrationCommands handles database migration CLI commands
func handleMigrationCommands(cfg *config.Config, logger *slog.Logger, up bool, down int, force int, status bool) {
	// Initialize database connection
	database, err := db.NewDatabase(cfg.Database, logger)
	if err != nil {
		logger.Error("Failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer database.Close()

	// Handle migration status command
	if status {
		version, dirty, err := database.GetMigrationVersion()
		if err != nil {
			logger.Error("Failed to get migration version", slog.Any("error", err))
			os.Exit(1)
		}

		if version == 0 {
			fmt.Println("No migrations have been applied")
		} else {
			dirtyStr := ""
			if dirty {
				dirtyStr = " (DIRTY - needs manual intervention)"
			}
			fmt.Printf("Current migration version: %d%s\n", version, dirtyStr)
		}
		return
	}

	// Handle migration force command
	if force >= 0 {
		if err := database.ForceMigrationVersion(force); err != nil {
			logger.Error("Failed to force migration version",
				slog.Int("version", force),
				slog.Any("error", err))
			os.Exit(1)
		}
		fmt.Printf("Successfully forced migration version to: %d\n", force)
		return
	}

	// Handle migration down command
	if down > 0 {
		if err := database.MigrateDown(down); err != nil {
			logger.Error("Failed to rollback migrations",
				slog.Int("steps", down),
				slog.Any("error", err))
			os.Exit(1)
		}
		fmt.Printf("Successfully rolled back %d migration steps\n", down)
		return
	}

	// Handle migration up command
	if up {
		err := database.RunMigrationsWithCallback(func(result db.MigrationResult) {
			if result.Error != nil {
				logger.Error("Migration failed",
					slog.Uint64("from_version", uint64(result.FromVersion)),
					slog.Uint64("to_version", uint64(result.ToVersion)),
					slog.Any("error", result.Error))
			} else if result.Applied {
				fmt.Printf("Migration applied: %d -> %d\n", result.FromVersion, result.ToVersion)
			} else {
				fmt.Println("No new migrations to apply")
			}
		})

		if err != nil {
			logger.Error("Failed to run migrations", slog.Any("error", err))
			os.Exit(1)
		}

		fmt.Println("Migration completed successfully")
		return
	}
}
