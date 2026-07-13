// =============================================================================
// AIStudio Launcher - 日志系统
// =============================================================================
// 功能：统一管理 Launcher / Backend / Engine 的日志输出
// 输出：同时写入控制台(stdout) 和日志文件(Runtime/logs/launcher.log)
// =============================================================================

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// -----------------------------------------------------------------------------
// 全局日志状态
// -----------------------------------------------------------------------------

var (
	logFile     *os.File           // 日志文件句柄
	logMu       sync.Mutex         // 保护日志输出的互斥锁
	debugEnabled bool              // 是否启用 DEBUG 级别日志
)

// -----------------------------------------------------------------------------
// InitLogger - 初始化日志系统
// -----------------------------------------------------------------------------
// 参数：
//   logDir - 日志目录路径（如 Runtime/logs）
// 返回：
//   error - 初始化失败时返回错误
// -----------------------------------------------------------------------------

func InitLogger(logDir string) error {
	// 确保日志目录存在
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 打开（或创建）日志文件，追加模式
	logPath := filepath.Join(logDir, "launcher.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}

	logFile = f

	// 配置标准 log 包：同时输出到控制台和文件
	multiWriter := io.MultiWriter(os.Stdout, f)
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime)

	return nil
}

// -----------------------------------------------------------------------------
// SetDebug - 启用或禁用 DEBUG 日志
// -----------------------------------------------------------------------------

func SetDebug(enabled bool) {
	logMu.Lock()
	debugEnabled = enabled
	logMu.Unlock()
}

// -----------------------------------------------------------------------------
// 核心日志函数 - Info / Warn / Error / Debug
// -----------------------------------------------------------------------------
// 参数：
//   msg - 日志消息
//   kv  - 键值对（可变参数），如 logInfo("启动服务", "name", "backend", "port", "8081")
//         输出格式：[INFO] 2024-01-01 12:00:00 启动服务 name=backend port=8081
// -----------------------------------------------------------------------------

func logInfo(msg string, kv ...any) {
	logMu.Lock()
	defer logMu.Unlock()
	log.Printf("[INFO]  %s%s", msg, formatKV(kv...))
}

func logWarn(msg string, kv ...any) {
	logMu.Lock()
	defer logMu.Unlock()
	log.Printf("[WARN]  %s%s", msg, formatKV(kv...))
}

func logError(msg string, kv ...any) {
	logMu.Lock()
	defer logMu.Unlock()
	log.Printf("[ERROR] %s%s", msg, formatKV(kv...))
}

func logDebug(msg string, kv ...any) {
	logMu.Lock()
	defer logMu.Unlock()
	if !debugEnabled {
		return
	}
	log.Printf("[DEBUG] %s%s", msg, formatKV(kv...))
}

// -----------------------------------------------------------------------------
// formatKV - 将键值对格式化为 " key=value" 字符串
// -----------------------------------------------------------------------------

func formatKV(kv ...any) string {
	if len(kv) == 0 {
		return ""
	}
	var sb strings.Builder
	for i := 0; i+1 < len(kv); i += 2 {
		sb.WriteString(fmt.Sprintf(" %v=%v", kv[i], kv[i+1]))
	}
	return sb.String()
}

// -----------------------------------------------------------------------------
// CloseLogger - 关闭日志文件，释放资源
// -----------------------------------------------------------------------------

func CloseLogger() {
	logMu.Lock()
	defer logMu.Unlock()
	if logFile != nil {
		log.SetOutput(os.Stdout) // 先切换回 stdout，避免写入已关闭的文件
		logFile.Close()
		logFile = nil
	}
}
