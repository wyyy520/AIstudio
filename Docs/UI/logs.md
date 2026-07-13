# AIStudio Logs Center 页面设计规范

# 1. 页面定位

**Logs Center 是 AIStudio 的智能日志中枢，不是终端日志查看器。**

传统日志系统让用户在大量 Python Traceback 中自行翻找问题。AIStudio 的日志中心要做的是：自动收集日志 → AI 理解日志 → 总结问题 → 提供解决方案 → 辅助修复。用户优先看到的是"发生了什么、为什么失败、如何解决"，而不是大量原始报错堆栈。

### 核心设计理念

日志中心包含三个层级：

| 层级 | 内容 | 面向用户 |
|------|------|----------|
| Level 1 | 任务执行状态 | 所有用户 |
| Level 2 | 人类可读日志 | 普通用户 |
| Level 3 | AI 错误分析 | 需要排错的用户 |

**用户优先看到：**

1. 发生了什么（任务状态）
2. 为什么失败（AI 分析）
3. 如何解决（修复方案）

**而不是：**

- 大量 Python Traceback
- 无结构的终端输出
- 需要人工逐行阅读的日志

### 设计参考

| 参考产品 | 借鉴点 |
|----------|--------|
| Codex Desktop | AI 错误分析与自动修复流程 |
| Claude Code | 自然语言日志解读与建议 |
| JetBrains Run Console | 结构化运行输出、测试结果 |
| MLFlow | 训练指标监控与可视化 |
| Docker Desktop | 容器状态驱动与日志面板 |
| Cursor | AI 辅助排错交互模式 |

### 与旧版"日志页面"的核心区别

| 维度 | 旧版（日志查看器） | 新版（智能日志中枢） |
|------|--------------------|----------------------|
| 定位 | 终端日志查看 | AI 驱动的日志理解与修复 |
| 布局 | 双栏（筛选 + 日志流） | 四区（任务 + 日志 + AI 分析 + 监控） |
| 日志呈现 | 原始文本流 | Human Log + Raw Log 双模式 |
| 错误处理 | 高亮显示 | AI 分析卡片 + 修复方案 |
| 训练支持 | 无 | Training Monitor 实时指标 |
| 工作流支持 | 无 | Workflow Timeline 节点执行追踪 |
| AI 集成 | 单一"分析"按钮 | 完整分析流程：理解 → 分类 → 方案 → 执行 |

---

# 2. 用户流程

## 2.1 核心用户角色

| 角色 | 典型场景 |
|------|----------|
| AI 工程师 | YOLO 训练失败，查看 AI 分析，一键修复 CUDA 版本问题 |
| 算法研究员 | 监控训练 Loss/Accuracy 曲线，发现过拟合，调整参数 |
| 仿真工程师 | SUMO 仿真运行异常，查看 Workflow Timeline 定位失败节点 |
| 系统管理员 | 系统服务异常，查看系统日志，AI 分析根因 |

## 2.2 主流程

```
任务执行 ──→ 产生日志 ──→ 自动分类 ──→ 用户查看
    │              │            │            │
    │              │            │            ├─→ Level 1: 任务状态
    │              │            │            ├─→ Level 2: Human Log
    │              │            │            └─→ Level 3: AI 分析
    │              │            │
    │              │            └─→ 错误自动识别
    │              │
    │              └─→ 实时流式收集
    │
    ├─→ Training Task → Training Monitor
    ├─→ Workflow Task → Timeline
    └─→ System Task → Raw Log
```

## 2.3 AI 自动分析流程

```
Raw Log 产生
    │
    ▼
LLM Analysis（自动触发 / 手动触发）
    │
    ▼
Error Classification（错误分类）
    │
    ├── CUDA Error
    ├── Dependency Error
    ├── Model Error
    ├── Config Error
    └── Runtime Error
    │
    ▼
Solution Generation（方案生成）
    │
    ├── 方案 1: 升级 CUDA
    ├── 方案 2: 安装对应 PyTorch 版本
    └── 方案 3: 修改配置文件
    │
    ▼
User Confirmation（用户确认）
    │
    ├── Apply Fix → 自动执行修复
    ├── Generate Command → 生成修复命令
    └── Ignore → 忽略此问题
    │
    ▼
Execute Fix（执行修复）
    │
    ▼
Verify（验证修复结果）
```

## 2.4 错误发现流程

```
用户发现任务失败
    │
    ▼
点击失败任务
    │
    ▼
查看 Human Log（发生了什么）
    │
    ▼
AI Analysis Panel 自动展开（为什么失败）
    │
    ▼
查看解决方案（如何修复）
    │
    ▼
点击 Apply Fix / Generate Command
    │
    ▼
修复执行 → 任务重新运行
```

---

# 3. 页面布局

## 3.1 整体四区布局

