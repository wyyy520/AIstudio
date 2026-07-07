import type {
  Task,
  LogEntry,
  ErrorAnalysis,
  TrainingMetrics,
  WorkflowTimeline,
  FixStep,
} from './types'

const now = new Date()
const t = (offsetMs: number) => new Date(now.getTime() - offsetMs).toISOString()

export const mockTasks: Task[] = [
  {
    id: 'task-001',
    name: 'YOLO Training',
    type: 'training',
    status: 'failed',
    startedAt: t(375000),
    completedAt: t(0),
    duration: 375,
    projectId: 'proj-001',
    workflowId: 'wf-001',
    metadata: { model: 'YOLOv8', dataset: 'coco-2024', epochs: 100, currentEpoch: 45 },
  },
  {
    id: 'task-002',
    name: 'Smart Traffic Prediction',
    type: 'training',
    status: 'running',
    startedAt: t(435000),
    duration: 435,
    projectId: 'proj-001',
    workflowId: 'wf-001',
    metadata: { model: 'LSTM', dataset: 'traffic-flow', epochs: 200, currentEpoch: 128 },
  },
  {
    id: 'task-003',
    name: 'Model Export',
    type: 'export',
    status: 'completed',
    startedAt: t(90000),
    completedAt: t(0),
    duration: 90,
    projectId: 'proj-001',
  },
  {
    id: 'task-004',
    name: 'SUMO Simulation',
    type: 'simulation',
    status: 'running',
    startedAt: t(600000),
    duration: 600,
    projectId: 'proj-002',
    workflowId: 'wf-002',
  },
  {
    id: 'task-005',
    name: 'Data Preprocessing',
    type: 'workflow',
    status: 'completed',
    startedAt: t(180000),
    completedAt: t(0),
    duration: 90,
    projectId: 'proj-001',
    workflowId: 'wf-001',
  },
  {
    id: 'task-006',
    name: 'Environment Check',
    type: 'system',
    status: 'warning',
    startedAt: t(30000),
    completedAt: t(0),
    duration: 30,
    projectId: 'proj-001',
  },
  {
    id: 'task-007',
    name: 'Agent Task: Auto Label',
    type: 'agent',
    status: 'completed',
    startedAt: t(120000),
    completedAt: t(0),
    duration: 120,
    projectId: 'proj-002',
  },
]

