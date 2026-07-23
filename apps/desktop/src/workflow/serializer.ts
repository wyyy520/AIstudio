import type { Node, Edge } from '@vue-flow/core'
import type { WorkflowSchema, WorkflowNodeSchema, WorkflowEdgeSchema, ViewportSchema } from './schema'
import type { WorkflowNodeData, NodeCategory } from '@/pages/Workflow/types/workflow'

function inferCategory(nodeType: string): NodeCategory {
  const categoryMap: Record<string, NodeCategory> = {
    'ai-vision': 'ai-vision',
    'ai-nlp': 'ai-nlp',
    'ai-timeseries': 'ai-timeseries',
    'ai-audio': 'ai-audio',
    data: 'data',
    dataset: 'data',
    training: 'training',
    deployment: 'deployment',
    logic: 'logic',
    system: 'system',
    simulation: 'simulation',
    mcp: 'mcp',
    input: 'input',
    output: 'output',
  }
  return categoryMap[nodeType] || 'system'
}

export function serializeNodesToVueFlow(nodes: WorkflowNodeSchema[]): Node[] {
  return nodes.map(n => ({
    id: n.id,
    type: 'custom',
    position: { x: n.position.x, y: n.position.y },
    data: {
      nodeType: n.type,
      label: n.name,
      description: n.description || '',
      category: inferCategory(n.type),
      status: n.status || 'idle',
      params: n.config,
      paramDefinitions: [],
      templateKey: n.type,
      inputs: n.inputs.map(i => ({ id: i.id, name: i.name, label: i.name, type: i.type, description: i.description, required: i.required })),
      outputs: n.outputs.map(o => ({ id: o.id, name: o.name, label: o.name, type: o.type, description: o.description, required: o.required })),
      enabled: n.enabled,
      plugin: n.plugin,
      metadata: n.metadata,
    } as WorkflowNodeData,
  }))
}

export function serializeEdgesToVueFlow(edges: WorkflowEdgeSchema[]): Edge[] {
  return edges.map(e => ({
    id: e.id,
    source: e.source_node,
    target: e.target_node,
    sourceHandle: e.source_port,
    targetHandle: e.target_port,
    label: e.label || undefined,
    type: 'smoothstep',
    data: { condition: e.condition, edgeType: e.type },
  }))
}

export function deserializeNodesFromVueFlow(nodes: Node[]): WorkflowNodeSchema[] {
  const now = new Date().toISOString()
  return nodes.map(n => {
    const data = (n.data || {}) as Record<string, unknown>
    return {
      id: n.id,
      type: (data.nodeType as string) || '',
      name: (data.label as string) || '',
      description: (data.description as string) || '',
      position: {
        x: n.position.x ?? 0,
        y: n.position.y ?? 0,
      },
      config: (data.params as Record<string, unknown>) || {},
      inputs: ((data.inputs as any[]) || []).map((i: any) => ({
        id: i.id || `port_${i.name}`,
        name: i.name || '',
        type: i.type || 'any',
        description: i.description,
        required: i.required ?? true,
      })),
      outputs: ((data.outputs as any[]) || []).map((o: any) => ({
        id: o.id || `port_${o.name}`,
        name: o.name || '',
        type: o.type || 'any',
        description: o.description,
        required: o.required ?? true,
      })),
      status: (data.status as WorkflowNodeSchema['status']) || 'idle',
      enabled: (data.enabled as boolean) ?? true,
      plugin: (data.plugin as string) || undefined,
      metadata: data.metadata as Record<string, unknown> | undefined,
      created_at: now,
      updated_at: now,
    }
  })
}

export function deserializeEdgesFromVueFlow(edges: Edge[]): WorkflowEdgeSchema[] {
  return edges.map(e => {
    const edgeData = (e.data || {}) as Record<string, unknown>
    return {
      id: e.id,
      source_node: e.source,
      source_port: e.sourceHandle || '',
      target_node: e.target,
      target_port: e.targetHandle || '',
      label: typeof e.label === 'string' ? e.label : undefined,
      type: (edgeData.edgeType as WorkflowEdgeSchema['type']) || 'data',
      condition: edgeData.condition as WorkflowEdgeSchema['condition'],
      metadata: edgeData.metadata as Record<string, unknown> | undefined,
    }
  })
}

export function toVueFlowData(workflow: WorkflowSchema): { nodes: Node[]; edges: Edge[]; viewport?: ViewportSchema } {
  return {
    nodes: serializeNodesToVueFlow(workflow.nodes),
    edges: serializeEdgesToVueFlow(workflow.edges),
    viewport: workflow.viewport,
  }
}

export function fromVueFlowData(
  id: string,
  name: string,
  domain: string,
  nodes: Node[],
  edges: Edge[],
  viewport?: ViewportSchema,
): WorkflowSchema {
  const now = new Date().toISOString()
  return {
    schema_version: '2.0.0',
    id,
    name,
    description: '',
    version: 1,
    domain: domain as WorkflowSchema['domain'],
    nodes: deserializeNodesFromVueFlow(nodes),
    edges: deserializeEdgesFromVueFlow(edges),
    viewport,
    created_at: now,
    updated_at: now,
  }
}
