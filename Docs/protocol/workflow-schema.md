# Workflow Schema

## 1. 概述

Workflow Schema 定义了 AIStudio 可视化工作流的完整数据结构。该结构被 Frontend 渲染、Backend 解析、Python Engine 执行、Agent 生成，是整个系统的核心数据协议。

**设计原则**：

- 前端可序列化保存为 JSON
- Backend 可校验并转换为执行 DAG
- Python Engine 可读取并执行节点
- Agent 可基于 Schema 自动生成完整工作流
- 支持版本演进，向后兼容

---

## 2. 数据类型系统

节点间传递的数据遵循统一类型系统：

| 类型标识 | 说明 | 序列化方式 | 典型来源 |
|---------|------|-----------|---------|
| `image` | 图像数据 | 文件路径 / base64 | YOLO, SAM, OCR |
| `text` | 文本字符串 | UTF-8 string | Transformer, LLM |
| `number` | 数值 | float64 | 逻辑判断, 阈值 |
| `boolean` | 布尔值 | true / false | 条件分支 |
| `json` | 结构化数据 | JSON object | 任意复合结果 |
| `file` | 文件引用 | 路径 string | 数据集, 模型文件 |
| `tensor` | 张量数据 | 序列化 bytes | 深度学习推理 |
| `stream` | 流式数据 | chunk 序列 | LLM token 流 |
| `any` | 动态类型 | 运行时推断 | Agent 生成节点 |

---

## 3. Workflow 顶层结构

```json
{
  "schema_version": "1.0.0",
  "id": "wf_a1b2c3d4e5f6",
  "name": "车辆检测工作流",
  "description": "基于 YOLO 的车辆检测与分类流程",
  "project_id": "proj_001",
  "version": 3,
  "created_at": "2026-07-01T10:00:00Z",
  "updated_at": "2026-07-07T14:00:00Z",
  "author": "user_001",
  "tags": ["vision", "yolo", "detection"],
  "metadata": {
    "template": "blank",
    "generator": "manual",
    "estimated_nodes": 5
  },
  "variables": {
    "confidence_threshold": 0.5,
    "device": "cuda"
  },
  "nodes": [],
  "edges": []
}
```

### 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `schema_version` | string | 是 | Schema 版本号，用于兼容性校验 |
| `id` | string | 是 | 工作流唯一标识，格式 `wf_{nanoid}` |
| `name` | string | 是 | 工作流名称 |
| `description` | string | 否 | 工作流描述 |
| `project_id` | string | 是 | 所属项目 ID |
| `version` | int | 是 | 版本号，每次保存递增 |
| `created_at` | datetime | 是 | 创建时间（ISO 8601） |
| `updated_at` | datetime | 是 | 最后更新时间 |
| `author` | string | 否 | 创建者 ID |
| `tags` | string[] | 否 | 标签数组 |
| `metadata` | object | 否 | 扩展元数据 |
| `variables` | object | 否 | 工作流级变量，节点可引用 |
| `nodes` | Node[] | 是 | 节点数组 |
| `edges` | Edge[] | 是 | 边（连线）数组 |

---

## 4. Node 结构

每个节点代表工作流中的一个处理单元，对应一个 Plugin 或内置功能。

```json
{
  "id": "n1",
  "type": "vision",
  "plugin": "yolo-detector",
  "name": "YOLO 目标检测",
  "description": "使用 YOLOv8 检测图像中的车辆",
  "position": { "x": 100, "y": 200 },
  "size": { "width": 200, "height": 120 },
  "parameters": {
    "model": "yolov8n.pt",
    "confidence": 0.5,
    "device": "cuda"
  },
  "inputs": [
    {
      "id": "in_image",
      "name": "image",
      "type": "image",
      "required": true,
      "description": "输入图像"
    }
  ],
  "outputs": [
    {
      "id": "out_detections",
      "name": "detections",
      "type": "json",
      "description": "检测结果 {boxes, scores, classes}"
    }
  ],
  "runtime": {
    "status": "idle",
    "last_run": null,
    "error": null
  },
  "constraints": {
    "gpu_required": false,
    "max_retries": 3,
    "timeout_ms": 30000
  }
}
```