```
┌──────────────────────────────────────────────────────────────────────┐
│  Toolbar: [🔍 Search] [Filter ▾] [Clear] [Export] [AI Analyze] [⚡ Auto Fix] │
├──────────────┬───────────────────────────────────┬───────────────────┤
│              │                                   │                   │
│  Task List   │  Log Workspace                    │  AI Analysis      │
│              │                                   │  Panel            │
│  240px       │  flex-1                           │  360px            │
│  (可折叠)     │  (自适应)                          │  (可折叠)          │
│              │                                   │                   │
│              │  ┌─────────────────────────────┐  │                   │
│              │  │ [Human Log] [Raw Log]        │  │                   │
│              │  ├─────────────────────────────┤  │                   │
│              │  │                             │  │                   │
│              │  │  Log Content                │  │                   │
│              │  │                             │  │                   │
│              │  │                             │  │                   │
│              │  └─────────────────────────────┘  │                   │
│              │                                   │                   │
├──────────────┴───────────────────────────────────┴───────────────────┤
│  Training Monitor / Workflow Timeline (可折叠, 默认收起)              │
└──────────────────────────────────────────────────────────────────────┘
```

## 3.2 区域尺寸规范

| 区域 | 默认宽度/高度 | 最小 | 最大 | 可调整 | 可折叠 |
|------|---------------|------|------|--------|--------|
| Task List | 240px | 200px | 320px | 拖拽 | 是（折叠为 48px 图标栏） |
| Log Workspace | flex-1 | 400px | - | 自适应 | 否 |
| AI Analysis Panel | 360px | 280px | 440px | 拖拽 | 是（折叠为 48px 图标栏） |
| Bottom Panel | 200px | 120px | 400px | 拖拽 | 是（完全收起） |

## 3.3 分割线

- 栏与栏之间使用可拖拽分割线
- 分割线宽度：1px
- 颜色：`border-subtle`
- 悬停时高亮为 `primary`，宽度变为 3px，提示可拖拽

---

# 4. 组件设计

## 4.1 Toolbar（顶部工具栏）

**位置**：页面顶部，横跨全宽

**高度**：48px

**内容**：

| 位置 | 元素 | 说明 |
|------|------|------|
| 左侧 | 搜索框 | 支持正则表达式，实时过滤日志 |
| 中间 | Filter 下拉 | 日志级别过滤：All / Error / Warning / Info / Debug |
| 右侧 | Clear 按钮 | 清空当前日志 |
| 右侧 | Export 按钮 | 导出日志（.log / .json） |
| 右侧 | AI Analyze 按钮 | 触发 AI 分析当前任务日志 |
| 右侧 | Auto Fix 按钮 | AI 自动修复（主色按钮） |

**Filter 下拉选项**：

| 选项 | 图标 | 颜色 | 说明 |
|------|------|------|------|
| All | `layers` | `text-secondary` | 显示全部日志 |
| Error | `alert-circle` | `error` | 仅显示错误 |
| Warning | `alert-triangle` | `warning` | 仅显示警告 |
| Info | `info` | `info` | 仅显示信息 |
| Debug | `bug` | `text-tertiary` | 仅显示调试 |

**样式**：
- 背景：`bg-secondary`
- 底部边框：`border-subtle` 1px
- 内边距：0 `spacing-4`

---

## 4.2 TaskList（左侧任务列表）

**宽度**：240px（默认）

**参考**：VS Code Debug Console 任务列表 + Docker Desktop 容器列表

### 结构

```
┌──────────────────────────────┐
│  🔍 Search tasks...          │  ← 搜索框
├──────────────────────────────┤
│  ▼ Running              (2)  │  ← 状态分组
│    ┌────────────────────┐    │
│    │ 🔵 YOLO Training   │    │
│    │    Training  12:30  │    │
│    │    ● Running  5m 32s│    │
│    └────────────────────┘    │
│    ┌────────────────────┐    │
│    │ 🔵 Smart Traffic   │    │
│    │    Simulation 12:28│    │
│    │    ● Running  7m 15│    │
│    └────────────────────┘    │
│  ▼ Failed               (1)  │
│    ┌────────────────────┐    │
│    │ 🔴 Model Export    │    │
│    │    Export   12:15   │    │
│    │    ✗ Failed   2m 10│    │
│    └────────────────────┘    │
│  ▼ Completed            (3)  │
│    ┌────────────────────┐    │
│    │ 🟢 Data Preprocess │    │
│    │    Pipeline  11:45  │    │
│    │    ✓ Done     1m 30│    │
│    └────────────────────┘    │
├──────────────────────────────┤
│  6 tasks  ·  2 running       │  ← 底部统计
└──────────────────────────────┘
```

### 任务卡片

每个任务卡片显示：

| 元素 | 样式 | 说明 |
|------|------|------|
| 状态图标 | 左侧，16px | 🔵 Running / 🟢 Success / 🔴 Failed / 🟡 Warning |
| 任务名称 | `body`，`text-primary` | 粗体 |
| 任务类型 | `caption`，类型色 | Training / Simulation / Export / System |
| 开始时间 | `caption`，`text-tertiary` | HH:mm 格式 |
| 状态文字 | `caption`，状态色 | Running / Failed / Done |
| 耗时 | `caption`，`text-tertiary`，等宽字体 | 5m 32s 格式 |

### 任务状态定义

| 状态 | 图标 | 颜色 | 动画 |
|------|------|------|------|
| Running | 脉冲圆点 | `info` | 呼吸动画 |
| Success | 对勾 | `success` | 无 |
| Failed | 叉号 | `error` | 无 |
| Warning | 感叹号 | `warning` | 无 |

### 任务类型

