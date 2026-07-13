# AIStudio 重构报告

> 生成日期：2026-07-09
> 阶段：代码整理与架构收敛

---

## 一、概述

本次重构聚焦于清理项目冗余内容、删除死代码、合并重复模块、统一目录规范，**不改变任何业务逻辑与核心功能**。

---

## 二、删除的文件和目录

### 2.1 Mock 数据文件（未被引用）

| 文件 | 原因 |
|------|------|
| `Frontend/src/mock/projects.ts` | 未被任何文件 import |
| `Frontend/src/pages/Logs/mock.ts` | 未被任何文件 import |
| `Frontend/src/pages/PluginStore/mock.ts` | 未被任何文件 import |
| `Frontend/src/pages/AIChat/mock.ts` | 仅被 `pages/AIChat/AIChat.vue` 引用（该页面未被 router 使用） |

### 2.2 重复目录

| 目录 | 原因 |
|------|------|
| `Frontend/src/styles/` | 与 `assets/styles/` 完全重复；`main.ts` 使用 `assets/styles/` |
| `Frontend/src/workflow/` | 空目录（仅含 `.gitkeep`），无实际文件 |

### 2.3 编译产物与临时文件

| 文件 | 原因 |
|------|------|
| `Backend/cmd.exe` | Go 编译产物，非源码 |
| `Backend/test_config.exe` | 测试编译产物 |
| `Backend/aistudio.db` | SQLite 测试数据库残留 |
| `Backend/internal/api/handlers/aistudio.db` | SQLite 测试数据库残留 |
| `Launcher/AIStudio.exe~` | 编译中间产物 |
| `Engine/sdk/__pycache__/` | Python 字节码缓存 |
| `package-lock.json`（项目根目录） | 空占位文件，`Frontend/package-lock.json` 为实际依赖锁 |

### 2.4 空目录（仅含 `.gitkeep`）

| 目录 | 原因 |
|------|------|
| `Backend/internal/logger/` | 空的包目录，未实现任何日志代码 |
| `Backend/internal/project/` | 空的包目录，无任何 Go 文件 |
| `Backend/pkg/` | 空的包目录，无任何共享库 |
| `Docs/mcp/` | 空文档目录，无实际内容 |
| `Docs/prd/` | 空文档目录，无实际内容 |
| `Docs/roadmap/` | 空文档目录，无实际内容 |

### 2.5 冗余 `.gitkeep` 文件

从以下已有实际文件的目录中删除了 `.gitkeep`（git 不再需要它们来占位）：

- `Backend/internal/agent/`
- `Backend/internal/common/`
- `Backend/internal/mcp/`
- `Backend/internal/workflow/`
- `Docs/api/`
- `Docs/architecture/`
- `Docs/plugin-sdk/`
- `Docs/workflow-sdk/`
- `Docs/UI/`
- `Docs/protocol/`
- `Frontend/public/`
- `Frontend/src/api/`
- `Frontend/src/assets/`
- `Frontend/src/components/`
- `Frontend/src/router/`
- `Frontend/src/store/`

### 2.6 保留的 `.gitkeep` 文件

以下目录为有意保留的空白脚手架，`.gitkeep` 维持 git 目录追踪：

- `Plugins/` 全部子目录（Logic, NLP, Vision 等）
- `Runtime/` 全部子目录（cache, logs, tasks, temp, workspace 等）
- `Scripts/` 全部子目录（build, ci, deploy, installer, release）
- `Storage/` 全部子目录（datasets, models, projects, settings, users）
- `Frontend/dist/`（构建输出目录）

---

## 三、模块合并评估

### 3.1 配置系统（Config）

| 位置 | 角色 | 说明 |
|------|------|------|
| `Config/app.yaml` | Launcher 顶层配置 | 启动顺序、路径解析 |
| `Config/backend.yaml` | Backend 运行时配置（Launcher 传入 `AISTUDIO_CONFIG`） | 含 `server`, `database`, `jwt`, `llm`, `mcp` 等 |
| `Backend/config/default.yaml` | Backend 备用配置 | 仅在直接运行 Backend（无 Launcher）时使用 |

**评估：** `Config/backend.yaml` 与 `Backend/config/default.yaml` 存在字段重叠（server.port, database, engine, plugin, log 等）。但 Launcher 通过环境变量显式指定使用 `Config/backend.yaml`，`Backend/config/default.yaml` 仅为直接运行 Backend 的 fallback。合并需修改 Launcher 启动逻辑 → **暂不合并**，建议在后续架构升级中统一配置中心。

### 3.2 状态管理（Store）

| 位置 | 角色 | 使用方 |
|------|------|--------|
| `Frontend/src/store/` | 旧 Store 系统（7 个 store） | `components/settings/*`, `components/logs/*`, `components/plugin/*`, `components/project/*`, `pages/*` |
| `Frontend/src/stores/` | 新 Store 系统（5 个 store） | `views/*`（路由激活的页面） |

**评估：** 两套 Store 系统同时活跃，分别服务于 `views/` 和 `pages/` + `components/`。合并需统一前端页面架构 → **暂不合并**，建议在路由迁移时统一。

### 3.3 API 封装

`Frontend/src/api/` 为单一 API 层，无重复封装。各模块职责清晰。

### 3.4 日志系统

- `Backend/internal/api/middleware/logger.go` — HTTP 请求日志中间件
- `Backend/internal/engine/task_logger.go` — Python 任务执行日志
- `Engine/sdk/logger.py` — Python SDK 日志
- `Frontend/src/api/log.ts` — 前端日志查询客户端

**评估：** 各日志模块职责不同（HTTP、任务引擎、Python SDK、前端查询），非重复实现。已删除空的 `Backend/internal/logger/` 目录。

### 3.5 组件重复（Flat vs Subdirectory）

