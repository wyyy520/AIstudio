import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type {
  Task,
  LogEntry,
  ErrorAnalysis,
  TrainingMetrics,
  WorkflowTimeline,
  LogTab,
  FilterLevel,
  FixStep,
  AgentPhase,
} from '@/pages/Logs/types'
import * as logApi from '@/api/log'
import type { ApiLogEntry } from '@/api/log'
import { getTasks, getTaskStatus, type ApiTask, type ApiTaskStatus } from '@/api/task'
import { analyzeError, repairError } from '@/api/error'
import { wsClient, type WebSocketEvent } from '@/api/websocket'

// Map Backend ApiTask to Frontend Task type
function mapApiTaskToTask(api: ApiTask): Task {
  return {
    id: api.id,
    name: api.name,
    type: api.handler as Task['type'] || 'workflow',
    status: mapTaskStatus(api.status),
    startedAt: api.startedAt || api.createdAt,
    completedAt: api.completedAt,
    duration: 0,
    projectId: '',
    workflowId: undefined,
    metadata: api.payload as Record<string, unknown> | undefined,
  }
}

function mapApiTaskStatusToTask(status: ApiTaskStatus): Partial<Task> {
  return {
    status: mapTaskStatus(status.status),
  }
}

function mapTaskStatus(status: string): Task['status'] {
  switch (status) {
    case 'waiting': return 'running'
    case 'running': return 'running'
    case 'completed': return 'success'
    case 'success': return 'success'
    case 'failed': return 'failed'
    case 'cancelled': return 'failed'
    case 'warning': return 'warning'
    default: return 'running'
  }
}

// Map Backend ApiLogEntry to Frontend LogEntry
function mapApiLogToLogEntry(api: ApiLogEntry): LogEntry {
  return {
    id: api.id,
    taskId: api.taskId,
    timestamp: api.timestamp,
    level: (api.level as LogEntry['level']) || 'info',
    source: (api.source as LogEntry['source']) || 'system',
    message: api.message,
    rawMessage: api.message,
    humanMessage: api.message,
    stepName: api.stepName || '',
    stepStatus: (api.stepStatus as LogEntry['stepStatus']) || 'pending',
    metadata: api.metadata,
  }
}

