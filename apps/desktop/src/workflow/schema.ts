export const CURRENT_SCHEMA_VERSION = '2.0.0'

export type Domain = 'python' | 'matlab' | 'stm32' | 'cpp' | 'java' | 'ros2' | 'unity' | 'docker' | 'solidworks' | 'ansys'

export interface WorkflowSchema {
  schema_version: string
  id: string
  name: string
  description?: string
  version: number
  domain: Domain
  target?: string
  author?: string
  tags?: string[]
  metadata?: Record<string, unknown>
  variables?: Record<string, unknown>
  plugins?: Record<string, PluginConfig>
  nodes: WorkflowNodeSchema[]
  edges: WorkflowEdgeSchema[]
  viewport?: ViewportSchema
  created_at: string
  updated_at: string
}

export interface PluginConfig {
  id: string
  name: string
  version: string
  enabled: boolean
  config?: Record<string, unknown>
}

export interface WorkflowNodeSchema {
  id: string
  type: string
  name: string
  description?: string
  position: PositionSchema
  size?: SizeSchema
  config: Record<string, unknown>
  inputs: PortSchema[]
  outputs: PortSchema[]
  status?: NodeStatus
  enabled: boolean
  plugin?: string
  metadata?: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface WorkflowEdgeSchema {
  id: string
  source_node: string
  source_port: string
  target_node: string
  target_port: string
  label?: string
  condition?: EdgeCondition
  type: EdgeType
  metadata?: Record<string, unknown>
}

export interface EdgeCondition {
  expression: string
  true_label?: string
  false_label?: string
}

export type EdgeType = 'data' | 'control' | 'condition'

export interface PositionSchema {
  x: number
  y: number
}

export interface SizeSchema {
  width: number
  height: number
}

export interface PortSchema {
  id: string
  name: string
  type: string
  description?: string
  required: boolean
}

export type NodeStatus = 'idle' | 'running' | 'success' | 'error' | 'waiting'

export interface ViewportSchema {
  x: number
  y: number
  zoom: number
}

export function createDefaultWorkflow(name: string, domain: Domain = 'python'): WorkflowSchema {
  const now = new Date().toISOString()
  return {
    schema_version: CURRENT_SCHEMA_VERSION,
    id: crypto.randomUUID(),
    name,
    description: '',
    version: 1,
    domain,
    nodes: [],
    edges: [],
    viewport: { x: 0, y: 0, zoom: 1 },
    created_at: now,
    updated_at: now,
  }
}
