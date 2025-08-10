package config

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Test loading with defaults
	config, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Verify default values
	assert.Equal(t, ":8080", config.Server.Address)
	assert.Equal(t, 30*time.Second, config.Server.ReadTimeout)
	assert.Equal(t, 30*time.Second, config.Server.WriteTimeout)
	assert.Equal(t, 1<<20, config.Server.MaxHeaderBytes)

	assert.Equal(t, "./levity.db", config.Database.Path)
	assert.Equal(t, 25, config.Database.MaxOpenConns)
	assert.Equal(t, 5, config.Database.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, config.Database.ConnMaxLifetime)

	assert.Equal(t, 60*time.Second, config.OCPP.HeartbeatInterval)
	assert.Equal(t, 1<<20, config.OCPP.MaxMessageSize)
	assert.Equal(t, 30*time.Second, config.OCPP.ConnectionTimeout)

	assert.Equal(t, "info", config.Log.Level)
	assert.Equal(t, "json", config.Log.Format)
	assert.Equal(t, "stdout", config.Log.Output)
	assert.True(t, config.Log.AddSource)
	assert.Equal(t, "2006-01-02T15:04:05.000Z07:00", config.Log.TimeFormat)

	assert.True(t, config.Monitoring.Enabled)
	assert.Equal(t, ":9090", config.Monitoring.Address)
}

func TestEnvironmentVariableOverride(t *testing.T) {
	// Set environment variables
	os.Setenv("SERVER_ADDRESS", ":9090")
	os.Setenv("DB_PATH", "/tmp/test.db")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FORMAT", "text")
	os.Setenv("LOG_ADD_SOURCE", "false")

	defer func() {
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("DB_PATH")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_FORMAT")
		os.Unsetenv("LOG_ADD_SOURCE")
	}()

	config, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Verify environment variable overrides
	assert.Equal(t, ":9090", config.Server.Address)
	assert.Equal(t, "/tmp/test.db", config.Database.Path)
	assert.Equal(t, "debug", config.Log.Level)
	assert.Equal(t, "text", config.Log.Format)
	assert.False(t, config.Log.AddSource)
}

func TestConfigureLogger(t *testing.T) {
	// Test JSON format logger
	logConfig := &LogConfig{
		Level:     "debug",
		Format:    "json",
		Output:    "stdout",
		AddSource: true,
	}

	logger, err := ConfigureLogger(logConfig)
	assert.NoError(t, err)
	assert.NotNil(t, logger)

	// Verify logger is set as default
	assert.Equal(t, logger, slog.Default())

	// Test text format logger
	logConfig.Format = "text"
	logger2, err := ConfigureLogger(logConfig)
	assert.NoError(t, err)
	assert.NotNil(t, logger2)

	// Test invalid log level (should default to info)
	logConfig.Level = "invalid"
	logger3, err := ConfigureLogger(logConfig)
	assert.NoError(t, err)
	assert.NotNil(t, logger3)
}

func TestConfigureLoggerWithDifferentLevels(t *testing.T) {
	testCases := []struct {
		level    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"invalid", slog.LevelInfo}, // should default to info
	}

	for _, tc := range testCases {
		t.Run(tc.level, func(t *testing.T) {
			logConfig := &LogConfig{
				Level:     tc.level,
				Format:    "json",
				Output:    "stdout",
				AddSource: false,
			}

			logger, err := ConfigureLogger(logConfig)
			assert.NoError(t, err)
			assert.NotNil(t, logger)
		})
	}
}

func TestGetEnvString(t *testing.T) {
	// Test with environment variable set
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	value := GetEnvString("TEST_KEY", "default")
	assert.Equal(t, "test_value", value)

	// Test with default value when environment variable not set
	value = GetEnvString("NONEXISTENT_KEY", "default_value")
	assert.Equal(t, "default_value", value)
}

func TestGetEnvInt(t *testing.T) {
	// Test with valid integer environment variable
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	value := GetEnvInt("TEST_INT", 0)
	assert.Equal(t, 42, value)

	// Test with default value when environment variable not set
	value = GetEnvInt("NONEXISTENT_INT", 100)
	assert.Equal(t, 100, value)

	// Test with invalid integer (should return default)
	os.Setenv("INVALID_INT", "not_a_number")
	value = GetEnvInt("INVALID_INT", 200)
	assert.Equal(t, 200, value)
	os.Unsetenv("INVALID_INT")
}

func TestGetEnvDuration(t *testing.T) {
	// Test with valid duration environment variable
	os.Setenv("TEST_DURATION", "30s")
	defer os.Unsetenv("TEST_DURATION")

	value := GetEnvDuration("TEST_DURATION", time.Minute)
	assert.Equal(t, 30*time.Second, value)

	// Test with default value when environment variable not set
	value = GetEnvDuration("NONEXISTENT_DURATION", 2*time.Minute)
	assert.Equal(t, 2*time.Minute, value)

	// Test with invalid duration (should return default)
	os.Setenv("INVALID_DURATION", "not_a_duration")
	value = GetEnvDuration("INVALID_DURATION", 3*time.Minute)
	assert.Equal(t, 3*time.Minute, value)
	os.Unsetenv("INVALID_DURATION")
}
