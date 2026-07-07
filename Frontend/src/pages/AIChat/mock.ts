import type {
  AIProvider,
  ChatMessage,
  Conversation,
  ChatContext,
  QuickAction,
} from './types'

export const mockProviders: AIProvider[] = [
  {
    id: 'openai',
    name: 'OpenAI',
    icon: 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 15h-2v-2h2v2zm0-4h-2V7h2v6z',
    apiBaseUrl: 'https://api.openai.com/v1',
    apiKey: 'sk-****',
    status: 'connected',
    models: [
      { id: 'gpt-4.1', name: 'GPT-4.1', providerId: 'openai', maxTokens: 128000, defaultTemperature: 0.7 },
      { id: 'gpt-4o', name: 'GPT-4o', providerId: 'openai', maxTokens: 128000, defaultTemperature: 0.7 },
      { id: 'gpt-4o-mini', name: 'GPT-4o Mini', providerId: 'openai', maxTokens: 128000, defaultTemperature: 0.7 },
    ],
  },
  {
    id: 'claude',
    name: 'Claude',
    icon: 'M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5',
    apiBaseUrl: 'https://api.anthropic.com/v1',
    apiKey: 'sk-ant-****',
    status: 'connected',
    models: [
      { id: 'claude-sonnet-4', name: 'Claude Sonnet 4', providerId: 'claude', maxTokens: 200000, defaultTemperature: 0.7 },
      { id: 'claude-opus-4', name: 'Claude Opus 4', providerId: 'claude', maxTokens: 200000, defaultTemperature: 0.7 },
      { id: 'claude-haiku-3.5', name: 'Claude Haiku 3.5', providerId: 'claude', maxTokens: 200000, defaultTemperature: 0.7 },
    ],
  },
  {
    id: 'deepseek',
    name: 'DeepSeek',
    icon: 'M13 2L3 14h9l-1 8 10-12h-9l1-8z',
    apiBaseUrl: 'https://api.deepseek.com/v1',
    apiKey: '',
    status: 'disconnected',
    models: [
      { id: 'deepseek-chat', name: 'DeepSeek V3', providerId: 'deepseek', maxTokens: 64000, defaultTemperature: 0.7 },
      { id: 'deepseek-reasoner', name: 'DeepSeek R1', providerId: 'deepseek', maxTokens: 64000, defaultTemperature: 0.7 },
    ],
  },
  {
    id: 'gemini',
    name: 'Gemini',
    icon: 'M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20zm0 18a8 8 0 1 1 0-16 8 8 0 0 1 0 16z',
    apiBaseUrl: 'https://generativelanguage.googleapis.com/v1',
    apiKey: '',
    status: 'disconnected',
    models: [
      { id: 'gemini-2.5-pro', name: 'Gemini 2.5 Pro', providerId: 'gemini', maxTokens: 1000000, defaultTemperature: 0.7 },
    ],
  },
  {
    id: 'qwen',
    name: '通义千问',
    icon: 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2z',
    apiBaseUrl: 'https://dashscope.aliyuncs.com/api/v1',
    apiKey: '',
    status: 'disconnected',
    models: [
      { id: 'qwen-max', name: 'Qwen Max', providerId: 'qwen', maxTokens: 32000, defaultTemperature: 0.7 },
    ],
  },
  {
    id: 'ollama',
    name: 'Ollama',
    icon: 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8z',
    apiBaseUrl: 'http://localhost:11434',
    apiKey: '',
    status: 'connected',
    models: [
      { id: 'llama3.1', name: 'Llama 3.1', providerId: 'ollama', maxTokens: 8192, defaultTemperature: 0.7 },
      { id: 'qwen2.5', name: 'Qwen 2.5', providerId: 'ollama', maxTokens: 8192, defaultTemperature: 0.7 },
    ],
  },
]

