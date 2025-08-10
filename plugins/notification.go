package plugins

import (
	"log/slog"

	"github.com/keeth/levity/config"
)

// NotificationPlugin handles notifications and alerts
type NotificationPlugin struct {
	config  *config.Config
	logger  *slog.Logger
	running bool
}

// NewNotificationPlugin creates a new notification plugin
func NewNotificationPlugin(cfg *config.Config, logger *slog.Logger) *NotificationPlugin {
	return &NotificationPlugin{
		config:  cfg,
		logger:  logger,
		running: false,
	}
}

// Name returns the plugin name
func (p *NotificationPlugin) Name() string {
	return "notification"
}

// Start starts the plugin
func (p *NotificationPlugin) Start() error {
	p.logger.Info("Notification plugin started")
	p.running = true
	return nil
}

// Stop stops the plugin
func (p *NotificationPlugin) Stop() error {
	p.logger.Info("Notification plugin stopped")
	p.running = false
	return nil
}

// IsRunning returns whether the plugin is currently running
func (p *NotificationPlugin) IsRunning() bool {
	return p.running
}
