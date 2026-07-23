<template>
  <div class="workflow-editor" :class="{ 'fullscreen': isFullscreen }">
    <!-- Run status bar -->
    <div v-if="isRunning || currentTaskId" class="run-status-bar" :class="`run-status-bar--${runStatus}`">
      <div class="run-status-left">
        <span class="run-status-text">{{ runStatusText }}</span>
      </div>
      <div class="run-status-right">
        <div v-if="isRunning" class="run-progress">
          <div class="run-progress-bar">
            <div class="run-progress-fill" :style="{ width: `${progress}%` }"></div>
          </div>
          <span class="run-progress-text">{{ progress }}%</span>
        </div>
        <button v-if="!isRunning && currentTaskId" class="view-logs-btn" @click="viewLogs">查看日志</button>
      </div>
    </div>

    <WorkflowToolbar
      v-show="!isFullscreen"
      :workflow-name="wfStore.workflow.name"
      :validation-errors="validationErrors"
      :is-saving="isSaving"
      @save="handleSave"
      @run="handleRun"
      @validate="handleValidate"
      @zoom-in="handleZoomIn"
      @zoom-out="handleZoomOut"
      @fit-view="handleFitView"
      @toggle-fullscreen="handleToggleFullscreen"
      @export-json="handleExportJSON"
    />
    <div class="editor-body">
      <NodePanel
        v-show="!isFullscreen"
        @add-node="handleAddNode"
      />
      <FlowCanvas
        ref="flowCanvasRef"
        v-model:nodes="bridge.vueFlowNodes"
        v-model:edges="bridge.vueFlowEdges"
        @node-click="handleNodeClick"
        @add-node="handleAddNodeFromDrop"
        @node-delete="handleDeleteNode"
        @node-drag-stop="bridge.onNodeDragStop"
        @connect="bridge.onConnect"
        @edge-dblclick="bridge.onEdgeDoubleClick"
        @pane-click="bridge.onPaneClick"
      />
      <PropertyPanel
        v-show="!isFullscreen"
        :selected-node="propertyPanelNode"
        :total-nodes="wfStore.nodeCount"
        :total-edges="wfStore.edgeCount"
        @update-node-label="handleUpdateNodeLabel"
        @update-node-params="handleUpdateNodeParams"
        @delete-node="handleDeleteNode"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import type { Node } from '@vue-flow/core'
import { useWorkflowStore } from '@/workflow/store'
import { useWorkflowBridge } from '@/workflow/services/bridge'
import { startAutoSave, stopAutoSave, flushSave } from '@/workflow/services/auto-save'
import { readWorkflow, saveWorkflow as apiSaveWorkflow } from '@/api/project'
import type { ValidationError } from '../validators/workflowValidator'
import { validateWorkflow } from '../validators/workflowValidator'
import WorkflowToolbar from './WorkflowToolbar.vue'
import NodePanel from './NodePanel.vue'
import FlowCanvas from './FlowCanvas.vue'
import PropertyPanel from './PropertyPanel.vue'

interface Props {
  projectId?: string
  workflowName?: string
}

const props = withDefaults(defineProps<Props>(), {
  projectId: undefined,
  workflowName: 'New Workflow',
})

const router = useRouter()
const flowCanvasRef = ref<InstanceType<typeof FlowCanvas>>()
const wfStore = useWorkflowStore()
const bridge = useWorkflowBridge()
const isFullscreen = ref(false)
const isRunning = ref(false)
const currentTaskId = ref<string | null>(null)
const runStatus = ref<'idle' | 'running' | 'success' | 'error'>('idle')
const progress = ref(0)
const isSaving = ref(false)
const validationErrors = ref<ValidationError[]>([])

const propertyPanelNode = computed(() => {
  const sel = wfStore.selectedNode
  if (!sel) return null
  return {
    id: sel.id,
    data: {
      label: sel.name,
      description: sel.description || '',
      nodeType: sel.type,
      params: sel.config,
      inputs: sel.inputs.map(i => ({ name: i.name, type: i.type, label: i.name })),
      outputs: sel.outputs.map(o => ({ name: o.name, type: o.type, label: o.name })),
      status: sel.status || 'idle',
    },
  } as unknown as Node
})

const runStatusText = computed(() => {
  switch (runStatus.value) {
    case 'running': return 'Running...'
    case 'success': return 'Success'
    case 'error': return 'Failed'
    default: return ''
  }
})

// ===== Lifecycle =====

onMounted(async () => {
  if (props.projectId) {
    await loadWorkflow(props.projectId)
  } else {
    wfStore.resetWorkflow(props.workflowName || 'Untitled')
  }
  startAutoSave()
  // Keyboard shortcuts
  document.addEventListener('keydown', handleKeyboardShortcut)
})

onUnmounted(() => {
  stopAutoSave()
  document.removeEventListener('keydown', handleKeyboardShortcut)
})

// Watch for projectId changes (re-load workflow when navigating)
watch(() => props.projectId, async (newId) => {
  if (newId) {
    await loadWorkflow(newId)
  }
})

async function loadWorkflow(projectId: string) {
  try {
    const wfData = await readWorkflow(projectId)
    if (wfData && wfData.nodes) {
      wfStore.loadWorkflowJSON(wfData)
      wfStore.project = {
        id: projectId,
        path: '',
        name: wfData.name || 'Untitled',
      }
    } else {
      wfStore.resetWorkflow(props.workflowName, 'python')
      wfStore.project = { id: projectId, path: '', name: props.workflowName }
    }
  } catch {
    wfStore.resetWorkflow(props.workflowName, 'python')
    wfStore.project = { id: projectId, path: '', name: props.workflowName }
  }
}

