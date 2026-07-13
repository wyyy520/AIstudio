# 快速开始：开发你的第一个插件

## 1. 创建插件目录

```
Plugins/
└── Vision/
    └── MyDetector/
        ├── plugin.json
        ├── main.py
        ├── requirements.txt
        └── README.md
```

## 2. 编写 plugin.json

```json
{
  "name": "my-detector",
  "version": "0.1.0",
  "author": "你的名字",
  "description": "我的第一个检测插件",
  "type": "vision",
  "icon": "",
  "language": "python",
  "entry": "main.py",
  "ports": {
    "inputs": [
      {
        "name": "image",
        "type": "image",
        "required": true,
        "description": "输入图像"
      }
    ],
    "outputs": [
      {
        "name": "result",
        "type": "json",
        "description": "检测结果"
      }
    ]
  },
  "config_schema": {
    "threshold": {
      "type": "number",
      "default": 0.5,
      "min": 0,
      "max": 1,
      "description": "检测阈值"
    }
  }
}
```

## 3. 编写 main.py

```python
import cv2
import numpy as np

class MyDetector:
    
    def setup(self, config):
        self.threshold = config.get("threshold", 0.5)
    
    def execute(self, inputs, config):
        image_path = inputs["image"]
        image = cv2.imread(image_path)
        
        # 简单边缘检测示例
        gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
        edges = cv2.Canny(gray, 100, 200)
        contours, _ = cv2.findContours(edges, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        
        boxes = []
        for cnt in contours:
            x, y, w, h = cv2.boundingRect(cnt)
            if w * h > 100:  # 过滤小区域
                boxes.append([x, y, x + w, y + h])
        
        return {
            "result": {
                "count": len(boxes),
                "boxes": boxes
            }
        }

plugin = MyDetector()
```

## 4. 编写 requirements.txt

```
opencv-python>=4.8.0
numpy>=1.24.0
```

## 5. 测试插件

在 AI Studio 中：

1. 打开插件管理页面
2. 点击"刷新插件列表"
3. 确认 `my-detector` 出现在列表中
4. 点击"测试"按钮
5. 选择一张测试图片
6. 查看输出结果

## 6. 在工作流中使用

1. 打开工作流编辑器
2. 从左侧节点面板拖入 `my-detector` 节点
3. 添加一个图像输入节点并连接
4. 添加一个输出节点并连接 `result` 端口
5. 点击运行

## 7. 调试技巧

- 在 `execute` 方法中使用 `print()` 输出调试信息（会显示在日志面板）
- 使用 `config` 参数获取用户在属性面板配置的值
- 大文件通过文件路径传递，不要直接传 base64
- 模型在 `setup` 中加载，避免每次执行都重新加载
