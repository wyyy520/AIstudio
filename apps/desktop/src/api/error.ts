import request from './request'
import type { ErrorAnalysis } from '@/pages/Logs/types'

export interface AnalyzeErrorRequest {
  taskId: string
  logIds?: string[]
}

export interface AnalyzeErrorResponse {
  success: boolean
  data: ErrorAnalysis[]
  message?: string
}

export interface RepairErrorRequest {
  analysisId: string
  solutionId: string
}

export interface RepairErrorResponse {
  success: boolean
  data: {
    fixId: string
    status: 'pending' | 'running' | 'completed' | 'failed'
    message?: string
  }
}

export function analyzeError(taskId: string, logIds?: string[]) {
  return request.post<AnalyzeErrorResponse>('/diagnostic/analyze', {
    taskId,
    logIds,
  })
}

export function repairError(analysisId: string, solutionId: string) {
  return request.post<RepairErrorResponse>('/diagnostic/repair', {
    analysisId,
    solutionId,
  })
}

export function getErrorAnalysis(taskId: string) {
  return request.get<AnalyzeErrorResponse>(`/diagnostic/analysis/${taskId}`)
}

export function getFixStatus(fixId: string) {
  return request.get(`/diagnostic/fix/${fixId}/status`)
}
