# Plugin Schema

## 1. 概述

Plugin Schema 定义了 AIStudio 插件系统的标准化数据协议。每个插件通过 `plugin.json` 描述自身能力，Backend 据此加载和注册插件，Frontend 据此动态渲染配置界面，Agent 据此理解插件能力并生成工作流。

**设计原则**：

- 声明式描述插件能力（端口、配置、依赖）
- 支持多语言插件（Python / Go / TypeScript / 进程隔离）
- 向后兼容，支持版本演进
- 支持 MCP 协议扩展外部能力

---

## 2. plugin.json 完整结构

```json
{
  "$schema": "https://aistudio.dev/schemas/plugin/v1.0.0.json",
  "name": "yolo-detector",
  "display_name": "YOLO 目标检测",
  "version": "1.2.0",
  "description": "基于 YOLOv8 的实时目标检测插件",
  "author": "AI Studio",
  "license": "MIT",
  "homepage": "https://github.com/aistudio/plugins/yolo-detector",
  "repository": "https://github.com/aistudio/plugins",
  "icon": "icons/yolo.png",
  "banner": "banners/yolo-detector.png",
  "tags": ["vision", "detection", "yolo", "realtime"],
  "type": "vision",
  "language": "python",
  "runtime": {
    "python": ">=3.10",
    "cuda": true,
    "gpu_memory_mb": 512,
    "timeout_ms": 60000
  },
  "entry": "main.py",
  "entry_function": "execute",
  "dependencies": {
    "python": {
      "ultralytics": ">=8.0.0",
      "opencv-python": ">=4.8.0",
      "torch": ">=2.0.0"
    },
    "system": {
      "ffmpeg": ">=4.0"
    }
  },
  "install": {
    "method": "pip",
    "command": "pip install -r requirements.txt",
    "requirements_file": "requirements.txt"
  },
  "ports": {
    "inputs": [
      {
        "name": "image",
        "type": "image",
        "required": true,
        "multiple": false,
        "description": "输入图像文件路径"
      },
      {
        "name": "confidence",
        "type": "number",
        "required": false,
        "default": 0.5,
        "description": "置信度阈值",
        "constraints": {
          "min": 0.0,
          "max": 1.0,
          "step": 0.05
        }
      }
    ],
    "outputs": [
      {
        "name": "detections",
        "type": "json",
        "description": "检测结果，包含 boxes、scores、classes"
      },
      {
        "name": "annotated_image",
        "type": "image",
        "description": "标注后的图像"
      }
    ]
  },
  "config_schema": {
    "model": {
      "type": "string",
      "label": "模型文件",
      "default": "yolov8n.pt",
      "description": "YOLO 模型权重文件",
      "options": [
        {"value": "yolov8n.pt", "label": "YOLOv8 Nano (3.2M)"},
        {"value": "yolov8s.pt", "label": "YOLOv8 Small (11.2M)"},
        {"value": "yolov8m.pt", "label": "YOLOv8 Medium (25.9M)"},
        {"value": "yolov8l.pt", "label": "YOLOv8 Large (43.7M)"},
        {"value": "yolov8x.pt", "label": "YOLOv8 X-Large (68.2M)"}
      ]
    },
    "device": {
      "type": "select",
      "label": "推理设备",
      "default": "auto",
      "options": [
        {"value": "auto", "label": "自动检测"},
        {"value": "cpu", "label": "CPU"},
        {"value": "cuda", "label": "GPU (CUDA)"}
      ]
    },
    "img_size": {
      "type": "number",
      "label": "输入尺寸",
      "default": 640,
      "description": "推理输入图像尺寸",
      "constraints": {
        "min": 320,
        "max": 1920,
        "step": 32
      }
    },
    "max_detections": {
      "type": "number",
      "label": "最大检测数",
      "default": 100,
      "constraints": {
        "min": 1,
        "max": 1000
      }
    }
  },
  "node_registration": {
    "type": "vision",
    "label": "YOLO 目标检测",
    "category": "Vision",
    "color": "#4CAF50",
    "icon": "icons/yolo-node.png",
    "description": "使用 YOLO 模型检测图像中的目标",
    "examples": [
      {
        "name": "车辆检测",
        "description": "检测图像中的车辆",
        "parameters": {"model": "yolov8n.pt", "confidence": 0.5}
      }
    ]
  },
  "compatibility": {
    "min_aistudio_version": "1.0.0",
    "max_aistudio_version": "",
    "platforms": ["windows", "linux", "macos"]
  },
  "changelog": {
    "1.2.0": "新增 annotated_image 输出端口",
    "1.1.0": "支持 RT-DETR 模型",
    "1.0.0": "初始版本"
  }
}
```

