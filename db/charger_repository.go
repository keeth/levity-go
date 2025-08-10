package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// chargerRepository implements ChargerRepository
type chargerRepository struct {
	db     Executor
	logger Logger
}

// Logger interface for dependency injection
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// NewChargerRepository creates a new charger repository
func NewChargerRepository(db Executor, logger Logger) ChargerRepository {
	return &chargerRepository{
		db:     db,
		logger: logger,
	}
}

// Create implements ChargerRepository.Create
func (r *chargerRepository) Create(ctx context.Context, req CreateChargerRequest) (*Charger, error) {
	query := `
		INSERT INTO chargers (
			id, name, vendor, model, serial_number, firmware_version, 
			iccid, imsi, status, is_connected, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'Unknown', 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, name, vendor, model, serial_number, firmware_version, 
				  iccid, imsi, status, is_connected, 
				  last_heartbeat_at, last_boot_at, last_connect_at, 
				  last_tx_start_at, last_tx_stop_at, created_at, updated_at`

	var charger Charger
	err := r.db.QueryRowContext(ctx, query,
		req.ID, req.Name, req.Vendor, req.Model, req.SerialNumber,
		req.FirmwareVersion, req.ICCID, req.IMSI,
	).Scan(
		&charger.ID, &charger.Name, &charger.Vendor, &charger.Model,
		&charger.SerialNumber, &charger.FirmwareVersion, &charger.ICCID,
		&charger.IMSI, &charger.Status, &charger.IsConnected,
		&charger.LastHeartbeatAt, &charger.LastBootAt, &charger.LastConnectAt,
		&charger.LastTxStartAt, &charger.LastTxStopAt, &charger.CreatedAt, &charger.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create charger", "charger_id", req.ID, "error", err)
		return nil, fmt.Errorf("failed to create charger: %w", err)
	}

	r.logger.Info("Created charger", "charger_id", charger.ID, "name", charger.Name)
	return &charger, nil
}

// GetByID implements ChargerRepository.GetByID
func (r *chargerRepository) GetByID(ctx context.Context, id string) (*Charger, error) {
	query := `
		SELECT id, name, vendor, model, serial_number, firmware_version, 
			   iccid, imsi, status, is_connected, 
			   last_heartbeat_at, last_boot_at, last_connect_at, 
			   last_tx_start_at, last_tx_stop_at, created_at, updated_at
		FROM chargers WHERE id = ?`

	var charger Charger
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&charger.ID, &charger.Name, &charger.Vendor, &charger.Model,
		&charger.SerialNumber, &charger.FirmwareVersion, &charger.ICCID,
		&charger.IMSI, &charger.Status, &charger.IsConnected,
		&charger.LastHeartbeatAt, &charger.LastBootAt, &charger.LastConnectAt,
		&charger.LastTxStartAt, &charger.LastTxStopAt, &charger.CreatedAt, &charger.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("charger not found: %s", id)
		}
		r.logger.Error("Failed to get charger", "charger_id", id, "error", err)
		return nil, fmt.Errorf("failed to get charger: %w", err)
	}

	return &charger, nil
}

// Update implements ChargerRepository.Update
func (r *chargerRepository) Update(ctx context.Context, id string, req UpdateChargerRequest) (*Charger, error) {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}

	if req.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *req.Name)
	}
	if req.Vendor != nil {
		setParts = append(setParts, "vendor = ?")
		args = append(args, *req.Vendor)
	}
	if req.Model != nil {
		setParts = append(setParts, "model = ?")
		args = append(args, *req.Model)
	}
	if req.SerialNumber != nil {
		setParts = append(setParts, "serial_number = ?")
		args = append(args, *req.SerialNumber)
	}
	if req.FirmwareVersion != nil {
		setParts = append(setParts, "firmware_version = ?")
		args = append(args, *req.FirmwareVersion)
	}
	if req.ICCID != nil {
		setParts = append(setParts, "iccid = ?")
		args = append(args, *req.ICCID)
	}
	if req.IMSI != nil {
		setParts = append(setParts, "imsi = ?")
		args = append(args, *req.IMSI)
	}
	if req.Status != nil {
		setParts = append(setParts, "status = ?")
		args = append(args, *req.Status)
	}
	if req.IsConnected != nil {
		connected := 0
		if *req.IsConnected {
			connected = 1
		}
		setParts = append(setParts, "is_connected = ?")
		args = append(args, connected)
	}

	if len(setParts) == 0 {
		return r.GetByID(ctx, id) // No updates, return current state
	}

	// Always update the updated_at timestamp
	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, id) // for WHERE clause

	query := fmt.Sprintf(`
		UPDATE chargers SET %s WHERE id = ?
		RETURNING id, name, vendor, model, serial_number, firmware_version, 
				  iccid, imsi, status, is_connected, 
				  last_heartbeat_at, last_boot_at, last_connect_at, 
				  last_tx_start_at, last_tx_stop_at, created_at, updated_at`,
		strings.Join(setParts, ", "))

	var charger Charger
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&charger.ID, &charger.Name, &charger.Vendor, &charger.Model,
		&charger.SerialNumber, &charger.FirmwareVersion, &charger.ICCID,
		&charger.IMSI, &charger.Status, &charger.IsConnected,
		&charger.LastHeartbeatAt, &charger.LastBootAt, &charger.LastConnectAt,
		&charger.LastTxStartAt, &charger.LastTxStopAt, &charger.CreatedAt, &charger.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("charger not found: %s", id)
		}
		r.logger.Error("Failed to update charger", "charger_id", id, "error", err)
		return nil, fmt.Errorf("failed to update charger: %w", err)
	}

	r.logger.Info("Updated charger", "charger_id", charger.ID)
	return &charger, nil
}

