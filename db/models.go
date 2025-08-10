package db

import (
	"time"
)

// Charger represents a charging point in the system
type Charger struct {
	ID              string     `json:"id" db:"id"`
	Name            string     `json:"name" db:"name"`
	Vendor          string     `json:"vendor" db:"vendor"`
	Model           string     `json:"model" db:"model"`
	SerialNumber    string     `json:"serial_number" db:"serial_number"`
	FirmwareVersion string     `json:"firmware_version" db:"firmware_version"`
	ICCID           string     `json:"iccid" db:"iccid"`
	IMSI            string     `json:"imsi" db:"imsi"`
	Status          string     `json:"status" db:"status"`
	IsConnected     bool       `json:"is_connected" db:"is_connected"`
	LastHeartbeatAt *time.Time `json:"last_heartbeat_at" db:"last_heartbeat_at"`
	LastBootAt      *time.Time `json:"last_boot_at" db:"last_boot_at"`
	LastConnectAt   *time.Time `json:"last_connect_at" db:"last_connect_at"`
	LastTxStartAt   *time.Time `json:"last_tx_start_at" db:"last_tx_start_at"`
	LastTxStopAt    *time.Time `json:"last_tx_stop_at" db:"last_tx_stop_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// ChargerConnector represents an individual connector on a charger
type ChargerConnector struct {
	ID              int       `json:"id" db:"id"`
	ChargerID       string    `json:"charger_id" db:"charger_id"`
	ConnectorID     int       `json:"connector_id" db:"connector_id"`
	Status          string    `json:"status" db:"status"`
	ErrorCode       string    `json:"error_code" db:"error_code"`
	VendorErrorCode string    `json:"vendor_error_code" db:"vendor_error_code"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// Transaction represents a charging session
type Transaction struct {
	ID              int        `json:"id" db:"id"`
	TransactionID   *int       `json:"transaction_id" db:"transaction_id"`
	ChargerID       string     `json:"charger_id" db:"charger_id"`
	ConnectorID     int        `json:"connector_id" db:"connector_id"`
	IDTag           string     `json:"id_tag" db:"id_tag"`
	StartTime       time.Time  `json:"start_time" db:"start_time"`
	StopTime        *time.Time `json:"stop_time" db:"stop_time"`
	MeterStart      int        `json:"meter_start" db:"meter_start"`
	MeterStop       *int       `json:"meter_stop" db:"meter_stop"`
	EnergyDelivered int        `json:"energy_delivered" db:"energy_delivered"`
	StopReason      string     `json:"stop_reason" db:"stop_reason"`
	Status          string     `json:"status" db:"status"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// MeterValue represents a meter reading sample
type MeterValue struct {
	ID            int       `json:"id" db:"id"`
	TransactionID *int      `json:"transaction_id" db:"transaction_id"`
	ChargerID     string    `json:"charger_id" db:"charger_id"`
	ConnectorID   int       `json:"connector_id" db:"connector_id"`
	Timestamp     time.Time `json:"timestamp" db:"timestamp"`
	Measurand     string    `json:"measurand" db:"measurand"`
	Value         float64   `json:"value" db:"value"`
	Unit          string    `json:"unit" db:"unit"`
	Context       string    `json:"context" db:"context"`
	Location      string    `json:"location" db:"location"`
	Phase         string    `json:"phase" db:"phase"`
	Format        string    `json:"format" db:"format"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// ChargerError represents an error event from a charger
type ChargerError struct {
	ID               int        `json:"id" db:"id"`
	ChargerID        string     `json:"charger_id" db:"charger_id"`
	ConnectorID      *int       `json:"connector_id" db:"connector_id"`
	ErrorCode        string     `json:"error_code" db:"error_code"`
	VendorErrorCode  string     `json:"vendor_error_code" db:"vendor_error_code"`
	ErrorDescription string     `json:"error_description" db:"error_description"`
	VendorErrorInfo  string     `json:"vendor_error_info" db:"vendor_error_info"`
	Timestamp        time.Time  `json:"timestamp" db:"timestamp"`
	ResolvedAt       *time.Time `json:"resolved_at" db:"resolved_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

// CreateChargerRequest represents the data needed to create a new charger
type CreateChargerRequest struct {
	ID              string `json:"id" validate:"required"`
	Name            string `json:"name"`
	Vendor          string `json:"vendor"`
	Model           string `json:"model"`
	SerialNumber    string `json:"serial_number"`
	FirmwareVersion string `json:"firmware_version"`
	ICCID           string `json:"iccid"`
	IMSI            string `json:"imsi"`
}

// UpdateChargerRequest represents the data that can be updated for a charger
type UpdateChargerRequest struct {
	Name            *string `json:"name,omitempty"`
	Vendor          *string `json:"vendor,omitempty"`
	Model           *string `json:"model,omitempty"`
	SerialNumber    *string `json:"serial_number,omitempty"`
	FirmwareVersion *string `json:"firmware_version,omitempty"`
	ICCID           *string `json:"iccid,omitempty"`
	IMSI            *string `json:"imsi,omitempty"`
	Status          *string `json:"status,omitempty"`
	IsConnected     *bool   `json:"is_connected,omitempty"`
}

// CreateTransactionRequest represents the data needed to create a new transaction
type CreateTransactionRequest struct {
	TransactionID *int   `json:"transaction_id,omitempty"`
	ChargerID     string `json:"charger_id" validate:"required"`
	ConnectorID   int    `json:"connector_id" validate:"required"`
	IDTag         string `json:"id_tag" validate:"required"`
	MeterStart    int    `json:"meter_start"`
}

// UpdateTransactionRequest represents the data that can be updated for a transaction
type UpdateTransactionRequest struct {
	MeterStop       *int       `json:"meter_stop,omitempty"`
	StopTime        *time.Time `json:"stop_time,omitempty"`
	EnergyDelivered *int       `json:"energy_delivered,omitempty"`
	StopReason      *string    `json:"stop_reason,omitempty"`
	Status          *string    `json:"status,omitempty"`
}

// CreateMeterValueRequest represents the data needed to create a meter value record
type CreateMeterValueRequest struct {
	TransactionID *int      `json:"transaction_id,omitempty"`
	ChargerID     string    `json:"charger_id" validate:"required"`
	ConnectorID   int       `json:"connector_id" validate:"required"`
	Timestamp     time.Time `json:"timestamp" validate:"required"`
	Measurand     string    `json:"measurand" validate:"required"`
	Value         float64   `json:"value" validate:"required"`
	Unit          string    `json:"unit"`
	Context       string    `json:"context"`
	Location      string    `json:"location"`
	Phase         string    `json:"phase"`
	Format        string    `json:"format"`
}

// CreateChargerErrorRequest represents the data needed to create an error record
type CreateChargerErrorRequest struct {
	ChargerID        string    `json:"charger_id" validate:"required"`
	ConnectorID      *int      `json:"connector_id,omitempty"`
	ErrorCode        string    `json:"error_code" validate:"required"`
	VendorErrorCode  string    `json:"vendor_error_code"`
	ErrorDescription string    `json:"error_description"`
	VendorErrorInfo  string    `json:"vendor_error_info"`
	Timestamp        time.Time `json:"timestamp"`
}

// ListOptions represents common options for list operations
type ListOptions struct {
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
	OrderBy string `json:"order_by"`
	SortDir string `json:"sort_dir"` // "ASC" or "DESC"
}

// DefaultListOptions returns sensible defaults for list operations
func DefaultListOptions() ListOptions {
	return ListOptions{
		Limit:   50,
		Offset:  0,
		OrderBy: "created_at",
		SortDir: "DESC",
	}
}

// ValidateSortDirection ensures sort direction is valid
func (opts *ListOptions) ValidateSortDirection() {
	if opts.SortDir != "ASC" && opts.SortDir != "DESC" {
		opts.SortDir = "DESC"
	}
}
