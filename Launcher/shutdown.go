// =============================================================================
// AIStudio Launcher - 优雅关闭管理
// =============================================================================
// 功能：提供优雅关闭的辅助函数
//       正常流程由 main.go 直接调用 svc.Stop() 完成
//       本模块提供 GracefulShutdown 作为备用关闭路径
// 关闭顺序：Frontend → Engine → Backend → 释放资源
// =============================================================================

package main

import (
	"fmt"
	"time"
)

// -----------------------------------------------------------------------------
// GracefulShutdown - 优雅关闭所有服务
// -----------------------------------------------------------------------------
// 按 Frontend → Engine → Backend 的顺序停止每个服务
// 每个服务最多等待 timeout 时间退出，超时则强制终止
// 最后释放日志等资源
//
// 参数：
//   pm      - 进程管理器
//   timeout - 每个服务的停止超时时间
// 返回：
//   error - 关闭过程中的错误（多个错误合并返回）
// -----------------------------------------------------------------------------

func GracefulShutdown(pm *ProcessManager, timeout time.Duration) error {
	logInfo("开始优雅关闭流程...")

	// 关闭顺序：先关闭依赖方，再关闭被依赖方
	// Frontend 依赖 Backend 和 Engine，所以最先关闭
	// Backend 是核心服务，最后关闭
	order := []struct {
		name string
	}{
		{"frontend"},
		{"engine"},
		{"backend"},
	}

	var errs []error

	for _, svc := range order {
		if pm.Check(svc.name) {
			logInfo("正在停止服务", "service", svc.name)
			if err := pm.Stop(svc.name, timeout); err != nil {
				logWarn("停止服务失败", "service", svc.name, "error", err)
				errs = append(errs, fmt.Errorf("%s: %w", svc.name, err))
			}
		} else {
			logInfo("服务未在运行，跳过", "service", svc.name)
		}
	}

	// 确保所有残留进程都被终止
	pm.KillAll()

	logInfo("所有服务已停止")

	// 释放日志资源
	CloseLogger()

	if len(errs) > 0 {
		return fmt.Errorf("关闭过程中出现错误: %v", errs)
	}
	return nil
}
