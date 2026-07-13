<template>
  <div
    class="flow-canvas"
    @dragover.prevent
    @drop="onDrop"
  >
    <VueFlow
      v-model:nodes="localNodes"
      v-model:edges="localEdges"
      :node-types="nodeTypes"
      :fit-view-on-init="true"
      @node-click="onNodeClick"
      @node-double-click="onNodeDoubleClick"
      @connect="onConnect"
      @connect-start="onConnectStart"
      @connect-end="onConnectEnd"
      @init="onInit"
      @pane-click="onPaneClick"
      @selection-change="onSelectionChange"
      @nodes-delete="onNodesDelete"
      :pan-on-scroll="false"
      :zoom-on-scroll="true"
      :pane-scale-bounds="{ min: 0.25, max: 2 }"
      :class="{ 'is-dragging': isDragging }"
      :default-edge-options="{
        type: 'step',
        style: { stroke: '#8b5cf6', strokeWidth: 2.5 },
        markerEnd: 'none',
        animated: false
      }"
      delete-key="Backspace"
    >
      <Background :gap="20" :color="'rgba(255, 255, 255, 0.05)'" />

      <template #node-custom="nodeProps">
        <NodeCard
          :data="nodeProps.data"
          :id="nodeProps.id"
          :selected="nodeProps.selected"
        />
      </template>
    </VueFlow>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import {
  VueFlow,
  useVueFlow,
  type Connection,
  type Node,
  type Edge,
  addEdge,
} from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import NodeCard from './NodeCard.vue'
import { PORT_COMPATIBILITY, type PortType } from '../types/workflow'
import type { WorkflowNodeData, ValidationError } from '../types/workflow'

import '@vue-flow/core/dist/style.css'

const nodeTypes = {
  custom: NodeCard,
}

const props = defineProps<{
  workflowId?: string
  nodes?: Node[]
  edges?: Edge[]
}>()

const emits = defineEmits<{
  'update:nodes': [nodes: Node[]]
  'update:edges': [edges: Edge[]]
  'node-click': [node: Node | null]
  'node-double-click': [node: Node]
  'add-node': [template: any, position: { x: number; y: number }]
  'node-delete': [nodeId: string]
  'nodes-delete': [nodeIds: string[]]
  'run-state-change': [nodeId: string, status: 'idle' | 'running' | 'success' | 'error']
  'validation-error': [error: ValidationError]
}>()

const localNodes = ref<Node[]>(props.nodes || [])
const localEdges = ref<Edge[]>(props.edges || [])
const selectedNodeId = ref<string | null>(null)
const isDragging = ref(false)
const isConnecting = ref(false)

const { fitView, zoomIn, zoomOut, screenToFlowPosition } = useVueFlow()

watch(() => props.nodes, (val) => {
  if (val) localNodes.value = val
}, { deep: true })

watch(() => props.edges, (val) => {
  if (val) localEdges.value = val
}, { deep: true })

watch(localNodes, (val) => {
  emits('update:nodes', val)
}, { deep: true })

watch(localEdges, (val) => {
  emits('update:edges', val)
}, { deep: true })

const colorMap: Record<string, string> = {
  dataset: 'var(--timeseries)',
  vision: 'var(--vision)',
  nlp: 'var(--nlp)',
  speech: 'var(--info)',
  timeseries: 'var(--timeseries)',
  logic: 'var(--logic)',
  system: 'var(--system)',
  simulation: 'var(--simulation)',
  mcp: 'var(--mcp)',
  agent: 'var(--agent)',
  input: 'var(--timeseries)',
  output: 'var(--mcp)',
  default: 'var(--neutral)',
}

function nodeColorResolver(node: Node): string {
  const nodeType = (node.data as any)?.nodeType || 'default'
  return colorMap[nodeType] || colorMap.default
}

function onInit(): void {
  if (localNodes.value.length === 0) {
    initSampleWorkflow()
    setTimeout(() => fitView({ padding: 0.2 }), 100)
  }
}

function onNodeClick({ node }: { node: Node }): void {
  selectedNodeId.value = node.id
  emits('node-click', node)
}

