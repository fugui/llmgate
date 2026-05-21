package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateTimeoutsAndFrontend(t *testing.T) {
	// Create a temporary config file
	tmpFile, err := os.CreateTemp("", "config-test-*.yaml")
	require.NoError(t, err)
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	initialCfg := &Config{
		Server: ServerConfig{
			Port:            8080,
			Mode:            "release",
			ReadTimeout:     60 * time.Second,
			WriteTimeout:    30 * time.Minute,
			IdleTimeout:     300 * time.Second,
			MaxHeaderBytes:  1048576,
			ShutdownTimeout: 30 * time.Second,
		},
		Database: DatabaseConfig{
			Path: "modelgate.db",
		},
		JWT: JWTConfig{
			Secret:      "test-secret-change-in-production",
			ExpireHours: 24,
		},
		Logs: LogConfig{
			DebugRawPayloads: "none",
		},
	}

	cm := NewManager(initialCfg, tmpPath)

	// Test UpdateTimeoutsAndFrontend
	newRead := 10 * time.Second
	newWrite := 15 * time.Minute
	newIdle := 180 * time.Second
	newFrontend := FrontendConfig{
		FeedbackURL:         "https://feedback.test.com",
		DevManualURL:        "https://docs.test.com",
		RegistrationEnabled: true,
	}

	err = cm.UpdateTimeoutsAndFrontend(newRead, newWrite, newIdle, newFrontend)
	require.NoError(t, err)

	// Verify in-memory state
	currentCfg := cm.GetConfig()
	assert.Equal(t, newRead, currentCfg.Server.ReadTimeout)
	assert.Equal(t, newWrite, currentCfg.Server.WriteTimeout)
	assert.Equal(t, newIdle, currentCfg.Server.IdleTimeout)
	assert.Equal(t, newFrontend, currentCfg.Frontend)

	// Verify file is saved properly
	savedCfg, err := Load(tmpPath)
	require.NoError(t, err)
	assert.Equal(t, newRead, savedCfg.Server.ReadTimeout)
	assert.Equal(t, newWrite, savedCfg.Server.WriteTimeout)
	assert.Equal(t, newIdle, savedCfg.Server.IdleTimeout)
	assert.Equal(t, newFrontend, savedCfg.Frontend)
}

func TestLoad_NotExistAutoCreate(t *testing.T) {
	// Create a unique temporary path that doesn't exist
	tmpFile, err := os.CreateTemp("", "config-missing-*.yaml")
	require.NoError(t, err)
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	os.Remove(tmpPath) // delete the file so it definitely does not exist
	defer os.Remove(tmpPath)

	// Call Load - it should automatically create and load the file!
	cfg, err := Load(tmpPath)
	require.NoError(t, err)

	// Verify values match defaults
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "release", cfg.Server.Mode)
	assert.Equal(t, "modelgate.db", cfg.Database.Path)
	assert.Equal(t, "default-secret-change-in-production", cfg.JWT.Secret)
	assert.Equal(t, 24, cfg.JWT.ExpireHours)

	// Verify the file was physically written to disk
	assert.FileExists(t, tmpPath)
}

