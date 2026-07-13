package workflow

import (
	"context"
	"log"

	"github.com/aistudio/backend/internal/engine"
	"github.com/aistudio/backend/internal/workflow/executors"
)

var (
	engineClient engine.EngineClient
)

func SetEngineClient(client engine.EngineClient) {
	engineClient = client
}

// ============================================================================
// Built-in Node Executors
// ============================================================================

// executeControlCondition is a placeholder executor for control.condition nodes.
// The actual logic is handled by the engine's topology controller.
func executeControlCondition(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"passed":  true,
		"failed":  false,
		"message": "condition evaluated",
	}, nil
}

// executeControlLoop is a placeholder executor for control.loop nodes.
func executeControlLoop(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"iterations": 0,
		"completed":  true,
		"message":    "loop completed",
	}, nil
}

// executeControlSwitch is a placeholder executor for control.switch nodes.
func executeControlSwitch(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"case":    "default",
		"message": "switch evaluated",
	}, nil
}

// executeControlRetry is a placeholder executor for control.retry nodes.
func executeControlRetry(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"retries": 0,
		"success": true,
		"message": "retry logic completed",
	}, nil
}

// noOpExecutor is a generic no-operation node executor.
// It simply passes inputs through and returns a status message.
func noOpExecutor(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{}, len(inputs)+2)
	for k, v := range inputs {
		result[k] = v
	}
	result["status"] = "completed"
	result["message"] = "node executed (declarative placeholder)"
	return result, nil
}

// ============================================================================
// Built-in Node Registry — DSL-Native Control Nodes
// ============================================================================

