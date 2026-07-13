package engine

import "time"

type Config struct {
	BaseURL    string
	Timeout    time.Duration
	RetryCount int
	RetryDelay time.Duration
}

func DefaultConfig() *Config {
	return &Config{
		BaseURL:    "http://localhost:8082",
		Timeout:    30 * time.Second,
		RetryCount: 3,
		RetryDelay: 1 * time.Second,
	}
}
