import request from './request'

export interface ApiLogEntry {
  id: string
  taskId?: string
  timestamp: string
  level: string
  source: string
  message: string
  stepName?: string
  stepStatus?: string
  metadata?: Record<string, unknown>
}

export interface LogEntry {
  id: string
  level: string
  message: string
  source: string
  timestamp: string
}

export function queryLogs(params?: { level?: string; source?: string; limit?: number; taskId?: string }) {
  return request.get('/logs', { params })
}

export function fetchTaskLogs(taskId: string) {
  return request.get(`/logs`, { params: { taskId } })
}