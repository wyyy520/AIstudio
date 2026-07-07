package workflow

import (
	"fmt"
	"sync"
)

type NodeFactory func() ExecutableNode

type NodeDefinition struct {
	Type        string
	Plugin      string
	Name        string
	Description string
	Inputs      []Port
	Outputs     []Port
	Factory     NodeFactory
}

type NodeRegistry struct {
	mu      sync.RWMutex
	nodes   map[string]NodeDefinition
}

func NewNodeRegistry() *NodeRegistry {
	return &NodeRegistry{
		nodes: make(map[string]NodeDefinition),
	}
}

func key(nodeType, plugin string) string {
	return nodeType + ":" + plugin
}

func (r *NodeRegistry) Register(def NodeDefinition) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodes[key(def.Type, def.Plugin)] = def
}

func (r *NodeRegistry) Get(nodeType, plugin string) (NodeDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	def, ok := r.nodes[key(nodeType, plugin)]
	return def, ok
}

func (r *NodeRegistry) Has(nodeType, plugin string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.nodes[key(nodeType, plugin)]
	return ok
}

func (r *NodeRegistry) List() []NodeDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]NodeDefinition, 0, len(r.nodes))
	for _, def := range r.nodes {
		result = append(result, def)
	}
	return result
}

func (r *NodeRegistry) CreateExecutable(nodeType, plugin string) (ExecutableNode, error) {
	r.mu.RLock()
	def, ok := r.nodes[key(nodeType, plugin)]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("node type %q with plugin %q is not registered", nodeType, plugin)
	}
	return def.Factory(), nil
}

var DefaultRegistry = NewNodeRegistry()

func RegisterDefaultNodes() {
	DefaultRegistry.Register(NodeDefinition{
		Type:        NodeTypeInput,
		Plugin:      "data-source",
		Name:        "数据源",
		Description: "提供数据输入",
		Inputs:      []Port{},
		Outputs: []Port{
			{ID: "out_image", Name: "image", Type: "image", Required: true},
			{ID: "out_text", Name: "text", Type: "text", Required: false},
			{ID: "out_tensor", Name: "data", Type: "tensor", Required: false},
		},
		Factory: func() ExecutableNode { return &DataSourceNode{} },
	})

	DefaultRegistry.Register(NodeDefinition{
		Type:        NodeTypeVision,
		Plugin:      "yolo-detector",
		Name:        "YOLO 目标检测",
		Description: "使用 YOLOv8 检测图像中的目标",
		Inputs: []Port{
			{ID: "in_image", Name: "image", Type: "image", Required: true},
		},
		Outputs: []Port{
			{ID: "out_detections", Name: "detections", Type: "json", Description: "检测结果 {boxes, scores, classes}"},
		},
		Factory: func() ExecutableNode { return &YOLODetectorNode{} },
	})

	DefaultRegistry.Register(NodeDefinition{
		Type:        NodeTypeSystem,
		Plugin:      "pytorch-train",
		Name:        "PyTorch 训练",
		Description: "使用 PyTorch 训练模型",
		Inputs: []Port{
			{ID: "in_dataset", Name: "dataset", Type: "dataset", Required: true},
		},
		Outputs: []Port{
			{ID: "out_model", Name: "model", Type: "model", Description: "训练好的模型"},
		},
		Factory: func() ExecutableNode { return &PyTorchNode{} },
	})

	DefaultRegistry.Register(NodeDefinition{
		Type:        NodeTypeNLP,
		Plugin:      "transformer",
		Name:        "Transformer 分类",
		Description: "使用 Transformer 进行文本分类",
		Inputs: []Port{
			{ID: "in_text", Name: "text", Type: "text", Required: true},
		},
		Outputs: []Port{
			{ID: "out_result", Name: "result", Type: "json", Description: "分类结果"},
		},
		Factory: func() ExecutableNode { return &TransformerNode{} },
	})

	DefaultRegistry.Register(NodeDefinition{
		Type:        NodeTypeTimeseries,
		Plugin:      "lstm-predict",
		Name:        "LSTM 时序预测",
		Description: "使用 LSTM 模型进行时序预测",
		Inputs: []Port{
			{ID: "in_data", Name: "data", Type: "tensor", Required: true},
		},
		Outputs: []Port{
			{ID: "out_prediction", Name: "prediction", Type: "json", Description: "预测结果"},
		},
		Factory: func() ExecutableNode { return &LSTMNode{} },
	})

	DefaultRegistry.Register(NodeDefinition{
		Type:        NodeTypeSystem,
		Plugin:      "model-export",
		Name:        "模型导出",
		Description: "将训练好的模型导出为指定格式",
		Inputs: []Port{
			{ID: "in_model", Name: "model", Type: "model", Required: true},
		},
		Outputs: []Port{
			{ID: "out_path", Name: "path", Type: "file", Description: "导出文件路径"},
		},
		Factory: func() ExecutableNode { return &ExportNode{} },
	})

	DefaultRegistry.Register(NodeDefinition{
		Type:        NodeTypeMCP,
		Plugin:      "mcp-client",
		Name:        "MCP 客户端",
		Description: "通过 MCP 协议调用外部工具",
		Inputs: []Port{
			{ID: "in_params", Name: "params", Type: "json", Required: true},
		},
		Outputs: []Port{
			{ID: "out_result", Name: "result", Type: "json", Description: "MCP 调用结果"},
		},
		Factory: func() ExecutableNode { return &MCPNode{} },
	})
}
