// =============================================================================
// AIStudio Launcher - 配置系统
// =============================================================================
// 功能：读取 Config/app.yaml，解析配置项，解析相对路径为绝对路径
// =============================================================================

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// -----------------------------------------------------------------------------
// 配置结构体定义
// -----------------------------------------------------------------------------

// Config 是顶层配置结构，对应 Config/app.yaml
type Config struct {
	App      AppConfig      `yaml:"app"`      // 应用元信息
	Paths    PathsConfig    `yaml:"paths"`    // 路径配置
	Launcher LauncherConfig `yaml:"launcher"` // Launcher 自身配置
	Backend  BackendConfig  `yaml:"backend"`  // Backend 服务配置
	Engine   EngineConfig   `yaml:"engine"`   // Python Engine 配置
	Frontend FrontendConfig `yaml:"frontend"` // Frontend (Tauri) 配置
}

// AppConfig - 应用元信息
type AppConfig struct {
	Name        string `yaml:"name"`         // 应用名称
	Version     string `yaml:"version"`      // 版本号
	Environment string `yaml:"environment"`  // 运行环境：development / staging / production
	LogLevel    string `yaml:"log_level"`    // 日志级别：debug / info / warn / error
}

// PathsConfig - 路径配置（相对于项目根目录）
type PathsConfig struct {
	Root        string `yaml:"root"`         // 项目根目录
	BackendDir  string `yaml:"backend_dir"`  // Backend 目录
	FrontendDir string `yaml:"frontend_dir"` // Frontend 目录
	EngineDir   string `yaml:"engine_dir"`   // Engine 目录
	PluginsDir  string `yaml:"plugins_dir"`  // Plugins 目录
	RuntimeDir  string `yaml:"runtime_dir"`  // Runtime 目录
	StorageDir  string `yaml:"storage_dir"`  // Storage 目录
	ConfigDir   string `yaml:"config_dir"`   // Config 目录
	LogsDir     string `yaml:"logs_dir"`     // 日志目录
}

// LauncherConfig - Launcher 自身行为配置
type LauncherConfig struct {
	StartupTimeout      int  `yaml:"startup_timeout"`       // 启动超时（秒），等待服务健康的最大时间
	ShutdownTimeout     int  `yaml:"shutdown_timeout"`      // 关闭超时（秒），等待进程退出的最大时间
	HealthCheckInterval int  `yaml:"health_check_interval"` // 健康检查间隔（秒）
	AutoRestart         bool `yaml:"auto_restart"`          // 是否自动重启崩溃的服务
	MaxRestartAttempts  int  `yaml:"max_restart_attempts"`  // 最大重启尝试次数
}

// BackendConfig - Backend 服务配置
type BackendConfig struct {
	Executable string `yaml:"executable"` // 可执行文件名（如 cmd.exe）
	Host       string `yaml:"host"`       // 监听地址
	Port       string `yaml:"port"`       // 监听端口
	HealthPath string `yaml:"health_path"` // 健康检查路径（如 /api/health）
}

// EngineConfig - Python Engine 配置
type EngineConfig struct {
	PythonPath string `yaml:"python_path"`   // Python 解释器路径
	Executable string `yaml:"executable"`    // Engine 服务脚本（如 server.py）
	EngineDir  string `yaml:"engine_dir"`    // Engine 工作目录
	VenvPath   string `yaml:"venv_path"`     // 虚拟环境路径（可选）
	Host       string `yaml:"host"`          // Engine HTTP 监听地址
	Port       string `yaml:"port"`          // Engine HTTP 服务端口
	HealthPath string `yaml:"health_path"`   // 健康检查路径
}

// FrontendConfig - Tauri Frontend 配置
type FrontendConfig struct {
	Executable  string `yaml:"executable"`   // Tauri 可执行文件名（如 ai-studio.exe）
	FrontendDir string `yaml:"frontend_dir"` // src-tauri 目录（相对 Frontend）
	BackendAddr string `yaml:"backend_addr"` // Backend 地址，传递给 Frontend
	DevServer   bool   `yaml:"dev_server"`   // 是否启动 Vite dev server（debug build 需要）
	VitePort    string `yaml:"vite_port"`    // Vite dev server 端口
	ViteCmd     string `yaml:"vite_cmd"`     // Vite 启动命令
	ViteDir     string `yaml:"vite_dir"`     // Vite 工作目录（相对于项目根目录）
}

