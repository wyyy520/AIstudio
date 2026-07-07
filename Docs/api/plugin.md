# 插件管理接口

## GET /plugins

获取已安装的插件列表。

### 查询参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 否 | 按类型筛选（vision/nlp/logic/system/simulation/mcp） |

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "name": "yolo-detector",
        "version": "1.0.0",
        "type": "vision",
        "author": "AI Studio",
        "description": "YOLO 目标检测插件",
        "icon": "yolo.png",
        "language": "python",
        "enabled": true,
        "loaded": true,
        "ports": {
          "inputs": [
            {"name": "image", "type": "image", "required": true},
            {"name": "confidence", "type": "number", "required": false, "default": 0.5}
          ],
          "outputs": [
            {"name": "detections", "type": "json"}
          ]
        }
      }
    ]
  }
}
```

---

## GET /plugins/:name

获取插件详情。

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "name": "yolo-detector",
    "version": "1.0.0",
    "type": "vision",
    "author": "AI Studio",
    "description": "YOLO 目标检测插件",
    "icon": "yolo.png",
    "language": "python",
    "entry": "main.py",
    "enabled": true,
    "loaded": true,
    "ports": {
      "inputs": [
        {"name": "image", "type": "image", "required": true, "description": "输入图像"},
        {"name": "confidence", "type": "number", "required": false, "default": 0.5, "description": "置信度阈值"}
      ],
      "outputs": [
        {"name": "detections", "type": "json", "description": "检测结果 {boxes, scores, classes}"}
      ]
    },
    "config_schema": {
      "model": {"type": "string", "default": "yolov8n.pt", "description": "模型文件"},
      "device": {"type": "string", "default": "auto", "options": ["auto", "cpu", "cuda"], "description": "推理设备"}
    }
  }
}
```

---

## POST /plugins/install

安装插件。

### 请求体

```json
{
  "source": "local",
  "path": "/storage/plugins/custom-plugin/"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| source | string | 是 | 安装来源（local / url / registry） |
| path | string | 是 | 本地路径或远程 URL |

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "name": "custom-plugin",
    "version": "0.1.0",
    "type": "vision",
    "enabled": true
  }
}
```

---

## DELETE /plugins/:name

卸载插件。

### 响应

```json
{
  "code": 0,
  "message": "success"
}
```

---

## GET /plugins/:name/config-schema

获取插件配置 Schema（用于前端动态生成配置表单）。

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "fields": [
      {
        "name": "model",
        "type": "string",
        "label": "模型文件",
        "default": "yolov8n.pt",
        "description": "YOLO 模型权重文件路径"
      },
      {
        "name": "device",
        "type": "select",
        "label": "推理设备",
        "default": "auto",
        "options": [
          {"value": "auto", "label": "自动"},
          {"value": "cpu", "label": "CPU"},
          {"value": "cuda", "label": "GPU (CUDA)"}
        ]
      },
      {
        "name": "confidence",
        "type": "number",
        "label": "置信度阈值",
        "default": 0.5,
        "min": 0,
        "max": 1,
        "step": 0.05
      }
    ]
  }
}
```

---

## POST /plugins/:name/test

测试插件执行（独立运行，不依赖工作流）。

### 请求体

```json
{
  "inputs": {
    "image": "/storage/datasets/test.jpg",
    "confidence": 0.5
  },
  "config": {
    "model": "yolov8n.pt",
    "device": "cuda"
  }
}
```

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "output": {
      "detections": {
        "boxes": [[100, 200, 300, 400]],
        "scores": [0.95],
        "classes": [2]
      }
    },
    "metrics": {
      "duration_ms": 120,
      "memory_mb": 256,
      "device": "cuda"
    }
  }
}
```