export const mockLogs: Record<string, LogEntry[]> = {
  'task-001': [
    { id: 'log-001', taskId: 'task-001', timestamp: t(375000), level: 'info', source: 'system', message: 'Checking Python environment...', rawMessage: '2026-07-07 10:32:00 [INFO] Checking Python environment...', humanMessage: '正在检查 Python 环境', stepName: 'Environment Check', stepStatus: 'completed' },
    { id: 'log-002', taskId: 'task-001', timestamp: t(374000), level: 'info', source: 'system', message: 'Python 3.11.2 found', rawMessage: '2026-07-07 10:32:01 [INFO] Python 3.11.2 found at /usr/bin/python3', humanMessage: 'Python 3.11.2 已就绪', stepName: 'Environment Check', stepStatus: 'completed' },
    { id: 'log-003', taskId: 'task-001', timestamp: t(370000), level: 'info', source: 'system', message: 'Installing PyTorch 2.1.0...', rawMessage: '2026-07-07 10:32:05 [INFO] Installing PyTorch 2.1.0 with CUDA 12.1 support...', humanMessage: '正在安装 PyTorch 2.1.0', stepName: 'Install Dependencies', stepStatus: 'completed' },
    { id: 'log-004', taskId: 'task-001', timestamp: t(250000), level: 'info', source: 'system', message: 'PyTorch installed successfully', rawMessage: '2026-07-07 10:33:15 [INFO] PyTorch 2.1.0+cu121 installed successfully', humanMessage: 'PyTorch 安装完成', stepName: 'Install Dependencies', stepStatus: 'completed' },
    { id: 'log-005', taskId: 'task-001', timestamp: t(240000), level: 'info', source: 'training', message: 'Downloading YOLOv8 model weights...', rawMessage: '2026-07-07 10:35:00 [INFO] Downloading YOLOv8n weights from ultralytics hub...', humanMessage: '正在下载 YOLOv8 模型权重', stepName: 'Download Model', stepStatus: 'completed' },
    { id: 'log-006', taskId: 'task-001', timestamp: t(180000), level: 'info', source: 'training', message: 'Model weights downloaded', rawMessage: '2026-07-07 10:35:22 [INFO] YOLOv8n weights downloaded (6.2 MB)', humanMessage: '模型权重下载完成', stepName: 'Download Model', stepStatus: 'completed' },
    { id: 'log-007', taskId: 'task-001', timestamp: t(170000), level: 'info', source: 'training', message: 'Starting YOLO training...', rawMessage: '2026-07-07 10:36:01 [INFO] Starting YOLO training with config:\n  epochs=100, batch=16, imgsz=640, lr=0.01', humanMessage: '正在开始 YOLO 训练', stepName: 'Start Training', stepStatus: 'running' },
    { id: 'log-008', taskId: 'task-001', timestamp: t(160000), level: 'info', source: 'training', message: 'Epoch 1/100 started', rawMessage: '2026-07-07 10:36:15 [INFO] Epoch 1/100 - loss: 0.892, acc: 0.234', humanMessage: 'Epoch 1/100 开始训练', stepName: 'Start Training', stepStatus: 'running' },
    { id: 'log-009', taskId: 'task-001', timestamp: t(15000), level: 'error', source: 'training', message: 'RuntimeError: CUDA version mismatch', rawMessage: 'RuntimeError: CUDA version mismatch\nExpected CUDA >= 12.0, got 11.8\n  at train.py:45, in <module>\n    model.train()\n  File "torch/nn/module.py", line 1511, in train\n    return self._apply(lambda m: m.train(mode))\n  File "torch/cuda/__init__.py", line 210, in _lazy_init\n    torch._C._cuda_init()', humanMessage: 'CUDA 版本不匹配，训练失败', stepName: 'Start Training', stepStatus: 'failed', metadata: { file: 'train.py', line: 45, function: 'train' } },
  ],
  'task-002': [
    { id: 'log-101', taskId: 'task-002', timestamp: t(435000), level: 'info', source: 'system', message: 'Checking Python environment...', rawMessage: '2026-07-07 10:28:00 [INFO] Checking Python environment...', humanMessage: '正在检查 Python 环境', stepName: 'Environment Check', stepStatus: 'completed' },
    { id: 'log-102', taskId: 'task-002', timestamp: t(434000), level: 'info', source: 'system', message: 'Python 3.11.2 found', rawMessage: '2026-07-07 10:28:01 [INFO] Python 3.11.2 found', humanMessage: 'Python 3.11.2 已就绪', stepName: 'Environment Check', stepStatus: 'completed' },
    { id: 'log-103', taskId: 'task-002', timestamp: t(430000), level: 'info', source: 'training', message: 'Loading traffic flow dataset...', rawMessage: '2026-07-07 10:28:05 [INFO] Loading traffic flow dataset from /data/traffic_flow/', humanMessage: '正在加载交通流量数据集', stepName: 'Load Dataset', stepStatus: 'completed' },
    { id: 'log-104', taskId: 'task-002', timestamp: t(400000), level: 'info', source: 'training', message: 'Dataset loaded: 50000 samples', rawMessage: '2026-07-07 10:30:00 [INFO] Dataset loaded: 50000 samples, 4 features', humanMessage: '数据集加载完成：50000 条样本', stepName: 'Load Dataset', stepStatus: 'completed' },
    { id: 'log-105', taskId: 'task-002', timestamp: t(395000), level: 'info', source: 'training', message: 'Building LSTM model...', rawMessage: '2026-07-07 10:30:05 [INFO] Building LSTM model: layers=3, hidden=128, dropout=0.2', humanMessage: '正在构建 LSTM 模型', stepName: 'Build Model', stepStatus: 'completed' },
    { id: 'log-106', taskId: 'task-002', timestamp: t(390000), level: 'info', source: 'training', message: 'Starting training...', rawMessage: '2026-07-07 10:30:10 [INFO] Starting LSTM training: epochs=200, batch=32, lr=0.001', humanMessage: '正在开始 LSTM 训练', stepName: 'Training', stepStatus: 'running' },
    { id: 'log-107', taskId: 'task-002', timestamp: t(10000), level: 'info', source: 'training', message: 'Epoch 128/200 - loss: 0.032, accuracy: 0.921', rawMessage: '2026-07-07 10:35:30 [INFO] Epoch 128/200 - loss: 0.032, accuracy: 0.921, gpu: 45%', humanMessage: 'Epoch 128/200 训练中', stepName: 'Training', stepStatus: 'running' },
  ],
  'task-003': [
    { id: 'log-201', taskId: 'task-003', timestamp: t(90000), level: 'info', source: 'system', message: 'Loading model for export...', rawMessage: '2026-07-07 11:00:00 [INFO] Loading YOLOv8 model for export...', humanMessage: '正在加载模型', stepName: 'Load Model', stepStatus: 'completed' },
    { id: 'log-202', taskId: 'task-003', timestamp: t(80000), level: 'info', source: 'system', message: 'Model loaded successfully', rawMessage: '2026-07-07 11:00:10 [INFO] Model loaded: YOLOv8n (6.2 MB)', humanMessage: '模型加载完成', stepName: 'Load Model', stepStatus: 'completed' },
    { id: 'log-203', taskId: 'task-003', timestamp: t(75000), level: 'info', source: 'system', message: 'Exporting to ONNX format...', rawMessage: '2026-07-07 11:00:15 [INFO] Exporting to ONNX format...', humanMessage: '正在导出为 ONNX 格式', stepName: 'Export', stepStatus: 'completed' },
    { id: 'log-204', taskId: 'task-003', timestamp: t(5000), level: 'info', source: 'system', message: 'Export completed', rawMessage: '2026-07-07 11:01:30 [INFO] Export completed: model.onnx (12.4 MB)', humanMessage: '导出完成', stepName: 'Export', stepStatus: 'completed' },
  ],
  'task-004': [
    { id: 'log-301', taskId: 'task-004', timestamp: t(600000), level: 'info', source: 'workflow', message: 'Initializing SUMO environment...', rawMessage: '2026-07-07 10:20:00 [INFO] Initializing SUMO simulation environment...', humanMessage: '正在初始化 SUMO 仿真环境', stepName: 'Init Simulation', stepStatus: 'completed' },
    { id: 'log-302', taskId: 'task-004', timestamp: t(590000), level: 'info', source: 'workflow', message: 'Loading road network...', rawMessage: '2026-07-07 10:20:10 [INFO] Loading road network from map.net.xml', humanMessage: '正在加载路网', stepName: 'Load Network', stepStatus: 'completed' },
    { id: 'log-303', taskId: 'task-004', timestamp: t(580000), level: 'info', source: 'workflow', message: 'Starting simulation step 0...', rawMessage: '2026-07-07 10:20:20 [INFO] Starting simulation step 0', humanMessage: '仿真步骤 0 开始', stepName: 'Simulation', stepStatus: 'running' },
    { id: 'log-304', taskId: 'task-004', timestamp: t(10000), level: 'info', source: 'workflow', message: 'Simulation step 3600 completed', rawMessage: '2026-07-07 10:30:00 [INFO] Simulation step 3600 completed, vehicles: 342', humanMessage: '仿真步骤 3600 完成', stepName: 'Simulation', stepStatus: 'running' },
  ],
  'task-005': [
    { id: 'log-401', taskId: 'task-005', timestamp: t(180000), level: 'info', source: 'workflow', message: 'Loading raw dataset...', rawMessage: '2026-07-07 11:45:00 [INFO] Loading raw dataset from /data/raw/', humanMessage: '正在加载原始数据集', stepName: 'Load Data', stepStatus: 'completed' },
    { id: 'log-402', taskId: 'task-005', timestamp: t(160000), level: 'info', source: 'workflow', message: 'Resizing images to 640x640...', rawMessage: '2026-07-07 11:45:20 [INFO] Resizing 1000 images to 640x640...', humanMessage: '正在调整图片尺寸', stepName: 'Preprocess', stepStatus: 'completed' },
    { id: 'log-403', taskId: 'task-005', timestamp: t(90000), level: 'info', source: 'workflow', message: 'Splitting dataset: train/val 80/20', rawMessage: '2026-07-07 11:46:30 [INFO] Splitting dataset: 800 train, 200 val', humanMessage: '正在划分训练集/验证集', stepName: 'Split', stepStatus: 'completed' },
    { id: 'log-404', taskId: 'task-005', timestamp: t(0), level: 'info', source: 'workflow', message: 'Preprocessing completed', rawMessage: '2026-07-07 11:48:00 [INFO] Preprocessing completed. Output: /data/processed/', humanMessage: '数据预处理完成', stepName: 'Complete', stepStatus: 'completed' },
  ],
  'task-006': [
    { id: 'log-501', taskId: 'task-006', timestamp: t(30000), level: 'info', source: 'system', message: 'Checking system dependencies...', rawMessage: '2026-07-07 12:00:00 [INFO] Checking system dependencies...', humanMessage: '正在检查系统依赖', stepName: 'Check Dependencies', stepStatus: 'completed' },
    { id: 'log-502', taskId: 'task-006', timestamp: t(25000), level: 'warning', source: 'system', message: 'CUDA 11.8 detected, recommended: 12.1+', rawMessage: '2026-07-07 12:00:05 [WARN] CUDA 11.8 detected. Recommended version: 12.1+', humanMessage: '检测到 CUDA 11.8，建议升级到 12.1+', stepName: 'Check CUDA', stepStatus: 'completed' },
    { id: 'log-503', taskId: 'task-006', timestamp: t(20000), level: 'warning', source: 'system', message: 'Disk space low: 15GB remaining', rawMessage: '2026-07-07 12:00:10 [WARN] Disk space low: 15GB remaining on /data partition', humanMessage: '磁盘空间不足：剩余 15GB', stepName: 'Check Disk', stepStatus: 'completed' },
    { id: 'log-504', taskId: 'task-006', timestamp: t(0), level: 'info', source: 'system', message: 'Environment check completed with warnings', rawMessage: '2026-07-07 12:00:30 [INFO] Environment check completed with 2 warnings', humanMessage: '环境检查完成，发现 2 个警告', stepName: 'Complete', stepStatus: 'completed' },
  ],
  'task-007': [
    { id: 'log-601', taskId: 'task-007', timestamp: t(120000), level: 'info', source: 'agent', message: 'Agent starting auto-label task...', rawMessage: '2026-07-07 13:00:00 [INFO] Agent starting auto-label task...', humanMessage: 'Agent 开始自动标注任务', stepName: 'Agent Start', stepStatus: 'completed' },
    { id: 'log-602', taskId: 'task-007', timestamp: t(110000), level: 'info', source: 'agent', message: 'Loading unlabeled images...', rawMessage: '2026-07-07 13:00:10 [INFO] Loading 500 unlabeled images...', humanMessage: '正在加载未标注图片', stepName: 'Load Images', stepStatus: 'completed' },
    { id: 'log-603', taskId: 'task-007', timestamp: t(60000), level: 'info', source: 'agent', message: 'Labeling images with YOLO model...', rawMessage: '2026-07-07 13:01:00 [INFO] Labeling images with YOLOv8 model...', humanMessage: '正在使用 YOLO 模型标注图片', stepName: 'Auto Label', stepStatus: 'completed' },
    { id: 'log-604', taskId: 'task-007', timestamp: t(0), level: 'info', source: 'agent', message: 'Auto-label completed: 500 images labeled', rawMessage: '2026-07-07 13:02:00 [INFO] Auto-label completed: 500/500 images labeled, avg confidence: 0.87', humanMessage: '自动标注完成：500 张图片已标注', stepName: 'Complete', stepStatus: 'completed' },
  ],
}