---

## 3. 字段详细说明

### 3.1 基本信息

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `$schema` | string | 否 | JSON Schema 地址 |
| `name` | string | 是 | 插件唯一标识（kebab-case） |
| `display_name` | string | 否 | 显示名称 |
| `version` | string | 是 | 语义版本号（SemVer） |
| `description` | string | 是 | 插件功能描述 |
| `author` | string | 是 | 作者名称 |
| `license` | string | 否 | 开源协议 |
| `homepage` | string | 否 | 主页 URL |
| `repository` | string | 否 | 代码仓库 URL |
| `icon` | string | 否 | 插件图标路径 |
| `banner` | string | 否 | 插件横幅图路径 |
| `tags` | string[] | 否 | 标签数组 |

### 3.2 分类与语言

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `type` | PluginType | 是 | 插件类型分类 |
| `language` | string | 是 | 插件开发语言 |
| `entry` | string | 是 | 入口文件 |
| `entry_function` | string | 否 | 入口函数名（默认 `execute`） |

```typescript
type PluginType =
  | "vision"        // 视觉处理
  | "nlp"           // 自然语言处理
  | "timeseries"    // 时序分析
  | "logic"         // 逻辑控制
  | "system"        // 系统操作
  | "simulation"    // 仿真
  | "mcp"           // MCP 协议
  | "data"          // 数据处理
  | "agent"         // Agent 能力
  | "utility";      // 工具类

type PluginLanguage =
  | "python"
  | "go"
  | "typescript"
  | "binary"        // 预编译二进制
  | "process";      // 进程隔离（stdin/stdout 通信）
```

### 3.3 运行时要求

```json
"runtime": {
  "python": ">=3.10",
  "cuda": true,
  "gpu_memory_mb": 512,
  "timeout_ms": 60000,
  "memory_mb": 2048,
  "platforms": ["windows", "linux"]
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `python` | string | Python 版本要求 |
| `cuda` | boolean | 是否需要 CUDA |
| `gpu_memory_mb` | int | 最低 GPU 显存要求 |
| `timeout_ms` | int | 默认超时时间 |
| `memory_mb` | int | 最低内存要求 |
| `platforms` | string[] | 支持的平台 |

### 3.4 依赖管理

```json
"dependencies": {
  "python": {
    "ultralytics": ">=8.0.0",
    "torch": ">=2.0.0"
  },
  "system": {
    "ffmpeg": ">=4.0"
  },
  "plugins": {
    "base-image-utils": ">=1.0.0"
  }
}
```

| 依赖类型 | 说明 |
|---------|------|
| `python` | Python pip 包 |
| `system` | 系统级依赖（需用户手动安装） |
| `plugins` | 其他 AIStudio 插件依赖 |

### 3.5 安装方式

```json
"install": {
  "method": "pip",
  "command": "pip install -r requirements.txt",
  "requirements_file": "requirements.txt"
}
```

```typescript
type InstallMethod =
  | "pip"           // pip install
  | "conda"         // conda install
  | "npm"           // npm install
  | "go"            // go install
  | "command"       // 自定义命令
  | "binary"        // 直接下载二进制
  | "git";          // git clone
