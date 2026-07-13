<template>
  <div class="workflow-editor" :class="{ 'fullscreen': isFullscreen }">
    <!-- 运行状态提示条 -->
    <div v-if="isRunning || currentTaskId" class="run-status-bar" :class="`run-status-bar--${runStatus}`">
      <div class="run-status-left">
        <svg v-if="isRunning" class="run-spinner" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 2v4 M12 18v4 M4.93 4.93l2.83 2.83 M16.24 16.24l2.83 2.83 M2 12h4 M18 12h4 M4.93 19.07l2.83-2.83 M16.24 7.76l2.83-2.83" />
        </svg>
        <svg v-else-if="runStatus === 'success'" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14 M22 4 12 14.01l-3-3" />
        </svg>
        <svg v-else-if="runStatus === 'error'" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10" /><path d="M15 9l-6 6 M9 9l6 6" />
        </svg>
        <span class="run-status-text">{{ runStatusText }}</span>
        <span v-if="currentTaskId" class="run-task-id">Task: {{ currentTaskId }}</span>
      </div>
      <div class="run-status-right">
        <!-- 进度条 -->
        <div v-if="isRunning" class="run-progress">
          <div class="run-progress-bar">
            <div class="run-progress-fill" :style="{ width: `${progress}%` }"></div>
          </div>
          <span class="run-progress-text">{{ progress }}%</span>
        </div>
        <button v-if="!isRunning && currentTaskId" class="view-logs-btn" @click="viewLogs">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" /><path d="M14 2v6h6" />
          </svg>
          查看日志
        </button>
      </div>
    </div>

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
    <!-- 编译结果提示条：显示生成的项目文件信息 -->
    <Transition name="slide">
      <div v-if="workflowStore.compileResult" class="compile-result-bar">
        <div class="compile-result-left">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="var(--success)" stroke-width="2">
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14 M22 4 12 14.01l-3-3" />
          </svg>
          <span class="compile-result-text">项目已生成</span>
          <span class="compile-result-info">{{ workflowStore.compileResult.entryPoints?.length || 0 }} 个入口</span>
          <span class="compile-result-info">{{ workflowStore.compileResult.files?.length || 0 }} 个文件</span>
          <span class="compile-result-info">{{ workflowStore.compileResult.projectRoot }}</span>
        </div>
        <div class="compile-result-right">
          <button class="compile-result-btn" @click="showCompileFiles = !showCompileFiles">
            {{ showCompileFiles ? '隐藏文件' : '查看文件' }}
          </button>
          <button class="compile-result-close" @click="workflowStore.compileResult = null">✕</button>
        </div>
      </div>
    </Transition>
    <!-- 编译文件树 -->
    <Transition name="slide">
      <div v-if="showCompileFiles && workflowStore.compileResult" class="compile-file-tree">
        <div class="compile-file-tree-header">生成的项目文件</div>
        <div class="compile-file-tree-body">
          <div
            v-for="(file, idx) in workflowStore.compileResult.files"
            :key="idx"
            class="compile-file-item"
          >
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z M14 2v6h6" />
            </svg>
            <span class="compile-file-path">{{ file.path }}</span>
          </div>
        </div>
        <div class="compile-file-tree-footer">
          <span class="compile-entry-label">入口文件:</span>
          <span
            v-for="ep in workflowStore.compileResult.entryPoints"
            :key="ep"
            class="compile-entry-path"
          >{{ ep }}</span>
        </div>
      </div>
    </Transition>
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
    <!-- 移除底部控制台 -->
  </div>
</template>