// -----------------------------------------------------------------------------
// LoadConfig - 读取并解析配置文件
// -----------------------------------------------------------------------------
// 返回：
//   *Config - 解析后的配置
//   error   - 读取或解析失败时返回错误
// -----------------------------------------------------------------------------

func LoadConfig() (*Config, error) {
	configPath := os.Getenv("AISTUDIO_CONFIG")
	if configPath == "" {
		configPath = findConfigFile("Config", "app.yaml")
	}

	if configPath == "" {
		return nil, fmt.Errorf("找不到配置文件 Config/app.yaml，请确认 AIStudio.exe 位于项目根目录")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败 [%s]: %w", configPath, err)
	}

	// 解析 YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	configDir := filepath.Dir(filepath.Dir(configPath))

	if err := cfg.resolvePathsFrom(configDir); err != nil {
		return nil, fmt.Errorf("解析路径失败: %w", err)
	}

	return &cfg, nil
}

// -----------------------------------------------------------------------------
// resolvePaths - 将所有相对路径转换为绝对路径
// -----------------------------------------------------------------------------

func (c *Config) resolvePathsFrom(baseDir string) error {
	root := c.Paths.Root
	if root == "" {
		root = baseDir
	}
	if !filepath.IsAbs(root) {
		root, err := filepath.Abs(root)
		if err != nil {
			return fmt.Errorf("无法获取根目录绝对路径: %w", err)
		}
		root = root
	}
	c.Paths.Root = root

	// 逐个解析路径字段
	paths := map[string]*string{
		"backend_dir":  &c.Paths.BackendDir,
		"frontend_dir": &c.Paths.FrontendDir,
		"engine_dir":   &c.Paths.EngineDir,
		"plugins_dir":  &c.Paths.PluginsDir,
		"runtime_dir":  &c.Paths.RuntimeDir,
		"storage_dir":  &c.Paths.StorageDir,
		"config_dir":   &c.Paths.ConfigDir,
		"logs_dir":     &c.Paths.LogsDir,
	}

	for _, ptr := range paths {
		if *ptr == "" {
			continue
		}
		if !filepath.IsAbs(*ptr) {
			*ptr = filepath.Join(root, *ptr)
		}
	}

	// 解析 Backend 可执行文件完整路径
	if c.Backend.Executable != "" && !filepath.IsAbs(c.Backend.Executable) {
		c.Backend.Executable = filepath.Join(c.Paths.BackendDir, c.Backend.Executable)
	}

	// 解析 Engine 相关路径
	if c.Engine.EngineDir != "" && !filepath.IsAbs(c.Engine.EngineDir) {
		c.Engine.EngineDir = filepath.Join(root, c.Engine.EngineDir)
	}
	if c.Engine.Executable != "" && !filepath.IsAbs(c.Engine.Executable) {
		c.Engine.Executable = filepath.Join(c.Paths.EngineDir, c.Engine.Executable)
	}
	if c.Engine.VenvPath != "" && !filepath.IsAbs(c.Engine.VenvPath) {
		c.Engine.VenvPath = filepath.Join(root, c.Engine.VenvPath)
	}
	if c.Engine.Host == "" {
		c.Engine.Host = "0.0.0.0"
	}
	if c.Engine.Port == "" {
		c.Engine.Port = "8082"
	}

	// 解析 Frontend 可执行文件路径
	if c.Frontend.FrontendDir != "" && !filepath.IsAbs(c.Frontend.FrontendDir) {
		c.Frontend.FrontendDir = filepath.Join(c.Paths.FrontendDir, c.Frontend.FrontendDir)
	}
	if c.Frontend.Executable != "" && !filepath.IsAbs(c.Frontend.Executable) {
		c.Frontend.Executable = filepath.Join(c.Frontend.FrontendDir, c.Frontend.Executable)
	}
	// 解析 Vite 工作目录
	if c.Frontend.ViteDir != "" && !filepath.IsAbs(c.Frontend.ViteDir) {
		c.Frontend.ViteDir = filepath.Join(root, c.Frontend.ViteDir)
	}

	return nil
}

// -----------------------------------------------------------------------------
// CheckDependencies - 检查运行环境依赖
// -----------------------------------------------------------------------------
// 检查项：
//   1. Backend 可执行文件是否存在
//   2. Python 解释器是否可用
//   3. Engine 脚本是否存在
//   4. Frontend 可执行文件是否存在（如果不存在尝试在 target/ 下查找）
// -----------------------------------------------------------------------------

