package db

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// transactionRepository implements TransactionRepository
type transactionRepository struct {
	db     Executor
	logger Logger
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db Executor, logger Logger) TransactionRepository {
	return &transactionRepository{
		db:     db,
		logger: logger,
	}
}

// Create implements TransactionRepository.Create
func (r *transactionRepository) Create(ctx context.Context, req CreateTransactionRequest) (*Transaction, error) {
	// Generate OCPP transaction ID if not provided
	var ocppTxID int
	var err error
	if req.TransactionID != nil {
		ocppTxID = *req.TransactionID
	} else {
		ocppTxID, err = r.GenerateTransactionID(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate transaction ID: %w", err)
		}
	}

	query := `
		INSERT INTO transactions (
			transaction_id, charger_id, connector_id, id_tag, 
			start_time, meter_start, energy_delivered, status, 
			created_at, updated_at
		) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, ?, 0, 'Active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, transaction_id, charger_id, connector_id, id_tag, 
				  start_time, stop_time, meter_start, meter_stop, 
				  energy_delivered, stop_reason, status, created_at, updated_at`

	var tx Transaction
	err = r.db.QueryRowContext(ctx, query,
		ocppTxID, req.ChargerID, req.ConnectorID, req.IDTag, req.MeterStart,
	).Scan(
		&tx.ID, &tx.TransactionID, &tx.ChargerID, &tx.ConnectorID, &tx.IDTag,
		&tx.StartTime, &tx.StopTime, &tx.MeterStart, &tx.MeterStop,
		&tx.EnergyDelivered, &tx.StopReason, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create transaction",
			"charger_id", req.ChargerID,
			"connector_id", req.ConnectorID,
			"ocpp_tx_id", ocppTxID,
			"error", err)
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	r.logger.Info("Created transaction",
		"id", tx.ID,
		"ocpp_tx_id", *tx.TransactionID,
		"charger_id", tx.ChargerID,
		"connector_id", tx.ConnectorID)
	return &tx, nil
}

// GetByID implements TransactionRepository.GetByID
func (r *transactionRepository) GetByID(ctx context.Context, id int) (*Transaction, error) {
	query := `
		SELECT id, transaction_id, charger_id, connector_id, id_tag, 
			   start_time, stop_time, meter_start, meter_stop, 
			   energy_delivered, stop_reason, status, created_at, updated_at
		FROM transactions WHERE id = ?`

	var tx Transaction
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tx.ID, &tx.TransactionID, &tx.ChargerID, &tx.ConnectorID, &tx.IDTag,
		&tx.StartTime, &tx.StopTime, &tx.MeterStart, &tx.MeterStop,
		&tx.EnergyDelivered, &tx.StopReason, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found: %d", id)
		}
		r.logger.Error("Failed to get transaction", "id", id, "error", err)
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &tx, nil
}

// GetByTransactionID implements TransactionRepository.GetByTransactionID
func (r *transactionRepository) GetByTransactionID(ctx context.Context, transactionID int) (*Transaction, error) {
	query := `
		SELECT id, transaction_id, charger_id, connector_id, id_tag, 
			   start_time, stop_time, meter_start, meter_stop, 
			   energy_delivered, stop_reason, status, created_at, updated_at
		FROM transactions WHERE transaction_id = ?`

	var tx Transaction
	err := r.db.QueryRowContext(ctx, query, transactionID).Scan(
		&tx.ID, &tx.TransactionID, &tx.ChargerID, &tx.ConnectorID, &tx.IDTag,
		&tx.StartTime, &tx.StopTime, &tx.MeterStart, &tx.MeterStop,
		&tx.EnergyDelivered, &tx.StopReason, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found with OCPP ID: %d", transactionID)
		}
		r.logger.Error("Failed to get transaction by OCPP ID", "ocpp_tx_id", transactionID, "error", err)
		return nil, fmt.Errorf("failed to get transaction by OCPP ID: %w", err)
	}

	return &tx, nil
}

// Update implements TransactionRepository.Update
func (r *transactionRepository) Update(ctx context.Context, id int, req UpdateTransactionRequest) (*Transaction, error) {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}

	if req.MeterStop != nil {
		setParts = append(setParts, "meter_stop = ?")
		args = append(args, *req.MeterStop)
	}
	if req.StopTime != nil {
		setParts = append(setParts, "stop_time = ?")
		args = append(args, *req.StopTime)
	}
	if req.EnergyDelivered != nil {
		setParts = append(setParts, "energy_delivered = ?")
		args = append(args, *req.EnergyDelivered)
	}
	if req.StopReason != nil {
		setParts = append(setParts, "stop_reason = ?")
		args = append(args, *req.StopReason)
	}
	if req.Status != nil {
		setParts = append(setParts, "status = ?")
		args = append(args, *req.Status)
	}

	if len(setParts) == 0 {
		return r.GetByID(ctx, id) // No updates, return current state
	}

	// Always update the updated_at timestamp
	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, id) // for WHERE clause

	query := fmt.Sprintf(`
		UPDATE transactions SET %s WHERE id = ?
		RETURNING id, transaction_id, charger_id, connector_id, id_tag, 
				  start_time, stop_time, meter_start, meter_stop, 
				  energy_delivered, stop_reason, status, created_at, updated_at`,
		strings.Join(setParts, ", "))

	var tx Transaction
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&tx.ID, &tx.TransactionID, &tx.ChargerID, &tx.ConnectorID, &tx.IDTag,
		&tx.StartTime, &tx.StopTime, &tx.MeterStart, &tx.MeterStop,
		&tx.EnergyDelivered, &tx.StopReason, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found: %d", id)
		}
		r.logger.Error("Failed to update transaction", "id", id, "error", err)
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	r.logger.Info("Updated transaction", "id", tx.ID)
	return &tx, nil
}

