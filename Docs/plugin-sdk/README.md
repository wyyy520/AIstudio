# 插件 SDK 开发文档

## 1. 概述

AI Studio 插件 SDK 用于开发可接入工作流编辑器的自定义功能节点。插件支持 Python 和 Go 两种语言。

## 2. 插件目录结构

```
Plugins/
└── Vision/
    └── YOLO/
        ├── plugin.json      # 插件清单（必需）
        ├── main.py          # 入口文件
        ├── requirements.txt # Python 依赖
        ├── README.md        # 插件说明
        └── assets/          # 图标等资源
            └── icon.png
```

## 3. 插件清单 (plugin.json)

```json
{
  "name": "yolo-detector",
  "version": "1.0.0",
  "author": "AI Studio",
  "description": "YOLO 目标检测插件",
  "type": "vision",
  "icon": "assets/icon.png",
  "language": "python",
  "entry": "main.py",
  "ports": {
    "inputs": [
      {
        "name": "image",
        "type": "image",
        "required": true,
        "description": "输入图像"
      },
      {
        "name": "confidence",
        "type": "number",
        "required": false,
        "default": 0.5,
        "description": "置信度阈值 (0-1)"
      }
    ],
    "outputs": [
      {
        "name": "detections",
        "type": "json",
        "description": "检测结果"
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
      "default": "yolov8n.pt",
      "description": "模型权重文件"
    },
    "device": {
      "type": "select",
      "default": "auto",
      "options": ["auto", "cpu", "cuda"],
      "description": "推理设备"
    }
  }
}
```

## 4. 数据类型系统

| 类型 | 标识 | Python 类型 | 说明 |
|------|------|-------------|------|
| image | `image` | str (文件路径) | 图像，以文件路径传递 |
| text | `text` | str | 文本字符串 |
| number | `number` | float / int | 数值 |
| json | `json` | dict / list | 结构化数据 |
| file | `file` | str (文件路径) | 任意文件 |
| tensor | `tensor` | numpy.ndarray | 张量（序列化传输） |
| stream | `stream` | Generator | 流式数据（LLM 等） |

## 5. Python 插件开发

### 5.1 基类

```python
from abc import ABC, abstractmethod
from typing import Any, Dict

class AIStudioPlugin(ABC):
    """AI Studio 插件基类"""
    
    @abstractmethod
    def execute(self, inputs: Dict[str, Any], config: Dict[str, Any]) -> Dict[str, Any]:
        """
        执行插件逻辑
        
        Args:
            inputs: 输入端口数据，key 为端口名
            config: 节点配置参数
        
        Returns:
            输出端口数据，key 为端口名
        """
        pass
    
    def setup(self, config: Dict[str, Any]):
        """初始化（可选，模型加载等）"""
        pass
    
    def teardown(self):
        """清理资源（可选）"""
        pass
```

### 5.2 完整示例：YOLO 检测插件

```python
# main.py
from ultralytics import YOLO
import cv2
import os

class YOLODetector(AIStudioPlugin):
    
    def setup(self, config: Dict[str, Any]):
        model_path = config.get("model", "yolov8n.pt")
        device = config.get("device", "auto")
        if device == "auto":
            device = "cuda" if torch.cuda.is_available() else "cpu"
        
        self.model = YOLO(model_path)
        self.model.to(device)
        self.device = device
    
    def execute(self, inputs: Dict[str, Any], config: Dict[str, Any]) -> Dict[str, Any]:
        image_path = inputs["image"]
        confidence = inputs.get("confidence", 0.5)
        
        # 推理
        results = self.model(image_path, conf=confidence)
        
        # 标注图像
        annotated = results[0].plot()
        annotated_path = os.path.join(
            os.path.dirname(image_path),
            "annotated_" + os.path.basename(image_path)
        )
        cv2.imwrite(annotated_path, annotated)
        
        # 返回输出
        return {
            "detections": {
                "boxes": results[0].boxes.xyxy.tolist(),
                "scores": results[0].boxes.conf.tolist(),
                "classes": results[0].boxes.cls.tolist(),
            },
            "annotated_image": annotated_path
        }
    
    def teardown(self):
        del self.model

# 插件入口
plugin = YOLODetector()
```

