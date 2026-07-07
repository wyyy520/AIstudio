# AIStudio 项目管理模块设计规范

# 页面定位

**项目管理模块是 AIStudio 的核心工作空间，负责管理 AI 项目的完整生命周期。**

它不是普通的文件夹管理或后台管理系统。

它是一个 **AI Project Workspace** —— 用户在此创建 AI 项目、编排 Workflow、管理 Dataset、训练 Model、记录 Experiment、配置 Environment，并由 AI Agent 辅助整个开发过程。

参考产品定位矩阵：

| 参考产品 | 借鉴能力 | 优先级 |
|----------|----------|--------|
| VS Code Workspace | 项目资源管理器、文件树、多标签 | ★★★★★ |
| Cursor | AI 辅助开发、上下文感知 | ★★★★★ |
| Codex Desktop | Agent 执行、任务卡片 | ★★★★★ |
| MLFlow | 实验追踪、参数对比、指标可视化 | ★★★★★ |
| HuggingFace Hub | 模型仓库、数据集浏览、社区共享 | ★★★★★ |
| PyCharm Project | 项目结构、运行配置、环境管理 | ★★★★☆ |
| Claude Desktop | 对话式交互、上下文面板 | ★★★★☆ |

---

# 用户目标

| 用户角色 | 核心目标 |
|----------|----------|
| AI 工程师 | 高效管理多个 AI 项目的 Workflow、模型、数据集 |
| 算法研究员 | 追踪训练实验、对比参数、分析指标 |
| 数据科学家 | 管理数据集、执行数据预处理、监控数据质量 |
| 交通仿真工程师 | 配置 SUMO/VISSIM 仿真环境、管理仿真 Workflow |
| 团队负责人 | 概览项目状态、查看训练进度、管理环境 |

---

# 整体页面布局

采用三栏桌面软件布局，参考 VS Code Workspace + Cursor 的空间组织方式。

```
┌──────────────────────────────────────────────────────────────────────┐
│  Toolbar（面包屑 + 搜索 + 操作按钮）                                  │
├────────────┬──────────────────────────────────┬──────────────────────┤
│            │                                  │                      │
│  Project   │       Project Workspace          │   AI Assistant       │
│  Explorer  │       (中间主区域)                │   Panel              │
│            │                                  │                      │
│  240px     │       flex: 1                    │   280px              │
│  (可拖拽)  │                                  │   (可拖拽)           │
│            │                                  │                      │
├────────────┴──────────────────────────────────┴──────────────────────┤
│  Status Bar（项目状态 + 环境状态 + AI Agent 状态）                     │
└──────────────────────────────────────────────────────────────────────┘
```

## 区域职责

### Toolbar（顶部工具栏）

- 高度：44px
- 背景：`bg-secondary`
- 内容：
  - 左侧：面包屑导航（`项目名称 / 资源类型 / 资源名称`）
  - 中间：全局搜索框（搜索项目内所有资源）
  - 右侧：操作按钮组（新建、运行、设置、AI 助手开关）

### Project Explorer（左侧资源管理器）

- 默认宽度：240px，可拖拽调整（180px ~ 360px）
- 背景：`bg-secondary`
- 边框：右侧 `border-subtle`
- 内容：AI 项目资源树，类似 VS Code Explorer 但显示 AI 项目专用资源
- 交互：折叠/展开、拖拽排序、右键菜单、搜索过滤

### Project Workspace（中间工作区域）

- 自适应宽度，占据剩余空间
- 背景：`bg-primary`
- 内容：根据左侧选中的资源类型显示对应工作界面
- 模式：Dashboard / Dataset / Workflow / Model / Experiment / Environment

### AI Assistant Panel（右侧 AI 助手）

- 默认宽度：280px，可拖拽调整（240px ~ 400px）
- 背景：`bg-secondary`
- 边框：左侧 `border-subtle`
- 内容：上下文感知的 AI 助手，理解当前项目状态并提供建议
- 可通过 Toolbar 按钮或快捷键 `Ctrl+Shift+A` 切换显示/隐藏

### Status Bar（底部状态栏）

- 高度：28px
- 背景：`bg-secondary`
- 边框：顶部 `border-subtle`
- 内容：
  - 左侧：项目状态指示（Ready / Running / Training）
  - 中间：当前环境（Python 3.10 · CUDA 11.8 · PyTorch 2.0）
  - 右侧：AI Agent 状态（Idle / Thinking / Executing）

---

# Project Explorer 设计

## 资源树结构

```
▼ AIStudio Traffic Project
  ├── ▶ Workflow (3)
  │   ├── YOLO Detection Pipeline
  │   ├── LSTM Prediction Flow
  │   └── SUMO Simulation Chain
  ├── ▶ Dataset (2)
  │   ├── traffic_detection_v3
  │   └── traffic_flow_csv
  ├── ▶ Models (4)
  │   ├── yolo_traffic_v2.1
  │   ├── lstm_flow_predictor
  │   ├── yolov8n_finetuned
  │   └── resnet50_baseline
  ├── ▶ Experiments (12)
  │   ├── exp_20260701_yolo_train
  │   ├── exp_20260628_lstm_tune
  │   └── ...
  ├── ▶ Environment
  │   ├── Python 3.10.12
  │   ├── CUDA 11.8
  │   └── Dependencies (48)
  ├── ▶ Outputs
  │   ├── checkpoints/
  │   ├── logs/
  │   └── exports/
  └── ▶ Logs (28)
```

## 节点设计

每个树节点结构：

```
[折叠图标 12px] [类型图标 16px] [名称] [计数标签] [操作按钮组]
```

### 节点尺寸与样式

| 属性 | 值 |
|------|-----|
| 节点高度 | 32px |
| 缩进量 | 每层 16px |
| 圆角 | 6px |
| 图标尺寸 | 16px |
| 名称字号 | 13px (`body-sm`) |

### 节点状态

