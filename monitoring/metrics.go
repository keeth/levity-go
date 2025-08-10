package monitoring

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all application metrics
type Metrics struct {
	// HTTP request metrics
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight *prometheus.GaugeVec

	// OCPP connection metrics
	ocppConnectionsTotal  *prometheus.CounterVec
	ocppConnectionsActive *prometheus.GaugeVec
	ocppMessagesTotal     *prometheus.CounterVec
	ocppMessageDuration   *prometheus.HistogramVec

	// Database metrics
	databaseConnectionsActive *prometheus.GaugeVec
	databaseQueryDuration     *prometheus.HistogramVec
	databaseQueriesTotal      *prometheus.CounterVec

	// Business metrics
	chargePointsTotal  *prometheus.GaugeVec
	transactionsTotal  *prometheus.CounterVec
	transactionsActive *prometheus.GaugeVec
}

// NewMetrics creates and registers new metrics
func NewMetrics() *Metrics {
	metrics := &Metrics{
		// HTTP metrics
		httpRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		httpRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		httpRequestsInFlight: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Current number of HTTP requests being processed",
			},
			[]string{"method", "endpoint"},
		),

		// OCPP metrics
		ocppConnectionsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ocpp_connections_total",
				Help: "Total number of OCPP connections",
			},
			[]string{"charge_point_id", "status"},
		),
		ocppConnectionsActive: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "ocpp_connections_active",
				Help: "Current number of active OCPP connections",
			},
			[]string{"charge_point_id"},
		),
		ocppMessagesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ocpp_messages_total",
				Help: "Total number of OCPP messages processed",
			},
			[]string{"charge_point_id", "message_type", "direction"},
		),
		ocppMessageDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ocpp_message_duration_seconds",
				Help:    "OCPP message processing duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"charge_point_id", "message_type"},
		),

		// Database metrics
		databaseConnectionsActive: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "database_connections_active",
				Help: "Current number of active database connections",
			},
			[]string{"database"},
		),
		databaseQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "database_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"database", "query_type"},
		),
		databaseQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "database_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"database", "query_type", "status"},
		),

		// Business metrics
		chargePointsTotal: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "charge_points_total",
				Help: "Total number of charge points",
			},
			[]string{"status"},
		),
		transactionsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "transactions_total",
				Help: "Total number of transactions",
			},
			[]string{"charge_point_id", "status"},
		),
		transactionsActive: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "transactions_active",
				Help: "Current number of active transactions",
			},
			[]string{"charge_point_id"},
		),
	}

	return metrics
}

// RecordHTTPRequest records an HTTP request metric
func (m *Metrics) RecordHTTPRequest(method, endpoint, status string) {
	m.httpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
}

// RecordHTTPRequestDuration records HTTP request duration
func (m *Metrics) RecordHTTPRequestDuration(method, endpoint string, duration float64) {
	m.httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

// SetHTTPRequestsInFlight sets the number of in-flight HTTP requests
func (m *Metrics) SetHTTPRequestsInFlight(method, endpoint string, count float64) {
	m.httpRequestsInFlight.WithLabelValues(method, endpoint).Set(count)
}

// RecordOCPPConnection records an OCPP connection metric
func (m *Metrics) RecordOCPPConnection(chargePointID, status string) {
	m.ocppConnectionsTotal.WithLabelValues(chargePointID, status).Inc()
}

// SetOCPPConnectionsActive sets the number of active OCPP connections
func (m *Metrics) SetOCPPConnectionsActive(chargePointID string, count float64) {
	m.ocppConnectionsActive.WithLabelValues(chargePointID).Set(count)
}

// RecordOCPPMessage records an OCPP message metric
func (m *Metrics) RecordOCPPMessage(chargePointID, messageType, direction string) {
	m.ocppMessagesTotal.WithLabelValues(chargePointID, messageType, direction).Inc()
}

// RecordOCPPMessageDuration records OCPP message processing duration
func (m *Metrics) RecordOCPPMessageDuration(chargePointID, messageType string, duration float64) {
	m.ocppMessageDuration.WithLabelValues(chargePointID, messageType).Observe(duration)
}

// SetDatabaseConnectionsActive sets the number of active database connections
func (m *Metrics) SetDatabaseConnectionsActive(database string, count float64) {
	m.databaseConnectionsActive.WithLabelValues(database).Set(count)
}

// RecordDatabaseQuery records a database query metric
func (m *Metrics) RecordDatabaseQuery(database, queryType, status string) {
	m.databaseQueriesTotal.WithLabelValues(database, queryType, status).Inc()
}

// RecordDatabaseQueryDuration records database query duration
func (m *Metrics) RecordDatabaseQueryDuration(database, queryType string, duration float64) {
	m.databaseQueryDuration.WithLabelValues(database, queryType).Observe(duration)
}

// SetChargePointsTotal sets the total number of charge points
func (m *Metrics) SetChargePointsTotal(status string, count float64) {
	m.chargePointsTotal.WithLabelValues(status).Set(count)
}

// RecordTransaction records a transaction metric
func (m *Metrics) RecordTransaction(chargePointID, status string) {
	m.transactionsTotal.WithLabelValues(chargePointID, status).Inc()
}

// SetTransactionsActive sets the number of active transactions
func (m *Metrics) SetTransactionsActive(chargePointID string, count float64) {
	m.transactionsActive.WithLabelValues(chargePointID).Set(count)
}

// Handler returns an HTTP handler for Prometheus metrics
func (m *Metrics) Handler(w http.ResponseWriter, r *http.Request) {
	promhttp.Handler().ServeHTTP(w, r)
}
