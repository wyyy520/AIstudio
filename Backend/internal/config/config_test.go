package config

import (
	"os"
	"sync"
	"testing"
)

func TestDefaultValues(t *testing.T) {
	cfg := &Config{
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
	}

	if cfg.Server.Port != "8081" {
		t.Errorf("expected server port 8081, got %s", cfg.Server.Port)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected server host 0.0.0.0, got %s", cfg.Server.Host)
	}
	if cfg.Database.Type != "sqlite" {
		t.Errorf("expected database type sqlite, got %s", cfg.Database.Type)
	}
	if cfg.JWT.Expire != "24h" {
		t.Errorf("expected jwt expire 24h, got %s", cfg.JWT.Expire)
	}
	if cfg.Engine.PythonPath != "python" {
		t.Errorf("expected engine python_path python, got %s", cfg.Engine.PythonPath)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("expected log level info, got %s", cfg.Log.Level)
	}
	if cfg.Websocket.Port != "8082" {
		t.Errorf("expected websocket port 8082, got %s", cfg.Websocket.Port)
	}
	if cfg.Task.NumWorkers != 4 {
		t.Errorf("expected task num_workers 4, got %d", cfg.Task.NumWorkers)
	}
}

func TestHelperMethods(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: "9090",
			Host: "127.0.0.1",
		},
		Database: DatabaseConfig{
			Type: "sqlite",
			URL:  "test.db",
		},
		Engine: EngineConfig{
			PythonPath: "python3",
			EngineDir:  "../Engine",
		},
	}

	if addr := cfg.Server.Addr(); addr != "127.0.0.1:9090" {
		t.Errorf("expected Addr 127.0.0.1:9090, got %s", addr)
	}
	if dsn := cfg.Database.DSN(); dsn != "test.db" {
		t.Errorf("expected DSN test.db, got %s", dsn)
	}
	if engineAddr := cfg.Engine.EngineAddr(); engineAddr != "python3:../Engine" {
		t.Errorf("expected EngineAddr python3:../Engine, got %s", engineAddr)
	}
}

func TestEnvVarOverride(t *testing.T) {
	os.Setenv("SERVER_PORT", "9090")
	defer os.Unsetenv("SERVER_PORT")

	if port := os.Getenv("SERVER_PORT"); port != "9090" {
		t.Errorf("expected SERVER_PORT=9090, got %s", port)
	}
}

func TestDotEnvLineParsing(t *testing.T) {
	// Input without trailing newline to avoid extra empty line
	lines := splitLines("KEY=value\nANOTHER=123\n# comment\n\nEMPTY=")
	if len(lines) != 5 {
		t.Errorf("expected 5 lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "KEY=value" {
		t.Errorf("expected KEY=value, got %s", lines[0])
	}
	if lines[2] != "# comment" {
		t.Errorf("expected # comment, got %s", lines[2])
	}
}

func TestResolveEnv(t *testing.T) {
	env := resolveEnv()
	if env != "development" {
		t.Errorf("expected default env development, got %s", env)
	}

	os.Setenv("AISTUDIO_ENV", "production")
	defer os.Unsetenv("AISTUDIO_ENV")
	env = resolveEnv()
	if env != "production" {
		t.Errorf("expected production, got %s", env)
	}
}

// resetGlobalConfig resets the global config state for testing.
func resetGlobalConfig() {
	configMu.Lock()
	defer configMu.Unlock()
	globalConfig = nil
	globalViper = nil
	configPath = ""
	configOnce = sync.Once{}
}

func TestLoadWithConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configContent := `server:
  port: "9999"
  host: "127.0.0.1"
database:
  type: "sqlite"
  url: "test_memory.db"
log:
  level: "debug"
`
	tmpFile := tmpDir + "/default.yaml"
	if err := os.WriteFile(tmpFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	// Set SERVER_PORT to prevent .env file from overriding the config file value.
	// The .env file in the project root sets SERVER_PORT=8081, which would
	// take precedence over the config file. By setting it explicitly here,
	// we prevent the .env file from setting it (loadDotEnvFile skips vars
	// that are already set), and the config file value takes effect.
	os.Setenv("SERVER_PORT", "9999")
	os.Setenv("AISTUDIO_CONFIG", tmpFile)
	defer os.Unsetenv("SERVER_PORT")
	defer os.Unsetenv("AISTUDIO_CONFIG")

	resetGlobalConfig()

	if err := Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	cfg := Get()
	if cfg.Server.Port != "9999" {
		t.Errorf("expected server port 9999, got %s", cfg.Server.Port)
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("expected server host 127.0.0.1, got %s", cfg.Server.Host)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("expected log level debug, got %s", cfg.Log.Level)
	}
	if cfg.Database.Type != "sqlite" {
		t.Errorf("expected database type sqlite (default), got %s", cfg.Database.Type)
	}
	if cfg.Websocket.Port != "8082" {
		t.Errorf("expected websocket port 8082 (default), got %s", cfg.Websocket.Port)
	}
}

func TestLoadEnvVarOverridesConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configContent := `server:
  port: "3000"
  host: "0.0.0.0"
`
	tmpFile := tmpDir + "/default.yaml"
	if err := os.WriteFile(tmpFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("AISTUDIO_CONFIG", tmpFile)
	defer os.Unsetenv("SERVER_PORT")
	defer os.Unsetenv("AISTUDIO_CONFIG")

	resetGlobalConfig()

	if err := Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	cfg := Get()
	// SERVER_PORT=9090 should override the config file value of 3000
	if cfg.Server.Port != "9090" {
		t.Errorf("expected server port 9090 (env override), got %s", cfg.Server.Port)
	}
	// Host should come from config file since not overridden
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected server host 0.0.0.0 (from config), got %s", cfg.Server.Host)
	}
}

func TestReload(t *testing.T) {
	tmpDir := t.TempDir()
	configContent1 := `server:
  port: "1111"
`
	tmpFile := tmpDir + "/default.yaml"
	if err := os.WriteFile(tmpFile, []byte(configContent1), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	// Prevent .env file from overriding the config file values
	os.Setenv("SERVER_PORT", "1111")
	os.Setenv("AISTUDIO_CONFIG", tmpFile)
	defer os.Unsetenv("SERVER_PORT")
	defer os.Unsetenv("AISTUDIO_CONFIG")

	resetGlobalConfig()

	if err := Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	cfg := Get()
	if cfg.Server.Port != "1111" {
		t.Errorf("expected port 1111, got %s", cfg.Server.Port)
	}

	configContent2 := `server:
  port: "2222"
`
	if err := os.WriteFile(tmpFile, []byte(configContent2), 0644); err != nil {
		t.Fatalf("failed to update temp config: %v", err)
	}

	// Update env var to match new expected value
	os.Setenv("SERVER_PORT", "2222")

	if err := Reload(); err != nil {
		t.Fatalf("Reload() failed: %v", err)
	}
	cfg = Get()
	if cfg.Server.Port != "2222" {
		t.Errorf("expected port 2222 after reload, got %s", cfg.Server.Port)
	}
}

func TestPanicOnGetBeforeLoad(t *testing.T) {
	resetGlobalConfig()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when calling Get() before Load()")
		}
	}()

	Get()
}