export const mockMessages: ChatMessage[] = [
  {
    id: 'msg-1',
    conversationId: 'conv-1',
    role: 'user',
    content: '帮我训练一个YOLO车辆检测模型，数据集是交通摄像头拍摄的城市道路画面。',
    createdAt: Date.now() - 300000,
  },
  {
    id: 'msg-2',
    conversationId: 'conv-1',
    role: 'assistant',
    content: `好的，我来帮你训练一个 YOLO 车辆检测模型。让我先分析一下需求：

## 任务分析

- **目标**：车辆检测（Car, Truck, Bus, Motorcycle）
- **数据源**：城市道路交通摄像头画面
- **模型**：YOLOv8n（轻量级，适合实时检测）

## 执行计划

1. ✅ 检查环境依赖（Python、CUDA、Ultralytics）
2. ✅ 下载预训练权重
3. ▶ 准备数据集（转换为 YOLO 格式）
4. ⏳ 开始训练

我已经为你生成了一个 Workflow，包含数据预处理、模型训练、评估三个节点。`,
    model: 'Claude Sonnet 4',
    createdAt: Date.now() - 295000,
    duration: 2400,
    tokenUsage: { prompt: 850, completion: 420, total: 1270 },
    task: {
      id: 'task-1',
      title: '训练 YOLO 车辆检测模型',
      steps: [
        { id: 's1', label: '检查环境', status: 'completed' },
        { id: 's2', label: '下载模型', status: 'completed' },
        { id: 's3', label: '准备数据', status: 'running' },
        { id: 's4', label: '开始训练', status: 'pending' },
      ],
      progress: 60,
      status: 'running',
    },
  },
  {
    id: 'msg-3',
    conversationId: 'conv-1',
    role: 'user',
    content: '数据集我已经准备好了，在 `/data/traffic_detection/` 目录下，包含 images 和 labels 文件夹。',
    createdAt: Date.now() - 200000,
  },
  {
    id: 'msg-4',
    conversationId: 'conv-1',
    role: 'assistant',
    content: `收到，我已经检测到数据集目录：

\`\`\`
/data/traffic_detection/
├── images/
│   ├── train/    (800 张)
│   └── val/      (200 张)
└── labels/
    ├── train/
    └── val/
\`\`\`

数据集结构正确。我正在生成 YOLO 训练配置文件 \`data.yaml\`：

\`\`\`yaml
train: /data/traffic_detection/images/train
val: /data/traffic_detection/images/val
nc: 4
names: ['car', 'truck', 'bus', 'motorcycle']
\`\`\`

训练参数：
- **Epochs**: 100
- **Batch Size**: 16
- **Image Size**: 640
- **Learning Rate**: 0.01

训练已开始，预计耗时约 15 分钟。`,
    model: 'Claude Sonnet 4',
    createdAt: Date.now() - 195000,
    duration: 1800,
    tokenUsage: { prompt: 1200, completion: 680, total: 1880 },
  },
]

export const mockConversations: Conversation[] = [
  { id: 'conv-1', title: 'YOLO 车辆检测训练', createdAt: Date.now() - 300000, updatedAt: Date.now() - 195000, messageCount: 4, model: 'Claude Sonnet 4', isFavorite: true },
  { id: 'conv-2', title: '交通流量优化分析', createdAt: Date.now() - 86400000, updatedAt: Date.now() - 86400000, messageCount: 12, model: 'GPT-4.1', isFavorite: false },
  { id: 'conv-3', title: 'SUMO 仿真配置', createdAt: Date.now() - 172800000, updatedAt: Date.now() - 172800000, messageCount: 8, model: 'Claude Sonnet 4', isFavorite: false },
  { id: 'conv-4', title: 'LSTM 时序预测调试', createdAt: Date.now() - 259200000, updatedAt: Date.now() - 259200000, messageCount: 15, model: 'DeepSeek V3', isFavorite: true },
]

export const mockContext: ChatContext = {
  project: { id: 'proj-1', name: 'AIStudio Traffic Project' },
  workflow: { id: 'wf-1', name: 'YOLO → LSTM → SUMO', nodes: ['YOLO Detection', 'LSTM Prediction', 'SUMO Simulation'] },
  files: [
    { name: 'train.py', path: '/src/train.py' },
    { name: 'data.yaml', path: '/config/data.yaml' },
    { name: 'config.json', path: '/config/config.json' },
  ],
  plugins: [
    { name: 'YOLO Plugin', status: 'active' },
    { name: 'SUMO Connector', status: 'active' },
  ],
  mcpServers: [
    { name: 'VISSIM Connected', status: 'connected' },
    { name: 'File System', status: 'connected' },
  ],
}

export const mockQuickActions: QuickAction[] = [
  { id: 'workflow', label: '生成 Workflow', icon: 'M6 3h3v6H6V3zm0 12h3v6H6v-6zm9-12h3v6h-3V3zm0 12h3v6h-3v-6zm-9 0V9m3 9v-3m3 3v-3m3 3V9', prompt: '帮我生成一个Workflow，用于' },
  { id: 'log', label: '分析日志', icon: 'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z M14 2v6h6 M16 13H8 M16 17H8 M10 9H8', prompt: '请分析以下日志中的错误和异常：' },
  { id: 'fix', label: '修复错误', icon: 'M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z', prompt: '请帮我修复以下代码中的错误：' },
  { id: 'code', label: '生成代码', icon: 'M16 18l6-6-6-6M8 6l-6 6 6 6', prompt: '请帮我生成以下功能的代码：' },
  { id: 'train', label: '训练模型', icon: 'M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20zm0 18a8 8 0 1 1 0-16 8 8 0 0 1 0 16z', prompt: '帮我训练一个AI模型，需求如下：' },
  { id: 'mcp', label: '调用 MCP', icon: 'M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71 M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71', prompt: '请调用MCP工具执行以下操作：' },
]
