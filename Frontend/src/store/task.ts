import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ApiTask, ApiTaskStatus, ApiTaskCreateRequest } from '@/api/task'
import {
  getTasks,
  getTaskById,
  getTaskStatus,
  createTask,
  cancelTask,
  deleteTask,
} from '@/api/task'
import { wsClient, type WebSocketEvent } from '@/api/websocket'

export interface TaskState {
  id: string
  name: string
  description: string
  priority: number
  status: string
  progress: number
  handler: string
  payload?: unknown
  result?: unknown
  error?: string
  createdAt: string
  updatedAt: string
  startedAt?: string
  completedAt?: string
}

function mapApiToState(api: ApiTask): TaskState {
  return {
    id: api.id,
    name: api.name,
    description: api.description || '',
    priority: api.priority || 0,
    status: api.status,
    progress: 0,
    handler: api.handler,
    payload: api.payload,
    result: api.result,
    error: api.error,
    createdAt: api.createdAt,
    updatedAt: api.updatedAt,
    startedAt: api.startedAt,
    completedAt: api.completedAt,
  }
}

function mapStatusToState(status: ApiTaskStatus): Partial<TaskState> {
  return {
    status: status.status,
    progress: status.progress || 0,
    result: status.result,
    error: status.error,
    startedAt: status.startedAt,
    completedAt: status.completedAt,
  }
}

export const useTaskStore = defineStore('task', () => {
  const tasks = ref<TaskState[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const pollingTasks = ref<Map<string, ReturnType<typeof setInterval>>>(new Map())

  let wsUnsubscribe: (() => void) | null = null

  const runningTasks = computed(() => tasks.value.filter(t => t.status === 'running'))
  const failedTasks = computed(() => tasks.value.filter(t => t.status === 'failed'))
  const completedTasks = computed(() => tasks.value.filter(t => t.status === 'completed'))
  const taskCount = computed(() => tasks.value.length)

  function connectWebSocket() {
    wsClient.connect()
    wsUnsubscribe = wsClient.subscribe((event: WebSocketEvent) => {
      handleWebSocketEvent(event)
    })
  }

  function disconnectWebSocket() {
    if (wsUnsubscribe) {
      wsUnsubscribe()
      wsUnsubscribe = null
    }
    wsClient.disconnect()
  }

  function handleWebSocketEvent(event: WebSocketEvent) {
    const task = tasks.value.find(t => t.id === event.taskId)
    if (!task) return

    switch (event.type) {
      case 'task_status':
        if (event.data.status) task.status = event.data.status
        break
      case 'task_progress':
        if (event.data.progress !== undefined) task.progress = event.data.progress
        break
      case 'task_complete':
        task.status = 'completed'
        task.completedAt = event.data.timestamp || new Date().toISOString()
        task.progress = 100
        stopPolling(event.taskId)
        break
      case 'task_error':
        task.status = 'failed'
        if (event.data.error) task.error = event.data.error
        stopPolling(event.taskId)
        break
    }
  }

  async function fetchTasks() {
    loading.value = true
    error.value = null
    try {
      const apiTasks = await getTasks()
      tasks.value = apiTasks.map(mapApiToState)
    } catch (e) {
      error.value = e instanceof Error ? e.message : '获取任务列表失败'
      console.error('[task-store] fetchTasks failed:', e)
    } finally {
      loading.value = false
    }
  }

  async function fetchTaskDetail(id: string) {
    error.value = null
    try {
      const apiTask = await getTaskById(id)
      const state = mapApiToState(apiTask)
      const idx = tasks.value.findIndex(t => t.id === id)
      if (idx >= 0) {
        tasks.value[idx] = state
      }
      return state
    } catch (e) {
      error.value = e instanceof Error ? e.message : '获取任务详情失败'
      return null
    }
  }

  async function fetchTaskStatus(id: string) {
    try {
      const status = await getTaskStatus(id)
      const idx = tasks.value.findIndex(t => t.id === id)
      if (idx >= 0) {
        Object.assign(tasks.value[idx], mapStatusToState(status))
      }
      return status
    } catch {
      return null
    }
  }

  async function createNewTask(data: ApiTaskCreateRequest) {
    loading.value = true
    error.value = null
    try {
      const result = await createTask(data)
      const newTask: TaskState = {
        id: result.task_id,
        name: data.name,
        description: data.description || '',
        priority: data.priority || 0,
        status: 'pending',
        progress: 0,
        handler: data.handler,
        payload: data.payload,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      }
      tasks.value.push(newTask)

      // Start polling for task status
      startPolling(result.task_id)

      return result.task_id
    } catch (e) {
      error.value = e instanceof Error ? e.message : '创建任务失败'
      return null
    } finally {
      loading.value = false
    }
  }

  async function cancelTaskById(id: string) {
    error.value = null
    try {
      await cancelTask(id)
      const task = tasks.value.find(t => t.id === id)
      if (task) {
        task.status = 'cancelled'
      }
      stopPolling(id)
      return true
    } catch (e) {
      error.value = e instanceof Error ? e.message : '取消任务失败'
      return false
    }
  }

  async function deleteTaskById(id: string) {
    error.value = null
    try {
      await deleteTask(id)
      tasks.value = tasks.value.filter(t => t.id !== id)
      stopPolling(id)
      return true
    } catch (e) {
      error.value = e instanceof Error ? e.message : '删除任务失败'
      return false
    }
  }

  function startPolling(taskId: string) {
    if (pollingTasks.value.has(taskId)) return

    const interval = setInterval(async () => {
      try {
        const status = await getTaskStatus(taskId)
        const idx = tasks.value.findIndex(t => t.id === taskId)
        if (idx >= 0) {
          Object.assign(tasks.value[idx], mapStatusToState(status))
          if (status.status === 'completed' || status.status === 'failed') {
            stopPolling(taskId)
          }
        } else {
          stopPolling(taskId)
        }
      } catch {
        stopPolling(taskId)
      }
    }, 2000)

    pollingTasks.value.set(taskId, interval)
  }

  function stopPolling(taskId: string) {
    const interval = pollingTasks.value.get(taskId)
    if (interval) {
      clearInterval(interval)
      pollingTasks.value.delete(taskId)
    }
  }

  function cleanup() {
    disconnectWebSocket()
    for (const [taskId] of pollingTasks.value) {
      stopPolling(taskId)
    }
  }

  return {
    tasks,
    loading,
    error,
    runningTasks,
    failedTasks,
    completedTasks,
    taskCount,
    connectWebSocket,
    disconnectWebSocket,
    fetchTasks,
    fetchTaskDetail,
    fetchTaskStatus,
    createNewTask,
    cancelTaskById,
    deleteTaskById,
    startPolling,
    stopPolling,
    cleanup,
  }
})