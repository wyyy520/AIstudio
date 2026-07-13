# AIStudio 可维护性审查报告

> **审查日期**: 2026-07-10  
> **审查范围**: 代码质量、重复代码、命名规范、废弃代码、TODO/FIXME、配置系统、日志系统

---

## 一、代码质量评分

| 维度 | 评分 | 说明 |
|------|------|------|
| 代码重复率 | ⭐⭐⭐ | 少量重复代码 |
| 命名规范 | ⭐⭐⭐⭐ | 总体良好，Go 和 Python 风格一致 |
| 函数长度 | ⭐⭐⭐ | 部分函数超过 100 行 |
| 文件长度 | ⭐⭐⭐ | 部分文件超过 300 行 |
| 注释质量 | ⭐⭐⭐ | 中文注释，但部分过时 |
| 错误处理 | ⭐⭐⭐⭐ | 错误处理较完善 |
| 测试覆盖 | ⭐⭐ | 测试文件较少 |

---

## 二、代码重复问题

### 2.1 重复的 Store 目录

**位置**: 
- `Frontend/src/store/` (7 个文件)
- `Frontend/src/stores/` (6 个文件)

**重复文件**:
```
store/project.ts  ↔  stores/project.ts
store/settings.ts ↔  stores/settings.ts
store/workflow.ts ↔  stores/workflow.ts
```

**建议**: 删除 `store/` 目录，保留 `stores/` 目录。

### 2.2 重复的 API 路由

**位置**: [router.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/api/router.go)

**重复路由**:
```
POST /api/plugin/install     (旧路径)
POST /api/plugins/install    (新路径)
GET  /api/plugin/:id         (旧路径)
GET  /api/plugins/:name      (新路径)
```

**建议**: 统一使用 `/api/plugins/*` 路径，添加重定向。

### 2.3 重复的配置系统

**位置**: 
- `Backend/config/` (Go Viper 配置)
- `Config/` (Launcher 配置)

**重复配置字段**:
```yaml
# Backend/config/default.yaml
server.port: 8081
engine.python_path: python
engine.engine_dir: Engine

# Config/app.yaml
backend.port: 8081
engine.python_path: python
engine.engine_dir: Engine
```

---

## 三、代码异味

### 3.1 超长函数

#### `main()` - 234 行

