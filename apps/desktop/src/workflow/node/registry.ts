import type { WorkflowNodeSchema, PortSchema, NodeStatus } from '../schema'

export interface NodeTemplate {
  type: string
  name: string
  category: string
  color: string
  icon: string
  description: string
  inputs: PortSchema[]
  outputs: PortSchema[]
  defaultConfig: Record<string, unknown>
  paramsLayout?: ParamLayout[]
}

export interface ParamLayout {
  key: string
  label: string
  type: 'text' | 'number' | 'select' | 'switch' | 'multi-select' | 'slider' | 'device-select'
  default?: unknown
  options?: Array<{ label: string; value: string }>
  min?: number
  max?: number
  step?: number
  required?: boolean
  placeholder?: string
  description?: string
}

const nodeTemplateRegistry = new Map<string, NodeTemplate>()

export function registerNodeTemplate(template: NodeTemplate) {
  nodeTemplateRegistry.set(template.type, template)
}

export function getNodeTemplate(type: string): NodeTemplate | undefined {
  return nodeTemplateRegistry.get(type)
}

export function getAllTemplates(): NodeTemplate[] {
  return Array.from(nodeTemplateRegistry.values())
}

export function getTemplatesByCategory(category: string): NodeTemplate[] {
  return getAllTemplates().filter(t => t.category === category)
}

export function getCategories(): string[] {
  const cats = new Set(getAllTemplates().map(t => t.category))
  return Array.from(cats)
}

export function createNodeFromTemplate(template: NodeTemplate, position: { x: number; y: number }): WorkflowNodeSchema {
  const now = new Date().toISOString()
  return {
    id: crypto.randomUUID(),
    type: template.type,
    name: template.name,
    description: template.description,
    position,
    config: { ...template.defaultConfig },
    inputs: template.inputs.map(p => ({ ...p })),
    outputs: template.outputs.map(p => ({ ...p })),
    status: 'idle',
    enabled: true,
    created_at: now,
    updated_at: now,
  }
}
