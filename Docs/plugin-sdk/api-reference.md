# 插件 SDK API 参考

## 1. 核心接口

### AIStudioPlugin (Python 基类)

```python
class AIStudioPlugin(ABC):
    """所有 Python 插件的基类"""
    
    @abstractmethod
    def execute(self, inputs: Dict[str, Any], config: Dict[str, Any]) -> Dict[str, Any]:
        """
        执行插件逻辑
        
        Args:
            inputs (Dict): 输入端口数据，key 为端口名
                        支持类型：str (image/text/file), dict (json), float (number)
            config (Dict): 节点配置参数（来自 plugin.json 的 config_schema）
        
        Returns:
            Dict: 输出端口数据，key 为端口名
                  返回格式：{"端口名": 数据}
        
        Raises:
            Exception: 执行失败时抛出异常，错误信息会显示在日志面板
        """
        pass
    
    def setup(self, config: Dict[str, Any]):
        """
        插件初始化（可选）
        在工作流启动时调用一次，适合加载模型、初始化连接等
        """
        pass
    
    def teardown(self):
        """
        清理资源（可选）
        在工作流停止时调用，适合释放显存、关闭连接等
        """
        pass
    
    def execute_stream(self, inputs: Dict[str, Any], config: Dict[str, Any]) -> Generator:
        """
        流式执行（可选，LLM 等插件使用）
        返回一个生成器，逐步 yield 输出块
        """
        yield from []
```

## 2. 数据类型转换

| plugin.json 端口类型 | Python 接收类型 | Python 返回类型 |
|---------------------|------------------|------------------|
| `image` | `str`（文件路径） | `str`（文件路径） |
| `text` | `str` | `str` |
| `number` | `float` / `int` | `float` / `int` |
| `json` | `dict` / `list` | `dict` / `list` |
| `file` | `str`（文件路径） | `str`（文件路径） |
| `tensor` | `numpy.ndarray` | `numpy.ndarray` |
| `stream` | - | `Generator`（流式输出） |

## 3. 插件上下文

引擎在执行插件时，会注入以下上下文变量：

```python
# 通过 inputs 传递的上下文
inputs = {
    # 用户定义的输入端口数据
    "image": "/storage/datasets/img001.jpg",
    "confidence": 0.5,
    
    # 引擎注入的上下文（以 _ 开头）
    "_context": {
        "workflow_id": "wf_001",
        "task_id": "task_20260707_001",
        "project_id": "proj_001",
        "work_dir": "/runtime/workspace/task_20260707_001/",
        "runtime": {}  # 节点间共享数据
    }
}
```

## 4. 工具函数

### 4.1 路径处理

```python
import os

def get_project_path(project_id: str) -> str:
    """获取项目根目录"""
    return os.path.join("/storage/projects", project_id)

def get_runtime_path(task_id: str) -> str:
    """获取任务运行时目录"""
    return os.path.join("/runtime/workspace", task_id)

def get_storage_path(subdir: str) -> str:
    """获取存储目录"""
    return os.path.join("/storage", subdir)
```

### 4.2 图像工具

```python
import cv2
import base64
import numpy as np

def image_to_base64(image_path: str) -> str:
    """图像文件转 base64（用于小图预览）"""
    with open(image_path, "rb") as f:
        return base64.b64encode(f.read()).decode()

def base64_to_image(b64_str: str, output_path: str):
    """base64 转图像文件"""
    data = base64.b64decode(b64_str)
    with open(output_path, "wb") as f:
        f.write(data)

def cv2_to_base64(image: np.ndarray) -> str:
    """OpenCV 图像转 base64"""
    _, buffer = cv2.imencode(".jpg", image)
    return base64.b64encode(buffer).decode()
```

### 4.3 日志输出

```python
# 在 execute 中使用 print() 输出日志
# 日志会自动推送到前端日志面板

def execute(self, inputs, config):
    print(f"[INFO] 开始处理图像: {inputs['image']}")
    # ... 处理逻辑 ...
    print(f"[INFO] 检测完成，共 {len(boxes)} 个目标")
    print(f"[DEBUG] 推理耗时: {duration_ms}ms")
    # 支持级别前缀：[INFO] [WARN] [ERROR] [DEBUG]
```

## 5. 错误码

| 错误类型 | 处理方式 | 说明 |
|----------|----------|------|
| 参数错误 | 抛出 ValueError | 输入参数不符合要求 |
| 模型错误 | 抛出 RuntimeError | 模型加载/推理失败 |
| 文件错误 | 抛出 FileNotFoundError | 输入文件不存在 |
| 超时错误 | 抛出 TimeoutError | 执行超时 |
| 其他错误 | 抛出 Exception | 通用错误 |

错误会被引擎捕获并推送到前端：

```json
{"type": "node_error", "node_id": "n1", "message": "Model file not found: yolov8n.pt"}
```

## 6. 插件元数据参考

```json
{
  "name": "插件标识符（小写，下划线分隔）",
  "version": "语义化版本号",
  "author": "作者名称",
  "description": "插件功能描述（支持中文）",
  "type": "vision | nlp | logic | system | simulation | mcp",
  "icon": "图标文件路径（相对于插件目录）",
  "language": "python | go",
  "entry": "入口文件名",
  "homepage": "插件主页 URL（可选）",
  "repository": "源码仓库 URL（可选）",
  "license": "MIT | Apache-2.0 | GPL-3.0（可选）",
  "ports": { ... },
  "config_schema": { ... }
}
```