| 状态 | 视觉表现 |
|------|----------|
| Default | `text-secondary` 文字，透明背景 |
| Hover | `bg-hover` 背景，`text-primary` 文字 |
| Selected | `bg-active` 背景 + 左侧 3px `primary` 边条 |
| Dragging | 透明度 0.5，跟随鼠标 |

### 类型图标颜色

| 资源类型 | 图标颜色 | 对应变量 |
|----------|----------|----------|
| Workflow | `#3b82f6` 蓝色 | `--nlp` |
| Dataset | `#10b981` 绿色 | `--timeseries` |
| Model | `#8b5cf6` 紫色 | `--primary` |
| Experiment | `#f59e0b` 橙色 | `--logic` |
| Environment | `#6b7280` 灰色 | `--system` |
| Outputs | `#06b6d4` 青色 | `--simulation` |
| Logs | `#ec4899` 粉色 | `--vision` |

### 计数标签

- 显示在类型名称右侧
- 样式：`bg-active` 背景，`text-tertiary` 文字，圆角 4px，字号 11px
- 例如：`Workflow (3)`

### 右键菜单

**项目级别：**
- 重命名
- 新建 Workflow
- 新建 Dataset
- 导入资源
- 导出项目
- 项目设置
- 在文件管理器中打开
- 删除

**Workflow 级别：**
- 打开编辑
- 运行
- 复制
- 重命名
- 导出
- 删除

**Dataset 级别：**
- 查看详情
- 数据预览
- 格式转换
- 重命名
- 导出
- 删除

**Model 级别：**
- 加载到 Workflow
- 部署
- 导出
- 查看训练记录
- 重命名
- 删除

**Experiment 级别：**
- 查看详情
- 对比实验
- 复制参数
- 导出报告
- 删除

---

# Project Dashboard 设计

中间区域默认显示项目主页，类似 HuggingFace Hub 项目页 + MLFlow 仪表盘。

## 布局结构

```
┌──────────────────────────────────────────────────────┐
│  项目名称 + 状态标签 + 创建时间                         │
│  项目描述文字                                          │
├──────┬──────┬──────┬──────┬──────────────────────────┤
│      │      │      │      │                          │
│ 统计 │ 统计 │ 统计 │ 统计 │     最近活动              │
│ 卡片 │ 卡片 │ 卡片 │ 卡片 │                          │
│      │      │      │      │                          │
├──────┴──────┴──────┴──────┴──────────────────────────┤
│                                                      │
│  快速操作按钮组                                        │
│                                                      │
├──────────────────────────┬───────────────────────────┤
│                          │                           │
│  当前 Workflow 概览       │   训练状态                 │
│  (节点缩略图 + 状态)      │   (最近 Experiment)       │
│                          │                           │
├──────────────────────────┼───────────────────────────┤
│                          │                           │
│  模型仓库概览             │   环境状态                 │
│  (模型列表 + 指标)        │   (GPU + 依赖)            │
│                          │                           │
└──────────────────────────┴───────────────────────────┘
```

## 项目头部

```
┌─────────────────────────────────────────────────────────────┐
│  [项目图标]                                                  │
│                                                             │
│  AIStudio Traffic Project                    [Ready]  ⋯   │
│  城市交通流量预测与仿真项目                                    │
│  创建于 2026-06-15 · 最后更新 2 小时前                         │
└─────────────────────────────────────────────────────────────┘
```

- 项目名称：20px，`font-semibold`，`text-primary`
- 状态标签：12px，圆角 6px，颜色根据状态变化
- 描述：14px，`text-secondary`
- 时间信息：12px，`text-tertiary`

## 统计卡片组

四列等宽网格布局：

```
┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│  🔵 Workflow │ │  🟢 Dataset  │ │  🟣 Model    │ │  🟠 Experiment│
│     3        │ │     2        │ │     4        │ │     12       │
│  个工作流    │ │  个数据集    │ │  个模型      │ │  个实验      │
└──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘
```

### 卡片规格

| 属性 | 值 |
|------|-----|
| 背景 | `bg-tertiary` |
| 圆角 | 12px |
| 内边距 | 16px |
| 最小高度 | 88px |
| 悬停 | 阴影增强 `shadow`，上移 2px |
| 点击 | 跳转到对应管理页面 |

### 卡片内容

- 顶部：类型图标（20px）+ 类型名称（12px，`text-tertiary`）
- 底部：数量（24px，`text-primary`，`font-semibold`）+ 说明（12px，`text-tertiary`）

## 快速操作按钮组

水平排列，间距 8px：

| 按钮 | 图标 | 类型 | 快捷键 |
|------|------|------|--------|
| Run Workflow | ▶ | Primary | `Ctrl+R` |
| Train Model | 🎯 | Secondary | `Ctrl+T` |
| Open Dataset | 📊 | Secondary | `Ctrl+D` |
| Deploy Model | 🚀 | Secondary | `Ctrl+Shift+D` |
| View Experiment | 📈 | Secondary | `Ctrl+E` |

按钮规格：高度 36px，圆角 10px，字号 13px，图标 16px

## 最近活动面板

位于统计卡片右侧，显示最近的操作记录：

```
最近活动
───────────────────────
✓ 训练完成  exp_20260701_yolo_train  2 小时前
▶ 训练中    exp_20260707_lstm_tune   进行中...
✓ 数据上传  traffic_detection_v3     昨天
✗ 训练失败  exp_20260625_resnet      3 天前
```

每条记录：状态图标 + 操作名称 + 资源名称 + 时间

---

# Dataset 管理设计

## Dataset 列表视图

网格/列表双视图切换，参考 HuggingFace Datasets 浏览。

### 网格视图

```
┌────────────────┐ ┌────────────────┐ ┌────────────────┐
│                │ │                │ │                │
│   [数据预览]   │ │   [数据预览]   │ │   [数据预览]   │
│                │ │                │ │                │
├────────────────┤ ├────────────────┤ ├────────────────┤
│ traffic_det_v3 │ │ traffic_flow   │ │ weather_csv    │
│ YOLO · 2.4GB   │ │ CSV · 180MB    │ │ CSV · 45MB     │
│ 800 张 · 4 类  │ │ 12 列 · 50K行  │ │ 8 列 · 10K行   │
│ 2 天前更新     │ │ 1 周前更新     │ │ 3 天前更新     │
└────────────────┘ └────────────────┘ └────────────────┘
```

