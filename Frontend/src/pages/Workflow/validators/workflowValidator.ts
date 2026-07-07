import type { Node, Edge } from '@vue-flow/core'
import type { WorkflowNodeData, ValidationError, PortDefinition } from '../types/workflow'
import { PORT_COMPATIBILITY, type PortType } from '../types/workflow'
import { nodeTemplates } from '../config/nodeTemplates'

// ============================================================
// 工作流校验器
// ============================================================

export { type ValidationError } from '../types/workflow'

/**
 * 校验整个工作流
 */
export function validateWorkflow(nodes: Node[], edges: Edge[]): ValidationError[] {
  const errors: ValidationError[] = []

  if (nodes.length === 0) {
    errors.push({
      type: 'missing-input',
      nodeId: 'workflow',
      message: '工作流中没有任何节点',
      severity: 'warning',
    })
    return errors
  }

  // 1. 校验每个节点的参数
  for (const node of nodes) {
    const data = node.data as WorkflowNodeData
    errors.push(...validateNodeParams(node.id, data))
  }

  // 2. 校验节点缺少输入
  for (const node of nodes) {
    const data = node.data as WorkflowNodeData
    errors.push(...validateNodeInputs(node.id, data, edges))
  }

  // 3. 校验连接类型匹配
  for (const edge of edges) {
    errors.push(...validateConnectionType(edge, nodes))
  }

  // 4. 校验环境
  for (const node of nodes) {
    const data = node.data as WorkflowNodeData
    errors.push(...validateEnvironment(node.id, data))
  }

  return errors
}

/**
 * 校验节点参数
 */
function validateNodeParams(nodeId: string, data: WorkflowNodeData): ValidationError[] {
  const errors: ValidationError[] = []
  const defs = data.paramDefinitions || []
  const params = data.params || {}

  for (const def of defs) {
    const value = params[def.name]

    // 必填项检查
    if (def.required && (value === undefined || value === '' || value === null)) {
      errors.push({
        type: 'param-error',
        nodeId,
        message: `"${data.label}" 的参数 "${def.label}" 为必填项`,
        severity: 'error',
        autoFix: `请为 "${def.label}" 设置一个值`,
        details: `参数 ${def.name} 标记为必填但当前为空`,
      })
      continue
    }

    // 自定义验证规则
    if (def.validation && value !== undefined && value !== '') {
      const { rule, message } = def.validation
      if (rule === '> 0' && parseFloat(value) <= 0) {
        errors.push({
          type: 'param-error',
          nodeId,
          message: `"${data.label}" 的 ${message}`,
          severity: 'error',
          autoFix: `将 "${def.label}" 设置为大于 0 的值`,
          details: `当前值: ${value}`,
        })
      }
    }

    // 数值范围检查
    if (def.type === 'number' && value !== undefined && value !== '') {
      const num = parseFloat(value)
      if (!isNaN(num)) {
        if (def.min !== undefined && num < def.min) {
          errors.push({
            type: 'param-error',
            nodeId,
            message: `"${data.label}" 的 "${def.label}" 不能小于 ${def.min}`,
            severity: 'error',
            autoFix: `将 "${def.label}" 设置为至少 ${def.min}`,
            details: `当前值: ${num}, 最小值: ${def.min}`,
          })
        }
        if (def.max !== undefined && num > def.max) {
          errors.push({
            type: 'param-error',
            nodeId,
            message: `"${data.label}" 的 "${def.label}" 不能大于 ${def.max}`,
            severity: 'error',
            autoFix: `将 "${def.label}" 设置为不超过 ${def.max}`,
            details: `当前值: ${num}, 最大值: ${def.max}`,
          })
        }
      }
    }
  }

  return errors
}

/**
 * 校验节点输入连接
 */
function validateNodeInputs(nodeId: string, data: WorkflowNodeData, edges: Edge[]): ValidationError[] {
  const errors: ValidationError[] = []
  const inputs = data.inputs || []

  // 检查每个输入端口是否已连接
  for (const input of inputs) {
    const hasConnection = edges.some(e => e.target === nodeId && e.targetHandle === input.name)
    if (!hasConnection) {
      // 对于有输入端口但未连接的节点，发出警告
      errors.push({
        type: 'missing-input',
        nodeId,
        message: `"${data.label}" 缺少 "${input.label}" 输入 (${input.type})`,
        severity: 'warning',
        autoFix: `请将 ${input.type} 类型的数据源连接到 "${input.label}" 端口`,
        details: `端口 ${input.name} 未连接任何数据源`,
      })
    }
  }

  return errors
}

