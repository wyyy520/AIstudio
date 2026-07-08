import http from './request'

// Backend log entry
export interface ApiLogEntry {
  id: string
  taskId: string
  timestamp: string
  level: string
  source: string
  message: string
  stepName: string
  stepStatus: string
  metadata?: Record<string, unknown>
}

export interface ApiLogQueryResult {
  entries: ApiLogEntry[]
  page: number
  size: number
  total: number
}

export interface ApiLogQueryParams {
  level?: string
  source?: string
  taskId?: string
  keyword?: string
  start?: string
  end?: string
  page?: number
  size?: number
}

// GET /api/logs — query logs
export async function queryLogs(params?: ApiLogQueryParams): Promise<ApiLogQueryResult> {
  const res = await http.get('/api/logs', { params })
  return (res as unknown as { data: ApiLogQueryResult }).data
}

// GET /api/logs?taskId=xxx — get logs for a specific task
export async function fetchTaskLogs(taskId: string): Promise<ApiLogEntry[]> {
  const result = await queryLogs({ taskId, size: 500 })
  return result.entries ?? []
}