// ===== Node operations =====

function handleNodeClick(node: Node | null) {
  if (node) wfStore.selectNode(node.id)
  else wfStore.selectNode(null)
}

function handleAddNode(template: any) {
  if (!flowCanvasRef.value) return
  const center = flowCanvasRef.value.getViewportCenter()
  const type = template.key || template.nodeType || template.type
  wfStore.addNode(type, { x: center.x - 100, y: center.y - 60 })
}

function handleAddNodeFromDrop(template: any, position: { x: number; y: number }) {
  const type = template.key || template.nodeType || template.type
  wfStore.addNode(type, position)
}

function handleUpdateNodeLabel(label: string) {
  if (wfStore.selectedNodeId) {
    wfStore.updateNodeName(wfStore.selectedNodeId, label)
  }
}

function handleUpdateNodeParams(params: Record<string, unknown>) {
  if (wfStore.selectedNodeId) {
    wfStore.updateNodeConfig(wfStore.selectedNodeId, params)
  }
}

function handleDeleteNode() {
  if (wfStore.selectedNodeId) {
    wfStore.removeNode(wfStore.selectedNodeId)
  }
}

// ===== Workflow operations =====

async function handleSave() {
  if (!props.projectId) return
  isSaving.value = true
  try {
    const wf = wfStore.getWorkflowJSON()
    await apiSaveWorkflow(props.projectId, wf)
    wfStore.isDirty = false
    wfStore.lastSaved = wf.updated_at || new Date().toISOString()
  } catch (err) {
    console.error('[workflow-editor] save failed:', err)
  } finally {
    isSaving.value = false
  }
}

function handleExportJSON() {
  const json = wfStore.toSaveJSON()
  const blob = new Blob([json], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${wfStore.workflow.name || 'workflow'}.json`
  a.click()
  URL.revokeObjectURL(url)
}

function handleValidate() {
  validationErrors.value = validateWorkflow(bridge.vueFlowNodes.value, bridge.vueFlowEdges.value)
}

async function handleRun() {
  validationErrors.value = validateWorkflow(bridge.vueFlowNodes.value, bridge.vueFlowEdges.value)
  if (validationErrors.value.length > 0) return
  if (!bridge.vueFlowNodes.value.length) return

  // Save first
  if (props.projectId) {
    isSaving.value = true
    try {
      await apiSaveWorkflow(props.projectId, wfStore.getWorkflowJSON())
    } finally {
      isSaving.value = false
    }
  }

  // Simple mock execution
  isRunning.value = true
  runStatus.value = 'running'
  progress.value = 0

  const nodeIds = bridge.vueFlowNodes.value.map(n => n.id)
  for (let i = 0; i < nodeIds.length; i++) {
    wfStore.setNodeStatus(nodeIds[i], 'running')
    await new Promise(r => setTimeout(r, 500))
    wfStore.setNodeStatus(nodeIds[i], 'success')
    progress.value = Math.round(((i + 1) / nodeIds.length) * 100)
  }

  isRunning.value = false
  runStatus.value = 'success'
  progress.value = 100
}

function viewLogs() {
  if (currentTaskId.value) {
    router.push({ path: '/logs', query: { taskId: currentTaskId.value } })
  }
}

function handleZoomIn() { flowCanvasRef.value?.zoomIn() }
function handleZoomOut() { flowCanvasRef.value?.zoomOut() }
function handleFitView() { flowCanvasRef.value?.fitView() }
function handleToggleFullscreen() { isFullscreen.value = !isFullscreen.value }

// Keyboard shortcuts
function handleKeyboardShortcut(e: KeyboardEvent) {
  const isCtrlOrCmd = e.ctrlKey || e.metaKey

  // Save: Ctrl+S
  if (isCtrlOrCmd && e.key === 's') {
    e.preventDefault()
    handleSave()
    return
  }

  // Undo: Ctrl+Z
  if (isCtrlOrCmd && !e.shiftKey && e.key === 'z') {
    e.preventDefault()
    wfStore.undo()
    return
  }

  // Redo: Ctrl+Shift+Z or Ctrl+Y
  if ((isCtrlOrCmd && e.shiftKey && e.key === 'z') || (isCtrlOrCmd && e.key === 'y')) {
    e.preventDefault()
    wfStore.redo()
    return
  }

  // Copy: Ctrl+C (only if a node is selected)
  if (isCtrlOrCmd && e.key === 'c' && wfStore.selectedNodeId) {
    e.preventDefault()
    wfStore.copyNode(wfStore.selectedNodeId)
    return
  }

  // Paste: Ctrl+V
  if (isCtrlOrCmd && e.key === 'v') {
    e.preventDefault()
    wfStore.pasteNode()
    return
  }

  // Delete: Backspace or Delete (handle in canvas, but add fallback)
  if (e.key === 'Delete' || e.key === 'Backspace') {
    // Prevent browser back on Backspace
    if (e.target === document.body) {
      e.preventDefault()
    }
  }
}
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

.run-status-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 44px;
  padding: 0 24px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-subtle);
}

.run-status-bar--running { background: rgba(59, 130, 246, 0.08); }
.run-status-bar--success { background: rgba(34, 197, 94, 0.08); }
.run-status-bar--error { background: rgba(239, 68, 68, 0.08); }

.run-status-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.run-status-text {
  font-size: var(--text-body-sm);
  font-weight: var(--font-medium);
  color: var(--text-primary);
}

.run-status-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

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
}
</style>