| 类型 | 标签色 | 说明 |
|------|--------|------|
| Training | `vision` / `nlp` / `timeseries` | 模型训练任务 |
| Simulation | `simulation` | 仿真运行任务 |
| Export | `system` | 模型导出任务 |
| Workflow | `primary` | 工作流执行任务 |
| System | `system` | 系统任务 |
| Agent | `agent` | Agent 执行任务 |

### 交互

- 点击任务 → 中间区域加载该任务的日志
- 搜索 → 实时过滤任务名称
- 状态分组可折叠

---

## 4.3 LogWorkspace（中间日志工作区）

**核心原则**：不要直接展示原始日志，提供 Human Log 和 Raw Log 双模式。

### Tab 切换

```
┌──────────────────────────────────────────────────┐
│  [Human Log]  [Raw Log]                          │  ← Tab 栏
├──────────────────────────────────────────────────┤
│                                                  │
│  日志内容区域                                     │
│                                                  │
└──────────────────────────────────────────────────┘
```

- Tab 栏高度：36px
- 背景：`bg-tertiary`
- 选中 Tab：`bg-active` + 底部 2px `primary` 指示线
- 未选中 Tab：`text-secondary`

---

### 4.3.1 Human Log（人类可读日志）

**面向**：普通用户

**原则**：将技术日志转化为人类可理解的操作步骤描述。

```
┌──────────────────────────────────────────────────┐
│  [10:32]  ✓ 正在检查 Python 环境                  │
│  [10:32]  ✓ Python 3.11.2 已就绪                  │
│  [10:33]  ✓ 正在安装 PyTorch 2.1.0                │
│  [10:35]  ✓ PyTorch 安装完成                       │
│  [10:35]  ● 正在下载 YOLOv8 模型权重               │  ← 进行中
│  [10:36]  ✓ 模型权重下载完成                       │
│  [10:36]  ● 正在开始 YOLO 训练                     │
│  [10:38]  ✗ CUDA 版本不匹配，训练失败              │  ← 失败
│                                                  │
│  ┌────────────────────────────────────────────┐  │
│  │  ⚠ CUDA Version Mismatch                   │  │  ← 错误卡片（内嵌）
│  │  Severity: High                             │  │
│  │  Problem: PyTorch 2.1 需要 CUDA ≥12.0      │  │
│  │  Current: CUDA 11.8                         │  │
│  │  [View AI Analysis →]                       │  │
│  └────────────────────────────────────────────┘  │
│                                                  │
│  [10:38]  ○ 注册 Workflow 节点                    │  ← 待执行
│  [10:38]  ○ 注册 Agent 工具                       │  ← 待执行
└──────────────────────────────────────────────────┘
```

**Human Log 行样式**：

| 元素 | 样式 |
|------|------|
| 时间戳 | `caption`，等宽字体，`text-tertiary` |
| 状态图标 | ✓ `success` / ● `info` / ✗ `error` / ○ `text-disabled` |
| 步骤描述 | `body-sm`，`text-primary` |
| 进行中行 | 左侧 2px `info` 竖线 + 呼吸动画 |
| 失败行 | 左侧 2px `error` 竖线 + 浅红背景 |
| 完成行 | 无特殊 |

**错误卡片（内嵌）**：

当 Human Log 中出现错误时，在对应步骤下方内嵌一张 ErrorCard，引导用户查看 AI 分析。

---

### 4.3.2 Raw Log（原始日志）

**面向**：开发者

**原则**：完整保留原始输出，支持搜索、复制、下载。

```
┌──────────────────────────────────────────────────┐
│  [Wrap ▾] [Copy All] [Download]                  │  ← Raw Log 工具栏
├──────────────────────────────────────────────────┤
│  1  2026-07-07 10:32:01 [INFO] Checking Python...│
│  2  2026-07-07 10:32:01 [INFO] Python 3.11.2     │
│  3  2026-07-07 10:33:15 [INFO] Installing torch  │
│  4  2026-07-07 10:35:22 [INFO] torch installed   │
│  5  2026-07-07 10:35:22 [INFO] Downloading model │
│  6  2026-07-07 10:36:01 [INFO] Model downloaded  │
│  7  2026-07-07 10:36:01 [INFO] Starting training │
│  8  2026-07-07 10:38:15 [ERROR] RuntimeError:    │
│  9  CUDA version mismatch                        │
│ 10  Expected CUDA >= 12.0, got 11.8              │
│ 11  Traceback (most recent call last):           │
│ 12    File "train.py", line 45, in <module>      │
│ 13      model.train()                            │
│ 14    File "torch/nn/module.py", line 1511       │
│ 15      ...                                      │
└──────────────────────────────────────────────────┘
```

**Raw Log 工具栏**：

| 按钮 | 功能 |
|------|------|
| Wrap | 切换自动换行 |
| Copy All | 复制全部日志到剪贴板 |
| Download | 下载为 .log 文件 |

**Raw Log 行样式**：

| 元素 | 样式 |
|------|------|
| 行号 | `caption`，等宽字体，`text-disabled`，右对齐，宽度 48px |
| 时间戳 | `caption`，等宽字体，`text-tertiary` |
| 级别标签 | `caption`，等宽字体，级别色 |
| 日志内容 | `code`，等宽字体，`text-primary` |
| ERROR 行 | 浅红背景 `error-bg` |
| WARN 行 | 浅黄背景 `warning-bg` |
| 选中行 | `bg-active` 背景 |

