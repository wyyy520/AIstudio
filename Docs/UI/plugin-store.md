# AIStudio Plugin Center 页面设计规范

# 1. 页面定位

**Plugin Center 是 AIStudio 的 AI 能力生态管理中枢，不是应用商店。**

插件在 AIStudio 中不是普通软件扩展，而是 **AI 能力模块** —— 它们代表模型能力、Workflow 节点、Agent 工具、环境配置和 MCP 工具。安装一个插件意味着：下载源码、安装依赖、配置环境、下载模型、注册 Workflow 节点、提供 Agent 调用能力。

Plugin Center 的核心使命是让用户 **感知、获取、管理和观察 AI 能力**，而非"浏览和购买应用"。

### 设计参考

| 参考产品 | 借鉴点 |
|----------|--------|
| VS Code Extension Marketplace | 左侧分类树 + 中间详情的三栏结构 |
| JetBrains Plugin Marketplace | 插件卡片的信息密度与层级 |
| Codex Desktop | 任务流式安装过程展示 |
| Claude Desktop | 简洁的深色桌面工具风格 |
| Cursor | 面板化布局与 Agent 集成 |
| Docker Desktop | 状态驱动的卡片与容器管理 |

### 与旧版"插件市场"的核心区别

| 维度 | 旧版（应用商店） | 新版（能力管理中心） |
|------|------------------|----------------------|
| 定位 | 浏览和安装插件 | 管理 AI 能力生态 |
| 布局 | 双栏（分类 + 网格） | 三栏（Explorer + Detail + Agent） |
| 安装展示 | 进度条 | 任务流卡片（可展开日志） |
| Agent 集成 | 无 | 右侧 Agent Panel 实时展示能力映射 |
| 信息焦点 | 评分、下载量、评论 | Capabilities、Workflow Nodes、Agent Tools |
| 交互模式 | 点击卡片 → 弹窗 | 点击卡片 → 中间面板详情 |

---

# 2. 用户流程

## 2.1 核心用户角色

| 角色 | 典型场景 |
|------|----------|
| AI 工程师 | 安装 YOLOv8，在 Workflow 中使用检测节点，让 Agent 调用推理能力 |
| 算法研究员 | 安装 Transformer + LSTM，组合实验不同模型 Pipeline |
| 仿真工程师 | 安装 SUMO/VISSIM，配置仿真环境，Agent 自动编排仿真流程 |
| 系统管理员 | 管理 CUDA/Docker 等系统插件，监控环境状态 |

## 2.2 主流程

```
浏览插件 ──→ 查看详情 ──→ 安装插件 ──→ 验证能力 ──→ 使用插件
    │              │            │             │            │
    │              │            │             │            ├─→ Workflow 节点可用
    │              │            │             │            ├─→ Agent 工具可用
    │              │            │             │            └─→ MCP 工具可用
    │              │            │             │
    │              │            └─→ 任务流展示安装过程
    │              │
    │              ├─→ Capabilities
    │              ├─→ Workflow Nodes
    │              ├─→ Dependencies
    │              └─→ Agent Tools
    │
    ├─→ 按分类浏览
    ├─→ 搜索插件
    └─→ 按状态筛选
```

## 2.3 自动安装流程

```
Workflow 编排 ──→ 拖入节点 ──→ 检测到缺失插件
                                    │
                                    ▼
                            ┌───────────────────┐
                            │  Missing Plugin    │
                            │                    │
                            │  Current workflow  │
                            │  requires:         │
                            │                    │
                            │  ● YOLOv8          │
                            │                    │
                            │  Install           │
                            │  automatically?    │
                            │                    │
                            │  [Install] [Cancel]│
                            └───────────────────┘
                                    │
                                    ▼ Install
                            安装任务流启动
                                    │
                                    ▼ 完成
                            节点自动加入 Node Library
                            Agent Tools 自动注册
```

---

# 3. 页面布局

## 3.1 整体三栏布局

```
┌──────────────────────────────────────────────────────────────────────┐
│  Toolbar: [← →] Plugin Center              🔍 Search...      [⚙]   │
├──────────────┬───────────────────────────┬───────────────────────────┤
│              │                           │                           │
│  Plugin      │  Plugin                   │  Agent                    │
│  Explorer    │  Detail                   │  Panel                    │
│              │                           │                           │
│  240px       │  flex-1                   │  320px                    │
│  (可折叠)     │  (自适应)                  │  (可折叠)                  │
│              │                           │                           │
│              │                           │                           │
│              │                           │                           │
│              │                           │                           │
│              │                           │                           │
│              │                           │                           │
└──────────────┴───────────────────────────┴───────────────────────────┘
```

## 3.2 区域尺寸规范

| 区域 | 默认宽度 | 最小宽度 | 最大宽度 | 可调整 | 可折叠 |
|------|----------|----------|----------|--------|--------|
| Plugin Explorer | 240px | 200px | 320px | 拖拽 | 是（折叠为 48px 图标栏） |
| Plugin Detail | flex-1 | 400px | - | 自适应 | 否 |
| Agent Panel | 320px | 260px | 400px | 拖拽 | 是（折叠为 48px 图标栏） |

## 3.3 分割线

- 栏与栏之间使用可拖拽分割线
- 分割线宽度：1px
- 颜色：`border-subtle`
- 悬停时高亮为 `primary`，宽度变为 3px，提示可拖拽

---

# 4. 组件设计

## 4.1 Toolbar

**位置**：页面顶部，横跨三栏

**高度**：48px

**内容**：

| 位置 | 元素 | 说明 |
|------|------|------|
| 左侧 | 导航按钮 | `[←]` `[→]` 前进/后退浏览历史 |
| 左侧 | 标题 | "Plugin Center"，字号 `h2`，字重 600 |
| 中间 | SearchBar | 搜索框，详见 4.9 |
| 右侧 | 设置按钮 | 插件源管理、代理配置、环境变量 |