卡片规格：宽度 240px，圆角 12px，背景 `bg-tertiary`

### 列表视图

| 名称 | 格式 | 大小 | 样本数 | 类别数 | 更新时间 | 操作 |
|------|------|------|--------|--------|----------|------|
| traffic_detection_v3 | YOLO | 2.4 GB | 800 张 | 4 | 2 天前 | ⋯ |
| traffic_flow_data | CSV | 180 MB | 50,000 | - | 1 周前 | ⋯ |
| weather_dataset | CSV | 45 MB | 10,000 | - | 3 天前 | ⋯ |

行高：40px，悬停 `bg-hover`，选中 `bg-active`

## Dataset 详情页

选中某个 Dataset 后显示详情：

```
┌──────────────────────────────────────────────────────────────┐
│  traffic_detection_v3                          [YOLO] [v3]  │
│  城市交通摄像头车辆检测数据集                                   │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐           │
│  │ 图片数量 │ │ 类别数量 │ │ 总大小  │ │ 格式    │           │
│  │  800    │ │    4    │ │ 2.4 GB  │ │ YOLO   │           │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘           │
│                                                              │
├──────────────────────────────────────────────────────────────┤
│  [数据预览]                    │  类别分布                    │
│                               │                              │
│  ┌────┐ ┌────┐ ┌────┐       │  car       ████████  45%     │
│  │img1│ │img2│ │img3│       │  truck     ████      25%     │
│  └────┘ └────┘ └────┘       │  bus       ███       20%     │
│  ┌────┐ ┌────┐ ┌────┐       │  motorcycle █        10%     │
│  │img4│ │img5│ │img6│       │                              │
│  └────┘ └────┘ └────┘       │                              │
│                              │                              │
├──────────────────────────────────────────────────────────────┤
│  操作：[格式转换] [数据增强] [导出] [删除]                      │
└──────────────────────────────────────────────────────────────┘
```

### 支持格式

| 格式 | 说明 | 图标颜色 |
|------|------|----------|
| YOLO | Ultralytics YOLO 格式 | `--timeseries` 绿色 |
| COCO | MS COCO JSON 格式 | `--nlp` 蓝色 |
| VOC | Pascal VOC XML 格式 | `--vision` 粉色 |
| CSV | 通用表格数据 | `--logic` 橙色 |

### 数据预览区

- 图片数据集：缩略图网格（6 列），悬停放大
- 表格数据集：表格前 100 行预览，支持排序和筛选
- 标注预览：在图片上叠加显示标注框

---

# Workflow 管理设计

## Workflow 列表

项目中包含多个 Workflow，每个 Workflow 是一个完整的 AI 处理流水线。

### 列表视图

```
┌──────────────────────────────────────────────────────────────┐
│  Workflow (3)                              [新建] [导入]    │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │ YOLO Detection Pipeline                [Ready]   [▶]  │  │
│  │ 3 个节点 · v1.2 · 最后运行 2 小时前 · 运行成功         │  │
│  │ ───────────────────────────────────────────────────── │  │
│  │ [YOLO Input] → [Preprocess] → [YOLO Detect] → [Output]│  │
│  └────────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │ LSTM Prediction Flow                  [Running]  [⏹]  │  │
│  │ 4 个节点 · v2.0 · 运行中 · 进度 65%                    │  │
│  │ ───────────────────────────────────────────────────── │  │
│  │ [Data Load] → [Feature Eng] → [LSTM] → [Predict]      │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │ SUMO Simulation Chain                 [Ready]   [▶]  │  │
│  │ 2 个节点 · v1.0 · 最后运行 3 天前 · 运行失败          │  │
│  │ ───────────────────────────────────────────────────── │  │
│  │ [Predict] → [SUMO Sim]                                 │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### Workflow 卡片规格

| 属性 | 值 |
|------|-----|
| 背景 | `bg-tertiary` |
| 圆角 | 12px |
| 内边距 | 16px |
| 悬停 | 阴影 `shadow` |
| 运行中 | 左侧 `warning` 色边条（动画） |
| 失败 | 左侧 `error` 色边条 |

### Workflow 节点预览

在卡片底部以简化流程图形式展示节点链：

```
[节点1] ─→ [节点2] ─→ [节点3] ─→ [节点4]
```

每个节点：圆角 6px，背景 `bg-active`，字号 11px，高度 24px

---

# Model 管理设计

## 模型仓库

项目中的模型仓库，类似 HuggingFace Model Hub 的本地版。

### 模型列表

```
┌──────────────────────────────────────────────────────────────┐
│  Models (4)                                  [导入] [训练]  │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │ 🟣 yolo_traffic_v2.1                                   │  │
│  │ PyTorch · 24.5 MB · 训练于 2026-07-01                  │  │
│  │ mAP@50: 0.89 · mAP@50-95: 0.72                        │  │
│  │ [加载] [部署] [导出] [⋯]                               │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │ 🟣 lstm_flow_predictor                                 │  │
│  │ PyTorch · 8.2 MB · 训练于 2026-06-28                   │  │
│  │ MSE: 0.023 · MAE: 0.112 · R²: 0.94                    │  │
│  │ [加载] [部署] [导出] [⋯]                               │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │ 🟣 yolov8n_finetuned                                   │  │
│  │ ONNX · 12.8 MB · 训练于 2026-06-25                     │  │
│  │ mAP@50: 0.85 · mAP@50-95: 0.68                        │  │
│  │ [加载] [部署] [导出] [⋯]                               │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### 模型卡片信息结构

```
[模型图标] 模型名称
框架 · 大小 · 训练时间
指标1: 值 · 指标2: 值 · 指标3: 值
[操作按钮组]
```

### 支持框架