function onNodeDoubleClick({ node }: { node: Node }): void {
  emits('node-double-click', node)
}

function onPaneClick(): void {
  selectedNodeId.value = null
  emits('node-click', null)
}

function onConnectStart(): void {
  isConnecting.value = true
}

function onConnectEnd(): void {
  isConnecting.value = false
}

function onConnect(params: Connection): void {
  // 校验连接类型是否匹配
  const sourceNode = localNodes.value.find(n => n.id === params.source)
  const targetNode = localNodes.value.find(n => n.id === params.target)

  if (!sourceNode) return
  if (!targetNode) return

  const sourceData = sourceNode.data as WorkflowNodeData
  const targetData = targetNode.data as WorkflowNodeData

  // 当 Vue Flow 发出 null handle 时，自动选择唯一的端口
  const resolvedSourceHandle = params.sourceHandle ?? (sourceData.outputs?.length === 1 ? sourceData.outputs[0].name : null)
  const resolvedTargetHandle = params.targetHandle ?? (targetData.inputs?.length === 1 ? targetData.inputs[0].name : null)

  if (!resolvedSourceHandle || !resolvedTargetHandle) return

  const sourcePort = sourceData.outputs?.find((o: any) => o.name === resolvedSourceHandle)
  const targetPort = targetData.inputs?.find((i: any) => i.name === resolvedTargetHandle)

  if (sourcePort && targetPort) {
    const sourceType = sourcePort.type as PortType
    const targetType = targetPort.type as PortType

    const compatible = PORT_COMPATIBILITY[sourceType]
    if (!compatible || !compatible.includes(targetType)) {
      emits('validation-error', {
        type: 'type-mismatch',
        nodeId: params.target,
        message: `数据类型不匹配: "${sourceData.label}" 输出 ${sourceType} 不能连接到 "${targetData.label}" 输入 ${targetType}`,
        severity: 'error',
        autoFix: `断开此连接，将 ${sourceType} 类型端口连接到兼容的 ${targetType} 输入`,
        details: `${sourceType} → ${targetType} 不在兼容列表中`,
      })
      return
    }
  }

  // 检查是否已存在相同连接
  const exists = localEdges.value.some(
    e => e.source === params.source &&
         e.target === params.target &&
         (e.sourceHandle ?? null) === resolvedSourceHandle &&
         (e.targetHandle ?? null) === resolvedTargetHandle
  )
  if (exists) return

  // 检查目标端口是否已被连接
  const targetPortConnected = localEdges.value.some(
    e => e.target === params.target && (e.targetHandle ?? null) === resolvedTargetHandle
  )
  if (targetPortConnected) {
    emits('validation-error', {
      type: 'connection-error',
      nodeId: params.target,
      message: `"${targetData.label}" 的 "${targetPort?.label || resolvedTargetHandle}" 端口已连接`,
      severity: 'warning',
      autoFix: `请先断开 "${targetData.label}" 的现有连接`,
      details: `一个输入端口只能连接一个输出`,
    })
    return
  }

  const newEdge: Edge = {
    id: `e_${params.source}_${params.target}_${Date.now()}`,
    source: params.source,
    target: params.target,
    sourceHandle: resolvedSourceHandle,
    targetHandle: resolvedTargetHandle,
    type: 'step',
  }
  localEdges.value = addEdge(newEdge, localEdges.value)
}

function onSelectionChange({ nodes, edges }: { nodes: Node[]; edges: Edge[] }): void {
  if (nodes.length === 1) {
    selectedNodeId.value = nodes[0].id
    emits('node-click', nodes[0])
  } else if (nodes.length === 0) {
    selectedNodeId.value = null
    emits('node-click', null)
  }
}

function onNodesDelete(deletedNodes: Node[]): void {
  const ids = deletedNodes.map(n => n.id)
  if (ids.length === 1) {
    emits('node-delete', ids[0])
  } else {
    emits('nodes-delete', ids)
  }
}

