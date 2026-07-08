package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App       AppConfig       `yaml:"app"`
	Paths     PathsConfig     `yaml:"paths"`
	Launcher  LauncherConfig  `yaml:"launcher"`
	Backend   BackendConfig   `yaml:"backend"`
	Engine    EngineConfig    `yaml:"engine"`
	Frontend  FrontendConfig  `yaml:"frontend"`
}

type AppConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
	LogLevel    string `yaml:"log_level"`
}

type PathsConfig struct {
	Root        string `yaml:"root"`
	BackendDir  string `yaml:"backend_dir"`
	FrontendDir string `yaml:"frontend_dir"`
	EngineDir   string `yaml:"engine_dir"`
	PluginsDir  string `yaml:"plugins_dir"`
	StorageDir  string `yaml:"storage_dir"`
	ConfigDir   string `yaml:"config_dir"`
	DataDir     string `yaml:"data_dir"`
	ModelsDir   string `yaml:"models_dir"`
	LogsDir     string `yaml:"logs_dir"`
}

type LauncherConfig struct {
	StartupTimeout      int  `yaml:"startup_timeout"`
	ShutdownTimeout     int  `yaml:"shutdown_timeout"`
	HealthCheckInterval int  `yaml:"health_check_interval"`
	AutoRestart         bool `yaml:"auto_restart"`
	MaxRestartAttempts int  `yaml:"max_restart_attempts"`
}

type BackendConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type EngineConfig struct {
	PythonPath   string `yaml:"python_path"`
	EngineDir    string `yaml:"engine_dir"`
	RunnerScript string `yaml:"runner_script"`
	VenvPath     string `yaml:"venv_path"`
	Port         string `yaml:"port"`
}

type FrontendConfig struct {
	Path        string `yaml:"path"`
	BackendAddr string `yaml:"backend_addr"`
}

func Load() (*Config, error) {
	configPath := os.Getenv("AISTUDIO_CONFIG")
	if configPath == "" {
		configPath = "Config/app.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := cfg.resolvePaths(); err != nil {
		return nil, fmt.Errorf("failed to resolve paths: %w", err)
	}

	return &cfg, nil
}

func (c *Config) resolvePaths() error {
	absPaths := map[string]*string{
		"root":          &c.Paths.Root,
		"backend_dir":   &c.Paths.BackendDir,
		"frontend_dir":  &c.Paths.FrontendDir,
		"engine_dir":    &c.Paths.EngineDir,
		"plugins_dir":  &c.Paths.PluginsDir,
		"storage_dir":  &c.Paths.StorageDir,
		"config_dir":   &c.Paths.ConfigDir,
		"data_dir":     &c.Paths.DataDir,
		"models_dir":   &c.Paths.ModelsDir,
		"logs_dir":     &c.Paths.LogsDir,
	}

	root := c.Paths.Root
	if root == "" {
		root = "."
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	for key, ptr := range absPaths {
		if *ptr == "" {
			continue
		}
		if !filepath.IsAbs(*ptr) {
			*ptr = filepath.Join(absRoot, *ptr)
		}
	}

	return nil
}