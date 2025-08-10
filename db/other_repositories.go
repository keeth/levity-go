package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Connector Repository Implementation

type chargerConnectorRepository struct {
	db     Executor
	logger Logger
}

func NewChargerConnectorRepository(db Executor, logger Logger) ChargerConnectorRepository {
	return &chargerConnectorRepository{db: db, logger: logger}
}

func (r *chargerConnectorRepository) Create(ctx context.Context, chargerID string, connectorID int) (*ChargerConnector, error) {
	query := `
		INSERT INTO charger_connectors (charger_id, connector_id, status, created_at, updated_at)
		VALUES (?, ?, 'Available', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, charger_id, connector_id, status, error_code, vendor_error_code, created_at, updated_at`

	var conn ChargerConnector
	err := r.db.QueryRowContext(ctx, query, chargerID, connectorID).Scan(
		&conn.ID, &conn.ChargerID, &conn.ConnectorID, &conn.Status,
		&conn.ErrorCode, &conn.VendorErrorCode, &conn.CreatedAt, &conn.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create connector: %w", err)
	}
	return &conn, nil
}

func (r *chargerConnectorRepository) GetByChargerAndConnector(ctx context.Context, chargerID string, connectorID int) (*ChargerConnector, error) {
	query := `
		SELECT id, charger_id, connector_id, status, error_code, vendor_error_code, created_at, updated_at
		FROM charger_connectors WHERE charger_id = ? AND connector_id = ?`

	var conn ChargerConnector
	err := r.db.QueryRowContext(ctx, query, chargerID, connectorID).Scan(
		&conn.ID, &conn.ChargerID, &conn.ConnectorID, &conn.Status,
		&conn.ErrorCode, &conn.VendorErrorCode, &conn.CreatedAt, &conn.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("connector not found: %s/%d", chargerID, connectorID)
		}
		return nil, fmt.Errorf("failed to get connector: %w", err)
	}
	return &conn, nil
}

func (r *chargerConnectorRepository) GetByChargerID(ctx context.Context, chargerID string) ([]*ChargerConnector, error) {
	query := `
		SELECT id, charger_id, connector_id, status, error_code, vendor_error_code, created_at, updated_at
		FROM charger_connectors WHERE charger_id = ? ORDER BY connector_id`

	rows, err := r.db.QueryContext(ctx, query, chargerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connectors: %w", err)
	}
	defer rows.Close()

	var connectors []*ChargerConnector
	for rows.Next() {
		var conn ChargerConnector
		err := rows.Scan(&conn.ID, &conn.ChargerID, &conn.ConnectorID, &conn.Status,
			&conn.ErrorCode, &conn.VendorErrorCode, &conn.CreatedAt, &conn.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan connector: %w", err)
		}
		connectors = append(connectors, &conn)
	}
	return connectors, nil
}

