# 工作流示例

## 1. 目标检测工作流

### 描述

读取图像 → YOLO 检测 → 条件判断（有目标？）→ 保存结果

### 节点图

```
[图像输入] → [YOLO 检测] → [If: 有目标?] → true → [保存结果]
                                        └→ false → [日志: 未检测到]
```

### JSON 定义

```json
{
  "id": "wf_detect",
  "name": "目标检测工作流",
  "graph": {
    "nodes": [
      {
        "id": "n1",
        "type": "input",
        "plugin": "image_input",
        "position": {"x": 100, "y": 200},
        "ports": {
          "outputs": [{"name": "image", "type": "image"}]
        }
      },
      {
        "id": "n2",
        "type": "vision",
        "plugin": "yolo",
        "position": {"x": 400, "y": 200},
        "config": {"model": "yolov8n.pt", "confidence": 0.5},
        "ports": {
          "inputs": [{"name": "image", "type": "image", "required": true}],
          "outputs": [{"name": "detections", "type": "json"}]
        }
      },
      {
        "id": "n3",
        "type": "logic",
        "plugin": "if",
        "position": {"x": 700, "y": 200},
        "config": {"condition": "len(input.get('boxes', [])) > 0"},
        "ports": {
          "inputs": [{"name": "input", "type": "json", "required": true}],
          "outputs": [{"name": "true", "type": "json"}, {"name": "false", "type": "json"}]
        }
      },
      {
        "id": "n4",
        "type": "system",
        "plugin": "file_save",
        "position": {"x": 1000, "y": 100},
        "ports": {
          "inputs": [{"name": "data", "type": "json", "required": true}]
        }
      },
      {
        "id": "n5",
        "type": "system",
        "plugin": "log",
        "position": {"x": 1000, "y": 300},
        "ports": {
          "inputs": [{"name": "message", "type": "text", "required": true}]
        }
      }
    ],
    "edges": [
      {"id": "e1", "from": "n1", "to": "n2", "from_port": "image", "to_port": "image"},
      {"id": "e2", "from": "n2", "to": "n3", "from_port": "detections", "to_port": "input"},
      {"id": "e3", "from": "n3", "to": "n4", "from_port": "true", "to_port": "data"},
      {"id": "e4", "from": "n3", "to": "n5", "from_port": "false", "to_port": "message"}
    ]
  }
}
```

---

## 2. LLM 对话工作流

### 描述

文本输入 → LLM 生成回复 → 流式输出

### 节点图

```
[文本输入] → [LLM 对话] → [流式输出]
```

### JSON 定义

```json
{
  "id": "wf_llm_chat",
  "name": "LLM 对话工作流",
  "graph": {
    "nodes": [
      {
        "id": "n1",
        "type": "input",
        "plugin": "text_input",
        "position": {"x": 100, "y": 200},
        "ports": {
          "outputs": [{"name": "text", "type": "text"}]
        }
      },
      {
        "id": "n2",
        "type": "nlp",
        "plugin": "llm",
        "position": {"x": 400, "y": 200},
        "config": {
          "model": "Qwen/Qwen2-7B",
          "temperature": 0.7,
          "max_tokens": 1024,
          "stream": true
        },
        "ports": {
          "inputs": [{"name": "prompt", "type": "text", "required": true}],
          "outputs": [{"name": "response", "type": "text"}, {"name": "usage", "type": "json"}]
        }
      },
      {
        "id": "n3",
        "type": "output",
        "plugin": "stream_output",
        "position": {"x": 700, "y": 200},
        "ports": {
          "inputs": [{"name": "text", "type": "text", "required": true}]
        }
      }
    ],
    "edges": [
      {"id": "e1", "from": "n1", "to": "n2", "from_port": "text", "to_port": "prompt"},
      {"id": "e2", "from": "n2", "to": "n3", "from_port": "response", "to_port": "text"}
    ]
  }
}
```

---

## 3. 批量图像处理工作流

### 描述

数据集列表 → Loop 迭代 → 逐个图像检测 → 汇总结果

### 节点图

```
[数据集列表] → [Loop: 迭代] → [YOLO 检测] → [汇总结果] → [保存 CSV]
```

### JSON 定义

```json
{
  "id": "wf_batch_detect",
  "name": "批量检测工作流",
  "graph": {
    "nodes": [
      {
        "id": "n1",
        "type": "input",
        "plugin": "dataset_input",
        "position": {"x": 100, "y": 200},
        "ports": {
          "outputs": [{"name": "items", "type": "json"}]
        }
      },
      {
        "id": "n2",
        "type": "logic",
        "plugin": "loop",
        "position": {"x": 400, "y": 200},
        "config": {"mode": "iterate"},
        "ports": {
          "inputs": [{"name": "input", "type": "json", "required": true}],
          "outputs": [{"name": "output", "type": "json"}, {"name": "completed", "type": "json"}]
        }
      },
      {
        "id": "n3",
        "type": "vision",
        "plugin": "yolo",
        "position": {"x": 700, "y": 200},
        "config": {"model": "yolov8n.pt"},
        "ports": {
          "inputs": [{"name": "image", "type": "image", "required": true}],
          "outputs": [{"name": "detections", "type": "json"}]
        }
      },
      {
        "id": "n4",
        "type": "system",
        "plugin": "collect",
        "position": {"x": 1000, "y": 200},
        "ports": {
          "inputs": [{"name": "item", "type": "json"}],
          "outputs": [{"name": "all", "type": "json"}]
        }
      },
      {
        "id": "n5",
        "type": "system",
        "plugin": "csv_export",
        "position": {"x": 1300, "y": 200},
        "ports": {
          "inputs": [{"name": "data", "type": "json", "required": true}]
        }
      }
    ],
    "edges": [
      {"id": "e1", "from": "n1", "to": "n2", "from_port": "items", "to_port": "input"},
      {"id": "e2", "from": "n2", "to": "n3", "from_port": "output", "to_port": "image"},
      {"id": "e3", "from": "n3", "to": "n4", "from_port": "detections", "to_port": "item"},
      {"id": "e4", "from": "n2", "to": "n4", "from_port": "completed", "to_port": "_trigger"},
      {"id": "e5", "from": "n4", "to": "n5", "from_port": "all", "to_port": "data"}
    ]
  }
}
```

---

## 4. 模型训练工作流

### 描述

数据集 → 预处理 → 划分训练/验证集 → 训练 → 评估

### 节点图

```
[数据集] → [预处理] → [划分] → [训练] → [评估] → [保存模型]
```

---

## 5. 仿真工作流（智能交通）

### 描述

交通数据 → SUMO 仿真 → 结果分析 → 可视化

### 节点图

```
[交通数据] → [SUMO 仿真] → [结果读取] → [时序分析] → [可视化图表]
```
