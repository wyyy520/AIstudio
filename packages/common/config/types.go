package config

import "time"

type Config struct {
	Server    ServerConfig    `yaml:"server" json:"server"`
	Database  DatabaseConfig  `yaml:"database" json:"database"`
	JWT       JWTConfig       `yaml:"jwt" json:"jwt"`
	Engine    EngineConfig    `yaml:"engine" json:"engine"`
	Plugin    PluginConfig    `yaml:"plugin" json:"plugin"`
	Log       LogConfig       `yaml:"log" json:"log"`
	Websocket WebsocketConfig `yaml:"websocket" json:"websocket"`
	Task      TaskConfig      `yaml:"task" json:"task"`
	LLM       LLMConfig       `yaml:"llm" json:"llm"`
	MCP       MCPConfig       `yaml:"mcp" json:"mcp"`
}

type ServerConfig struct {
	Port string `yaml:"port" json:"port"`
	Host string `yaml:"host" json:"host"`
}

func (c *ServerConfig) Addr() string {
	if c.Host == "" {
		return ":" + c.Port
	}
	return c.Host + ":" + c.Port
}

type DatabaseConfig struct {
	Type string `yaml:"type" json:"type"`
	URL  string `yaml:"url" json:"url"`
}

func (c *DatabaseConfig) DSN() string {
	return c.URL
}

type JWTConfig struct {
	Secret string `yaml:"secret" json:"secret"`
	Expire string `yaml:"expire" json:"expire"`
}

type EngineConfig struct {
	PythonPath string        `yaml:"python_path" json:"python_path"`
	EngineDir  string        `yaml:"engine_dir" json:"engine_dir"`
	Timeout    time.Duration `yaml:"timeout" json:"timeout"`
}

type PluginConfig struct {
	Directory string `yaml:"directory" json:"directory"`
}

type LogConfig struct {
	Level string `yaml:"level" json:"level"`
}

type WebsocketConfig struct {
	Port string `yaml:"port" json:"port"`
}

type TaskConfig struct {
	NumWorkers int `yaml:"num_workers" json:"num_workers"`
}

type LLMConfig struct {
	Provider    string        `yaml:"provider" json:"provider"`
	APIKey      string        `yaml:"api_key" json:"api_key"`
	BaseURL     string        `yaml:"base_url" json:"base_url"`
	Model       string        `yaml:"model" json:"model"`
	MaxTokens   int           `yaml:"max_tokens" json:"max_tokens"`
	Temperature float64       `yaml:"temperature" json:"temperature"`
	Timeout     time.Duration `yaml:"timeout" json:"timeout"`
}

type MCPConfig struct {
	ConfigPath     string `yaml:"config_path" json:"config_path"`
	AutoConnect    bool   `yaml:"auto_connect" json:"auto_connect"`
	DefaultTimeout int    `yaml:"default_timeout" json:"default_timeout"`
}
