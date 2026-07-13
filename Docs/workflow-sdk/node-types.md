# 节点类型定义

## 1. 节点分类

| 分类 | 说明 | 示例 |
|------|------|------|
| 输入节点 | 工作流的数据入口 | 图像输入、文本输入、文件输入 |
| 处理节点 | 核心处理单元（插件） | YOLO 检测、LLM 对话、条件判断 |
| 输出节点 | 工作流的数据出口 | JSON 输出、文件输出、可视化输出 |
| 控制节点 | 流程控制 | If/Else/Switch/Loop/Retry |
| 注释节点 | 流程图注释（不执行） | 文字说明 |

## 2. 输入节点

### 图像输入节点

```json
{
  "id": "n_input_img",
  "type": "input",
  "plugin": "image_input",
  "ports": {
    "outputs": [
      {"name": "image", "type": "image", "description": "输出图像路径"}
    ]
  },
  "config": {
    "source": "file"  // file / camera / url
  }
}
```

### 文本输入节点

```json
{
  "id": "n_input_text",
  "type": "input",
  "plugin": "text_input",
  "ports": {
    "outputs": [
      {"name": "text", "type": "text", "description": "输出文本"}
    ]
  }
}
```

### 文件输入节点

```json
{
  "id": "n_input_file",
  "type": "input",
  "plugin": "file_input",
  "ports": {
    "outputs": [
      {"name": "file", "type": "file", "description": "输出文件路径"}
    ]
  }
}
```

## 3. 处理节点（插件节点）

### 视觉处理节点

| 插件 | 输入端口 | 输出端口 |
|-------|-----------|-----------|
| yolo | image, confidence | detections, annotated_image |
| rt-detr | image, confidence | detections |
| sam | image, points | masks |
| ocr | image, languages | text, regions |

### NLP 处理节点

| 插件 | 输入端口 | 输出端口 |
|-------|-----------|-----------|
| transformer | text, task | result |
| llm | prompt, context, max_tokens | response, usage |

### 逻辑控制节点

| 插件 | 输入端口 | 输出端口 |
|-------|-----------|-----------|
| if | input, condition | true, false |
| switch | input, key, cases | case_0, case_1, ..., default |
| loop | input, count/mode | output, completed |
| retry | input, max_retries | output, failed |

## 4. 输出节点

### JSON 输出节点

```json
{
  "id": "n_output_json",
  "type": "output",
  "plugin": "json_output",
  "ports": {
    "inputs": [
      {"name": "input", "type": "json", "required": true}
    ]
  }
}
```

### 文件输出节点

```json
{
  "id": "n_output_file",
  "type": "output",
  "plugin": "file_output",
  "ports": {
    "inputs": [
      {"name": "input", "type": "file", "required": true}
    ]
  },
  "config": {
    "output_path": "/storage/results/"
  }
}
```

### 可视化输出节点

```json
{
  "id": "n_output_viz",
  "type": "output",
  "plugin": "visualization",
  "ports": {
    "inputs": [
      {"name": "image", "type": "image"},
      {"name": "data", "type": "json"}
    ]
  }
}
```

## 5. 控制节点

### If 节点

```json
{
  "id": "n_if",
  "type": "logic",
  "plugin": "if",
  "ports": {
    "inputs": [
      {"name": "input", "type": "json", "required": true}
    ],
    "outputs": [
      {"name": "true", "type": "json"},
      {"name": "false", "type": "json"}
    ]
  },
  "config": {
    "condition": "len(input.get('boxes', [])) > 0"
  }
}
```

### Loop 节点

```json
{
  "id": "n_loop",
  "type": "logic",
  "plugin": "loop",
  "ports": {
    "inputs": [
      {"name": "input", "type": "json", "required": true},
      {"name": "count", "type": "number"}
    ],
    "outputs": [
      {"name": "output", "type": "json"},
      {"name": "completed", "type": "json"}
    ]
  },
  "config": {
    "mode": "iterate"  // iterate / count / while
  }
}
```

## 6. 注释节点

```json
{
  "id": "n_note",
  "type": "note",
  "label": "预处理步骤：\n1. 图像缩放\n2. 归一化\n3. 转 tensor",
  "position": {"x": 250, "y": 100},
  "style": {
    "width": 200,
    "height": 100,
    "background": "#fff3cd",
    "border": "1px solid #ffc107"
  }
}
```

## 7. 节点样式规范

| 属性 | 类型 | 说明 |
|------|------|------|
| background | string | 节点背景色 |
| border | string | 边框样式 |
| width | number | 节点宽度（px） |
| height | number | 节点高度（px） |
| icon | string | 节点图标 URL |
| color | string | 文字颜色 |

### 节点类型默认颜色

| 类型 | 默认背景色 |
|------|------------|
| input | #e3f2fd |
| vision | #e8f5e9 |
| nlp | #fce4ec |
| logic | #fff9c4 |
| system | #f3e5f5 |
| simulation | #e0f7fa |
| output | #fbe9e7 |
| note | #fff3cd |
