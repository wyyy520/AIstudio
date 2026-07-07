import type { NodeTemplate } from '../types/workflow'

// ============================================================
// 颜色快捷常量
// ============================================================

const C = {
  vision: 'var(--vision)',
  nlp: 'var(--nlp)',
  timeseries: 'var(--timeseries)',
  info: 'var(--info)',
  warning: 'var(--warning)',
  success: 'var(--success)',
  error: 'var(--error)',
  logic: 'var(--logic)',
  system: 'var(--neutral)',
  simulation: 'var(--simulation)',
  mcp: 'var(--mcp)',
  agent: 'var(--agent)',
}

// ============================================================
// 所有节点模板
// ============================================================

export const nodeTemplates: NodeTemplate[] = [

  // ======================== AI 模型 - Vision ========================
  {
    key: 'yolo_training',
    label: 'YOLO 训练',
    description: 'YOLO 目标检测模型训练',
    category: 'ai-vision',
    nodeType: 'vision',
    color: C.vision,
    icon: 'M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z M12 9a3 3 0 1 0 0 6 3 3 0 0 0 0-6z',
    inputs: [
      { name: 'dataset', label: '数据集', type: 'dataset', description: 'YOLO格式数据集' },
      { name: 'model_weight', label: '预训练权重', type: 'model', description: '可选预训练权重' },
    ],
    outputs: [
      { name: 'best_pt', label: 'best.pt', type: 'model' },
      { name: 'last_pt', label: 'last.pt', type: 'model' },
      { name: 'training_results', label: '训练结果', type: 'result' },
    ],
    params: [
      { name: 'model_version', label: '模型版本', type: 'select', default: 'yolov8', category: '基础设置', options: [
        { label: 'YOLOv5', value: 'yolov5' }, { label: 'YOLOv8', value: 'yolov8' },
        { label: 'YOLOv9', value: 'yolov9' }, { label: 'YOLOv10', value: 'yolov10' },
        { label: 'YOLOv11', value: 'yolov11' },
      ]},
      { name: 'task_type', label: '任务类型', type: 'select', default: 'detection', category: '基础设置', options: [
        { label: 'Detection', value: 'detection' }, { label: 'Segmentation', value: 'segmentation' },
        { label: 'Classification', value: 'classification' }, { label: 'Pose', value: 'pose' },
      ]},
      { name: 'epochs', label: 'Epoch', type: 'number', default: 100, min: 1, max: 10000, category: '训练参数', validation: { rule: '> 0', message: 'Epoch 必须大于 0' } },
      { name: 'batch_size', label: 'Batch Size', type: 'number', default: 16, min: 1, max: 512, category: '训练参数' },
      { name: 'image_size', label: 'Image Size', type: 'number', default: 640, min: 32, max: 2048, step: 32, category: '训练参数' },
      { name: 'optimizer', label: 'Optimizer', type: 'select', default: 'SGD', category: '训练参数', options: [
        { label: 'SGD', value: 'SGD' }, { label: 'Adam', value: 'Adam' }, { label: 'AdamW', value: 'AdamW' },
      ]},
      { name: 'learning_rate', label: 'Learning Rate', type: 'number', default: 0.01, min: 0.0001, max: 1, step: 0.0001, category: '训练参数' },
      { name: 'device', label: '设备', type: 'select', default: 'cuda', category: '基础设置', options: [
        { label: 'CPU', value: 'cpu' }, { label: 'CUDA GPU', value: 'cuda' },
      ]},
      { name: 'gpu_select', label: 'GPU 选择', type: 'select', default: '0', category: '基础设置', options: [
        { label: 'RTX 3060', value: '0' }, { label: 'RTX 4090', value: '1' }, { label: 'Auto', value: 'auto' },
      ]},
      { name: 'mosaic', label: 'Mosaic 增强', type: 'switch', default: true, category: '高级设置' },
      { name: 'flip', label: 'Flip 翻转', type: 'switch', default: true, category: '高级设置' },
      { name: 'rotation', label: 'Rotation', type: 'switch', default: false, category: '高级设置' },
      { name: 'hsv', label: 'HSV 增强', type: 'switch', default: true, category: '高级设置' },
      { name: 'output_onnx', label: '导出 ONNX', type: 'switch', default: false, category: '输出' },
    ],
    paramsLayout: 'collapsible',
  },
  {
    key: 'yolo_inference',
    label: 'YOLO 推理',
    description: 'YOLO 模型推理',
    category: 'ai-vision',
    nodeType: 'vision',
    color: C.vision,
    icon: 'M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z M12 9a3 3 0 1 0 0 6 3 3 0 0 0 0-6z',
    inputs: [
      { name: 'image', label: '图像', type: 'image' },
      { name: 'model', label: '模型', type: 'model' },
    ],
    outputs: [
      { name: 'detections', label: '检测结果', type: 'result' },
      { name: 'annotated', label: '标注图像', type: 'image' },
    ],
    params: [
      { name: 'confidence', label: '置信度阈值', type: 'slider', default: 0.5, min: 0, max: 1, step: 0.05, category: '推理参数' },
      { name: 'iou', label: 'IoU 阈值', type: 'slider', default: 0.45, min: 0, max: 1, step: 0.05, category: '推理参数' },
      { name: 'device', label: '设备', type: 'select', default: 'cuda', category: '基础设置', options: [
        { label: 'CPU', value: 'cpu' }, { label: 'CUDA GPU', value: 'cuda' },
      ]},
    ],
  },
  {
    key: 'cnn',
    label: 'CNN',
    description: '卷积神经网络',
    category: 'ai-vision',
    nodeType: 'vision',
    color: C.vision,
    icon: 'M3 3h18v18H3z M7 7h4v4H7z M13 7h4v4h-4z M7 13h4v4H7z M13 13h4v4h-4z',
    inputs: [{ name: 'image', label: '图像', type: 'image' }],
    outputs: [{ name: 'features', label: '特征图', type: 'tensor' }],
    params: [
      { name: 'kernel_size', label: 'Kernel Size', type: 'number', default: 3, min: 1, max: 15, step: 2, category: '卷积参数' },
      { name: 'stride', label: 'Stride', type: 'number', default: 1, min: 1, max: 5, category: '卷积参数' },
      { name: 'padding', label: 'Padding', type: 'number', default: 1, min: 0, max: 5, category: '卷积参数' },
      { name: 'channels', label: '输出通道', type: 'number', default: 64, min: 1, max: 2048, step: 16, category: '卷积参数' },
      { name: 'activation', label: '激活函数', type: 'select', default: 'relu', category: '基础设置', options: [
        { label: 'ReLU', value: 'relu' }, { label: 'LeakyReLU', value: 'leaky_relu' },
        { label: 'SiLU', value: 'silu' }, { label: 'GELU', value: 'gelu' },
      ]},
    ],
  },
  {
    key: 'resnet',
    label: 'ResNet',
    description: '残差网络',
    category: 'ai-vision',
    nodeType: 'vision',
    color: C.vision,
    icon: 'M3 3h7v7H3V3zm0 11h7v7H3v-7zm11-11h7v7h-7V3zm0 11h7v7h-7v-7z',
    inputs: [{ name: 'image', label: '图像', type: 'image' }],
    outputs: [
      { name: 'features', label: '特征', type: 'tensor' },
      { name: 'logits', label: '分类结果', type: 'result' },
    ],
    params: [
      { name: 'variant', label: '模型变体', type: 'select', default: 'resnet50', category: '基础设置', options: [
        { label: 'ResNet-18', value: 'resnet18' }, { label: 'ResNet-34', value: 'resnet34' },
        { label: 'ResNet-50', value: 'resnet50' }, { label: 'ResNet-101', value: 'resnet101' },
      ]},
      { name: 'pretrained', label: '预训练权重', type: 'switch', default: true, category: '基础设置' },
      { name: 'num_classes', label: '类别数', type: 'number', default: 1000, min: 1, max: 100000, category: '基础设置' },
    ],
  },
  {
    key: 'sam',
    label: 'SAM 分割',
    description: 'Segment Anything Model',
    category: 'ai-vision',
    nodeType: 'vision',
    color: C.vision,
    icon: 'M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z M12 9a3 3 0 1 0 0 6 3 3 0 0 0 0-6z',
    inputs: [
      { name: 'image', label: '图像', type: 'image' },
      { name: 'points', label: '提示点', type: 'json' },
    ],
    outputs: [
      { name: 'masks', label: '分割掩码', type: 'image' },
      { name: 'scores', label: '置信度', type: 'json' },
    ],
    params: [
      { name: 'model_type', label: '模型类型', type: 'select', default: 'vit_h', category: '基础设置', options: [
        { label: 'ViT-H', value: 'vit_h' }, { label: 'ViT-L', value: 'vit_l' }, { label: 'ViT-B', value: 'vit_b' },
      ]},
      { name: 'multimask', label: '多掩码输出', type: 'switch', default: true, category: '高级设置' },
    ],
  },
  {
    key: 'detr',
    label: 'DETR',
    description: 'Detection Transformer',
    category: 'ai-vision',
    nodeType: 'vision',
    color: C.vision,
    icon: 'M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z M12 9a3 3 0 1 0 0 6 3 3 0 0 0 0-6z',
    inputs: [{ name: 'image', label: '图像', type: 'image' }],
    outputs: [{ name: 'detections', label: '检测结果', type: 'result' }],
    params: [
      { name: 'num_queries', label: '查询数量', type: 'number', default: 100, min: 1, max: 500, category: '基础设置' },
      { name: 'backbone', label: '骨干网络', type: 'select', default: 'resnet50', category: '基础设置', options: [
        { label: 'ResNet-50', value: 'resnet50' }, { label: 'ResNet-101', value: 'resnet101' },
      ]},
    ],
  },

  // ======================== AI 模型 - NLP ========================
  {
    key: 'transformer',
    label: 'Transformer',
    description: '通用 Transformer 模型',
    category: 'ai-nlp',
    nodeType: 'nlp',
    color: C.nlp,
    icon: 'M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z',
    inputs: [{ name: 'text', label: '文本', type: 'text' }],
    outputs: [{ name: 'output', label: '输出', type: 'tensor' }],
    params: [
      { name: 'model_type', label: '模型', type: 'select', default: 'bert', category: '基础设置', options: [
        { label: 'BERT', value: 'bert' }, { label: 'GPT', value: 'gpt' }, { label: 'LLaMA', value: 'llama' },
      ]},
      { name: 'batch_size', label: 'Batch Size', type: 'number', default: 32, min: 1, max: 512, category: '训练参数' },
      { name: 'learning_rate', label: 'Learning Rate', type: 'number', default: 0.0001, min: 0.00001, max: 0.1, step: 0.00001, category: '训练参数' },
      { name: 'epochs', label: 'Epoch', type: 'number', default: 3, min: 1, max: 100, category: '训练参数', validation: { rule: '> 0', message: 'Epoch 必须大于 0' } },
      { name: 'max_length', label: '最大长度', type: 'number', default: 512, min: 16, max: 4096, category: '基础设置' },
    ],
    paramsLayout: 'collapsible',
  },
  {
    key: 'bert',
    label: 'BERT',
    description: 'BERT 预训练模型',
    category: 'ai-nlp',
    nodeType: 'nlp',
    color: C.nlp,
    icon: 'M4 4h16v16H4z M8 8h8 M8 12h4 M8 16h8',
    inputs: [{ name: 'text', label: '文本', type: 'text' }],
    outputs: [
      { name: 'embeddings', label: '嵌入向量', type: 'tensor' },
      { name: 'logits', label: '分类结果', type: 'result' },
    ],
    params: [
      { name: 'variant', label: '模型变体', type: 'select', default: 'bert_base', category: '基础设置', options: [
        { label: 'BERT-Base', value: 'bert_base' }, { label: 'BERT-Large', value: 'bert_large' },
        { label: 'RoBERTa', value: 'roberta' }, { label: 'DistilBERT', value: 'distilbert' },
      ]},
      { name: 'task', label: '任务', type: 'select', default: 'classification', category: '基础设置', options: [
        { label: '文本分类', value: 'classification' }, { label: '命名实体识别', value: 'ner' }, { label: '问答', value: 'qa' },
      ]},
      { name: 'num_labels', label: '标签数', type: 'number', default: 2, min: 2, max: 1000, category: '基础设置' },
    ],
  },
  {
    key: 'llm',
    label: 'LLM 对话',
    description: '大语言模型推理',
    category: 'ai-nlp',
    nodeType: 'nlp',
    color: C.nlp,
    icon: 'M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z M8 9h8 M8 12h6',
    inputs: [
      { name: 'prompt', label: '提示词', type: 'text' },
      { name: 'context', label: '上下文', type: 'text' },
    ],
    outputs: [
      { name: 'response', label: '回复', type: 'text' },
      { name: 'usage', label: '用量', type: 'json' },
    ],
    params: [
      { name: 'model', label: '模型', type: 'select', default: 'gpt-4', category: '基础设置', options: [
        { label: 'GPT-4', value: 'gpt-4' }, { label: 'GPT-4o', value: 'gpt-4o' },
        { label: 'Claude 3.5', value: 'claude-3.5' }, { label: 'LLaMA 3', value: 'llama-3' },
      ]},
      { name: 'temperature', label: 'Temperature', type: 'slider', default: 0.7, min: 0, max: 2, step: 0.1, category: '推理参数' },
      { name: 'max_tokens', label: 'Max Tokens', type: 'number', default: 2048, min: 1, max: 128000, category: '推理参数' },
      { name: 'system_prompt', label: 'System Prompt', type: 'text', default: '', category: '高级设置', placeholder: '可选的系统提示词...' },
    ],
  },

  // ======================== AI 模型 - Time Series ========================
  {
    key: 'lstm',
    label: 'LSTM',
    description: '长短期记忆网络',
    category: 'ai-timeseries',
    nodeType: 'timeseries',
    color: C.timeseries,
    icon: 'M3 12h3 M18 12h3 M12 3v3 M12 18v3 M8 8l4 4-4 4 M16 8l-4 4 4 4',
    inputs: [{ name: 'sequence', label: '序列', type: 'tensor' }],
    outputs: [
      { name: 'output', label: '输出', type: 'tensor' },
      { name: 'hidden', label: '隐状态', type: 'tensor' },
    ],
    params: [
      { name: 'input_size', label: 'Input Size', type: 'number', default: 128, min: 1, max: 4096, category: '基础设置' },
      { name: 'hidden_size', label: 'Hidden Size', type: 'number', default: 256, min: 16, max: 4096, category: '基础设置' },
      { name: 'num_layers', label: 'Layers', type: 'number', default: 2, min: 1, max: 10, category: '基础设置' },
      { name: 'seq_length', label: 'Sequence Length', type: 'number', default: 100, min: 1, max: 10000, category: '基础设置' },
      { name: 'epochs', label: 'Epoch', type: 'number', default: 50, min: 1, max: 10000, category: '训练参数', validation: { rule: '> 0', message: 'Epoch 必须大于 0' } },
      { name: 'bidirectional', label: '双向', type: 'switch', default: false, category: '高级设置' },
      { name: 'dropout', label: 'Dropout', type: 'slider', default: 0.2, min: 0, max: 0.9, step: 0.05, category: '高级设置' },
    ],
    paramsLayout: 'collapsible',
  },
  {
    key: 'gru',
    label: 'GRU',
    description: '门控循环单元',
    category: 'ai-timeseries',
    nodeType: 'timeseries',
    color: C.timeseries,
    icon: 'M3 12h3 M18 12h3 M12 3v3 M12 18v3 M8 8l4 4-4 4 M16 8l-4 4 4 4',
    inputs: [{ name: 'sequence', label: '序列', type: 'tensor' }],
    outputs: [{ name: 'output', label: '输出', type: 'tensor' }],
    params: [
      { name: 'input_size', label: 'Input Size', type: 'number', default: 128, min: 1, max: 4096, category: '基础设置' },
      { name: 'hidden_size', label: 'Hidden Size', type: 'number', default: 256, min: 16, max: 4096, category: '基础设置' },
      { name: 'num_layers', label: 'Layers', type: 'number', default: 2, min: 1, max: 10, category: '基础设置' },
      { name: 'epochs', label: 'Epoch', type: 'number', default: 50, min: 1, max: 10000, category: '训练参数', validation: { rule: '> 0', message: 'Epoch 必须大于 0' } },
    ],
  },
  {
    key: 'transformer_forecast',
    label: 'Transformer Forecast',
    description: 'Transformer 时序预测',
    category: 'ai-timeseries',
    nodeType: 'timeseries',
    color: C.timeseries,
    icon: 'M3 16.5 9 10.5 13 14.5 21 6.5 M21 6.5 13 14.5 9 10.5 3 16.5',
    inputs: [{ name: 'sequence', label: '历史序列', type: 'tensor' }],
    outputs: [{ name: 'forecast', label: '预测值', type: 'tensor' }],
    params: [
      { name: 'pred_len', label: '预测长度', type: 'number', default: 96, min: 1, max: 1000, category: '基础设置' },
      { name: 'd_model', label: '模型维度', type: 'number', default: 512, min: 64, max: 2048, category: '基础设置' },
      { name: 'n_heads', label: '注意力头数', type: 'number', default: 8, min: 1, max: 32, category: '基础设置' },
      { name: 'epochs', label: 'Epoch', type: 'number', default: 10, min: 1, max: 1000, category: '训练参数', validation: { rule: '> 0', message: 'Epoch 必须大于 0' } },
    ],
  },

  // ======================== AI 模型 - Audio ========================
  {
    key: 'whisper',
    label: 'Whisper',
    description: 'OpenAI 语音识别',
    category: 'ai-audio',
    nodeType: 'speech',
    color: C.warning,
    icon: 'M4 12h3 M17 12h3 M12 4v3 M12 17v3 M8 8l4 4-4 4 M16 8l-4 4 4 4',
    inputs: [{ name: 'audio', label: '音频', type: 'audio' }],
    outputs: [
      { name: 'text', label: '转录文本', type: 'text' },
      { name: 'segments', label: '片段', type: 'json' },
    ],
    params: [
      { name: 'model_size', label: '模型大小', type: 'select', default: 'medium', category: '基础设置', options: [
        { label: 'Tiny', value: 'tiny' }, { label: 'Base', value: 'base' },
        { label: 'Small', value: 'small' }, { label: 'Medium', value: 'medium' }, { label: 'Large', value: 'large' },
      ]},
      { name: 'language', label: '语言', type: 'select', default: 'zh', category: '基础设置', options: [
        { label: '中文', value: 'zh' }, { label: '英文', value: 'en' },
        { label: '日语', value: 'ja' }, { label: '自动检测', value: 'auto' },
      ]},
      { name: 'task', label: '任务', type: 'select', default: 'transcribe', category: '基础设置', options: [
        { label: '转录', value: 'transcribe' }, { label: '翻译', value: 'translate' },
      ]},
    ],
  },
  {
    key: 'speech_recognition',
    label: '语音识别',
    description: '通用语音识别',
    category: 'ai-audio',
    nodeType: 'speech',
    color: C.warning,
    icon: 'M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z M19 10v2a7 7 0 0 1-14 0v-2 M12 19v3',
    inputs: [{ name: 'audio', label: '音频', type: 'audio' }],
    outputs: [{ name: 'text', label: '文本', type: 'text' }],
    params: [
      { name: 'engine', label: '引擎', type: 'select', default: 'whisper', category: '基础设置', options: [
        { label: 'Whisper', value: 'whisper' }, { label: 'DeepSpeech', value: 'deepspeech' }, { label: 'Wav2Vec', value: 'wav2vec' },
      ]},
      { name: 'sample_rate', label: '采样率', type: 'number', default: 16000, min: 8000, max: 48000, category: '基础设置' },
    ],
  },

  // ======================== 数据处理 ========================
  {
    key: 'dataset_loader',
    label: 'Dataset 加载器',
    description: '加载数据集',
    category: 'data',
    nodeType: 'dataset',
    color: C.timeseries,
    icon: 'M3 3v18h18 M18.5 9h.01 M15.5 17h.01 M11.5 14h.01 M8.5 11h.01',
    inputs: [],
    outputs: [
      { name: 'dataset', label: '数据集', type: 'dataset' },
      { name: 'info', label: '数据集信息', type: 'json' },
    ],
    params: [
      { name: 'data_path', label: '数据路径', type: 'text', default: '/data/dataset', category: '基础设置', required: true },
      { name: 'data_format', label: '数据格式', type: 'select', default: 'yolo', category: '基础设置', options: [
        { label: 'YOLO', value: 'yolo' }, { label: 'COCO', value: 'coco' }, { label: 'VOC', value: 'voc' },
      ]},
      { name: 'train_ratio', label: '训练集比例', type: 'slider', default: 0.8, min: 0.1, max: 0.95, step: 0.05, category: '数据划分' },
      { name: 'val_ratio', label: '验证集比例', type: 'slider', default: 0.1, min: 0.05, max: 0.5, step: 0.05, category: '数据划分' },
      { name: 'num_classes', label: '类别数量', type: 'number', default: 80, min: 1, max: 10000, category: '基础设置' },
    ],
    paramsLayout: 'collapsible',
  },
  {
    key: 'image_loader',
    label: '图像加载器',
    description: '加载图像文件',
    category: 'data',
    nodeType: 'dataset',
    color: C.info,
    icon: 'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z M14 2v6h6 M16 13H8 M16 17H8 M10 9H8',
    inputs: [],
    outputs: [{ name: 'images', label: '图像列表', type: 'image' }],
    params: [
      { name: 'folder_path', label: '文件夹路径', type: 'text', default: '/data/images', category: '基础设置', required: true },
      { name: 'resize', label: 'Resize', type: 'number', default: 640, min: 32, max: 4096, category: '预处理' },
      { name: 'grayscale', label: '灰度化', type: 'switch', default: false, category: '预处理' },
    ],
  },
  {
    key: 'csv_loader',
    label: 'CSV 加载器',
    description: '加载 CSV 数据',
    category: 'data',
    nodeType: 'dataset',
    color: C.timeseries,
    icon: 'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z M14 2v6h6 M16 13H8 M16 17H8 M10 9H8',
    inputs: [],
    outputs: [{ name: 'data', label: '数据', type: 'dataset' }],
    params: [
      { name: 'file_path', label: '文件路径', type: 'text', default: '/data/data.csv', category: '基础设置', required: true },
      { name: 'delimiter', label: '分隔符', type: 'select', default: ',', category: '基础设置', options: [
        { label: '逗号 (,)', value: ',' }, { label: '制表符', value: '\t' }, { label: '分号 (;)', value: ';' },
      ]},
      { name: 'has_header', label: '包含表头', type: 'switch', default: true, category: '基础设置' },
    ],
  },
  {
    key: 'data_cleaning',
    label: '数据清洗',
    description: '数据预处理与清洗',
    category: 'data',
    nodeType: 'dataset',
    color: C.info,
    icon: 'M22 3H2l8 9.46V19l4 2v-8.54L22 3z',
    inputs: [{ name: 'data', label: '原始数据', type: 'dataset' }],
    outputs: [{ name: 'cleaned', label: '清洗后数据', type: 'dataset' }],
    params: [
      { name: 'remove_null', label: '移除空值', type: 'switch', default: true, category: '基础设置' },
      { name: 'remove_duplicates', label: '去重', type: 'switch', default: true, category: '基础设置' },
      { name: 'normalize', label: '标准化', type: 'switch', default: false, category: '高级设置' },
    ],
  },
  {
    key: 'data_augmentation',
    label: '数据增强',
    description: '数据增强处理',
    category: 'data',
    nodeType: 'dataset',
    color: C.info,
    icon: 'M22 3H2l8 9.46V19l4 2v-8.54L22 3z',
    inputs: [{ name: 'data', label: '原始数据', type: 'dataset' }],
    outputs: [{ name: 'augmented', label: '增强后数据', type: 'dataset' }],
    params: [
      { name: 'flip_h', label: '水平翻转', type: 'switch', default: true, category: '图像增强' },
      { name: 'flip_v', label: '垂直翻转', type: 'switch', default: false, category: '图像增强' },
      { name: 'rotation', label: '旋转角度', type: 'number', default: 15, min: 0, max: 180, category: '图像增强' },
      { name: 'brightness', label: '亮度调整', type: 'slider', default: 0.2, min: 0, max: 1, step: 0.05, category: '图像增强' },
    ],
  },
  {
    key: 'data_split',
    label: '数据划分',
    description: '划分训练/验证/测试集',
    category: 'data',
    nodeType: 'dataset',
    color: C.info,
    icon: 'M8 6h8M8 12h8M8 18h8 M16 3l5 3-5 3 M8 3L3 6l5 3',
    inputs: [{ name: 'data', label: '数据集', type: 'dataset' }],
    outputs: [
      { name: 'train', label: '训练集', type: 'dataset' },
      { name: 'val', label: '验证集', type: 'dataset' },
      { name: 'test', label: '测试集', type: 'dataset' },
    ],
    params: [
      { name: 'train_ratio', label: '训练集比例', type: 'slider', default: 0.7, min: 0.1, max: 0.9, step: 0.05, category: '基础设置' },
      { name: 'val_ratio', label: '验证集比例', type: 'slider', default: 0.15, min: 0.05, max: 0.5, step: 0.05, category: '基础设置' },
      { name: 'seed', label: '随机种子', type: 'number', default: 42, min: 0, max: 99999, category: '高级设置' },
    ],
  },

  // ======================== 训练 ========================
  {
    key: 'train',
    label: '训练',
    description: '模型训练',
    category: 'training',
    nodeType: 'training',
    color: C.success,
    icon: 'M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z M2 12h3 M19 12h3 M12 2v3 M12 19v3',
    inputs: [
      { name: 'model', label: '模型', type: 'model' },
      { name: 'dataset', label: '训练数据', type: 'dataset' },
    ],
    outputs: [
      { name: 'trained_model', label: '训练后模型', type: 'model' },
      { name: 'metrics', label: '训练指标', type: 'result' },
    ],
    params: [
      { name: 'epochs', label: 'Epoch', type: 'number', default: 100, min: 1, max: 10000, category: '基础设置', validation: { rule: '> 0', message: 'Epoch 必须大于 0' } },
      { name: 'batch_size', label: 'Batch Size', type: 'number', default: 32, min: 1, max: 512, category: '基础设置' },
      { name: 'lr', label: 'Learning Rate', type: 'number', default: 0.001, min: 0.00001, max: 1, step: 0.0001, category: '基础设置' },
      { name: 'optimizer', label: 'Optimizer', type: 'select', default: 'Adam', category: '基础设置', options: [
        { label: 'Adam', value: 'Adam' }, { label: 'AdamW', value: 'AdamW' }, { label: 'SGD', value: 'SGD' },
      ]},
      { name: 'early_stopping', label: 'Early Stopping', type: 'switch', default: true, category: '高级设置' },
      { name: 'patience', label: 'Patience', type: 'number', default: 10, min: 1, max: 100, category: '高级设置' },
    ],
  },
  {
    key: 'validation',
    label: '验证',
    description: '模型验证',
    category: 'training',
    nodeType: 'training',
    color: C.info,
    icon: 'M22 11.08V12a10 10 0 1 1-5.93-9.14 M22 4 12 14.01l-3-3',
    inputs: [
      { name: 'model', label: '模型', type: 'model' },
      { name: 'dataset', label: '验证数据', type: 'dataset' },
    ],
    outputs: [{ name: 'metrics', label: '验证指标', type: 'result' }],
    params: [
      { name: 'batch_size', label: 'Batch Size', type: 'number', default: 32, min: 1, max: 512, category: '基础设置' },
    ],
  },
  {
    key: 'test',
    label: '测试',
    description: '模型测试',
    category: 'training',
    nodeType: 'training',
    color: C.warning,
    icon: 'M5 12h14 M12 5l7 7-7 7',
    inputs: [
      { name: 'model', label: '模型', type: 'model' },
      { name: 'dataset', label: '测试数据', type: 'dataset' },
    ],
    outputs: [{ name: 'results', label: '测试结果', type: 'result' }],
    params: [
      { name: 'batch_size', label: 'Batch Size', type: 'number', default: 32, min: 1, max: 512, category: '基础设置' },
    ],
  },
  {
    key: 'checkpoint',
    label: 'Checkpoint',
    description: '模型检查点保存',
    category: 'training',
    nodeType: 'training',
    color: C.success,
    icon: 'M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z M17 21v-8H7v8 M7 3v5h8',
    inputs: [{ name: 'model', label: '模型', type: 'model' }],
    outputs: [{ name: 'checkpoint', label: '检查点', type: 'model' }],
    params: [
      { name: 'save_path', label: '保存路径', type: 'text', default: '/checkpoints', category: '基础设置' },
      { name: 'save_freq', label: '保存频率(Epoch)', type: 'number', default: 10, min: 1, max: 1000, category: '基础设置' },
      { name: 'save_best_only', label: '仅保存最优', type: 'switch', default: true, category: '高级设置' },
    ],
  },
  {
    key: 'export_model',
    label: '导出模型',
    description: '导出模型格式',
    category: 'training',
    nodeType: 'deployment',
    color: C.mcp,
    icon: 'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z M14 2v6h6',
    inputs: [{ name: 'model', label: '模型', type: 'model' }],
    outputs: [{ name: 'exported', label: '导出模型', type: 'model' }],
    params: [
      { name: 'format', label: '导出格式', type: 'select', default: 'torchscript', category: '基础设置', options: [
        { label: 'TorchScript', value: 'torchscript' }, { label: 'ONNX', value: 'onnx' },
        { label: 'TensorRT', value: 'tensorrt' }, { label: 'TFLite', value: 'tflite' },
      ]},
      { name: 'half', label: 'FP16 半精度', type: 'switch', default: false, category: '高级设置' },
      { name: 'dynamic', label: '动态尺寸', type: 'switch', default: false, category: '高级设置' },
    ],
  },

  // ======================== 部署 ========================
  {
    key: 'onnx_export',
    label: 'ONNX 导出',
    description: '导出 ONNX 格式',
    category: 'deployment',
    nodeType: 'deployment',
    color: C.mcp,
    icon: 'M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z',
    inputs: [{ name: 'model', label: '模型', type: 'model' }],
    outputs: [{ name: 'onnx_model', label: 'ONNX 模型', type: 'model' }],
    params: [
      { name: 'opset', label: 'Opset 版本', type: 'number', default: 17, min: 9, max: 20, category: '基础设置' },
      { name: 'simplify', label: '简化模型', type: 'switch', default: true, category: '高级设置' },
    ],
  },
  {
    key: 'tensorrt',
    label: 'TensorRT',
    description: 'TensorRT 推理加速',
    category: 'deployment',
    nodeType: 'deployment',
    color: C.mcp,
    icon: 'M4 4h16v12H4z M7 7h4v4H7z M13 7h4v4h-4z M7 13h10v2H7z',
    inputs: [{ name: 'model', label: '模型', type: 'model' }],
    outputs: [{ name: 'trt_engine', label: 'TRT 引擎', type: 'model' }],
    params: [
      { name: 'precision', label: '精度', type: 'select', default: 'fp16', category: '基础设置', options: [
        { label: 'FP32', value: 'fp32' }, { label: 'FP16', value: 'fp16' }, { label: 'INT8', value: 'int8' },
      ]},
      { name: 'max_batch', label: '最大 Batch', type: 'number', default: 16, min: 1, max: 256, category: '基础设置' },
    ],
  },
  {
    key: 'api_server',
    label: 'API Server',
    description: '部署为 API 服务',
    category: 'deployment',
    nodeType: 'deployment',
    color: C.mcp,
    icon: 'M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71 M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71',
    inputs: [{ name: 'model', label: '模型', type: 'model' }],
    outputs: [{ name: 'endpoint', label: 'API 端点', type: 'service' }],
    params: [
      { name: 'port', label: '端口', type: 'number', default: 8000, min: 1024, max: 65535, category: '基础设置' },
      { name: 'host', label: 'Host', type: 'text', default: '0.0.0.0', category: '基础设置' },
      { name: 'workers', label: 'Worker 数', type: 'number', default: 4, min: 1, max: 32, category: '基础设置' },
    ],
  },
  {
    key: 'docker',
    label: 'Docker',
    description: 'Docker 容器化部署',
    category: 'deployment',
    nodeType: 'deployment',
    color: C.mcp,
    icon: 'M4 11h16v10H4z M4 6h3v3H4z M9 6h3v3H9z M14 6h3v3h-3z M9 11h3v3H9z M14 11h3v3h-3z',
    inputs: [
      { name: 'model', label: '模型', type: 'model' },
      { name: 'config', label: '配置', type: 'json' },
    ],
    outputs: [{ name: 'image', label: 'Docker 镜像', type: 'service' }],
    params: [
      { name: 'base_image', label: '基础镜像', type: 'select', default: 'python:3.10', category: '基础设置', options: [
        { label: 'Python 3.10', value: 'python:3.10' }, { label: 'Python 3.11', value: 'python:3.11' },
        { label: 'CUDA 12.1', value: 'nvidia/cuda:12.1-runtime' }, { label: 'CUDA 12.4', value: 'nvidia/cuda:12.4-runtime' },
      ]},
      { name: 'tag', label: '镜像标签', type: 'text', default: 'latest', category: '基础设置' },
      { name: 'expose_port', label: '暴露端口', type: 'number', default: 8000, min: 1024, max: 65535, category: '基础设置' },
    ],
  },
  {
    key: 'edge_deployment',
    label: '边缘部署',
    description: '边缘设备部署',
    category: 'deployment',
    nodeType: 'deployment',
    color: C.mcp,
    icon: 'M21 12a9 9 0 0 1-9 9m9-9a9 9 0 0 0-9-9m9 9H3',
    inputs: [{ name: 'model', label: '模型', type: 'model' }],
    outputs: [{ name: 'bundle', label: '部署包', type: 'service' }],
    params: [
      { name: 'target', label: '目标设备', type: 'select', default: 'jetson', category: '基础设置', options: [
        { label: 'Jetson Nano', value: 'jetson_nano' }, { label: 'Jetson Orin', value: 'jetson_orin' },
        { label: 'Raspberry Pi', value: 'raspberry_pi' }, { label: 'Mobile', value: 'mobile' },
      ]},
      { name: 'quantization', label: '量化', type: 'select', default: 'int8', category: '高级设置', options: [
        { label: 'FP16', value: 'fp16' }, { label: 'INT8', value: 'int8' }, { label: 'INT4', value: 'int4' },
      ]},
    ],
  },

  // ======================== 逻辑控制 ========================
  {
    key: 'if_condition',
    label: 'IF 条件',
    description: '条件判断分支',
    category: 'logic',
    nodeType: 'logic',
    color: C.logic,
    icon: 'M4 4h16v16H4z M8 8h8 M8 12h8 M8 16h8',
    inputs: [{ name: 'input', label: '输入', type: 'any' }],
    outputs: [
      { name: 'true', label: 'True', type: 'any' },
      { name: 'false', label: 'False', type: 'any' },
    ],
    params: [
      { name: 'condition', label: '条件表达式', type: 'text', default: 'value > 0.5', category: '基础设置', placeholder: '例如: value > 0.5', required: true },
      { name: 'field', label: '判断字段', type: 'text', default: 'score', category: '基础设置' },
    ],
  },
  {
    key: 'switch',
    label: 'Switch',
    description: '多路分支选择',
    category: 'logic',
    nodeType: 'logic',
    color: C.logic,
    icon: 'M4 4h16v16H4z M8 8h8 M8 12h8 M8 16h8',
    inputs: [{ name: 'input', label: '输入', type: 'any' }],
    outputs: [
      { name: 'case_0', label: 'Case 0', type: 'any' },
      { name: 'case_1', label: 'Case 1', type: 'any' },
      { name: 'case_2', label: 'Case 2', type: 'any' },
      { name: 'default', label: 'Default', type: 'any' },
    ],
    params: [
      { name: 'num_cases', label: '分支数量', type: 'number', default: 3, min: 1, max: 10, category: '基础设置' },
      { name: 'field', label: '判断字段', type: 'text', default: 'class_id', category: '基础设置' },
    ],
  },
  {
    key: 'loop',
    label: 'Loop',
    description: '循环迭代',
    category: 'logic',
    nodeType: 'logic',
    color: C.logic,
    icon: 'M17 2v4h-4 M7 22v-4h4 M2 12h3 M19 12h3',
    inputs: [{ name: 'input', label: '输入', type: 'any' }],
    outputs: [
      { name: 'iteration', label: '迭代输出', type: 'any' },
      { name: 'completed', label: '完成', type: 'any' },
    ],
    params: [
      { name: 'max_iterations', label: '最大迭代次数', type: 'number', default: 100, min: 1, max: 100000, category: '基础设置' },
      { name: 'parallel', label: '并行执行', type: 'switch', default: false, category: '高级设置' },
    ],
  },
  {
    key: 'condition',
    label: 'Condition',
    description: '通用条件节点',
    category: 'logic',
    nodeType: 'logic',
    color: C.logic,
    icon: 'M9 5H7a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2h-2 M9 5a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2v0a2 2 0 0 1-2 2h-2a2 2 0 0 1-2-2z M9 14l2 2 4-4',
    inputs: [{ name: 'input', label: '输入', type: 'any' }],
    outputs: [
      { name: 'pass', label: '通过', type: 'any' },
      { name: 'fail', label: '不通过', type: 'any' },
    ],
    params: [
      { name: 'condition', label: '条件表达式', type: 'text', default: 'result.status == "success"', category: '基础设置', required: true },
    ],
  },
  {
    key: 'retry',
    label: 'Retry',
    description: '失败重试',
    category: 'logic',
    nodeType: 'logic',
    color: C.logic,
    icon: 'M17 2v4h-4 M7 22v-4h4',
    inputs: [{ name: 'input', label: '输入', type: 'any' }],
    outputs: [
      { name: 'output', label: '输出', type: 'any' },
      { name: 'failed', label: '最终失败', type: 'any' },
    ],
    params: [
      { name: 'max_retries', label: '最大重试次数', type: 'number', default: 3, min: 1, max: 100, category: '基础设置' },
      { name: 'delay', label: '重试延迟(秒)', type: 'number', default: 5, min: 0, max: 3600, category: '基础设置' },
      { name: 'exponential_backoff', label: '指数退避', type: 'switch', default: true, category: '高级设置' },
    ],
  },

  // ======================== 系统工具 ========================
  {
    key: 'python_env',
    label: 'Python 环境',
    description: '配置 Python 运行环境',
    category: 'system',
    nodeType: 'system',
    color: C.system,
    icon: 'M12 2C7.6 2 4 4 4 4s0 3.6 0 5h8c0 0 0-1 2-1s4 0 4 0 0-2 0-4c0-2-2.5-3-6-3z M12 22c4.4 0 8-2 8-2s0-3.6 0-5h-8c0 0 0 1-2 1s-4 0-4 0 0 2 0 4c0 2 2.5 3 6 3z',
    inputs: [],
    outputs: [{ name: 'env', label: '环境', type: 'service' }],
    params: [
      { name: 'python_version', label: 'Python 版本', type: 'select', default: '3.10', category: '基础设置', options: [
        { label: 'Python 3.9', value: '3.9' }, { label: 'Python 3.10', value: '3.10' },
        { label: 'Python 3.11', value: '3.11' }, { label: 'Python 3.12', value: '3.12' },
      ]},
      { name: 'env_name', label: '环境名称', type: 'text', default: 'aistudio_env', category: '基础设置' },
      { name: 'requirements', label: '依赖包', type: 'text', default: 'torch torchvision', category: '高级设置', placeholder: 'pip install 包名, 换行分隔...' },
    ],
  },
  {
    key: 'cuda_check',
    label: 'CUDA 检查',
    description: '检查 CUDA 环境',
    category: 'system',
    nodeType: 'system',
    color: C.system,
    icon: 'M12 2L2 7l10 5 10-5-10-5z M2 17l10 5 10-5 M2 12l10 5 10-5',
    inputs: [],
    outputs: [
      { name: 'cuda_info', label: 'CUDA 信息', type: 'json' },
      { name: 'available', label: '是否可用', type: 'result' },
    ],
    params: [
      { name: 'min_version', label: '最低 CUDA 版本', type: 'text', default: '11.8', category: '基础设置' },
      { name: 'check_cudnn', label: '检查 cuDNN', type: 'switch', default: true, category: '高级设置' },
    ],
  },
  {
    key: 'install_dependency',
    label: '安装依赖',
    description: '安装 Python 依赖',
    category: 'system',
    nodeType: 'system',
    color: C.system,
    icon: 'M4 6h16M4 12h16M4 18h12 M8 2v4 M8 20v2 M16 2v4M16 20v2',
    inputs: [],
    outputs: [{ name: 'result', label: '安装结果', type: 'result' }],
    params: [
      { name: 'packages', label: '包列表', type: 'text', default: 'numpy\npandas\ntorch', category: '基础设置', placeholder: '每行一个包名', required: true },
      { name: 'upgrade', label: '升级已有包', type: 'switch', default: false, category: '高级设置' },
    ],
  },
  {
    key: 'git_clone',
    label: 'Git Clone',
    description: '克隆 Git 仓库',
    category: 'system',
    nodeType: 'system',
    color: C.system,
    icon: 'M15 22v-4a4.8 4.8 0 0 0-1-3.5c3 0 6-2 6-5.5.08-1.25-.27-2.48-1-3.5.28-1.15.28-2.35 0-3.5 0 0-1 0-3 1.5-2.64-.5-5.36-.5-8 0C6 2 5 2 5 2c-.3 1.15-.3 2.35 0 3.5A5.403 5.403 0 0 0 4 9c0 3.5 3 5.5 6 5.5-.39.49-.68 1.05-.85 1.65-.17.6-.22 1.23-.15 1.85v4 M12 18c-1 0-3 1-3 3',
    inputs: [],
    outputs: [{ name: 'repo_path', label: '仓库路径', type: 'service' }],
    params: [
      { name: 'repo_url', label: '仓库 URL', type: 'text', default: 'https://github.com/user/repo.git', category: '基础设置', required: true },
      { name: 'branch', label: '分支', type: 'text', default: 'main', category: '基础设置' },
      { name: 'target_dir', label: '目标目录', type: 'text', default: '/workspace', category: '基础设置' },
      { name: 'depth', label: '浅克隆深度', type: 'number', default: 1, min: 1, max: 100, category: '高级设置' },
    ],
  },
  {
    key: 'download_model',
    label: '下载模型',
    description: '下载预训练模型',
    category: 'system',
    nodeType: 'system',
    color: C.system,
    icon: 'M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4 M7 10l5 5 5-5 M12 15V3',
    inputs: [],
    outputs: [{ name: 'model', label: '模型', type: 'model' }],
    params: [
      { name: 'source', label: '来源', type: 'select', default: 'huggingface', category: '基础设置', options: [
        { label: 'HuggingFace', value: 'huggingface' }, { label: 'ModelScope', value: 'modelscope' },
        { label: 'PyTorch Hub', value: 'torchhub' }, { label: 'URL', value: 'url' },
      ]},
      { name: 'model_id', label: '模型 ID', type: 'text', default: 'ultralytics/yolov8n', category: '基础设置', required: true },
      { name: 'cache_dir', label: '缓存目录', type: 'text', default: '/models', category: '高级设置' },
    ],
  },

  // ======================== 仿真 ========================
  {
    key: 'sumo',
    label: 'SUMO',
    description: 'SUMO 交通仿真',
    category: 'simulation',
    nodeType: 'simulation',
    color: C.simulation,
    icon: 'M12 2L2 7l10 5 10-5-10-5z M2 17l10 5 10-5 M2 12l10 5 10-5',
    inputs: [
      { name: 'config', label: '配置文件', type: 'json' },
    ],
    outputs: [
      { name: 'sim_data', label: '仿真数据', type: 'result' },
    ],
    params: [
      { name: 'sumo_cfg', label: 'SUMO 配置路径', type: 'text', default: '/sim/sumo.sumocfg', category: '基础设置', required: true },
      { name: 'step_length', label: '步长(秒)', type: 'number', default: 1, min: 0.1, max: 60, step: 0.1, category: '基础设置' },
      { name: 'gui', label: 'GUI 模式', type: 'switch', default: false, category: '高级设置' },
    ],
  },
  {
    key: 'matlab',
    label: 'MATLAB',
    description: 'MATLAB 仿真',
    category: 'simulation',
    nodeType: 'simulation',
    color: C.simulation,
    icon: 'M4 6h16 M4 12h16 M4 18h16 M8 2v4 M8 20v2 M16 2v4 M16 20v2',
    inputs: [
      { name: 'script', label: '脚本', type: 'text' },
    ],
    outputs: [
      { name: 'result', label: '结果', type: 'result' },
    ],
    params: [
      { name: 'script_path', label: '脚本路径', type: 'text', default: '/sim/script.m', category: '基础设置', required: true },
      { name: 'engine', label: '引擎', type: 'select', default: 'octave', category: '基础设置', options: [
        { label: 'MATLAB', value: 'matlab' }, { label: 'Octave', value: 'octave' },
      ]},
    ],
  },
  {
    key: 'ros',
    label: 'ROS',
    description: 'ROS 机器人仿真',
    category: 'simulation',
    nodeType: 'simulation',
    color: C.simulation,
    icon: 'M12 2a10 10 0 1 0 10 10A10 10 0 0 0 12 2z M12 6a4 4 0 1 0 4 4 4 4 0 0 0-4-4z M12 14a2 2 0 1 0 2 2 2 2 0 0 0-2-2z',
    inputs: [
      { name: 'topic_in', label: '输入 Topic', type: 'json' },
    ],
    outputs: [
      { name: 'topic_out', label: '输出 Topic', type: 'json' },
    ],
    params: [
      { name: 'ros_version', label: 'ROS 版本', type: 'select', default: 'ros2', category: '基础设置', options: [
        { label: 'ROS 1', value: 'ros1' }, { label: 'ROS 2', value: 'ros2' },
      ]},
      { name: 'node_name', label: '节点名称', type: 'text', default: 'aistudio_node', category: '基础设置' },
      { name: 'topic_sub', label: '订阅 Topic', type: 'text', default: '/input', category: '基础设置' },
      { name: 'topic_pub', label: '发布 Topic', type: 'text', default: '/output', category: '基础设置' },
    ],
  },

  // ======================== MCP ========================
  {
    key: 'mcp_client',
    label: 'MCP Client',
    description: 'MCP 客户端调用',
    category: 'mcp',
    nodeType: 'mcp',
    color: C.mcp,
    icon: 'M13 2L3 14h9l-1 8 10-12h-9l1-8z',
    inputs: [
      { name: 'input', label: '输入', type: 'json' },
    ],
    outputs: [
      { name: 'output', label: '输出', type: 'json' },
    ],
    params: [
      { name: 'server_url', label: '服务地址', type: 'text', default: 'http://localhost:3000', category: '基础设置', required: true },
      { name: 'tool_name', label: '工具名称', type: 'select', default: 'query', category: '基础设置', options: [
        { label: 'Query', value: 'query' }, { label: 'Execute', value: 'execute' },
        { label: 'Search', value: 'search' }, { label: 'Analyze', value: 'analyze' },
      ]},
      { name: 'timeout', label: '超时(秒)', type: 'number', default: 30, min: 1, max: 300, category: '基础设置' },
      { name: 'retry', label: '失败重试', type: 'switch', default: false, category: '高级设置' },
    ],
  },
  {
    key: 'mcp_server',
    label: 'MCP Server',
    description: 'MCP 服务器节点',
    category: 'mcp',
    nodeType: 'mcp',
    color: C.mcp,
    icon: 'M13 2L3 14h9l-1 8 10-12h-9l1-8z',
    inputs: [
      { name: 'config', label: '配置', type: 'json' },
    ],
    outputs: [
      { name: 'service', label: '服务', type: 'service' },
    ],
    params: [
      { name: 'port', label: '端口', type: 'number', default: 3000, min: 1024, max: 65535, category: '基础设置' },
      { name: 'name', label: '服务名称', type: 'text', default: 'AIStudio MCP Server', category: '基础设置' },
      { name: 'tools', label: '工具列表', type: 'text', default: 'query\nexecute\nsearch', category: '基础设置', placeholder: '每行一个工具名' },
      { name: 'auth', label: '认证', type: 'select', default: 'none', category: '高级设置', options: [
        { label: '无', value: 'none' }, { label: 'API Key', value: 'api_key' }, { label: 'OAuth 2.0', value: 'oauth' },
      ]},
    ],
  },
  {
    key: 'mcp_tool_calling',
    label: 'Tool Calling',
    description: 'MCP 工具调用',
    category: 'mcp',
    nodeType: 'mcp',
    color: C.mcp,
    icon: 'M13 2L3 14h9l-1 8 10-12h-9l1-8z',
    inputs: [
      { name: 'params', label: '参数', type: 'json' },
    ],
    outputs: [
      { name: 'result', label: '结果', type: 'json' },
    ],
    params: [
      { name: 'tool_name', label: '工具名称', type: 'select', default: 'search', category: '基础设置', options: [
        { label: 'Search', value: 'search' }, { label: 'Analyze', value: 'analyze' },
        { label: 'Generate', value: 'generate' }, { label: 'Translate', value: 'translate' },
      ]},
      { name: 'max_tokens', label: 'Max Tokens', type: 'number', default: 4096, min: 1, max: 128000, category: '基础设置' },
      { name: 'stream', label: '流式输出', type: 'switch', default: false, category: '高级设置' },
    ],
  },
]