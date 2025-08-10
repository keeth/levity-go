-- Charge Point (created automatically the first time a charger connects)
CREATE TABLE cp (
	id TEXT PRIMARY KEY,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL,
	name TEXT NOT NULL,
	status TEXT,
	is_connected INTEGER NOT NULL
);

-- Charge Point Hardware (one record per charger, updated from boot notification)
CREATE TABLE cp_hw (
    cp_id TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    vendor TEXT NOT NULL,
    model TEXT NOT NULL,
    serial TEXT NOT NULL,
    firmware TEXT NOT NULL,
    iccid TEXT NOT NULL,
    imsi TEXT NOT NULL,
    FOREIGN KEY (cp_id) REFERENCES cp(id)
);

-- Charge Point Errors (many records per charger)
CREATE TABLE cp_err (
    cp_id TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
	code TEXT,
	vendor_code TEXT,
	vendor_status_info TEXT,
	vendor_status_id TEXT,
    FOREIGN KEY (cp_id) REFERENCES cp(id)
);

-- Charge Point Status (one record per charger)
CREATE TABLE cp_stat (
    cp_id TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
	last_heartbeat_at DATETIME,
	last_boot_at DATETIME,
	last_connect_at DATETIME,
	last_tx_start_at DATETIME,
	last_tx_stop_at DATETIME,
    FOREIGN KEY (cp_id) REFERENCES cp(id)
);

-- Transaction (many records per charger)
CREATE TABLE tx (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    cp_id TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    started_at DATETIME,
    stopped_at DATETIME,
    connector_id TEXT NOT NULL,
    meter_start TEXT,
    meter_stop TEXT,
    stop_reason TEXT,
    FOREIGN KEY (cp_id) REFERENCES cp(id)
);

-- Meter Values (many records per transaction)
CREATE TABLE mtr (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tx_id TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
	context TEXT NOT NULL,
	format TEXT NOT NULL,
	location TEXT NOT NULL,
	measurand TEXT NOT NULL,
	phase TEXT NOT NULL,
	unit TEXT NOT NULL,
	value REAL NOT NULL,
    FOREIGN KEY (tx_id) REFERENCES tx(id)
);
