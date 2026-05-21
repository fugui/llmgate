package user

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"modelgate/internal/config"
)

func TestSystemConfigHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create temporary config file
	tmpFile, err := os.CreateTemp("", "config-test-*.yaml")
	require.NoError(t, err)
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	initialCfg := &config.Config{
		Server: config.ServerConfig{
			Port:            8080,
			Mode:            "release",
			ReadTimeout:     60 * time.Second,
			WriteTimeout:    30 * time.Minute,
			IdleTimeout:     300 * time.Second,
			MaxHeaderBytes:  1048576,
			ShutdownTimeout: 30 * time.Second,
		},
		Database: config.DatabaseConfig{
			Path: "modelgate.db",
		},
		JWT: config.JWTConfig{
			Secret:      "test-secret-change-in-production",
			ExpireHours: 24,
		},
		Logs: config.LogConfig{
			DebugRawPayloads: "none",
		},
	}
	cm := config.NewManager(initialCfg, tmpPath)

	h := &Handler{
		cm: cm,
	}

	router := gin.New()
	admin := router.Group("/admin")
	{
		configGrp := admin.Group("/config")
		configGrp.GET("/system", h.GetSystemConfig)
		configGrp.PUT("/system", h.UpdateSystemConfig)
	}

	// 1. Test GET /admin/config/system
	req, _ := http.NewRequest("GET", "/admin/config/system", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var getResponse struct {
		Data RefinedSystemConfigJSON `json:"data"`
	}
	err = json.Unmarshal(resp.Body.Bytes(), &getResponse)
	require.NoError(t, err)
	assert.Equal(t, "1m0s", getResponse.Data.Server.ReadTimeout)
	assert.Equal(t, "30m0s", getResponse.Data.Server.WriteTimeout)
	assert.Equal(t, "5m0s", getResponse.Data.Server.IdleTimeout)

	// 2. Test PUT /admin/config/system (Valid)
	updatePayload := RefinedSystemConfigJSON{
		Server: RefinedServerConfigJSON{
			ReadTimeout:  "15s",
			WriteTimeout: "10m",
			IdleTimeout:  "120s",
		},
		Frontend: config.FrontendConfig{
			FeedbackURL:         "https://feedback.new.com",
			DevManualURL:        "https://docs.new.com",
			RegistrationEnabled: true,
		},
	}
	payloadBytes, _ := json.Marshal(updatePayload)
	req, _ = http.NewRequest("PUT", "/admin/config/system", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	// Verify manager state
	currentCfg := cm.GetConfig()
	assert.Equal(t, 15*time.Second, currentCfg.Server.ReadTimeout)
	assert.Equal(t, 10*time.Minute, currentCfg.Server.WriteTimeout)
	assert.Equal(t, 120*time.Second, currentCfg.Server.IdleTimeout)
	assert.Equal(t, "https://feedback.new.com", currentCfg.Frontend.FeedbackURL)
	assert.True(t, currentCfg.Frontend.RegistrationEnabled)

	// 3. Test PUT /admin/config/system (Invalid Timeout string)
	invalidPayload := RefinedSystemConfigJSON{
		Server: RefinedServerConfigJSON{
			ReadTimeout:  "invalid-duration",
			WriteTimeout: "10m",
			IdleTimeout:  "120s",
		},
		Frontend: config.FrontendConfig{
			FeedbackURL: "https://feedback.new.com",
		},
	}
	invalidBytes, _ := json.Marshal(invalidPayload)
	req, _ = http.NewRequest("PUT", "/admin/config/system", bytes.NewBuffer(invalidBytes))
	req.Header.Set("Content-Type", "application/json")
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, resp.Body.String(), "invalid read_timeout")
}
