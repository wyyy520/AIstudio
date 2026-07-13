# 视觉插件开发指南

## 1. 支持的视觉插件类型

| 插件 | 目录 | 功能 |
|------|------|------|
| YOLO | Plugins/Vision/YOLO/ | 目标检测（实时） |
| RT-DETR | Plugins/Vision/RT-DETR/ | 目标检测（高精度） |
| SAM | Plugins/Vision/SAM/ | 图像分割（任意目标） |
| OCR | Plugins/Vision/OCR/ | 文字识别 |

## 2. 通用端口定义

### 输入端口

| 端口名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| image | image | 是 | 输入图像（文件路径） |
| confidence | number | 否 | 置信度阈值，默认 0.5 |

### 输出端口

| 端口名 | 类型 | 说明 |
|--------|------|------|
| detections | json | 检测结果 |
| annotated_image | image | 标注后的图像 |

### detections 数据结构

```json
{
  "boxes": [[x1, y1, x2, y2], ...],
  "scores": [0.95, 0.87, ...],
  "classes": [0, 2, 7, ...],
  "labels": ["person", "car", "truck", ...]
}
```

## 3. YOLO 插件开发

```python
from ultralytics import YOLO
import cv2
import os

class YOLOPlugin:
    
    def setup(self, config):
        model_path = config.get("model", "yolov8n.pt")
        device = config.get("device", "auto")
        self.model = YOLO(model_path)
        if device != "auto":
            self.model.to(device)
    
    def execute(self, inputs, config):
        image_path = inputs["image"]
        conf = inputs.get("confidence", 0.5)
        
        results = self.model(image_path, conf=conf)
        result = results[0]
        
        # 标注图像
        annotated = result.plot()
        annotated_path = self._save_annotated(image_path, annotated)
        
        return {
            "detections": {
                "boxes": result.boxes.xyxy.tolist(),
                "scores": result.boxes.conf.tolist(),
                "classes": result.boxes.cls.tolist(),
                "labels": [self.model.names[int(c)] for c in result.boxes.cls]
            },
            "annotated_image": annotated_path
        }
    
    def _save_annotated(self, original_path, image):
        dir_name = os.path.dirname(original_path)
        base_name = os.path.basename(original_path)
        output_path = os.path.join(dir_name, f"annotated_{base_name}")
        cv2.imwrite(output_path, image)
        return output_path
```

## 4. SAM 插件开发

```python
from segment_anything import sam_model_registry, SamPredictor
import cv2
import numpy as np
import torch

class SAMPlugin:
    
    def setup(self, config):
        model_type = config.get("model_type", "vit_h")
        checkpoint = config.get("checkpoint", "sam_vit_h.pth")
        device = "cuda" if torch.cuda.is_available() else "cpu"
        
        sam = sam_model_registry[model_type](checkpoint=checkpoint)
        sam.to(device)
        self.predictor = SamPredictor(sam)
    
    def execute(self, inputs, config):
        image_path = inputs["image"]
        points = inputs.get("points", [])  # [[x, y], ...]
        
        image = cv2.imread(image_path)
        image_rgb = cv2.cvtColor(image, cv2.COLOR_BGR2RGB)
        
        self.predictor.set_image(image_rgb)
        
        if points:
            input_points = np.array(points)
            input_labels = np.ones(len(points))
            
            masks, scores, _ = self.predictor.predict(
                point_coords=input_points,
                point_labels=input_labels,
                multimask_output=True
            )
            
            return {
                "detections": {
                    "masks": masks.tolist(),
                    "scores": scores.tolist()
                }
            }
        
        return {"detections": {"masks": [], "scores": []}}
```

## 5. OCR 插件开发

```python
import easyocr

class OCRPlugin:
    
    def setup(self, config):
        languages = config.get("languages", ["ch_sim", "en"])
        self.reader = easyocr.Reader(languages)
    
    def execute(self, inputs, config):
        image_path = inputs["image"]
        
        results = self.reader.readtext(image_path)
        
        texts = []
        boxes = []
        for (bbox, text, conf) in results:
            texts.append(text)
            boxes.append({
                "bbox": [[int(p[0]), int(p[1])] for p in bbox],
                "text": text,
                "confidence": float(conf)
            })
        
        return {
            "detections": {
                "full_text": "\n".join(texts),
                "regions": boxes
            }
        }
```

## 6. 配置 Schema 参考

```json
{
  "config_schema": {
    "model": {
      "type": "string",
      "default": "yolov8n.pt",
      "description": "模型权重文件路径"
    },
    "device": {
      "type": "select",
      "default": "auto",
      "options": [
        {"value": "auto", "label": "自动"},
        {"value": "cpu", "label": "CPU"},
        {"value": "cuda", "label": "GPU (CUDA)"}
      ],
      "description": "推理设备"
    },
    "confidence": {
      "type": "number",
      "default": 0.5,
      "min": 0,
      "max": 1,
      "step": 0.05,
      "description": "置信度阈值"
    },
    "iou_threshold": {
      "type": "number",
      "default": 0.45,
      "min": 0,
      "max": 1,
      "step": 0.05,
      "description": "NMS IOU 阈值"
    }
  }
}
```

## 7. 性能优化建议

| 优化项 | 方法 |
|--------|------|
| 模型预加载 | 在 `setup` 中加载，避免每次执行重新加载 |
| 批量推理 | 支持多图输入，使用 batch inference |
| 半精度推理 | 使用 FP16 减少显存占用 |
| 模型量化 | INT8 量化提升 CPU 推理速度 |
| 图像缩放 | 输入前 resize 到模型输入尺寸 |
| 结果缓存 | 相同输入参数的推理结果可缓存 |