| 框架 | 图标 | 颜色 |
|------|------|------|
| PyTorch | 🔥 | `#ee4c2c` |
| ONNX | 📦 | `#005CED` |
| TensorRT | ⚡ | `#76B900` |

### 指标显示规则

- 检测模型：显示 mAP@50、mAP@50-95
- 分类模型：显示 Accuracy、Precision、Recall、F1
- 回归模型：显示 MSE、MAE、R²
- 自定义指标：根据 Experiment 配置动态显示

### 模型操作

| 操作 | 说明 | 目标位置 |
|------|------|----------|
| 加载 | 加载模型到 Workflow 节点 | Workflow 编辑器 |
| 部署 | 将模型部署为推理服务 | 部署面板 |
| 导出 | 导出模型文件到本地 | 文件系统 |
| 对比 | 与另一个模型对比指标 | 实验对比视图 |

---

# Experiment 实验管理设计

参考 MLFlow Tracking 的设计，每次训练生成一个 Experiment 记录。

## 实验列表

```
┌──────────────────────────────────────────────────────────────┐
│  Experiments (12)                        [对比] [筛选] [导出]│
├──────────────────────────────────────────────────────────────┤
│  筛选：[全部 ▼] [模型 ▼] [状态 ▼]     搜索: [________]     │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ID          模型             参数    Epoch  Loss   Acc  GPU │
│  ─────────────────────────────────────────────────────────── │
│  exp_0701    yolo_v2.1       lr=0.01  100   0.032  0.89  A100│
│  exp_0628    lstm_flow       lr=0.001 200   0.023  0.94  3090│
│  exp_0625    resnet50        lr=0.005  50   0.156  0.82  3090│
│  exp_0620    yolov8n_ft      lr=0.01   80   0.045  0.85  A100│
│  exp_0615    lstm_v2         lr=0.002  150   0.031  0.91  3090│
│  ...                                                          │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### 表格规格

| 属性 | 值 |
|------|-----|
| 表头高度 | 40px，`bg-secondary`，`border-subtle` 底部边框 |
| 行高 | 40px |
| 行边框 | `border-subtle` 底部 |
| 悬停行 | `bg-hover` |
| 选中行 | `bg-active` |
| 排序 | 点击表头排序，显示排序箭头 |

### 实验详情页

选中实验后展开详情面板：

```
┌──────────────────────────────────────────────────────────────┐
│  exp_20260701_yolo_train                    [已完成 ✓]      │
│  YOLO Detection Pipeline · 训练于 2026-07-01 14:30          │
│  耗时: 45 分钟 · GPU: NVIDIA A100 40GB                      │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  参数                          │  指标                       │
│  ────────────────────────────  │  ──────────────────────────│
│  learning_rate:    0.01        │  mAP@50:       0.89        │
│  batch_size:       16          │  mAP@50-95:    0.72        │
│  epochs:           100         │  precision:    0.91        │
│  image_size:       640         │  recall:       0.85        │
│  optimizer:        SGD         │  loss_final:   0.032       │
│  momentum:         0.937       │                              │
│                                                              │
├──────────────────────────────────────────────────────────────┤
│  训练曲线                                                    │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  Loss 曲线（折线图）                                     │  │
│  │  ▁▂▃▄▅▆▇██████████                                    │  │
│  └────────────────────────────────────────────────────────┘  │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  mAP 曲线（折线图）                                     │  │
│  │  ▇▆▅▄▃▂▁▁▁▁▁▁▁▁▁▁▁                                  │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                              │
├──────────────────────────────────────────────────────────────┤
│  操作：[加载模型] [复制参数] [导出报告] [对比实验]              │
└──────────────────────────────────────────────────────────────┘
```

## 实验对比视图

选择两个或多个实验进行对比：

```
┌──────────────────────────────────────────────────────────────┐
│  实验对比                               [关闭] [添加实验 +]  │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│              │  exp_0701_yolo   │  exp_0620_yolov8n         │
│  ────────────┼─────────────────┼───────────────────────────│
│  模型        │  yolo_traffic   │  yolov8n_finetuned        │
│  学习率      │  0.01           │  0.01                      │
│  Batch Size  │  16             │  16                        │
│  Epochs      │  100            │  80                        │
│  ────────────┼─────────────────┼───────────────────────────│
│  mAP@50      │  0.89  ★        │  0.85                      │
│  mAP@50-95   │  0.72  ★        │  0.68                      │
│  Precision   │  0.91  ★        │  0.87                      │
│  Recall      │  0.85           │  0.83                      │
│  Loss Final  │  0.032  ★       │  0.045                     │
│  训练时间    │  45 min         │  35 min                    │
│  ────────────┼─────────────────┼───────────────────────────│
│  结论        │  更优 (4/6)     │  训练更快                   │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### 对比规则

- 较优指标标记 `★`（绿色）
- 指标数值使用等宽字体 `font-family-mono`
- 表头固定，内容可滚动
- 支持添加更多实验列

---

# Environment 环境管理设计

体现 AIStudio 作为 AI 开发平台的特色，不是简单的依赖列表。

## 环境概览

