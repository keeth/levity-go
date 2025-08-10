# MCPP Central System - Product Requirements Document

## Executive Summary

The MCPP (Minimal Central Power Point) system is a lightweight OCPP-compliant central management system for smart EV chargers. Designed for resource-constrained environments, it provides essential charging station management with automated transaction handling, comprehensive monitoring, and extensible plugin architecture.

## Problem Statement

Current OCPP central systems are often heavyweight, complex, and require significant infrastructure investment. Many deployments need only basic functionality with reliable monitoring and automated operation, particularly for single-owner charging stations where authentication and payment complexity is unnecessary.

## Product Vision

A minimalist, reliable OCPP central system that handles essential charging operations with zero manual intervention, comprehensive observability, and extensible automation capabilities while running efficiently on commodity hardware.

## Target Users

- Small-scale EV charging operators
- Fleet managers with private charging infrastructure  
- Property managers with resident/tenant charging
- Developers building OCPP-based solutions
- System integrators needing lightweight OCPP backends

## Core Functional Requirements

### 1. OCPP Protocol Support

#### 1.1 Message Handling
- **Boot Notification**: Accept and respond to charger boot sequences
- **Status Notification**: Track charger availability and status changes
- **Start/Stop Transaction**: Handle transaction lifecycle management
- **Meter Values**: Collect and store energy consumption data
- **Heartbeat**: Maintain connection health monitoring

#### 1.2 Protocol Implementation
- **OCPP-J 1.6**: Full WebSocket/JSON implementation as primary target
- **Message Queuing**: Single in-flight message per charger with proper queueing
- **Connection Management**: Handle WebSocket lifecycle, reconnections, and timeouts
- **Future Extensibility**: Architecture supporting OCPP 2.0.1 addition

### 2. Data Persistence

#### 2.1 SQLite3 Storage
- **Charger State**: Current status, configuration, and connection state
- **Transaction Records**: Complete transaction lifecycle with timestamps
- **Meter Data**: Energy consumption readings and historical data

#### 2.2 Data Schema
- Normalized schema supporting multiple OCPP versions
- Efficient indexing for monitoring queries
- Automatic schema migrations for updates

### 3. Automated Behaviors

#### 3.1 Auto-Start Transactions
- **Trigger**: Charger status change to "Preparing"
- **Action**: Send RemoteStartTransaction request
- **Configuration**: Configurable per charger or globally

#### 3.2 Orphaned Transaction Recovery
- **Detection**: New transaction start with existing open transaction
- **Recovery**: Auto-close orphaned transactions using last meter value
- **Reason Code**: Set stop_reason to "Other" for recovered transactions

#### 3.3 Plugin Architecture
- **Hook System**: Extensible automation framework
- **Event-Driven**: React to OCPP events and state changes
- **Modular**: Enable/disable behaviors per deployment
- **Custom Logic**: Support site-specific automation requirements

### 4. Monitoring and Observability

#### 4.1 Prometheus Metrics
**System Metrics:**
- Service uptime and health
- Active connections count
- Message throughput and latency
- Error rates and types
- Database performance metrics

**Charger Metrics:**
- Connection status and uptime
- Boot/reboot events
- Status change frequencies
- Transaction counts and durations
- Error conditions and fault codes

**Energy Metrics:**
- Current power draw (Watts)
- Voltage and current readings
- Energy delivered per transaction (kWh)
- Total energy delivered per charger
- Peak power utilization

#### 4.2 Structured Logging
- **JSON Format**: Machine-readable log entries
- **Correlation IDs**: Track message flows across chargers
- **Log Levels**: Configurable verbosity (DEBUG, INFO, WARN, ERROR)
- **Event Categories**: System, OCPP, Transaction, Error classifications
- **Performance Tracking**: Response times and throughput metrics
- **Message Logs**: Structured logging of all OCPP communications
- **System Events**: Errors, anomalies, and operational events

## Technical Architecture

### 5. Technology Stack

#### 5.1 Core Technologies
- **Language**: Go (Golang) for performance and resource efficiency
- **Database**: SQLite3 for zero-administration persistence
- **Protocols**: WebSocket, JSON marshaling
- **Monitoring**: Prometheus client library