// BuiltInNodeDefinitions returns the full list of built-in node definitions.
// These are the nodes that ship with the AIStudio engine.
func BuiltInNodeDefinitions() []NodeDefinition {
	return []NodeDefinition{
		// ---- Logic / Control nodes ----
		{
			Type:        "control.condition",
			Plugin:      "core",
			Name:        "IF 条件",
			Description: "条件判断分支",
			Category:    "logic",
			Inputs: []Port{
				{ID: "input", Name: "输入", Type: "any", Required: false},
			},
			Outputs: []Port{
				{ID: "true", Name: "True", Type: "any"},
				{ID: "false", Name: "False", Type: "any"},
			},
			Factory: func() ExecutableNode { return executableFunc(executors.ConditionExecutor()) },
		},
		{
			Type:        "control.switch",
			Plugin:      "core",
			Name:        "Switch",
			Description: "多路分支选择",
			Category:    "logic",
			Inputs: []Port{
				{ID: "input", Name: "输入", Type: "any", Required: false},
			},
			Outputs: []Port{
				{ID: "case_0", Name: "Case 0", Type: "any"},
				{ID: "case_1", Name: "Case 1", Type: "any"},
				{ID: "case_2", Name: "Case 2", Type: "any"},
				{ID: "default", Name: "Default", Type: "any"},
			},
			Factory: func() ExecutableNode { return executableFunc(executors.SwitchExecutor()) },
		},
		{
			Type:        "control.loop",
			Plugin:      "core",
			Name:        "Loop",
			Description: "循环迭代",
			Category:    "logic",
			Inputs: []Port{
				{ID: "input", Name: "输入", Type: "any", Required: false},
			},
			Outputs: []Port{
				{ID: "iteration", Name: "迭代输出", Type: "any"},
				{ID: "completed", Name: "完成", Type: "any"},
			},
			Factory: func() ExecutableNode { return executableFunc(executors.LoopExecutor()) },
		},
		{
			Type:        "control.retry",
			Plugin:      "core",
			Name:        "Retry",
			Description: "失败重试",
			Category:    "logic",
			Inputs: []Port{
				{ID: "input", Name: "输入", Type: "any", Required: false},
			},
			Outputs: []Port{
				{ID: "output", Name: "输出", Type: "any"},
				{ID: "failed", Name: "最终失败", Type: "any"},
			},
			Factory: func() ExecutableNode { return executableFunc(executors.RetryExecutor(nil)) },
		},

		// ---- Data processing nodes ----
		{
			Type:        "data.loader",
			Plugin:      "core",
			Name:        "数据集加载",
			Description: "加载数据集",
			Category:    "data",
			Inputs:      []Port{},
			Outputs: []Port{
				{ID: "dataset", Name: "数据集", Type: "dataset"},
				{ID: "info", Name: "数据集信息", Type: "json"},
			},
			Factory: func() ExecutableNode { return executableFunc(executors.DataLoaderExecutor()) },
		},
		{
			Type:        "data.split",
			Plugin:      "core",
			Name:        "数据划分",
			Description: "划分训练/验证/测试集",
			Category:    "data",
			Inputs: []Port{
				{ID: "data", Name: "数据集", Type: "dataset", Required: true},
			},
			Outputs: []Port{
				{ID: "train", Name: "训练集", Type: "dataset"},
				{ID: "val", Name: "验证集", Type: "dataset"},
				{ID: "test", Name: "测试集", Type: "dataset"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
		{
			Type:        "data.augmentation",
			Plugin:      "core",
			Name:        "数据增强",
			Description: "数据增强处理",
			Category:    "data",
			Inputs: []Port{
				{ID: "data", Name: "原始数据", Type: "dataset", Required: true},
			},
			Outputs: []Port{
				{ID: "augmented", Name: "增强后数据", Type: "dataset"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},

		// ---- Vision / Model nodes ----
		{
			Type:        "vision.yolo_train",
			Plugin:      "core",
			Name:        "YOLO 训练",
			Description: "YOLO 目标检测模型训练",
			Category:    "ai-vision",
			Inputs: []Port{
				{ID: "dataset", Name: "数据集", Type: "dataset", Required: true},
				{ID: "model_weight", Name: "预训练权重", Type: "model", Required: false},
			},
			Outputs: []Port{
				{ID: "best_pt", Name: "best.pt", Type: "model"},
				{ID: "training_results", Name: "训练结果", Type: DataTypeJSON},
			},
			Factory: func() ExecutableNode { return executableFunc(executors.YOLOTrainExecutor(engineClient)) },
		},
		{
			Type:        "vision.yolo_inference",
			Plugin:      "core",
			Name:        "YOLO 推理",
			Description: "YOLO 模型推理",
			Category:    "ai-vision",
			Inputs: []Port{
				{ID: "image", Name: "图像", Type: "image", Required: true},
				{ID: "model", Name: "模型", Type: "model", Required: true},
			},
			Outputs: []Port{
				{ID: "detections", Name: "检测结果", Type: DataTypeJSON},
				{ID: "annotated", Name: "标注图像", Type: "image"},
			},
			Factory: func() ExecutableNode { return executableFunc(executors.YOLOPredictExecutor(engineClient)) },
		},

		// ---- NLP nodes ----
		{
			Type:        "nlp.transformer",
			Plugin:      "core",
			Name:        "Transformer",
			Description: "通用 Transformer 模型",
			Category:    "ai-nlp",
			Inputs: []Port{
				{ID: "text", Name: "文本", Type: "text", Required: true},
			},
			Outputs: []Port{
				{ID: "output", Name: "输出", Type: "tensor"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
		{
			Type:        "nlp.llm",
			Plugin:      "core",
			Name:        "LLM 对话",
			Description: "大语言模型推理",
			Category:    "ai-nlp",
			Inputs: []Port{
				{ID: "prompt", Name: "提示词", Type: "text", Required: true},
				{ID: "context", Name: "上下文", Type: "text", Required: false},
			},
			Outputs: []Port{
				{ID: "response", Name: "回复", Type: "text"},
				{ID: "usage", Name: "用量", Type: DataTypeJSON},
			},
			Factory: func() ExecutableNode { return executableFunc(executors.NLPExecutor(engineClient)) },
		},

		// ---- Training nodes ----
		{
			Type:        "training.train",
			Plugin:      "core",
			Name:        "训练",
			Description: "模型训练",
			Category:    "training",
			Inputs: []Port{
				{ID: "model", Name: "模型", Type: "model", Required: true},
				{ID: "dataset", Name: "训练数据", Type: "dataset", Required: true},
			},
			Outputs: []Port{
				{ID: "trained_model", Name: "训练后模型", Type: "model"},
				{ID: "metrics", Name: "训练指标", Type: DataTypeJSON},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
		{
			Type:        "training.export",
			Plugin:      "core",
			Name:        "导出模型",
			Description: "导出模型格式",
			Category:    "training",
			Inputs: []Port{
				{ID: "model", Name: "模型", Type: "model", Required: true},
			},
			Outputs: []Port{
				{ID: "exported", Name: "导出模型", Type: "model"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},

		// ---- System nodes ----
		{
			Type:        "system.python_env",
			Plugin:      "core",
			Name:        "Python 环境",
			Description: "配置 Python 运行环境",
			Category:    "system",
			Inputs:      []Port{},
			Outputs: []Port{
				{ID: "env", Name: "环境", Type: DataTypeStream},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
		{
			Type:        "system.install_dep",
			Plugin:      "core",
			Name:        "安装依赖",
			Description: "安装 Python 依赖",
			Category:    "system",
			Inputs:      []Port{},
			Outputs: []Port{
				{ID: "result", Name: "安装结果", Type: DataTypeJSON},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},

		// ---- Deployment nodes ----
		{
			Type:        "deployment.api_server",
			Plugin:      "core",
			Name:        "API Server",
			Description: "部署为 API 服务",
			Category:    "deployment",
			Inputs: []Port{
				{ID: "model", Name: "模型", Type: DataTypeModel, Required: true},
			},
			Outputs: []Port{
				{ID: "endpoint", Name: "API 端点", Type: DataTypeStream},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
		{
			Type:        "deployment.docker",
			Plugin:      "core",
			Name:        "Docker",
			Description: "Docker 容器化部署",
			Category:    "deployment",
			Inputs: []Port{
				{ID: "model", Name: "模型", Type: DataTypeModel, Required: true},
				{ID: "config", Name: "配置", Type: DataTypeJSON, Required: false},
			},
			Outputs: []Port{
				{ID: "image", Name: "Docker 镜像", Type: DataTypeStream},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},

		// ---- Image Classification nodes ----
		{
			Type:        "vision.resnet",
			Plugin:      "core",
			Name:        "ResNet 分类",
			Description: "ResNet 图像分类模型",
			Category:    "ai-vision",
			Inputs: []Port{
				{ID: "image", Name: "图像", Type: "image", Required: true},
			},
			Outputs: []Port{
				{ID: "class", Name: "类别", Type: "text"},
				{ID: "confidence", Name: "置信度", Type: "number"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
		{
			Type:        "vision.efficientnet",
			Plugin:      "core",
			Name:        "EfficientNet 分类",
			Description: "EfficientNet 高效图像分类",
			Category:    "ai-vision",
			Inputs: []Port{
				{ID: "image", Name: "图像", Type: "image", Required: true},
			},
			Outputs: []Port{
				{ID: "class", Name: "类别", Type: "text"},
				{ID: "probabilities", Name: "概率分布", Type: DataTypeJSON},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
		{
			Type:        "vision.vgg",
			Plugin:      "core",
			Name:        "VGG 分类",
			Description: "VGG 图像分类模型",
			Category:    "ai-vision",
			Inputs: []Port{
				{ID: "image", Name: "图像", Type: "image", Required: true},
			},
			Outputs: []Port{
				{ID: "class", Name: "类别", Type: "text"},
				{ID: "features", Name: "特征向量", Type: "tensor"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},

		// ---- Object Detection nodes ----
		{
			Type:        "vision.faster_rcnn",
			Plugin:      "core",
			Name:        "Faster R-CNN",
			Description: "Faster R-CNN 目标检测",
			Category:    "ai-vision",
			Inputs: []Port{
				{ID: "image", Name: "图像", Type: "image", Required: true},
			},
			Outputs: []Port{
				{ID: "boxes", Name: "边界框", Type: DataTypeJSON},
				{ID: "labels", Name: "标签", Type: DataTypeJSON},
				{ID: "scores", Name: "分数", Type: DataTypeJSON},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
		{
			Type:        "vision.ssd",
			Plugin:      "core",
			Name:        "SSD 检测",
			Description: "SSD 单阶段目标检测",
			Category:    "ai-vision",
			Inputs: []Port{
				{ID: "image", Name: "图像", Type: "image", Required: true},
			},
			Outputs: []Port{
				{ID: "detections", Name: "检测结果", Type: DataTypeJSON},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},

		// ---- Image Segmentation nodes ----
		{
			Type:        "vision.unet",
			Plugin:      "core",
			Name:        "U-Net 分割",
			Description: "U-Net 语义分割",
			Category:    "ai-vision",
			Inputs: []Port{
				{ID: "image", Name: "图像", Type: "image", Required: true},
			},
			Outputs: []Port{
				{ID: "mask", Name: "分割掩码", Type: "image"},
				{ID: "overlay", Name: "叠加图", Type: "image"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
		{
			Type:        "vision.mask_rcnn",
			Plugin:      "core",
			Name:        "Mask R-CNN",
			Description: "Mask R-CNN 实例分割",
			Category:    "ai-vision",
			Inputs: []Port{
				{ID: "image", Name: "图像", Type: "image", Required: true},
			},
			Outputs: []Port{
				{ID: "instances", Name: "实例", Type: DataTypeJSON},
				{ID: "masks", Name: "掩码", Type: "image"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},

		// ---- NLP nodes ----
		{
			Type:        "nlp.bert",
			Plugin:      "core",
			Name:        "BERT",
			Description: "BERT 预训练模型",
			Category:    "ai-nlp",
			Inputs: []Port{
				{ID: "text", Name: "文本", Type: "text", Required: true},
			},
			Outputs: []Port{
				{ID: "embedding", Name: "嵌入向量", Type: "tensor"},
				{ID: "pooled", Name: "池化输出", Type: "tensor"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
		{
			Type:        "nlp.lstm",
			Plugin:      "core",
			Name:        "LSTM",
			Description: "LSTM 文本处理",
			Category:    "ai-nlp",
			Inputs: []Port{
				{ID: "text", Name: "文本", Type: "text", Required: true},
			},
			Outputs: []Port{
				{ID: "output", Name: "输出", Type: "tensor"},
				{ID: "hidden", Name: "隐藏状态", Type: "tensor"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
		{
			Type:        "nlp.text_classification",
			Plugin:      "core",
			Name:        "文本分类",
			Description: "文本情感/分类分析",
			Category:    "ai-nlp",
			Inputs: []Port{
				{ID: "text", Name: "文本", Type: "text", Required: true},
			},
			Outputs: []Port{
				{ID: "label", Name: "类别", Type: "text"},
				{ID: "score", Name: "分数", Type: "number"},
			},
			Factory: func() ExecutableNode { return executableFunc(executors.NLPExecutor(engineClient)) },
		},
		{
			Type:        "nlp.ner",
			Plugin:      "core",
			Name:        "命名实体识别",
			Description: "NER 实体识别",
			Category:    "ai-nlp",
			Inputs: []Port{
				{ID: "text", Name: "文本", Type: "text", Required: true},
			},
			Outputs: []Port{
				{ID: "entities", Name: "实体列表", Type: DataTypeJSON},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
		{
			Type:        "nlp.summarization",
			Plugin:      "core",
			Name:        "文本摘要",
			Description: "自动文本摘要生成",
			Category:    "ai-nlp",
			Inputs: []Port{
				{ID: "text", Name: "原文", Type: "text", Required: true},
			},
			Outputs: []Port{
				{ID: "summary", Name: "摘要", Type: "text"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
		{
			Type:        "nlp.translation",
			Plugin:      "core",
			Name:        "机器翻译",
			Description: "文本翻译",
			Category:    "ai-nlp",
			Inputs: []Port{
				{ID: "text", Name: "源文本", Type: "text", Required: true},
				{ID: "target_lang", Name: "目标语言", Type: "text", Required: true},
			},
			Outputs: []Port{
				{ID: "translated", Name: "翻译结果", Type: "text"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
		{
			Type:        "nlp.tokenizer",
			Plugin:      "core",
			Name:        "分词器",
			Description: "文本分词处理",
			Category:    "ai-nlp",
			Inputs: []Port{
				{ID: "text", Name: "文本", Type: "text", Required: true},
			},
			Outputs: []Port{
				{ID: "tokens", Name: "词元", Type: DataTypeJSON},
				{ID: "ids", Name: "词元ID", Type: DataTypeJSON},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},

		// ---- Speech nodes ----
		{
			Type:        "speech.asr",
			Plugin:      "core",
			Name:        "语音识别",
			Description: "语音转文字 (ASR)",
			Category:    "ai-speech",
			Inputs: []Port{
				{ID: "audio", Name: "音频", Type: "audio", Required: true},
			},
			Outputs: []Port{
				{ID: "text", Name: "文本", Type: "text"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
		{
			Type:        "speech.tts",
			Plugin:      "core",
			Name:        "语音合成",
			Description: "文字转语音 (TTS)",
			Category:    "ai-speech",
			Inputs: []Port{
				{ID: "text", Name: "文本", Type: "text", Required: true},
			},
			Outputs: []Port{
				{ID: "audio", Name: "音频", Type: "audio"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},

		// ---- Data Processing nodes ----
		{
			Type:        "data.csv_reader",
			Plugin:      "core",
			Name:        "CSV 读取",
			Description: "读取 CSV 数据文件",
			Category:    "data",
			Inputs: []Port{
				{ID: "file_path", Name: "文件路径", Type: "text", Required: true},
			},
			Outputs: []Port{
				{ID: "data", Name: "数据", Type: "dataset"},
				{ID: "columns", Name: "列名", Type: DataTypeJSON},
			},
			Factory: func() ExecutableNode { return executableFunc(executors.DataLoaderExecutor()) },
		},
		{
			Type:        "data.json_reader",
			Plugin:      "core",
			Name:        "JSON 读取",
			Description: "读取 JSON 数据文件",
			Category:    "data",
			Inputs: []Port{
				{ID: "file_path", Name: "文件路径", Type: "text", Required: true},
			},
			Outputs: []Port{
				{ID: "data", Name: "数据", Type: DataTypeJSON},
			},
			Factory: func() ExecutableNode { return executableFunc(executors.DataLoaderExecutor()) },
		},
		{
			Type:        "data.image_loader",
			Plugin:      "core",
			Name:        "图像加载",
			Description: "加载图像文件",
			Category:    "data",
			Inputs: []Port{
				{ID: "path", Name: "路径", Type: "text", Required: true},
			},
			Outputs: []Port{
				{ID: "image", Name: "图像", Type: "image"},
				{ID: "metadata", Name: "元数据", Type: DataTypeJSON},
			},
			Factory: func() ExecutableNode { return executableFunc(executors.DataLoaderExecutor()) },
		},

		// ---- Feature Engineering nodes ----
		{
			Type:        "feature.scaler",
			Plugin:      "core",
			Name:        "特征归一化",
			Description: "数据标准化/归一化",
			Category:    "feature",
			Inputs: []Port{
				{ID: "data", Name: "输入数据", Type: "dataset", Required: true},
			},
			Outputs: []Port{
				{ID: "scaled", Name: "归一化数据", Type: "dataset"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
		{
			Type:        "feature.encoder",
			Plugin:      "core",
			Name:        "特征编码",
			Description: "类别特征编码",
			Category:    "feature",
			Inputs: []Port{
				{ID: "data", Name: "输入数据", Type: "dataset", Required: true},
			},
			Outputs: []Port{
				{ID: "encoded", Name: "编码后数据", Type: "dataset"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},

		// ---- Evaluation nodes ----
		{
			Type:        "eval.classification",
			Plugin:      "core",
			Name:        "分类评估",
			Description: "分类模型评估指标",
			Category:    "evaluation",
			Inputs: []Port{
				{ID: "predictions", Name: "预测", Type: DataTypeJSON, Required: true},
				{ID: "labels", Name: "真实标签", Type: DataTypeJSON, Required: true},
			},
			Outputs: []Port{
				{ID: "accuracy", Name: "准确率", Type: "number"},
				{ID: "report", Name: "评估报告", Type: DataTypeJSON},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
		{
			Type:        "eval.detection",
			Plugin:      "core",
			Name:        "检测评估",
			Description: "目标检测评估指标",
			Category:    "evaluation",
			Inputs: []Port{
				{ID: "predictions", Name: "预测", Type: DataTypeJSON, Required: true},
				{ID: "ground_truth", Name: "真实框", Type: DataTypeJSON, Required: true},
			},
			Outputs: []Port{
				{ID: "map", Name: "mAP", Type: "number"},
				{ID: "metrics", Name: "评估指标", Type: DataTypeJSON},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},

		// ---- Visualization nodes ----
		{
			Type:        "viz.chart",
			Plugin:      "core",
			Name:        "图表",
			Description: "数据可视化图表",
			Category:    "visualization",
			Inputs: []Port{
				{ID: "data", Name: "数据", Type: "dataset", Required: true},
			},
			Outputs: []Port{
				{ID: "chart", Name: "图表", Type: "image"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
		{
			Type:        "viz.plot",
			Plugin:      "core",
			Name:        "绘图",
			Description: "自定义数据绘图",
			Category:    "visualization",
			Inputs: []Port{
				{ID: "x", Name: "X轴数据", Type: DataTypeJSON, Required: true},
				{ID: "y", Name: "Y轴数据", Type: DataTypeJSON, Required: true},
			},
			Outputs: []Port{
				{ID: "plot", Name: "图像", Type: "image"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},

		// ---- Model nodes ----
		{
			Type:        "model.load",
			Plugin:      "core",
			Name:        "加载模型",
			Description: "加载预训练模型",
			Category:    "model",
			Inputs: []Port{
				{ID: "path", Name: "模型路径", Type: "text", Required: true},
			},
			Outputs: []Port{
				{ID: "model", Name: "模型", Type: DataTypeModel},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
		{
			Type:        "model.save",
			Plugin:      "core",
			Name:        "保存模型",
			Description: "保存训练好的模型",
			Category:    "model",
			Inputs: []Port{
				{ID: "model", Name: "模型", Type: DataTypeModel, Required: true},
				{ID: "path", Name: "保存路径", Type: "text", Required: true},
			},
			Outputs: []Port{
				{ID: "saved", Name: "保存成功", Type: "boolean"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},

		// ---- Input/Output nodes ----
		{
			Type:        "io.input",
			Plugin:      "core",
			Name:        "用户输入",
			Description: "接收用户输入",
			Category:    "io",
			Inputs:      []Port{},
			Outputs: []Port{
				{ID: "value", Name: "输入值", Type: "any"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
		{
			Type:        "io.output",
			Plugin:      "core",
			Name:        "输出",
			Description: "输出结果",
			Category:    "io",
			Inputs: []Port{
				{ID: "data", Name: "数据", Type: "any", Required: true},
			},
			Outputs: []Port{},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
		{
			Type:        "io.file",
			Plugin:      "core",
			Name:        "文件输出",
			Description: "保存数据到文件",
			Category:    "io",
			Inputs: []Port{
				{ID: "data", Name: "数据", Type: "any", Required: true},
				{ID: "path", Name: "文件路径", Type: "text", Required: true},
			},
			Outputs: []Port{
				{ID: "success", Name: "成功", Type: "boolean"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},

		// ---- Math nodes ----
		{
			Type:        "math.add",
			Plugin:      "core",
			Name:        "加法",
			Description: "数值加法运算",
			Category:    "math",
			Inputs: []Port{
				{ID: "a", Name: "A", Type: "number", Required: true},
				{ID: "b", Name: "B", Type: "number", Required: true},
			},
			Outputs: []Port{
				{ID: "result", Name: "结果", Type: "number"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
		{
			Type:        "math.multiply",
			Plugin:      "core",
			Name:        "乘法",
			Description: "数值乘法运算",
			Category:    "math",
			Inputs: []Port{
				{ID: "a", Name: "A", Type: "number", Required: true},
				{ID: "b", Name: "B", Type: "number", Required: true},
			},
			Outputs: []Port{
				{ID: "result", Name: "结果", Type: "number"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},

		// ---- Logic nodes ----
		{
			Type:        "logic.compare",
			Plugin:      "core",
			Name:        "比较",
			Description: "数值比较运算",
			Category:    "logic",
			Inputs: []Port{
				{ID: "a", Name: "A", Type: "any", Required: true},
				{ID: "b", Name: "B", Type: "any", Required: true},
			},
			Outputs: []Port{
				{ID: "equal", Name: "相等", Type: "boolean"},
				{ID: "greater", Name: "大于", Type: "boolean"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
		{
			Type:        "logic.merge",
			Plugin:      "core",
			Name:        "合并",
			Description: "合并多个输入",
			Category:    "logic",
			Inputs: []Port{
				{ID: "input1", Name: "输入1", Type: "any", Required: false},
				{ID: "input2", Name: "输入2", Type: "any", Required: false},
				{ID: "input3", Name: "输入3", Type: "any", Required: false},
			},
			Outputs: []Port{
				{ID: "output", Name: "输出", Type: "any"},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},

		// ---- Simulation nodes ----
		{
			Type:        "simulation.sumo",
			Plugin:      "core",
			Name:        "SUMO 交通仿真",
			Description: "SUMO 交通仿真",
			Category:    "simulation",
			Inputs: []Port{
				{ID: "config", Name: "配置", Type: DataTypeJSON, Required: true},
			},
			Outputs: []Port{
				{ID: "simulation_result", Name: "仿真结果", Type: DataTypeJSON},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},

		// ---- MCP nodes ----
		{
			Type:        "mcp.tool",
			Plugin:      "core",
			Name:        "MCP 工具",
			Description: "调用 MCP 服务器工具",
			Category:    "mcp",
			Inputs: []Port{
				{ID: "server", Name: "服务器", Type: DataTypeStream, Required: true},
				{ID: "input", Name: "输入参数", Type: DataTypeJSON, Required: false},
			},
			Outputs: []Port{
				{ID: "output", Name: "工具输出", Type: DataTypeJSON},
			},
			Factory: func() ExecutableNode { return executableFunc(noOpExecutor) },
		},
	}
}

// ============================================================================
// Registry Helpers — Bootstrap the engine with built-in definitions
// ============================================================================

// RegisterBuiltInNodes registers all built-in node definitions with a registry.
func RegisterBuiltInNodes(reg *NodeRegistry) {
	defs := BuiltInNodeDefinitions()
	log.Printf("[builtin] Registering %d built-in nodes", len(defs))
	for _, def := range defs {
		reg.Register(def)
		log.Printf("[builtin] Registered node: %s (%s)", def.Type, def.Name)
	}
}

// NewEngineWithBuiltIns creates a new workflow engine with built-in nodes registered.
func NewEngineWithBuiltIns() *Engine {
	engine := NewEngine()
	RegisterBuiltInNodes(engine.registry)
	return engine
}

// ============================================================================
// Executable wrapper — allows plain functions as node executors
// ============================================================================

type executableFunc func(ctx context.Context, inputs map[string]interface{}, params map[string]interface{}) (map[string]interface{}, error)

func (f executableFunc) Execute(ctx context.Context, inputs map[string]interface{}, params map[string]interface{}) (map[string]interface{}, error) {
	return f(ctx, inputs, params)
}
