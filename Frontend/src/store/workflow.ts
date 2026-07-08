import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type {
  ApiWorkflow,
  ApiNodeType,
  ApiWorkflowRunResult,
} from '@/api/workflow'
import {
  getWorkflows,
  getWorkflowById,
  createWorkflow as apiCreateWorkflow,
  updateWorkflow as apiUpdateWorkflow,
  deleteWorkflow as apiDeleteWorkflow,
  runWorkflow as apiRunWorkflow,
  getNodeTypes,
} from '@/api/workflow'

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

export const useWorkflowStore = defineStore('workflow', () => {
  const workflows = ref<WorkflowState[]>([])
  const currentWorkflow = ref<WorkflowState | null>(null)
  const nodeTypes = ref<ApiNodeType[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const isRunning = ref(false)
  const lastRunTaskId = ref<string | null>(null)

  const workflowCount = computed(() => workflows.value.length)

  const nodeTypeMap = computed(() => {
    const map = new Map<string, ApiNodeType>()
    for (const nt of nodeTypes.value) {
      map.set(nt.type, nt)
    }
    return map
  })

  // Parse definition JSON string to WorkflowDefinition
  function parseDefinition(def: string): WorkflowDefinition {
    if (!def) return { nodes: [], edges: [] }
    try {
      return JSON.parse(def)
    } catch {
      return { nodes: [], edges: [] }
    }
  }

  // Serialize WorkflowDefinition to JSON string
  function serializeDefinition(def: WorkflowDefinition): string {
    return JSON.stringify(def)
  }

  // Map API workflow to store state
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
      console.error('[workflow-store] fetchWorkflows failed:', e)
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
      console.error('[workflow-store] fetchWorkflowById failed:', e)
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
      console.error('[workflow-store] createWorkflow failed:', e)
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
      console.error('[workflow-store] saveWorkflow failed:', e)
      return false
    }
  }

  async function runWorkflow(id: string) {
    isRunning.value = true
    error.value = null
    try {
      const result = await apiRunWorkflow(id)
      lastRunTaskId.value = result.task_id

      // Update workflow status
      if (currentWorkflow.value?.id === id) {
        currentWorkflow.value.status = 'running'
      }
      const idx = workflows.value.findIndex(w => w.id === id)
      if (idx >= 0) {
        workflows.value[idx].status = 'running'
      }

      return result.task_id
    } catch (e) {
      error.value = e instanceof Error ? e.message : '运行工作流失败'
      console.error('[workflow-store] runWorkflow failed:', e)
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
      console.error('[workflow-store] fetchNodeTypes failed:', e)
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

  function updateNodePosition(nodeId: string, x: number, y: number) {
    if (!currentWorkflow.value) return
    const node = currentWorkflow.value.definition.nodes.find(n => n.id === nodeId)
    if (node) {
      node.x = x
      node.y = y
    }
  }

  function addNode(node: WorkflowNode) {
    if (!currentWorkflow.value) return
    currentWorkflow.value.definition.nodes.push(node)
  }

  function removeNode(nodeId: string) {
    if (!currentWorkflow.value) return
    currentWorkflow.value.definition.nodes = currentWorkflow.value.definition.nodes.filter(n => n.id !== nodeId)
    currentWorkflow.value.definition.edges = currentWorkflow.value.definition.edges.filter(
      e => e.source !== nodeId && e.target !== nodeId
    )
  }

  function addEdge(edge: WorkflowEdge) {
    if (!currentWorkflow.value) return
    currentWorkflow.value.definition.edges.push(edge)
  }

  function removeEdge(edgeId: string) {
    if (!currentWorkflow.value) return
    currentWorkflow.value.definition.edges = currentWorkflow.value.definition.edges.filter(e => e.id !== edgeId)
  }

  return {
    workflows,
    currentWorkflow,
    nodeTypes,
    loading,
    error,
    isRunning,
    lastRunTaskId,
    workflowCount,
    nodeTypeMap,
    fetchWorkflows,
    fetchWorkflowById,
    createWorkflow,
    saveWorkflow,
    runWorkflow,
    deleteWorkflow,
    fetchNodeTypes,
    initNewWorkflow,
    updateNodePosition,
    addNode,
    removeNode,
    addEdge,
    removeEdge,
    parseDefinition,
    serializeDefinition,
  }
})