export const useLogStore = defineStore('log', () => {
  const tasks = ref<Task[]>([])
  const selectedTaskId = ref<string | null>(null)
  const logs = ref<LogEntry[]>([])
  const errorAnalyses = ref<ErrorAnalysis[]>([])
  const trainingMetrics = ref<TrainingMetrics | null>(null)
  const workflowTimeline = ref<WorkflowTimeline | null>(null)

  const activeTab = ref<LogTab>('human')
  const filterLevel = ref<FilterLevel>('all')
  const searchQuery = ref('')

  const isAnalyzing = ref(false)
  const agentPhase = ref<AgentPhase>('idle')
  const fixSteps = ref<FixStep[]>([])
  const isFixing = ref(false)
  const showFixDialog = ref(false)
  const currentFixSolutionId = ref<string | null>(null)
  const currentFixAnalysisId = ref<string | null>(null)

  const isLoadingTasks = ref(false)
  const isLoadingLogs = ref(false)
  const error = ref<string | null>(null)

  // WebSocket unsubscribe functions
  let wsUnsubscribe: (() => void) | null = null
  let taskWsUnsubscribe: (() => void) | null = null

  const selectedTask = computed<Task | null>(() => {
    if (!selectedTaskId.value) return null
    return tasks.value.find(t => t.id === selectedTaskId.value) ?? null
  })

  const filteredLogs = computed<LogEntry[]>(() => {
    let result = logs.value
    if (filterLevel.value !== 'all') {
      result = result.filter(l => l.level === filterLevel.value)
    }
    if (searchQuery.value.trim()) {
      const q = searchQuery.value.toLowerCase().trim()
      result = result.filter(l =>
        l.message.toLowerCase().includes(q) ||
        l.humanMessage.toLowerCase().includes(q) ||
        l.rawMessage.toLowerCase().includes(q)
      )
    }
    return result
  })

  const taskGroups = computed(() => {
    const groups: Record<string, Task[]> = {
      running: tasks.value.filter(t => t.status === 'running'),
      failed: tasks.value.filter(t => t.status === 'failed'),
      warning: tasks.value.filter(t => t.status === 'warning'),
      completed: tasks.value.filter(t => t.status === 'success'),
    }
    return groups
  })

  const runningCount = computed(() => tasks.value.filter(t => t.status === 'running').length)
  const failedCount = computed(() => tasks.value.filter(t => t.status === 'failed').length)

  // Connect to WebSocket
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
    if (taskWsUnsubscribe) {
      taskWsUnsubscribe()
      taskWsUnsubscribe = null
    }
    wsClient.disconnect()
  }

  function handleWebSocketEvent(event: WebSocketEvent) {
    switch (event.type) {
      case 'task_status': {
        const task = tasks.value.find(t => t.id === event.taskId)
        if (task && event.data.status) {
          task.status = mapTaskStatus(event.data.status)
        }
        break
      }
      case 'task_progress': {
        const task = tasks.value.find(t => t.id === event.taskId)
        if (task && event.data.progress !== undefined) {
          // progress can be used for UI
        }
        break
      }
      case 'task_log': {
        if (event.data.message) {
          const logEntry: LogEntry = {
            id: `${event.taskId}-${Date.now()}`,
            taskId: event.taskId,
            timestamp: event.data.timestamp || new Date().toISOString(),
            level: (event.data.level as LogEntry['level']) || 'info',
            source: 'system',
            message: event.data.message,
            rawMessage: event.data.message,
            humanMessage: event.data.message,
            stepName: event.data.step || '',
            stepStatus: 'running',
          }
          if (selectedTaskId.value === event.taskId) {
            logs.value.push(logEntry)
          }
        }
        break
      }
      case 'task_error': {
        const task = tasks.value.find(t => t.id === event.taskId)
        if (task) {
          task.status = 'failed'
        }
        if (event.data.error) {
          error.value = event.data.error
        }
        // 任务失败后自动触发AI分析
        if (selectedTaskId.value === event.taskId && !isAnalyzing.value) {
          setTimeout(() => {
            analyzeCurrentTask()
          }, 500)
        }
        break
      }
      case 'task_complete': {
        const task = tasks.value.find(t => t.id === event.taskId)
        if (task) {
          task.status = 'success'
          task.completedAt = event.data.timestamp || new Date().toISOString()
        }
        break
      }
    }
  }

  async function loadTasks() {
    isLoadingTasks.value = true
    error.value = null
    try {
      const apiTasks = await getTasks()
      tasks.value = apiTasks.map(mapApiTaskToTask)
    } catch (e) {
      error.value = e instanceof Error ? e.message : '获取任务列表失败'
      console.error('[log-store] loadTasks failed:', e)
    } finally {
      isLoadingTasks.value = false
    }
  }

  async function selectTask(taskId: string) {
    selectedTaskId.value = taskId
    activeTab.value = 'human'
    filterLevel.value = 'all'
    searchQuery.value = ''
    errorAnalyses.value = []
    trainingMetrics.value = null
    workflowTimeline.value = null
    agentPhase.value = 'idle'
    error.value = null

    isLoadingLogs.value = true
    try {
      // Fetch logs from backend
      const apiLogs = await logApi.fetchTaskLogs(taskId)
      logs.value = apiLogs.map(mapApiLogToLogEntry)

      // Poll task status for real-time updates
      try {
        const status = await getTaskStatus(taskId)
        const existingTask = tasks.value.find(t => t.id === taskId)
        if (existingTask) {
          Object.assign(existingTask, mapApiTaskStatusToTask(status))
        }
      } catch {
        // status endpoint may not be available
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : '获取任务日志失败'
      console.error('[log-store] selectTask failed:', e)
    } finally {
      isLoadingLogs.value = false
    }

    // Subscribe to task-specific WebSocket events
    if (taskWsUnsubscribe) taskWsUnsubscribe()
    taskWsUnsubscribe = wsClient.subscribeTask(taskId, (event: WebSocketEvent) => {
      handleWebSocketEvent(event)
    })
  }

  async function analyzeCurrentTask() {
    if (!selectedTaskId.value) return
    isAnalyzing.value = true
    agentPhase.value = 'thinking'
    error.value = null
    try {
      agentPhase.value = 'analyzing'
      const response = await analyzeError(selectedTaskId.value)
      if (response.success && response.data) {
        errorAnalyses.value = response.data
      }
      agentPhase.value = 'completed'
    } catch (e) {
      agentPhase.value = 'failed'
      error.value = e instanceof Error ? e.message : 'AI分析失败'
      console.error('[log-store] analyzeCurrentTask failed:', e)
    } finally {
      isAnalyzing.value = false
    }
  }

  async function startFix(analysisId: string, solutionId: string) {
    currentFixAnalysisId.value = analysisId
    currentFixSolutionId.value = solutionId
    showFixDialog.value = true
    fixSteps.value = [
      { id: 'fix-1', label: '正在分析修复方案...', status: 'pending' },
      { id: 'fix-2', label: '正在执行修复...', status: 'pending' },
      { id: 'fix-3', label: '验证修复结果...', status: 'pending' },
    ]
  }

  async function executeFix() {
    if (!currentFixSolutionId.value || !currentFixAnalysisId.value) return
    isFixing.value = true
    try {
      for (let i = 0; i < fixSteps.value.length; i++) {
        fixSteps.value[i].status = 'running'
        await delay(600 + Math.random() * 400)
        fixSteps.value[i].status = 'completed'
      }
      const response = await repairError(currentFixAnalysisId.value, currentFixSolutionId.value)
      if (response.success) {
        const analysis = errorAnalyses.value.find(a => a.id === currentFixAnalysisId.value)
        if (analysis) {
          analysis.status = 'fixed'
        }
      }
      agentPhase.value = 'completed'
    } catch (e) {
      error.value = e instanceof Error ? e.message : '修复失败'
      console.error('[log-store] executeFix failed:', e)
    } finally {
      isFixing.value = false
      showFixDialog.value = false
      currentFixSolutionId.value = null
      currentFixAnalysisId.value = null
    }
  }

  function cancelFix() {
    showFixDialog.value = false
    currentFixSolutionId.value = null
    currentFixAnalysisId.value = null
    fixSteps.value = []
  }

  function ignoreAnalysis(analysisId: string) {
    const analysis = errorAnalyses.value.find(a => a.id === analysisId)
    if (analysis) {
      analysis.status = 'ignored'
    }
  }

  function setFilterLevel(level: FilterLevel) {
    filterLevel.value = level
  }

  function setSearchQuery(query: string) {
    searchQuery.value = query
  }

  function setActiveTab(tab: LogTab) {
    activeTab.value = tab
  }

  function clearLogs() {
    logs.value = []
  }

  async function exportCurrentLogs(format: 'log' | 'json' | 'csv') {
    if (!selectedTaskId.value || logs.value.length === 0) return
    const content = format === 'json'
      ? JSON.stringify(logs.value, null, 2)
      : logs.value.map(l => `[${l.timestamp}] [${l.level.toUpperCase()}] ${l.message}`).join('\n')
    const blob = new Blob([content], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `logs-${selectedTaskId.value}.${format}`
    a.click()
    URL.revokeObjectURL(url)
  }

  // Cleanup on store destroy
  function cleanup() {
    disconnectWebSocket()
  }

  return {
    tasks,
    selectedTaskId,
    logs,
    errorAnalyses,
    trainingMetrics,
    workflowTimeline,
    activeTab,
    filterLevel,
    searchQuery,
    isAnalyzing,
    agentPhase,
    fixSteps,
    isFixing,
    showFixDialog,
    currentFixSolutionId,
    currentFixAnalysisId,
    isLoadingTasks,
    isLoadingLogs,
    error,
    selectedTask,
    filteredLogs,
    taskGroups,
    runningCount,
    failedCount,
    connectWebSocket,
    disconnectWebSocket,
    loadTasks,
    selectTask,
    analyzeCurrentTask,
    startFix,
    executeFix,
    cancelFix,
    ignoreAnalysis,
    setFilterLevel,
    setSearchQuery,
    setActiveTab,
    clearLogs,
    exportCurrentLogs,
    cleanup,
  }
})

function delay(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms))
}