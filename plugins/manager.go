package plugins

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/keeth/levity/config"
)

// Plugin represents a plugin interface
type Plugin interface {
	Name() string
	Start() error
	Stop() error
	IsRunning() bool
}

// Manager manages all plugins in the system
type Manager struct {
	config  *config.Config
	logger  *slog.Logger
	plugins map[string]Plugin
	mu      sync.RWMutex
}

// NewManager creates a new plugin manager
func NewManager(cfg *config.Config, logger *slog.Logger) (*Manager, error) {
	manager := &Manager{
		config:  cfg,
		logger:  logger,
		plugins: make(map[string]Plugin),
	}

	// Initialize built-in plugins
	if err := manager.initializeBuiltinPlugins(); err != nil {
		return nil, fmt.Errorf("failed to initialize built-in plugins: %w", err)
	}

	logger.Info("Plugin manager initialized successfully")
	return manager, nil
}

// initializeBuiltinPlugins initializes the built-in plugins
func (m *Manager) initializeBuiltinPlugins() error {
	// Auto-start transaction plugin
	autoStartPlugin := NewAutoStartTransactionPlugin(m.config, m.logger)
	if err := m.RegisterPlugin(autoStartPlugin); err != nil {
		return fmt.Errorf("failed to register auto-start transaction plugin: %w", err)
	}

	// Orphaned transaction recovery plugin
	orphanedRecoveryPlugin := NewOrphanedTransactionRecoveryPlugin(m.config, m.logger)
	if err := m.RegisterPlugin(orphanedRecoveryPlugin); err != nil {
		return fmt.Errorf("failed to register orphaned transaction recovery plugin: %w", err)
	}

	return nil
}

// RegisterPlugin registers a new plugin
func (m *Manager) RegisterPlugin(plugin Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[plugin.Name()]; exists {
		return fmt.Errorf("plugin %s already registered", plugin.Name())
	}

	m.plugins[plugin.Name()] = plugin
	m.logger.Info("Plugin registered", slog.String("name", plugin.Name()))
	return nil
}

// UnregisterPlugin unregisters a plugin
func (m *Manager) UnregisterPlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// Stop the plugin if it's running
	if plugin.IsRunning() {
		if err := plugin.Stop(); err != nil {
			m.logger.Error("Failed to stop plugin", slog.String("name", name), slog.Any("error", err))
		}
	}

	delete(m.plugins, name)
	m.logger.Info("Plugin unregistered", slog.String("name", name))
	return nil
}

// GetPlugin returns a plugin by name
func (m *Manager) GetPlugin(name string) (Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, exists := m.plugins[name]
	return plugin, exists
}

// ListPlugins returns a list of all registered plugins
func (m *Manager) ListPlugins() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.plugins))
	for name := range m.plugins {
		names = append(names, name)
	}
	return names
}

// StartAll starts all registered plugins
func (m *Manager) StartAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, plugin := range m.plugins {
		if !plugin.IsRunning() {
			if err := plugin.Start(); err != nil {
				m.logger.Error("Failed to start plugin", slog.String("name", name), slog.Any("error", err))
				return fmt.Errorf("failed to start plugin %s: %w", name, err)
			}
			m.logger.Info("Plugin started", slog.String("name", name))
		}
	}

	return nil
}

// StopAll stops all running plugins
func (m *Manager) StopAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, plugin := range m.plugins {
		if plugin.IsRunning() {
			if err := plugin.Stop(); err != nil {
				m.logger.Error("Failed to stop plugin", slog.String("name", name), slog.Any("error", err))
				return fmt.Errorf("failed to stop plugin %s: %w", name, err)
			}
			m.logger.Info("Plugin stopped", slog.String("name", name))
		}
	}

	return nil
}

// Shutdown gracefully shuts down the plugin manager
func (m *Manager) Shutdown() error {
	m.logger.Info("Shutting down plugin manager...")

	if err := m.StopAll(); err != nil {
		m.logger.Error("Failed to stop all plugins", slog.Any("error", err))
		return err
	}

	m.logger.Info("Plugin manager shutdown complete")
	return nil
}
