package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Models   []ModelConfig  `yaml:"models"`
	Policies []PolicyConfig `yaml:"quota_policies"`
	Admin    AdminConfig    `yaml:"admin"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

type ModelConfig struct {
	ID      string `yaml:"id"`
	Name    string `yaml:"name"`
	Backend string `yaml:"backend"`
	Enabled bool   `yaml:"enabled"`
	Weight  int    `yaml:"weight"`
}

type PolicyConfig struct {
	Name            string   `yaml:"name"`
	RateLimit       int      `yaml:"rate_limit"`
	RateLimitWindow int      `yaml:"rate_limit_window"`
	TokenQuotaDaily int64    `yaml:"token_quota_daily"`
	Models          []string `yaml:"models"`
	Description     string   `yaml:"description"`
}

type AdminConfig struct {
	DefaultEmail    string `yaml:"default_email"`
	DefaultPassword string `yaml:"default_password"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "release"
	}
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "default-secret-change-in-production"
	}
	if cfg.JWT.ExpireHours == 0 {
		cfg.JWT.ExpireHours = 24
	}

	return &cfg, nil
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