**样式**：
- 背景：`bg-secondary`
- 底部边框：`border-subtle` 1px
- 内边距：0 16px

---

## 4.2 PluginExplorer（左侧面板）

**宽度**：240px（默认）

**参考**：VS Code Extension 侧边栏 + JetBrains Plugin 分类树

### 结构

```
┌──────────────────────────┐
│  🔍 Search plugins...    │  ← SearchBar（内嵌）
├──────────────────────────┤
│  ▼ AI Vision        (5)  │  ← 分类折叠组
│    ● YOLOv8        8.2   │  ← 插件列表项
│    ● SAM           1.2   │
│    ○ RT-DETR       2.0   │
│    ● CNN           3.1   │
│    ○ OCR           1.0   │
│  ▶ NLP             (3)  │
│  ▶ Time Series     (2)  │
│  ▶ Speech          (1)  │
│  ▶ Simulation      (2)  │
│  ▶ System          (3)  │
│  ▶ MCP             (4)  │
├──────────────────────────┤
│  ▼ Installed        (6)  │  ← 状态筛选组
│  ▶ Updates Available (1) │
│  ▶ Errors           (0)  │
└──────────────────────────┘
```

### 分类定义

| 分类 | 图标 | 类型色 | 子分类示例 |
|------|------|--------|-----------|
| AI Vision | `eye` | `vision` (#ec4899) | YOLO, CNN, SAM, OCR, RT-DETR |
| NLP | `type` | `nlp` (#3b82f6) | Transformer, LLM, BERT |
| Time Series | `clock` | `timeseries` (#10b981) | LSTM, GRU, Prophet |
| Speech | `mic` | `nlp` (#3b82f6) | Whisper, TTS, STT |
| Simulation | `box` | `simulation` (#06b6d4) | SUMO, VISSIM, MATLAB |
| System | `terminal` | `system` (#6b7280) | CUDA, Docker, Python, Git |
| MCP | `link` | `mcp` (#8b5cf6) | Tools, Servers |

### 插件列表项

每个插件显示：

| 元素 | 样式 |
|------|------|
| 状态指示器 | 左侧圆点，6px，颜色见状态定义 |
| 插件名称 | `body` 字号，`text-primary` |
| 版本号 | `caption` 字号，`text-tertiary`，右对齐 |

**列表项交互**：
- 悬停：背景变为 `bg-hover`
- 点击：背景变为 `bg-active`，中间面板加载详情
- 当前选中：左侧 2px `primary` 色竖线指示

### 状态筛选组

| 筛选项 | 图标 | 说明 |
|--------|------|------|
| Installed | `check-circle` | 显示所有已安装插件 |
| Updates Available | `refresh` | 显示有可用更新的插件，附带更新数量徽章 |
| Errors | `alert-circle` | 显示安装/运行出错的插件，附带错误数量徽章 |

---

## 4.3 PluginCard（插件卡片）

**参考**：Codex 任务卡 + JetBrains 插件卡片

**用于**：Plugin Explorer 中选中分类后的列表视图，或搜索结果

### 卡片布局

```
┌─────────────────────────────────────┐
│                                     │
│   [图标]                            │
│   48×48                             │
│                                     │
│   YOLOv8                            │  ← 名称，h3
│   Vision AI                         │  ← 分类标签，caption，vision 色
│                                     │
│   Version 8.2                       │  ← 版本，caption
│   CUDA Supported                    │  ← 关键特性标签
│                                     │
│   ┌──────────┐                      │
│   │ Installed │                     │  ← StatusBadge
│   └──────────┘                      │
│                                     │
└─────────────────────────────────────┘
```

### 卡片尺寸

| 模式 | 宽度 | 高度 |
|------|------|------|
| 紧凑（Explorer 列表内） | 自适应父容器 | 80px |
| 标准（网格视图） | 200px | 180px |

### 卡片样式

| 属性 | 值 |
|------|-----|
| 圆角 | 16px |
| 背景 | `bg-tertiary` |
| 内边距 | 16px |
| 阴影 | `0 1px 3px rgba(0,0,0,0.2)` |
| 悬停阴影 | `0 4px 12px rgba(0,0,0,0.3)` |
| 悬停位移 | `translateY(-2px)` |
| 过渡 | `all 150ms ease` |

### 紧凑模式布局（Explorer 列表）

```
┌──────────────────────────────────────────────┐
│  [图标]  YOLOv8                  v8.2  [●]   │
│  32×32   Vision AI                    Installed│
└──────────────────────────────────────────────┘
```

- 图标：32×32
- 名称：`body-sm`，`text-primary`
- 分类：`caption`，类型色
- 版本：`caption`，`text-tertiary`
- 状态徽章：右侧对齐

---

## 4.4 PluginDetail（中间面板）

**参考**：VS Code Extension 详情页 + JetBrains 插件详情

### 顶部信息区

```
┌──────────────────────────────────────────────────────────┐
│                                                          │
│  [图标 64×64]   YOLOv8                                   │
│                 Vision AI · v8.2.0                        │
│                 by ultralytics                            │
│                 Source: GitHub                            │
│                                                          │
│  ┌─────────┐ ┌────────┐ ┌─────────┐ ┌───────────┐       │
│  │ Install  │ │ Update │ │ Remove  │ │ Configure │       │
│  └─────────┘ └────────┘ └─────────┘ └───────────┘       │
│                                                          │
│  ┌──────────┐                                           │
│  │ Installed │  ← StatusBadge                           │
│  └──────────┘                                           │
│                                                          │
├──────────────────────────────────────────────────────────┤
```

**按钮状态逻辑**：

| 插件状态 | Install | Update | Remove | Configure |
|----------|---------|--------|--------|-----------|
| Not Installed | Primary 可用 | 隐藏 | 隐藏 | 隐藏 |
| Installed | 隐藏 | 有更新时可见 | Danger 可用 | 可用 |
| Updating | Disabled | Disabled | Disabled | Disabled |
| Error | 可用（重试） | 隐藏 | 可用 | 可用 |

### 信息区域（可滚动）

采用分段卡片式布局，每段一个信息模块：

#### Description 段

```
┌──────────────────────────────────────────────────────────┐
│  Description                                             │
│                                                          │
│  YOLOv8 object detection model.                          │
│  State-of-the-art real-time object detection             │
│  and image segmentation model.                           │
└──────────────────────────────────────────────────────────┘
```

#### Capabilities 段

```
┌──────────────────────────────────────────────────────────┐
│  Capabilities                                            │
│                                                          │
│  ✓ Image Detection                                       │
│  ✓ Training                                              │
│  ✓ Export                                                 │
│  ✓ Inference                                              │
│  ✗ Segmentation                         ← 不可用灰色显示  │
└──────────────────────────────────────────────────────────┘
```

- ✓ 可用：`success` 色
- ✗ 不可用：`text-disabled` 色

#### Supported Workflow Nodes 段

```
┌──────────────────────────────────────────────────────────┐
│  Supported Workflow Nodes                                │
│                                                          │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐     │
│  │ [●] YOLO     │ │ [●] YOLO    │ │ [●] YOLO     │     │
│  │     Train    │ │     Detect  │ │     Export    │     │
│  └──────────────┘ └──────────────┘ └──────────────┘     │
│                                                          │
│  点击节点 → 跳转至 Workflow 编辑器并定位该节点            │
└──────────────────────────────────────────────────────────┘
```

- 每个节点以小卡片展示
- 节点卡片左侧带类型色圆点
- 点击节点卡片跳转至 Workflow 编辑器

#### Dependencies 段

```
┌──────────────────────────────────────────────────────────┐
│  Dependencies                                            │
│                                                          │
│  ┌────────────────────────────────────────────────┐      │
│  │  Python     ≥3.8      ✓ Installed              │      │
│  │  PyTorch    ≥2.0      ✓ Installed              │      │
│  │  CUDA       ≥11.8     ✓ Installed              │      │
│  │  OpenCV     ≥4.8      ✗ Not Installed           │      │
│  └────────────────────────────────────────────────┘      │
└──────────────────────────────────────────────────────────┘
```

- 依赖项以列表展示
- 每项显示：名称、版本要求、当前状态
- 状态：✓ 已满足（`success`）、✗ 未满足（`error`）、⚠ 版本不匹配（`warning`）

### 信息段样式

| 属性 | 值 |
|------|-----|
| 段间距 | 16px |
| 段标题 | `h3`，`text-primary` |
| 段内容 | `body`，`text-secondary` |
| 段卡片圆角 | 12px |
| 段卡片背景 | `bg-tertiary` |
| 段卡片内边距 | 16px |

---

## 4.5 InstallTask（安装任务流）

**参考**：Codex Agent Task 执行界面

**核心原则**：不要简单显示 loading，要展示 **任务流**，让用户感知安装的每一步。

### 任务流布局

```
┌──────────────────────────────────────────────────────────┐
│  Installing YOLOv8                              [✕ Cancel]│
│                                                          │
│  ┌──────────────────────────────────────────────────┐    │
│  │  ✓ Checking Python                    0.3s       │    │
│  └──────────────────────────────────────────────────┘    │
│  ┌──────────────────────────────────────────────────┐    │
│  │  ✓ Checking CUDA                      0.2s       │    │
│  └──────────────────────────────────────────────────┘    │
│  ┌──────────────────────────────────────────────────┐    │
│  │  ✓ Creating Environment               1.2s       │    │
│  │  ┌──────────────────────────────────────────┐    │    │
│  │  │  $ python -m venv .venv                  │    │    │  ← 展开日志
│  │  │  Creating virtual environment...          │    │    │
│  │  │  Done.                                    │    │    │
│  │  └──────────────────────────────────────────┘    │    │
│  └──────────────────────────────────────────────────┘    │
│  ┌──────────────────────────────────────────────────┐    │
│  │  ✓ Installing PyTorch                 12.5s      │    │
│  └──────────────────────────────────────────────────┘    │
│  ┌──────────────────────────────────────────────────┐    │
│  │  ▶ Downloading Model                  3.2s...    │    │  ← 当前步骤
│  │  ┌──────────────────────────────────────────┐    │    │
│  │  │  Downloading yolov8n.pt... 67%           │    │    │  ← 展开日志
│  │  │  ████████████░░░░░░  34.2MB / 51.1MB     │    │    │
│  │  └──────────────────────────────────────────┘    │    │
│  └──────────────────────────────────────────────────┘    │
│  ┌──────────────────────────────────────────────────┐    │
│  │  ○ Registering Workflow Node                     │    │  ← 待执行
│  └──────────────────────────────────────────────────┘    │
│  ┌──────────────────────────────────────────────────┐    │
│  │  ○ Registering Agent Tools                       │    │  ← 待执行
│  └──────────────────────────────────────────────────┘    │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

### 任务步骤状态

| 状态 | 图标 | 颜色 | 说明 |
|------|------|------|------|
| Completed | ✓ | `success` | 步骤完成，显示耗时 |
| In Progress | ▶ | `primary` | 当前执行中，带脉冲动画 |
| Pending | ○ | `text-tertiary` | 等待执行 |
| Failed | ✗ | `error` | 执行失败，可展开查看错误 |

### 任务卡片样式

| 属性 | 值 |
|------|-----|
| 圆角 | 12px |
| 背景 | `bg-tertiary` |
| 内边距 | 12px 16px |
| 间距 | 8px |
| 展开日志背景 | `bg-primary` |
| 展开日志圆角 | 8px |
| 日志字体 | `font-family-mono`，`code` 字号 |

### 任务步骤定义

安装一个插件的标准任务流：

| 序号 | 步骤 | 说明 |
|------|------|------|
| 1 | Checking Dependencies | 检查依赖是否满足 |
| 2 | Checking Environment | 检查运行环境（Python、CUDA 等） |
| 3 | Creating Environment | 创建虚拟环境 |
| 4 | Installing Dependencies | 安装 Python/系统依赖包 |
| 5 | Downloading Source | 下载插件源码 |
| 6 | Downloading Model | 下载预训练模型（如适用） |
| 7 | Configuring Plugin | 执行插件配置脚本 |
| 8 | Registering Workflow Nodes | 注册 Workflow 节点到 Node Library |
| 9 | Registering Agent Tools | 注册 Agent 工具到 Tool Registry |
| 10 | Verifying Installation | 验证安装完整性 |

### 交互

- 点击任务卡片：展开/折叠日志
- 当前步骤自动展开日志
- 已完成步骤默认折叠，点击可展开
- 失败步骤自动展开，显示错误信息
- Cancel 按钮：取消当前及后续步骤

---

## 4.6 DependencyList（依赖列表）

**用于**：PluginDetail 的 Dependencies 段、InstallTask 的依赖检查步骤

### 布局

```
┌────────────────────────────────────────────┐
│  Python      ≥3.8      ✓ 3.11.2 Installed  │
│  PyTorch     ≥2.0      ✓ 2.1.0 Installed   │
│  CUDA        ≥11.8     ✓ 12.1 Available    │
│  OpenCV      ≥4.8      ✗ Not Installed      │
│  numpy       ≥1.24     ⚠ 1.23.0 Mismatch   │
└────────────────────────────────────────────┘
```

### 依赖状态

| 状态 | 图标 | 颜色 | 说明 |
|------|------|------|------|
| Satisfied | ✓ | `success` | 已安装且版本满足 |
| Not Installed | ✗ | `error` | 未安装 |
| Version Mismatch | ⚠ | `warning` | 已安装但版本不满足 |
| Checking | ◌ | `primary` | 正在检查（旋转动画） |

### 行样式

| 属性 | 值 |
|------|-----|
| 行高 | 32px |
| 内边距 | 4px 12px |
| 名称列宽 | 120px |
| 版本要求列宽 | 80px |
| 交替行背景 | 奇数 `bg-tertiary`，偶数 `bg-secondary` |

---

## 4.7 AgentPanel（右侧面板）

**这是 Plugin Center 的核心差异化面板。**

展示 AI Agent 如何使用当前选中的插件，将"插件"与"Agent 能力"建立可视化映射。

**宽度**：320px（默认）

### 结构

```
┌──────────────────────────────────────┐
│  Agent Integration                   │  ← 面板标题
├──────────────────────────────────────┤
│                                      │
│  Agent Status                        │
│  ┌──────────────────────────────┐    │
│  │  ● Ready                     │    │  ← 状态指示
│  └──────────────────────────────┘    │
│                                      │
│  Available Tools                     │
│  ┌──────────────────────────────┐    │
│  │  [●] YOLO Detection          │    │  ← ToolList
│  │  [●] YOLO Training           │    │
│  │  [●] Model Export            │    │
│  └──────────────────────────────┘    │
│                                      │
│  Recent Invocations                  │
│  ┌──────────────────────────────┐    │
│  │  User:                       │    │
│  │  训练车辆检测模型              │    │
│  │                              │    │
│  │  Agent:                      │    │
│  │  Selected: YOLOv8 Plugin     │    │
│  │                              │    │
│  │  Generated Workflow:         │    │
│  │  Dataset → YOLOv8 → Export   │    │  ← 迷你流程图
│  └──────────────────────────────┘    │
│  ┌──────────────────────────────┐    │
│  │  User:                       │    │
│  │  检测图片中的车辆              │    │
│  │                              │    │
│  │  Agent:                      │    │
│  │  Tool: YOLO Detection        │    │
│  │  Result: 12 objects found    │    │
│  └──────────────────────────────┘    │
│                                      │
└──────────────────────────────────────┘
```

### Agent Status 区域

| 状态 | 图标 | 颜色 | 说明 |
|------|------|------|------|
| Ready | ● | `success` | 插件已安装，Agent 可调用 |
| Not Available | ○ | `text-disabled` | 插件未安装，Agent 无法调用 |
| Busy | ● | `warning` | Agent 正在使用该插件 |
| Error | ● | `error` | 插件异常，Agent 调用失败 |

### Recent Invocations 区域

- 显示最近 5 条 Agent 调用记录
- 每条记录以对话卡片展示
- 用户消息：左对齐，`bg-tertiary` 背景
- Agent 响应：左对齐，`primary-bg` 背景
- 迷你流程图：水平排列节点，箭头连接

### 空状态

当插件未安装时：

```
┌──────────────────────────────────────┐
│                                      │
│         [插件图标]                    │
│                                      │
│    Install this plugin to enable     │
│    Agent integration                 │
│                                      │
│    [Install Plugin]                  │
│                                      │
└──────────────────────────────────────┘
```

---

## 4.8 ToolList（工具列表）

**用于**：AgentPanel 的 Available Tools 区域

### 布局

```
┌──────────────────────────────────────┐
│  [●] YOLO Detection                  │  ← 工具项
│      Detect objects in images        │
│  [●] YOLO Training                   │
│      Train custom YOLO models        │
│  [●] Model Export                    │
│      Export model to various formats │
└──────────────────────────────────────┘
```

### 工具项样式

| 元素 | 样式 |
|------|------|
| 类型色圆点 | 6px，对应插件类型色 |
| 工具名称 | `body-sm`，`text-primary` |
| 工具描述 | `caption`，`text-tertiary` |
| 行高 | 40px |
| 悬停 | 背景 `bg-hover` |
| 点击 | 展开工具参数 Schema |

### 工具参数展开

```
┌──────────────────────────────────────┐
│  [●] YOLO Detection                  │
│      Detect objects in images        │
│  ┌──────────────────────────────┐    │
│  │  Parameters:                 │    │
│  │  image     : image  (req)    │    │
│  │  confidence: number (opt)    │    │
│  │  model     : string (opt)    │    │
│  │                              │    │
│  │  Returns:                    │    │
│  │  detections : json           │    │
│  └──────────────────────────────┘    │
└──────────────────────────────────────┘
```

---

## 4.9 StatusBadge（状态徽章）

**用于**：PluginCard、PluginDetail 顶部

### 状态定义

| 状态 | 文本 | 背景色 | 文字色 | 图标 |
|------|------|--------|--------|------|
| Installed | Installed | `success-bg` | `success` | `check-circle` |
| Not Installed | Not Installed | `bg-hover` | `text-secondary` | `download` |
| Updating | Updating | `info-bg` | `info` | `refresh`（旋转动画） |
| Error | Error | `error-bg` | `error` | `alert-circle` |
| Installing | Installing | `primary-bg` | `primary` | `loader`（旋转动画） |

### 徽章样式

| 属性 | 值 |
|------|-----|
| 圆角 | 8px |
| 内边距 | 4px 10px |
| 字号 | `caption` |
| 字重 | 500 |
| 图标与文字间距 | 6px |

---

## 4.10 SearchBar（搜索栏）

**用于**：Toolbar 和 PluginExplorer 顶部

### 布局

```
┌──────────────────────────────────────────────┐
│  🔍  Search plugins...              [✕]      │
└──────────────────────────────────────────────┘
```

### 搜索范围

搜索匹配以下字段：
- 插件名称（权重最高）
- 插件描述
- 分类标签
- Capabilities
- Workflow Node 名称
- Agent Tool 名称

### 交互

| 操作 | 行为 |
|------|------|
| 输入文字 | 实时过滤（debounce 300ms） |
| 清空 | 点击 [✕] 或 Esc 清空搜索 |
| 聚焦 | 边框高亮为 `primary` |
| 无结果 | Explorer 显示空状态提示 |

### 样式

| 属性 | 值 |
|------|-----|
| 高度 | 36px |
| 圆角 | 10px |
| 背景 | `bg-tertiary` |
| 聚焦边框 | `primary` 1px |
| 占位符 | `text-tertiary` |
| 图标颜色 | `text-tertiary` |
| 内边距 | 0 12px |

---

# 5. 数据结构

## 5.1 Plugin

```typescript
interface Plugin {
  id: string
  name: string
  version: string
  author: string
  source: 'github' | 'local' | 'registry'
  sourceUrl: string
  description: string
  category: PluginCategory
  icon: string
  status: PluginStatus
  capabilities: string[]
  workflowNodes: WorkflowNodeDef[]
  dependencies: Dependency[]
  agentTools: AgentToolDef[]
  configSchema: ConfigSchema
  installedAt?: string
  updatedAt?: string
  error?: string
}

type PluginCategory =
  | 'vision'
  | 'nlp'
  | 'timeseries'
  | 'speech'
  | 'simulation'
  | 'system'
  | 'mcp'

type PluginStatus =
  | 'installed'
  | 'not-installed'
  | 'installing'
  | 'updating'
  | 'error'
```

## 5.2 Dependency

```typescript
interface Dependency {
  name: string
  versionRequired: string
  versionInstalled?: string
  status: 'satisfied' | 'not-installed' | 'version-mismatch' | 'checking'
}
```

## 5.3 InstallTask

```typescript
interface InstallTask {
  id: string
  pluginId: string
  pluginName: string
  status: 'running' | 'completed' | 'failed' | 'cancelled'
  steps: InstallStep[]
  startedAt: string
  completedAt?: string
}

interface InstallStep {
  id: string
  name: string
  status: 'completed' | 'in-progress' | 'pending' | 'failed'
  startedAt?: string
  completedAt?: string
  duration?: number
  logs: LogEntry[]
  error?: string
}

interface LogEntry {
  timestamp: string
  level: 'info' | 'warn' | 'error' | 'debug'
  message: string
}
```

## 5.4 AgentToolDef

```typescript
interface AgentToolDef {
  name: string
  description: string
  parameters: ToolParameter[]
  returns: string
}

interface ToolParameter {
  name: string
  type: string
  required: boolean
  description?: string
  default?: any
}
```

## 5.5 WorkflowNodeDef

```typescript
interface WorkflowNodeDef {
  name: string
  type: string
  category: PluginCategory
  inputs: PortDef[]
  outputs: PortDef[]
}

interface PortDef {
  name: string
  type: 'image' | 'text' | 'number' | 'json' | 'file' | 'model'
  required: boolean
  description: string
}
```

## 5.6 AgentInvocation

```typescript
interface AgentInvocation {
  id: string
  pluginId: string
  userMessage: string
  agentResponse: string
  toolUsed?: string
  workflowGenerated?: string[]
  result?: string
  timestamp: string
}
```

## 5.7 PluginCategoryGroup

```typescript
interface PluginCategoryGroup {
  category: PluginCategory
  label: string
  icon: string
  color: string
  plugins: Plugin[]
}
```

---

# 6. 交互逻辑

## 6.1 插件选择

| 操作 | 行为 |
|------|------|
| 点击 Explorer 中的插件 | 中间面板加载详情，右侧面板加载 Agent 信息 |
| 点击分类 | Explorer 展开该分类，折叠其他分类 |
| 点击状态筛选 | Explorer 仅显示符合状态的插件 |
| 搜索 | Explorer 实时过滤，跨分类搜索 |

## 6.2 安装交互

| 操作 | 行为 |
|------|------|
| 点击 Install | PluginDetail 区域切换为 InstallTask 视图 |
| InstallTask 运行中 | 操作按钮全部 Disabled |
| InstallTask 完成 | 自动切换回 PluginDetail，状态更新为 Installed |
| InstallTask 失败 | 显示失败步骤，提供 Retry 按钮 |
| 点击 Cancel | 弹出确认对话框，确认后取消安装 |

## 6.3 更新交互

| 操作 | 行为 |
|------|------|
| 点击 Update | PluginDetail 切换为 InstallTask（更新模式） |
| 更新完成 | 状态更新，版本号更新 |
| Explorer 中 Updates Available | 显示更新数量徽章 |

## 6.4 卸载交互

| 操作 | 行为 |
|------|------|
| 点击 Remove | 弹出确认对话框 |
| 确认卸载 | 执行卸载，显示卸载任务流 |
| 卸载完成 | 状态变为 Not Installed，Workflow 节点移除，Agent Tools 注销 |

## 6.5 配置交互

| 操作 | 行为 |
|------|------|
| 点击 Configure | 右侧 Agent Panel 切换为配置表单 |
| 修改配置 | 实时保存（debounce 500ms） |
| 配置验证 | 提交时验证，错误字段标红 |

## 6.6 Workflow 联动

| 场景 | 行为 |
|------|------|
| Workflow 中拖入未安装插件的节点 | 弹出 Missing Plugin 对话框 |
| 确认安装 | 跳转 Plugin Center，自动开始安装 |
| 安装完成 | 节点自动可用，通知用户 |
| Agent 自动安装 | Agent 判断需要插件时，弹出确认后自动安装 |

---

# 7. 安装流程设计

## 7.1 完整安装流程

```
用户点击 Install
        │
        ▼
PluginDetail → InstallTask 视图切换（过渡动画 200ms）
        │
        ▼
┌─ Step 1: Checking Dependencies ──────────────────────────┐
│  检查所有 Dependency 状态                                  │
│  全部满足 → 进入 Step 2                                   │
│  有不满足 → 标记失败步骤，提示用户                         │
└──────────────────────────────────────────────────────────┘
        │
        ▼
┌─ Step 2: Checking Environment ───────────────────────────┐
│  检查 Python、CUDA、Docker 等环境                         │
│  全部可用 → 进入 Step 3                                   │
│  不可用 → 标记失败，提示安装环境                           │
└──────────────────────────────────────────────────────────┘
        │
        ▼
┌─ Step 3: Creating Environment ───────────────────────────┐
│  创建虚拟环境 / 容器                                      │
│  成功 → 进入 Step 4                                       │
│  失败 → 标记失败，显示错误日志                             │
└──────────────────────────────────────────────────────────┘
        │
        ▼
┌─ Step 4: Installing Dependencies ────────────────────────┐
│  pip install / apt install 等                              │
│  成功 → 进入 Step 5                                       │
│  失败 → 标记失败，显示错误日志                             │
└──────────────────────────────────────────────────────────┘
        │
        ▼
┌─ Step 5: Downloading Source ─────────────────────────────┐
│  从 GitHub / Registry 下载插件源码                         │
│  成功 → 进入 Step 6                                       │
│  失败 → 标记失败，提供重试                                 │
└──────────────────────────────────────────────────────────┘
        │
        ▼
┌─ Step 6: Downloading Model ──────────────────────────────┐
│  下载预训练模型文件（如适用）                               │
│  成功 → 进入 Step 7                                       │
│  失败 → 标记失败，提供镜像源切换                           │
│  不适用 → 跳过此步骤                                      │
└──────────────────────────────────────────────────────────┘
        │
        ▼
┌─ Step 7: Configuring Plugin ─────────────────────────────┐
│  执行插件配置脚本                                          │
│  成功 → 进入 Step 8                                       │
│  失败 → 标记失败，显示配置错误                             │
└──────────────────────────────────────────────────────────┘
        │
        ▼
┌─ Step 8: Registering Workflow Nodes ─────────────────────┐
│  将插件提供的节点注册到 Node Library                       │
│  注册后 Workflow 编辑器立即可用                             │
└──────────────────────────────────────────────────────────┘
        │
        ▼
┌─ Step 9: Registering Agent Tools ────────────────────────┐
│  将插件提供的工具注册到 Agent Tool Registry                │
│  注册后 Agent 立即可调用                                   │
└──────────────────────────────────────────────────────────┘
        │
        ▼
┌─ Step 10: Verifying Installation ────────────────────────┐
│  运行验证脚本，确认安装完整                                 │
│  通过 → 安装完成，切换回 PluginDetail                      │
│  失败 → 标记警告，提示手动验证                             │
└──────────────────────────────────────────────────────────┘
```

## 7.2 Missing Plugin 自动安装对话框

当 Workflow 需要 未安装插件时弹出：

```
┌──────────────────────────────────────────────┐
│                                              │
│  Missing Plugin                              │
│                                              │
│  Current workflow requires:                  │
│                                              │
│  [图标] YOLOv8                               │
│         Vision AI · v8.2                     │
│                                              │
│  This will automatically:                    │
│  • Download source code                      │
│  • Install dependencies                      │
│  • Download model weights                    │
│  • Register workflow nodes                   │
│  • Register agent tools                      │
│                                              │
│            [Install]       [Cancel]          │
│                                              │
└──────────────────────────────────────────────┘
```

- 弹窗圆角：16px
- 背景：`bg-secondary`
- Install 按钮：Primary 类型
- Cancel 按钮：Ghost 类型

---

# 8. Agent 调用流程

## 8.1 Agent 使用插件的完整流程

```
用户向 Agent 发出指令
        │
        ▼
Agent 解析意图，判断需要哪些工具
        │
        ▼
Agent 查询 Tool Registry
        │
        ├─→ 工具已注册（插件已安装）→ 直接调用
        │
        └─→ 工具未注册（插件未安装）→ 弹出 Missing Plugin 对话框
                │
                ▼ 用户确认安装
            Plugin Center 安装插件
                │
                ▼ 安装完成
            工具自动注册到 Tool Registry
                │
                ▼
            Agent 继续执行，调用工具
```

## 8.2 Agent Panel 中的调用展示

### 场景 1：Agent 直接调用工具

```
┌──────────────────────────────────────┐
│  User:                               │
│  检测这张图片中的车辆                  │
│                                      │
│  Agent:                              │
│  Using tool: YOLO Detection          │
│                                      │
│  Input:                              │
│  image: traffic.jpg                  │
│  confidence: 0.5                     │
│                                      │
│  Output:                             │
│  12 objects detected                 │
│  - car: 8                            │
│  - truck: 3                          │
│  - bus: 1                            │
│                                      │
│  Duration: 1.2s                      │
└──────────────────────────────────────┘
```

### 场景 2：Agent 生成 Workflow

```
┌──────────────────────────────────────┐
│  User:                               │
│  训练车辆检测模型                      │
│                                      │
│  Agent:                              │
│  Selected: YOLOv8 Plugin             │
│                                      │
│  Generated Workflow:                 │
│                                      │
│  [Dataset] → [YOLOv8 Train] → [Export]│  ← 迷你流程图
│                                      │
│  [Open in Workflow Editor]           │  ← 跳转按钮
└──────────────────────────────────────┘
```

### 场景 3：Agent 需要安装插件

```
┌──────────────────────────────────────┐
│  User:                               │
│  分析这段语音的内容                    │
│                                      │
│  Agent:                              │
│  Required plugin not found:          │
│  Whisper (Speech)                    │
│                                      │
│  [Install & Continue]  [Cancel]      │
└──────────────────────────────────────┘
```

---

# 9. Vue 组件拆分

## 9.1 组件目录

```
PluginCenter/
├── PluginExplorer.vue        # 左侧面板：分类树 + 插件列表
├── PluginCard.vue            # 插件卡片（紧凑/标准两种模式）
├── PluginDetail.vue          # 中间面板：插件详情页
├── InstallTask.vue           # 安装任务流视图
├── DependencyList.vue        # 依赖列表
├── AgentPanel.vue            # 右侧面板：Agent 集成信息
├── ToolList.vue              # Agent 工具列表
├── StatusBadge.vue           # 状态徽章
└── SearchBar.vue             # 搜索栏
```

## 9.2 组件职责与 Props/Emits

### PluginExplorer.vue

**职责**：左侧面板，展示分类树和插件列表

| Props | 类型 | 说明 |
|-------|------|------|
| categories | `PluginCategoryGroup[]` | 分类数据 |
| plugins | `Plugin[]` | 当前筛选的插件列表 |
| selectedPluginId | `string \| null` | 当前选中的插件 ID |
| searchQuery | `string` | 搜索关键词 |

| Emits | 参数 | 说明 |
|-------|------|------|
| select-plugin | `pluginId: string` | 选中某个插件 |
| select-category | `category: PluginCategory` | 选中某个分类 |
| search | `query: string` | 搜索输入 |

### PluginCard.vue

**职责**：插件卡片，支持紧凑和标准两种模式

| Props | 类型 | 说明 |
|-------|------|------|
| plugin | `Plugin` | 插件数据 |
| mode | `'compact' \| 'standard'` | 显示模式 |
| selected | `boolean` | 是否选中 |

| Emits | 参数 | 说明 |
|-------|------|------|
| click | - | 点击卡片 |
| install | `pluginId: string` | 点击安装 |
| remove | `pluginId: string` | 点击卸载 |

### PluginDetail.vue

**职责**：中间面板，展示插件详情或安装任务流

| Props | 类型 | 说明 |
|-------|------|------|
| plugin | `Plugin \| null` | 当前插件数据 |
| installTask | `InstallTask \| null` | 当前安装任务 |

| Emits | 参数 | 说明 |
|-------|------|------|
| install | `pluginId: string` | 安装 |
| update | `pluginId: string` | 更新 |
| remove | `pluginId: string` | 卸载 |
| configure | `pluginId: string` | 配置 |
| navigate-workflow | `nodeName: string` | 跳转 Workflow 节点 |

### InstallTask.vue

**职责**：安装任务流视图，展示每一步的状态和日志

| Props | 类型 | 说明 |
|-------|------|------|
| task | `InstallTask` | 安装任务数据 |

| Emits | 参数 | 说明 |
|-------|------|------|
| cancel | `taskId: string` | 取消安装 |
| retry | `taskId: string` | 重试失败的步骤 |

### DependencyList.vue

**职责**：依赖列表，展示依赖项及其状态

| Props | 类型 | 说明 |
|-------|------|------|
| dependencies | `Dependency[]` | 依赖数据 |

无 Emits（纯展示组件）

### AgentPanel.vue

**职责**：右侧面板，展示 Agent 集成信息

| Props | 类型 | 说明 |
|-------|------|------|
| plugin | `Plugin \| null` | 当前插件 |
| agentStatus | `'ready' \| 'not-available' \| 'busy' \| 'error'` | Agent 状态 |
| invocations | `AgentInvocation[]` | 最近调用记录 |

| Emits | 参数 | 说明 |
|-------|------|------|
| install | `pluginId: string` | 从空状态安装 |
| open-workflow | `workflowId: string` | 跳转 Workflow |

### ToolList.vue

**职责**：Agent 工具列表，展示可用工具及参数

| Props | 类型 | 说明 |
|-------|------|------|
| tools | `AgentToolDef[]` | 工具列表 |
| categoryColor | `string` | 类型色 |

| Emits | 参数 | 说明 |
|-------|------|------|
| select-tool | `toolName: string` | 选中工具 |

### StatusBadge.vue

**职责**：状态徽章

| Props | 类型 | 说明 |
|-------|------|------|
| status | `PluginStatus` | 插件状态 |

无 Emits（纯展示组件）

### SearchBar.vue

**职责**：搜索输入框

| Props | 类型 | 说明 |
|-------|------|------|
| modelValue | `string` | 搜索关键词（v-model） |
| placeholder | `string` | 占位文本 |

| Emits | 参数 | 说明 |
|-------|------|------|
| update:modelValue | `value: string` | 输入变化 |
| search | `query: string` | 搜索提交 |
| clear | - | 清空搜索 |

## 9.3 组件层级关系

```
PluginCenter (页面)
├── Toolbar
│   └── SearchBar
├── PluginExplorer (左栏)
│   ├── SearchBar
│   ├── CategoryTree
│   │   └── CategoryGroup (多个)
│   │       └── PluginCard (compact mode) (多个)
│   └── StatusFilter
├── PluginDetail (中栏)
│   ├── PluginHeader
│   │   └── StatusBadge
│   ├── ActionButtons
│   ├── DescriptionSection
│   ├── CapabilitiesSection
│   ├── WorkflowNodesSection
│   ├── DependencyList
│   └── InstallTask (安装时替换上述内容)
└── AgentPanel (右栏)
    ├── AgentStatus
    ├── ToolList
    │   └── ToolItem (多个)
    └── InvocationList
        └── InvocationCard (多个)
```

---

# 动画规范

| 动画 | 时长 | 缓动 | 说明 |
|------|------|------|------|
| 面板折叠/展开 | 200ms | ease | 宽度过渡 |
| 插件选中 | 150ms | ease | 背景色过渡 |
| Detail 视图切换 | 250ms | ease | 淡入淡出 + 轻微位移 |
| InstallTask 步骤状态变更 | 200ms | ease | 图标/颜色过渡 |
| InstallTask 日志展开 | 200ms | ease | 高度过渡 |
| 卡片悬停 | 150ms | ease | 位移 + 阴影 |
| Agent Panel 调用记录新增 | 300ms | ease-out | 从上方滑入 |
| StatusBadge 状态变更 | 200ms | ease | 颜色过渡 |
| 搜索结果过滤 | 200ms | ease | 列表重排 |
| 分类展开/折叠 | 200ms | ease | 高度过渡 |

---

# 错误状态设计

## 安装失败

- 失败步骤自动展开日志
- 日志中错误行以 `error` 色高亮
- 提供 Retry 按钮（重试失败步骤）
- 提供 View Full Log 按钮
- 后续 Pending 步骤显示为 Blocked 状态

## 依赖不满足

- DependencyList 中未满足项标红
- 提供 "Install Missing Dependencies" 按钮
- 点击后自动安装缺失依赖

## 网络错误

- 下载步骤失败时显示网络错误
- 提供 Retry 按钮
- 提供 Change Mirror 按钮（切换下载镜像源）

## 环境缺失

- 环境检查步骤失败时显示缺失项
- 提供 "How to Install" 链接（跳转文档）
- 提供 "Skip & Continue" 按钮（警告）

---

# 空状态设计

## Explorer 无搜索结果

```
┌──────────────────────────┐
│                          │
│     [搜索图标]            │
│                          │
│  No plugins found for    │
│  "xxx"                   │
│                          │
│  Try different keywords  │
│                          │
└──────────────────────────┘
```

## Detail 未选中插件

```
┌──────────────────────────────────────────┐
│                                          │
│           [插件图标]                      │
│                                          │
│     Select a plugin to view details      │
│                                          │
│     Browse categories on the left        │
│                                          │
└──────────────────────────────────────────┘
```

## Agent Panel 插件未安装

```
┌──────────────────────────────────────┐
│                                      │
│         [插件图标]                    │
│                                      │
│    Install this plugin to enable     │
│    Agent integration                 │
│                                      │
│    [Install Plugin]                  │
│                                      │
└──────────────────────────────────────┘
```

---

# 页面跳转关系

| 来源 | 触发 | 目标 |
|------|------|------|
| Dashboard | 点击 Plugin Center 入口 | Plugin Center |
| Workflow Editor | 拖入未安装插件节点 | Plugin Center + Missing Plugin 弹窗 |
| Workflow Editor | 节点面板 "Get More Nodes" | Plugin Center |
| Plugin Center | 点击 Workflow Node 卡片 | Workflow Editor（定位到该节点） |
| Plugin Center | Agent Panel "Open in Workflow Editor" | Workflow Editor |
| Plugin Center | Configure 按钮 | Agent Panel 切换为配置表单 |
| AI Chat | Agent 判断需要插件 | Plugin Center + Missing Plugin 弹窗 |

---

# API 交互

| API | 方法 | 用途 |
|-----|------|------|
| `plugin/list` | GET | 获取插件列表（支持分类、状态筛选） |
| `plugin/detail` | GET | 获取插件详情 |
| `plugin/search` | GET | 搜索插件 |
| `plugin/install` | POST | 安装插件（返回任务 ID） |
| `plugin/uninstall` | POST | 卸载插件 |
| `plugin/update` | POST | 更新插件 |
| `plugin/configure` | PUT | 更新插件配置 |
| `plugin/task/status` | GET | 获取安装任务状态（SSE 推送） |
| `plugin/task/cancel` | POST | 取消安装任务 |
| `plugin/task/retry` | POST | 重试失败步骤 |
| `plugin/dependencies/check` | GET | 检查依赖状态 |
| `agent/tools` | GET | 获取 Agent 可用工具列表 |
| `agent/invocations` | GET | 获取 Agent 调用记录 |

---

# 后续扩展

1. **插件依赖图**：可视化展示插件之间的依赖关系，类似 Docker Compose 依赖图
2. **插件环境隔离**：每个插件独立虚拟环境，避免依赖冲突
3. **插件开发模式**：内置插件开发模板、调试工具、热加载
4. **企业私有源**：支持配置私有插件 Registry
5. **插件能力推荐**：基于当前 Workflow 自动推荐所需插件
6. **插件合集**：官方场景化合集（如"自动驾驶工具链"），一键安装
7. **MCP 服务管理**：MCP 插件提供独立的 Server 管理面板
8. **插件性能监控**：展示插件运行时的资源占用（CPU/GPU/Memory）