import type { Node, Edge } from '@vue-flow/core'

// ============================================================
// 端口类型系统
// ============================================================

export const PORT_TYPES = [
  'image', 'tensor', 'dataset', 'model', 'text', 'audio',
  'result', 'number', 'json', 'trigger', 'service', 'any',
] as const

export type PortType = typeof PORT_TYPES[number]

export const PORT_TYPE_LABELS: Record<PortType, string> = {
  image: 'Image',
  tensor: 'Tensor',
  dataset: 'Dataset',
  model: 'Model',
  text: 'Text',
  audio: 'Audio',
  result: 'Result',
  number: 'Number',
  json: 'JSON',
  trigger: 'Trigger',
  service: 'Service',
  any: 'Any',
}

export const PORT_TYPE_COLORS: Record<PortType, string> = {
  image: 'var(--vision)',
  tensor: 'var(--nlp)',
  dataset: 'var(--timeseries)',
  model: 'var(--mcp)',
  text: 'var(--info)',
  audio: 'var(--warning)',
  result: 'var(--success)',
  number: 'var(--logic)',
  json: 'var(--system)',
  trigger: 'var(--agent)',
  service: 'var(--simulation)',
  any: 'var(--neutral)',
}

// 端口连接兼容性矩阵
export const PORT_COMPATIBILITY: Record<PortType, PortType[]> = {
  image: ['image', 'tensor', 'any'],
  tensor: ['tensor', 'model', 'any'],
  dataset: ['dataset', 'any'],
  model: ['model', 'service', 'any'],
  text: ['text', 'json', 'audio', 'any'],
  audio: ['audio', 'any'],
  result: ['result', 'json', 'any'],
  number: ['number', 'json', 'any'],
  json: ['json', 'text', 'any'],
  trigger: ['trigger', 'any'],
  service: ['service', 'any'],
  any: PORT_TYPES as unknown as PortType[],
}

// ============================================================
// 参数定义
// ============================================================

export type ParamType = 'text' | 'number' | 'select' | 'switch' | 'multi-select' | 'slider' | 'device-select'

export interface ParamOption {
  label: string
  value: string | number | boolean
}

export interface ParamValidation {
  rule: string
  message: string
}

export interface ParamDefinition {
  name: string
  label: string
  type: ParamType
  default: any
  options?: ParamOption[]
  min?: number
  max?: number
  step?: number
  required?: boolean
  category?: string
  validation?: ParamValidation
  placeholder?: string
  hint?: string
}

// ============================================================
// 端口定义
// ============================================================

export interface PortDefinition {
  name: string
  label: string
  type: PortType
  description?: string
}

// ============================================================
// 节点分类
// ============================================================

export type NodeCategory =
  | 'ai-vision'
  | 'ai-nlp'
  | 'ai-timeseries'
  | 'ai-audio'
  | 'data'
  | 'training'
  | 'deployment'
  | 'logic'
  | 'system'
  | 'simulation'
  | 'mcp'
  | 'input'
  | 'output'

export const NODE_CATEGORY_LABELS: Record<NodeCategory, string> = {
  'ai-vision': 'AI 视觉',
  'ai-nlp': 'AI 自然语言',
  'ai-timeseries': 'AI 时序',
  'ai-audio': 'AI 语音',
  'data': '数据处理',
  'training': '训练',
  'deployment': '部署',
  'logic': '逻辑控制',
  'system': '系统工具',
  'simulation': '仿真',
  'mcp': 'MCP',
  'input': '输入',
  'output': '输出',
}

export const NODE_CATEGORY_COLORS: Record<NodeCategory, string> = {
  'ai-vision': 'var(--vision)',
  'ai-nlp': 'var(--nlp)',
  'ai-timeseries': 'var(--timeseries)',
  'ai-audio': 'var(--warning)',
  'data': 'var(--info)',
  'training': 'var(--success)',
  'deployment': 'var(--mcp)',
  'logic': 'var(--logic)',
  'system': 'var(--neutral)',
  'simulation': 'var(--simulation)',
  'mcp': 'var(--mcp)',
  'input': 'var(--timeseries)',
  'output': 'var(--mcp)',
}

// ============================================================
// 节点模板
// ============================================================

export interface NodeTemplate {
  key: string
  label: string
  description: string
  category: NodeCategory
  nodeType: string
  color: string
  icon: string
  inputs: PortDefinition[]
  outputs: PortDefinition[]
  params: ParamDefinition[]
  paramsLayout?: 'default' | 'collapsible'
}

// ============================================================
// 节点实例数据
// ============================================================

export interface WorkflowNodeData {
  label: string
  description: string
  nodeType: string
  category: NodeCategory
  status: NodeStatus
  inputs: PortDefinition[]
  outputs: PortDefinition[]
  params: Record<string, any>
  paramDefinitions: ParamDefinition[]
  templateKey: string
}

export type NodeStatus = 'idle' | 'running' | 'success' | 'error' | 'waiting'

export const NODE_STATUS_COLORS: Record<NodeStatus, string> = {
  idle: 'var(--neutral)',
  running: 'var(--info)',
  success: 'var(--success)',
  error: 'var(--error)',
  waiting: 'var(--warning)',
}

// ============================================================
// 校验
// ============================================================

export type ValidationSeverity = 'error' | 'warning' | 'info'

export interface ValidationError {
  type: 'missing-input' | 'param-error' | 'env-error' | 'env-warning' | 'connection-error' | 'type-mismatch'
  nodeId: string
  message: string
  severity: ValidationSeverity
  autoFix?: string
  details?: string
}

export interface ValidationResult {
  valid: boolean
  errors: ValidationError[]
  warnings: ValidationError[]
}

// ============================================================
// 工作流 JSON
// ============================================================

export interface WorkflowJSON {
  id: string
  name: string
  description: string
  version: string
  nodes: WorkflowNodeJSON[]
  edges: WorkflowEdgeJSON[]
  createdAt: string
  updatedAt: string
}

export interface WorkflowNodeJSON {
  id: string
  templateKey: string
  label: string
  type: string
  category: NodeCategory
  position: { x: number; y: number }
  params: Record<string, any>
  inputs: PortDefinition[]
  outputs: PortDefinition[]
}

export interface WorkflowEdgeJSON {
  id: string
  source: string
  target: string
  sourceHandle: string
  targetHandle: string
  sourceType: PortType
  targetType: PortType
}

// ============================================================
// JSON 导出辅助
// ============================================================

export function toWorkflowJSON(
  id: string,
  name: string,
  description: string,
  nodes: Node[],
  edges: Edge[],
): WorkflowJSON {
  return {
    id,
    name,
    description,
    version: '1.0.0',
    nodes: nodes.map(n => ({
      id: n.id,
      templateKey: (n.data as WorkflowNodeData).templateKey || '',
      label: (n.data as WorkflowNodeData).label || '',
      type: (n.data as WorkflowNodeData).nodeType || '',
      category: (n.data as WorkflowNodeData).category || 'system',
      position: n.position,
      params: (n.data as WorkflowNodeData).params || {},
      inputs: (n.data as WorkflowNodeData).inputs || [],
      outputs: (n.data as WorkflowNodeData).outputs || [],
    })),
    edges: edges.map(e => ({
      id: e.id,
      source: e.source,
      target: e.target,
      sourceHandle: e.sourceHandle || '',
      targetHandle: e.targetHandle || '',
      sourceType: (e as any).sourceType || 'any',
      targetType: (e as any).targetType || 'any',
    })),
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  }
}