### Node 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | string | 是 | 节点唯一标识，格式 `n{序号}` |
| `type` | NodeType | 是 | 节点类型（见下方枚举） |
| `plugin` | string | 否 | 绑定的插件名称，内置节点可为空 |
| `name` | string | 是 | 节点显示名称 |
| `description` | string | 否 | 节点功能描述 |
| `position` | Point | 是 | 画布位置坐标 |
| `size` | Size | 否 | 节点显示尺寸 |
| `parameters` | object | 否 | 节点配置参数（由 Plugin schema 定义） |
| `inputs` | Port[] | 是 | 输入端口定义 |
| `outputs` | Port[] | 是 | 输出端口定义 |
| `runtime` | NodeRuntime | 否 | 运行时状态（执行时填充） |
| `constraints` | NodeConstraints | 否 | 执行约束条件 |

### NodeType 枚举

```typescript
type NodeType =
  | "vision"        // 视觉处理
  | "nlp"           // 自然语言处理
  | "timeseries"    // 时序分析
  | "logic"         // 逻辑控制
  | "system"        // 系统操作
  | "simulation"    // 仿真
  | "mcp"           // MCP 协议
  | "input"         // 数据输入
  | "output"        // 数据输出
  | "agent"         // Agent 节点
  | "subworkflow";  // 子工作流
```

### 内置节点类型

| type | plugin | 说明 | 输入 | 输出 |
|------|--------|------|------|------|
| `vision` | yolo-detector | YOLO 目标检测 | image | json |
| `vision` | sam-segmenter | SAM 图像分割 | image | json |
| `vision` | ocr-reader | OCR 文字识别 | image | text |
| `vision` | rtdetr-detector | RT-DETR 检测 | image | json |
| `nlp` | transformer | 文本分类/NER | text | json |
| `nlp` | llm-chat | 大语言模型 | text | stream |
| `timeseries` | lstm-predict | LSTM 时序预测 | tensor | json |
| `logic` | if-else | 条件判断 | any | boolean |
| `logic` | switch | 多路分支 | any | any |
| `logic` | loop | 循环控制 | any | any |
| `logic` | merge | 数据合并 | any[] | json |
| `system` | python-exec | Python 脚本 | any | any |
| `system` | terminal | 终端命令 | text | text |
| `system` | file-io | 文件读写 | file | any |
| `simulation` | sumo | SUMO 交通仿真 | json | json |
| `mcp` | mcp-client | MCP 协议客户端 | json | json |
| `input` | data-source | 数据源输入 | - | any |
| `output` | result-sink | 结果输出 | any | - |
| `agent` | auto-generate | Agent 自动生成 | text | json |

---

## 5. Port 结构（端口）

