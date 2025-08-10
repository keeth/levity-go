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

// NewDatabase creates a new database connection with SQLite optimizations
func NewDatabase(cfg config.DatabaseConfig, logger *slog.Logger) (*Database, error) {
	// Construct SQLite connection string with performance optimizations
	// Enable WAL mode, foreign keys, and other optimizations in connection string
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000&_foreign_keys=ON&_temp_store=MEMORY&_busy_timeout=30000", cfg.Path)

	// Open SQLite database
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for SQLite optimization
	// SQLite with WAL mode can handle more concurrent readers
	maxOpenConns := cfg.MaxOpenConns
	if maxOpenConns <= 0 {
		maxOpenConns = 25 // Default reasonable value
	}

	maxIdleConns := cfg.MaxIdleConns
	if maxIdleConns <= 0 {
		maxIdleConns = 5 // Keep fewer idle connections for SQLite
	}

	// Ensure max idle is not greater than max open
	if maxIdleConns > maxOpenConns {
		maxIdleConns = maxOpenConns
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &Database{
		db:     db,
		config: cfg,
		logger: logger,
	}

	// Apply additional SQLite performance pragmas
	if err := database.applyPerformanceSettings(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to apply performance settings: %w", err)
	}

	// Log connection pool settings
	logger.Info("Database connection established with optimizations",
		slog.String("path", cfg.Path),
		slog.Int("max_open_conns", maxOpenConns),
		slog.Int("max_idle_conns", maxIdleConns),
		slog.Duration("conn_max_lifetime", cfg.ConnMaxLifetime))

	return database, nil
}

// applyPerformanceSettings applies SQLite-specific performance optimizations
func (d *Database) applyPerformanceSettings() error {
	pragmas := []string{
		// WAL mode checkpoint settings for better performance
		"PRAGMA wal_autocheckpoint = 1000", // Checkpoint every 1000 pages

		// Memory and performance optimizations
		"PRAGMA mmap_size = 268435456", // 256MB memory mapped I/O
		"PRAGMA optimize",              // Analyze and optimize query planner

		// Additional safety and performance settings
		"PRAGMA trusted_schema = OFF", // Security: don't trust schema
	}

	for _, pragma := range pragmas {
		if _, err := d.db.Exec(pragma); err != nil {
			d.logger.Warn("Failed to apply pragma",
				slog.String("pragma", pragma),
				slog.Any("error", err))
			// Continue with other pragmas even if one fails
		} else {
			d.logger.Debug("Applied pragma", slog.String("pragma", pragma))
		}
	}

	// Verify WAL mode is enabled
	var journalMode string
	err := d.db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		return fmt.Errorf("failed to check journal mode: %w", err)
	}

	d.logger.Info("SQLite configuration applied",
		slog.String("journal_mode", journalMode))

	return nil
}

// GetDB returns the underlying sql.DB instance
func (d *Database) GetDB() *sql.DB {
	return d.db
}

// HealthCheck performs a comprehensive database health check
func (d *Database) HealthCheck() error {
	// Basic connectivity test
	if err := d.db.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Test a simple query
	var result int
	err := d.db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("test query failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("test query returned unexpected result: %d", result)
	}

	// Check database integrity (quick check)
	var integrityCheck string
	err = d.db.QueryRow("PRAGMA quick_check(1)").Scan(&integrityCheck)
	if err != nil {
		return fmt.Errorf("integrity check failed: %w", err)
	}

	if integrityCheck != "ok" {
		return fmt.Errorf("database integrity check failed: %s", integrityCheck)
	}

	return nil
}

