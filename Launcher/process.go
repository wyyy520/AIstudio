// =============================================================================
// AIStudio Launcher - 进程管理器
// =============================================================================
// 功能：统一管理 Backend / Engine / Frontend 三个子进程的生命周期
// 支持：启动、停止（优雅+强制）、重启、存活检查、批量终止
// =============================================================================

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// -----------------------------------------------------------------------------
// Process - 单个被管理的子进程
// -----------------------------------------------------------------------------

type Process struct {
	Name   string       // 进程名称（backend / engine / frontend）
	Cmd    *exec.Cmd    // exec.Cmd 句柄
	PID    int          // 进程 ID
	Status string       // 状态：running / stopping / stopped / crashed
	done   chan struct{} // 进程退出信号通道（Wait 返回后关闭）
}

// -----------------------------------------------------------------------------
// ProcessManager - 进程管理器
// -----------------------------------------------------------------------------

type ProcessManager struct {
	mu        sync.Mutex              // 保护 processes map
	processes map[string]*Process     // 进程名称 -> Process
}

// NewProcessManager - 创建进程管理器
func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		processes: make(map[string]*Process),
	}
}

// -----------------------------------------------------------------------------
// Start - 启动一个子进程
// -----------------------------------------------------------------------------
// 参数：
//   name    - 进程名称（唯一标识，如 "backend"）
//   cmd     - 已配置好的 exec.Cmd（路径、参数、环境变量、工作目录）
//   logFile - 日志文件路径（子进程的 stdout/stderr 会重定向到此文件）
// 返回：
//   error - 启动失败时返回错误
// -----------------------------------------------------------------------------

func (pm *ProcessManager) Start(name string, cmd *exec.Cmd, logFile string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 检查是否已有同名进程在运行
	if p, exists := pm.processes[name]; exists {
		if p.Status == "running" {
			return fmt.Errorf("进程 %s 已在运行 (PID=%d)", name, p.PID)
		}
		// 清理已退出的旧记录
		delete(pm.processes, name)
	}

	// 配置日志输出：将子进程的 stdout/stderr 重定向到日志文件
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("打开日志文件失败 [%s]: %w", logFile, err)
		}
		cmd.Stdout = f
		cmd.Stderr = f
		// 注意：文件句柄在进程退出后由 GC 自动关闭，这里不 defer Close
	}

	// 启动进程
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动进程 %s 失败: %w", name, err)
	}

	// 创建 Process 记录
	p := &Process{
		Name:   name,
		Cmd:    cmd,
		PID:    cmd.Process.Pid,
		Status: "running",
		done:   make(chan struct{}),
	}
	pm.processes[name] = p

	// 启动后台 goroutine 等待进程退出
	// 当 cmd.Wait() 返回时，关闭 done 通道，通知其他 goroutine 进程已退出
	go func() {
		err := cmd.Wait()
		close(p.done)
		pm.mu.Lock()
		if proc, exists := pm.processes[name]; exists && proc == p {
			if err != nil {
				p.Status = "crashed"
				logWarn("进程异常退出", "name", name, "pid", p.PID, "error", err)
			} else {
				p.Status = "stopped"
				logInfo("进程正常退出", "name", name, "pid", p.PID)
			}
		}
		pm.mu.Unlock()
	}()

	logInfo("进程已启动", "name", name, "pid", p.PID)
	return nil
}

// -----------------------------------------------------------------------------
// Stop - 停止指定进程
// -----------------------------------------------------------------------------
// 停止流程：
//   1. 标记状态为 stopping
//   2. 发送 Kill 信号（Windows 上等同于 TerminateProcess）
//   3. 等待进程退出（最长 timeout 时间）
//   4. 超时则强制终止整个进程树（taskkill /F /T /PID）
//   5. 从管理器中移除进程记录
// -----------------------------------------------------------------------------

