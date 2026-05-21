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