func CheckDependencies(cfg *Config) error {
	logInfo("开始检查运行环境依赖...")

	var errs []error

	// 1. 检查 Backend 可执行文件
	backendExe := cfg.Backend.Executable
	if backendExe == "" {
		backendExe = filepath.Join(cfg.Paths.BackendDir, "aistudio-backend.exe")
	}
	if !fileExists(backendExe) {
		// Try go build as fallback
		logWarn("Backend 可执行文件不存在，尝试编译", "expected", backendExe)
	}

	// 2. 检查 Python 解释器
	pythonPath := cfg.Engine.PythonPath
	if pythonPath == "" {
		pythonPath = "python"
	}
	if _, err := exec.LookPath(pythonPath); err != nil {
		// LookPath 失败，尝试直接检查文件是否存在
		if !fileExists(pythonPath) {
			errs = append(errs, fmt.Errorf("Python 解释器不可用: %s", pythonPath))
		} else {
			logInfo("Python 解释器检查通过", "path", pythonPath)
		}
	} else {
		logInfo("Python 解释器检查通过", "path", pythonPath)
	}

	// 3. 检查 Engine 脚本
	engineScript := cfg.Engine.Executable
	if engineScript == "" {
		engineScript = filepath.Join(cfg.Paths.EngineDir, "server.py")
	}
	if !fileExists(engineScript) {
		errs = append(errs, fmt.Errorf("Engine 脚本不存在: %s", engineScript))
	} else {
		logInfo("Engine 脚本检查通过", "path", engineScript)
	}

	// 4. 检查 Frontend 可执行文件
	frontendExe := cfg.Frontend.Executable
	if frontendExe == "" {
		frontendExe = findFrontendExe(cfg.Paths.FrontendDir)
	}
	if frontendExe == "" || !fileExists(frontendExe) {
		// Frontend 可执行文件不存在，尝试自动查找
		frontendExe = findFrontendExe(cfg.Paths.FrontendDir)
	}
	if frontendExe == "" {
		logWarn("Frontend 可执行文件未找到，将在启动时跳过", "dir", cfg.Paths.FrontendDir)
	} else {
		logInfo("Frontend 可执行文件检查通过", "path", frontendExe)
		cfg.Frontend.Executable = frontendExe // 更新为找到的路径
	}

	if len(errs) > 0 {
		return fmt.Errorf("依赖检查失败: %v", errs)
	}

	logInfo("运行环境依赖检查全部通过")
	return nil
}

// -----------------------------------------------------------------------------
// findFrontendExe - 在 Frontend 目录下查找 Tauri 可执行文件
// -----------------------------------------------------------------------------
// 查找顺序：
//   1. Frontend/src-tauri/target/release/ai-studio.exe
//   2. Frontend/src-tauri/target/debug/ai-studio.exe
//   3. Frontend/ai-studio.exe
// -----------------------------------------------------------------------------

func findFrontendExe(frontendDir string) string {
	candidates := []string{
		filepath.Join(frontendDir, "src-tauri", "target", "release", "ai-studio.exe"),
		filepath.Join(frontendDir, "src-tauri", "target", "debug", "ai-studio.exe"),
		filepath.Join(frontendDir, "ai-studio.exe"),
	}
	for _, p := range candidates {
		if fileExists(p) {
			return p
		}
	}
	return ""
}

// -----------------------------------------------------------------------------
// fileExists - 检查文件是否存在
// -----------------------------------------------------------------------------

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func findConfigFile(dir, name string) string {
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		candidate := filepath.Join(exeDir, dir, name)
		if fileExists(candidate) {
			return candidate
		}
		parentDir := filepath.Dir(exeDir)
		candidate = filepath.Join(parentDir, dir, name)
		if fileExists(candidate) {
			return candidate
		}
	}

	wd, err := os.Getwd()
	if err == nil {
		candidate := filepath.Join(wd, dir, name)
		if fileExists(candidate) {
			return candidate
		}
		parentDir := filepath.Dir(wd)
		candidate = filepath.Join(parentDir, dir, name)
		if fileExists(candidate) {
			return candidate
		}
	}

	candidate := filepath.Join(dir, name)
	if fileExists(candidate) {
		return candidate
	}

	return ""
}