<script setup lang="ts">
import { ref, shallowRef, watch, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import type { Node, Edge } from '@vue-flow/core'
import { useWorkflowStore } from '@/stores/workflow'
import { useTaskStore } from '@/stores/task'
import type { WorkflowNode, WorkflowEdge, WorkflowDefinition } from '@/stores/workflow'
import WorkflowToolbar from './WorkflowToolbar.vue'
import NodePanel from './NodePanel.vue'
import FlowCanvas from './FlowCanvas.vue'
import PropertyPanel from './PropertyPanel.vue'
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
import { wsClient, type WebSocketEvent } from '@/api/websocket'

interface Props {
  workflowId?: string
  workflowName?: string
}

const props = withDefaults(defineProps<Props>(), {
  workflowId: undefined,
  workflowName: '未命名工作流',
})

const router = useRouter()
const flowCanvasRef = ref<InstanceType<typeof FlowCanvas>>()
const workflowStore = useWorkflowStore()
const taskStore = useTaskStore()
const workflowId = ref(props.workflowId)
const workflowName = ref(props.workflowName)
const selectedNode = shallowRef<Node | null>(null)
const nodes = ref<Node[]>([])
const edges = ref<Edge[]>([])
const isFullscreen = ref(false)
const isRunning = ref(false)
const currentTaskId = ref<string | null>(null)
const runStatus = ref<'idle' | 'running' | 'success' | 'error'>('idle')
const progress = ref(0)
const validationErrors = ref<ValidationError[]>([])
const aiSuggestions = ref<{ message: string; actions: string[] }[]>([])
const showCompileFiles = ref(false)

let nodeIdCounter = 0
let wsUnsubscribe: (() => void) | null = null

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
    // Save to local storage as draft when no workflow ID is available
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

  // Save first, then compile, then run
  if (workflowId.value) {
    await handleSave()
    isRunning.value = true
    runStatus.value = 'running'
    progress.value = 0
    validationErrors.value = []

    // Compile then run
    const taskId = await workflowStore.compileAndRun(workflowId.value)
    if (!taskId) {
      validationErrors.value = [{ nodeId: '', message: workflowStore.error || '运行失败' }]
      isRunning.value = false
      runStatus.value = 'error'
      return
    }
    
    currentTaskId.value = taskId
    setupWebSocketForTask(taskId)
  } else {
    // Mock for new/unsaved workflows
    isRunning.value = true
    runStatus.value = 'running'
    progress.value = 0
    validationErrors.value = []
    currentTaskId.value = 'mock_' + Date.now()

    const sortedIds = topologicalSort(nodes.value, edges.value)
    const total = sortedIds.length
    let completed = 0

    for (const nodeId of sortedIds) {
      const node = nodes.value.find(n => n.id === nodeId)
      if (!node) continue
      node.data.status = 'running'
      await new Promise(resolve => setTimeout(resolve, 600))
      node.data.status = 'success'
      completed++
      progress.value = Math.round((completed / total) * 100)
      await new Promise(resolve => setTimeout(resolve, 150))
    }
    
    isRunning.value = false
    runStatus.value = 'success'
    progress.value = 100
  }
}

function setupWebSocketForTask(taskId: string): void {
  // 先取消之前的订阅
  if (wsUnsubscribe) {
    wsUnsubscribe()
    wsUnsubscribe = null
  }

  // 连接WebSocket
  wsClient.connect()
  
  wsUnsubscribe = wsClient.subscribe((event: WebSocketEvent) => {
    if (event.taskId !== taskId) return

    switch (event.type) {
      case 'task_status':
        if (event.data.status === 'running') {
          runStatus.value = 'running'
          isRunning.value = true
        }
        break
      case 'task_progress':
        if (event.data.progress !== undefined) {
          progress.value = Math.round(event.data.progress)
        }
        break
      case 'task_complete':
        isRunning.value = false
        runStatus.value = 'success'
        progress.value = 100
        nodes.value.forEach(node => {
          node.data.status = 'success'
        })
        if (wsUnsubscribe) {
          wsUnsubscribe()
          wsUnsubscribe = null
        }
        break
      case 'task_error':
        isRunning.value = false
        runStatus.value = 'error'
        if (event.data.error) {
          validationErrors.value = [{
            nodeId: '',
            message: event.data.error || '运行失败',
            severity: 'error',
          }]
        }
        nodes.value.forEach(node => {
          if (node.data.status === 'running') {
            node.data.status = 'error'
          }
        })
        if (wsUnsubscribe) {
          wsUnsubscribe()
          wsUnsubscribe = null
        }
        break
    }
  })
}

function viewLogs(): void {
  if (currentTaskId.value) {
    // 跳转到Logs页面，并传递taskId
    router.push({
      path: '/logs',
      query: { taskId: currentTaskId.value },
    })
  }
}

const runStatusText = computed(() => {
  switch (runStatus.value) {
    case 'running': return '工作流运行中...'
    case 'success': return '工作流运行成功'
    case 'error': return '工作流运行失败'
    default: return ''
  }
})

function handlePause(): void {
  isRunning.value = false
  runStatus.value = 'idle'
}

