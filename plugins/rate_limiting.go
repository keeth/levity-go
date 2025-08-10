package plugins

import (
	"log/slog"

	"github.com/keeth/levity/config"
)

// RateLimitingPlugin handles rate limiting for API requests
type RateLimitingPlugin struct {
	config  *config.Config
	logger  *slog.Logger
	running bool
}

// NewRateLimitingPlugin creates a new rate limiting plugin
func NewRateLimitingPlugin(cfg *config.Config, logger *slog.Logger) *RateLimitingPlugin {
	return &RateLimitingPlugin{
		config:  cfg,
		logger:  logger,
		running: false,
	}
}

// Name returns the plugin name
func (p *RateLimitingPlugin) Name() string {
	return "rate_limiting"
}

// Start starts the plugin
func (p *RateLimitingPlugin) Start() error {
	p.logger.Info("Rate limiting plugin started")
	p.running = true
	return nil
}

// Stop stops the plugin
func (p *RateLimitingPlugin) Stop() error {
	p.logger.Info("Rate limiting plugin stopped")
	p.running = false
	return nil
}

// IsRunning returns whether the plugin is currently running
func (p *RateLimitingPlugin) IsRunning() bool {
	return p.running
}
