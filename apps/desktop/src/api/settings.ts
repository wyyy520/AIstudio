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
  const res = await http.get('/settings')
  return (res as unknown as { data: AppSettings }).data
}

export async function updateSettings(data: Partial<AppSettings>): Promise<AppSettings> {
  const res = await http.put('/settings', data)
  return (res as unknown as { data: AppSettings }).data
}

export async function getEngineConfig(): Promise<AppSettings['engine']> {
  const res = await http.get('/settings/engine')
  return (res as unknown as { data: AppSettings['engine'] }).data
}

export async function updateEngineConfig(data: Partial<AppSettings['engine']>): Promise<AppSettings['engine']> {
  const res = await http.put('/settings/engine', data)
  return (res as unknown as { data: AppSettings['engine'] }).data
}

export async function testEngineConnection(config: Partial<AppSettings['engine']>): Promise<{ success: boolean; message: string }> {
  try {
    const res = await http.post('/settings/engine/test', config)
    const data = res as unknown as { data: { success: boolean; message?: string } }
    return { success: data.data?.success ?? false, message: data.data?.message || '' }
  } catch (e: any) {
    return { success: false, message: e?.message || 'Connection failed' }
  }
}