/**
 * 校验连接类型匹配
 */
function validateConnectionType(edge: Edge, nodes: Node[]): ValidationError[] {
  const errors: ValidationError[] = []

  const sourceNode = nodes.find(n => n.id === edge.source)
  const targetNode = nodes.find(n => n.id === edge.target)
  if (!sourceNode || !targetNode) return errors

  const sourceData = sourceNode.data as WorkflowNodeData
  const targetData = targetNode.data as WorkflowNodeData

  const sourcePort = sourceData.outputs?.find(o => o.name === edge.sourceHandle)
  const targetPort = targetData.inputs?.find(i => i.name === edge.targetHandle)

  if (!sourcePort || !targetPort) return errors

  const sourceType = sourcePort.type as PortType
  const targetType = targetPort.type as PortType

  const compatible = PORT_COMPATIBILITY[sourceType]
  if (!compatible || !compatible.includes(targetType)) {
    errors.push({
      type: 'type-mismatch',
      nodeId: edge.target,
      message: `数据类型不匹配: "${sourceData.label}" 输出 ${sourceType} 不能连接到 "${targetData.label}" 输入 ${targetType}`,
      severity: 'error',
      autoFix: `断开此连接，将 ${sourceType} 类型端口连接到兼容的 ${targetType} 输入`,
      details: `${sourceType} → ${targetType} 不在兼容列表中。${sourceType} 可连接: ${compatible?.join(', ') || '无'}`,
    })
  }

  return errors
}

/**
 * 校验环境相关
 */
function validateEnvironment(nodeId: string, data: WorkflowNodeData): ValidationError[] {
  const errors: ValidationError[] = []
  const params = data.params || {}

  // 检查 CUDA 相关节点
  if (data.templateKey === 'cuda_check') {
    // Mock: 模拟 CUDA 不可用
    const mockCudaAvailable = false
    if (!mockCudaAvailable) {
      errors.push({
        type: 'env-error',
        nodeId,
        message: '检测不到 GPU 环境，将使用 CPU',
        severity: 'warning',
        autoFix: '安装 CUDA 驱动或切换到 CPU 模式',
        details: '未检测到 CUDA 设备，建议安装 CUDA 12.4 或更高版本',
      })
    }
  }

  // 检查 YOLO 训练节点的 GPU 设置
  if (data.templateKey === 'yolo_training' && params.device === 'cuda') {
    const mockCudaAvailable = false
    if (!mockCudaAvailable) {
      errors.push({
        type: 'env-error',
        nodeId,
        message: `"${data.label}" 配置了 CUDA 但检测不到 GPU 环境`,
        severity: 'warning',
        autoFix: '将设备切换为 CPU 或安装 CUDA 驱动',
        details: '当前设备配置为 CUDA GPU，但环境中未检测到可用 GPU',
      })
    }
  }

  return errors
}

/**
 * 校验单个连接（实时校验，在连接时调用）
 */
export function validateSingleConnection(
  sourceNode: Node,
  targetNode: Node,
  sourceHandle: string,
  targetHandle: string,
): ValidationError | null {
  const sourceData = sourceNode.data as WorkflowNodeData
  const targetData = targetNode.data as WorkflowNodeData

  const sourcePort = sourceData.outputs?.find(o => o.name === sourceHandle)
  const targetPort = targetData.inputs?.find(i => i.name === targetHandle)

  if (!sourcePort || !targetPort) return null

  const sourceType = sourcePort.type as PortType
  const targetType = targetPort.type as PortType

  const compatible = PORT_COMPATIBILITY[sourceType]
  if (!compatible || !compatible.includes(targetType)) {
    return {
      type: 'type-mismatch',
      nodeId: targetNode.id,
      message: `数据类型不匹配: ${sourceType} → ${targetType}`,
      severity: 'error',
      autoFix: `断开此连接`,
      details: `${sourceData.label}(${sourceType}) 不能连接到 ${targetData.label}(${targetType})`,
    }
  }

  return null
}

// ============================================================
// AI 修复建议生成器
// ============================================================