**日志级别颜色**：

| 级别 | 标签颜色 | 行背景 |
|------|----------|--------|
| INFO | `info` | 无特殊 |
| WARN | `warning` | `warning-bg` |
| ERROR | `error` | `error-bg` |
| DEBUG | `text-tertiary` | 无特殊 |

**默认隐藏**：Raw Log 默认不激活，需点击 Tab 切换。

---

## 4.4 AIAnalysisPanel（右侧 AI 分析面板）

**宽度**：360px（默认）

**参考**：Codex Error Fix + Claude Code 分析面板

**核心原则**：不是简单的日志搜索，而是 AI 理解日志后给出的结构化分析与修复方案。

### 结构

```
┌──────────────────────────────────────┐
│  🤖 AI Analysis                      │  ← 面板标题
├──────────────────────────────────────┤
│                                      │
│  ┌──────────────────────────────┐    │
│  │  🔴 CUDA Version Error       │    │  ← ErrorCard
│  │  Severity: Critical          │    │
│  │                              │    │
│  │  Problem:                    │    │
│  │  YOLO 训练失败               │    │
│  │                              │    │
│  │  Cause:                      │    │
│  │  CUDA 版本不匹配             │    │
│  │                              │    │
│  │  Detail:                     │    │
│  │  当前 PyTorch 需要 CUDA 12.1 │    │
│  │  系统检测到 CUDA 11.8        │    │
│  │                              │    │
│  │  Solutions:                  │    │
│  │  ┌──────────────────────┐    │    │
│  │  │ 1. 升级 CUDA 到 12.1 │    │    │
│  │  │   [Apply Fix]        │    │    │
│  │  └──────────────────────┘    │    │
│  │  ┌──────────────────────┐    │    │
│  │  │ 2. 安装对应 PyTorch  │    │    │
│  │  │   [Apply Fix]        │    │    │
│  │  └──────────────────────┘    │    │
│  │                              │    │
│  │  [Generate Command] [Ignore] │    │
│  └──────────────────────────────┘    │
│                                      │
│  ┌──────────────────────────────┐    │
│  │  🟡 Deprecation Warning      │    │  ← 另一张 ErrorCard
│  │  Severity: Warning           │    │
│  │  ...                         │    │
│  └──────────────────────────────┘    │
│                                      │
├──────────────────────────────────────┤
│  Agent Status: ● Analyzing...       │  ← AI Agent 状态
└──────────────────────────────────────┘
```

### ErrorCard（错误卡片）

每个检测到的错误生成一张 ErrorCard。

**卡片内容**：

| 区域 | 内容 | 样式 |
|------|------|------|
| 标题 | 错误类型名称 | `h3`，`text-primary` |
| Severity | Critical / Warning / Info | 徽章样式，对应颜色 |
| Problem | 问题描述 | `body-sm`，`text-secondary` |
| Cause | 原因分析 | `body-sm`，`text-secondary` |
| Detail | 详细解释 | `body-sm`，`text-tertiary`，可折叠 |
| Solutions | 修复方案列表 | 每个方案一张子卡片 |
| Actions | Apply Fix / Generate Command / Ignore | 按钮组 |

**Severity 徽章**：

| 级别 | 颜色 | 背景 | 图标 |
|------|------|------|------|
| Critical | `error` | `error-bg` | `alert-circle` |
| Warning | `warning` | `warning-bg` | `alert-triangle` |
| Info | `info` | `info-bg` | `info` |

**Solution 子卡片**：

```
┌──────────────────────────────────┐
│  方案 1: 升级 CUDA               │
│  安装 CUDA Toolkit 12.1          │
│  预计耗时: ~5 分钟               │
│                      [Apply Fix] │
└──────────────────────────────────┘
```

- 背景：`bg-tertiary`
- 圆角：`radius-lg`
- Apply Fix 按钮：Primary 样式

**Action 按钮**：

| 按钮 | 样式 | 功能 |
|------|------|------|
| Apply Fix | Primary | AI 自动执行修复 |
| Generate Command | Secondary | 生成修复命令，用户手动执行 |
| Ignore | Ghost | 忽略此问题 |

### AI Agent 状态栏

面板底部显示 AI Agent 当前状态：

| 状态 | 图标 | 颜色 | 动画 |
|------|------|------|------|
| Idle | 空心圆 | `text-tertiary` | 无 |
| Thinking | 脑图标 | `primary` | 脉冲 |
| Analyzing | 搜索图标 | `info` | 旋转 |
| Calling Tool | 工具图标 | `warning` | 闪烁 |
| Executing | 播放图标 | `success` | 脉冲 |
| Completed | 对勾 | `success` | 无 |
| Failed | 叉号 | `error` | 无 |

---

## 4.5 TrainingMonitor（训练监控面板）

**位置**：底部可折叠面板

**触发**：当选中任务类型为 Training 时自动展开

**参考**：MLFlow + TensorBoard

### 结构

