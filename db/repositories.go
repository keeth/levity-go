package db

import (
	"context"
	"database/sql"
	"time"
)

// Executor interface that both sql.DB and sql.Tx implement
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// ChargerRepository defines the interface for charger data operations
type ChargerRepository interface {
	// Create a new charger
	Create(ctx context.Context, req CreateChargerRequest) (*Charger, error)

	// Get charger by ID
	GetByID(ctx context.Context, id string) (*Charger, error)

	// Update charger
	Update(ctx context.Context, id string, req UpdateChargerRequest) (*Charger, error)

	// Delete charger
	Delete(ctx context.Context, id string) error

	// List chargers with optional filtering
	List(ctx context.Context, opts ListOptions) ([]*Charger, error)

	// Count total chargers
	Count(ctx context.Context) (int, error)

	// Update connection status
	UpdateConnectionStatus(ctx context.Context, id string, connected bool) error

	// Update status
	UpdateStatus(ctx context.Context, id string, status string) error

	// Update timestamps
	UpdateLastHeartbeat(ctx context.Context, id string, timestamp time.Time) error
	UpdateLastBoot(ctx context.Context, id string, timestamp time.Time) error
	UpdateLastConnect(ctx context.Context, id string, timestamp time.Time) error
	UpdateLastTxStart(ctx context.Context, id string, timestamp time.Time) error
	UpdateLastTxStop(ctx context.Context, id string, timestamp time.Time) error

	// Get connected chargers
	GetConnected(ctx context.Context) ([]*Charger, error)

	// Get chargers by status
	GetByStatus(ctx context.Context, status string) ([]*Charger, error)
}

// ChargerConnectorRepository defines the interface for connector data operations
type ChargerConnectorRepository interface {
	// Create connector
	Create(ctx context.Context, chargerID string, connectorID int) (*ChargerConnector, error)

	// Get connector by charger and connector ID
	GetByChargerAndConnector(ctx context.Context, chargerID string, connectorID int) (*ChargerConnector, error)

	// Get all connectors for a charger
	GetByChargerID(ctx context.Context, chargerID string) ([]*ChargerConnector, error)

	// Update connector status
	UpdateStatus(ctx context.Context, chargerID string, connectorID int, status string) error

	// Update connector error
	UpdateError(ctx context.Context, chargerID string, connectorID int, errorCode, vendorErrorCode string) error

	// Clear connector error
	ClearError(ctx context.Context, chargerID string, connectorID int) error

	// Delete connectors for charger
	DeleteByChargerID(ctx context.Context, chargerID string) error
}

// TransactionRepository defines the interface for transaction data operations
type TransactionRepository interface {
	// Create a new transaction
	Create(ctx context.Context, req CreateTransactionRequest) (*Transaction, error)

	// Get transaction by ID
	GetByID(ctx context.Context, id int) (*Transaction, error)

	// Get transaction by OCPP transaction ID
	GetByTransactionID(ctx context.Context, transactionID int) (*Transaction, error)

	// Update transaction
	Update(ctx context.Context, id int, req UpdateTransactionRequest) (*Transaction, error)

	// Delete transaction
	Delete(ctx context.Context, id int) error

	// List transactions with optional filtering
	List(ctx context.Context, opts ListOptions) ([]*Transaction, error)

	// Get transactions by charger
	GetByChargerID(ctx context.Context, chargerID string, opts ListOptions) ([]*Transaction, error)

	// Get active transactions
	GetActive(ctx context.Context) ([]*Transaction, error)

	// Get active transaction for connector
	GetActiveByConnector(ctx context.Context, chargerID string, connectorID int) (*Transaction, error)

	// Stop transaction
	Stop(ctx context.Context, id int, meterStop int, stopTime time.Time, stopReason string) error

	// Count transactions
	Count(ctx context.Context) (int, error)

	// Count transactions by charger
	CountByChargerID(ctx context.Context, chargerID string) (int, error)

	// Generate next OCPP transaction ID
	GenerateTransactionID(ctx context.Context) (int, error)
}

// MeterValueRepository defines the interface for meter value data operations
type MeterValueRepository interface {
	// Create meter value record
	Create(ctx context.Context, req CreateMeterValueRequest) (*MeterValue, error)

	// Get meter value by ID
	GetByID(ctx context.Context, id int) (*MeterValue, error)

	// List meter values with filtering
	List(ctx context.Context, opts ListOptions) ([]*MeterValue, error)

	// Get meter values by transaction
	GetByTransactionID(ctx context.Context, transactionID int, opts ListOptions) ([]*MeterValue, error)

	// Get meter values by charger
	GetByChargerID(ctx context.Context, chargerID string, opts ListOptions) ([]*MeterValue, error)

	// Get meter values by time range
	GetByTimeRange(ctx context.Context, chargerID string, start, end time.Time, opts ListOptions) ([]*MeterValue, error)

	// Get latest meter value for connector
	GetLatestByConnector(ctx context.Context, chargerID string, connectorID int) (*MeterValue, error)

	// Get meter values by measurand
	GetByMeasurand(ctx context.Context, chargerID string, measurand string, opts ListOptions) ([]*MeterValue, error)

	// Delete old meter values (for cleanup)
	DeleteOlderThan(ctx context.Context, cutoff time.Time) (int, error)

	// Count meter values
	Count(ctx context.Context) (int, error)
}

// ChargerErrorRepository defines the interface for error data operations
type ChargerErrorRepository interface {
	// Create error record
	Create(ctx context.Context, req CreateChargerErrorRequest) (*ChargerError, error)

	// Get error by ID
	GetByID(ctx context.Context, id int) (*ChargerError, error)

	// List errors with filtering
	List(ctx context.Context, opts ListOptions) ([]*ChargerError, error)

	// Get errors by charger
	GetByChargerID(ctx context.Context, chargerID string, opts ListOptions) ([]*ChargerError, error)

	// Get active (unresolved) errors
	GetActive(ctx context.Context) ([]*ChargerError, error)

	// Get active errors by charger
	GetActiveByChargerID(ctx context.Context, chargerID string) ([]*ChargerError, error)

	// Resolve error
	Resolve(ctx context.Context, id int, resolvedAt time.Time) error

	// Resolve errors by error code
	ResolveByErrorCode(ctx context.Context, chargerID string, errorCode string, resolvedAt time.Time) (int, error)

	// Delete old resolved errors (for cleanup)
	DeleteOldResolved(ctx context.Context, cutoff time.Time) (int, error)

	// Count errors
	Count(ctx context.Context) (int, error)

	// Count active errors
	CountActive(ctx context.Context) (int, error)
}

// RepositoryManager aggregates all repositories
type RepositoryManager interface {
	Chargers() ChargerRepository
	Connectors() ChargerConnectorRepository
	Transactions() TransactionRepository
	MeterValues() MeterValueRepository
	Errors() ChargerErrorRepository

	// Transaction management
	BeginTx(ctx context.Context) (TxManager, error)

	// Health check
	HealthCheck(ctx context.Context) error
}

// TxManager provides transactional repository access
type TxManager interface {
	Chargers() ChargerRepository
	Connectors() ChargerConnectorRepository
	Transactions() TransactionRepository
	MeterValues() MeterValueRepository
	Errors() ChargerErrorRepository

	// Transaction control
	Commit() error
	Rollback() error
}