export function generateAiFix(
  errors: ValidationError[],
  nodes: Node[],
  edges: Edge[],
): { message: string; actions: string[] }[] {
  if (errors.length === 0) {
    return [{
      message: '工作流配置正常，无需修复。',
      actions: [],
    }]
  }

  const suggestions: { message: string; actions: string[] }[] = []

  // 分类错误
  const paramErrors = errors.filter(e => e.type === 'param-error')
  const missingInputs = errors.filter(e => e.type === 'missing-input')
  const typeMismatches = errors.filter(e => e.type === 'type-mismatch')
  const envErrors = errors.filter(e => e.type === 'env-error')

  if (paramErrors.length > 0) {
    suggestions.push({
      message: `检测到 ${paramErrors.length} 个参数配置问题`,
      actions: paramErrors.slice(0, 3).map(e => e.autoFix || '请检查参数配置'),
    })
  }

  if (missingInputs.length > 0) {
    suggestions.push({
      message: `检测到 ${missingInputs.length} 个节点缺少输入连接`,
      actions: [
        ...missingInputs.slice(0, 3).map(e => e.autoFix || '请连接数据源'),
        '建议从左侧节点库拖入对应数据源节点',
      ],
    })
  }

  if (typeMismatches.length > 0) {
    suggestions.push({
      message: `检测到 ${typeMismatches.length} 个数据类型不匹配`,
      actions: [
        ...typeMismatches.slice(0, 3).map(e => e.autoFix || '请重新连接'),
        '在连接前确认端口类型兼容性',
      ],
    })
  }

  if (envErrors.length > 0) {
    suggestions.push({
      message: `检测到 ${envErrors.length} 个环境配置问题`,
      actions: [
        ...envErrors.slice(0, 3).map(e => e.autoFix || '请检查环境配置'),
        '建议: 升级 CUDA 12.4',
      ],
    })
  }

  // 如果没有具体分类，提供通用建议
  if (suggestions.length === 0) {
    suggestions.push({
      message: `检测到 ${errors.length} 个问题`,
      actions: errors.slice(0, 5).map(e => e.message),
    })
  }

  return suggestions
}

/**
 * 自动修复工作流（Mock）
 */
export function autoFixWorkflow(
  errors: ValidationError[],
  nodes: Node[],
  edges: Edge[],
): { nodes: Node[]; edges: Edge[]; fixed: string[] } {
  const fixed: string[] = []
  let newNodes = [...nodes]
  let newEdges = [...edges]

  for (const error of errors) {
    switch (error.type) {
      case 'param-error': {
        // 自动修复参数错误
        const node = newNodes.find(n => n.id === error.nodeId)
        if (node) {
          const data = node.data as WorkflowNodeData
          const defs = data.paramDefinitions || []
          for (const def of defs) {
            if (error.message.includes(def.label)) {
              if (def.type === 'number' && def.min !== undefined) {
                data.params[def.name] = def.min
                node.data = { ...data }
                fixed.push(`已将 "${data.label}" 的 "${def.label}" 设置为 ${def.min}`)
              } else if (def.type === 'number' && def.default !== undefined) {
                data.params[def.name] = def.default
                node.data = { ...data }
                fixed.push(`已将 "${data.label}" 的 "${def.label}" 重置为默认值 ${def.default}`)
              }
              break
            }
          }
        }
        break
      }
      case 'type-mismatch': {
        // 移除不兼容的连接
        newEdges = newEdges.filter(e => {
          const node = newNodes.find(n => n.id === e.target)
          const srcNode = newNodes.find(n => n.id === e.source)
          if (node && srcNode && error.message.includes((srcNode.data as WorkflowNodeData).label || '')) {
            fixed.push(`已移除不兼容的连线: ${(srcNode.data as WorkflowNodeData).label} → ${(node.data as WorkflowNodeData).label}`)
            return false
          }
          return true
        })
        break
      }
      case 'env-error': {
        // 自动切换设备为 CPU
        const node = newNodes.find(n => n.id === error.nodeId)
        if (node) {
          const data = node.data as WorkflowNodeData
          if (data.params.device === 'cuda') {
            data.params.device = 'cpu'
            node.data = { ...data }
            fixed.push(`已将 "${data.label}" 的设备切换为 CPU`)
          }
        }
        break
      }
    }
  }

  return { nodes: newNodes, edges: newEdges, fixed }
}