async function handleStop(): Promise<void> {
  if (currentTaskId.value) {
    try {
      const { stopProject } = await import('@/api/workflow')
      await stopProject({ runId: currentTaskId.value })
    } catch { }
  }
  isRunning.value = false
  runStatus.value = 'idle'
  progress.value = 0
  currentTaskId.value = null
  nodes.value.forEach(node => {
    node.data.status = 'idle'
  })
  if (wsUnsubscribe) {
    wsUnsubscribe()
    wsUnsubscribe = null
  }
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

// 生命周期
onMounted(() => {
  // 连接WebSocket
  wsClient.connect()
  // 加载节点类型数据
  workflowStore.fetchNodeTypes()
})

// 自动保存：节点或连线发生变化时，自动同步 workflow.json 到后端
// Workflow 永远作为唯一事实来源（Single Source of Truth）
let autoSaveTimer: ReturnType<typeof setTimeout> | null = null

watch([nodes, edges], () => {
  if (!workflowId.value) return
  if (autoSaveTimer) clearTimeout(autoSaveTimer)
  autoSaveTimer = setTimeout(async () => {
    const definition = toStoreDefinition()
    await workflowStore.saveWorkflow(workflowId.value!, definition)
  }, 2000) // 2s debounce 防抖
}, { deep: true })

onUnmounted(() => {
  if (autoSaveTimer) clearTimeout(autoSaveTimer)
  // 清理WebSocket订阅
  if (wsUnsubscribe) {
    wsUnsubscribe()
    wsUnsubscribe = null
  }
})

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

/* ===== 运行状态条 ===== */
.run-status-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 44px;
  padding: 0 24px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-subtle);
  transition: all var(--transition-fast);
}

.run-status-bar--running {
  background: rgba(59, 130, 246, 0.08);
  border-bottom-color: rgba(59, 130, 246, 0.3);
}

.run-status-bar--success {
  background: rgba(34, 197, 94, 0.08);
  border-bottom-color: rgba(34, 197, 94, 0.3);
}

.run-status-bar--error {
  background: rgba(239, 68, 68, 0.08);
  border-bottom-color: rgba(239, 68, 68, 0.3);
}

.run-status-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.run-spinner {
  animation: spin 1s linear infinite;
  color: var(--info);
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.run-status-text {
  font-size: var(--text-body-sm);
  font-weight: var(--font-medium);
  color: var(--text-primary);
}

.run-task-id {
  font-size: var(--text-caption);
  font-family: var(--font-family-mono);
  color: var(--text-tertiary);
  background: var(--bg-tertiary);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
}

.run-status-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

/* 进度条 */
.run-progress {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.run-progress-bar {
  width: 120px;
  height: 6px;
  background: var(--bg-tertiary);
  border-radius: 3px;
  overflow: hidden;
}

.run-progress-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--info), var(--primary));
  border-radius: 3px;
  transition: width 0.3s ease;
}

.run-progress-text {
  font-size: var(--text-caption);
  font-family: var(--font-family-mono);
  color: var(--text-secondary);
  min-width: 36px;
  text-align: right;
}

/* 查看日志按钮 */
.view-logs-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  height: 28px;
  padding: 0 var(--spacing-3);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  font-size: var(--text-caption);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.view-logs-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
  border-color: var(--border-strong);
}

/* ===== 编译结果提示条 ===== */
.compile-result-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 40px;
  padding: 0 20px;
  background: rgba(34, 197, 94, 0.08);
  border-bottom: 1px solid rgba(34, 197, 94, 0.2);
  flex-shrink: 0;
}

.compile-result-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.compile-result-text {
  font-size: 13px;
  font-weight: 600;
  color: var(--success);
}

.compile-result-info {
  font-size: 12px;
  color: var(--text-secondary);
  font-family: var(--font-family-mono);
}

.compile-result-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.compile-result-btn {
  height: 26px;
  padding: 0 12px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--bg-tertiary);
  color: var(--text-primary);
  font-size: 12px;
  cursor: pointer;
  transition: all var(--transition-fast);
  font-family: var(--font-family-sans);
}

.compile-result-btn:hover {
  border-color: var(--primary);
  color: var(--primary);
  background: var(--primary-bg);
}

.compile-result-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-tertiary);
  cursor: pointer;
  font-size: 12px;
  transition: all var(--transition-fast);
}

.compile-result-close:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

/* ===== 编译文件树 ===== */
.compile-file-tree {
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-subtle);
  flex-shrink: 0;
  max-height: 200px;
  overflow-y: auto;
}

.compile-file-tree-header {
  padding: 8px 20px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
  border-bottom: 1px solid var(--border-subtle);
  background: var(--bg-tertiary);
}

.compile-file-tree-body {
  padding: 4px 0;
}

.compile-file-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 20px;
  font-size: 12px;
  font-family: var(--font-family-mono);
  color: var(--text-secondary);
  transition: background var(--transition-fast);
}

.compile-file-item:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.compile-file-path {
  white-space: nowrap;
}

.compile-file-tree-footer {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 20px;
  border-top: 1px solid var(--border-subtle);
  background: var(--bg-tertiary);
  font-size: 12px;
}

.compile-entry-label {
  color: var(--text-tertiary);
  font-weight: 500;
}

.compile-entry-path {
  font-family: var(--font-family-mono);
  color: var(--primary);
}

/* ===== Slide transitions ===== */
.slide-enter-active,
.slide-leave-active {
  transition: all 0.2s ease;
}
.slide-enter-from,
.slide-leave-to {
  opacity: 0;
  max-height: 0;
}
</style>