```
┌──────────────────────────────────────────────────────────────────────┐
│  Training Monitor                                    [—] [✕]        │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────┐│
│  │  Epoch       │  │  Loss        │  │  Accuracy    │  │  GPU     ││
│  │  45 / 100    │  │  0.034       │  │  92.1%       │  │  85%     ││
│  │  ██████░░░░  │  │  ↓ 0.012     │  │  ↑ 1.3%      │  │  ████░░  ││
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────┘│
│                                                                      │
│  ┌─────────────────────────────┐  ┌─────────────────────────────┐   │
│  │  Loss Curve                 │  │  Accuracy Curve             │   │
│  │  ╲                          │  │               ╱             │   │
│  │   ╲                         │  │             ╱               │   │
│  │    ╲___                     │  │          ╱                  │   │
│  │        ───                  │  │       ╱                     │   │
│  │           ──                │  │    ╱                        │   │
│  └─────────────────────────────┘  └─────────────────────────────┘   │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

### 指标卡片

| 指标 | 格式 | 图标 | 颜色 |
|------|------|------|------|
| Epoch | `45 / 100` | `activity` | `info` |
| Loss | `0.034` | `trending-down` | `success`（下降时）/ `error`（上升时） |
| Accuracy | `92.1%` | `trending-up` | `success`（上升时）/ `error`（下降时） |
| GPU Usage | `85%` | `cpu` | `warning`（>90%）/ `info` |
| Learning Rate | `0.001` | `gauge` | `text-secondary` |
| Memory | `6.2 GB` | `hard-drive` | `text-secondary` |

**指标卡片样式**：

| 属性 | 值 |
|------|-----|
| 圆角 | `radius-xl` |
| 背景 | `bg-tertiary` |
| 内边距 | `spacing-4` |
| 数值字号 | `h1` |
| 标签字号 | `caption` |
| 趋势箭头 | 数值右侧，颜色跟随趋势 |

### 图表

- Loss Curve：折线图，X 轴 Epoch，Y 轴 Loss
- Accuracy Curve：折线图，X 轴 Epoch，Y 轴 Accuracy
- GPU Usage：面积图，X 轴时间，Y 轴百分比

**图表样式**：
- 背景：`bg-tertiary`
- 圆角：`radius-xl`
- 网格线：`border-subtle`
- 折线颜色：`primary`
- 面积填充：`primary-bg`
- 坐标轴文字：`caption`，`text-tertiary`

---

## 4.6 Timeline（工作流执行时间线）

**位置**：底部可折叠面板

**触发**：当选中任务类型为 Workflow 时自动展开

**参考**：Agent Timeline + CI/CD Pipeline 可视化

### 结构

```
┌──────────────────────────────────────────────────────────────────────┐
│  Workflow Timeline                                    [—] [✕]       │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ● Dataset Loader ──── ✓ Completed ────── 1.2s                      │
│  │                                                                   │
│  ● YOLO Training ──── ● Running ──────── 5m 32s                     │
│  │                     ████████████░░░░  45%                         │
│  │                                                                   │
│  ○ Export Model ────── ○ Waiting                                       │
│  │                                                                   │
│  ○ Upload Results ──── ○ Waiting                                       │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

### 节点行样式

| 元素 | 样式 |
|------|------|
| 状态圆点 | 12px，Completed=`success`，Running=`info`+脉冲，Failed=`error`，Waiting=`text-disabled` |
| 节点名称 | `body-sm`，`text-primary` |
| 状态文字 | `caption`，状态色 |
| 耗时 | `caption`，等宽字体，`text-tertiary` |
| 进度条 | Running 节点显示，4px 高，`primary` 色 |
| 连接线 | 1px，`border-default`，垂直连接节点 |

---

## 4.7 LogFilter（日志过滤器）

**位置**：Toolbar 内的下拉组件

### 结构

```
┌─────────────────────────────────┐
│  Level                          │
│  [✓] All                        │
│  [✓] Error          (3)         │
│  [✓] Warning        (5)         │
│  [✓] Info           (128)       │
│  [ ] Debug          (42)        │
│                                  │
│  Source                          │
│  [✓] System                     │
│  [✓] Workflow Engine            │
│  [✓] Plugin Runtime             │
│  [✓] Training                   │
│  [ ] Agent                      │
│                                  │
│  Time Range                      │
│  [Last 1h ▾]                    │
│                                  │
│  [Apply]  [Reset]               │
└─────────────────────────────────┘
```

**下拉样式**：
- 圆角：`radius-lg`
- 背景：`bg-secondary`
- 阴影：`shadow-lg`
- 边框：`border-default`

---

## 4.8 FixDialog（修复确认弹窗）

**触发**：点击 ErrorCard 中的 Apply Fix 按钮

### 结构

```
┌──────────────────────────────────────────┐
│  ⚡ Apply Fix                             │
├──────────────────────────────────────────┤
│                                          │
│  AI suggests the following fix:          │
│                                          │
│  ┌──────────────────────────────────┐    │
│  │  $ pip install torch==2.1.0      │    │  ← 命令预览
│  │    --index-url https://...       │    │
│  └──────────────────────────────────┘    │
│                                          │
│  This will:                              │
│  • Install PyTorch 2.1.0 with CUDA 12   │
│  • Remove current PyTorch 2.0.0         │
│  • Require ~2.3 GB download             │
│                                          │
│  ⚠ This action will restart the         │
│    training task.                        │
│                                          │
├──────────────────────────────────────────┤
│            [Cancel]  [Apply Fix]         │
└──────────────────────────────────────────┘
```

