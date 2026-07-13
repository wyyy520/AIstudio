import { apiClient } from './client'

export interface EnvironmentStatus {
  python: { version: string; installed: boolean }
  cuda: { version: string; installed: boolean }
  pytorch: { version: string; installed: boolean }
  gpu: { available: boolean; devices: string[] }
}

export async function getEnvironment(): Promise<EnvironmentStatus> {
  return apiClient.get<EnvironmentStatus>('/environment')
}
