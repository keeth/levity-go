package plugins

import (
	"log/slog"

	"github.com/keeth/levity/config"
)

// OrphanedTransactionRecoveryPlugin handles recovery of orphaned transactions
type OrphanedTransactionRecoveryPlugin struct {
	config  *config.Config
	logger  *slog.Logger
	running bool
}

// NewOrphanedTransactionRecoveryPlugin creates a new orphaned transaction recovery plugin
func NewOrphanedTransactionRecoveryPlugin(cfg *config.Config, logger *slog.Logger) *OrphanedTransactionRecoveryPlugin {
	return &OrphanedTransactionRecoveryPlugin{
		config:  cfg,
		logger:  logger,
		running: false,
	}
}

// Name returns the plugin name
func (p *OrphanedTransactionRecoveryPlugin) Name() string {
	return "orphaned_transaction_recovery"
}

// Start starts the plugin
func (p *OrphanedTransactionRecoveryPlugin) Start() error {
	p.logger.Info("Orphaned transaction recovery plugin started")
	p.running = true
	return nil
}

// Stop stops the plugin
func (p *OrphanedTransactionRecoveryPlugin) Stop() error {
	p.logger.Info("Orphaned transaction recovery plugin stopped")
	p.running = false
	return nil
}

// IsRunning returns whether the plugin is currently running
func (p *OrphanedTransactionRecoveryPlugin) IsRunning() bool {
	return p.running
}