```
┌──────────────────────────────────────────────────────────────┐
│  Environment                                    [Repair] [⟳]│
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  运行环境状态                                    [Ready ✓]   │
│                                                              │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐       │
│  │ Python   │ │ CUDA     │ │ PyTorch  │ │ GPU      │       │
│  │ 3.10.12  │ │ 11.8     │ │ 2.0.1    │ │ A100 40G │       │
│  │ ✓        │ │ ✓        │ │ ✓        │ │ 78°C     │       │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘       │
│                                                              │
├──────────────────────────────────────────────────────────────┤
│  GPU 状态                                                    │
│  ┌────────────────────────────────────────────────────────┐  │
│  │ NVIDIA A100 40GB                                       │  │
│  │ 利用率: ████████████░░░░░░░░ 65%                       │  │
│  │ 显存:   ██████████░░░░░░░░░░ 12.8 / 40.0 GB           │  │
│  │ 温度:   78°C                                           │  │
│  │ 功耗:   280W / 400W                                    │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                              │
├──────────────────────────────────────────────────────────────┤
│  依赖状态                                          [48 已安装]│
│                                                              │
│  已安装依赖：                                                 │
│  ┌─────────────────┬──────────┬──────────┬────────────────┐  │
│  │ 包名            │ 当前版本  │ 最新版本  │ 状态           │  │
│  ├─────────────────┼──────────┼──────────┼────────────────┤  │
│  │ ultralytics     │ 8.0.120  │ 8.0.120  │ ✓ 最新         │  │
│  │ torch           │ 2.0.1    │ 2.1.0    │ ↑ 可升级       │  │
│  │ torchvision     │ 0.15.2   │ 0.16.0   │ ↑ 可升级       │  │
│  │ numpy           │ 1.24.3   │ 1.24.3   │ ✓ 最新         │  │
│  │ opencv-python   │ 4.8.0    │ 4.8.1    │ ↑ 可升级       │  │
│  │ pandas          │ 2.0.3    │ 2.0.3    │ ✓ 最新         │  │
│  └─────────────────┴──────────┴──────────┴────────────────┘  │
│                                                              │
├──────────────────────────────────────────────────────────────┤
│  操作：[安装依赖] [升级全部] [导出 requirements.txt] [重装环境] │
└──────────────────────────────────────────────────────────────┘
```

## 环境状态指示

| 状态 | 颜色 | 图标 | 说明 |
|------|------|------|------|
| Ready | `success` 绿色 | ✓ | 所有依赖就绪，可以运行 |
| Warning | `warning` 橙色 | ⚠ | 部分依赖可升级，不影响运行 |
| Error | `error` 红色 | ✗ | 缺少必要依赖或版本不兼容 |

## GPU 监控

实时显示 GPU 状态：

- 利用率进度条：`bg-hover` 背景 + `primary` 填充
- 显存进度条：`bg-hover` 背景 + `primary` 填充
- 温度：数值显示，>85°C 时变为 `error` 颜色
- 功耗：数值显示

## 依赖操作

| 操作 | 说明 |
|------|------|
| 安装依赖 | 弹出搜索框，搜索并安装 Python 包 |
| 升级全部 | 一键升级所有可更新的依赖 |
| 导出 | 导出 requirements.txt 或 environment.yml |
| 重装环境 | 重新创建 Python 虚拟环境并安装所有依赖 |
| 修复环境 | 自动检测并修复环境问题（AI 建议） |

---

# AI Assistant Panel 设计

右侧 AI 助手面板，类似 Codex Agent，能够理解当前项目上下文并提供建议。

## 面板结构

```
┌──────────────────────────────────┐
│  AI Assistant            [Ctrl+A]│
├──────────────────────────────────┤
│                                  │
│  上下文信息                       │
│  ┌────────────────────────────┐  │
│  │ 项目: Traffic Project      │  │
│  │ Workflow: YOLO Pipeline    │  │
│  │ 当前文件: train.py         │  │
│  │ 环境: Ready                │  │
│  └────────────────────────────┘  │
│                                  │
├──────────────────────────────────┤
│                                  │
│  AI 建议                         │
│  ┌────────────────────────────┐  │
│  │ ⚠ 检测到环境问题           │  │
│  │                            │  │
│  │ CUDA 版本 (11.8) 与        │  │
│  │ PyTorch 2.0.1 存在         │  │
│  │ 兼容性风险                  │  │
│  │                            │  │
│  │ 建议：升级到 CUDA 12.1     │  │
│  │                            │  │
│  │ [Apply Fix]  [Ignore]      │  │
│  └────────────────────────────┘  │
│                                  │
│  ┌────────────────────────────┐  │
│  │ 💡 优化建议                 │  │
│  │                            │  │
│  │ 模型 yolo_traffic_v2.1     │  │
│  │ 的 mAP 可通过增加数据增强   │  │
│  │ 进一步提升                  │  │
│  │                            │  │
│  │ [查看建议]  [忽略]          │  │
│  └────────────────────────────┘  │
│                                  │
│  ┌────────────────────────────┐  │
│  │ 🔍 最近错误分析             │  │
│  │                            │  │
│  │ exp_0625 训练失败           │  │
│  │ 原因: GPU 显存不足          │  │
│  │ 建议: 减小 batch_size      │  │
│  │                            │  │
│  │ [自动修复]  [查看详情]       │  │
│  └────────────────────────────┘  │
│                                  │
├──────────────────────────────────┤
│  [Ask AI...]               [→]  │
└──────────────────────────────────┘
```

## 上下文感知

AI 助手自动收集以下上下文：

| 上下文 | 来源 | 用途 |
|--------|------|------|
| 当前项目 | Project Explorer 选中项 | 理解项目范围 |
| 当前 Workflow | Workspace 打开的 Workflow | 提供 Workflow 优化建议 |
| 当前文件 | 编辑器打开的文件 | 代码辅助 |
| 环境状态 | Environment 面板 | 检测环境问题 |
| 最近错误 | Logs 面板 | 错误分析和修复建议 |
| 训练状态 | Experiment 面板 | 训练进度和优化建议 |

## 建议卡片类型

| 类型 | 图标颜色 | 触发条件 |
|------|----------|----------|
| 环境问题 | `warning` 橙色 | 依赖版本不兼容、GPU 异常 |
| 优化建议 | `info` 蓝色 | 模型可优化、参数可调优 |
| 错误分析 | `error` 红色 | 训练失败、运行错误 |
| 快捷操作 | `primary` 紫色 | 常用操作快捷入口 |

## 交互

- 点击 `Apply Fix`：自动执行修复操作
- 点击 `查看详情`：展开建议详情或跳转到相关页面
- 点击 `Ignore`：关闭该建议卡片
- 底部输入框：自由提问，AI 基于项目上下文回答

---

# 项目创建流程设计

## New Project 弹窗

