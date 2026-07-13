package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	globalConfig *Config
	configMu     sync.RWMutex
	configOnce   sync.Once
	configPath   string
)

var defaultConfig = Config{
	Server: ServerConfig{
		Port: "8081",
		Host: "0.0.0.0",
	},
	Database: DatabaseConfig{
		Type: "sqlite",
		URL:  "aistudio.db",
	},
	JWT: JWTConfig{
		Secret: "aistudio-default-secret-change-in-production",
		Expire: "24h",
	},
	Engine: EngineConfig{
		PythonPath: "python",
		EngineDir:  "../Engine",
	},
	Plugin: PluginConfig{
		Directory: "../Plugins",
	},
	Log: LogConfig{
		Level: "info",
	},
	Websocket: WebsocketConfig{
		Port: "8082",
	},
	Task: TaskConfig{
		NumWorkers: 4,
	},
	MCP: MCPConfig{
		ConfigPath:     "config/mcp.json",
		AutoConnect:    true,
		DefaultTimeout: 30000,
	},
	LLM: LLMConfig{
		Provider:    "mock",
		Model:       "gpt-4o-mini",
		MaxTokens:   4096,
		Temperature: 0.7,
	},
}

func Get() *Config {
	configMu.RLock()
	defer configMu.RUnlock()
	if globalConfig == nil {
		panic("config: Get() called before Load(). Call config.Load() first.")
	}
	return globalConfig
}

func Load() error {
	var err error
	configOnce.Do(func() {
		err = loadInternal()
	})
	return err
}

func Reload() error {
	configMu.Lock()
	defer configMu.Unlock()
	configOnce = sync.Once{}
	configOnce.Do(func() {
		err := loadInternal()
		if err != nil {
			log.Printf("[config] reload failed: %v", err)
		} else {
			log.Printf("[config] configuration reloaded from %s", configPath)
		}
	})
	return nil
}

func loadInternal() error {
	cfg := defaultConfig

	path := resolveConfigPath()
	configPath = path

	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return fmt.Errorf("config: failed to parse %s: %w", path, err)
			}
			log.Printf("[config] loaded base config from %s", path)
		} else {
			log.Printf("[config] warning: could not load config file %s: %v", path, err)
		}
	} else {
		log.Println("[config] no config file found, using defaults")
	}

	env := resolveEnv()
	envPath := findEnvConfigPath(path, env)
	if envPath != "" {
		data, err := os.ReadFile(envPath)
		if err == nil {
			envCfg := Config{}
			if err := yaml.Unmarshal(data, &envCfg); err != nil {
				log.Printf("[config] warning: could not parse %s config file %s: %v", env, envPath, err)
			} else {
				cfg = mergeConfig(cfg, envCfg)
				log.Printf("[config] loaded %s config from %s", env, envPath)
			}
		}
	}

	applyEnvOverrides(&cfg)

	if env == "production" && cfg.JWT.Secret == "aistudio-default-secret-change-in-production" {
		log.Println("[config] WARNING: Using default JWT secret in production! Set AISTUDIO_JWT_SECRET environment variable.")
	}

	globalConfig = &cfg
	return nil
}

