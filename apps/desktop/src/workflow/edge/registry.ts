import type { WorkflowEdgeSchema, EdgeType, EdgeCondition } from '../schema'

const PORT_COMPATIBILITY: Record<string, string[]> = {
  image: ['image', 'tensor'],
  tensor: ['tensor', 'model'],
  dataset: ['dataset'],
  model: ['model'],
  text: ['text', 'json'],
  number: ['number', 'json'],
  audio: ['audio'],
  result: ['result'],
  json: ['json', 'text'],
  trigger: ['trigger', 'any'],
  service: ['service'],
  any: ['any'],
}

export function isPortTypeCompatible(sourceType: string, targetType: string): boolean {
  if (sourceType === 'any' || targetType === 'any' || sourceType === targetType) return true
  const compatible = PORT_COMPATIBILITY[sourceType]
  return compatible ? compatible.includes(targetType) : false
}

export function createEdge(
  id: string,
  sourceNode: string,
  sourcePort: string,
  targetNode: string,
  targetPort: string,
  type: EdgeType = 'data',
  label?: string,
): WorkflowEdgeSchema {
  return {
    id,
    source_node: sourceNode,
    source_port: sourcePort,
    target_node: targetNode,
    target_port: targetPort,
    type,
    label,
  }
}