```

---

## 4. Port 定义（端口）

### 4.1 端口结构

```json
{
  "name": "image",
  "type": "image",
  "required": true,
  "multiple": false,
  "default": null,
  "description": "输入图像文件路径",
  "accepts": ["image"],
  "constraints": {
    "max_size_mb": 50,
    "formats": ["jpg", "png", "bmp", "webp"]
  }
}
```

### 4.2 端口类型兼容矩阵

下游端口可接收上游端口的类型：

| 上游 \ 下游 | image | text | number | boolean | json | file | tensor | stream | any |
|------------|-------|------|--------|---------|------|------|--------|--------|-----|
| image | ✓ | - | - | - | - | ✓ | - | - | ✓ |
| text | - | ✓ | - | - | ✓ | ✓ | - | - | ✓ |
| number | - | - | ✓ | - | ✓ | - | - | - | ✓ |
| boolean | - | - | - | ✓ | ✓ | - | - | - | ✓ |
| json | - | - | - | - | ✓ | - | - | - | ✓ |
| file | ✓ | ✓ | - | - | ✓ | ✓ | - | - | ✓ |
| tensor | - | - | - | - | ✓ | - | ✓ | - | ✓ |
| stream | - | - | - | - | ✓ | - | - | ✓ | ✓ |

---

## 5. Config Schema（配置模式）

Config Schema 定义插件的配置项，Frontend 据此动态生成配置表单。

### 5.1 字段类型

```typescript
type FieldType =
  | "string"     // 文本输入
  | "number"     // 数字输入
  | "boolean"    // 开关
  | "select"     // 下拉选择
  | "multiselect"// 多选
  | "color"      // 颜色选择
  | "file"       // 文件选择
  | "directory"  // 目录选择
  | "slider"     // 滑块
  | "textarea"   // 多行文本
  | "json"       // JSON 编辑器
  | "password"   // 密码输入
  | "group";     // 字段分组