```
┌──────────────────────────────────────────────────────────────┐
│  新建项目                                              [✕]  │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  项目名称: [________________________]                        │
│  项目描述: [________________________]                        │
│  项目路径: [/Users/.../AIStudio/________] [浏览]             │
│                                                              │
├──────────────────────────────────────────────────────────────┤
│  选择模板                                                    │
│                                                              │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐       │
│  │          │ │  🚗      │ │  🖼️      │ │  📈      │       │
│  │  Empty   │ │  YOLO    │ │  Image   │ │  Time    │       │
│  │  Project │ │Detection │ │  Classif │ │  Series  │       │
│  │          │ │          │ │          │ │          │       │
│  │ 空白项目 │ │车辆检测   │ │图像分类   │ │时序预测   │       │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘       │
│                                                              │
│  ┌──────────┐                                               │
│  │  🚦      │                                               │
│  │  Smart   │                                               │
│  │ Traffic  │                                               │
│  │          │                                               │
│  │ 智慧交通 │                                               │
│  └──────────┘                                               │
│                                                              │
├──────────────────────────────────────────────────────────────┤
│  项目配置                                                    │
│                                                              │
│  Framework:  [PyTorch ▼]    Plugin: [YOLO Plugin ▼]         │
│  Dataset:    [无 ▼]          GPU:    [Auto ▼]               │
│                                                              │
├──────────────────────────────────────────────────────────────┤
│                                          [取消]  [创建项目]   │
└──────────────────────────────────────────────────────────────┘
```

### 模板卡片规格

| 属性 | 值 |
|------|-----|
| 宽度 | 140px |
| 高度 | 120px |
| 圆角 | 12px |
| 背景 | `bg-tertiary` |
| 选中 | `primary` 边框 + `primary-bg` 背景 |
| 悬停 | `shadow` 阴影 |

### 创建流程

```
用户点击"新建项目"
    ↓
弹出 New Project 弹窗
    ↓
填写项目名称、描述、路径
    ↓
选择项目模板（或空白项目）
    ↓
选择 Framework（PyTorch / TensorFlow / JAX）
    ↓
选择 Plugin（可选，后续可安装）
    ↓
选择 Dataset（可选）
    ↓
点击"创建项目"
    ↓
自动创建项目目录结构
    ↓
安装必要依赖（如选择了模板）
    ↓
生成默认 Workflow（模板预设）
    ↓
进入 Project Dashboard
```

---

# 状态系统设计

## 项目状态

| 状态 | 颜色 | 动画 | 说明 |
|------|------|------|------|
| Ready | `success` 绿色 | 无 | 项目就绪，可执行操作 |
| Running | `info` 蓝色 | 呼吸动画 | Workflow 正在执行 |
| Training | `warning` 橙色 | 呼吸动画 | 模型正在训练 |
| Failed | `error` 红色 | 无 | 执行失败 |
| Completed | `success` 绿色 | 完成闪烁一次 | 任务完成 |

### 状态标签样式

- 圆角：6px
- 内边距：2px 8px
- 字号：12px
- 字重：600
- 背景：对应颜色的 `*-bg` 变量
- 文字：对应颜色

## Workflow 状态

| 状态 | 颜色 | 说明 |
|------|------|------|
| Idle | `neutral` 灰色 | 未运行 |
| Running | `info` 蓝色 | 正在执行 |
| Success | `success` 绿色 | 执行成功 |
| Failed | `error` 红色 | 执行失败 |
| Paused | `warning` 橙色 | 已暂停 |

## Experiment 状态

| 状态 | 颜色 | 说明 |
|------|------|------|
| Running | `info` 蓝色 + 旋转图标 | 正在训练 |
| Completed | `success` 绿色 | 训练完成 |
| Failed | `error` 红色 | 训练失败 |
| Cancelled | `neutral` 灰色 | 用户取消 |

---

# 组件设计

## Vue 组件结构

```
src/pages/Project/
├── Project.vue                          # 页面根组件（三栏布局）
│
├── components/
│   ├── toolbar/
│   │   ├── ProjectToolbar.vue           # 顶部工具栏（面包屑 + 搜索 + 操作）
│   │   └── BreadcrumbNav.vue            # 面包屑导航
│   │
│   ├── explorer/
│   │   ├── ProjectExplorer.vue          # 左侧资源管理器容器
│   │   ├── ExplorerTree.vue             # 树形控件
│   │   ├── TreeNode.vue                 # 单个树节点
│   │   └── ContextMenu.vue              # 右键菜单
│   │
│   ├── dashboard/
│   │   ├── ProjectDashboard.vue         # 项目 Dashboard 主页
│   │   ├── StatCard.vue                 # 统计卡片
│   │   ├── QuickActions.vue             # 快速操作按钮组
│   │   ├── RecentActivity.vue           # 最近活动面板
│   │   └── ProjectHeader.vue            # 项目头部信息
│   │
│   ├── dataset/
│   │   ├── DatasetManager.vue           # 数据集管理页面
│   │   ├── DatasetCard.vue              # 数据集卡片
│   │   ├── DatasetDetail.vue            # 数据集详情
│   │   ├── DataPreview.vue              # 数据预览（图片/表格）
│   │   └── FormatConverter.vue          # 格式转换工具
│   │
│   ├── workflow/
│   │   ├── WorkflowManager.vue          # Workflow 列表页面
│   │   ├── WorkflowCard.vue             # Workflow 卡片
│   │   └── WorkflowPreview.vue          # Workflow 节点预览
│   │
│   ├── model/
│   │   ├── ModelManager.vue             # 模型仓库页面
│   │   ├── ModelCard.vue                # 模型卡片
│   │   └── ModelDetail.vue              # 模型详情面板
│   │
│   ├── experiment/
│   │   ├── ExperimentTable.vue          # 实验列表表格
│   │   ├── ExperimentDetail.vue         # 实验详情面板
│   │   ├── ExperimentCompare.vue        # 实验对比视图
│   │   ├── TrainingCurve.vue            # 训练曲线图表
│   │   └── MetricBadge.vue              # 指标徽章
│   │
│   ├── environment/
│   │   ├── EnvironmentPanel.vue         # 环境管理面板
│   │   ├── GPUMonitor.vue               # GPU 状态监控
│   │   ├── DependencyList.vue           # 依赖列表
│   │   └── StatusIndicator.vue          # 环境状态指示器
│   │
│   ├── assistant/
│   │   ├── AIAssistantPanel.vue         # AI 助手面板容器
│   │   ├── ContextInfo.vue              # 上下文信息卡片
│   │   ├── SuggestionCard.vue           # AI 建议卡片
│   │   └── AssistantInput.vue           # AI 对话输入框
│   │
│   ├── dialog/
│   │   ├── ProjectCreateDialog.vue      # 新建项目弹窗
│   │   ├── TemplateSelector.vue         # 模板选择器
│   │   └── ProjectSettings.vue          # 项目设置弹窗
│   │
│   └── shared/
│       ├── StatusBadge.vue              # 状态标签（通用）
│       ├── EmptyState.vue               # 空状态组件
│       ├── SearchBox.vue                # 搜索框
│       └── ViewToggle.vue               # 网格/列表视图切换
│
└── composables/
    ├── useProject.ts                    # 项目状态管理
    ├── useExplorer.ts                   # 资源管理器逻辑
    ├── useEnvironment.ts                # 环境监控逻辑
    └── useExperiment.ts                 # 实验管理逻辑
```