func (r *chargerConnectorRepository) UpdateStatus(ctx context.Context, chargerID string, connectorID int, status string) error {
	query := `UPDATE charger_connectors SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE charger_id = ? AND connector_id = ?`
	result, err := r.db.ExecContext(ctx, query, status, chargerID, connectorID)
	if err != nil {
		return fmt.Errorf("failed to update connector status: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("connector not found: %s/%d", chargerID, connectorID)
	}
	return nil
}

func (r *chargerConnectorRepository) UpdateError(ctx context.Context, chargerID string, connectorID int, errorCode, vendorErrorCode string) error {
	query := `UPDATE charger_connectors SET error_code = ?, vendor_error_code = ?, updated_at = CURRENT_TIMESTAMP WHERE charger_id = ? AND connector_id = ?`
	result, err := r.db.ExecContext(ctx, query, errorCode, vendorErrorCode, chargerID, connectorID)
	if err != nil {
		return fmt.Errorf("failed to update connector error: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("connector not found: %s/%d", chargerID, connectorID)
	}
	return nil
}

func (r *chargerConnectorRepository) ClearError(ctx context.Context, chargerID string, connectorID int) error {
	return r.UpdateError(ctx, chargerID, connectorID, "", "")
}

func (r *chargerConnectorRepository) DeleteByChargerID(ctx context.Context, chargerID string) error {
	query := `DELETE FROM charger_connectors WHERE charger_id = ?`
	_, err := r.db.ExecContext(ctx, query, chargerID)
	return err
}

// Meter Value Repository Implementation

type meterValueRepository struct {
	db     Executor
	logger Logger
}

func NewMeterValueRepository(db Executor, logger Logger) MeterValueRepository {
	return &meterValueRepository{db: db, logger: logger}
}

func (r *meterValueRepository) Create(ctx context.Context, req CreateMeterValueRequest) (*MeterValue, error) {
	query := `
		INSERT INTO meter_values (
			transaction_id, charger_id, connector_id, timestamp, measurand, value, 
			unit, context, location, phase, format, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		RETURNING id, transaction_id, charger_id, connector_id, timestamp, measurand, 
				  value, unit, context, location, phase, format, created_at`

	var mv MeterValue
	err := r.db.QueryRowContext(ctx, query,
		req.TransactionID, req.ChargerID, req.ConnectorID, req.Timestamp,
		req.Measurand, req.Value, req.Unit, req.Context, req.Location, req.Phase, req.Format,
	).Scan(
		&mv.ID, &mv.TransactionID, &mv.ChargerID, &mv.ConnectorID, &mv.Timestamp,
		&mv.Measurand, &mv.Value, &mv.Unit, &mv.Context, &mv.Location, &mv.Phase, &mv.Format, &mv.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create meter value: %w", err)
	}
	return &mv, nil
}

func (r *meterValueRepository) GetByID(ctx context.Context, id int) (*MeterValue, error) {
	query := `
		SELECT id, transaction_id, charger_id, connector_id, timestamp, measurand, 
			   value, unit, context, location, phase, format, created_at
		FROM meter_values WHERE id = ?`

	var mv MeterValue
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&mv.ID, &mv.TransactionID, &mv.ChargerID, &mv.ConnectorID, &mv.Timestamp,
		&mv.Measurand, &mv.Value, &mv.Unit, &mv.Context, &mv.Location, &mv.Phase, &mv.Format, &mv.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("meter value not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get meter value: %w", err)
	}
	return &mv, nil
}

func (r *meterValueRepository) List(ctx context.Context, opts ListOptions) ([]*MeterValue, error) {
	query := `
		SELECT id, transaction_id, charger_id, connector_id, timestamp, measurand, 
			   value, unit, context, location, phase, format, created_at
		FROM meter_values ORDER BY timestamp DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, opts.Limit, opts.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list meter values: %w", err)
	}
	defer rows.Close()

	var values []*MeterValue
	for rows.Next() {
		var mv MeterValue
		err := rows.Scan(&mv.ID, &mv.TransactionID, &mv.ChargerID, &mv.ConnectorID, &mv.Timestamp,
			&mv.Measurand, &mv.Value, &mv.Unit, &mv.Context, &mv.Location, &mv.Phase, &mv.Format, &mv.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan meter value: %w", err)
		}
		values = append(values, &mv)
	}
	return values, nil
}

func (r *meterValueRepository) GetByTransactionID(ctx context.Context, transactionID int, opts ListOptions) ([]*MeterValue, error) {
	query := `
		SELECT id, transaction_id, charger_id, connector_id, timestamp, measurand, 
			   value, unit, context, location, phase, format, created_at
		FROM meter_values WHERE transaction_id = ? ORDER BY timestamp ASC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, transactionID, opts.Limit, opts.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get meter values by transaction: %w", err)
	}
	defer rows.Close()

	var values []*MeterValue
	for rows.Next() {
		var mv MeterValue
		err := rows.Scan(&mv.ID, &mv.TransactionID, &mv.ChargerID, &mv.ConnectorID, &mv.Timestamp,
			&mv.Measurand, &mv.Value, &mv.Unit, &mv.Context, &mv.Location, &mv.Phase, &mv.Format, &mv.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan meter value: %w", err)
		}
		values = append(values, &mv)
	}
	return values, nil
}

