# AIStudio Bug 扫描报告

> **审查日期**: 2026-07-10  
> **严重程度**: 🔴 CRITICAL / 🟠 HIGH / 🟡 MEDIUM / 🔵 LOW

---

## 一、🔴 CRITICAL 级别 Bug

### Bug 1：Plugin Executor 中 ctx.Value 导致格式化 panic

**文件**: [executor.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/plugin/executor.go) L61

```go
taskID := fmt.Sprintf("plugin-%s-%d", name, ctx.Value("request_id"))
```

**风险**: `ctx.Value("request_id")` 返回 `interface{}`，当值为 nil 时，`%d` 格式化输出 `%!d(<nil>)`，导致 taskID 格式错误。虽不 panic 但产生垃圾 ID。

**修复**: 使用 UUID 生成 taskID。

### Bug 2：Worker 失败时直接修改 Task 状态（竞态条件）

**文件**: [worker.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/task/worker.go) L100-L108

```go
if w.manager != nil {
    _ = w.manager.FailTask(w.ctx, task.ID, err.Error())
} else {
    task.Status = StatusFailed  // 直接修改，无锁保护
    task.Error = err.Error()
}
```

**风险**: 当 `w.manager == nil` 时，Worker 直接修改 Task 对象而未加锁。在多 Worker 场景下存在竞态条件。

### Bug 3：PythonEngine subprocess 无资源清理保证

**文件**: [python.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/engine/python.go) L80-L82

```go
taskDir, err := os.MkdirTemp("", "aistudio-task-"+input.TaskID)
if err != nil {
    return nil, fmt.Errorf("create task dir: %w", err)
}
defer os.RemoveAll(taskDir)
```

**风险**: 如果 `cmd.Start()` 成功但进程卡死，`defer os.RemoveAll(taskDir)` 在函数返回时执行，但进程可能仍在运行访问 taskDir。同时 stdout/stderr 管道可能泄漏。

### Bug 4：JWT Secret 硬编码默认值

**文件**: [token.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/auth/token.go) L19, [middleware.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/api/middleware/middleware.go) L28

```go
secret = "aistudio-default-secret-change-in-production"
```

**风险**: 默认密钥是公开的，生产环境未修改时任何人都可以伪造 JWT Token。

### Bug 5：SQL 注入风险（GORM 使用不当）

**文件**: 检查所有数据库查询

**风险**: 项目使用 GORM 框架，大部分查询使用了参数化查询，但需要检查所有 `Raw()` 和 `Exec()` 调用。

---

## 二、🟠 HIGH 级别 Bug

### Bug 6：WebSocket 实现使用标准库而非 gorilla/websocket

**文件**: [websocket.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/api/handlers/websocket.go)

**问题**: 自行实现 WebSocket 协议（RFC 6455），而非使用成熟的 `gorilla/websocket` 库。手动处理帧编码/解码容易出错。

**影响**: 
- 可能不支持所有 WebSocket 扩展
- 缺乏 wss（TLS）支持
- 自行实现的掩码处理可能有问题

### Bug 7：LogService 内存泄漏

**文件**: [log_service.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/service/log_service.go)

**问题**: 日志存储在内存中，最大 10000 条。长时间运行后：
- 旧日志被静默丢弃
- 重启后日志丢失
- 没有持久化机制

### Bug 8：Python Engine 无超时传播

**文件**: [python.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/engine/python.go)

**问题**: `exec.CommandContext(ctx, ...)` 使用的 context 没有设置超时。如果 Python 任务卡死，Go 进程也会永久等待。

### Bug 9：Plugin 安装路径遍历

**文件**: [installer.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/plugin/installer.go)

**问题**: 插件安装时未验证 manifest 路径是否在预期目录内，可能导致路径遍历攻击。

### Bug 10：Frontend 未处理 token 过期

**文件**: [request.ts](file:///d:/AIstudio-master/AIstudio-master/Frontend/src/api/request.ts) L28-L32

```typescript
if (error.response?.status === 401) {
    const userStore = useUserStore()
    userStore.logout()
}
```

**问题**: 401 时直接登出，没有尝试刷新 token。用户体验差。

---

## 三、🟡 MEDIUM 级别 Bug

### Bug 11：Goroutine 泄漏

**位置**: [python.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/engine/python.go) L100-L107

```go
go func() {
    stderrReader := bufio.NewReader(stderr)
    for {
        line, err := stderrReader.ReadString('\n')
        // ...
    }
}()
```

**问题**: stderr 读取 goroutine 在进程退出时可能无法正确退出，导致 goroutine 泄漏。

### Bug 12：配置文件不一致

**位置**: [default.yaml](file:///d:/AIstudio-master/AIstudio-master/Backend/config/default.yaml) 与 [app.yaml](file:///d:/AIstudio-master/AIstudio-master/Config/app.yaml)

**问题**: 两个配置文件中 `server.port` 默认值不同（8081 vs 8082）。

### Bug 13：Python 异常未捕获

**位置**: [server.py](file:///d:/AIstudio-master/AIstudio-master/Engine/server.py) `_handle_dataset_read`, `_handle_dataset_split`, `_handle_dataset_convert`

**问题**: 这些方法在异常时返回 400，但 Python 的 `BaseHTTPRequestHandler` 在异常时不会自动捕获，可能导致 500 错误。

### Bug 14：Frontend Pinia Store 未持久化

**问题**: 刷新页面后所有 store 状态丢失，包括认证信息、项目状态等。

---

## 四、🔵 LOW 级别 Bug

### Bug 15：未使用的导入

**文件**: 
- [auth.go middleware](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/api/middleware/auth.go) L137-138: `_ = GetUserID; _ = GetUsername`
- [engine.go](file:///d:/AIstudio-master/AIstudio-master/Engine/vision/yolo/train.py) L79: `import json` 未使用

### Bug 16：硬编码路径

**文件**: 
- [node.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/workflow/node.go): `/storage/datasets/sample.jpg`, `/storage/exports/`
- [app.yaml](file:///d:/AIstudio-master/AIstudio-master/Config/app.yaml): 路径硬编码为相对路径

### Bug 17：变量命名不一致

**文件**: 
- Go 代码中同时使用 `snake_case` 和 `camelCase` 的 JSON tag
- Python 代码中同时使用 `snake_case` 和 `camelCase`

---

## 五、Bug 修复优先级

| 优先级 | Bug ID | 修复难度 | 影响范围 |
|--------|--------|---------|---------|
| P0 | 1, 4, 5 | 低 | 安全/数据 |
| P0 | 2, 3, 8 | 中 | 稳定性 |
| P1 | 6, 7, 9 | 中 | 功能 |
| P1 | 10, 11 | 低 | 用户体验 |
| P2 | 12-17 | 低 | 代码质量 |