// GetConnectionStats returns current connection pool statistics
func (d *Database) GetConnectionStats() map[string]interface{} {
	stats := d.db.Stats()

	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration_ms":     stats.WaitDuration.Milliseconds(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

// OptimizeDatabase runs ANALYZE and VACUUM to optimize database performance
func (d *Database) OptimizeDatabase() error {
	d.logger.Info("Starting database optimization...")

	// Run ANALYZE to update statistics for query planner
	if _, err := d.db.Exec("ANALYZE"); err != nil {
		return fmt.Errorf("failed to analyze database: %w", err)
	}

	// Run incremental VACUUM to reclaim space without blocking
	if _, err := d.db.Exec("PRAGMA incremental_vacuum"); err != nil {
		d.logger.Warn("Incremental vacuum failed, trying regular vacuum", slog.Any("error", err))

		// Fall back to regular VACUUM (this will block)
		if _, err := d.db.Exec("VACUUM"); err != nil {
			return fmt.Errorf("failed to vacuum database: %w", err)
		}
	}

	d.logger.Info("Database optimization completed")
	return nil
}

// WALCheckpoint forces a WAL checkpoint to ensure data is written to main database
func (d *Database) WALCheckpoint() error {
	var checkpointResult struct {
		busy         int
		log          int
		checkpointed int
	}

	err := d.db.QueryRow("PRAGMA wal_checkpoint(TRUNCATE)").Scan(
		&checkpointResult.busy,
		&checkpointResult.log,
		&checkpointResult.checkpointed)

	if err != nil {
		return fmt.Errorf("WAL checkpoint failed: %w", err)
	}

	d.logger.Debug("WAL checkpoint completed",
		slog.Int("busy", checkpointResult.busy),
		slog.Int("log_size", checkpointResult.log),
		slog.Int("checkpointed", checkpointResult.checkpointed))

	return nil
}

// Close closes the database connection with graceful shutdown
func (d *Database) Close() error {
	if d.db != nil {
		// Perform final WAL checkpoint before closing
		if err := d.WALCheckpoint(); err != nil {
			d.logger.Warn("Failed to checkpoint WAL before closing", slog.Any("error", err))
		}

		// Log final connection statistics
		stats := d.GetConnectionStats()
		d.logger.Info("Closing database connection",
			slog.Any("final_stats", stats))

		return d.db.Close()
	}
	return nil
}

// MigrationResult contains the result of a migration operation
type MigrationResult struct {
	FromVersion uint
	ToVersion   uint
	Applied     bool
	Error       error
}

// RunMigrations runs database migrations up to the latest version
func (d *Database) RunMigrations() error {
	return d.RunMigrationsWithCallback(nil)
}

// RunMigrationsWithCallback runs database migrations with a callback for progress
func (d *Database) RunMigrationsWithCallback(callback func(MigrationResult)) error {
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
	defer m.Close()

	// Get current version
	currentVersion, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	if dirty {
		d.logger.Warn("Database is in dirty state, attempting to fix",
			slog.Uint64("version", uint64(currentVersion)))

		// Force version to clean state
		if err := m.Force(int(currentVersion)); err != nil {
			return fmt.Errorf("failed to force clean migration state: %w", err)
		}
		d.logger.Info("Cleaned dirty migration state")
	}

	// Log current state
	if err == migrate.ErrNilVersion {
		d.logger.Info("No migrations have been applied yet")
		currentVersion = 0
	} else {
		d.logger.Info("Current migration version", slog.Uint64("version", uint64(currentVersion)))
	}

	// Run migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		// Log specific migration error details
		if callback != nil {
			callback(MigrationResult{
				FromVersion: currentVersion,
				ToVersion:   currentVersion, // same version since migration failed
				Applied:     false,
				Error:       err,
			})
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Get final version
	finalVersion, _, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get final migration version: %w", err)
	}

	if err == migrate.ErrNilVersion {
		finalVersion = 0
	}

	// Log results
	if err == migrate.ErrNoChange {
		d.logger.Info("No new migrations to apply", slog.Uint64("version", uint64(finalVersion)))
	} else {
		d.logger.Info("Database migrations completed successfully",
			slog.Uint64("from_version", uint64(currentVersion)),
			slog.Uint64("to_version", uint64(finalVersion)))
	}

	// Call callback with successful result
	if callback != nil {
		callback(MigrationResult{
			FromVersion: currentVersion,
			ToVersion:   finalVersion,
			Applied:     err != migrate.ErrNoChange,
			Error:       nil,
		})
	}

	return nil
}

// MigrateDown rolls back database migrations by N steps
func (d *Database) MigrateDown(steps int) error {
	d.logger.Info("Rolling back database migrations", slog.Int("steps", steps))

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
	defer m.Close()

	// Get current version
	currentVersion, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			d.logger.Info("No migrations to roll back")
			return nil
		}
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	if dirty {
		return fmt.Errorf("database is in dirty state, cannot rollback safely")
	}

	// Calculate target version
	targetVersion := int(currentVersion) - steps
	if targetVersion < 0 {
		targetVersion = 0
	}

	d.logger.Info("Rolling back migrations",
		slog.Uint64("from_version", uint64(currentVersion)),
		slog.Int("to_version", targetVersion))

	// Perform rollback
	if targetVersion == 0 {
		err = m.Down()
	} else {
		err = m.Migrate(uint(targetVersion))
	}

	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	d.logger.Info("Migration rollback completed successfully")
	return nil
}

// GetMigrationVersion returns the current migration version
func (d *Database) GetMigrationVersion() (uint, bool, error) {
	// Create sqlite3 driver instance
	driver, err := sqlite3.WithInstance(d.db, &sqlite3.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("failed to create sqlite3 driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://sql/migrations",
		"sqlite3",
		driver,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err == migrate.ErrNilVersion {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, dirty, nil
}

// ForceMigrationVersion forces the migration version (use with caution)
func (d *Database) ForceMigrationVersion(version int) error {
	d.logger.Warn("Forcing migration version", slog.Int("version", version))

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
	defer m.Close()

	if err := m.Force(version); err != nil {
		return fmt.Errorf("failed to force migration version: %w", err)
	}

	d.logger.Info("Migration version forced successfully", slog.Int("version", version))
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
