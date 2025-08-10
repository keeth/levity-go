package plugins

import (
	"log/slog"

	"github.com/keeth/levity/config"
)

// LoadBalancingPlugin handles load balancing between multiple charging stations
type LoadBalancingPlugin struct {
	config  *config.Config
	logger  *slog.Logger
	running bool
}

// NewLoadBalancingPlugin creates a new load balancing plugin
func NewLoadBalancingPlugin(cfg *config.Config, logger *slog.Logger) *LoadBalancingPlugin {
	return &LoadBalancingPlugin{
		config:  cfg,
		logger:  logger,
		running: false,
	}
}

// Name returns the plugin name
func (p *LoadBalancingPlugin) Name() string {
	return "load_balancing"
}

// Start starts the plugin
func (p *LoadBalancingPlugin) Start() error {
	p.logger.Info("Load balancing plugin started")
	p.running = true
	return nil
}

// Stop stops the plugin
func (p *LoadBalancingPlugin) Stop() error {
	p.logger.Info("Load balancing plugin stopped")
	p.running = false
	return nil
}

// IsRunning returns whether the plugin is currently running
func (p *LoadBalancingPlugin) IsRunning() bool {
	return p.running
}