export const mockErrorAnalyses: Record<string, ErrorAnalysis[]> = {
  'task-001': [
    {
      id: 'analysis-001',
      taskId: 'task-001',
      logEntryIds: ['log-009'],
      severity: 'critical',
      errorType: 'CUDA_ERROR',
      problem: 'YOLO 训练失败',
      cause: 'CUDA 版本不匹配',
      detail: '当前 PyTorch 2.1.0 需要 CUDA ≥12.0，系统检测到 CUDA 11.8。PyTorch 编译时使用了 CUDA 12.1，无法在 CUDA 11.8 上运行。',
      solutions: [
        { id: 'sol-001', title: '升级 CUDA', description: '安装 CUDA Toolkit 12.1', command: 'sudo apt install cuda-toolkit-12-1', estimatedTime: '~5 分钟', risk: 'low', autoFixable: true },
        { id: 'sol-002', title: '安装对应 PyTorch 版本', description: '安装兼容 CUDA 11.8 的 PyTorch 版本', command: 'pip install torch==2.0.0+cu118 --index-url https://download.pytorch.org/whl/cu118', estimatedTime: '~3 分钟', risk: 'low', autoFixable: true },
      ],
      status: 'pending',
      analyzedAt: t(0),
    },
    {
      id: 'analysis-002',
      taskId: 'task-001',
      logEntryIds: ['log-009'],
      severity: 'warning',
      errorType: 'DEPRECATION_WARNING',
      problem: 'PyTorch API 即将弃用',
      cause: '使用了即将在 PyTorch 2.2 中移除的 API',
      detail: 'torch.cuda.amp.autocast() 将在 PyTorch 2.2 中弃用，建议使用 torch.amp.autocast("cuda") 替代。',
      solutions: [
        { id: 'sol-003', title: '更新 API 调用', description: '将 torch.cuda.amp.autocast() 替换为 torch.amp.autocast("cuda")', estimatedTime: '~1 分钟', risk: 'low', autoFixable: true },
      ],
      status: 'pending',
      analyzedAt: t(0),
    },
  ],
  'task-006': [
    {
      id: 'analysis-003',
      taskId: 'task-006',
      logEntryIds: ['log-502', 'log-503'],
      severity: 'warning',
      errorType: 'ENVIRONMENT_WARNING',
      problem: '系统环境存在潜在问题',
      cause: 'CUDA 版本过低且磁盘空间不足',
      detail: 'CUDA 11.8 可能导致训练任务失败，磁盘空间 15GB 可能不足以存储大型模型和数据集。',
      solutions: [
        { id: 'sol-004', title: '升级 CUDA', description: '安装 CUDA Toolkit 12.1', command: 'sudo apt install cuda-toolkit-12-1', estimatedTime: '~5 分钟', risk: 'low', autoFixable: false },
        { id: 'sol-005', title: '清理磁盘空间', description: '删除不需要的模型缓存和临时文件', command: 'rm -rf ~/.cache/torch/hub/checkpoints/*', estimatedTime: '~1 分钟', risk: 'medium', autoFixable: true },
      ],
      status: 'pending',
      analyzedAt: t(0),
    },
  ],
}

