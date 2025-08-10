package core

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/keeth/levity/config"
	"github.com/keeth/levity/db"
	"github.com/keeth/levity/plugins"
)

// System represents the core OCPP Central System
type System struct {
	config    *config.Config
	logger    *slog.Logger
	db        *db.Database
	repos     db.RepositoryManager
	plugins   *plugins.Manager
	mu        sync.RWMutex
	healthyDB bool
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

	// Initialize repository manager with logger adapter
	loggerAdapter := &slogAdapter{logger: logger}
	system.repos = db.NewRepositoryManager(database, loggerAdapter)

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

	// Perform initial health check
	if err := system.healthCheck(); err != nil {
		logger.Warn("Initial health check failed", slog.Any("error", err))
		system.healthyDB = false
	} else {
		system.healthyDB = true
		logger.Info("Database health check passed")
	}

	logger.Info("Core system initialized successfully")
	return system, nil
}

// GetDatabase returns the database instance
func (s *System) GetDatabase() *db.Database {
	return s.db
}

// GetRepositories returns the repository manager
func (s *System) GetRepositories() db.RepositoryManager {
	return s.repos
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

// IsHealthy returns the current health status of the database
func (s *System) IsHealthy() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.healthyDB
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

// healthCheck performs a comprehensive health check of the system
func (s *System) healthCheck() error {
	// Check database health
	if err := s.repos.HealthCheck(context.TODO()); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Update health status
	s.mu.Lock()
	s.healthyDB = true
	s.mu.Unlock()

	return nil
}

// PerformHealthCheck executes a health check and updates status
func (s *System) PerformHealthCheck() error {
	return s.healthCheck()
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

// slogAdapter adapts slog.Logger to db.Logger interface
type slogAdapter struct {
	logger *slog.Logger
}

func (s *slogAdapter) Debug(msg string, args ...interface{}) {
	s.logger.Debug(msg, args...)
}

func (s *slogAdapter) Info(msg string, args ...interface{}) {
	s.logger.Info(msg, args...)
}

func (s *slogAdapter) Warn(msg string, args ...interface{}) {
	s.logger.Warn(msg, args...)
}

func (s *slogAdapter) Error(msg string, args ...interface{}) {
	s.logger.Error(msg, args...)
}
