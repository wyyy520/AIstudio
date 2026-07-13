# 插件系统架构

## 1. 设计目标

- **可扩展**：第三方开发者可轻松编写并注册新插件
- **类型安全**：强类型接口定义输入/输出
- **隔离性**：插件崩溃不影响主进程
- **热加载**：运行时动态加载/卸载插件
- **跨语言**：支持 Go / Python / TypeScript 插件

## 2. 插件分类

```
Plugins/
├── Vision/          # 视觉类插件
│   ├── YOLO/        # 目标检测
│   ├── RT-DETR/     # 实时目标检测
│   ├── SAM/         # 图像分割
│   └── OCR/         # 文字识别
├── NLP/             # 自然语言处理
│   ├── Transformer/ # 文本分类/NER
│   └── LLM/         # 大语言模型
├── TimeSeries/      # 时序分析
│   └── LSTM/        # 时序预测
├── Logic/           # 逻辑控制
│   ├── If/          # 条件判断
│   ├── Else/        # 条件分支
│   ├── Switch/      # 多路分支
│   ├── Loop/        # 循环
│   └── Retry/       # 重试
├── System/          # 系统操作
│   ├── Python/      # 执行 Python 脚本
│   ├── Git/         # Git 操作
│   ├── CUDA/        # GPU 状态
│   ├── Download/    # 文件下载
│   ├── Terminal/    # 终端命令
│   └── Environment/ # 环境变量管理
├── Simulation/      # 仿真
│   ├── SUMO/        # 交通仿真
│   ├── MATLAB/      # MATLAB 仿真
│   └── VISSIM/      # 交通微观仿真
├── MCP/             # MCP 协议插件
└── SDK/             # 插件开发 SDK
```

## 3. 插件接口规范

### 3.1 Go 插件接口

```go
// 所有插件必须实现此接口
type Plugin interface {
    // 元信息
    Meta() PluginMeta
    // 输入/输出端口定义
    Ports() PluginPorts
    // 执行
    Execute(ctx context.Context, input *PluginInput) (*PluginOutput, error)
    // 配置 schema（可选）
    ConfigSchema() *ConfigSchema
}

type PluginMeta struct {
    Name        string
    Version     string
    Author      string
    Description string
    Type        PluginType
    Icon        string
}

type PluginPorts struct {
    Inputs  []Port
    Outputs []Port
}

type Port struct {
    Name        string
    Type        PortType  // image / text / number / json / file
    Required    bool
    Description string
}
```

### 3.2 Python 插件接口

```python
from abc import ABC, abstractmethod

class Plugin(ABC):
    """所有 Python 插件的基类"""
    
    @abstractmethod
    def meta(self) -> dict:
        """返回插件元信息"""
        pass
    
    @abstractmethod
    def ports(self) -> dict:
        """返回输入/输出端口定义"""
        pass
    
    @abstractmethod
    def execute(self, input_data: dict, config: dict) -> dict:
        """执行插件逻辑"""
        pass
```

### 3.3 插件清单文件 (plugin.json)

```json
{
  "name": "yolo-detector",
  "version": "1.0.0",
  "author": "AI Studio",
  "description": "YOLO 目标检测插件",
  "type": "vision",
  "icon": "yolo.png",
  "language": "python",
  "entry": "main.py",
  "ports": {
    "inputs": [
      {"name": "image", "type": "image", "required": true, "description": "输入图像"},
      {"name": "confidence", "type": "number", "required": false, "default": 0.5}
    ],
    "outputs": [
      {"name": "detections", "type": "json", "description": "检测结果"}
    ]
  },
  "config_schema": {
    "model": {"type": "string", "default": "yolov8n.pt"},
    "device": {"type": "string", "default": "auto", "options": ["auto", "cpu", "cuda"]}
  }
}
```

## 4. 插件生命周期

```
加载阶段                    运行阶段                    卸载阶段
┌─────────┐              ┌─────────┐              ┌─────────┐
│  发现    │              │  执行    │              │  清理    │
│ (扫描目录) │              │ (调用    │              │ (释放    │
└────┬────┘              │  Execute) │              │  资源)   │
     │                   └────┬────┘              └─────────┘
┌────┴────┐                   │
│  注册    │              ┌────┴────┐
│ (写入    │              │  状态    │
│  registry)│              │  管理    │
└────┬────┘              └─────────┘
     │
┌────┴────┐
│  初始化  │
│ (加载    │
│  依赖)   │
└─────────┘
```

## 5. 插件加载机制

### 5.1 Python 插件（通过 Engine）

```go
// Backend 调用 Engine 执行 Python 插件
func (pm *PluginManager) executePythonPlugin(ctx context.Context, p *PythonPlugin, input *PluginInput) (*PluginOutput, error) {
    req := &engine.InferRequest{
        ModelType:  "plugin",
        ModelName:  p.Name(),
        InputData:  input.Serialize(),
        Params:     input.Config,
    }
    resp, err := pm.engineClient.Infer(ctx, req)
    if err != nil {
        return nil, err
    }
    return DeserializeOutput(resp.OutputData), nil
}
```

### 5.2 Go 插件（原生加载）

```go
// 使用 Go plugin 包加载 .so 文件
func (pm *PluginManager) loadGoPlugin(path string) error {
    p, err := plugin.Open(path)
    if err != nil {
        return err
    }
    sym, err := p.Lookup("Plugin")
    if err != nil {
        return err
    }
    pluginInstance, ok := sym.(Plugin)
    if !ok {
        return errors.New("invalid plugin interface")
    }
    pm.registry[pluginInstance.Meta().Name] = pluginInstance
    return nil
}
```

### 5.3 进程隔离插件

对于不信任的第三方插件，使用独立进程执行：

```go
type ProcessPlugin struct {
    cmd    *exec.Cmd
    stdin  io.WriteCloser
    stdout io.ReadCloser
}

// 通过 stdin/stdout JSON 通信
func (p *ProcessPlugin) Execute(ctx context.Context, input *PluginInput) (*PluginOutput, error) {
    json.NewEncoder(p.stdin).Encode(input)
    var output PluginOutput
    json.NewDecoder(p.stdout).Decode(&output)
    return &output, nil
}
```

## 6. 插件通信数据格式

```go
type PluginInput struct {
    NodeID  string                 `json:"node_id"`
    Data    map[string]interface{} `json:"data"`     // 端口数据
    Config  map[string]interface{} `json:"config"`   // 节点配置
    Context *ExecutionContext      `json:"context"`   // 执行上下文
}

type PluginOutput struct {
    Data    map[string]interface{} `json:"data"`     // 输出端口数据
    Status  string                 `json:"status"`    // success / error
    Error   string                 `json:"error"`     // 错误信息
    Metrics *ExecutionMetrics      `json:"metrics"`   // 执行指标
}

type ExecutionContext struct {
    WorkflowID string
    TaskID     string
    ProjectID  string
    WorkDir    string
    Runtime    map[string]interface{} // 运行时共享数据
}
```