func resolveConfigPath() string {
	if path := os.Getenv("AISTUDIO_CONFIG"); path != "" {
		return path
	}
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

func resolveEnv() string {
	env := os.Getenv("AISTUDIO_ENV")
	if env == "" {
		env = "development"
	}
	return env
}

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

func mergeConfig(base, override Config) Config {
	result := base
	if override.Server.Port != "" {
		result.Server.Port = override.Server.Port
	}
	if override.Server.Host != "" {
		result.Server.Host = override.Server.Host
	}
	if override.Database.Type != "" {
		result.Database.Type = override.Database.Type
	}
	if override.Database.URL != "" {
		result.Database.URL = override.Database.URL
	}
	if override.JWT.Secret != "" {
		result.JWT.Secret = override.JWT.Secret
	}
	if override.JWT.Expire != "" {
		result.JWT.Expire = override.JWT.Expire
	}
	if override.Engine.PythonPath != "" {
		result.Engine.PythonPath = override.Engine.PythonPath
	}
	if override.Engine.EngineDir != "" {
		result.Engine.EngineDir = override.Engine.EngineDir
	}
	if override.Engine.Timeout != 0 {
		result.Engine.Timeout = override.Engine.Timeout
	}
	if override.Plugin.Directory != "" {
		result.Plugin.Directory = override.Plugin.Directory
	}
	if override.Log.Level != "" {
		result.Log.Level = override.Log.Level
	}
	if override.Websocket.Port != "" {
		result.Websocket.Port = override.Websocket.Port
	}
	if override.Task.NumWorkers != 0 {
		result.Task.NumWorkers = override.Task.NumWorkers
	}
	if override.LLM.Provider != "" {
		result.LLM.Provider = override.LLM.Provider
	}
	if override.LLM.APIKey != "" {
		result.LLM.APIKey = override.LLM.APIKey
	}
	if override.LLM.BaseURL != "" {
		result.LLM.BaseURL = override.LLM.BaseURL
	}
	if override.LLM.Model != "" {
		result.LLM.Model = override.LLM.Model
	}
	if override.LLM.MaxTokens != 0 {
		result.LLM.MaxTokens = override.LLM.MaxTokens
	}
	if override.LLM.Temperature != 0 {
		result.LLM.Temperature = override.LLM.Temperature
	}
	if override.LLM.Timeout != 0 {
		result.LLM.Timeout = override.LLM.Timeout
	}
	if override.MCP.ConfigPath != "" {
		result.MCP.ConfigPath = override.MCP.ConfigPath
	}
	if override.MCP.AutoConnect {
		result.MCP.AutoConnect = override.MCP.AutoConnect
	}
	if override.MCP.DefaultTimeout != 0 {
		result.MCP.DefaultTimeout = override.MCP.DefaultTimeout
	}
	return result
}

func applyEnvOverrides(cfg *Config) {
	envBindings := map[string]func(string){
		"SERVER_PORT":      func(v string) { cfg.Server.Port = v },
		"SERVER_HOST":      func(v string) { cfg.Server.Host = v },
		"DATABASE_TYPE":    func(v string) { cfg.Database.Type = v },
		"DATABASE_URL":     func(v string) { cfg.Database.URL = v },
		"JWT_SECRET":       func(v string) { cfg.JWT.Secret = v },
		"JWT_EXPIRE":       func(v string) { cfg.JWT.Expire = v },
		"ENGINE_PYTHON_PATH": func(v string) { cfg.Engine.PythonPath = v },
		"ENGINE_DIR":       func(v string) { cfg.Engine.EngineDir = v },
		"PLUGIN_DIRECTORY": func(v string) { cfg.Plugin.Directory = v },
		"LOG_LEVEL":        func(v string) { cfg.Log.Level = v },
		"WEBSOCKET_PORT":   func(v string) { cfg.Websocket.Port = v },
		"TASK_NUM_WORKERS": func(v string) { cfg.Task.NumWorkers = parseInt(v) },
		"LLM_PROVIDER":     func(v string) { cfg.LLM.Provider = v },
		"LLM_API_KEY":      func(v string) { cfg.LLM.APIKey = v },
		"LLM_BASE_URL":     func(v string) { cfg.LLM.BaseURL = v },
		"LLM_MODEL":        func(v string) { cfg.LLM.Model = v },
		"LLM_MAX_TOKENS":   func(v string) { cfg.LLM.MaxTokens = parseInt(v) },
		"LLM_TEMPERATURE":  func(v string) { cfg.LLM.Temperature = parseFloat(v) },
		"MCP_CONFIG_PATH":  func(v string) { cfg.MCP.ConfigPath = v },
		"MCP_AUTO_CONNECT": func(v string) { cfg.MCP.AutoConnect = v == "true" || v == "1" },
		"MCP_DEFAULT_TIMEOUT": func(v string) { cfg.MCP.DefaultTimeout = parseInt(v) },
	}
	for envName, setter := range envBindings {
		if v, ok := os.LookupEnv(envName); ok {
			setter(v)
		}
	}
}

func parseInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

func SetConfigPath(path string) {
	configPath = path
}