**弹窗样式**：
- 圆角：`radius-2xl`
- 背景：`bg-secondary`
- 阴影：`shadow-xl`
- 命令预览：等宽字体，`bg-tertiary`，`radius-lg`
- Apply Fix 按钮：Primary 样式
- Cancel 按钮：Ghost 样式

---

# 5. 数据结构设计

## 5.1 Task（任务）

```json
{
  "id": "task-001",
  "name": "YOLO Training",
  "type": "training",
  "status": "failed",
  "startedAt": "2026-07-07T10:32:00Z",
  "completedAt": "2026-07-07T10:38:15Z",
  "duration": 375,
  "projectId": "proj-001",
  "workflowId": "wf-001",
  "metadata": {
    "model": "YOLOv8",
    "dataset": "coco-2024",
    "epochs": 100,
    "currentEpoch": 45
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 任务唯一 ID |
| name | string | 任务名称 |
| type | TaskType | training / simulation / export / workflow / system / agent |
| status | TaskStatus | running / success / failed / warning |
| startedAt | ISO 8601 | 开始时间 |
| completedAt | ISO 8601 | 结束时间（可空） |
| duration | number | 耗时（秒） |
| projectId | string | 所属项目 ID |
| workflowId | string | 关联工作流 ID（可空） |
| metadata | object | 任务类型特定元数据 |

## 5.2 LogEntry（日志条目）

```json
{
  "id": "log-001",
  "taskId": "task-001",
  "timestamp": "2026-07-07T10:32:01Z",
  "level": "error",
  "source": "training",
  "message": "RuntimeError: CUDA version mismatch",
  "rawMessage": "RuntimeError: CUDA version mismatch\nExpected CUDA >= 12.0, got 11.8\n  at train.py:45",
  "humanMessage": "CUDA 版本不匹配，训练失败",
  "stepName": "Start Training",
  "stepStatus": "failed",
  "metadata": {
    "file": "train.py",
    "line": 45,
    "function": "train"
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 日志唯一 ID |
| taskId | string | 所属任务 ID |
| timestamp | ISO 8601 | 时间戳 |
| level | LogLevel | info / warning / error / debug |
| source | LogSource | system / workflow / plugin / training / agent |
| message | string | 日志消息（原始） |
| rawMessage | string | 完整原始输出（含 Traceback） |
| humanMessage | string | 人类可读描述（AI 生成或规则转换） |
| stepName | string | 所属步骤名称 |
| stepStatus | StepStatus | completed / running / failed / pending |
| metadata | object | 附加元数据（文件、行号等） |

## 5.3 ErrorAnalysis（AI 错误分析）

```json
{
  "id": "analysis-001",
  "taskId": "task-001",
  "logEntryIds": ["log-045", "log-046", "log-047"],
  "severity": "critical",
  "errorType": "CUDA_ERROR",
  "problem": "YOLO 训练失败",
  "cause": "CUDA 版本不匹配",
  "detail": "当前 PyTorch 2.1.0 需要 CUDA ≥12.0，系统检测到 CUDA 11.8",
  "solutions": [
    {
      "id": "sol-001",
      "title": "升级 CUDA",
      "description": "安装 CUDA Toolkit 12.1",
      "command": "sudo apt install cuda-toolkit-12-1",
      "estimatedTime": "~5 分钟",
      "risk": "low",
      "autoFixable": true
    },
    {
      "id": "sol-002",
      "title": "安装对应 PyTorch 版本",
      "description": "安装兼容 CUDA 11.8 的 PyTorch 版本",
      "command": "pip install torch==2.0.0+cu118",
      "estimatedTime": "~3 分钟",
      "risk": "low",
      "autoFixable": true
    }
  ],
  "status": "pending",
  "analyzedAt": "2026-07-07T10:38:20Z"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 分析唯一 ID |
| taskId | string | 所属任务 ID |
| logEntryIds | string[] | 关联的日志条目 ID |
| severity | Severity | critical / warning / info |
| errorType | string | 错误分类 |
| problem | string | 问题描述 |
| cause | string | 原因分析 |
| detail | string | 详细解释 |
| solutions | Solution[] | 修复方案列表 |
| status | AnalysisStatus | pending / fixing / fixed / ignored |
| analyzedAt | ISO 8601 | 分析时间 |

## 5.4 Solution（修复方案）

```json
{
  "id": "sol-001",
  "title": "升级 CUDA",
  "description": "安装 CUDA Toolkit 12.1",
  "command": "sudo apt install cuda-toolkit-12-1",
  "estimatedTime": "~5 分钟",
  "risk": "low",
  "autoFixable": true
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 方案唯一 ID |
| title | string | 方案标题 |
| description | string | 方案描述 |
| command | string | 执行命令（可空） |
| estimatedTime | string | 预计耗时 |
| risk | RiskLevel | low / medium / high |
| autoFixable | boolean | 是否可自动修复 |

## 5.5 TrainingMetrics（训练指标）

```json
{
  "taskId": "task-001",
  "currentEpoch": 45,
  "totalEpochs": 100,
  "metrics": {
    "loss": 0.034,
    "accuracy": 0.921,
    "learningRate": 0.001,
    "gpuUsage": 0.85,
    "memoryUsage": 6.2
  },
  "history": [
    { "epoch": 1, "loss": 0.892, "accuracy": 0.234, "gpuUsage": 0.72 },
    { "epoch": 2, "loss": 0.756, "accuracy": 0.345, "gpuUsage": 0.78 }
  ],
  "updatedAt": "2026-07-07T10:37:30Z"
}
```

## 5.6 WorkflowTimeline（工作流时间线）

```json
{
  "taskId": "task-002",
  "workflowId": "wf-001",
  "nodes": [
    {
      "nodeId": "node-001",
      "name": "Dataset Loader",
      "type": "data",
      "status": "completed",
      "startedAt": "2026-07-07T10:30:00Z",
      "completedAt": "2026-07-07T10:30:01Z",
      "duration": 1.2
    },
    {
      "nodeId": "node-002",
      "name": "YOLO Training",
      "type": "training",
      "status": "running",
      "startedAt": "2026-07-07T10:30:01Z",
      "progress": 0.45
    }
  ]
}
```

---

# 6. 交互规范

## 6.1 基本交互

| 操作 | 行为 |
|------|------|
| 点击任务 | 中间区域加载该任务日志，右侧显示 AI 分析 |
| 点击 Human Log 错误行 | 右侧 AI Analysis Panel 自动滚动到对应 ErrorCard |
| 点击 ErrorCard 的 View AI Analysis | 右侧面板展开并聚焦 |
| 点击 Apply Fix | 弹出 FixDialog 确认弹窗 |
| 点击 Generate Command | 复制修复命令到剪贴板，显示 Toast 提示 |
| 点击 Raw Log Tab | 切换到原始日志视图 |
| 点击 Export | 弹出导出选项（.log / .json / .csv） |
| 点击 Clear | 弹出确认弹窗，确认后清空当前任务日志 |

## 6.2 搜索交互

| 操作 | 行为 |
|------|------|
| 输入搜索词 | 实时高亮匹配结果，显示匹配数量 |
| Enter | 跳转到下一个匹配 |
| Shift+Enter | 跳转到上一个匹配 |
| 正则模式 | 切换正则表达式搜索 |
| 清除搜索 | 恢复全部显示 |

## 6.3 过滤交互

| 操作 | 行为 |
|------|------|
| 切换级别过滤 | 立即过滤日志列表，动画过渡 200ms |
| 切换来源过滤 | 立即过滤日志列表 |
| 切换时间范围 | 重新加载指定范围的日志 |
| Reset | 重置所有过滤条件 |

## 6.4 实时日志

| 操作 | 行为 |
|------|------|
| 新日志到达 | Human Log 自动追加步骤，Raw Log 自动滚动到底部 |
| 用户向上滚动 | 停止自动滚动，显示"回到最新"按钮 |
| 点击"回到最新" | 滚动到底部，恢复自动滚动 |
| 任务完成 | 停止自动滚动，显示完成状态 |

## 6.5 AI 分析交互

| 操作 | 行为 |
|------|------|
| 任务失败 | 自动触发 AI 分析（可配置） |
| 手动触发 | 点击 AI Analyze 按钮 |
| 分析进行中 | Agent Status 显示 Analyzing，ErrorCard 骨架屏 |
| 分析完成 | ErrorCard 逐步淡入，带 200ms 动画 |
| Apply Fix 执行中 | Agent Status 显示 Executing，按钮变为 Loading |
| 修复完成 | ErrorCard 状态变为 Fixed，显示绿色对勾 |
| 修复失败 | ErrorCard 状态保持，显示重试按钮 |

---

# 7. 动画规范

| 动画 | 时长 | 缓动 | 说明 |
|------|------|------|------|
| 任务卡片悬停 | 150ms | ease-out | 背景色变化 |
| 日志行进入 | 即时 | - | 新日志直接追加 |
| Tab 切换 | 200ms | ease-in-out | 内容淡入淡出 |
| ErrorCard 出现 | 200ms | ease-in-out | 从下方滑入 + 淡入 |
| FixDialog 弹出 | 250ms | cubic-bezier(0.4, 0, 0.2, 1) | 从中心放大 + 淡入 |
| AI Agent 状态变化 | 300ms | ease-in-out | 图标切换 + 颜色过渡 |
| Training Monitor 展开 | 200ms | ease-in-out | 从底部滑入 |
| 搜索高亮 | 即时 | - | 匹配文字闪烁一下 |
| 过滤切换 | 200ms | ease-in-out | 日志列表淡入淡出 |
| 进度条更新 | 300ms | ease-out | 宽度平滑变化 |

---

# 8. 状态设计

## 8.1 日志流状态

| 状态 | 表现 |
|------|------|
| 实时接收 | 新日志不断添加，自动滚动，Agent Status: Idle |
| 暂停滚动 | 固定在当前视图，显示"回到最新"浮动按钮 |
| 分析中 | Agent Status: Analyzing，ErrorCard 骨架屏 |
| 修复中 | Agent Status: Executing，Apply Fix 按钮 Loading |
| 断开 | 显示"日志流已断开"提示 + 重连按钮 |

## 8.2 空状态

| 场景 | 图标 | 标题 | 描述 |
|------|------|------|------|
| 无任务 | `inbox` | No Tasks | 运行工作流或训练任务后将在此处显示 |
| 无日志 | `file-text` | No Logs | 选中任务后查看日志 |
| 无错误 | `check-circle` | No Errors | 当前任务没有检测到错误 |
| AI 未分析 | `sparkles` | AI Analysis Ready | 点击 AI Analyze 开始分析日志 |

## 8.3 加载状态

| 场景 | 表现 |
|------|------|
| 加载任务列表 | 骨架屏（3 行卡片骨架） |
| 加载日志 | 骨架屏（5 行日志行骨架） |
| AI 分析中 | ErrorCard 骨架屏 + Agent Status 脉冲 |
| 修复执行中 | Apply Fix 按钮 Loading 旋转 |

---

# 9. 日志存储设计

日志按照项目保存，结构如下：

```
Project/
├── logs/
│   ├── workflow/
│   │   ├── wf-001.json        ← 工作流执行日志
│   │   └── wf-002.json
│   ├── training/
│   │   ├── task-001.json      ← 训练任务日志 + 指标
│   │   └── task-002.json
│   ├── system/
│   │   ├── 2026-07-07.json    ← 按日期存储系统日志
│   │   └── 2026-07-06.json
│   └── error/
│       ├── analysis-001.json  ← AI 错误分析结果
│       └── analysis-002.json
```

---

# 10. Vue 组件拆分

```
Logs/
├── LogCenter.vue            ← 主页面容器，四区布局
├── TaskList.vue             ← 左侧任务列表
├── LogViewer.vue            ← 中间日志工作区（Tab 容器）
├── RawLogViewer.vue         ← Raw Log 视图
├── AIAnalysisPanel.vue      ← 右侧 AI 分析面板
├── ErrorCard.vue            ← 错误分析卡片
├── Timeline.vue             ← 工作流执行时间线
├── TrainingMonitor.vue      ← 训练监控面板
├── LogFilter.vue            ← 日志过滤器下拉
└── FixDialog.vue            ← 修复确认弹窗
```

### 组件职责

| 组件 | 职责 |
|------|------|
| LogCenter | 四区布局容器，管理面板折叠/展开，协调子组件通信 |
| TaskList | 任务列表展示、搜索、状态分组、点击选中 |
| LogViewer | Human Log / Raw Log Tab 切换，日志内容渲染 |
| RawLogViewer | 原始日志渲染、行号、搜索高亮、复制/下载 |
| AIAnalysisPanel | AI 分析结果展示、ErrorCard 列表、Agent 状态 |
| ErrorCard | 单个错误分析卡片：问题、原因、方案、操作按钮 |
| Timeline | 工作流节点执行时间线、状态、进度 |
| TrainingMonitor | 训练指标卡片、Loss/Accuracy 曲线图 |
| LogFilter | 级别/来源/时间过滤下拉面板 |
| FixDialog | 修复确认弹窗：命令预览、影响说明、确认/取消 |

---

# 11. API 交互设计

| API | Method | 用途 |
|-----|--------|------|
| `/tasks` | GET | 获取任务列表 |
| `/tasks/:id` | GET | 获取任务详情 |
| `/tasks/:id/logs` | GET | 获取任务日志（分页） |
| `/tasks/:id/logs/stream` | WebSocket | 实时日志流 |
| `/tasks/:id/analyze` | POST | AI 分析任务日志 |
| `/tasks/:id/fix` | POST | 执行修复方案 |
| `/tasks/:id/metrics` | GET | 获取训练指标 |
| `/tasks/:id/metrics/stream` | WebSocket | 实时训练指标流 |
| `/tasks/:id/timeline` | GET | 获取工作流时间线 |
| `/logs/export` | POST | 导出日志 |
| `/logs/clear` | POST | 清空日志 |

---

# 12. 页面跳转关系

| 来源 | 目标 | 触发 |
|------|------|------|
| Dashboard | Logs Center | 点击日志入口 |
| Workflow Editor | Logs Center | 点击 Console 面板的"查看完整日志" |
| Plugin Center | Logs Center | 安装任务失败时跳转 |
| AI Chat | Logs Center | AI 建议查看日志详情 |
| Logs Center | Workflow Editor | 点击 Timeline 节点跳转 |
| Logs Center | Plugin Center | Apply Fix 涉及插件操作时跳转 |

---

# 13. 后续扩展

1. **日志告警规则**：支持关键词/模式告警，条件触发桌面通知
2. **日志仪表盘**：可视化日志统计（错误率趋势、任务成功率、平均耗时）
3. **日志对比**：支持两次运行的日志 Diff 对比
4. **日志关联链**：自动关联多条相关日志，形成问题因果链
5. **远程日志**：支持查看远程部署节点的运行日志
6. **自定义 Human Log 模板**：允许插件开发者自定义日志的人类可读转换规则
7. **日志回放**：支持回放历史任务的日志执行过程
8. **协作标注**：团队成员可对错误日志添加标注和评论