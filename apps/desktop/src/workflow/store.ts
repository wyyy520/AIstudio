import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { WorkflowSchema, WorkflowNodeSchema, WorkflowEdgeSchema, ViewportSchema, Domain } from './schema'
import { createDefaultWorkflow } from './schema'
import { createNodeFromTemplate, getNodeTemplate } from './node/registry'
import { createEdge } from './edge/registry'
import { toVueFlowData, fromVueFlowData } from './serializer'
import type { Node, Edge } from '@vue-flow/core'
import type {
  ApiWorkflow,
  ApiNodeType,
  ApiWorkflowRunResult,
  CompileResultData,
} from '@/api/workflow'
import {
  getWorkflows,
  getWorkflowById,
  createWorkflow as apiCreateWorkflow,
  updateWorkflow as apiUpdateWorkflow,
  deleteWorkflow as apiDeleteWorkflow,
  runWorkflow as apiRunWorkflow,
  compileProject as apiCompileProject,
  runCompiledProject as apiRunCompiledProject,
  detectEnvironment as apiDetectEnvironment,
  getNodeTypes,
} from '@/api/workflow'

export interface WorkflowProject {
  id: string    // Backend project ID
  path: string  // Project root path on filesystem
  name: string
}

export const useWorkflowStore = defineStore('workflow', () => {
  const workflow = ref<WorkflowSchema>(createDefaultWorkflow('Untitled'))
  const project = ref<WorkflowProject | null>(null)
  const isDirty = ref(false)
  const lastSaved = ref<string>('')
  const loading = ref(false)
  const error = ref<string | null>(null)
  const selectedNodeId = ref<string | null>(null)
  const selectedEdgeId = ref<string | null>(null)

  // Undo/redo stacks
  const undoStack = ref<WorkflowSchema[]>([])
  const redoStack = ref<WorkflowSchema[]>([])
  const MAX_HISTORY = 50

  const canUndo = computed(() => undoStack.value.length > 0)
  const canRedo = computed(() => redoStack.value.length > 0)

  function pushSnapshot() {
    undoStack.value.push(JSON.parse(JSON.stringify(workflow.value)))
    if (undoStack.value.length > MAX_HISTORY) undoStack.value.shift()
    redoStack.value = []
  }

  function undo() {
    if (undoStack.value.length === 0) return
    redoStack.value.push(JSON.parse(JSON.stringify(workflow.value)))
    const prev = undoStack.value.pop()!
    workflow.value = prev
    isDirty.value = true
  }

  function redo() {
    if (redoStack.value.length === 0) return
    undoStack.value.push(JSON.parse(JSON.stringify(workflow.value)))
    const next = redoStack.value.pop()!
    workflow.value = next
    isDirty.value = true
  }

  const nodeCount = computed(() => workflow.value.nodes.length)
  const edgeCount = computed(() => workflow.value.edges.length)

  const selectedNode = computed(() => {
    if (!selectedNodeId.value) return null
    return workflow.value.nodes.find(n => n.id === selectedNodeId.value) || null
  })

  function setWorkflow(w: WorkflowSchema) {
    workflow.value = w
    isDirty.value = false
    lastSaved.value = w.updated_at
  }

  function resetWorkflow(name: string, domain: Domain = 'python') {
    workflow.value = createDefaultWorkflow(name, domain)
    project.value = null
    isDirty.value = false
    selectedNodeId.value = null
    selectedEdgeId.value = null
  }

  function addNode(type: string, position: { x: number; y: number }): string | null {
    pushSnapshot()
    const template = getNodeTemplate(type)
    if (!template) {
      error.value = `Unknown node type: ${type}`
      return null
    }
    const node = createNodeFromTemplate(template, position)
    workflow.value.nodes.push(node)
    workflow.value.updated_at = new Date().toISOString()
    isDirty.value = true
    return node.id
  }

  function removeNode(nodeId: string) {
    pushSnapshot()
    workflow.value.nodes = workflow.value.nodes.filter(n => n.id !== nodeId)
    workflow.value.edges = workflow.value.edges.filter(e => e.source_node !== nodeId && e.target_node !== nodeId)
    workflow.value.updated_at = new Date().toISOString()
    isDirty.value = true
    if (selectedNodeId.value === nodeId) selectedNodeId.value = null
  }

  function updateNodePosition(nodeId: string, x: number, y: number) {
    const node = workflow.value.nodes.find(n => n.id === nodeId)
    if (!node) return
    pushSnapshot()
    node.position.x = x
    node.position.y = y
    node.updated_at = new Date().toISOString()
    workflow.value.updated_at = new Date().toISOString()
    isDirty.value = true
  }

  function updateNodeConfig(nodeId: string, config: Record<string, unknown>) {
    pushSnapshot()
    const node = workflow.value.nodes.find(n => n.id === nodeId)
    if (!node) return
    node.config = { ...node.config, ...config }
    node.updated_at = new Date().toISOString()
    workflow.value.updated_at = new Date().toISOString()
    isDirty.value = true
  }

  function updateNodeName(nodeId: string, name: string) {
    pushSnapshot()
    const node = workflow.value.nodes.find(n => n.id === nodeId)
    if (!node) return
    node.name = name
    node.updated_at = new Date().toISOString()
    workflow.value.updated_at = new Date().toISOString()
    isDirty.value = true
  }

  function setNodeStatus(nodeId: string, status: WorkflowNodeSchema['status']) {
    const node = workflow.value.nodes.find(n => n.id === nodeId)
    if (!node) return
    node.status = status
  }

  function addEdge(
    sourceNode: string,
    sourcePort: string,
    targetNode: string,
    targetPort: string,
  ): string | null {
    pushSnapshot()
    const id = crypto.randomUUID()
    const edge = createEdge(id, sourceNode, sourcePort, targetNode, targetPort)
    workflow.value.edges.push(edge)
    workflow.value.updated_at = new Date().toISOString()
    isDirty.value = true
    return id
  }

  function removeEdge(edgeId: string) {
    pushSnapshot()
    workflow.value.edges = workflow.value.edges.filter(e => e.id !== edgeId)
    workflow.value.updated_at = new Date().toISOString()
    isDirty.value = true
    if (selectedEdgeId.value === edgeId) selectedEdgeId.value = null
  }

  function updateViewport(viewport: ViewportSchema) {
    workflow.value.viewport = viewport
  }

  function selectNode(nodeId: string | null) {
    selectedNodeId.value = nodeId
    if (nodeId) selectedEdgeId.value = null
  }

  function selectEdge(edgeId: string | null) {
    selectedEdgeId.value = edgeId
    if (edgeId) selectedNodeId.value = null
  }

  // Clipboard for copy/paste
  const clipboard = ref<WorkflowNodeSchema | null>(null)

  function copyNode(nodeId: string) {
    const node = workflow.value.nodes.find(n => n.id === nodeId)
    if (node) clipboard.value = JSON.parse(JSON.stringify(node))
  }

  function pasteNode(position?: { x: number; y: number }): string | null {
    if (!clipboard.value) return null
    pushSnapshot()
    const newNode = JSON.parse(JSON.stringify(clipboard.value)) as WorkflowNodeSchema
    newNode.id = crypto.randomUUID()
    const offset = position || { x: clipboard.value.position.x + 50, y: clipboard.value.position.y + 50 }
    newNode.position = offset
    newNode.created_at = new Date().toISOString()
    newNode.updated_at = new Date().toISOString()
    workflow.value.nodes.push(newNode)
    workflow.value.updated_at = new Date().toISOString()
    isDirty.value = true
    return newNode.id
  }

  const vueFlowNodes = computed(() => {
    return toVueFlowData(workflow.value).nodes
  })

  const vueFlowEdges = computed(() => {
    return toVueFlowData(workflow.value).edges
  })

  function syncFromVueFlow(nodes: Node[], edges: Edge[], viewport?: ViewportSchema) {
    if (isDirty.value) return
    const updated = fromVueFlowData(
      workflow.value.id,
      workflow.value.name,
      workflow.value.domain,
      nodes,
      edges,
      viewport || workflow.value.viewport,
    )
    updated.schema_version = workflow.value.schema_version
    updated.created_at = workflow.value.created_at
    updated.description = workflow.value.description
    updated.version = workflow.value.version
    updated.metadata = workflow.value.metadata
    workflow.value = updated
  }

  function getWorkflowJSON(): WorkflowSchema {
    return JSON.parse(JSON.stringify(workflow.value))
  }

  function loadWorkflowJSON(json: WorkflowSchema) {
    setWorkflow(json)
  }

  function toSaveJSON(): string {
    workflow.value.updated_at = new Date().toISOString()
    return JSON.stringify(workflow.value, null, 2)
  }

  // ============================================================================
  // API state: node types, workflow list, compile & run pipeline
  // ============================================================================

  const workflows = ref<WorkflowState[]>([])
  const currentWorkflow = ref<WorkflowState | null>(null)
  const nodeTypes = ref<ApiNodeType[]>([])
  const isRunning = ref(false)
  const lastRunTaskId = ref<string | null>(null)

  const compileResult = ref<CompileResultData | null>(null)
  const compileLoading = ref(false)
  const compileError = ref<string | null>(null)

  const workflowCount = computed(() => workflows.value.length)

  const nodeTypeMap = computed(() => {
    const map = new Map<string, ApiNodeType>()
    for (const nt of nodeTypes.value) {
      map.set(nt.type, nt)
    }
    return map
  })

  function parseDefinition(def: string): WorkflowDefinition {
    if (!def) return { nodes: [], edges: [] }
    try {
      return JSON.parse(def)
    } catch {
      return { nodes: [], edges: [] }
    }
  }

  function serializeDefinition(def: WorkflowDefinition): string {
    return JSON.stringify(def)
  }

  function mapApiToState(api: ApiWorkflow): WorkflowState {
    return {
      id: String(api.id),
      name: api.name,
      projectId: api.projectId,
      definition: parseDefinition(api.definition),
      status: api.status,
      createdAt: api.createdAt,
      updatedAt: api.updatedAt,
    }
  }

  async function fetchWorkflows(projectId?: string) {
    loading.value = true
    error.value = null
    try {
      const apiWorkflows = await getWorkflows(projectId)
      workflows.value = apiWorkflows.map(mapApiToState)
    } catch (e) {
      error.value = e instanceof Error ? e.message : '获取工作流列表失败'
    } finally {
      loading.value = false
    }
  }

  async function fetchWorkflowById(id: string) {
    loading.value = true
    error.value = null
    try {
      const apiWorkflow = await getWorkflowById(id)
      currentWorkflow.value = mapApiToState(apiWorkflow)
    } catch (e) {
      error.value = e instanceof Error ? e.message : '获取工作流详情失败'
    } finally {
      loading.value = false
    }
  }

  async function createWorkflow(data: { projectId: number; name: string; definition?: WorkflowDefinition }) {
    loading.value = true
    error.value = null
    try {
      const apiWorkflow = await apiCreateWorkflow({
        projectId: data.projectId,
        name: data.name,
        definition: data.definition ? serializeDefinition(data.definition) : undefined,
      })
      const state = mapApiToState(apiWorkflow)
      workflows.value.push(state)
      currentWorkflow.value = state
      return state
    } catch (e) {
      error.value = e instanceof Error ? e.message : '创建工作流失败'
      return null
    } finally {
      loading.value = false
    }
  }

  async function saveWorkflow(id: string, definition: WorkflowDefinition) {
    error.value = null
    try {
      const apiWorkflow = await apiUpdateWorkflow(id, {
        definition: serializeDefinition(definition),
      })
      const idx = workflows.value.findIndex(w => w.id === id)
      if (idx >= 0) {
        workflows.value[idx] = mapApiToState(apiWorkflow)
      }
      if (currentWorkflow.value?.id === id) {
        currentWorkflow.value = mapApiToState(apiWorkflow)
      }
      return true
    } catch (e) {
      error.value = e instanceof Error ? e.message : '保存工作流失败'
      return false
    }
  }

  async function compileWorkflow(projectId: string, target?: string) {
    compileLoading.value = true
    compileError.value = null
    compileResult.value = null
    try {
      const response = await apiCompileProject(projectId, target)
      const data = response.data || response
      if (data.projectRoot) {
        compileResult.value = data as CompileResultData
      } else if (data.generatorId !== undefined) {
        compileResult.value = data as CompileResultData
      }
      return data
    } catch (e) {
      compileError.value = e instanceof Error ? e.message : '编译工作流失败'
      return null
    } finally {
      compileLoading.value = false
    }
  }

  async function compileAndRun(projectId: string, target?: string) {
    isRunning.value = true
    error.value = null
    compileResult.value = null
    try {
      if (currentWorkflow.value) {
        const saved = await saveWorkflow(projectId, currentWorkflow.value.definition)
        if (!saved) {
          error.value = '保存工作流失败'
          return null
        }
      }

      compileLoading.value = true
      const compileResp = await apiCompileProject(projectId, target)
      const compileData = compileResp.data || compileResp
      if (compileData.projectRoot) {
        compileResult.value = compileData as CompileResultData
      }
      compileLoading.value = false

      if (!compileData || !compileData.projectRoot) {
        error.value = '编译失败：未生成项目文件'
        return null
      }

      const runResp = await apiRunCompiledProject(projectId, target)
      const runData = runResp.data || runResp
      lastRunTaskId.value = runData.runId

      if (currentWorkflow.value) {
        currentWorkflow.value.status = 'running'
      }

      return runData
    } catch (e) {
      error.value = e instanceof Error ? e.message : '运行工作流失败'
      lastRunTaskId.value = null
      return null
    } finally {
      isRunning.value = false
      compileLoading.value = false
    }
  }

  async function detectEnvironment() {
    try {
      const resp = await apiDetectEnvironment()
      return resp.data || resp
    } catch (e) {
      return null
    }
  }

  async function runWorkflow(id: string) {
    isRunning.value = true
    error.value = null
    try {
      const result = await apiRunWorkflow(id)
      const taskId = result.data?.id || result.id
      lastRunTaskId.value = taskId

      if (currentWorkflow.value?.id === id) {
        currentWorkflow.value.status = 'running'
      }
      const idx = workflows.value.findIndex(w => w.id === id)
      if (idx >= 0) {
        workflows.value[idx].status = 'running'
      }

      return taskId
    } catch (e) {
      error.value = e instanceof Error ? e.message : '运行工作流失败'
      return null
    } finally {
      isRunning.value = false
    }
  }

  async function deleteWorkflow(id: string) {
    error.value = null
    try {
      await apiDeleteWorkflow(id)
      workflows.value = workflows.value.filter(w => w.id !== id)
      if (currentWorkflow.value?.id === id) {
        currentWorkflow.value = null
      }
      return true
    } catch (e) {
      error.value = e instanceof Error ? e.message : '删除工作流失败'
      return false
    }
  }

  async function fetchNodeTypes() {
    error.value = null
    try {
      nodeTypes.value = await getNodeTypes()
    } catch (e) {
      error.value = e instanceof Error ? e.message : '获取节点类型失败'
    }
  }

  function initNewWorkflow(projectId: number, name: string): WorkflowState {
    const state: WorkflowState = {
      id: '',
      name,
      projectId,
      definition: { nodes: [], edges: [] },
      status: 'draft',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }
    currentWorkflow.value = state
    return state
  }

  return {
    workflow, project, isDirty, lastSaved, loading, error,
    selectedNodeId, selectedEdgeId,
    undoStack, redoStack, canUndo, canRedo,
    nodeCount, edgeCount, selectedNode,
    vueFlowNodes, vueFlowEdges,
    clipboard,
    setWorkflow, resetWorkflow,
    addNode, removeNode, updateNodePosition, updateNodeConfig, updateNodeName, setNodeStatus,
    addEdge, removeEdge,
    updateViewport,
    selectNode, selectEdge,
    copyNode, pasteNode,
    undo, redo,
    syncFromVueFlow,
    getWorkflowJSON, loadWorkflowJSON, toSaveJSON,
    // API state
    workflows, currentWorkflow, nodeTypes, workflowCount, nodeTypeMap,
    isRunning, lastRunTaskId,
    compileResult, compileLoading, compileError,
    fetchWorkflows, fetchWorkflowById, createWorkflow, saveWorkflow,
    compileWorkflow, compileAndRun, runWorkflow, deleteWorkflow,
    fetchNodeTypes, initNewWorkflow, detectEnvironment,
    parseDefinition, serializeDefinition,
  }
})

export interface WorkflowNode {
  id: string
  type: string
  name: string
  plugin: string
  description: string
  inputs: Array<{ name: string; type: string; required: boolean }>
  outputs: Array<{ name: string; type: string }>
  x: number
  y: number
  config: Record<string, unknown>
}

export interface WorkflowEdge {
  id: string
  source: string
  target: string
  sourceHandle: string
  targetHandle: string
}

export interface WorkflowDefinition {
  nodes: WorkflowNode[]
  edges: WorkflowEdge[]
}

export interface WorkflowState {
  id: string
  name: string
  projectId: number
  definition: WorkflowDefinition
  status: string
  createdAt: string
  updatedAt: string
}