`Frontend/src/components/` 中存在扁平文件与子目录两种模式：

| 组件 | 扁平文件 | 子目录文件 | 使用情况 |
|------|---------|-----------|---------|
| `AppButton` | `AppButton.vue` | `AppButton/AppButton.vue` | 两者均被使用（views 用扁平，pages/components 用子目录） |
| `AppTag` | `AppTag.vue` | `AppTag/AppTag.vue` | 同上 |
| `AppInput` | `AppInput.vue` | `AppInput/AppInput.vue` | 同上 |
| `AppModal` | `AppModal.vue` | 仅扁平 | 仅扁平存在 |
| `AppSwitch` | 仅子目录 | `AppSwitch/AppSwitch.vue` | 被 components 引用 |

**评估：** 两者均活跃使用，不能直接删除任一版本 → **暂不合并**，建议统一为子目录模式后迁移。

---

## 四、目录规范对照

### Frontend

| 规范要求 | 实际存在 | 状态 |
|----------|---------|------|
| `components/` | ✅ | 已清理 `.gitkeep` |
| `pages/` | ✅ | 未被路由引用但类型被多处 import |
| `api/` | ✅ | 职责清晰 |
| `store/` | ✅ | 与 `stores/` 并行存在 |
| `utils/` | ⚠️ | 仅 `pages/AIChat/utils/` 存在，非顶层 |
| `stores/` | ✅ | 新系统，路由页面使用 |
| `views/` | ✅ | 路由当前指向 |

### Backend

| 规范要求 | 实际存在 | 状态 |
|----------|---------|------|
| `api/` | `internal/api/` | ✅ |
| `service/` | `internal/service/` | ✅ |
| `workflow/` | `internal/workflow/` | ✅ |
| `task/` | `internal/task/` | ✅ |
| `plugin/` | `internal/plugin/` | ✅ |
| `agent/` | `internal/agent/` | ✅ |

### Engine

| 规范要求 | 实际存在 | 状态 |
|----------|---------|------|
| `runtime/` | `Engine/runtime/` | ✅ |
| `sdk/` | `Engine/sdk/` | ✅ |
| `models/` | ❌ 不存在 | README 中有描述但实际未实现 |

---

## 五、潜在问题

### 5.1 依赖风险

以下删除操作无依赖风险（已通过 grep 确认无任何 `import` 或路线引用）：

- ✅ `Frontend/src/mock/projects.ts` — 0 处引用
- ✅ `Frontend/src/pages/Logs/mock.ts` — 0 处引用
- ✅ `Frontend/src/pages/PluginStore/mock.ts` — 0 处引用
- ✅ `Frontend/src/pages/AIChat/mock.ts` — 仅被未激活的 `pages/AIChat/AIChat.vue` 引用
- ✅ `Frontend/src/styles/` — 0 处引用 (`main.ts` 使用 `assets/styles/`)
- ✅ 所有删除的文件 — 已通过 grep 确认无 import 链

### 5.2 需要关注的结构问题

#### 5.2.1 `pages/` vs `views/` 双轨制

`pages/` 包含更完整的模块化实现（含子组件、类型、config），但路由使用 `views/`（单体组件）。
- `pages/` 的类型文件被 `stores/`, `api/`, `components/` 引用
- `views/` 的单体页面功能与 `pages/` 中的对应页面功能重叠

**建议：** 后续将路由从 `views/` 迁移到 `pages/`，并删除 `views/`

#### 5.2.2 `Config/` 与 `Backend/config/` 配置重叠

Launcher `service.go:136` 显式传递 `AISTUDIO_CONFIG=Config/backend.yaml`，覆盖了 Backend 默认的 `Backend/config/default.yaml`。

**建议：** 将 Backend 配置统一到 `Backend/config/` 下，Launcher 指向该路径

#### 5.2.3 `Engine/README.md` 与实际目录不一致

README 描述存在 `grpc/`, `models/`, `handlers/`, `manager/`, `proto/` 子目录，但实际文件系统中不存在。

**建议：** 更新 README 以匹配实际代码结构

#### 5.2.4 `Backend/config/README.md` 字段过时

`engine.address` 和 `engine.grpc_port` 字段在 README 中有说明，但实际 Config struct 使用 `python_path` 和 `engine_dir`。

**建议：** 更新 README

#### 5.2.5 `Backend/internal/common/errors.go` 孤立文件

`internal/common/` 目录仅含 `errors.go`（定义 `AppError` 类型），无包内其他文件或导出链确认。

**建议：** 确认是否被引用，如无引用可删除

---

## 六、统计汇总

| 类别 | 数量 |
|------|------|
| 删除 Mock 数据文件 | 4 个 |
| 删除重复样式目录 | 1 个 |
| 删除空占位目录 | 1 个 |
| 删除编译产物 | 5 个 |
| 删除测试数据库 | 2 个 |
| 删除 Python 缓存 | 1 个（2 个 .pyc 文件） |
| 删除空包目录 | 3 个 |
| 删除空文档目录 | 3 个 |
| 删除根目录占位文件 | 1 个 |
| 清理冗余 `.gitkeep` | 16 个 |
| **合计** | **37 项** |

---

## 七、后续建议

| 优先級 | 建议 |
|--------|------|
| P0 | 统一组件目录为子目录模式，删除扁平版本 |
| P0 | 统一 Store 系统，删除 `store/` 或 `stores/` |
| P1 | 将路由从 `views/` 迁移至 `pages/` |
| P1 | 合并 `Config/` 与 `Backend/config/` 配置 |
| P2 | 清理 `Backend/config/README.md` 过期文档 |
| P2 | 同步 `Engine/README.md` 与实际目录结构 |
| P2 | 删除 `Backend/internal/common/errors.go`（如确无引用） |