func (r *meterValueRepository) GetByChargerID(ctx context.Context, chargerID string, opts ListOptions) ([]*MeterValue, error) {
	query := `
		SELECT id, transaction_id, charger_id, connector_id, timestamp, measurand, 
			   value, unit, context, location, phase, format, created_at
		FROM meter_values WHERE charger_id = ? ORDER BY timestamp DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, chargerID, opts.Limit, opts.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get meter values by charger: %w", err)
	}
	defer rows.Close()

	var values []*MeterValue
	for rows.Next() {
		var mv MeterValue
		err := rows.Scan(&mv.ID, &mv.TransactionID, &mv.ChargerID, &mv.ConnectorID, &mv.Timestamp,
			&mv.Measurand, &mv.Value, &mv.Unit, &mv.Context, &mv.Location, &mv.Phase, &mv.Format, &mv.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan meter value: %w", err)
		}
		values = append(values, &mv)
	}
	return values, nil
}

func (r *meterValueRepository) GetByTimeRange(ctx context.Context, chargerID string, start, end time.Time, opts ListOptions) ([]*MeterValue, error) {
	query := `
		SELECT id, transaction_id, charger_id, connector_id, timestamp, measurand, 
			   value, unit, context, location, phase, format, created_at
		FROM meter_values WHERE charger_id = ? AND timestamp BETWEEN ? AND ? 
		ORDER BY timestamp ASC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, chargerID, start, end, opts.Limit, opts.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get meter values by time range: %w", err)
	}
	defer rows.Close()

	var values []*MeterValue
	for rows.Next() {
		var mv MeterValue
		err := rows.Scan(&mv.ID, &mv.TransactionID, &mv.ChargerID, &mv.ConnectorID, &mv.Timestamp,
			&mv.Measurand, &mv.Value, &mv.Unit, &mv.Context, &mv.Location, &mv.Phase, &mv.Format, &mv.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan meter value: %w", err)
		}
		values = append(values, &mv)
	}
	return values, nil
}

func (r *meterValueRepository) GetLatestByConnector(ctx context.Context, chargerID string, connectorID int) (*MeterValue, error) {
	query := `
		SELECT id, transaction_id, charger_id, connector_id, timestamp, measurand, 
			   value, unit, context, location, phase, format, created_at
		FROM meter_values WHERE charger_id = ? AND connector_id = ? 
		ORDER BY timestamp DESC LIMIT 1`

	var mv MeterValue
	err := r.db.QueryRowContext(ctx, query, chargerID, connectorID).Scan(
		&mv.ID, &mv.TransactionID, &mv.ChargerID, &mv.ConnectorID, &mv.Timestamp,
		&mv.Measurand, &mv.Value, &mv.Unit, &mv.Context, &mv.Location, &mv.Phase, &mv.Format, &mv.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No meter values is not an error
		}
		return nil, fmt.Errorf("failed to get latest meter value: %w", err)
	}
	return &mv, nil
}

func (r *meterValueRepository) GetByMeasurand(ctx context.Context, chargerID string, measurand string, opts ListOptions) ([]*MeterValue, error) {
	query := `
		SELECT id, transaction_id, charger_id, connector_id, timestamp, measurand, 
			   value, unit, context, location, phase, format, created_at
		FROM meter_values WHERE charger_id = ? AND measurand = ? 
		ORDER BY timestamp DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, chargerID, measurand, opts.Limit, opts.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get meter values by measurand: %w", err)
	}
	defer rows.Close()

	var values []*MeterValue
	for rows.Next() {
		var mv MeterValue
		err := rows.Scan(&mv.ID, &mv.TransactionID, &mv.ChargerID, &mv.ConnectorID, &mv.Timestamp,
			&mv.Measurand, &mv.Value, &mv.Unit, &mv.Context, &mv.Location, &mv.Phase, &mv.Format, &mv.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan meter value: %w", err)
		}
		values = append(values, &mv)
	}
	return values, nil
}

func (r *meterValueRepository) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int, error) {
	query := `DELETE FROM meter_values WHERE created_at < ?`
	result, err := r.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old meter values: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}

func (r *meterValueRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM meter_values`
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count meter values: %w", err)
	}
	return count, nil
}

