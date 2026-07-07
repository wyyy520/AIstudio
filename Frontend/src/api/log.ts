import type { Task, LogEntry, ErrorAnalysis, TrainingMetrics, WorkflowTimeline } from '@/pages/Logs/types'
import { mockTasks, mockLogs, mockErrorAnalyses, mockTrainingMetrics, mockWorkflowTimelines } from '@/pages/Logs/mock'

export async function fetchTasks(): Promise<Task[]> {
  await delay(300)
  return [...mockTasks]
}

export async function fetchTaskLogs(taskId: string): Promise<LogEntry[]> {
  await delay(200)
  return mockLogs[taskId] ? [...mockLogs[taskId]] : []
}

export async function fetchErrorAnalyses(taskId: string): Promise<ErrorAnalysis[]> {
  await delay(400)
  return mockErrorAnalyses[taskId] ? [...mockErrorAnalyses[taskId]] : []
}

export async function fetchTrainingMetrics(taskId: string): Promise<TrainingMetrics | null> {
  await delay(200)
  return mockTrainingMetrics[taskId] ?? null
}

export async function fetchWorkflowTimeline(taskId: string): Promise<WorkflowTimeline | null> {
  await delay(200)
  return mockWorkflowTimelines[taskId] ?? null
}

export async function analyzeLogs(taskId: string): Promise<ErrorAnalysis[]> {
  await delay(1500)
  return mockErrorAnalyses[taskId] ? [...mockErrorAnalyses[taskId]] : []
}

export async function applyFix(_analysisId: string, _solutionId: string): Promise<{ success: boolean; message: string }> {
  await delay(2000)
  return { success: true, message: 'Fix applied successfully' }
}

export async function exportLogs(taskId: string, format: 'log' | 'json' | 'csv'): Promise<Blob> {
  await delay(500)
  const logs = mockLogs[taskId] || []
  const content = format === 'json'
    ? JSON.stringify(logs, null, 2)
    : logs.map(l => `[${l.timestamp}] [${l.level.toUpperCase()}] ${l.message}`).join('\n')
  return new Blob([content], { type: 'text/plain' })
}

function delay(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms))
}