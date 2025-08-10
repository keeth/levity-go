package db

import (
	"context"
	"database/sql"
	"fmt"
)

// repositoryManager implements RepositoryManager
type repositoryManager struct {
	db              *Database
	chargerRepo     ChargerRepository
	connectorRepo   ChargerConnectorRepository
	transactionRepo TransactionRepository
	meterValueRepo  MeterValueRepository
	errorRepo       ChargerErrorRepository
}

// txRepositoryManager implements TxManager for transactional operations
type txRepositoryManager struct {
	tx              *sql.Tx
	chargerRepo     ChargerRepository
	connectorRepo   ChargerConnectorRepository
	transactionRepo TransactionRepository
	meterValueRepo  MeterValueRepository
	errorRepo       ChargerErrorRepository
}

// NewRepositoryManager creates a new repository manager
func NewRepositoryManager(database *Database, logger Logger) RepositoryManager {
	db := database.GetDB()

	return &repositoryManager{
		db:              database,
		chargerRepo:     NewChargerRepository(db, logger),
		connectorRepo:   NewChargerConnectorRepository(db, logger),
		transactionRepo: NewTransactionRepository(db, logger),
		meterValueRepo:  NewMeterValueRepository(db, logger),
		errorRepo:       NewChargerErrorRepository(db, logger),
	}
}

// Chargers implements RepositoryManager.Chargers
func (rm *repositoryManager) Chargers() ChargerRepository {
	return rm.chargerRepo
}

// Connectors implements RepositoryManager.Connectors
func (rm *repositoryManager) Connectors() ChargerConnectorRepository {
	return rm.connectorRepo
}

// Transactions implements RepositoryManager.Transactions
func (rm *repositoryManager) Transactions() TransactionRepository {
	return rm.transactionRepo
}

// MeterValues implements RepositoryManager.MeterValues
func (rm *repositoryManager) MeterValues() MeterValueRepository {
	return rm.meterValueRepo
}

// Errors implements RepositoryManager.Errors
func (rm *repositoryManager) Errors() ChargerErrorRepository {
	return rm.errorRepo
}

// BeginTx implements RepositoryManager.BeginTx
func (rm *repositoryManager) BeginTx(ctx context.Context) (TxManager, error) {
	tx, err := rm.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Create a simple logger adapter for transaction context
	txLogger := &txLoggerAdapter{logger: rm.db.logger}

	return &txRepositoryManager{
		tx:              tx,
		chargerRepo:     NewChargerRepository(tx, txLogger),
		connectorRepo:   NewChargerConnectorRepository(tx, txLogger),
		transactionRepo: NewTransactionRepository(tx, txLogger),
		meterValueRepo:  NewMeterValueRepository(tx, txLogger),
		errorRepo:       NewChargerErrorRepository(tx, txLogger),
	}, nil
}

// HealthCheck implements RepositoryManager.HealthCheck
func (rm *repositoryManager) HealthCheck(ctx context.Context) error {
	return rm.db.HealthCheck()
}

// TxManager implementation

// Chargers implements TxManager.Chargers
func (tm *txRepositoryManager) Chargers() ChargerRepository {
	return tm.chargerRepo
}

// Connectors implements TxManager.Connectors
func (tm *txRepositoryManager) Connectors() ChargerConnectorRepository {
	return tm.connectorRepo
}

// Transactions implements TxManager.Transactions
func (tm *txRepositoryManager) Transactions() TransactionRepository {
	return tm.transactionRepo
}

// MeterValues implements TxManager.MeterValues
func (tm *txRepositoryManager) MeterValues() MeterValueRepository {
	return tm.meterValueRepo
}

// Errors implements TxManager.Errors
func (tm *txRepositoryManager) Errors() ChargerErrorRepository {
	return tm.errorRepo
}

// Commit implements TxManager.Commit
func (tm *txRepositoryManager) Commit() error {
	return tm.tx.Commit()
}

// Rollback implements TxManager.Rollback
func (tm *txRepositoryManager) Rollback() error {
	return tm.tx.Rollback()
}

// txLoggerAdapter adapts the Database logger for use in transactions
type txLoggerAdapter struct {
	logger interface{} // This would be the actual logger from Database
}

func (tla *txLoggerAdapter) Debug(msg string, args ...interface{}) {
	// Implement based on your actual logger interface
	// For now, we'll use a simple approach
	if logger, ok := tla.logger.(Logger); ok {
		logger.Debug(msg, args...)
	}
}

func (tla *txLoggerAdapter) Info(msg string, args ...interface{}) {
	if logger, ok := tla.logger.(Logger); ok {
		logger.Info(msg, args...)
	}
}

func (tla *txLoggerAdapter) Warn(msg string, args ...interface{}) {
	if logger, ok := tla.logger.(Logger); ok {
		logger.Warn(msg, args...)
	}
}

func (tla *txLoggerAdapter) Error(msg string, args ...interface{}) {
	if logger, ok := tla.logger.(Logger); ok {
		logger.Error(msg, args...)
	}
}
