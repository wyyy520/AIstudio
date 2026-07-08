// =============================================================================
// AIStudio Launcher - 健康检查与自动恢复
// =============================================================================
// 功能：定期检查 Backend 和 Engine 的健康状态
//       如果服务不可用且开启了 auto_restart，则自动重启
// 检查方式：
//   1. 进程存活检查（ProcessManager.Check）
//   2. HTTP 健康端点检查（GET /health 或 /api/health）
// 自动恢复：
//   - 服务故障时自动重启，最多重试 max_restart_attempts 次
//   - 重启成功后重置计数器
// =============================================================================

package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// -----------------------------------------------------------------------------
// HealthMonitor - 健康监控器
// -----------------------------------------------------------------------------

type HealthMonitor struct {
	cfg           *Config          // 全局配置
	pm            *ProcessManager  // 进程管理器
	svc           *LauncherService // 服务编排器（用于重启服务）
	mu            sync.Mutex       // 保护 restartCounts
	restartCounts map[string]int   // 各服务累计重启次数
}

// -----------------------------------------------------------------------------
// StartHealthMonitor - 启动健康监控（阻塞式，应在单独的 goroutine 中调用）
// -----------------------------------------------------------------------------
// 参数：
//   ctx - 上下文（ctx 取消时停止监控）
//   cfg - 全局配置
//   pm  - 进程管理器
//   svc - 服务编排器（用于自动重启）
// -----------------------------------------------------------------------------

func StartHealthMonitor(ctx context.Context, cfg *Config, pm *ProcessManager, svc *LauncherService) {
	hm := &HealthMonitor{
		cfg:           cfg,
		pm:            pm,
		svc:           svc,
		restartCounts: make(map[string]int),
	}

	// 健康检查间隔（最少 2 秒，默认 5 秒）
	interval := time.Duration(cfg.Launcher.HealthCheckInterval) * time.Second
	if interval < 2*time.Second {
		interval = 5 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logInfo("健康监控已启动", "interval", interval)

	for {
		select {
		case <-ctx.Done():
			logInfo("健康监控已停止")
			return
		case <-ticker.C:
			hm.checkAll(ctx)
		}
	}
}

// -----------------------------------------------------------------------------
// checkAll - 执行一次完整的健康检查（Backend + Engine）
// -----------------------------------------------------------------------------

func (hm *HealthMonitor) checkAll(ctx context.Context) {
	hm.checkBackend(ctx)
	hm.checkEngine(ctx)
}

// -----------------------------------------------------------------------------
// checkBackend - 检查 Backend 服务健康状态
// -----------------------------------------------------------------------------

func (hm *HealthMonitor) checkBackend(ctx context.Context) {
	serviceName := "backend"
	healthURL := fmt.Sprintf("http://localhost:%s%s", hm.cfg.Backend.Port, hm.cfg.Backend.HealthPath)

	// 1. 检查进程是否存活
	if !hm.pm.Check(serviceName) {
		logWarn("Backend 进程未运行")
		hm.handleServiceFailure(ctx, serviceName, "进程未运行")
		return
	}

	// 2. 检查 HTTP 健康端点
	if !checkHTTP(healthURL) {
		logWarn("Backend 健康检查失败", "url", healthURL)
		hm.handleServiceFailure(ctx, serviceName, "健康检查无响应")
		return
	}

	// 健康检查通过，重置重启计数
	hm.resetRestartCount(serviceName)
	logDebug("Backend 健康检查通过")
}

// -----------------------------------------------------------------------------
// checkEngine - 检查 Engine 服务健康状态
// -----------------------------------------------------------------------------

func (hm *HealthMonitor) checkEngine(ctx context.Context) {
	serviceName := "engine"
	healthURL := fmt.Sprintf("http://localhost:%s%s", hm.cfg.Engine.Port, hm.cfg.Engine.HealthPath)

	// 1. 检查进程是否存活
	if !hm.pm.Check(serviceName) {
		logWarn("Engine 进程未运行")
		hm.handleServiceFailure(ctx, serviceName, "进程未运行")
		return
	}

	// 2. 检查 HTTP 健康端点
	if !checkHTTP(healthURL) {
		logWarn("Engine 健康检查失败", "url", healthURL)
		hm.handleServiceFailure(ctx, serviceName, "健康检查无响应")
		return
	}

	// 健康检查通过，重置重启计数
	hm.resetRestartCount(serviceName)
	logDebug("Engine 健康检查通过")
}

// -----------------------------------------------------------------------------
// handleServiceFailure - 处理服务故障（自动重启逻辑）
// -----------------------------------------------------------------------------
// 参数：
//   ctx         - 上下文
//   serviceName - 服务名称（backend / engine）
//   reason      - 故障原因描述
// 逻辑：
//   1. 如果未开启 auto_restart，仅记录日志
//   2. 如果已超过 max_restart_attempts，停止重试并记录错误
//   3. 否则尝试重启服务，递增重启计数
//   4. 重启成功后重置计数器
// -----------------------------------------------------------------------------

func (hm *HealthMonitor) handleServiceFailure(ctx context.Context, serviceName string, reason string) {
	// 检查是否开启自动重启
	if !hm.cfg.Launcher.AutoRestart {
		logWarn("服务故障但未开启自动重启", "service", serviceName, "reason", reason)
		return
	}

	hm.mu.Lock()
	count := hm.restartCounts[serviceName]
	maxRestarts := hm.cfg.Launcher.MaxRestartAttempts
	if maxRestarts <= 0 {
		maxRestarts = 3 // 默认最多重启 3 次
	}

	// 超过最大重启次数，停止重试
	if count >= maxRestarts {
		hm.mu.Unlock()
		logError("服务重启次数已达上限，停止重试", "service", serviceName, "attempts", count, "max", maxRestarts)
		return
	}

	// 递增重启计数
	hm.restartCounts[serviceName] = count + 1
	currentAttempt := hm.restartCounts[serviceName]
	hm.mu.Unlock()

	logWarn("尝试自动重启服务", "service", serviceName, "reason", reason, "attempt", currentAttempt, "max", maxRestarts)

	// 执行重启
	var restartErr error
	switch serviceName {
	case "backend":
		restartErr = hm.svc.RestartBackend(ctx)
	case "engine":
		restartErr = hm.svc.RestartEngine(ctx)
	default:
		logWarn("未知服务，无法重启", "service", serviceName)
		return
	}

	if restartErr != nil {
		logError("服务重启失败", "service", serviceName, "attempt", currentAttempt, "error", restartErr)
	} else {
		logInfo("服务重启成功", "service", serviceName, "attempt", currentAttempt)
		// 重启成功后等待健康检查通过再重置计数
		// 计数器会在下次健康检查通过时自动重置
	}
}

// -----------------------------------------------------------------------------
// resetRestartCount - 重置指定服务的重启计数
// -----------------------------------------------------------------------------

func (hm *HealthMonitor) resetRestartCount(serviceName string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	if hm.restartCounts[serviceName] > 0 {
		hm.restartCounts[serviceName] = 0
	}
}
