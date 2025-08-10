package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/keeth/levity/config"
	"github.com/keeth/levity/core"
	"github.com/keeth/levity/monitoring"
)

// Server represents the HTTP server for the MCPP Central System
type Server struct {
	config     *config.Config
	coreSystem *core.System
	metrics    *monitoring.Metrics
	logger     *slog.Logger
	httpServer *http.Server
	router     *gin.Engine
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config, coreSystem *core.System, metrics *monitoring.Metrics, logger *slog.Logger) *Server {
	// Set Gin mode based on environment
	if cfg.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(loggingMiddleware(logger))
	router.Use(corsMiddleware())

	server := &Server{
		config:     cfg,
		coreSystem: coreSystem,
		metrics:    metrics,
		logger:     logger,
		router:     router,
	}

	// Setup routes
	server.setupRoutes()

	// Create HTTP server
	server.httpServer = &http.Server{
		Addr:           cfg.Server.Address,
		Handler:        router,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	return server
}

// setupRoutes configures all the routes for the server
func (s *Server) setupRoutes() {
	// Health check endpoint
	s.router.GET("/health", s.healthCheck)

	// Metrics endpoint
	s.router.GET("/metrics", s.metricsHandler)

	// OCPP WebSocket endpoint
	s.router.GET("/ocpp/:chargePointId", s.ocppWebSocketHandler)

	// API endpoints
	api := s.router.Group("/api/v1")
	{
		api.GET("/chargepoints", s.listChargePoints)
		api.GET("/chargepoints/:id", s.getChargePoint)
		api.GET("/transactions", s.listTransactions)
		api.GET("/transactions/:id", s.getTransaction)
		api.GET("/status", s.getSystemStatus)
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server", slog.String("addr", s.config.Server.Address))
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// healthCheck handles health check requests
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	})
}

// metricsHandler handles metrics requests
func (s *Server) metricsHandler(c *gin.Context) {
	if s.metrics != nil {
		s.metrics.Handler(c.Writer, c.Request)
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Metrics not available"})
	}
}

// ocppWebSocketHandler handles OCPP WebSocket connections
func (s *Server) ocppWebSocketHandler(c *gin.Context) {
	chargePointId := c.Param("chargePointId")
	if chargePointId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Charge point ID is required"})
		return
	}

	// Upgrade to WebSocket connection
	// This will be implemented in the OCPP server package
	c.JSON(http.StatusNotImplemented, gin.H{"error": "WebSocket upgrade not yet implemented"})
}

// listChargePoints lists all charge points
func (s *Server) listChargePoints(c *gin.Context) {
	// This will be implemented when we have the database layer
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not yet implemented"})
}

// getChargePoint gets a specific charge point
func (s *Server) getChargePoint(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Charge point ID is required"})
		return
	}

	// This will be implemented when we have the database layer
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not yet implemented"})
}

// listTransactions lists all transactions
func (s *Server) listTransactions(c *gin.Context) {
	// This will be implemented when we have the database layer
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not yet implemented"})
}

// getTransaction gets a specific transaction
func (s *Server) getTransaction(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction ID is required"})
		return
	}

	// This will be implemented when we have the database layer
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not yet implemented"})
}

// getSystemStatus gets the overall system status
func (s *Server) getSystemStatus(c *gin.Context) {
	// This will be implemented when we have the monitoring layer
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not yet implemented"})
}

// loggingMiddleware adds logging to all requests
func loggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.Info("HTTP Request",
			slog.String("client_ip", param.ClientIP),
			slog.String("timestamp", param.TimeStamp.Format(time.RFC3339)),
			slog.String("method", param.Method),
			slog.String("path", param.Path),
			slog.String("protocol", param.Request.Proto),
			slog.Int("status_code", param.StatusCode),
			slog.Duration("latency", param.Latency),
			slog.String("user_agent", param.Request.UserAgent()),
			slog.String("error", param.ErrorMessage),
		)
		return ""
	})
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