### 5.3 流式插件示例（LLM）

```python
# main.py
from typing import Generator

class LLMChatPlugin(AIStudioPlugin):
    
    def setup(self, config: Dict[str, Any]):
        from transformers import AutoModelForCausalLM, AutoTokenizer
        model_name = config.get("model", "gpt2")
        self.tokenizer = AutoTokenizer.from_pretrained(model_name)
        self.model = AutoModelForCausalLM.from_pretrained(model_name)
    
    def execute(self, inputs: Dict[str, Any], config: Dict[str, Any]) -> Dict[str, Any]:
        prompt = inputs["text"]
        max_tokens = config.get("max_tokens", 512)
        
        inputs_ids = self.tokenizer(prompt, return_tensors="pt")
        output = self.model.generate(**inputs_ids, max_new_tokens=max_tokens)
        text = self.tokenizer.decode(output[0], skip_special_tokens=True)
        
        return {"text": text}
    
    def execute_stream(self, inputs: Dict[str, Any], config: Dict[str, Any]) -> Generator[str, None, None]:
        """流式输出（用于 WebSocket 推送）"""
        prompt = inputs["text"]
        inputs_ids = self.tokenizer(prompt, return_tensors="pt")
        
        for token in self.model.generate(**inputs_ids, streamer=True):
            yield self.tokenizer.decode(token, skip_special_tokens=True)
```

## 6. Go 插件开发

```go
package main

import (
    "context"
    "github.com/aistudio/backend/pkg/plugin"
)

type GitPlugin struct{}

func (p *GitPlugin) Meta() plugin.PluginMeta {
    return plugin.PluginMeta{
        Name:        "git-operations",
        Version:     "1.0.0",
        Author:      "AI Studio",
        Description: "Git 操作插件",
        Type:        plugin.PluginSystem,
    }
}

func (p *GitPlugin) Ports() plugin.PluginPorts {
    return plugin.PluginPorts{
        Inputs: []plugin.Port{
            {Name: "command", Type: plugin.PortText, Required: true, Description: "Git 命令"},
            {Name: "repo_path", Type: plugin.PortFile, Required: false, Description: "仓库路径"},
        },
        Outputs: []plugin.Port{
            {Name: "output", Type: plugin.PortText, Description: "命令输出"},
            {Name: "exit_code", Type: plugin.PortNumber, Description: "退出码"},
        },
    }
}

func (p *GitPlugin) Execute(ctx context.Context, input *plugin.PluginInput) (*plugin.PluginOutput, error) {
    cmd := input.Data["command"].(string)
    repoPath, _ := input.Data["repo_path"].(string)
    
    // 执行 git 命令
    output, err := execGitCommand(ctx, cmd, repoPath)
    if err != nil {
        return &plugin.PluginOutput{
            Data:   map[string]interface{}{"output": err.Error(), "exit_code": 1},
            Status: "error",
            Error:  err.Error(),
        }, nil
    }
    
    return &plugin.PluginOutput{
        Data:   map[string]interface{}{"output": output, "exit_code": 0},
        Status: "success",
    }, nil
}

// 导出插件实例
var Plugin plugin.Plugin = &GitPlugin{}
```

## 7. 插件测试

```python
# test_plugin.py
import unittest
from main import YOLODetector

class TestYOLODetector(unittest.TestCase):
    
    def setUp(self):
        self.plugin = YOLODetector()
        self.plugin.setup({"model": "yolov8n.pt", "device": "cpu"})
    
    def test_detect(self):
        result = self.plugin.execute(
            inputs={"image": "test_data/test.jpg", "confidence": 0.5},
            config={"model": "yolov8n.pt", "device": "cpu"}
        )
        self.assertIn("detections", result)
        self.assertIn("annotated_image", result)
        self.assertTrue(os.path.exists(result["annotated_image"]))
    
    def tearDown(self):
        self.plugin.teardown()

if __name__ == "__main__":
    unittest.main()
```

## 8. 插件发布

1. 确保目录结构完整（plugin.json + 入口文件）
2. 编写 README.md
3. 填写 requirements.txt（Python 插件）
4. 将插件目录放入 `Plugins/<类型>/` 下
5. 在 AI Studio 插件管理界面点击"刷新"，自动发现新插件
