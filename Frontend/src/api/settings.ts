import http from './request'

export interface AppSettings {
  language: string
  autoSave: boolean
  autoSaveInterval: number
  startupBehavior: 'dashboard' | 'last-project' | 'empty'
  theme: 'light' | 'dark' | 'system'

  engine: {
    provider: 'openai' | 'claude' | 'local'
    model: string
    apiKey: string
    endpoint: string
    timeout: number
  }

  shortcuts: Record<string, string>
}

export async function getSettings(): Promise<AppSettings> {
  const res = await http.get('/api/settings')
  return (res as unknown as { data: AppSettings }).data
}

export async function updateSettings(data: Partial<AppSettings>): Promise<AppSettings> {
  const res = await http.put('/api/settings', data)
  return (res as unknown as { data: AppSettings }).data
}

export async function getEngineConfig(): Promise<AppSettings['engine']> {
  const res = await http.get('/api/settings/engine')
  return (res as unknown as { data: AppSettings['engine'] }).data
}

export async function updateEngineConfig(data: Partial<AppSettings['engine']>): Promise<AppSettings['engine']> {
  const res = await http.put('/api/settings/engine', data)
  return (res as unknown as { data: AppSettings['engine'] }).data
}

export async function testEngineConnection(data: { provider: string; endpoint: string; apiKey: string }): Promise<{ success: boolean; message: string }> {
  const res = await http.post('/api/settings/engine/test', data)
  return (res as unknown as { data: { success: boolean; message: string } }).data
}