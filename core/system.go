package core

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/keeth/levity/config"
	"github.com/keeth/levity/db"
	"github.com/keeth/levity/plugins"
)

// System represents the core MCPP Central System
type System struct {
	config  *config.Config
	logger  *slog.Logger
	db      *db.Database
	plugins *plugins.Manager
	mu      sync.RWMutex
}

// NewSystem creates and initializes a new core system
func NewSystem(cfg *config.Config, logger *slog.Logger) (*System, error) {
	system := &System{
		config: cfg,
		logger: logger,
	}

	// Initialize database
	database, err := db.NewDatabase(cfg.Database, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	system.db = database

	// Initialize plugin manager
	pluginManager, err := plugins.NewManager(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize plugin manager: %w", err)
	}
	system.plugins = pluginManager

	// Initialize database schema
	if err := system.initializeSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	logger.Info("Core system initialized successfully")
	return system, nil
}

// GetDatabase returns the database instance
func (s *System) GetDatabase() *db.Database {
	return s.db
}

// GetPluginManager returns the plugin manager
func (s *System) GetPluginManager() *plugins.Manager {
	return s.plugins
}

// GetConfig returns the configuration
func (s *System) GetConfig() *config.Config {
	return s.config
}

// GetLogger returns the logger
func (s *System) GetLogger() *slog.Logger {
	return s.logger
}

// initializeSchema ensures the database schema is up to date
func (s *System) initializeSchema() error {
	s.logger.Info("Initializing database schema...")

	// Run migrations
	if err := s.db.RunMigrations(); err != nil {
		return fmt.Errorf("failed to run database migrations: %w", err)
	}

	s.logger.Info("Database schema initialized successfully")
	return nil
}

// Shutdown gracefully shuts down the core system
func (s *System) Shutdown() error {
	s.logger.Info("Shutting down core system...")

	// Shutdown plugins
	if s.plugins != nil {
		if err := s.plugins.Shutdown(); err != nil {
			s.logger.Error("Failed to shutdown plugin manager", slog.Any("error", err))
		}
	}

	// Close database connection
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			s.logger.Error("Failed to close database", slog.Any("error", err))
		}
	}

	s.logger.Info("Core system shutdown complete")
	return nil
}
