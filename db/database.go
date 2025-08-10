package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/keeth/levity/config"
	_ "github.com/mattn/go-sqlite3"
)

// Database represents the database connection and operations
type Database struct {
	db     *sql.DB
	config config.DatabaseConfig
	logger *slog.Logger
}

// NewDatabase creates a new database connection
func NewDatabase(cfg config.DatabaseConfig, logger *slog.Logger) (*Database, error) {
	// Open SQLite database
	db, err := sql.Open("sqlite3", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &Database{
		db:     db,
		config: cfg,
		logger: logger,
	}

	logger.Info("Database connection established", slog.String("path", cfg.Path))
	return database, nil
}

// GetDB returns the underlying sql.DB instance
func (d *Database) GetDB() *sql.DB {
	return d.db
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// RunMigrations runs database migrations
func (d *Database) RunMigrations() error {
	d.logger.Info("Running database migrations...")

	// Create sqlite3 driver instance
	driver, err := sqlite3.WithInstance(d.db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create sqlite3 driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://sql/migrations",
		"sqlite3",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	d.logger.Info("Database migrations completed successfully")
	return nil
}

// Begin starts a new transaction
func (d *Database) Begin() (*sql.Tx, error) {
	return d.db.Begin()
}

// Exec executes a query without returning rows
func (d *Database) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

// Query executes a query that returns rows
func (d *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

// QueryRow executes a query that returns a single row
func (d *Database) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.db.QueryRow(query, args...)
}

// Prepare creates a prepared statement
func (d *Database) Prepare(query string) (*sql.Stmt, error) {
	return d.db.Prepare(query)
}

// Stats returns database statistics
func (d *Database) Stats() sql.DBStats {
	return d.db.Stats()
}

// Ping verifies the database connection
func (d *Database) Ping() error {
	return d.db.Ping()
}

// SetConnMaxLifetime sets the maximum amount of time a connection may be reused
func (d *Database) SetConnMaxLifetime(duration time.Duration) {
	d.db.SetConnMaxLifetime(duration)
}

// SetMaxIdleConns sets the maximum number of connections in the idle connection pool
func (d *Database) SetMaxIdleConns(n int) {
	d.db.SetMaxIdleConns(n)
}

// SetMaxOpenConns sets the maximum number of open connections to the database
func (d *Database) SetMaxOpenConns(n int) {
	d.db.SetMaxOpenConns(n)
}
