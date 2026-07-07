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
      @connect="onConnect"
      @init="onInit"
      @pane-click="onPaneClick"
      :pan-on-scroll="false"
      :zoom-on-scroll="true"
      :pane-scale-bounds="{ min: 0.25, max: 2 }"
      :class="{ 'is-dragging': isDragging }"
    >
      <Background :gap="20" :color="'rgba(255, 255, 255, 0.05)'" />
      <MiniMap
        class="custom-minimap"
        :position="Position.BottomRight"
        :node-color="nodeColorResolver"
      />
      <Controls :position="Position.BottomRight" />

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
  Position,
} from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { MiniMap } from '@vue-flow/minimap'
import { Controls } from '@vue-flow/controls'
import NodeCard from './NodeCard.vue'

import '@vue-flow/core/dist/style.css'
import '@vue-flow/controls/dist/style.css'
import '@vue-flow/minimap/dist/style.css'

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
  'add-node': [template: any, position: { x: number; y: number }]
  'node-delete': [nodeId: string]
  'run-state-change': [nodeId: string, status: 'idle' | 'running' | 'success' | 'error']
}>()

const localNodes = ref<Node[]>(props.nodes || [])
const localEdges = ref<Edge[]>(props.edges || [])
const selectedNodeId = ref<string | null>(null)
const isDragging = ref(false)

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

function onPaneClick(): void {
  selectedNodeId.value = null
  emits('node-click', null)
}

function onConnect(params: Connection): void {
  const newEdge = {
    id: `e_${params.source}_${params.target}_${Date.now()}`,
    source: params.source,
    target: params.target,
    sourceHandle: params.sourceHandle,
    targetHandle: params.targetHandle,
  }
  localEdges.value = addEdge(newEdge, localEdges.value)
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
    console.error('Failed to parse dropped node', e)
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
    },
    {
      id: 'e2',
      source: 'n2',
      target: 'n3',
      sourceHandle: 'detections',
      targetHandle: 'detections',
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

/* 自定义 MiniMap */
:deep(.vue-flow__minimap) {
  background: var(--bg-secondary) !important;
  border: 1px solid var(--border-subtle) !important;
  border-radius: var(--radius-md) !important;
  right: 16px !important;
  bottom: 16px !important;
  box-shadow: var(--shadow) !important;
}

/* 自定义 Controls */
:deep(.vue-flow__controls) {
  background: var(--bg-secondary) !important;
  border: 1px solid var(--border-subtle) !important;
  border-radius: var(--radius-md) !important;
  box-shadow: var(--shadow) !important;
  right: 16px !important;
  bottom: 140px !important;
}

:deep(.vue-flow__controls button) {
  background: transparent !important;
  border-color: var(--border-default) !important;
  color: var(--text-secondary) !important;
}

:deep(.vue-flow__controls button:hover) {
  background: var(--bg-hover) !important;
  color: var(--text-primary) !important;
}

:deep(.vue-flow__edge-path) {
  stroke: var(--border-strong);
  stroke-width: 2;
}

:deep(.vue-flow__edge.selected .vue-flow__edge-path) {
  stroke: var(--primary);
  stroke-width: 3;
}

:deep(.vue-flow__edge:hover .vue-flow__edge-path) {
  stroke: var(--primary);
  stroke-width: 3;
}
</style>