```

### 5.2 完整字段结构

```json
{
  "name": "confidence",
  "type": "number",
  "label": "置信度阈值",
  "description": "检测目标的最低置信度",
  "default": 0.5,
  "required": true,
  "group": "Detection",
  "constraints": {
    "min": 0.0,
    "max": 1.0,
    "step": 0.05
  },
  "validation": {
    "pattern": "^[0-9]+(\\.[0-9]+)?$",
    "message": "请输入有效的数值"
  },
  "depends_on": {
    "field": "model",
    "condition": "not_equals",
    "value": "custom"
  },
  "hidden": false,
  "disabled": false
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | string | 字段标识 |
| `type` | FieldType | 字段类型 |
| `label` | string | 显示标签 |
| `description` | string | 字段说明 |
| `default` | any | 默认值 |
| `required` | boolean | 是否必填 |
| `group` | string | 分组名称 |
| `constraints` | object | 值约束 |
| `validation` | object | 校验规则 |
| `depends_on` | object | 依赖字段（条件显示） |
| `hidden` | boolean | 是否隐藏 |
| `disabled` | boolean | 是否禁用 |

---

## 6. Node Registration（节点注册）

插件通过 `node_registration` 声明如何在工作流编辑器中呈现。

```json
"node_registration": {
  "type": "vision",
  "label": "YOLO 目标检测",
  "category": "Vision",
  "color": "#4CAF50",
  "icon": "icons/yolo-node.png",
  "description": "使用 YOLO 模型检测图像中的目标",
  "order": 10,
  "examples": [
    {
      "name": "车辆检测",
      "description": "检测图像中的车辆",
      "parameters": {"model": "yolov8n.pt"},
      "workflow_template": "templates/vehicle-detection.json"
    }
  ]
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `type` | string | 节点类型（对应 Workflow NodeType） |
| `label` | string | 节点面板显示名称 |
| `category` | string | 分类目录 |
| `color` | string | 节点颜色（HEX） |
| `icon` | string | 节点图标路径 |
| `description` | string | 节点描述 |
| `order` | int | 排序权重 |
| `examples` | object[] | 预设示例 |

---

## 7. 多语言插件规范

### 7.1 Python 插件

**目录结构**：

```
plugins/yolo-detector/
├── plugin.json              # 插件清单
├── main.py                  # 入口文件
├── requirements.txt         # Python 依赖
├── README.md                # 文档
├── icons/                   # 图标资源
├── banners/                 # 横幅资源
└── tests/                   # 测试用例
```

**入口文件约定**：

```python
# main.py
def execute(input_data: dict, config: dict, context: dict) -> dict:
    """
    插件执行入口。

    Args:
        input_data: 输入端口数据 {"image": "/path/to/img.jpg", "confidence": 0.5}
        config: 插件配置 {"model": "yolov8n.pt", "device": "cuda"}
        context: 执行上下文 {"workflow_id": "...", "task_id": "...", "work_dir": "..."}

    Returns:
        输出端口数据 {"detections": {...}, "annotated_image": "/path/to/annotated.jpg"}
    """
    pass
```

### 7.2 Go 插件

**接口要求**：

```go
type Plugin interface {
    Meta() PluginMeta
    Ports() PluginPorts
    Execute(ctx context.Context, input *PluginInput) (*PluginOutput, error)
    ConfigSchema() *ConfigSchema
}
```

### 7.3 进程隔离插件

通过 stdin/stdout JSON 通信，适用于不信任的第三方插件：

```
Parent Process                Plugin Process
    │                              │
    │── JSON Request ─────────────→│
    │                              │── 执行逻辑
    │←── JSON Response ────────────│
    │                              │
```

### 7.4 MCP 插件

MCP 插件通过 JSON-RPC 与外部软件通信：

```json
{
  "mcp_config": {
    "server": "matlab-simulink",
    "transport": "stdio",
    "command": "matlab-mcp-server",
    "args": ["--license", "standard"],
    "env": {
      "MATLAB_ROOT": "/usr/local/MATLAB/R2026a"
    },
    "tools": [
      {
        "name": "run_simulation",
        "description": "运行 Simulink 仿真",
        "input_schema": {
          "type": "object",
          "properties": {
            "model_name": {"type": "string"},
            "parameters": {"type": "object"}
          },
          "required": ["model_name"]
        }
      }
    ]
  }
}
```

---

## 8. 插件生命周期

```
发现 → 校验 → 注册 → 加载 → 初始化 → 就绪 → 执行 → 卸载
 │       │       │       │       │       │       │       │
 │       │       │       │       │       │       │       └── 释放资源
 │       │       │       │       │       │       └── 调用 Execute()
 │       │       │       │       │       └── 端口/配置注册完成
 │       │       │       │       └── 加载依赖、初始化模型
 │       │       │       └── 加载入口文件
 │       │       └── 写入 Plugin Registry
 │       └── 校验 plugin.json 合法性
 └── 扫描插件目录
```

### 生命周期状态

```typescript
type PluginLifecycleStatus =
  | "discovered"    // 已发现
  | "validating"    // 校验中
  | "registering"   // 注册中
  | "loading"       // 加载中
  | "initializing"  // 初始化中
  | "ready"         // 就绪
  | "running"       // 执行中
  | "error"         // 错误
  | "unloaded";     // 已卸载
```

---

## 9. 插件示例

### 9.1 YOLO Plugin

```json
{
  "name": "yolo-detector",
  "version": "1.2.0",
  "type": "vision",
  "language": "python",
  "author": "AI Studio",
  "description": "YOLO 目标检测",
  "entry": "main.py",
  "runtime": {"python": ">=3.10", "cuda": true},
  "dependencies": {
    "python": {"ultralytics": ">=8.0.0", "opencv-python": ">=4.8.0"}
  },
  "ports": {
    "inputs": [
      {"name": "image", "type": "image", "required": true},
      {"name": "confidence", "type": "number", "required": false, "default": 0.5}
    ],
    "outputs": [
      {"name": "detections", "type": "json"},
      {"name": "annotated_image", "type": "image"}
    ]
  },
  "config_schema": {
    "model": {"type": "select", "label": "模型", "default": "yolov8n.pt",
      "options": [{"value": "yolov8n.pt", "label": "Nano"}, {"value": "yolov8s.pt", "label": "Small"}]},
    "device": {"type": "select", "label": "设备", "default": "auto",
      "options": [{"value": "auto", "label": "自动"}, {"value": "cuda", "label": "GPU"}]}
  },
  "node_registration": {
    "type": "vision",
    "label": "YOLO 检测",
    "category": "Vision",
    "color": "#4CAF50"
  }
}
```

### 9.2 PyTorch Plugin

```json
{
  "name": "pytorch-inference",
  "version": "1.0.0",
  "type": "vision",
  "language": "python",
  "author": "AI Studio",
  "description": "通用 PyTorch 模型推理",
  "entry": "main.py",
  "runtime": {"python": ">=3.10", "cuda": true, "gpu_memory_mb": 1024},
  "dependencies": {
    "python": {"torch": ">=2.0.0", "torchvision": ">=0.15.0"}
  },
  "ports": {
    "inputs": [
      {"name": "input_tensor", "type": "tensor", "required": true},
      {"name": "model_path", "type": "file", "required": true}
    ],
    "outputs": [
      {"name": "output_tensor", "type": "tensor"},
      {"name": "predictions", "type": "json"}
    ]
  },
  "config_schema": {
    "model_type": {"type": "select", "label": "模型类型", "default": "classification",
      "options": [{"value": "classification", "label": "分类"}, {"value": "detection", "label": "检测"}, {"value": "segmentation", "label": "分割"}]},
    "device": {"type": "select", "label": "设备", "default": "auto"}
  },
  "node_registration": {
    "type": "vision",
    "label": "PyTorch 推理",
    "category": "Vision",
    "color": "#FF6F00"
  }
}
```

### 9.3 MCP Plugin

```json
{
  "name": "mcp-matlab",
  "version": "1.0.0",
  "type": "mcp",
  "language": "process",
  "author": "AI Studio",
  "description": "通过 MCP 协议连接 MATLAB Simulink",
  "entry": "mcp-bridge",
  "mcp_config": {
    "server": "matlab-simulink",
    "transport": "stdio",
    "command": "npx",
    "args": ["-y", "@anthropic/matlab-mcp-server"],
    "env": {"MATLAB_ROOT": "/usr/local/MATLAB/R2026a"},
    "tools": [
      {
        "name": "run_simulation",
        "description": "运行 Simulink 仿真",
        "input_schema": {
          "type": "object",
          "properties": {
            "model_name": {"type": "string", "description": "Simulink 模型名称"},
            "parameters": {"type": "object", "description": "仿真参数"}
          },
          "required": ["model_name"]
        },
        "output_schema": {
          "type": "object",
          "properties": {
            "status": {"type": "string"},
            "results": {"type": "object"}
          }
        }
      }
    ]
  },
  "ports": {
    "inputs": [
      {"name": "params", "type": "json", "required": true, "description": "仿真参数"}
    ],
    "outputs": [
      {"name": "result", "type": "json", "description": "仿真结果"}
    ]
  },
  "node_registration": {
    "type": "mcp",
    "label": "MATLAB 仿真",
    "category": "Simulation",
    "color": "#FF4081"
  }
}
```

---

## 10. 插件版本管理

| 版本变化 | 兼容性 | 说明 |
|---------|--------|------|
| 新增端口 | 向后兼容 | 旧工作流不受影响 |
| 删除端口 | 不兼容 | 需要 major 版本升级 |
| 修改端口类型 | 不兼容 | 需要 major 版本升级 |
| 新增配置项 | 向后兼容 | 使用默认值 |
| 删除配置项 | 不兼容 | 需要 major 版本升级 |
| 修改默认值 | 可能影响 | 需要 minor 版本升级 |
| 新增依赖 | 向后兼容 | 需要重新安装 |
| 删除依赖 | 向后兼容 | 需要重新安装 |
