package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var (
	flagDev        bool
	flagNoFrontend bool
	flagDebug      bool
	flagConfig     string
	flagVersion    bool
)

func init() {
	flag.BoolVar(&flagDev, "dev", false, "开发模式：启动 Vite dev server")
	flag.BoolVar(&flagNoFrontend, "no-frontend", false, "不启动 Frontend")
	flag.BoolVar(&flagDebug, "debug", false, "启用 DEBUG 日志")
	flag.StringVar(&flagConfig, "config", "", "配置文件路径 (默认: Config/app.yaml)")
	flag.BoolVar(&flagVersion, "version", false, "显示版本信息")
}

func main() {
	flag.Parse()

	if flagVersion {
		fmt.Println("AIStudio Launcher v1.0.0")
		return
	}

	chdirToExeDir()
	printBanner()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n[Launcher] 收到退出信号，开始关闭...")
		cancel()
	}()

	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n[Launcher] 加载配置失败: %v\n", err)
		waitExit()
		return
	}

	if flagConfig != "" {
		os.Setenv("AISTUDIO_CONFIG", flagConfig)
		cfg2, err2 := LoadConfig()
		if err2 != nil {
			fmt.Fprintf(os.Stderr, "\n[Launcher] 加载指定配置失败: %v\n", err2)
			waitExit()
			return
		}
		cfg = cfg2
	}

	if flagDebug || cfg.App.LogLevel == "debug" {
		SetDebug(true)
	}

	if flagDev {
		cfg.Frontend.DevServer = true
		if cfg.App.Environment == "" {
			cfg.App.Environment = "development"
		}
		logInfo("已启用开发模式（--dev）")
	}

	if err := InitLogger(cfg.Paths.LogsDir); err != nil {
		fmt.Fprintf(os.Stderr, "\n[Launcher] 初始化日志失败: %v\n", err)
		waitExit()
		return
	}
	logInfo("日志系统已初始化", "log_dir", cfg.Paths.LogsDir)
	logInfo("配置加载完成", "app", cfg.App.Name, "version", cfg.App.Version, "env", cfg.App.Environment)

	if err := checkDirectories(cfg); err != nil {
		logError("目录检查失败", "error", err)
		fmt.Fprintf(os.Stderr, "\n[Launcher] 目录检查失败: %v\n", err)
		waitExit()
		return
	}
	logInfo("目录检查通过")

	if err := CheckDependencies(cfg); err != nil {
		logError("依赖检查失败", "error", err)
		fmt.Fprintf(os.Stderr, "\n[Launcher] 依赖检查失败: %v\n", err)
		waitExit()
		return
	}

	pm := NewProcessManager()
	svc := NewLauncherService(cfg, pm)

	if flagNoFrontend {
		logInfo("跳过 Frontend 启动（--no-frontend）")
	}

	if err := svc.Start(ctx, flagNoFrontend); err != nil {
		logError("服务启动失败", "error", err)
		pm.KillAll()
		fmt.Fprintf(os.Stderr, "\n[Launcher] 启动失败: %v\n", err)
		waitExit()
		return
	}

	fmt.Println()
	fmt.Println("[Launcher] AIStudio 正在运行...")
	displayBackendHost := cfg.Backend.Host
	if displayBackendHost == "0.0.0.0" {
		displayBackendHost = "127.0.0.1"
	}
	displayEngineHost := cfg.Engine.Host
	if displayEngineHost == "0.0.0.0" {
		displayEngineHost = "127.0.0.1"
	}
	fmt.Printf("[Launcher] Backend:  http://%s:%s\n", displayBackendHost, cfg.Backend.Port)
	fmt.Printf("[Launcher] Engine:   http://%s:%s\n", displayEngineHost, cfg.Engine.Port)
	if cfg.Frontend.DevServer {
		fmt.Printf("[Launcher] Vite:     http://localhost:%s\n", cfg.Frontend.VitePort)
	}
	fmt.Println("[Launcher] 按 Ctrl+C 停止")
	fmt.Println()

	go StartHealthMonitor(ctx, cfg, pm, svc)

	<-ctx.Done()

	logInfo("开始关闭所有服务...")
	svc.Stop()
	logInfo("关闭完成")
	CloseLogger()
}

func chdirToExeDir() {
	exePath, err := os.Executable()
	if err != nil {
		return
	}
	exeDir := filepath.Dir(exePath)

	configInExeDir := filepath.Join(exeDir, "Config", "app.yaml")
	configInParentDir := filepath.Join(filepath.Dir(exeDir), "Config", "app.yaml")

	if _, err := os.Stat(configInExeDir); err == nil {
		os.Chdir(exeDir)
	} else if _, err := os.Stat(configInParentDir); err == nil {
		os.Chdir(filepath.Dir(exeDir))
	} else {
		os.Chdir(exeDir)
	}
}

func waitExit() {
	fmt.Println()
	fmt.Println("按回车键退出...")
	var b [1]byte
	os.Stdin.Read(b[:])
}

func printBanner() {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║        AIStudio Launcher v1.0.0          ║")
	fmt.Println("║        Unified Process Manager           ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()
}

func checkDirectories(cfg *Config) error {
	dirs := []string{
		cfg.Paths.Root,
		cfg.Paths.BackendDir,
		cfg.Paths.FrontendDir,
		cfg.Paths.EngineDir,
		cfg.Paths.StorageDir,
		cfg.Paths.ConfigDir,
		cfg.Paths.LogsDir,
	}

	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		if info, err := os.Stat(dir); os.IsNotExist(err) {
			logWarn("目录不存在，自动创建", "dir", dir)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("创建目录失败 %s: %w", dir, err)
			}
		} else if err != nil {
			return fmt.Errorf("检查目录失败 %s: %w", dir, err)
		} else if !info.IsDir() {
			return fmt.Errorf("路径不是目录: %s", dir)
		}
	}

	return nil
}