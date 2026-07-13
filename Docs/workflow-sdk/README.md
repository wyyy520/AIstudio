# 工作流 SDK 文档

## 1. 概述

工作流 SDK 定义了 AI Studio 中工作流的图结构、节点类型、执行引擎等核心规范。开发者可以通过 SDK 编程方式创建工作流、添加节点、定义连线。

## 2. 工作流数据结构

```json
{
  "id": "wf_001",
  "name": "车辆检测工作流",
  "version": 1,
  "graph": {
    "nodes": [ ... ],
    "edges": [ ... ]
  }
}
```

## 3. 节点 (Node)

### 节点基本结构

```json
{
  "id": "n1",
  "type": "vision",
  "plugin": "yolo",
  "label": "YOLO 检测",
  "position": {"x": 100, "y": 200},
  "config": {
    "model": "yolov8n.pt",
    "confidence": 0.5
  },
  "ports": {
    "inputs": [
      {"name": "image", "type": "image", "required": true}
    ],
    "outputs": [
      {"name": "detections", "type": "json"}
    ]
  }
}
```

### 节点 ID 规范

- 格式：`n` + 数字，如 `n1`, `n2`, `n10`
- 推荐：按创建顺序自增
- 连线时使用 `from`/`to` 引用节点 ID

## 4. 连线 (Edge)

### 连线基本结构

```json
{
  "id": "e1",
  "from": "n1",
  "to": "n2",
  "from_port": "detections",
  "to_port": "input"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 连线唯一标识符 |
| from | string | 源节点 ID |
| to | string | 目标节点 ID |
| from_port | string | 源节点输出端口名 |
| to_port | string | 目标节点输入端口名 |
| label | string | 连线标签（可选） |

## 5. 编程方式创建工作流

### Python SDK

```python
from aistudio_sdk import WorkflowBuilder

# 创建工作流
builder = WorkflowBuilder("车辆检测工作流")

# 添加节点
builder.add_node(
    node_id="n1",
    node_type="input",
    plugin="image_input",
    position=(100, 200)
)

builder.add_node(
    node_id="n2",
    node_type="vision",
    plugin="yolo",
    position=(400, 200),
    config={"model": "yolov8n.pt", "confidence": 0.5}
)

builder.add_node(
    node_id="n3",
    node_type="output",
    plugin="json_output",
    position=(700, 200)
)

# 添加连线
builder.add_edge(from_node="n1", from_port="image",
                 to_node="n2", to_port="image")

builder.add_edge(from_node="n2", from_port="detections",
                 to_node="n3", to_port="input")

# 导出工作流 JSON
workflow_json = builder.build()
print(workflow_json)
```

### Go SDK

```go
import "github.com/aistudio/sdk/go/workflow"

func BuildWorkflow() *workflow.Workflow {
    wf := workflow.New("车辆检测工作流")
    
    // 添加节点
    wf.AddNode(workflow.Node{
        ID:     "n1",
        Type:   "input",
        Plugin: "image_input",
        Position: &workflow.Position{X: 100, Y: 200},
    })
    
    wf.AddNode(workflow.Node{
        ID:     "n2",
        Type:   "vision",
        Plugin: "yolo",
        Position: &workflow.Position{X: 400, Y: 200},
        Config: map[string]interface{}{
            "model": "yolov8n.pt",
            "confidence": 0.5,
        },
    })
    
    // 添加连线
    wf.AddEdge("n1", "image", "n2", "image")
    
    return wf
}
```

## 6. 执行模式

| 模式 | 说明 | 适用场景 |
|------|------|----------|
| 同步执行 | 等待所有节点完成后返回 | 批处理、离线任务 |
| 流式执行 | 节点完成后立即推送结果 | 实时预览、LLM 对话 |
| 调试执行 | 逐步执行，可单步调试 | 工作流开发调试 |
| 断点续跑 | 从失败节点恢复 | 长时任务容错 |