## 组件职责说明

| 组件 | 职责 | Props |
|------|------|-------|
| Project.vue | 三栏布局编排，全局状态管理 | - |
| ProjectToolbar.vue | 面包屑、搜索、全局操作 | `project: Project` |
| ProjectExplorer.vue | 左侧资源树容器，管理展开/选中状态 | `resources: ResourceTree` |
| TreeNode.vue | 单个树节点渲染 | `node: TreeNode` `level: number` |
| ProjectDashboard.vue | 项目主页，组合各子面板 | `project: Project` |
| StatCard.vue | 统计数据卡片 | `icon: string` `label: string` `value: number` |
| DatasetManager.vue | 数据集列表和管理 | `datasets: Dataset[]` |
| DatasetCard.vue | 单个数据集卡片 | `dataset: Dataset` |
| WorkflowManager.vue | Workflow 列表和管理 | `workflows: Workflow[]` |
| WorkflowCard.vue | 单个 Workflow 卡片 | `workflow: Workflow` |
| ModelManager.vue | 模型仓库列表 | `models: Model[]` |
| ModelCard.vue | 单个模型卡片 | `model: Model` |
| ExperimentTable.vue | 实验列表表格 | `experiments: Experiment[]` |
| ExperimentCompare.vue | 多实验对比视图 | `experiments: Experiment[]` |
| EnvironmentPanel.vue | 环境状态和依赖管理 | `environment: Environment` |
| GPUMonitor.vue | GPU 实时监控 | `gpus: GPU[]` |
| AIAssistantPanel.vue | AI 助手面板容器 | `context: AIContext` |
| SuggestionCard.vue | AI 建议卡片 | `suggestion: Suggestion` |
| ProjectCreateDialog.vue | 新建项目弹窗 | `visible: boolean` |

---

# 数据结构设计

## Project

```typescript
interface Project {
  id: string
  name: string
  description: string
  path: string
  template: ProjectTemplate
  status: ProjectStatus
  createdAt: number
  updatedAt: number
  workflows: Workflow[]
  datasets: Dataset[]
  models: Model[]
  experiments: Experiment[]
  environment: Environment
}

type ProjectStatus = 'ready' | 'running' | 'training' | 'failed' | 'completed'

type ProjectTemplate =
  | 'empty'
  | 'yolo-detection'
  | 'image-classification'
  | 'time-series'
  | 'smart-traffic'
```

## Workflow

```typescript
interface Workflow {
  id: string
  name: string
  projectId: string
  version: string
  nodes: WorkflowNode[]
  edges: WorkflowEdge[]
  status: WorkflowStatus
  lastRunAt?: number
  lastRunResult?: 'success' | 'failed'
  createdAt: number
  updatedAt: number
}

type WorkflowStatus = 'idle' | 'running' | 'success' | 'failed' | 'paused'

interface WorkflowNode {
  id: string
  type: string
  label: string
  config: Record<string, unknown>
  position: { x: number; y: number }
}

interface WorkflowEdge {
  id: string
  source: string
  target: string
}
```

## Dataset

```typescript
interface Dataset {
  id: string
  name: string
  projectId: string
  description: string
  format: 'yolo' | 'coco' | 'voc' | 'csv'
  size: number                    // bytes
  sampleCount: number
  classCount: number
  classNames: string[]
  path: string
  previewImages?: string[]
  createdAt: number
  updatedAt: number
}
```

## Model

```typescript
interface Model {
  id: string
  name: string
  projectId: string
  framework: 'pytorch' | 'onnx' | 'tensorrt'
  version: string
  size: number                    // bytes
  filePath: string
  metrics: Record<string, number>
  experimentId?: string
  trainedAt: number
  createdAt: number
}
```

## Experiment

```typescript
interface Experiment {
  id: string
  name: string
  projectId: string
  workflowId: string
  modelName: string
  status: 'running' | 'completed' | 'failed' | 'cancelled'
  parameters: Record<string, unknown>
  metrics: Record<string, number>
  gpu: string
  duration: number                // ms
  epoch?: number
  maxEpoch?: number
  lossHistory: number[]
  metricHistory: Record<string, number[]>
  startedAt: number
  finishedAt?: number
  createdAt: number
}
```

## Environment

```typescript
interface Environment {
  status: 'ready' | 'warning' | 'error'
  python: { version: string; path: string }
  cuda: { version: string; path: string }
  pytorch: { version: string }
  gpus: GPU[]
  dependencies: Dependency[]
}

interface GPU {
  id: string
  name: string
  memoryTotal: number             // GB
  memoryUsed: number
  utilization: number             // 0-100
  temperature: number
  power: number
  powerMax: number
}

interface Dependency {
  name: string
  currentVersion: string
  latestVersion: string
  status: 'latest' | 'upgradable' | 'missing'
}
```

