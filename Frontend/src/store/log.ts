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
import { createMockFixSteps } from '@/pages/Logs/mock'
import * as logApi from '@/api/log'

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

  const isLoadingTasks = ref(false)
  const isLoadingLogs = ref(false)

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

  async function loadTasks() {
    isLoadingTasks.value = true
    try {
      tasks.value = await logApi.fetchTasks()
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

    isLoadingLogs.value = true
    try {
      const [taskLogs, analyses, task] = await Promise.all([
        logApi.fetchTaskLogs(taskId),
        logApi.fetchErrorAnalyses(taskId),
        tasks.value.find(t => t.id === taskId),
      ])
      logs.value = taskLogs
      errorAnalyses.value = analyses

      if (task?.type === 'training') {
        trainingMetrics.value = await logApi.fetchTrainingMetrics(taskId)
      }
      if (task?.workflowId) {
        workflowTimeline.value = await logApi.fetchWorkflowTimeline(taskId)
      }
    } finally {
      isLoadingLogs.value = false
    }
  }

  async function analyzeCurrentTask() {
    if (!selectedTaskId.value) return
    isAnalyzing.value = true
    agentPhase.value = 'thinking'
    try {
      await delay(500)
      agentPhase.value = 'analyzing'
      const analyses = await logApi.analyzeLogs(selectedTaskId.value)
      agentPhase.value = 'completed'
      errorAnalyses.value = analyses
    } catch {
      agentPhase.value = 'failed'
    } finally {
      isAnalyzing.value = false
    }
  }

  async function startFix(solutionId: string) {
    currentFixSolutionId.value = solutionId
    showFixDialog.value = true
    fixSteps.value = createMockFixSteps()
  }

  async function executeFix() {
    if (!currentFixSolutionId.value) return
    isFixing.value = true
    fixSteps.value = createMockFixSteps()

    for (let i = 0; i < fixSteps.value.length; i++) {
      fixSteps.value[i].status = 'running'
      await delay(600 + Math.random() * 400)
      fixSteps.value[i].status = 'completed'
    }

    if (selectedTaskId.value) {
      const analysis = errorAnalyses.value.find(a =>
        a.solutions.some(s => s.id === currentFixSolutionId.value)
      )
      if (analysis) {
        analysis.status = 'fixed'
      }
    }

    isFixing.value = false
    showFixDialog.value = false
    currentFixSolutionId.value = null
    agentPhase.value = 'completed'
  }

  function cancelFix() {
    showFixDialog.value = false
    currentFixSolutionId.value = null
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
    if (!selectedTaskId.value) return
    const blob = await logApi.exportLogs(selectedTaskId.value, format)
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `logs-${selectedTaskId.value}.${format}`
    a.click()
    URL.revokeObjectURL(url)
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
    isLoadingTasks,
    isLoadingLogs,
    selectedTask,
    filteredLogs,
    taskGroups,
    runningCount,
    failedCount,
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
  }
})

function delay(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms))
}