func (pm *ProcessManager) Stop(name string, timeout time.Duration) error {
	pm.mu.Lock()
	p, exists := pm.processes[name]
	pm.mu.Unlock()

	if !exists {
		return fmt.Errorf("进程 %s 不存在", name)
	}

	if p.Cmd.Process == nil {
		delete(pm.processes, name)
		return fmt.Errorf("进程 %s 没有进程句柄", name)
	}

	logInfo("正在停止进程", "name", name, "pid", p.PID)
	p.Status = "stopping"

	// 先检查进程是否已退出（避免对已退出进程调用 Kill 导致 invalid argument）
	select {
	case <-p.done:
		// 进程已经退出，无需 Kill
		logInfo("进程已自行退出", "name", name)
		pm.mu.Lock()
		delete(pm.processes, name)
		pm.mu.Unlock()
		return nil
	default:
	}

	// 进程仍在运行，发送 Kill 信号终止
	if err := p.Cmd.Process.Kill(); err != nil {
		logWarn("Kill 信号发送失败，尝试强制终止", "name", name, "error", err)
		// 即使 Kill 失败也尝试 taskkill 强制终止
		killProcessTree(p.PID)
	}

	// 等待进程退出或超时
	select {
	case <-p.done:
		// 进程已退出
		logInfo("进程已停止", "name", name)
	case <-time.After(timeout):
		// 超时，强制终止整个进程树
		logWarn("进程停止超时，强制终止进程树", "name", name, "pid", p.PID)
		killProcessTree(p.PID)
		<-p.done // 等待 done 通道关闭
	}

	// 从管理器中移除
	pm.mu.Lock()
	delete(pm.processes, name)
	pm.mu.Unlock()

	return nil
}

// -----------------------------------------------------------------------------
// Restart - 重启进程
// -----------------------------------------------------------------------------
// 参数：
//   name    - 进程名称
//   newCmd  - 新的 exec.Cmd（重启时需要重新构造命令）
//   logFile - 日志文件路径
//   timeout - 停止超时时间
// -----------------------------------------------------------------------------

func (pm *ProcessManager) Restart(name string, newCmd *exec.Cmd, logFile string, timeout time.Duration) error {
	logInfo("正在重启进程", "name", name)

	// 先停止旧进程（忽略错误，因为可能已经退出）
	if _, exists := pm.processes[name]; exists {
		if err := pm.Stop(name, timeout); err != nil {
			logWarn("停止旧进程失败，继续启动新进程", "name", name, "error", err)
		}
	}

	// 启动新进程
	return pm.Start(name, newCmd, logFile)
}

// -----------------------------------------------------------------------------
// Check - 检查进程是否存活
// -----------------------------------------------------------------------------

func (pm *ProcessManager) Check(name string) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p, exists := pm.processes[name]
	if !exists {
		return false
	}

	// 通过 done 通道检查进程是否已退出
	select {
	case <-p.done:
		return false
	default:
		return p.Status == "running"
	}
}

// -----------------------------------------------------------------------------
// GetPID - 获取进程 PID
// -----------------------------------------------------------------------------

func (pm *ProcessManager) GetPID(name string) (int, bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p, exists := pm.processes[name]
	if !exists {
		return 0, false
	}
	return p.PID, true
}

// -----------------------------------------------------------------------------
// KillAll - 终止所有进程（紧急关闭时使用）
// -----------------------------------------------------------------------------

func (pm *ProcessManager) KillAll() {
	pm.mu.Lock()
	names := make([]string, 0, len(pm.processes))
	for name := range pm.processes {
		names = append(names, name)
	}
	pm.mu.Unlock()

	for _, name := range names {
		pm.Stop(name, 3*time.Second)
	}
}

// -----------------------------------------------------------------------------
// List - 列出所有进程的名称和 PID
// -----------------------------------------------------------------------------

func (pm *ProcessManager) List() map[string]int {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	result := make(map[string]int)
	for name, p := range pm.processes {
		result[name] = p.PID
	}
	return result
}

// -----------------------------------------------------------------------------
// killProcessTree - 强制终止进程及其所有子进程
// -----------------------------------------------------------------------------
// Windows 实现：使用 taskkill /F /T /PID <pid>
//   /F - 强制终止
//   /T - 终止指定进程及其子进程
// -----------------------------------------------------------------------------

func killProcessTree(pid int) {
	// 构造 taskkill 命令
	cmd := exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", pid))
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		logWarn("taskkill 执行失败", "pid", pid, "error", err)
	}
}

// -----------------------------------------------------------------------------
// findProcessOnPort - 查找占用指定端口的进程 PID
// -----------------------------------------------------------------------------
// Windows 实现：使用 netstat -ano 查找 LISTENING 状态的端口占用
// 返回 PID（0 表示端口空闲）
// -----------------------------------------------------------------------------

func findProcessOnPort(port string) int {
	// 使用 netstat -ano 查找端口占用
	cmd := exec.Command("cmd", "/c", fmt.Sprintf("netstat -ano | findstr :%s | findstr LISTENING", port))
	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		return 0
	}

	// 解析 netstat 输出，提取 PID
	// 输出格式示例：  TCP    127.0.0.1:8081         0.0.0.0:0              LISTENING       12345
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			// 最后一个字段是 PID
			pid, err := strconv.Atoi(fields[len(fields)-1])
			if err == nil && pid > 0 {
				return pid
			}
		}
	}
	return 0
}