export const mockTrainingMetrics: Record<string, TrainingMetrics> = {
  'task-001': {
    taskId: 'task-001',
    currentEpoch: 45,
    totalEpochs: 100,
    metrics: { loss: 0.034, accuracy: 0.921, learningRate: 0.001, gpuUsage: 0.85, memoryUsage: 6.2 },
    history: Array.from({ length: 45 }, (_, i) => ({
      epoch: i + 1,
      loss: Math.max(0.02, 0.9 - i * 0.02 + Math.random() * 0.01),
      accuracy: Math.min(0.95, 0.2 + i * 0.016 + Math.random() * 0.005),
      gpuUsage: 0.7 + Math.random() * 0.2,
    })),
    updatedAt: t(15000),
  },
  'task-002': {
    taskId: 'task-002',
    currentEpoch: 128,
    totalEpochs: 200,
    metrics: { loss: 0.032, accuracy: 0.921, learningRate: 0.0005, gpuUsage: 0.45, memoryUsage: 3.1 },
    history: Array.from({ length: 128 }, (_, i) => ({
      epoch: i + 1,
      loss: Math.max(0.02, 0.8 - i * 0.006 + Math.random() * 0.008),
      accuracy: Math.min(0.95, 0.15 + i * 0.006 + Math.random() * 0.004),
      gpuUsage: 0.35 + Math.random() * 0.15,
    })),
    updatedAt: t(10000),
  },
}

