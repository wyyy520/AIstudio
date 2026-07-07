// Package config provides a unified configuration management system for AIStudio Backend.
//
// Built on top of Viper (github.com/spf13/viper), it supports:
//   - YAML configuration files (default.yaml, development.yaml, production.yaml)
//   - Environment variable overrides (highest priority)
//   - Default values (lowest priority)
//   - Multi-environment profiles
//   - Hot-reload via config.Reload()
//
// Loading priority (highest to lowest):
//  1. Environment variables (e.g. SERVER_PORT=9090)
//  2. Environment-specific config file (config/{AISTUDIO_ENV}.yaml)
//  3. Base config file (config/default.yaml)
//  4. Default values (hardcoded)
//
// Usage:
//
//	import "github.com/aistudio/backend/internal/config"
//
//	// Load config at startup
//	if err := config.Load(); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Access config values
//	serverPort := config.Get().Server.Port
//	dbURL := config.Get().Database.URL
//
//	// Reload configuration at runtime
//	if err := config.Reload(); err != nil {
//	    log.Printf("reload failed: %v", err)
//	}
package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

// ---------------------------------------------------------------------------
// Configuration Structure
// ---------------------------------------------------------------------------

// Config is the top-level configuration structure for the entire backend.
// Each field maps to a YAML section and can be overridden by environment variables.
type Config struct {
	Server    ServerConfig    `mapstructure:"server" json:"server"`
	Database  DatabaseConfig  `mapstructure:"database" json:"database"`
	JWT       JWTConfig       `mapstructure:"jwt" json:"jwt"`
	Engine    EngineConfig    `mapstructure:"engine" json:"engine"`
	Plugin    PluginConfig    `mapstructure:"plugin" json:"plugin"`
	Log       LogConfig       `mapstructure:"log" json:"log"`
	Websocket WebsocketConfig `mapstructure:"websocket" json:"websocket"`
	Task      TaskConfig      `mapstructure:"task" json:"task"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port string `mapstructure:"port" json:"port"`
	Host string `mapstructure:"host" json:"host"`
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Type string `mapstructure:"type" json:"type"`
	URL  string `mapstructure:"url" json:"url"`
}

// JWTConfig holds JWT authentication settings.
type JWTConfig struct {
	Secret string `mapstructure:"secret" json:"secret"`
	Expire string `mapstructure:"expire" json:"expire"` // duration string, e.g. "24h"
}

// EngineConfig holds AI Engine (gRPC) connection settings.
type EngineConfig struct {
	Address  string `mapstructure:"address" json:"address"`
	GrpcPort int    `mapstructure:"grpc_port" json:"grpc_port"`
}

// PluginConfig holds plugin system settings.
type PluginConfig struct {
	Directory string `mapstructure:"directory" json:"directory"`
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level string `mapstructure:"level" json:"level"` // "debug", "info", "warn", "error"
}

// WebsocketConfig holds WebSocket server settings.
type WebsocketConfig struct {
	Port string `mapstructure:"port" json:"port"`
}

// TaskConfig holds task manager settings.
type TaskConfig struct {
	NumWorkers int `mapstructure:"num_workers" json:"num_workers"`
}

// ---------------------------------------------------------------------------
// Helper Methods
// ---------------------------------------------------------------------------

// Addr returns the server listen address (host:port).
func (c *ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// DSN returns the GORM-compatible database data source name from the config.
func (c *DatabaseConfig) DSN() string {
	return c.URL
}

// GrpcAddr returns the gRPC engine address (host:port).
func (c *EngineConfig) GrpcAddr() string {
	return fmt.Sprintf("%s:%d", c.Address, c.GrpcPort)
}

// ---------------------------------------------------------------------------
// Default Configuration
// ---------------------------------------------------------------------------

// setDefaults registers all default configuration values with Viper.
// These are used as the lowest priority (fallback) when no config file or env var is set.
func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", "8081")
	v.SetDefault("server.host", "0.0.0.0")

	v.SetDefault("database.type", "sqlite")
	v.SetDefault("database.url", "aistudio.db")

	v.SetDefault("jwt.secret", "aistudio-default-secret-change-in-production")
	v.SetDefault("jwt.expire", "24h")

	v.SetDefault("engine.address", "localhost")
	v.SetDefault("engine.grpc_port", 50051)

	v.SetDefault("plugin.directory", "../Plugins")

	v.SetDefault("log.level", "info")

	v.SetDefault("websocket.port", "8082")

	v.SetDefault("task.num_workers", 4)
}

// ---------------------------------------------------------------------------
// Global State
// ---------------------------------------------------------------------------

var (
	globalConfig *Config
	globalViper  *viper.Viper
	configOnce   sync.Once
	configMu     sync.RWMutex
	configPath   string // resolved path to the base config file
)

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// Get returns the current global configuration.
// It panics if Load() has not been called yet.
// The returned pointer is safe for concurrent reads.
func Get() *Config {
	configMu.RLock()
	defer configMu.RUnlock()

	if globalConfig == nil {
		panic("config: Get() called before Load(). Call config.Load() first.")
	}
	return globalConfig
}

// Load initializes the configuration system using Viper.
//
// It loads configuration from the default path (config/default.yaml) or the
// path specified by the AISTUDIO_CONFIG environment variable.
// The environment variable AISTUDIO_ENV selects the active environment profile
// (e.g., "development", "production").
//
// Loading order:
//  1. Default values (hardcoded via setDefaults)
//  2. Base config file (config/default.yaml)
//  3. Environment-specific config file (config/{env}.yaml)
//  4. Environment variables (highest priority)
func Load() error {
	var err error
	configOnce.Do(func() {
		err = loadInternal("")
	})
	return err
}

// Reload re-reads the configuration from disk and merges it with the current
// environment variables. This is useful for hot-reloading config changes.
// It returns an error if the config file cannot be read or parsed.
func Reload() error {
	configMu.Lock()
	defer configMu.Unlock()

	configOnce = sync.Once{} // reset sync.Once to allow reload

	v := viper.New()
	setDefaults(v)

	cfgPath := resolveConfigPath()
	if cfgPath != "" {
		v.SetConfigFile(cfgPath)
		if err := v.ReadInConfig(); err != nil {
			return fmt.Errorf("config reload failed: %w", err)
		}
	}

	// Load environment-specific config
	env := resolveEnv()
	envPath := findEnvConfigPath(cfgPath, env)
	if envPath != "" {
		tmpViper := viper.New()
		tmpViper.SetConfigFile(envPath)
		if err := tmpViper.ReadInConfig(); err == nil {
			if err := v.MergeConfigMap(tmpViper.AllSettings()); err != nil {
				log.Printf("[config] warning: merge %s failed: %v", envPath, err)
			}
		}
	}

	// Bind environment variables
	bindEnvVars(v)

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("config unmarshal failed: %w", err)
	}

	globalConfig = cfg
	globalViper = v
	log.Printf("[config] configuration reloaded from %s (env=%s)", cfgPath, env)
	return nil
}

// ---------------------------------------------------------------------------
// Internal: Load Logic
// ---------------------------------------------------------------------------

// loadInternal performs the actual config loading logic using Viper.
func loadInternal(explicitPath string) error {
	v := viper.New()

	// Step 1: Register defaults
	setDefaults(v)

	// Step 2: Determine config path
	cfgPath := explicitPath
	if cfgPath == "" {
		cfgPath = resolveConfigPath()
	}
	configPath = cfgPath

	// Step 3: Load base config file
	if cfgPath != "" {
		v.SetConfigFile(cfgPath)
		if err := v.ReadInConfig(); err != nil {
			log.Printf("[config] warning: could not load config file %s: %v", cfgPath, err)
		} else {
			log.Printf("[config] loaded base config from %s", cfgPath)
		}
	} else {
		log.Println("[config] no config file found, using defaults")
	}

	// Step 4: Load environment-specific config file
	env := resolveEnv()
	envPath := findEnvConfigPath(cfgPath, env)
	if envPath != "" {
		tmpViper := viper.New()
		tmpViper.SetConfigFile(envPath)
		if err := tmpViper.ReadInConfig(); err != nil {
			log.Printf("[config] warning: could not load %s config file %s: %v", env, envPath, err)
		} else {
			if err := v.MergeConfigMap(tmpViper.AllSettings()); err != nil {
				log.Printf("[config] warning: merge %s failed: %v", envPath, err)
			} else {
				log.Printf("[config] loaded %s config from %s", env, envPath)
			}
		}
	}

	// Step 5: Load .env file (for local development convenience)
	tryLoadDotEnv()

	// Step 6: Bind environment variables (highest priority)
	bindEnvVars(v)

	// Step 7: Unmarshal into Config struct
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("config unmarshal failed: %w", err)
	}

	globalConfig = cfg
	globalViper = v
	log.Printf("[config] configuration initialized (env=%s)", env)
	return nil
}

// ---------------------------------------------------------------------------
// Internal: Path Resolution
// ---------------------------------------------------------------------------

// resolveConfigPath determines the config file path from env var or searches
// standard locations.
func resolveConfigPath() string {
	if path := os.Getenv("AISTUDIO_CONFIG"); path != "" {
		return path
	}

	// Search standard locations
	candidates := []string{
		"config/default.yaml",
		"config/default.yml",
		"../config/default.yaml",
		"../../config/default.yaml",
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	// Check relative to the executable
	exec, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(exec)
		for _, c := range []string{
			"config/default.yaml",
			"../config/default.yaml",
			"../../config/default.yaml",
		} {
			abs := filepath.Join(execDir, c)
			if _, err := os.Stat(abs); err == nil {
				return abs
			}
		}
	}

	return ""
}

// resolveEnv returns the current environment profile name.
func resolveEnv() string {
	env := os.Getenv("AISTUDIO_ENV")
	if env == "" {
		env = "development"
	}
	return env
}

// findEnvConfigPath locates the environment-specific config file.
func findEnvConfigPath(basePath, env string) string {
	baseDir := filepath.Dir(basePath)
	if baseDir == "." {
		baseDir = "config"
	}

	candidates := []string{
		filepath.Join(baseDir, env+".yaml"),
		filepath.Join(baseDir, env+".yml"),
		filepath.Join("config", env+".yaml"),
		filepath.Join("config", env+".yml"),
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	return ""
}

// ---------------------------------------------------------------------------
// Internal: Environment Variable Binding
// ---------------------------------------------------------------------------

// bindEnvVars tells Viper to automatically read environment variables and
// map them to config keys. Viper uses the following rules:
//   - SERVER_PORT  → server.port
//   - DATABASE_URL → database.url
//   - Environment variables take precedence over config files.
func bindEnvVars(v *viper.Viper) {
	v.AutomaticEnv()

	// Replace underscores with dots for nested config keys
	// This allows SERVER_PORT to map to server.port
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Explicit bindings for documentation clarity
	envBindings := map[string]string{
		"SERVER_PORT":      "server.port",
		"SERVER_HOST":      "server.host",
		"DATABASE_TYPE":    "database.type",
		"DATABASE_URL":     "database.url",
		"JWT_SECRET":       "jwt.secret",
		"JWT_EXPIRE":       "jwt.expire",
		"ENGINE_ADDRESS":   "engine.address",
		"ENGINE_GRPC_PORT": "engine.grpc_port",
		"PLUGIN_DIRECTORY": "plugin.directory",
		"LOG_LEVEL":        "log.level",
		"WEBSOCKET_PORT":   "websocket.port",
		"TASK_NUM_WORKERS": "task.num_workers",
	}

	for envName, configKey := range envBindings {
		_ = v.BindEnv(configKey, envName)
	}
}

// ---------------------------------------------------------------------------
// Internal: .env File Support
// ---------------------------------------------------------------------------

// tryLoadDotEnv attempts to load the .env file for local development.
// This is a convenience for developers who prefer .env over YAML for secrets.
func tryLoadDotEnv() {
	locations := []string{".env", "../.env", "../../.env"}
	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			loadDotEnvFile(loc)
			return
		}
	}
}

// loadDotEnvFile reads a .env file and sets environment variables.
// Only sets variables that are not already set in the environment.
func loadDotEnvFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	lines := splitLines(string(data))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		if key != "" && os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

// splitLines splits a string into lines, handling both \r\n and \n.
func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.Split(s, "\n")
}