// Charger Error Repository Implementation

type chargerErrorRepository struct {
	db     Executor
	logger Logger
}

func NewChargerErrorRepository(db Executor, logger Logger) ChargerErrorRepository {
	return &chargerErrorRepository{db: db, logger: logger}
}

func (r *chargerErrorRepository) Create(ctx context.Context, req CreateChargerErrorRequest) (*ChargerError, error) {
	timestamp := req.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	query := `
		INSERT INTO charger_errors (
			charger_id, connector_id, error_code, vendor_error_code, 
			error_description, vendor_error_info, timestamp, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		RETURNING id, charger_id, connector_id, error_code, vendor_error_code, 
				  error_description, vendor_error_info, timestamp, resolved_at, created_at`

	var err ChargerError
	scanErr := r.db.QueryRowContext(ctx, query,
		req.ChargerID, req.ConnectorID, req.ErrorCode, req.VendorErrorCode,
		req.ErrorDescription, req.VendorErrorInfo, timestamp,
	).Scan(
		&err.ID, &err.ChargerID, &err.ConnectorID, &err.ErrorCode, &err.VendorErrorCode,
		&err.ErrorDescription, &err.VendorErrorInfo, &err.Timestamp, &err.ResolvedAt, &err.CreatedAt,
	)
	if scanErr != nil {
		return nil, fmt.Errorf("failed to create charger error: %w", scanErr)
	}
	return &err, nil
}

func (r *chargerErrorRepository) GetByID(ctx context.Context, id int) (*ChargerError, error) {
	query := `
		SELECT id, charger_id, connector_id, error_code, vendor_error_code, 
			   error_description, vendor_error_info, timestamp, resolved_at, created_at
		FROM charger_errors WHERE id = ?`

	var err ChargerError
	scanErr := r.db.QueryRowContext(ctx, query, id).Scan(
		&err.ID, &err.ChargerID, &err.ConnectorID, &err.ErrorCode, &err.VendorErrorCode,
		&err.ErrorDescription, &err.VendorErrorInfo, &err.Timestamp, &err.ResolvedAt, &err.CreatedAt,
	)
	if scanErr != nil {
		if scanErr == sql.ErrNoRows {
			return nil, fmt.Errorf("charger error not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get charger error: %w", scanErr)
	}
	return &err, nil
}

func (r *chargerErrorRepository) List(ctx context.Context, opts ListOptions) ([]*ChargerError, error) {
	query := `
		SELECT id, charger_id, connector_id, error_code, vendor_error_code, 
			   error_description, vendor_error_info, timestamp, resolved_at, created_at
		FROM charger_errors ORDER BY timestamp DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, opts.Limit, opts.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list charger errors: %w", err)
	}
	defer rows.Close()

	var errors []*ChargerError
	for rows.Next() {
		var cerr ChargerError
		err := rows.Scan(&cerr.ID, &cerr.ChargerID, &cerr.ConnectorID, &cerr.ErrorCode, &cerr.VendorErrorCode,
			&cerr.ErrorDescription, &cerr.VendorErrorInfo, &cerr.Timestamp, &cerr.ResolvedAt, &cerr.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan charger error: %w", err)
		}
		errors = append(errors, &cerr)
	}
	return errors, nil
}

func (r *chargerErrorRepository) GetByChargerID(ctx context.Context, chargerID string, opts ListOptions) ([]*ChargerError, error) {
	query := `
		SELECT id, charger_id, connector_id, error_code, vendor_error_code, 
			   error_description, vendor_error_info, timestamp, resolved_at, created_at
		FROM charger_errors WHERE charger_id = ? ORDER BY timestamp DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, chargerID, opts.Limit, opts.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get charger errors by charger ID: %w", err)
	}
	defer rows.Close()

	var errors []*ChargerError
	for rows.Next() {
		var cerr ChargerError
		err := rows.Scan(&cerr.ID, &cerr.ChargerID, &cerr.ConnectorID, &cerr.ErrorCode, &cerr.VendorErrorCode,
			&cerr.ErrorDescription, &cerr.VendorErrorInfo, &cerr.Timestamp, &cerr.ResolvedAt, &cerr.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan charger error: %w", err)
		}
		errors = append(errors, &cerr)
	}
	return errors, nil
}