**文件**: [main.go](file:///d:/AIstudio-master/AIstudio-master/Backend/cmd/main.go)

**问题**: 整个启动逻辑在 `main()` 函数中，包含配置加载、数据库初始化、Task Manager、Plugin Manager、Workflow Engine、Environment Manager、Python Engine、Agent、Services、HTTP Server 等所有初始化逻辑。

**建议**: 拆分为 `initXxx()` 函数。

#### `PythonRunner.Run()` - 约 150 行

**文件**: [python.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/engine/python.go)

**问题**: 包含文件写入、进程创建、stdout 解析、stderr 读取、结果处理等所有逻辑。

**建议**: 拆分为 `writeTaskFile()`, `startProcess()`, `parseOutput()`, `handleResult()` 等。

### 3.2 超长文件

| 文件 | 行数 | 问题 |
|------|------|------|
| [python.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/engine/python.go) | 348 | 包含 Runner、TaskHandler、Logger 三个类 |
| [server.py](file:///d:/AIstudio-master/AIstudio-master/Engine/server.py) | 400+ | 所有路由处理在一个类中 |
| [manager.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/plugin/manager.go) | 200+ | 包含 Manager 所有方法 |

### 3.3 Magic Number

| 位置 | 值 | 建议 |
|------|-----|------|
| [service.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/service/service.go) L25 | `24*60*60*1e9` | 定义为常量 `AccessTokenTTL` |
| [log_service.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/service/log_service.go) L22 | `10000` | 定义为常量 `MaxLogEntries` |
| [websocket.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/api/handlers/websocket.go) L130 | `30 * time.Second` | 定义为常量 `PingInterval` |

### 3.4 废弃代码

| 位置 | 废弃内容 | 说明 |
|------|---------|------|
| [auth.go middleware](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/api/middleware/auth.go) L137-138 | `_ = GetUserID; _ = GetUsername` | 用于抑制未使用警告，应删除 |
| [manager.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/plugin/manager.go) L63-68 | `SetWorkflowRegistry(wr interface{})` | 空实现 |
| [interfaces.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/workflow/interfaces.go) | `TaskManager`, `PluginManager`, `PythonEngine`, `EngineOptions` | 定义但未使用 |

### 3.5 TODO/FIXME 统计

| 位置 | 类型 | 内容 |
|------|------|------|
| [server.py](file:///d:/AIstudio-master/AIstudio-master/Engine/server.py) | TODO | 多处注释标记待完善功能 |
| [runner.py](file:///d:/AIstudio-master/AIstudio-master/Engine/runner.py) | TODO | 需要扩展插件注册表 |
| [gpu_manager.py](file:///d:/AIstudio-master/AIstudio-master/Engine/runtime/gpu_manager.py) | TODO | CPU 模式下的内存监控 |

---

## 四、命名规范问题

### 4.1 Go 代码

| 位置 | 问题 | 建议 |
|------|------|------|
| [models.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/plugin/models.go) | `PluginType` 常量 `PluginTypeVision` | 需移除 `PluginType` 前缀 |
| [router.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/api/router.go) | 混合使用 `snake_case` 和 `camelCase` | 统一为 `camelCase` |

### 4.2 Python 代码

| 位置 | 问题 | 建议 |
|------|------|------|
| [server.py](file:///d:/AIstudio-master/AIstudio-master/Engine/server.py) | 中文注释混用 | 统一为英文注释 |
| [runner.py](file:///d:/AIstudio-master/AIstudio-master/Engine/runner.py) | 函数名 `get_subsystem` 和 `init_engine` 风格不一致 | 统一为 `snake_case` |

### 4.3 TypeScript 代码

| 位置 | 问题 | 建议 |
|------|------|------|
| [workflow.ts](file:///d:/AIstudio-master/AIstudio-master/Frontend/src/api/workflow.ts) | 同时存在 `getWorkflow` 和 `getWorkflowById` 两个别名 | 删除重复 |
| [task.ts](file:///d:/AIstudio-master/AIstudio-master/Frontend/src/api/task.ts) | 同时存在 `getTaskById` 和 `getTask` | 删除重复 |

---

## 五、配置系统统一建议

### 5.1 当前状态

```
配置来源:
├── Config/app.yaml          (Launcher 主配置)
├── Config/backend.yaml      (Backend 配置)
├── Config/engine.yaml       (Engine 配置)
├── Config/plugin.yaml       (Plugin 配置)
├── Backend/config/default.yaml    (Go 默认值)
├── Backend/config/development.yaml (开发环境)
├── Backend/config/production.yaml  (生产环境)
├── 环境变量 (最高优先级)
└── 代码默认值 (最低优先级)
```

### 5.2 建议统一方案

```
配置来源:
├── Config/
│   ├── default.yaml         (所有模块的默认值)
│   ├── development.yaml     (开发环境覆盖)
│   ├── production.yaml      (生产环境覆盖)
│   └── secret.yaml          (密钥，不提交到 Git)
├── 环境变量 (最高优先级)
└── 代码默认值 (最低优先级)
```

### 5.3 需要统一的环境变量命名

当前无统一规范，建议：

```
AISTUDIO_SERVER_PORT=8081
AISTUDIO_DATABASE_URL=sqlite://aistudio.db
AISTUDIO_LOG_LEVEL=debug
AISTUDIO_JWT_SECRET=xxx
AISTUDIO_LLM_API_KEY=xxx
```

---

## 六、日志系统统一建议

### 6.1 当前日志系统碎片化

| 日志组件 | 存储位置 | 查询方式 | 问题 |
|---------|---------|---------|------|
| Go `log` 标准库 | stdout | 无 | 无结构化，无法查询 |
| `LogService` | 内存 | `/api/logs` | 最大 10000 条，重启丢失 |
| `TaskLogger` | 内存 | 仅通过 Go 代码 | 无清理机制 |
| Python `sdk.logger` | stdout JSON | 由 Go 解析 | 不直接暴露给用户 |
| Gin Logger | stdout | 无 | 仅请求日志 |

### 6.2 统一方案

```
日志流:
Python Engine → stdout JSON → Go TaskLogger → LogService → API → Frontend
                                     ↓
                               WebSocket → Frontend (实时)
```

### 6.3 日志格式标准

```json
{
    "timestamp": "2026-07-10T10:00:00Z",
    "level": "INFO",
    "source": "engine/python",
    "message": "Training started",
    "taskId": "abc-123",
    "workflowId": "wf-456",
    "pluginId": "yolo-detector",
    "detail": "Epoch 1/100, loss=0.5234"
}
```

### 6.4 日志等级标准

| 等级 | 数值 | 说明 |
|------|------|------|
| DEBUG | 0 | 调试信息，仅开发环境 |
| INFO | 1 | 正常操作信息 |
| WARN | 2 | 警告，不影响功能 |
| ERROR | 3 | 错误，功能受影响 |
| FATAL | 4 | 致命错误，服务退出 |

---

## 七、重构建议优先级

### 7.1 P0 - 立即处理

1. 合并 `store/` 和 `stores/` 目录
2. 删除重复 API 路由
3. 修复 `SetWorkflowRegistry` 空实现
4. 删除废弃代码和未使用变量

### 7.2 P1 - 本周内处理

1. 统一配置系统
2. 统一日志系统
3. 拆分 `main()` 函数
4. 替换 Magic Number 为常量

### 7.3 P2 - 本月内处理

1. 减少代码重复
2. 统一命名规范
3. 添加单元测试
4. 清理 TODO/FIXME