// Delete implements ChargerRepository.Delete
func (r *chargerRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM chargers WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete charger", "charger_id", id, "error", err)
		return fmt.Errorf("failed to delete charger: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("charger not found: %s", id)
	}

	r.logger.Info("Deleted charger", "charger_id", id)
	return nil
}

// List implements ChargerRepository.List
func (r *chargerRepository) List(ctx context.Context, opts ListOptions) ([]*Charger, error) {
	opts.ValidateSortDirection()

	// Validate order by field for security
	validOrderFields := map[string]bool{
		"id": true, "name": true, "vendor": true, "model": true, "status": true,
		"is_connected": true, "created_at": true, "updated_at": true,
		"last_heartbeat_at": true, "last_boot_at": true, "last_connect_at": true,
	}

	if !validOrderFields[opts.OrderBy] {
		opts.OrderBy = "created_at"
	}

	query := fmt.Sprintf(`
		SELECT id, name, vendor, model, serial_number, firmware_version, 
			   iccid, imsi, status, is_connected, 
			   last_heartbeat_at, last_boot_at, last_connect_at, 
			   last_tx_start_at, last_tx_stop_at, created_at, updated_at
		FROM chargers 
		ORDER BY %s %s 
		LIMIT ? OFFSET ?`, opts.OrderBy, opts.SortDir)

	rows, err := r.db.QueryContext(ctx, query, opts.Limit, opts.Offset)
	if err != nil {
		r.logger.Error("Failed to list chargers", "error", err)
		return nil, fmt.Errorf("failed to list chargers: %w", err)
	}
	defer rows.Close()

	var chargers []*Charger
	for rows.Next() {
		var charger Charger
		err := rows.Scan(
			&charger.ID, &charger.Name, &charger.Vendor, &charger.Model,
			&charger.SerialNumber, &charger.FirmwareVersion, &charger.ICCID,
			&charger.IMSI, &charger.Status, &charger.IsConnected,
			&charger.LastHeartbeatAt, &charger.LastBootAt, &charger.LastConnectAt,
			&charger.LastTxStartAt, &charger.LastTxStopAt, &charger.CreatedAt, &charger.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan charger row", "error", err)
			return nil, fmt.Errorf("failed to scan charger: %w", err)
		}
		chargers = append(chargers, &charger)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Row iteration error", "error", err)
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return chargers, nil
}

// Count implements ChargerRepository.Count
func (r *chargerRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM chargers`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		r.logger.Error("Failed to count chargers", "error", err)
		return 0, fmt.Errorf("failed to count chargers: %w", err)
	}

	return count, nil
}

// UpdateConnectionStatus implements ChargerRepository.UpdateConnectionStatus
func (r *chargerRepository) UpdateConnectionStatus(ctx context.Context, id string, connected bool) error {
	connectedVal := 0
	if connected {
		connectedVal = 1
	}

	query := `
		UPDATE chargers 
		SET is_connected = ?, last_connect_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, connectedVal, id)
	if err != nil {
		r.logger.Error("Failed to update connection status", "charger_id", id, "connected", connected, "error", err)
		return fmt.Errorf("failed to update connection status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("charger not found: %s", id)
	}

	r.logger.Debug("Updated connection status", "charger_id", id, "connected", connected)
	return nil
}

// UpdateStatus implements ChargerRepository.UpdateStatus
func (r *chargerRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE chargers SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		r.logger.Error("Failed to update status", "charger_id", id, "status", status, "error", err)
		return fmt.Errorf("failed to update status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("charger not found: %s", id)
	}

	r.logger.Debug("Updated status", "charger_id", id, "status", status)
	return nil
}

// UpdateLastHeartbeat implements ChargerRepository.UpdateLastHeartbeat
func (r *chargerRepository) UpdateLastHeartbeat(ctx context.Context, id string, timestamp time.Time) error {
	query := `UPDATE chargers SET last_heartbeat_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, timestamp, id)
	if err != nil {
		r.logger.Error("Failed to update last heartbeat", "charger_id", id, "error", err)
		return fmt.Errorf("failed to update last heartbeat: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("charger not found: %s", id)
	}

	return nil
}

// UpdateLastBoot implements ChargerRepository.UpdateLastBoot
func (r *chargerRepository) UpdateLastBoot(ctx context.Context, id string, timestamp time.Time) error {
	query := `UPDATE chargers SET last_boot_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, timestamp, id)
	if err != nil {
		r.logger.Error("Failed to update last boot", "charger_id", id, "error", err)
		return fmt.Errorf("failed to update last boot: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("charger not found: %s", id)
	}

	return nil
}

// UpdateLastConnect implements ChargerRepository.UpdateLastConnect
func (r *chargerRepository) UpdateLastConnect(ctx context.Context, id string, timestamp time.Time) error {
	query := `UPDATE chargers SET last_connect_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, timestamp, id)
	if err != nil {
		r.logger.Error("Failed to update last connect", "charger_id", id, "error", err)
		return fmt.Errorf("failed to update last connect: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("charger not found: %s", id)
	}

	return nil
}

// UpdateLastTxStart implements ChargerRepository.UpdateLastTxStart
func (r *chargerRepository) UpdateLastTxStart(ctx context.Context, id string, timestamp time.Time) error {
	query := `UPDATE chargers SET last_tx_start_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, timestamp, id)
	if err != nil {
		r.logger.Error("Failed to update last tx start", "charger_id", id, "error", err)
		return fmt.Errorf("failed to update last tx start: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("charger not found: %s", id)
	}

	return nil
}

// UpdateLastTxStop implements ChargerRepository.UpdateLastTxStop
func (r *chargerRepository) UpdateLastTxStop(ctx context.Context, id string, timestamp time.Time) error {
	query := `UPDATE chargers SET last_tx_stop_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, timestamp, id)
	if err != nil {
		r.logger.Error("Failed to update last tx stop", "charger_id", id, "error", err)
		return fmt.Errorf("failed to update last tx stop: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("charger not found: %s", id)
	}

	return nil
}

// GetConnected implements ChargerRepository.GetConnected
func (r *chargerRepository) GetConnected(ctx context.Context) ([]*Charger, error) {
	query := `
		SELECT id, name, vendor, model, serial_number, firmware_version, 
			   iccid, imsi, status, is_connected, 
			   last_heartbeat_at, last_boot_at, last_connect_at, 
			   last_tx_start_at, last_tx_stop_at, created_at, updated_at
		FROM chargers WHERE is_connected = 1 
		ORDER BY last_connect_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("Failed to get connected chargers", "error", err)
		return nil, fmt.Errorf("failed to get connected chargers: %w", err)
	}
	defer rows.Close()

	var chargers []*Charger
	for rows.Next() {
		var charger Charger
		err := rows.Scan(
			&charger.ID, &charger.Name, &charger.Vendor, &charger.Model,
			&charger.SerialNumber, &charger.FirmwareVersion, &charger.ICCID,
			&charger.IMSI, &charger.Status, &charger.IsConnected,
			&charger.LastHeartbeatAt, &charger.LastBootAt, &charger.LastConnectAt,
			&charger.LastTxStartAt, &charger.LastTxStopAt, &charger.CreatedAt, &charger.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan charger row", "error", err)
			return nil, fmt.Errorf("failed to scan charger: %w", err)
		}
		chargers = append(chargers, &charger)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Row iteration error", "error", err)
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return chargers, nil
}

// GetByStatus implements ChargerRepository.GetByStatus
func (r *chargerRepository) GetByStatus(ctx context.Context, status string) ([]*Charger, error) {
	query := `
		SELECT id, name, vendor, model, serial_number, firmware_version, 
			   iccid, imsi, status, is_connected, 
			   last_heartbeat_at, last_boot_at, last_connect_at, 
			   last_tx_start_at, last_tx_stop_at, created_at, updated_at
		FROM chargers WHERE status = ? 
		ORDER BY updated_at DESC`

	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		r.logger.Error("Failed to get chargers by status", "status", status, "error", err)
		return nil, fmt.Errorf("failed to get chargers by status: %w", err)
	}
	defer rows.Close()

	var chargers []*Charger
	for rows.Next() {
		var charger Charger
		err := rows.Scan(
			&charger.ID, &charger.Name, &charger.Vendor, &charger.Model,
			&charger.SerialNumber, &charger.FirmwareVersion, &charger.ICCID,
			&charger.IMSI, &charger.Status, &charger.IsConnected,
			&charger.LastHeartbeatAt, &charger.LastBootAt, &charger.LastConnectAt,
			&charger.LastTxStartAt, &charger.LastTxStopAt, &charger.CreatedAt, &charger.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan charger row", "error", err)
			return nil, fmt.Errorf("failed to scan charger: %w", err)
		}
		chargers = append(chargers, &charger)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Row iteration error", "error", err)
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return chargers, nil
}
