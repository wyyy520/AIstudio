// =============================================================================
// AIStudio Launcher - 服务编排
// =============================================================================
// 功能：按顺序启动 Backend → Engine → Frontend，等待各服务健康后继续
// 关闭：按逆序停止 Frontend → Engine → Backend
// =============================================================================

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// -----------------------------------------------------------------------------
// LauncherService - 服务编排器
// -----------------------------------------------------------------------------

type LauncherService struct {
	cfg *Config          // 全局配置
	pm  *ProcessManager  // 进程管理器
}

// NewLauncherService - 创建服务编排器
func NewLauncherService(cfg *Config, pm *ProcessManager) *LauncherService {
	return &LauncherService{
		cfg: cfg,
		pm:  pm,
	}
}

// -----------------------------------------------------------------------------
// Start - 启动所有服务（按依赖顺序）
// -----------------------------------------------------------------------------
// 启动流程：
//   1. 启动 Backend，等待 /api/health 响应
//   2. 启动 Python Engine，等待 /health 响应
//   3. 启动 Tauri Frontend（无需健康检查，GUI 应用）
// -----------------------------------------------------------------------------

func (s *LauncherService) Start(ctx context.Context) error {
	logInfo("开始启动 AIStudio 服务...")

	// ---- 1. 启动 Backend ----
	if err := s.startBackend(ctx); err != nil {
		return fmt.Errorf("启动 Backend 失败: %w", err)
	}

	// 等待 Backend 健康检查通过
	backendURL := fmt.Sprintf("http://localhost:%s%s", s.cfg.Backend.Port, s.cfg.Backend.HealthPath)
	if err := s.waitForHealth(ctx, "Backend", backendURL); err != nil {
		return fmt.Errorf("Backend 健康检查失败: %w", err)
	}

	// ---- 2. 启动 Python Engine ----
	if err := s.startEngine(ctx); err != nil {
		return fmt.Errorf("启动 Engine 失败: %w", err)
	}

	// 等待 Engine 健康检查通过（Engine 健康检查失败时仅警告，不阻塞启动）
	engineURL := fmt.Sprintf("http://localhost:%s%s", s.cfg.Engine.Port, s.cfg.Engine.HealthPath)
	if err := s.waitForHealth(ctx, "Engine", engineURL); err != nil {
		logWarn("Engine 健康检查未通过，继续启动（Engine 可能需要手动检查）", "error", err)
	}

	// ---- 3. 启动 Frontend ----
	if err := s.startFrontend(ctx); err != nil {
		return fmt.Errorf("启动 Frontend 失败: %w", err)
	}

	return nil
}

// -----------------------------------------------------------------------------
// Stop - 停止所有服务（按逆序）
// -----------------------------------------------------------------------------

func (s *LauncherService) Stop() {
	logInfo("开始停止 AIStudio 服务...")

	shutdownTimeout := time.Duration(s.cfg.Launcher.ShutdownTimeout) * time.Second

	// 按逆序停止：Frontend → Vite → Engine → Backend
	services := []struct {
		name    string
		timeout time.Duration
	}{
		{"frontend", shutdownTimeout},
		{"vite", shutdownTimeout},
		{"engine", shutdownTimeout},
		{"backend", shutdownTimeout},
	}

	for _, svc := range services {
		if s.pm.Check(svc.name) {
			if err := s.pm.Stop(svc.name, svc.timeout); err != nil {
				logWarn("停止服务失败", "name", svc.name, "error", err)
			}
		}
	}

	logInfo("所有服务已停止")
}

// -----------------------------------------------------------------------------
// startBackend - 启动 Backend 服务
// -----------------------------------------------------------------------------

func (s *LauncherService) startBackend(ctx context.Context) error {
	logInfo("正在启动 Backend 服务...")

	// 启动前清理可能残留的占用端口的进程
	if pid := findProcessOnPort(s.cfg.Backend.Port); pid > 0 {
		logWarn("检测到端口被占用，清理残留进程", "port", s.cfg.Backend.Port, "pid", pid)
		killProcessTree(pid)
		time.Sleep(500 * time.Millisecond) // 等待端口释放
	}

	// 获取 Backend 可执行文件路径
	backendExe := s.cfg.Backend.Executable
	if backendExe == "" {
		backendExe = filepath.Join(s.cfg.Paths.BackendDir, "cmd.exe")
	}

	// 构造启动命令
	cmd := exec.CommandContext(ctx, backendExe)
	cmd.Dir = s.cfg.Paths.BackendDir // 工作目录设为 Backend/

	// 传递环境变量：配置文件路径
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("AISTUDIO_CONFIG=%s", filepath.Join(s.cfg.Paths.ConfigDir, "backend.yaml")),
	)

	// 日志文件
	backendLog := filepath.Join(s.cfg.Paths.LogsDir, "backend.log")

	// 启动进程
	if err := s.pm.Start("backend", cmd, backendLog); err != nil {
		return err
	}

	logInfo("Backend 服务已启动", "exe", backendExe, "port", s.cfg.Backend.Port)
	return nil
}

