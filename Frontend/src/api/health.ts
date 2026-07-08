import { apiClient } from './client'

export interface ModuleStatus {
  name: string
  status: string
  message?: string
}

export interface HealthCheckData {
  status: string
  uptime: string
  modules: ModuleStatus[]
  version: string
  go: string
  platform: string
}

export interface HealthCheckResponse {
  code: number
  message: string
  data: HealthCheckData
}

export async function getHealth(): Promise<HealthCheckData> {
  return apiClient.get<HealthCheckData>('/api/health')
}

export async function isBackendReady(): Promise<boolean> {
  try {
    const health = await getHealth()
    return health.status === 'running'
  } catch {
    return false
  }
}