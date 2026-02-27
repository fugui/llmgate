package config

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the complete application configuration
type Config struct {
	Server    ServerConfig             `yaml:"server"`
	Providers map[string]ProviderConfig `yaml:"providers"`
	Routing   RoutingConfig            `yaml:"routing"`
	RateLimit RateLimitConfig          `yaml:"rate_limit"`
	Logging   LoggingConfig            `yaml:"logging"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

// ProviderConfig holds LLM provider configuration
type ProviderConfig struct {
	BaseURL  string        `yaml:"base_url"`
	APIKey   string        `yaml:"api_key"`
	Timeout  time.Duration `yaml:"timeout"`
	Priority int           `yaml:"priority"`
}

// RoutingRule defines routing rules for specific paths
type RoutingRule struct {
	Path      string   `yaml:"path"`
	Providers []string `yaml:"providers"`
}

// RoutingConfig holds request routing configuration
type RoutingConfig struct {
	DefaultProvider string        `yaml:"default_provider"`
	Rules           []RoutingRule `yaml:"rules"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool  `yaml:"enabled"`
	RequestsPerSecond int   `yaml:"requests_per_second"`
	BurstSize         int   `yaml:"burst_size"`
	PerIP             bool  `yaml:"per_ip"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level          string `yaml:"level"`
	Format         string `yaml:"format"`
	Output         string `yaml:"output"`
	FilePath       string `yaml:"file_path"`
	RequestLogging bool   `yaml:"request_logging"`
	ResponseLogging bool  `yaml:"response_logging"`
}

var envVarRegex = regexp.MustCompile(`\$\{([^}]+)\}`)

// Load reads configuration from the specified file path
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables
	expanded := expandEnvVars(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults
	setDefaults(&cfg)

	return &cfg, nil
}

func expandEnvVars(content string) string {
	return envVarRegex.ReplaceAllStringFunc(content, func(match string) string {
		varName := match[2 : len(match)-1] // Remove ${ and }
		if value := os.Getenv(varName); value != "" {
			return value
		}
		return match
	})
}

func setDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 30 * time.Second
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 30 * time.Second
	}
	if cfg.Server.IdleTimeout == 0 {
		cfg.Server.IdleTimeout = 120 * time.Second
	}
	if cfg.RateLimit.RequestsPerSecond == 0 {
		cfg.RateLimit.RequestsPerSecond = 10
	}
	if cfg.RateLimit.BurstSize == 0 {
		cfg.RateLimit.BurstSize = 20
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}
	if cfg.Logging.Output == "" {
		cfg.Logging.Output = "stdout"
	}
}

// GetProvider returns the configuration for a specific provider
func (c *Config) GetProvider(name string) (ProviderConfig, bool) {
	provider, ok := c.Providers[name]
	return provider, ok
}

// GetRoutingRule returns the routing rule for a given path
func (c *Config) GetRoutingRule(path string) *RoutingRule {
	for _, rule := range c.Routing.Rules {
		if rule.Path == path {
			return &rule
		}
	}
	return nil
}