// -----------------------------------------------------------------------------
// startEngine - 启动 Python Engine 服务
// -----------------------------------------------------------------------------

func (s *LauncherService) startEngine(ctx context.Context) error {
	logInfo("正在启动 Python Engine 服务...")

	// 确定 Python 解释器路径
	pythonPath := s.cfg.Engine.PythonPath
	if pythonPath == "" {
		pythonPath = "python"
	}

	// 如果配置了虚拟环境，优先使用虚拟环境的 Python
	venvPython := filepath.Join(s.cfg.Engine.VenvPath, "Scripts", "python.exe")
	if s.cfg.Engine.VenvPath != "" && fileExists(venvPython) {
		pythonPath = venvPython
		logInfo("使用虚拟环境 Python", "path", venvPython)
	}

	// 获取 Engine 服务脚本路径
	engineScript := s.cfg.Engine.Executable
	if engineScript == "" {
		engineScript = filepath.Join(s.cfg.Paths.EngineDir, "server.py")
	}

	// 构造启动命令：python server.py --port <port>
	cmd := exec.CommandContext(ctx, pythonPath, engineScript,
		"--port", s.cfg.Engine.Port,
	)
	cmd.Dir = s.cfg.Paths.EngineDir // 工作目录设为 Engine/

	// 传递环境变量
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PYTHONPATH=%s", s.cfg.Paths.EngineDir),
		"PYTHONUNBUFFERED=1", // 禁用输出缓冲，确保日志实时写入
	)

	// 日志文件
	engineLog := filepath.Join(s.cfg.Paths.LogsDir, "engine.log")

	// 启动进程
	if err := s.pm.Start("engine", cmd, engineLog); err != nil {
		return err
	}

	logInfo("Engine 服务已启动", "python", pythonPath, "script", engineScript, "port", s.cfg.Engine.Port)
	return nil
}

// -----------------------------------------------------------------------------
// startFrontend - 启动 Tauri Frontend
// -----------------------------------------------------------------------------

func (s *LauncherService) startFrontend(ctx context.Context) error {
	logInfo("正在启动 Frontend 服务...")

	// 如果配置了 Vite dev server，先启动 Vite
	if s.cfg.Frontend.DevServer {
		if err := s.startViteDevServer(ctx); err != nil {
			logWarn("Vite dev server 启动失败，继续启动 Frontend", "error", err)
		}
	}

	// 获取 Frontend 可执行文件路径
	frontendExe := s.cfg.Frontend.Executable
	if frontendExe == "" {
		frontendExe = findFrontendExe(s.cfg.Paths.FrontendDir)
	}

	// 如果找不到可执行文件，跳过 Frontend 启动
	if frontendExe == "" || !fileExists(frontendExe) {
		logWarn("Frontend 可执行文件未找到，跳过 Frontend 启动", "dir", s.cfg.Paths.FrontendDir)
		return nil
	}

	// 构造启动命令
	cmd := exec.CommandContext(ctx, frontendExe)
	cmd.Dir = s.cfg.Paths.FrontendDir // 工作目录设为 Frontend 目录

	// 传递环境变量：Backend 地址
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("VITE_API_BASE_URL=%s", s.cfg.Frontend.BackendAddr),
	)

	// 日志文件
	frontendLog := filepath.Join(s.cfg.Paths.LogsDir, "frontend.log")

	// 启动进程
	if err := s.pm.Start("frontend", cmd, frontendLog); err != nil {
		return err
	}

	logInfo("Frontend 服务已启动", "exe", frontendExe, "backend_addr", s.cfg.Frontend.BackendAddr)
	return nil
}

// -----------------------------------------------------------------------------
// startViteDevServer - 启动 Vite dev server
// -----------------------------------------------------------------------------
// 在配置的 ViteDir 目录下执行 ViteCmd，等待 Vite 在 VitePort 上可用
// -----------------------------------------------------------------------------

