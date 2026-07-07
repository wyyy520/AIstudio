# AI Studio Engine

Python-based AI execution engine with gRPC service interface.

## Architecture

```
Backend (Go)
    │
    ▼ (gRPC)
Engine (Python)
    │
    ▼
Task Router
    │
    ├─► Vision Handler
    ├─► NLP Handler
    └─► Timeseries Handler
            │
            ▼
        Model Manager
```

## Directory Structure

```
Engine/
├── grpc/
│   └── server.py              # gRPC server implementation
├── models/                    # Model storage / configs
├── handlers/
│   ├── vision/handler.py      # Image tasks (detect, classify, OCR)
│   ├── nlp/handler.py         # Text tasks (generate, chat, embedding)
│   └── timeseries/handler.py  # Time series tasks (forecast, anomaly)
├── manager/
│   └── model_manager.py       # Model lifecycle management
├── proto/
│   ├── aiengine.proto         # gRPC service definition
│   ├── aiengine_pb2.py        # Generated protobuf code
│   └── aiengine_pb2_grpc.py   # Generated gRPC stubs
├── task_router.py             # Task routing to handlers
├── main.py                    # Entry point
└── requirements.txt
```

## Quick Start

```bash
# Install dependencies
pip install -r requirements.txt

# Start the engine
python main.py --port 50051
```

## gRPC API

The engine exposes these RPC methods:

| Method | Description |
|--------|-------------|
| `ExecuteTask` | Execute an AI task |
| `LoadModel` | Load a model into memory |
| `UnloadModel` | Unload a model |
| `GetModelStatus` | Check model status |
| `HealthCheck` | Health probe |

### ExecuteTask

```protobuf
message TaskRequest {
  string task_id = 1;
  string task_type = 2;
  string input = 3;
  map<string, string> config = 4;
}

message TaskResponse {
  string task_id = 1;
  string status = 2;
  string result = 3;
  map<string, string> metadata = 4;
}
```

## Supported Task Types

| Category | Task Type | Description |
|----------|-----------|-------------|
| Vision | `vision.detect` | Object detection |
| Vision | `vision.classify` | Image classification |
| Vision | `vision.ocr` | Text extraction |
| NLP | `nlp.generate` | Text generation |
| NLP | `nlp.chat` | Chat completion |
| NLP | `nlp.embedding` | Text embedding |
| Timeseries | `timeseries.forecast` | Time series forecasting |
| Timeseries | `timeseries.anomaly` | Anomaly detection |
| Timeseries | `timeseries.trend` | Trend analysis |

## Model Manager

Supports loading/unloading models by type:

```python
from manager.model_manager import ModelManager, ModelType

manager = ModelManager()
manager.register_model("my-model", ModelType.LLM, path="/path/to/model")
manager.load_model("my-model")
model = manager.get_model("my-model")
```

## Regenerating Proto

```bash
cd Engine/proto
python -m grpc_tools.protoc -I . --python_out=. --grpc_python_out=. aiengine.proto
```
