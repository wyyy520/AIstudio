# AIStudio 数据流审计：实际 vs silu.md 要求

## silu.md 要求的数据流

```
Workflow Editor → Workflow Store → workflow.json → Compiler → Execution Plan → Generator → Real Project → Runtime → Log Center → Diagnose
```

## 当前实际数据流

```
Frontend Editor (Vue Flow)
    ↓ 用户编辑
Workflow Store (Pinia: apps/desktop/src/workflow/store.ts)
    ↓ 实时同步 (bridge.ts)
workflow.json on disk (通过 PUT /api/projects/:id/workflow 保存)
    ↓
runtime_service.go: readWorkflow() → json.Unmarshal → workflow.Workflow struct
    ↓
internal/compiler.Compile() → 直接调 generator.Generate(ctx, wf, opts)
    ↓
generator 内部用 fmt.Sprintf() 拼字符串 → 写入项目文件
    ↓
(没有 Execution Plan, 没有 Template Engine)
```

## 各环节与 silu.md 对照

| silu.md 要求 | 当前状态 | 说明 |
|---|---|---|
| **Workflow Store** — 单一状态管理中心 | ✅ **OK** | Pinia store (`workflow/store.ts`) + bridge 实现双向同步 |
| **实时数据绑定** — 无"保存"按钮 | ✅ **OK** | PropertyPanel 用 @input 实时更新 |
| **workflow.json** 是唯一事实源 | ⚠️ **部分** | 前端存了，API 也能读写，但 Compiler 未做 validation |
| **自动保存** | ✅ **OK** | Debounce 1秒的 auto-save |
| **Compiler 验证工作流** | ❌ **缺失** | internal/compiler 不调 DAG 排序，不检查参数合法性 |
| **Compiler 生成 Execution Plan** | ❌ **缺失** | 直接传 raw workflow 给 Generator |
| **Generator 消费 Execution Plan** | ❌ **缺失** | Generator 读的是 workflow.Workflow struct |
| **Template Engine** | ❌ **缺失** | 各 Generator 用 fmt.Sprintf 拼字符串，无模板引擎 |
| **多域联合生成** | ❌ **缺失** | 一次 Compile() 只支持一个 target |
| **Runtime 环境检测** | ⚠️ 部分 | Detect() 接口存在但 localRuntime 没调真实检测 |
| **Log Center 持久化** | ✅ **新增** | PersistentLogCenter 已完成 |
| **Log Center 分类** | ✅ **OK** | 按 source 映射到 system/runtime/compiler/... |
| **Diagnose Center** | ❌ **缺失** | 未与 Compiler/Runtime 错误对接 |
| **Undo/Redo** | ✅ **新增** | 50步历史 + 剪贴板 |
| **键盘快捷键** | ✅ **新增** | Ctrl+Z/Y/S/C/V |
| **Executor（进程管理）** | ✅ **新增** | ProcessManager 已完成 |

## 最大的架构裂缝

**问题：有两个平行的 Compiler 体系**

1. `packages/compiler/` — 共享库，我在这里加了 ExecutionPlan 和 BuildExecutionPlan
2. `apps/backend/internal/compiler/` — 后端实际使用的编译器

**两者完全不同**：
- `internal/compiler` 有自己一套 CompilerOptions/CompileResult/Generator 类型
- `internal/compiler` 调 `generator.Generate(ctx, wf, opts)` — 没有 ExecutionPlan
- `internal/compiler` 的 Generator 用 `fmt.Sprintf()` 拼代码，不是模板

**修复方案（下一轮要做的事）：**

1. 让 `apps/backend/internal/compiler/` 使用 `packages/compiler` 的 ExecutionPlan + BuildExecutionPlan
2. 把 `internal/compiler/generators/python/generator.go` 从 fmt.Sprintf 改成调 Template Engine
3. 在 Compile 流程中加入 workflow 验证 (DAG topological sort + 参数检查)
4. 连接 Log Center → Diagnose Center 的管道
5. 把 `internal/workflow/types.go` 删掉，统一用 `packages/workflow`

## 当前可以运行的端到端流程

```
POST  /api/projects           — 创建项目 (含12个目录)
PUT   /api/projects/:id/workflow — 保存 workflow.json
GET   /api/projects/:id/workflow — 读取 workflow.json
POST  /api/projects/:id/compile  — 编译 (internal/compiler) → 生成项目文件
POST  /api/projects/:id/run      — 编译 + 运行
GET   /api/runtime/status/:runId — 运行状态
POST  /api/runtime/detect        — 环境检测
```
