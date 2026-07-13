# Python 执行引擎架构 (Engine)

## 1. 技术选型

| 技术 | 用途 |
|------|------|
| Python 3.11+ | 引擎主语言 |
| PyTorch 2.1+ | 深度学习推理框架 |
| Transformers | NLP 模型加载 |
| Ultralytics | YOLO 等视觉模型 |
| OpenCV | 图像处理 |
| gRPC | 与 Go Backend 通信 |

## 2. 目录结构说明

```
Engine/
├── vision/         # 视觉推理模块
├── nlp/            # 自然语言处理模块
├── speech/         # 语音处理模块
├── timeseries/     # 时序分析模块
├── deploy/         # 模型部署工具
├── runtime/        # 运行时管理
├── sdk/            # 插件 SDK（Python 端）
└── requirements.txt
```

## 3. 架构设计

```
┌──────────────────────────────────────────┐
│              gRPC Server                  │
│  ┌────────────────────────────────────┐  │
│  │         Request Router              │  │
│  │  ┌─────────┐ ┌─────────┐          │  │
│  │  │ Vision  │ │  NLP    │          │  │
│  │  │ Handler │ │ Handler │          │  │
│  │  └────┬────┘ └────┬────┘          │  │
│  │  ┌─────┴──────────┴────────┐      │  │
│  │  │    Model Manager          │      │  │
│  │  │  ┌──────────────────┐    │      │  │
│  │  │  │  Model Registry   │    │      │  │  模型注册/查找
│  │  │  └────────┬─────────┘    │      │  │
│  │  │  ┌────────┴─────────┐    │      │  │
│  │  │  │  Model Loader     │    │      │  │  模型加载/缓存
│  │  │  └────────┬─────────┘    │      │  │
│  │  └───────────┼──────────────┘      │  │
│  │  ┌───────────┴──────────────┐      │  │
│  │  │    Device Manager          │      │  │  GPU/CPU 调度
│  │  └────────────────────────────┘      │  │
│  └────────────────────────────────────┘  │
├──────────────────────────────────────────┤
│            Runtime Layer                  │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │ Memory   │ │  Cache   │ │  Logger  │ │
│  │ Monitor  │ │ Manager  │ │          │ │
│  └──────────┘ └──────────┘ └──────────┘ │
└──────────────────────────────────────────┘
```

## 4. 核心模块设计

### 4.1 模型管理器

```python
from enum import Enum
from typing import Any, Dict, Optional
import torch

class ModelType(Enum):
    YOLO = "yolo"
    RT_DETR = "rt-detr"
    SAM = "sam"
    OCR = "ocr"
    TRANSFORMER = "transformer"
    LLM = "llm"
    LSTM = "lstm"

class ModelManager:
    """模型注册、加载、缓存管理"""
    
    def __init__(self):
        self._registry: Dict[str, Any] = {}     # 已注册模型
        self._cache: Dict[str, Any] = {}         # 已加载模型缓存
        self._device = self._detect_device()
    
    def register(self, name: str, model_type: ModelType, path: str):
        """注册模型"""
        self._registry[name] = {
            "type": model_type,
            "path": path,
            "loaded": False,
        }
    
    def load(self, name: str) -> Any:
        """加载模型（带缓存）"""
        if name in self._cache:
            return self._cache[name]
        
        info = self._registry[name]
        model = self._load_model(info)
        model.to(self._device)
        self._cache[name] = model
        return model
    
    def unload(self, name: str):
        """卸载模型释放显存"""
        if name in self._cache:
            del self._cache[name]
            torch.cuda.empty_cache()
    
    def _detect_device(self) -> torch.device:
        if torch.cuda.is_available():
            return torch.device("cuda")
        return torch.device("cpu")
```

### 4.2 视觉推理模块

```python
class VisionHandler:
    """视觉模型推理处理器"""
    
    def __init__(self, model_manager: ModelManager):
        self.mm = model_manager
    
    def detect(self, model_name: str, image: bytes, 
               conf: float = 0.5) -> dict:
        """目标检测"""
        model = self.mm.load(model_name)
        results = model(image, conf=conf)
        return {
            "boxes": results.boxes.xyxy.tolist(),
            "scores": results.boxes.conf.tolist(),
            "classes": results.boxes.cls.tolist(),
        }
    
    def segment(self, model_name: str, image: bytes) -> dict:
        """图像分割"""
        model = self.mm.load(model_name)
        results = model(image)
        return {"masks": results.masks.data.tolist()}
    
    def ocr(self, model_name: str, image: bytes) -> dict:
        """文字识别"""
        model = self.mm.load(model_name)
        result = model(image)
        return {"text": result.text, "boxes": result.boxes}
```

### 4.3 gRPC 服务定义

```python
# engine.proto
# service EngineService {
#   rpc Infer(InferRequest) returns (InferResponse);
#   rpc InferStream(InferRequest) returns (stream InferChunk);
#   rpc LoadModel(LoadModelRequest) returns (LoadModelResponse);
#   rpc UnloadModel(UnloadModelRequest) returns (UnloadModelResponse);
#   rpc GetStatus(StatusRequest) returns (StatusResponse);
# }

class EngineServicer(engine_pb2_grpc.EngineServiceServicer):
    
    def Infer(self, request, context):
        """统一推理入口"""
        handler = self.router.get_handler(request.model_type)
        result = handler.execute(
            model_name=request.model_name,
            input_data=request.input_data,
            params=json.loads(request.params),
        )
        return engine_pb2.InferResponse(
            output_data=json.dumps(result),
            metadata="",
        )
    
    def InferStream(self, request, context):
        """流式推理（LLM 等）"""
        handler = self.router.get_handler(request.model_type)
        for chunk in handler.execute_stream(request):
            yield engine_pb2.InferChunk(data=chunk)
```

### 4.4 设备管理

```python
class DeviceManager:
    """GPU/CPU 设备管理"""
    
    def get_status(self) -> dict:
        status = {
            "device": "cuda" if torch.cuda.is_available() else "cpu",
            "gpu_count": torch.cuda.device_count(),
        }
        if torch.cuda.is_available():
            for i in range(torch.cuda.device_count()):
                status[f"gpu_{i}"] = {
                    "name": torch.cuda.get_device_name(i),
                    "memory_total": torch.cuda.get_device_properties(i).total_memory,
                    "memory_used": torch.cuda.memory_allocated(i),
                    "memory_cached": torch.cuda.memory_reserved(i),
                }
        return status
```

## 5. 模型加载策略

| 策略 | 说明 | 适用场景 |
|------|------|----------|
| 预加载 | 启动时加载常用模型 | 高频使用的小模型 |
| 懒加载 | 首次推理时加载 | 大模型、低频使用 |
| LRU 缓存 | 超过显存限制时淘汰最久未用 | 显存有限场景 |
| 共享加载 | 多工作流共享同一模型实例 | 并发推理场景 |