export const mockWorkflowTimelines: Record<string, WorkflowTimeline> = {
  'task-001': {
    taskId: 'task-001',
    workflowId: 'wf-001',
    nodes: [
      { nodeId: 'node-001', name: 'Dataset Loader', type: 'data', status: 'completed', startedAt: t(375000), completedAt: t(374000), duration: 1.2 },
      { nodeId: 'node-002', name: 'YOLO Training', type: 'training', status: 'failed', startedAt: t(373000), completedAt: t(15000), duration: 358, progress: 0.45 },
      { nodeId: 'node-003', name: 'Export Model', type: 'export', status: 'pending' },
      { nodeId: 'node-004', name: 'Upload Results', type: 'system', status: 'pending' },
    ],
  },
  'task-002': {
    taskId: 'task-002',
    workflowId: 'wf-001',
    nodes: [
      { nodeId: 'node-101', name: 'Data Loader', type: 'data', status: 'completed', startedAt: t(435000), completedAt: t(400000), duration: 35 },
      { nodeId: 'node-102', name: 'LSTM Training', type: 'training', status: 'running', startedAt: t(395000), progress: 0.64 },
      { nodeId: 'node-103', name: 'Evaluate Model', type: 'training', status: 'pending' },
    ],
  },
  'task-004': {
    taskId: 'task-004',
    workflowId: 'wf-002',
    nodes: [
      { nodeId: 'node-201', name: 'Init SUMO', type: 'system', status: 'completed', startedAt: t(600000), completedAt: t(590000), duration: 10 },
      { nodeId: 'node-202', name: 'Load Network', type: 'data', status: 'completed', startedAt: t(590000), completedAt: t(580000), duration: 10 },
      { nodeId: 'node-203', name: 'Run Simulation', type: 'simulation', status: 'running', startedAt: t(580000), progress: 0.6 },
      { nodeId: 'node-204', name: 'Collect Results', type: 'data', status: 'pending' },
    ],
  },
  'task-005': {
    taskId: 'task-005',
    workflowId: 'wf-001',
    nodes: [
      { nodeId: 'node-301', name: 'Load Data', type: 'data', status: 'completed', startedAt: t(180000), completedAt: t(160000), duration: 20 },
      { nodeId: 'node-302', name: 'Preprocess', type: 'data', status: 'completed', startedAt: t(160000), completedAt: t(90000), duration: 70 },
      { nodeId: 'node-303', name: 'Split Dataset', type: 'data', status: 'completed', startedAt: t(90000), completedAt: t(0), duration: 90 },
    ],
  },
}

export function createMockFixSteps(): FixStep[] {
  return [
    { id: 'fix-1', label: '分析错误原因', status: 'pending' },
    { id: 'fix-2', label: '修改 CUDA 配置', status: 'pending' },
    { id: 'fix-3', label: '更新依赖版本', status: 'pending' },
    { id: 'fix-4', label: '重启运行环境', status: 'pending' },
  ]
}