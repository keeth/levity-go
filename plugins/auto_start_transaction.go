package plugins

import (
	"log/slog"

	"github.com/keeth/levity/config"
)

// AutoStartTransactionPlugin automatically starts transactions based on configuration
type AutoStartTransactionPlugin struct {
	config  *config.Config
	logger  *slog.Logger
	running bool
}

// NewAutoStartTransactionPlugin creates a new auto-start transaction plugin
func NewAutoStartTransactionPlugin(cfg *config.Config, logger *slog.Logger) *AutoStartTransactionPlugin {
	return &AutoStartTransactionPlugin{
		config:  cfg,
		logger:  logger,
		running: false,
	}
}

// Name returns the plugin name
func (p *AutoStartTransactionPlugin) Name() string {
	return "auto_start_transaction"
}

// Start starts the plugin
func (p *AutoStartTransactionPlugin) Start() error {
	p.logger.Info("Auto-start transaction plugin started")
	p.running = true
	return nil
}

// Stop stops the plugin
func (p *AutoStartTransactionPlugin) Stop() error {
	p.logger.Info("Auto-start transaction plugin stopped")
	p.running = false
	return nil
}

// IsRunning returns whether the plugin is currently running
func (p *AutoStartTransactionPlugin) IsRunning() bool {
	return p.running
}