func (r *chargerErrorRepository) GetActive(ctx context.Context) ([]*ChargerError, error) {
	query := `
		SELECT id, charger_id, connector_id, error_code, vendor_error_code, 
			   error_description, vendor_error_info, timestamp, resolved_at, created_at
		FROM charger_errors WHERE resolved_at IS NULL ORDER BY timestamp DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active charger errors: %w", err)
	}
	defer rows.Close()

	var errors []*ChargerError
	for rows.Next() {
		var cerr ChargerError
		err := rows.Scan(&cerr.ID, &cerr.ChargerID, &cerr.ConnectorID, &cerr.ErrorCode, &cerr.VendorErrorCode,
			&cerr.ErrorDescription, &cerr.VendorErrorInfo, &cerr.Timestamp, &cerr.ResolvedAt, &cerr.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan charger error: %w", err)
		}
		errors = append(errors, &cerr)
	}
	return errors, nil
}

func (r *chargerErrorRepository) GetActiveByChargerID(ctx context.Context, chargerID string) ([]*ChargerError, error) {
	query := `
		SELECT id, charger_id, connector_id, error_code, vendor_error_code, 
			   error_description, vendor_error_info, timestamp, resolved_at, created_at
		FROM charger_errors WHERE charger_id = ? AND resolved_at IS NULL ORDER BY timestamp DESC`

	rows, err := r.db.QueryContext(ctx, query, chargerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active charger errors by charger ID: %w", err)
	}
	defer rows.Close()

	var errors []*ChargerError
	for rows.Next() {
		var cerr ChargerError
		err := rows.Scan(&cerr.ID, &cerr.ChargerID, &cerr.ConnectorID, &cerr.ErrorCode, &cerr.VendorErrorCode,
			&cerr.ErrorDescription, &cerr.VendorErrorInfo, &cerr.Timestamp, &cerr.ResolvedAt, &cerr.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan charger error: %w", err)
		}
		errors = append(errors, &cerr)
	}
	return errors, nil
}

func (r *chargerErrorRepository) Resolve(ctx context.Context, id int, resolvedAt time.Time) error {
	query := `UPDATE charger_errors SET resolved_at = ? WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, resolvedAt, id)
	if err != nil {
		return fmt.Errorf("failed to resolve charger error: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("charger error not found: %d", id)
	}
	return nil
}

func (r *chargerErrorRepository) ResolveByErrorCode(ctx context.Context, chargerID string, errorCode string, resolvedAt time.Time) (int, error) {
	query := `UPDATE charger_errors SET resolved_at = ? WHERE charger_id = ? AND error_code = ? AND resolved_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, resolvedAt, chargerID, errorCode)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve charger errors by error code: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}

func (r *chargerErrorRepository) DeleteOldResolved(ctx context.Context, cutoff time.Time) (int, error) {
	query := `DELETE FROM charger_errors WHERE resolved_at IS NOT NULL AND resolved_at < ?`
	result, err := r.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old resolved errors: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}

func (r *chargerErrorRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM charger_errors`
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count charger errors: %w", err)
	}
	return count, nil
}

func (r *chargerErrorRepository) CountActive(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM charger_errors WHERE resolved_at IS NULL`
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count active charger errors: %w", err)
	}
	return count, nil
}
