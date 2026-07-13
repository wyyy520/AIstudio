import { defineStore } from 'pinia'
import { ref } from 'vue'
import { compileProject, stopProject, getRuntimeStatus } from '@/api/workflow'
import { wsClient, type WebSocketEvent } from '@/api/websocket'

export type RunStatus = 'idle' | 'compiling' | 'running' | 'completed' | 'failed'

export const useRuntimeStore = defineStore('runtime', () => {
  const status = ref<RunStatus>('idle')
  const currentRunId = ref<string | null>(null)
  const currentProjectId = ref<string | null>(null)
  const error = ref<string | null>(null)
  const progress = ref(0)

  let wsUnsubscribe: (() => void) | null = null

  function connectWebSocket() {
    wsClient.connect()
    wsUnsubscribe = wsClient.subscribe((event: WebSocketEvent) => {
      if (event.taskId !== currentRunId.value) return
      switch (event.type) {
        case 'task_status':
          if (event.data.status === 'running') status.value = 'running'
          break
        case 'task_progress':
          if (event.data.progress !== undefined) progress.value = event.data.progress
          break
        case 'task_complete':
          status.value = 'completed'
          progress.value = 100
          break
        case 'task_error':
          status.value = 'failed'
          error.value = event.data.error || 'Run failed'
          break
      }
    })
  }

  function disconnectWebSocket() {
    if (wsUnsubscribe) {
      wsUnsubscribe()
      wsUnsubscribe = null
    }
  }

  async function compileAndRun(projectId: string) {
    status.value = 'compiling'
    currentProjectId.value = projectId
    error.value = null
    progress.value = 0

    try {
      await compileProject(projectId)
      status.value = 'running'
      currentRunId.value = `run_${projectId}_${Date.now()}`
      connectWebSocket()
    } catch (e) {
      status.value = 'failed'
      error.value = e instanceof Error ? e.message : 'Compile failed'
    }
  }

  async function stop() {
    if (currentRunId.value) {
      try {
        await stopProject({ runId: currentRunId.value })
      } catch { }
    }
    status.value = 'idle'
    currentRunId.value = null
    progress.value = 0
    error.value = null
    disconnectWebSocket()
  }

  function reset() {
    status.value = 'idle'
    currentRunId.value = null
    currentProjectId.value = null
    error.value = null
    progress.value = 0
    disconnectWebSocket()
  }

  return {
    status, currentRunId, currentProjectId, error, progress,
    compileAndRun, stop, reset, connectWebSocket, disconnectWebSocket,
  }
})
