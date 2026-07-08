<template>
  <div class="workflow-editor" :class="{ 'fullscreen': isFullscreen }">
    <WorkflowToolbar
      v-show="!isFullscreen"
      :workflow-name="workflowName"
      :validation-errors="validationErrors"
      @save="handleSave"
      @run="handleRun"
      @pause="handlePause"
      @stop="handleStop"
      @zoom-in="handleZoomIn"
      @zoom-out="handleZoomOut"
      @fit-view="handleFitView"
      @toggle-fullscreen="handleToggleFullscreen"
      @validate="handleValidate"
      @ai-fix="handleAiFix"
      @export-json="handleExportJSON"
    />
    <div class="editor-body">
      <NodePanel
        v-show="!isFullscreen"
        @add-node="handleAddNode"
      />
      <FlowCanvas
        ref="flowCanvasRef"
        :workflow-id="workflowId"
        v-model:nodes="nodes"
        v-model:edges="edges"
        @node-click="handleNodeClick"
        @add-node="handleAddNodeFromDrop"
        @node-delete="handleDeleteNode"
        @run-state-change="handleRunStateChange"
        @validation-error="handleConnectionError"
      />
      <PropertyPanel
        v-show="!isFullscreen"
        :selected-node="selectedNode"
        :total-nodes="nodes.length"
        :total-edges="edges.length"
        @update-node-label="handleUpdateNodeLabel"
        @update-node-params="handleUpdateNodeParams"
        @delete-node="handleDeleteNode"
      />
    </div>
    <WorkflowConsole
      v-show="!isFullscreen"
      :validation-errors="validationErrors"
      :ai-suggestions="aiSuggestions"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, shallowRef, onMounted } from 'vue'
import type { Node, Edge } from '@vue-flow/core'
import { useWorkflowStore } from '@/store/workflow'
import type { WorkflowNode, WorkflowEdge, WorkflowDefinition } from '@/store/workflow'
import WorkflowToolbar from './WorkflowToolbar.vue'
import NodePanel from './NodePanel.vue'
import FlowCanvas from './FlowCanvas.vue'
import PropertyPanel from './PropertyPanel.vue'
import WorkflowConsole from './WorkflowConsole.vue'
import { nodeTemplates } from '../config/nodeTemplates'
import {
  validateWorkflow,
  generateAiFix,
  type ValidationError,
} from '../validators/workflowValidator'
import {
  toWorkflowJSON,
  type WorkflowNodeData,
  type NodeTemplate,
} from '../types/workflow'

interface Props {
  workflowId?: string
  workflowName?: string
}

const props = withDefaults(defineProps<Props>(), {
  workflowId: undefined,
  workflowName: '未命名工作流',
})

const flowCanvasRef = ref<InstanceType<typeof FlowCanvas>>()
const workflowStore = useWorkflowStore()
const workflowId = ref(props.workflowId)
const workflowName = ref(props.workflowName)
const selectedNode = shallowRef<Node | null>(null)
const nodes = ref<Node[]>([])
const edges = ref<Edge[]>([])
const isFullscreen = ref(false)
const isRunning = ref(false)
const validationErrors = ref<ValidationError[]>([])
const aiSuggestions = ref<{ message: string; actions: string[] }[]>([])

let nodeIdCounter = 0

function generateNodeId(): string {
  return `node_${Date.now()}_${++nodeIdCounter}`
}

function buildNodeData(template: NodeTemplate): WorkflowNodeData {
  const initialParams: Record<string, any> = {}
  for (const param of template.params) {
    initialParams[param.name] = param.default
  }
  return {
    label: template.label,
    description: template.description,
    nodeType: template.nodeType,
    category: template.category,
    templateKey: template.key,
    status: 'idle',
    inputs: template.inputs.map(i => ({ ...i })),
    outputs: template.outputs.map(o => ({ ...o })),
    params: initialParams,
    paramDefinitions: template.params.map(p => ({
      name: p.name,
      label: p.label,
      type: p.type,
      default: p.default,
      options: p.options,
      min: p.min,
      max: p.max,
      step: p.step,
      required: p.required,
      category: p.category,
      validation: p.validation,
      placeholder: p.placeholder,
      hint: p.hint,
    })),
  }
}

function handleNodeClick(node: Node | null): void {
  selectedNode.value = node
}

function handleAddNode(template: any): void {
  if (!flowCanvasRef.value) return
  const center = flowCanvasRef.value.getViewportCenter()
  const nodeData = buildNodeData(template)
  const newNode: Node = {
    id: generateNodeId(),
    type: 'custom',
    position: { x: center.x - 100, y: center.y - 60 },
    data: nodeData,
  }
  nodes.value = [...nodes.value, newNode]
}

function handleAddNodeFromDrop(template: any, position: { x: number; y: number }): void {
  const nodeData = buildNodeData(template)
  const newNode: Node = {
    id: generateNodeId(),
    type: 'custom',
    position,
    data: nodeData,
  }
  nodes.value = [...nodes.value, newNode]
}

function handleUpdateNodeLabel(label: string): void {
  if (selectedNode.value) {
    selectedNode.value.data.label = label
  }
}

function handleUpdateNodeParams(params: Record<string, any>): void {
  if (selectedNode.value) {
    selectedNode.value.data.params = { ...params }
  }
}

function handleDeleteNode(): void {
  if (selectedNode.value) {
    const id = selectedNode.value.id
    nodes.value = nodes.value.filter(n => n.id !== id)
    edges.value = edges.value.filter(e => e.source !== id && e.target !== id)
    selectedNode.value = null
  }
}

