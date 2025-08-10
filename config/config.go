package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	OCPP       OCPPConfig       `mapstructure:"ocpp"`
	Log        LogConfig        `mapstructure:"log"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Address        string        `mapstructure:"address"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	MaxHeaderBytes int           `mapstructure:"max_header_bytes"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Path            string        `mapstructure:"path"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// OCPPConfig holds OCPP-specific configuration
type OCPPConfig struct {
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval"`
	MaxMessageSize    int           `mapstructure:"max_message_size"`
	ConnectionTimeout time.Duration `mapstructure:"connection_timeout"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	AddSource  bool   `mapstructure:"add_source"`
	TimeFormat string `mapstructure:"time_format"`
}

// MonitoringConfig holds monitoring configuration
type MonitoringConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Address string `mapstructure:"address"`
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set defaults
	setDefaults()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Bind environment variables
	bindEnvVars()

	// Override with environment variables
	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.address", ":8080")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.max_header_bytes", 1<<20)

	// Database defaults
	viper.SetDefault("database.path", "./levity.db")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 25)
	viper.SetDefault("database.conn_max_lifetime", "5m")

	// OCPP defaults
	viper.SetDefault("ocpp.heartbeat_interval", "60s")
	viper.SetDefault("ocpp.max_message_size", 1024*1024) // 1MB
	viper.SetDefault("ocpp.connection_timeout", "30s")

	// Log defaults
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("log.add_source", true)
	viper.SetDefault("log.time_format", "2006-01-02T15:04:05.000Z07:00")

	// Monitoring defaults
	viper.SetDefault("monitoring.enabled", true)
	viper.SetDefault("monitoring.address", ":9090")
}

func bindEnvVars() {
	// Server
	viper.BindEnv("server.address", "SERVER_ADDRESS")
	viper.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	viper.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")
	viper.BindEnv("server.max_header_bytes", "SERVER_MAX_HEADER_BYTES")

	// Database
	viper.BindEnv("database.path", "DB_PATH")
	viper.BindEnv("database.max_open_conns", "DB_MAX_OPEN_CONNS")
	viper.BindEnv("database.max_idle_conns", "DB_MAX_IDLE_CONNS")
	viper.BindEnv("database.conn_max_lifetime", "DB_CONN_MAX_LIFETIME")

	// OCPP
	viper.BindEnv("ocpp.heartbeat_interval", "OCPP_HEARTBEAT_INTERVAL")
	viper.BindEnv("ocpp.max_message_size", "OCPP_MAX_MESSAGE_SIZE")
	viper.BindEnv("ocpp.connection_timeout", "OCPP_CONNECTION_TIMEOUT")

	// Log
	viper.BindEnv("log.level", "LOG_LEVEL")
	viper.BindEnv("log.format", "LOG_FORMAT")
	viper.BindEnv("log.output", "LOG_OUTPUT")
	viper.BindEnv("log.add_source", "LOG_ADD_SOURCE")
	viper.BindEnv("log.time_format", "LOG_TIME_FORMAT")

	// Monitoring
	viper.BindEnv("monitoring.enabled", "MONITORING_ENABLED")
	viper.BindEnv("monitoring.address", "MONITORING_ADDRESS")
}

func validateConfig(config *Config) error {
	// Validate log level
	validLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true, "fatal": true, "panic": true,
	}
	if !validLevels[strings.ToLower(config.Log.Level)] {
		return fmt.Errorf("invalid log level: %s", config.Log.Level)
	}

	// Validate log format
	validFormats := map[string]bool{"text": true, "json": true}
	if !validFormats[strings.ToLower(config.Log.Format)] {
		return fmt.Errorf("invalid log format: %s", config.Log.Format)
	}

	// Validate log output
	validOutputs := map[string]bool{"stdout": true, "stderr": true, "file": true}
	if !validOutputs[strings.ToLower(config.Log.Output)] {
		return fmt.Errorf("invalid log output: %s", config.Log.Output)
	}

	// Validate database path
	if config.Database.Path == "" {
		return fmt.Errorf("database path cannot be empty")
	}

	// Validate server address
	if config.Server.Address == "" {
		return fmt.Errorf("server address cannot be empty")
	}

	return nil
}

// ConfigureLogger sets up the structured logger based on the configuration
func ConfigureLogger(config *LogConfig) (*slog.Logger, error) {
	// Parse log level
	var level slog.Level
	switch strings.ToLower(config.Level) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: config.AddSource,
	}

	// Create handler based on format
	var handler slog.Handler
	switch strings.ToLower(config.Format) {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	// Create logger
	logger := slog.New(handler)

	// Set as default logger
	slog.SetDefault(logger)

	return logger, nil
}

// GetEnvString returns an environment variable value or default
func GetEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt returns an environment variable value as int or default
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetEnvDuration returns an environment variable value as duration or default
func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