function onDrop(event: DragEvent): void {
  const data = event.dataTransfer?.getData('application/json')
  if (!data) return

  try {
    const nodeTemplate = JSON.parse(data)
    const position = screenToFlowPosition({
      x: event.clientX,
      y: event.clientY,
    })
    if (position) {
      emits('add-node', nodeTemplate, position)
    }
  } catch (e) {
    // Failed to parse dropped node data - silently ignore invalid drops
  }
}

function getViewportCenter(): { x: number; y: number } {
  const el = document.querySelector('.vue-flow__viewport') as HTMLElement
  if (!el) return { x: 200, y: 200 }
  const rect = el.getBoundingClientRect()
  const transform = new DOMMatrixReadOnly(el.style.transform)
  const cx = (-transform.e + rect.width / 2) / transform.a
  const cy = (-transform.f + rect.height / 2) / transform.d
  return { x: cx, y: cy }
}

function initSampleWorkflow(): void {
  localNodes.value = [
    {
      id: 'n1',
      type: 'custom',
      position: { x: 100, y: 150 },
      data: {
        label: '图像输入',
        description: '加载输入图像',
        nodeType: 'input',
        status: 'idle',
        inputs: [],
        outputs: [{ name: 'image', label: '图像', type: 'image' }],
      },
    },
    {
      id: 'n2',
      type: 'custom',
      position: { x: 400, y: 150 },
      data: {
        label: 'YOLO 检测',
        description: '目标检测',
        nodeType: 'vision',
        status: 'idle',
        inputs: [{ name: 'image', label: '图像', type: 'image' }],
        outputs: [{ name: 'detections', label: '检测结果', type: 'json' }],
      },
    },
    {
      id: 'n3',
      type: 'custom',
      position: { x: 700, y: 150 },
      data: {
        label: '输出结果',
        description: '保存检测结果',
        nodeType: 'output',
        status: 'idle',
        inputs: [{ name: 'detections', label: '结果', type: 'json' }],
        outputs: [],
      },
    },
  ]

  localEdges.value = [
    {
      id: 'e1',
      source: 'n1',
      target: 'n2',
      sourceHandle: 'image',
      targetHandle: 'image',
      type: 'step',
    },
    {
      id: 'e2',
      source: 'n2',
      target: 'n3',
      sourceHandle: 'detections',
      targetHandle: 'detections',
      type: 'step',
    },
  ]
}

defineExpose({
  fitView,
  zoomIn,
  zoomOut,
  getViewportCenter,
})
</script>

<style scoped>
.flow-canvas {
  position: relative;
  width: 100%;
  height: 100%;
  background: var(--bg-primary);
  overflow: hidden;
}

/* ---- Simulink 风格直角折线 ---- */
:deep(.vue-flow__edge-path) {
  fill: none;
  stroke: #8b5cf6;
  stroke-width: 2.5;
  stroke-linejoin: round;
  stroke-linecap: round;
  transition: stroke-width 0.12s ease, stroke 0.12s ease;
}

:deep(.vue-flow__edge.selected .vue-flow__edge-path) {
  stroke: #ef4444;
  stroke-width: 4;
  filter: drop-shadow(0 0 4px rgba(239, 68, 68, 0.4));
}

:deep(.vue-flow__edge:hover .vue-flow__edge-path) {
  stroke: #a78bfa;
  stroke-width: 4;
}

/* ---- 移除动画，Simulink 风格纯色静止 ---- */
:deep(.vue-flow__edge.animated path) {
  animation: none !important;
  stroke-dasharray: none !important;
}

/* ---- 选中路径状态指示 ---- */
:deep(.vue-flow__edge.selected) {
  z-index: 10;
}

/* ---- 节点层级（比连线高） ---- */
:deep(.vue-flow__node) {
  z-index: 20;
}

/* ---- 画布背景 ---- */
:deep(.vue-flow__background pattern path) {
  stroke: rgba(255, 255, 255, 0.06);
}

/* ---- 去掉任何多余装饰 ---- */
:deep(.vue-flow__minimap),
:deep(.vue-flow__controls) {
  display: none !important;
}
</style>