function handleRunStateChange(nodeId: string, status: 'idle' | 'running' | 'success' | 'error'): void {
  const node = nodes.value.find(n => n.id === nodeId)
  if (node) {
    node.data.status = status
  }
}

function handleConnectionError(error: ValidationError): void {
  validationErrors.value = [...validationErrors.value, error]
  setTimeout(() => {
    validationErrors.value = validationErrors.value.filter(e => e !== error)
  }, 5000)
}

// ===== 工作流操作 =====

function toStoreDefinition(): WorkflowDefinition {
  const workflowNodes: WorkflowNode[] = nodes.value.map(n => ({
    id: n.id,
    type: n.data?.nodeType || 'unknown',
    name: n.data?.label || n.id,
    plugin: n.data?.templateKey || 'unknown',
    description: n.data?.description || '',
    inputs: (n.data?.inputs || []).map((i: { name: string; type: string; required: boolean }) => ({
      name: i.name,
      type: i.type,
      required: i.required,
    })),
    outputs: (n.data?.outputs || []).map((o: { name: string; type: string }) => ({
      name: o.name,
      type: o.type,
    })),
    x: n.position.x,
    y: n.position.y,
    config: n.data?.params || {},
  }))

  const workflowEdges: WorkflowEdge[] = edges.value.map(e => ({
    id: e.id,
    source: e.source,
    target: e.target,
    sourceHandle: e.sourceHandle || '',
    targetHandle: e.targetHandle || '',
  }))

  return { nodes: workflowNodes, edges: workflowEdges }
}

async function handleSave(): Promise<void> {
  const definition = toStoreDefinition()
  const json = toWorkflowJSON(
    workflowId.value || 'workflow_1',
    workflowName.value,
    '',
    nodes.value,
    edges.value,
  )

  if (workflowId.value) {
    const success = await workflowStore.saveWorkflow(workflowId.value, definition)
    if (!success) {
      validationErrors.value = [{ nodeId: '', message: workflowStore.error || '保存失败' }]
    }
  } else {
    console.log('Save workflow (no id):', json)
    localStorage.setItem(`workflow_draft`, JSON.stringify(json))
  }
}

function handleExportJSON(): void {
  const json = toWorkflowJSON(
    workflowId.value || 'workflow_1',
    workflowName.value,
    '',
    nodes.value,
    edges.value,
  )
  const blob = new Blob([JSON.stringify(json, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${workflowName.value}.json`
  a.click()
  URL.revokeObjectURL(url)
}

function handleValidate(): void {
  validationErrors.value = validateWorkflow(nodes.value, edges.value)
}

async function handleRun(): Promise<void> {
  // 先校验
  const errors = validateWorkflow(nodes.value, edges.value)
  if (errors.length > 0) {
    validationErrors.value = errors
    return
  }

  if (!nodes.value.length) return

  // Save first, then run
  if (workflowId.value) {
    await handleSave()
    isRunning.value = true
    validationErrors.value = []

    const taskId = await workflowStore.runWorkflow(workflowId.value)
    if (!taskId) {
      validationErrors.value = [{ nodeId: '', message: workflowStore.error || '运行失败' }]
      isRunning.value = false
    }
    // isRunning will be set to false by WebSocket event or polling
  } else {
    // Mock for new/unsaved workflows
    isRunning.value = true
    validationErrors.value = []

    const sortedIds = topologicalSort(nodes.value, edges.value)
    for (const nodeId of sortedIds) {
      const node = nodes.value.find(n => n.id === nodeId)
      if (!node) continue
      node.data.status = 'running'
      await new Promise(resolve => setTimeout(resolve, 600))
      node.data.status = 'success'
      await new Promise(resolve => setTimeout(resolve, 150))
    }
    isRunning.value = false
  }
}

function handlePause(): void {
  isRunning.value = false
}

function handleStop(): void {
  isRunning.value = false
  nodes.value.forEach(node => {
    node.data.status = 'idle'
  })
}

function handleAiFix(): void {
  const errors = validateWorkflow(nodes.value, edges.value)
  validationErrors.value = errors
  aiSuggestions.value = generateAiFix(errors, nodes.value, edges.value)
}

function handleZoomIn(): void {
  flowCanvasRef.value?.zoomIn()
}

function handleZoomOut(): void {
  flowCanvasRef.value?.zoomOut()
}

function handleFitView(): void {
  flowCanvasRef.value?.fitView()
}

function handleToggleFullscreen(): void {
  isFullscreen.value = !isFullscreen.value
}

function topologicalSort(allNodes: Node[], allEdges: Edge[]): string[] {
  const inDegree: Record<string, number> = {}
  const adj: Record<string, string[]> = {}

  for (const n of allNodes) {
    inDegree[n.id] = 0
    adj[n.id] = []
  }
  for (const e of allEdges) {
    if (adj[e.source]) adj[e.source].push(e.target)
    if (inDegree[e.target] !== undefined) inDegree[e.target]++
  }

  const queue: string[] = []
  for (const n of allNodes) {
    if (inDegree[n.id] === 0) queue.push(n.id)
  }

  const result: string[] = []
  while (queue.length > 0) {
    const id = queue.shift()!
    result.push(id)
    for (const next of adj[id] || []) {
      inDegree[next]--
      if (inDegree[next] === 0) queue.push(next)
    }
  }
  return result
}

defineExpose({
  nodes,
  edges,
})
</script>

<style scoped>
.workflow-editor {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  overflow: hidden;
  background: var(--bg-primary);
}

.editor-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}
</style>