#### 5.2 Module Structure
```
mcpp/
├── cmd/           # CLI utilities and main entry points
├── core/          # OCPP protocol and business logic
├── db/            # Database layer and migrations  
├── server/        # WebSocket server and connection management
├── plugins/       # Automation behavior modules
├── monitoring/    # Prometheus metrics and health checks
└── config/        # Configuration management
```

### 6. System Design

#### 6.0 Key Golang Dependencies

- [ocpp-go](https://github.com/lorenzodonini/ocpp-go): Open Charge Point Protocol implementation in Go
- [sqlc](https://github.com/sqlc-dev/sqlc): Generate type-safe code from SQL
- [sqlitestdb](https://github.com/terinjokes/sqlitestdb): quickly run tests in their own temporary, isolated, SQLite databases
- [golang-migrate](https://github.com/golang-migrate/migrate): Database migrations. CLI and Golang library.
- [go-sqlite3](https://github.com/mattn/go-sqlite3): sqlite3 driver for go using database/sql
- [prometheus](https://github.com/prometheus/client_golang): Prometheus client library
- [gin](https://github.com/gin-gonic/gin): Web framework for Go

#### 6.1 Component Architecture
- **OCPP Server**: WebSocket endpoint handling multiple charger connections
- **Protocol Engine**: Message parsing, validation, and response generation
- **State Manager**: Charger and transaction state tracking
- **Plugin System**: Event-driven automation framework
- **Metrics Collector**: Prometheus metrics aggregation
- **Database Layer**: SQLite operations with connection pooling

#### 6.2 Message Flow
1. Charger connects via WebSocket to `/ocpp/{chargerID}` endpoint (create a new charger record in the database, if it doesn't exist)
2. Boot notification establishes charger registration
3. Incoming messages processed by protocol engine
4. State updates trigger plugin hooks
5. Responses queued and sent maintaining single-flight rule
6. All interactions logged and metrics updated

#### 6.3 Database Schema

Can be found in the `sql/migrations/001_initial.up.sql` file.

Modify the schema as needed, but it should have most of the fields that are needed for the MVP.

The schema contains hints about specific events and "last X at" timestamps that should be recorded, like 
last_heartbeat_at, last_boot_at, last_connect_at, last_tx_start_at, last_tx_stop_at, etc.

Raw messages or other events should not be stored in the database, but should be logged with structured logging and monitored with Prometheus.

For now, don't generate new migrations, just modify the existing one.

#### 6.4 Runtime

Build the service around the ocpp-go library.

Here's an example of how to use the ocpp-go library for a central system:

```go
package main

import (
	"fmt"
	"time"

	"github.com/lorenzodonini/ocpp-go/ocpp1.6/logging"
	"github.com/lorenzodonini/ocpp-go/ocpp1.6/securefirmware"
	"github.com/lorenzodonini/ocpp-go/ocpp1.6/security"
	"github.com/sirupsen/logrus"

	"github.com/lorenzodonini/ocpp-go/ocpp1.6/core"
	"github.com/lorenzodonini/ocpp-go/ocpp1.6/firmware"
	"github.com/lorenzodonini/ocpp-go/ocpp1.6/types"
)

var (
	nextTransactionId = 1
)

// TransactionInfo contains info about a transaction
type TransactionInfo struct {
	id          int
	startTime   *types.DateTime
	endTime     *types.DateTime
	startMeter  int
	endMeter    int
	connectorId int
	idTag       string
}

func (ti *TransactionInfo) hasTransactionEnded() bool {
	return ti.endTime != nil && !ti.endTime.IsZero()
}

// ConnectorInfo contains status and ongoing transaction ID for a connector
type ConnectorInfo struct {
	status             core.ChargePointStatus
	currentTransaction int
}

func (ci *ConnectorInfo) hasTransactionInProgress() bool {
	return ci.currentTransaction >= 0
}

// ChargePointState contains some simple state for a connected charge point
type ChargePointState struct {
	status            core.ChargePointStatus
	diagnosticsStatus firmware.DiagnosticsStatus
	firmwareStatus    firmware.FirmwareStatus
	connectors        map[int]*ConnectorInfo // No assumptions about the # of connectors
	transactions      map[int]*TransactionInfo
	errorCode         core.ChargePointErrorCode
}

func (cps *ChargePointState) getConnector(id int) *ConnectorInfo {
	ci, ok := cps.connectors[id]
	if !ok {
		ci = &ConnectorInfo{currentTransaction: -1}
		cps.connectors[id] = ci
	}
	return ci
}

// CentralSystemHandler contains some simple state that a central system may want to keep.
// In production this will typically be replaced by database/API calls.
type CentralSystemHandler struct {
	chargePoints map[string]*ChargePointState
}

// ------------- Core profile callbacks -------------

func (handler *CentralSystemHandler) OnAuthorize(chargePointId string, request *core.AuthorizeRequest) (confirmation *core.AuthorizeConfirmation, err error) {
	logDefault(chargePointId, request.GetFeatureName()).Infof("client authorized")
	return core.NewAuthorizationConfirmation(types.NewIdTagInfo(types.AuthorizationStatusAccepted)), nil
}

func (handler *CentralSystemHandler) OnBootNotification(chargePointId string, request *core.BootNotificationRequest) (confirmation *core.BootNotificationConfirmation, err error) {
	logDefault(chargePointId, request.GetFeatureName()).Infof("boot confirmed")
	return core.NewBootNotificationConfirmation(types.NewDateTime(time.Now()), defaultHeartbeatInterval, core.RegistrationStatusAccepted), nil
}

func (handler *CentralSystemHandler) OnDataTransfer(chargePointId string, request *core.DataTransferRequest) (confirmation *core.DataTransferConfirmation, err error) {
	logDefault(chargePointId, request.GetFeatureName()).Infof("received data %v", request.Data)
	return core.NewDataTransferConfirmation(core.DataTransferStatusAccepted), nil
}

func (handler *CentralSystemHandler) OnHeartbeat(chargePointId string, request *core.HeartbeatRequest) (confirmation *core.HeartbeatConfirmation, err error) {
	logDefault(chargePointId, request.GetFeatureName()).Infof("heartbeat handled")
	return core.NewHeartbeatConfirmation(types.NewDateTime(time.Now())), nil
}

func (handler *CentralSystemHandler) OnMeterValues(chargePointId string, request *core.MeterValuesRequest) (confirmation *core.MeterValuesConfirmation, err error) {
	logDefault(chargePointId, request.GetFeatureName()).Infof("received meter values for connector %v. Meter values:\n", request.ConnectorId)
	for _, mv := range request.MeterValue {
		logDefault(chargePointId, request.GetFeatureName()).Printf("%v", mv)
	}
	return core.NewMeterValuesConfirmation(), nil
}

func (handler *CentralSystemHandler) OnStatusNotification(chargePointId string, request *core.StatusNotificationRequest) (confirmation *core.StatusNotificationConfirmation, err error) {
	info, ok := handler.chargePoints[chargePointId]
	if !ok {
		return nil, fmt.Errorf("unknown charge point %v", chargePointId)
	}
	info.errorCode = request.ErrorCode
	if request.ConnectorId > 0 {
		connectorInfo := info.getConnector(request.ConnectorId)
		connectorInfo.status = request.Status
		logDefault(chargePointId, request.GetFeatureName()).Infof("connector %v updated status to %v", request.ConnectorId, request.Status)
	} else {
		info.status = request.Status
		logDefault(chargePointId, request.GetFeatureName()).Infof("all connectors updated status to %v", request.Status)
	}
	return core.NewStatusNotificationConfirmation(), nil
}

func (handler *CentralSystemHandler) OnStartTransaction(chargePointId string, request *core.StartTransactionRequest) (confirmation *core.StartTransactionConfirmation, err error) {
	info, ok := handler.chargePoints[chargePointId]
	if !ok {
		return nil, fmt.Errorf("unknown charge point %v", chargePointId)
	}
	connector := info.getConnector(request.ConnectorId)
	if connector.currentTransaction >= 0 {
		return nil, fmt.Errorf("connector %v is currently busy with another transaction", request.ConnectorId)
	}
	transaction := &TransactionInfo{}
	transaction.idTag = request.IdTag
	transaction.connectorId = request.ConnectorId
	transaction.startMeter = request.MeterStart
	transaction.startTime = request.Timestamp
	transaction.id = nextTransactionId
	nextTransactionId += 1
	connector.currentTransaction = transaction.id
	info.transactions[transaction.id] = transaction
	// TODO: check billable clients
	logDefault(chargePointId, request.GetFeatureName()).Infof("started transaction %v for connector %v", transaction.id, transaction.connectorId)
	return core.NewStartTransactionConfirmation(types.NewIdTagInfo(types.AuthorizationStatusAccepted), transaction.id), nil
}

func (handler *CentralSystemHandler) OnStopTransaction(chargePointId string, request *core.StopTransactionRequest) (confirmation *core.StopTransactionConfirmation, err error) {
	info, ok := handler.chargePoints[chargePointId]
	if !ok {
		return nil, fmt.Errorf("unknown charge point %v", chargePointId)
	}
	transaction, ok := info.transactions[request.TransactionId]
	if ok {
		connector := info.getConnector(transaction.connectorId)
		connector.currentTransaction = -1
		transaction.endTime = request.Timestamp
		transaction.endMeter = request.MeterStop
		// TODO: bill charging period to client
	}
	logDefault(chargePointId, request.GetFeatureName()).Infof("stopped transaction %v - %v", request.TransactionId, request.Reason)
	for _, mv := range request.TransactionData {
		logDefault(chargePointId, request.GetFeatureName()).Printf("%v", mv)
	}
	return core.NewStopTransactionConfirmation(), nil
}

// ------------- Firmware management profile callbacks -------------

func (handler *CentralSystemHandler) OnDiagnosticsStatusNotification(chargePointId string, request *firmware.DiagnosticsStatusNotificationRequest) (confirmation *firmware.DiagnosticsStatusNotificationConfirmation, err error) {
	info, ok := handler.chargePoints[chargePointId]
	if !ok {
		return nil, fmt.Errorf("unknown charge point %v", chargePointId)
	}
	info.diagnosticsStatus = request.Status
	logDefault(chargePointId, request.GetFeatureName()).Infof("updated diagnostics status to %v", request.Status)
	return firmware.NewDiagnosticsStatusNotificationConfirmation(), nil
}

func (handler *CentralSystemHandler) OnFirmwareStatusNotification(chargePointId string, request *firmware.FirmwareStatusNotificationRequest) (confirmation *firmware.FirmwareStatusNotificationConfirmation, err error) {
	info, ok := handler.chargePoints[chargePointId]
	if !ok {
		return nil, fmt.Errorf("unknown charge point %v", chargePointId)
	}
	info.firmwareStatus = request.Status
	logDefault(chargePointId, request.GetFeatureName()).Infof("updated firmware status to %v", request.Status)
	return &firmware.FirmwareStatusNotificationConfirmation{}, nil
}

// No callbacks for Local Auth management, Reservation, Remote trigger or Smart Charging profile on central system

func (handler *CentralSystemHandler) OnSecurityEventNotification(chargingStationID string, request *security.SecurityEventNotificationRequest) (response *security.SecurityEventNotificationResponse, err error) {
	logDefault(chargingStationID, request.GetFeatureName()).Infof("security event notification received")
	return security.NewSecurityEventNotificationResponse(), nil
}

func (handler *CentralSystemHandler) OnSignCertificate(chargingStationID string, request *security.SignCertificateRequest) (response *security.SignCertificateResponse, err error) {
	logDefault(chargingStationID, request.GetFeatureName()).Infof("certificate signing request received")
	return security.NewSignCertificateResponse(types.GenericStatusAccepted), nil
}

func (handler *CentralSystemHandler) OnSignedFirmwareStatusNotification(chargingStationID string, request *securefirmware.SignedFirmwareStatusNotificationRequest) (response *securefirmware.SignedFirmwareStatusNotificationResponse, err error) {
	logDefault(chargingStationID, request.GetFeatureName()).Infof("signed firmware status notification received")
	return securefirmware.NewFirmwareStatusNotificationResponse(), nil
}

func (handler *CentralSystemHandler) OnLogStatusNotification(chargingStationID string, request *logging.LogStatusNotificationRequest) (response *logging.LogStatusNotificationResponse, err error) {
	logDefault(chargingStationID, request.GetFeatureName()).Infof("log status notification received")
	return logging.NewLogStatusNotificationResponse(), nil
}

// Utility functions

func logDefault(chargePointId string, feature string) *logrus.Entry {
	return log.WithFields(logrus.Fields{"client": chargePointId, "message": feature})
}

```

#### 6.5 Configuration

For now, the service should be configured through a combination of env vars and CLI flags.
