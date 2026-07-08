// =============================================================================
// AIStudio Launcher - 程序入口
// =============================================================================
// 启动流程：
//   1. 加载配置 (Config/app.yaml)
//   2. 初始化日志 (Runtime/logs/launcher.log)
//   3. 检查目录结构
//   4. 检查运行依赖
//   5. 启动服务 (Backend → Engine → Frontend)
//   6. 启动健康监控
//   7. 等待退出信号
//   8. 优雅关闭
// =============================================================================

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// -----------------------------------------------------------------------------
// main - 程序入口
// -----------------------------------------------------------------------------

func main() {
	// 打印启动横幅
	printBanner()

	// 创建根上下文，用于控制整个生命周期
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听退出信号（Ctrl+C 或 SIGTERM）
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n[Launcher] 收到退出信号，开始关闭...")
		cancel()
	}()

	// ---- 1. 加载配置 ----
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("[Launcher] 加载配置失败: %v", err)
	}

	// 根据配置的日志级别决定是否开启 DEBUG
	if cfg.App.LogLevel == "debug" {
		SetDebug(true)
	}

	// ---- 2. 初始化日志系统 ----
	if err := InitLogger(cfg.Paths.LogsDir); err != nil {
		log.Fatalf("[Launcher] 初始化日志失败: %v", err)
	}
	logInfo("日志系统已初始化", "log_dir", cfg.Paths.LogsDir)
	logInfo("配置加载完成", "app", cfg.App.Name, "version", cfg.App.Version, "env", cfg.App.Environment)

	// ---- 3. 检查目录结构 ----
	if err := checkDirectories(cfg); err != nil {
		logError("目录检查失败", "error", err)
		log.Fatalf("[Launcher] 目录检查失败: %v", err)
	}
	logInfo("目录检查通过")

	// ---- 4. 检查运行依赖 ----
	if err := CheckDependencies(cfg); err != nil {
		logError("依赖检查失败", "error", err)
		log.Fatalf("[Launcher] 依赖检查失败: %v", err)
	}

	// ---- 5. 创建进程管理器和服务编排器 ----
	pm := NewProcessManager()
	svc := NewLauncherService(cfg, pm)

	// ---- 6. 启动所有服务 ----
	if err := svc.Start(ctx); err != nil {
		logError("服务启动失败", "error", err)
		pm.KillAll()
		log.Fatalf("[Launcher] 启动失败: %v", err)
	}

	// 打印运行状态
	fmt.Println()
	fmt.Println("[Launcher] AIStudio 正在运行...")
	fmt.Printf("[Launcher] Backend:  http://localhost:%s\n", cfg.Backend.Port)
	fmt.Printf("[Launcher] Engine:   http://localhost:%s\n", cfg.Engine.Port)
	fmt.Println("[Launcher] 按 Ctrl+C 停止")
	fmt.Println()

	// ---- 7. 启动健康监控（后台 goroutine）----
	go StartHealthMonitor(ctx, cfg, pm, svc)

	// ---- 8. 等待退出信号 ----
	<-ctx.Done()

	// ---- 9. 优雅关闭 ----
	logInfo("开始关闭所有服务...")
	svc.Stop()
	logInfo("关闭完成")
	CloseLogger()
}

// -----------------------------------------------------------------------------
// printBanner - 打印启动横幅
// -----------------------------------------------------------------------------

func printBanner() {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║        AIStudio Launcher v1.0.0          ║")
	fmt.Println("║        Unified Process Manager           ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()
}

// -----------------------------------------------------------------------------
// checkDirectories - 检查必要的目录是否存在，不存在则自动创建
// -----------------------------------------------------------------------------

func checkDirectories(cfg *Config) error {
	// 需要检查的目录列表
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
