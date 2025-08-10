-- Drop indexes first
DROP INDEX IF EXISTS idx_charger_errors_resolved;
DROP INDEX IF EXISTS idx_charger_errors_timestamp;
DROP INDEX IF EXISTS idx_charger_errors_charger_id;

DROP INDEX IF EXISTS idx_meter_values_measurand;
DROP INDEX IF EXISTS idx_meter_values_timestamp;
DROP INDEX IF EXISTS idx_meter_values_charger_id;
DROP INDEX IF EXISTS idx_meter_values_transaction_id;

DROP INDEX IF EXISTS idx_transactions_transaction_id;
DROP INDEX IF EXISTS idx_transactions_start_time;
DROP INDEX IF EXISTS idx_transactions_status;
DROP INDEX IF EXISTS idx_transactions_id_tag;
DROP INDEX IF EXISTS idx_transactions_connector_id;
DROP INDEX IF EXISTS idx_transactions_charger_id;

DROP INDEX IF EXISTS idx_connector_status;
DROP INDEX IF EXISTS idx_connector_charger_id;

DROP INDEX IF EXISTS idx_chargers_last_heartbeat;
DROP INDEX IF EXISTS idx_chargers_connected;
DROP INDEX IF EXISTS idx_chargers_status;

-- Drop tables in reverse order of creation (respecting foreign key constraints)
DROP TABLE IF EXISTS charger_errors;
DROP TABLE IF EXISTS meter_values;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS charger_connectors;
DROP TABLE IF EXISTS chargers;