func (s *LauncherService) startViteDevServer(ctx context.Context) error {
	vitePort := s.cfg.Frontend.VitePort
	if vitePort == "" {
		vitePort = "5173"
	}

	// 检查 Vite 是否已经在运行
	viteURL := fmt.Sprintf("http://localhost:%s", vitePort)
	if checkHTTP(viteURL) {
		logInfo("Vite dev server 已在运行，跳过启动", "url", viteURL)
		return nil
	}

	// 清理可能残留的占用 Vite 端口的进程
	if pid := findProcessOnPort(vitePort); pid > 0 {
		logWarn("检测到 Vite 端口被占用，清理残留进程", "port", vitePort, "pid", pid)
		killProcessTree(pid)
		time.Sleep(500 * time.Millisecond)
	}

	logInfo("正在启动 Vite dev server...", "port", vitePort)

	// 构造 Vite 启动命令
	viteDir := s.cfg.Frontend.ViteDir
	if viteDir == "" {
		viteDir = s.cfg.Paths.FrontendDir
	}

	cmd := exec.CommandContext(ctx, "cmd", "/c", s.cfg.Frontend.ViteCmd)
	cmd.Dir = viteDir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PORT=%s", vitePort),
	)

	// Vite 日志文件
	viteLog := filepath.Join(s.cfg.Paths.LogsDir, "vite.log")

	if err := s.pm.Start("vite", cmd, viteLog); err != nil {
		return fmt.Errorf("启动 Vite 失败: %w", err)
	}

	// 等待 Vite 启动（最多 15 秒）
	timeout := 15 * time.Second
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if checkHTTP(viteURL) {
				logInfo("Vite dev server 已就绪", "url", viteURL)
				return nil
			}
			if time.Now().After(deadline) {
				return fmt.Errorf("Vite dev server 启动超时（%v 内未响应）", timeout)
			}
		}
	}
}

// -----------------------------------------------------------------------------
// waitForHealth - 等待服务健康检查通过
// -----------------------------------------------------------------------------
// 参数：
//   ctx       - 上下文（用于取消）
//   name      - 服务名称（用于日志）
//   healthURL - 健康检查 URL（如 http://localhost:8081/api/health）
// 逻辑：
//   每 500ms 发送一次 HTTP GET 请求，成功则返回 nil
//   超过 startup_timeout 秒未成功则返回错误
// -----------------------------------------------------------------------------

func (s *LauncherService) waitForHealth(ctx context.Context, name string, healthURL string) error {
	logInfo("等待服务健康检查...", "service", name, "url", healthURL)

	timeout := time.Duration(s.cfg.Launcher.StartupTimeout) * time.Second
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if checkHTTP(healthURL) {
				logInfo("服务健康检查通过", "service", name)
				return nil
			}
			if time.Now().After(deadline) {
				return fmt.Errorf("%s 健康检查超时（%v 内未响应）", name, timeout)
			}
		}
	}
}

// -----------------------------------------------------------------------------
// RestartBackend - 重启 Backend 服务
// -----------------------------------------------------------------------------

func (s *LauncherService) RestartBackend(ctx context.Context) error {
	logInfo("正在重启 Backend...")

	backendExe := s.cfg.Backend.Executable
	if backendExe == "" {
		backendExe = filepath.Join(s.cfg.Paths.BackendDir, "cmd.exe")
	}

	cmd := exec.CommandContext(ctx, backendExe)
	cmd.Dir = s.cfg.Paths.BackendDir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("AISTUDIO_CONFIG=%s", filepath.Join(s.cfg.Paths.ConfigDir, "backend.yaml")),
	)

	backendLog := filepath.Join(s.cfg.Paths.LogsDir, "backend.log")
	timeout := time.Duration(s.cfg.Launcher.ShutdownTimeout) * time.Second

	return s.pm.Restart("backend", cmd, backendLog, timeout)
}

// -----------------------------------------------------------------------------
// RestartEngine - 重启 Engine 服务
// -----------------------------------------------------------------------------

func (s *LauncherService) RestartEngine(ctx context.Context) error {
	logInfo("正在重启 Engine...")

	pythonPath := s.cfg.Engine.PythonPath
	if pythonPath == "" {
		pythonPath = "python"
	}
	venvPython := filepath.Join(s.cfg.Engine.VenvPath, "Scripts", "python.exe")
	if s.cfg.Engine.VenvPath != "" && fileExists(venvPython) {
		pythonPath = venvPython
	}

	engineScript := s.cfg.Engine.Executable
	if engineScript == "" {
		engineScript = filepath.Join(s.cfg.Paths.EngineDir, "server.py")
	}

	cmd := exec.CommandContext(ctx, pythonPath, engineScript, "--port", s.cfg.Engine.Port)
	cmd.Dir = s.cfg.Paths.EngineDir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PYTHONPATH=%s", s.cfg.Paths.EngineDir),
		"PYTHONUNBUFFERED=1",
	)

	engineLog := filepath.Join(s.cfg.Paths.LogsDir, "engine.log")
	timeout := time.Duration(s.cfg.Launcher.ShutdownTimeout) * time.Second

	return s.pm.Restart("engine", cmd, engineLog, timeout)
}

// -----------------------------------------------------------------------------
// checkHTTP - 发送 HTTP GET 请求检查服务健康
// -----------------------------------------------------------------------------
// 返回 true 表示服务健康（HTTP 200），false 表示不可用
// -----------------------------------------------------------------------------

func checkHTTP(url string) bool {
	client := &http.Client{
		Timeout: 3 * time.Second, // 健康检查请求超时
	}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