```json
{
  "id": "in_image",
  "name": "image",
  "type": "image",
  "required": true,
  "multiple": false,
  "default": null,
  "description": "输入图像文件路径",
  "accepts": ["image"],
  "constraints": {
    "max_size_mb": 50,
    "formats": ["jpg", "png", "bmp"]
  }
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | string | 是 | 端口唯一标识 |
| `name` | string | 是 | 端口名称 |
| `type` | string | 是 | 数据类型 |
| `required` | boolean | 否 | 是否必填，默认 true |
| `multiple` | boolean | 否 | 是否接受多条连线 |
| `default` | any | 否 | 默认值 |
| `description` | string | 否 | 端口描述 |
| `accepts` | string[] | 否 | 允许接收的数据类型（类型兼容约束） |
| `constraints` | object | 否 | 数据约束 |

---

## 6. Edge 结构（连线）

```json
{
  "id": "e1",
  "source": {
    "node_id": "n1",
    "port_id": "out_detections"
  },
  "target": {
    "node_id": "n2",
    "port_id": "in_input"
  },
  "label": "",
  "animated": false,
  "condition": null
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | string | 是 | 边唯一标识，格式 `e{序号}` |
| `source.node_id` | string | 是 | 源节点 ID |
| `source.port_id` | string | 是 | 源端口 ID |
| `target.node_id` | string | 是 | 目标节点 ID |
| `target.port_id` | string | 是 | 目标端口 ID |
| `label` | string | 否 | 连线标签 |
| `animated` | boolean | 否 | 是否显示动画（前端展示用） |
| `condition` | string | 否 | 条件表达式（条件连线） |

### 连线规则

1. 源端口的 `type` 必须与目标端口的 `type` 兼容
2. 一个输入端口只能接收一条连线（除非 `multiple: true`）
3. 不允许自环（node_id 不能相同）
4. 不允许形成环路（Backend 校验 DAG）

---

## 7. NodeRuntime 结构（运行时状态）

执行过程中由 Backend 填充，前端实时渲染。

```json
{
  "node_id": "n1",
  "status": "running",
  "progress": 0.45,
  "started_at": "2026-07-07T14:00:00Z",
  "finished_at": null,
  "duration_ms": null,
  "input_snapshot": {
    "image": "/storage/datasets/test.jpg"
  },
  "output_snapshot": null,
  "error": null,
  "metrics": {
    "cpu_percent": 35.2,
    "memory_mb": 512,
    "gpu_memory_mb": 1024,
    "gpu_utilization": 78.5
  },
  "logs": [
    {
      "timestamp": "2026-07-07T14:00:00Z",
      "level": "info",
      "message": "Loading model yolov8n.pt..."
    }
  ]
}
```

### NodeStatus 枚举

```typescript
type NodeStatus =
  | "idle"       // 未执行
  | "pending"    // 等待输入
  | "running"    // 执行中
  | "success"    // 执行成功
  | "error"      // 执行失败
  | "cancelled"  // 已取消
  | "skipped";   // 已跳过（条件分支）
```

---

## 8. 完整 JSON 示例

### 8.1 YOLO 车辆检测工作流

```json
{
  "schema_version": "1.0.0",
  "id": "wf_yolo_vehicle_detection",
  "name": "车辆检测工作流",
  "description": "输入图像 → YOLO 检测 → 条件判断 → 结果输出",
  "project_id": "proj_001",
  "version": 1,
  "created_at": "2026-07-07T10:00:00Z",
  "updated_at": "2026-07-07T10:00:00Z",
  "tags": ["vision", "yolo", "detection"],
  "variables": {
    "confidence_threshold": 0.5
  },
  "nodes": [
    {
      "id": "n1",
      "type": "input",
      "plugin": "data-source",
      "name": "图像输入",
      "position": { "x": 50, "y": 200 },
      "parameters": {},
      "inputs": [],
      "outputs": [
        {
          "id": "out_image",
          "name": "image",
          "type": "image",
          "required": true
        }
      ]
    },
    {
      "id": "n2",
      "type": "vision",
      "plugin": "yolo-detector",
      "name": "YOLO 车辆检测",
      "position": { "x": 350, "y": 200 },
      "parameters": {
        "model": "yolov8n.pt",
        "confidence": "$variables.confidence_threshold",
        "device": "cuda"
      },
      "inputs": [
        {
          "id": "in_image",
          "name": "image",
          "type": "image",
          "required": true
        }
      ],
      "outputs": [
        {
          "id": "out_detections",
          "name": "detections",
          "type": "json",
          "description": "检测结果 {boxes, scores, classes}"
        }
      ],
      "constraints": {
        "gpu_required": false,
        "timeout_ms": 30000
      }
    },
    {
      "id": "n3",
      "type": "logic",
      "plugin": "if-else",
      "name": "是否有车辆",
      "position": { "x": 650, "y": 200 },
      "parameters": {
        "condition": "len(detections.boxes) > 0"
      },
      "inputs": [
        {
          "id": "in_detections",
          "name": "detections",
          "type": "json",
          "required": true
        }
      ],
      "outputs": [
        {
          "id": "out_true",
          "name": "true",
          "type": "boolean"
        },
        {
          "id": "out_false",
          "name": "false",
          "type": "boolean"
        }
      ]
    },
    {
      "id": "n4",
      "type": "system",
      "plugin": "terminal",
      "name": "输出检测数量",
      "position": { "x": 950, "y": 100 },
      "parameters": {
        "command": "echo \"Detected ${n2.detections.boxes.length} vehicles\""
      },
      "inputs": [
        {
          "id": "in_trigger",
          "name": "trigger",
          "type": "boolean",
          "required": true
        }
      ],
      "outputs": [
        {
          "id": "out_message",
          "name": "message",
          "type": "text"
        }
      ]
    }
  ],
  "edges": [
    {
      "id": "e1",
      "source": { "node_id": "n1", "port_id": "out_image" },
      "target": { "node_id": "n2", "port_id": "in_image" }
    },
    {
      "id": "e2",
      "source": { "node_id": "n2", "port_id": "out_detections" },
      "target": { "node_id": "n3", "port_id": "in_detections" }
    },
    {
      "id": "e3",
      "source": { "node_id": "n3", "port_id": "out_true" },
      "target": { "node_id": "n4", "port_id": "in_trigger" }
    }
  ]
}
```

### 8.2 Transformer 文本分类工作流

```json
{
  "schema_version": "1.0.0",
  "id": "wf_nlp_classification",
  "name": "文本情感分类",
  "description": "输入文本 → Tokenize → Transformer 分类 → 结果输出",
  "project_id": "proj_001",
  "version": 1,
  "nodes": [
    {
      "id": "n1",
      "type": "input",
      "plugin": "data-source",
      "name": "文本输入",
      "position": { "x": 50, "y": 200 },
      "parameters": {},
      "inputs": [],
      "outputs": [
        { "id": "out_text", "name": "text", "type": "text" }
      ]
    },
    {
      "id": "n2",
      "type": "nlp",
      "plugin": "transformer",
      "name": "情感分类",
      "position": { "x": 350, "y": 200 },
      "parameters": {
        "model": "bert-base-chinese",
        "task": "sentiment",
        "labels": ["positive", "negative", "neutral"]
      },
      "inputs": [
        { "id": "in_text", "name": "text", "type": "text", "required": true }
      ],
      "outputs": [
        { "id": "out_result", "name": "result", "type": "json" }
      ]
    }
  ],
  "edges": [
    {
      "id": "e1",
      "source": { "node_id": "n1", "port_id": "out_text" },
      "target": { "node_id": "n2", "port_id": "in_text" }
    }
  ]
}
```

### 8.3 LSTM 时序预测工作流

```json
{
  "schema_version": "1.0.0",
  "id": "wf_lstm_forecast",
  "name": "时序预测",
  "description": "输入时序数据 → LSTM 预测 → 结果可视化",
  "project_id": "proj_001",
  "version": 1,
  "nodes": [
    {
      "id": "n1",
      "type": "input",
      "plugin": "data-source",
      "name": "时序数据",
      "position": { "x": 50, "y": 200 },
      "parameters": {},
      "inputs": [],
      "outputs": [
        { "id": "out_tensor", "name": "data", "type": "tensor" }
      ]
    },
    {
      "id": "n2",
      "type": "timeseries",
      "plugin": "lstm-predict",
      "name": "LSTM 预测",
      "position": { "x": 350, "y": 200 },
      "parameters": {
        "sequence_length": 60,
        "forecast_horizon": 10,
        "hidden_size": 128,
        "num_layers": 2
      },
      "inputs": [
        { "id": "in_data", "name": "data", "type": "tensor", "required": true }
      ],
      "outputs": [
        { "id": "out_prediction", "name": "prediction", "type": "json" }
      ]
    }
  ],
  "edges": [
    {
      "id": "e1",
      "source": { "node_id": "n1", "port_id": "out_tensor" },
      "target": { "node_id": "n2", "port_id": "in_data" }
    }
  ]
}
```

### 8.4 MCP 外部工具调用工作流

```json
{
  "schema_version": "1.0.0",
  "id": "wf_mcp_external",
  "name": "MCP 外部工具调用",
  "description": "通过 MCP 协议调用外部软件（如 MATLAB、SUMO）",
  "project_id": "proj_001",
  "version": 1,
  "nodes": [
    {
      "id": "n1",
      "type": "input",
      "plugin": "data-source",
      "name": "参数输入",
      "position": { "x": 50, "y": 200 },
      "parameters": {},
      "inputs": [],
      "outputs": [
        { "id": "out_params", "name": "params", "type": "json" }
      ]
    },
    {
      "id": "n2",
      "type": "mcp",
      "plugin": "mcp-client",
      "name": "MATLAB 仿真",
      "position": { "x": 350, "y": 200 },
      "parameters": {
        "server": "matlab-simulink",
        "tool": "run_simulation",
        "timeout_ms": 60000
      },
      "inputs": [
        { "id": "in_params", "name": "params", "type": "json", "required": true }
      ],
      "outputs": [
        { "id": "out_result", "name": "result", "type": "json" }
      ]
    }
  ],
  "edges": [
    {
      "id": "e1",
      "source": { "node_id": "n1", "port_id": "out_params" },
      "target": { "node_id": "n2", "port_id": "in_params" }
    }
  ]
}
```

### 8.5 Agent 自动生成工作流

```json
{
  "schema_version": "1.0.0",
  "id": "wf_agent_generated",
  "name": "Agent 生成：图像预处理流水线",
  "description": "由 Agent 根据用户自然语言描述自动生成",
  "project_id": "proj_001",
  "version": 1,
  "metadata": {
    "generator": "agent",
    "agent_session": "sess_001",
    "user_prompt": "创建一个图像预处理流程：裁剪 → 缩放 → 归一化",
    "confidence": 0.92
  },
  "nodes": [
    {
      "id": "n1",
      "type": "input",
      "plugin": "data-source",
      "name": "原始图像",
      "position": { "x": 50, "y": 200 },
      "parameters": {},
      "inputs": [],
      "outputs": [
        { "id": "out_image", "name": "image", "type": "image" }
      ]
    },
    {
      "id": "n2",
      "type": "vision",
      "plugin": "image-crop",
      "name": "图像裁剪",
      "position": { "x": 350, "y": 200 },
      "parameters": {
        "crop_type": "center",
        "ratio": 0.8
      },
      "inputs": [
        { "id": "in_image", "name": "image", "type": "image", "required": true }
      ],
      "outputs": [
        { "id": "out_image", "name": "image", "type": "image" }
      ]
    },
    {
      "id": "n3",
      "type": "vision",
      "plugin": "image-resize",
      "name": "图像缩放",
      "position": { "x": 650, "y": 200 },
      "parameters": {
        "target_size": [640, 640],
        "interpolation": "bilinear"
      },
      "inputs": [
        { "id": "in_image", "name": "image", "type": "image", "required": true }
      ],
      "outputs": [
        { "id": "out_image", "name": "image", "type": "image" }
      ]
    },
    {
      "id": "n4",
      "type": "vision",
      "plugin": "image-normalize",
      "name": "归一化",
      "position": { "x": 950, "y": 200 },
      "parameters": {
        "mean": [0.485, 0.456, 0.406],
        "std": [0.229, 0.224, 0.225]
      },
      "inputs": [
        { "id": "in_image", "name": "image", "type": "image", "required": true }
      ],
      "outputs": [
        { "id": "out_tensor", "name": "tensor", "type": "tensor" }
      ]
    }
  ],
  "edges": [
    {
      "id": "e1",
      "source": { "node_id": "n1", "port_id": "out_image" },
      "target": { "node_id": "n2", "port_id": "in_image" }
    },
    {
      "id": "e2",
      "source": { "node_id": "n2", "port_id": "out_image" },
      "target": { "node_id": "n3", "port_id": "in_image" }
    },
    {
      "id": "e3",
      "source": { "node_id": "n3", "port_id": "out_image" },
      "target": { "node_id": "n4", "port_id": "in_image" }
    }
  ]
}
```

---

## 9. Schema 校验规则

Backend 在保存和执行前执行以下校验：

| 规则 | 说明 |
|------|------|
| DAG 校验 | 图必须是有向无环图，不允许环路 |
| 端口类型兼容 | 连线两端的端口类型必须兼容 |
| 必填端口 | `required: true` 的输入端口必须有连线 |
| 节点 ID 唯一 | 同一工作流内节点 ID 不可重复 |
| 插件存在性 | 引用的插件必须已安装 |
| 参数合法性 | 根据 Plugin config_schema 校验参数 |

---

## 10. 版本兼容策略

| 版本变化 | 兼容性 | 处理方式 |
|---------|--------|---------|
| 新增可选字段 | 向后兼容 | 旧版本忽略未知字段 |
| 新增节点类型 | 向后兼容 | 旧版本渲染为 "unsupported" 节点 |
| 删除字段 | 不兼容 | 升级 schema_version，提供迁移脚本 |
| 修改端口类型 | 不兼容 | 强制要求重新校验连线 |