## Resource Tree (Explorer)

```typescript
interface ResourceTreeNode {
  id: string
  type: 'project' | 'workflow' | 'dataset' | 'model' | 'experiment' | 'environment' | 'outputs' | 'logs'
  label: string
  icon: string
  color: string
  count?: number
  children?: ResourceTreeNode[]
  isExpanded?: boolean
  isSelected?: boolean
}
```

---

# 交互流程

## 完整项目生命周期

```
用户打开 AIStudio
    ↓
进入 Dashboard → 点击"新建项目"
    ↓
ProjectCreateDialog 弹出
    ↓
选择模板（如 YOLO Detection）
    ↓
填写项目名称、路径
    ↓
选择 Framework + Plugin + Dataset
    ↓
点击"创建项目"
    ↓
系统创建项目目录 + 安装依赖 + 生成默认 Workflow
    ↓
进入 Project 页面 → 显示 Project Dashboard
    ↓
用户在 Explorer 中选择 Workflow → 打开编辑器
    ↓
用户编辑 Workflow 节点 → 配置参数
    ↓
用户点击"运行" → Workflow 状态变为 Running
    ↓
执行完成 → 自动创建 Experiment 记录
    ↓
用户查看 Experiment → 分析训练曲线和指标
    ↓
用户满意 → 模型自动保存到 Models
    ↓
用户选择模型 → 点击"部署"
    ↓
模型部署为推理服务
```

## 快捷键映射

| 快捷键 | 操作 |
|--------|------|
| `Ctrl+N` | 新建项目 |
| `Ctrl+Shift+N` | 新建 Workflow |
| `Ctrl+R` | 运行当前 Workflow |
| `Ctrl+S` | 保存当前编辑 |
| `Ctrl+Shift+A` | 切换 AI Assistant |
| `Ctrl+P` | 快速打开资源（类似 VS Code Ctrl+P） |
| `Ctrl+Shift+E` | 聚焦 Explorer |
| `F2` | 重命名选中资源 |
| `Delete` | 删除选中资源 |
| `Ctrl+Z` | 撤销 |
| `Ctrl+Shift+Z` | 重做 |

---

# 响应式适配

| 窗口宽度 | 布局调整 |
|----------|----------|
| > 1400px | 三栏全部显示，Explorer 240px + Assistant 280px |
| 1100px ~ 1400px | Assistant 可折叠，Explorer 200px |
| 900px ~ 1100px | Explorer 折叠为图标栏（48px），Assistant 隐藏 |
| < 900px | 仅 Workspace，Explorer 和 Assistant 通过按钮呼出 |

折叠状态下：
- Explorer 显示为 48px 宽的图标栏，只显示类型图标
- 点击图标展开为完整面板
- Assistant 通过 Toolbar 按钮或快捷键呼出为覆盖层

---

# 错误与空状态设计

## 空状态

### 无项目

```
┌──────────────────────────────────────┐
│                                      │
│         [文件夹图标 64px]             │
│                                      │
│         还没有项目                    │
│         创建你的第一个 AI 项目        │
│                                      │
│         [新建项目]                    │
│                                      │
└──────────────────────────────────────┘
```

### 无 Workflow

```
┌──────────────────────────────────────┐
│                                      │
│         [工作流图标 48px]             │
│                                      │
│         此项目还没有 Workflow         │
│         创建一个 Workflow 开始构建    │
│                                      │
│         [新建 Workflow] [从模板创建]  │
│                                      │
└──────────────────────────────────────┘
```

### 搜索无结果

```
┌──────────────────────────────────────┐
│                                      │
│         [搜索图标 48px]               │
│                                      │
│         没有找到匹配的内容            │
│         尝试其他搜索关键词            │
│                                      │
└──────────────────────────────────────┘
```

## 错误状态

### 环境错误

- 红色错误边框
- 错误图标 + 错误描述
- 修复建议
- `Repair Environment` 按钮

### 训练失败

- 实验行显示红色背景
- 错误摘要
- `查看日志` 按钮
- `重试` 按钮

### 网络/文件错误

- Toast 消息提示（右上角）
- 自动重试 3 次
- 失败后显示 `重试` 按钮

---

# 动画规范

遵循 `Docs/UI/design-system.md` 动画规范，针对项目管理模块的具体补充：

| 操作 | 动画 | 时长 | 缓动 |
|------|------|------|------|
| 树节点展开/折叠 | 高度平滑过渡 | 200ms | `ease-in-out` |
| 选中项切换 | 背景色过渡 | 150ms | `ease-out` |
| 卡片悬停 | 上移 2px + 阴影增强 | 150ms | `ease-out` |
| 页面切换（Dashboard → Dataset） | 淡入淡出 | 200ms | `ease-in-out` |
| 弹窗打开 | 中心放大 + 淡入 | 250ms | `cubic-bezier(0.4, 0, 0.2, 1)` |
| 弹窗关闭 | 缩小 + 淡出 | 200ms | `ease-in` |
| 状态标签变化 | 颜色过渡 | 300ms | `ease` |
| 进度条更新 | 宽度平滑过渡 | 500ms | `ease` |
| GPU 利用率变化 | 宽度平滑过渡 | 1000ms | `linear` |
| 训练曲线绘制 | 从左到右绘制 | 800ms | `ease-out` |
| 建议卡片出现 | 从右侧滑入 + 淡入 | 250ms | `ease-out` |
| 建议卡片消失 | 向右滑出 + 淡出 | 200ms | `ease-in` |

---

# 后续扩展

1. **项目版本控制**：集成 Git，支持版本提交、分支管理、回滚
2. **团队协作**：支持项目共享、权限管理、评论批注
3. **项目模板市场**：社区共享项目模板
4. **自动化测试**：模型评估自动化、回归测试
5. **CI/CD 集成**：自动训练、自动部署流水线
6. **多项目对比**：跨项目的模型和实验对比
7. **项目导入/导出**：支持从 MLFlow、W&B 导入实验数据
8. **离线模式**：无网络环境下正常使用核心功能