// Delete implements TransactionRepository.Delete
func (r *transactionRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM transactions WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete transaction", "id", id, "error", err)
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transaction not found: %d", id)
	}

	r.logger.Info("Deleted transaction", "id", id)
	return nil
}

// List implements TransactionRepository.List
func (r *transactionRepository) List(ctx context.Context, opts ListOptions) ([]*Transaction, error) {
	opts.ValidateSortDirection()

	// Validate order by field for security
	validOrderFields := map[string]bool{
		"id": true, "transaction_id": true, "charger_id": true, "connector_id": true,
		"id_tag": true, "start_time": true, "stop_time": true, "status": true,
		"created_at": true, "updated_at": true,
	}

	if !validOrderFields[opts.OrderBy] {
		opts.OrderBy = "created_at"
	}

	query := fmt.Sprintf(`
		SELECT id, transaction_id, charger_id, connector_id, id_tag, 
			   start_time, stop_time, meter_start, meter_stop, 
			   energy_delivered, stop_reason, status, created_at, updated_at
		FROM transactions 
		ORDER BY %s %s 
		LIMIT ? OFFSET ?`, opts.OrderBy, opts.SortDir)

	rows, err := r.db.QueryContext(ctx, query, opts.Limit, opts.Offset)
	if err != nil {
		r.logger.Error("Failed to list transactions", "error", err)
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		var tx Transaction
		err := rows.Scan(
			&tx.ID, &tx.TransactionID, &tx.ChargerID, &tx.ConnectorID, &tx.IDTag,
			&tx.StartTime, &tx.StopTime, &tx.MeterStart, &tx.MeterStop,
			&tx.EnergyDelivered, &tx.StopReason, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan transaction row", "error", err)
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, &tx)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Row iteration error", "error", err)
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return transactions, nil
}

// GetByChargerID implements TransactionRepository.GetByChargerID
func (r *transactionRepository) GetByChargerID(ctx context.Context, chargerID string, opts ListOptions) ([]*Transaction, error) {
	opts.ValidateSortDirection()

	// Validate order by field for security
	validOrderFields := map[string]bool{
		"id": true, "transaction_id": true, "connector_id": true,
		"id_tag": true, "start_time": true, "stop_time": true, "status": true,
		"created_at": true, "updated_at": true,
	}

	if !validOrderFields[opts.OrderBy] {
		opts.OrderBy = "created_at"
	}

	query := fmt.Sprintf(`
		SELECT id, transaction_id, charger_id, connector_id, id_tag, 
			   start_time, stop_time, meter_start, meter_stop, 
			   energy_delivered, stop_reason, status, created_at, updated_at
		FROM transactions 
		WHERE charger_id = ?
		ORDER BY %s %s 
		LIMIT ? OFFSET ?`, opts.OrderBy, opts.SortDir)

	rows, err := r.db.QueryContext(ctx, query, chargerID, opts.Limit, opts.Offset)
	if err != nil {
		r.logger.Error("Failed to get transactions by charger", "charger_id", chargerID, "error", err)
		return nil, fmt.Errorf("failed to get transactions by charger: %w", err)
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		var tx Transaction
		err := rows.Scan(
			&tx.ID, &tx.TransactionID, &tx.ChargerID, &tx.ConnectorID, &tx.IDTag,
			&tx.StartTime, &tx.StopTime, &tx.MeterStart, &tx.MeterStop,
			&tx.EnergyDelivered, &tx.StopReason, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan transaction row", "error", err)
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, &tx)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Row iteration error", "error", err)
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return transactions, nil
}

// GetActive implements TransactionRepository.GetActive
func (r *transactionRepository) GetActive(ctx context.Context) ([]*Transaction, error) {
	query := `
		SELECT id, transaction_id, charger_id, connector_id, id_tag, 
			   start_time, stop_time, meter_start, meter_stop, 
			   energy_delivered, stop_reason, status, created_at, updated_at
		FROM transactions 
		WHERE status = 'Active'
		ORDER BY start_time DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("Failed to get active transactions", "error", err)
		return nil, fmt.Errorf("failed to get active transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		var tx Transaction
		err := rows.Scan(
			&tx.ID, &tx.TransactionID, &tx.ChargerID, &tx.ConnectorID, &tx.IDTag,
			&tx.StartTime, &tx.StopTime, &tx.MeterStart, &tx.MeterStop,
			&tx.EnergyDelivered, &tx.StopReason, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan transaction row", "error", err)
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, &tx)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Row iteration error", "error", err)
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return transactions, nil
}

// GetActiveByConnector implements TransactionRepository.GetActiveByConnector
func (r *transactionRepository) GetActiveByConnector(ctx context.Context, chargerID string, connectorID int) (*Transaction, error) {
	query := `
		SELECT id, transaction_id, charger_id, connector_id, id_tag, 
			   start_time, stop_time, meter_start, meter_stop, 
			   energy_delivered, stop_reason, status, created_at, updated_at
		FROM transactions 
		WHERE charger_id = ? AND connector_id = ? AND status = 'Active'
		ORDER BY start_time DESC
		LIMIT 1`

	var tx Transaction
	err := r.db.QueryRowContext(ctx, query, chargerID, connectorID).Scan(
		&tx.ID, &tx.TransactionID, &tx.ChargerID, &tx.ConnectorID, &tx.IDTag,
		&tx.StartTime, &tx.StopTime, &tx.MeterStart, &tx.MeterStop,
		&tx.EnergyDelivered, &tx.StopReason, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active transaction is not an error
		}
		r.logger.Error("Failed to get active transaction by connector",
			"charger_id", chargerID, "connector_id", connectorID, "error", err)
		return nil, fmt.Errorf("failed to get active transaction by connector: %w", err)
	}

	return &tx, nil
}

// Stop implements TransactionRepository.Stop
func (r *transactionRepository) Stop(ctx context.Context, id int, meterStop int, stopTime time.Time, stopReason string) error {
	// Calculate energy delivered
	energyQuery := `SELECT meter_start FROM transactions WHERE id = ?`
	var meterStart int
	err := r.db.QueryRowContext(ctx, energyQuery, id).Scan(&meterStart)
	if err != nil {
		r.logger.Error("Failed to get meter start for transaction", "id", id, "error", err)
		return fmt.Errorf("failed to get meter start: %w", err)
	}

	energyDelivered := meterStop - meterStart
	if energyDelivered < 0 {
		energyDelivered = 0 // Prevent negative energy
	}

	query := `
		UPDATE transactions 
		SET meter_stop = ?, stop_time = ?, stop_reason = ?, 
			energy_delivered = ?, status = 'Completed', updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, meterStop, stopTime, stopReason, energyDelivered, id)
	if err != nil {
		r.logger.Error("Failed to stop transaction", "id", id, "error", err)
		return fmt.Errorf("failed to stop transaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transaction not found: %d", id)
	}

	r.logger.Info("Stopped transaction",
		"id", id,
		"meter_stop", meterStop,
		"energy_delivered", energyDelivered,
		"stop_reason", stopReason)
	return nil
}

// Count implements TransactionRepository.Count
func (r *transactionRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM transactions`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		r.logger.Error("Failed to count transactions", "error", err)
		return 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	return count, nil
}

// CountByChargerID implements TransactionRepository.CountByChargerID
func (r *transactionRepository) CountByChargerID(ctx context.Context, chargerID string) (int, error) {
	query := `SELECT COUNT(*) FROM transactions WHERE charger_id = ?`

	var count int
	err := r.db.QueryRowContext(ctx, query, chargerID).Scan(&count)
	if err != nil {
		r.logger.Error("Failed to count transactions by charger", "charger_id", chargerID, "error", err)
		return 0, fmt.Errorf("failed to count transactions by charger: %w", err)
	}

	return count, nil
}

// GenerateTransactionID implements TransactionRepository.GenerateTransactionID
func (r *transactionRepository) GenerateTransactionID(ctx context.Context) (int, error) {
	// Strategy: Use a combination of timestamp and random number for uniqueness
	// This ensures we don't have collisions even in high-frequency scenarios

	for attempts := 0; attempts < 10; attempts++ {
		// Generate a candidate ID using timestamp + random component
		now := time.Now().Unix()
		random := rand.Int31n(1000) // 0-999
		candidateID := int(now*1000 + int64(random))

		// Ensure it's positive and within reasonable bounds
		if candidateID <= 0 {
			candidateID = int(time.Now().Unix() % 2147483647) // Keep within int32 range
		}

		// Check if this ID is already used
		checkQuery := `SELECT COUNT(*) FROM transactions WHERE transaction_id = ?`
		var count int
		err := r.db.QueryRowContext(ctx, checkQuery, candidateID).Scan(&count)
		if err != nil {
			r.logger.Error("Failed to check transaction ID uniqueness", "candidate_id", candidateID, "error", err)
			continue // Try again
		}

		if count == 0 {
			r.logger.Debug("Generated transaction ID", "id", candidateID)
			return candidateID, nil
		}

		// ID collision, try again with a different random component
		r.logger.Debug("Transaction ID collision, retrying", "candidate_id", candidateID, "attempt", attempts+1)
	}

	return 0, fmt.Errorf("failed to generate unique transaction ID after 10 attempts")
}
