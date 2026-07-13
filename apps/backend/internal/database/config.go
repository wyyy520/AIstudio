package database

import (
	"fmt"
	"os"
)

// Config holds database configuration.
type Config struct {
	Type string // "sqlite" or "postgres"
	URL  string // connection string
}

// LoadConfig loads database configuration from environment variables.
// Defaults to SQLite with local file storage.
//
// Deprecated: Use config.Load() + config.Get().Database instead.
// This function is kept for backward compatibility.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		Type: os.Getenv("DATABASE_TYPE"),
		URL:  os.Getenv("DATABASE_URL"),
	}

	// Apply defaults
	if cfg.Type == "" {
		cfg.Type = "sqlite"
	}
	if cfg.URL == "" {
		cfg.URL = "aistudio.db"
	}

	// Validate
	switch cfg.Type {
	case "sqlite", "postgres":
		// valid
	default:
		return nil, fmt.Errorf("unsupported database type: %s (supported: sqlite, postgres)", cfg.Type)
	}

	if cfg.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

// DSN returns the GORM-compatible data source name.
func (c *Config) DSN() string {
	return c.URL
}