import http from './request'

export interface ApiPlugin {
  name: string
  version: string
  description: string
  entry: string
  status: string
  path: string
  createdAt: string
  updatedAt: string
}

export async function getPlugins(): Promise<ApiPlugin[]> {
  const res = await http.get('/api/plugins')
  return (res as unknown as { data: ApiPlugin[] }).data
}

export async function getPluginByName(name: string): Promise<ApiPlugin> {
  const res = await http.get(`/api/plugins/${name}`)
  return (res as unknown as { data: ApiPlugin }).data
}

export async function enablePlugin(name: string): Promise<void> {
  await http.post(`/api/plugins/${name}/enable`)
}

export async function disablePlugin(name: string): Promise<void> {
  await http.post(`/api/plugins/${name}/disable`)
}

export async function uninstallPlugin(name: string): Promise<void> {
  await http.delete(`/api/plugins/${name}`)
}

export async function installPluginAction(name: string): Promise<void> {
  await http.post('/api/plugins/install', { name })
}

export async function updatePluginStatus(name: string, status: string): Promise<void> {
  await http.put(`/api/plugins/${name}/status`, { status })
}

export async function executePlugin(name: string, input: Record<string, unknown>): Promise<unknown> {
  const res = await http.post(`/api/plugins/${name}/execute`, { input })
  return (res as unknown as { data: unknown }).data
}

// Aliases for store compatibility
export async function installPlugin(pluginId: string): Promise<unknown> {
  const res = await http.post('/api/plugins/install', { name: pluginId })
  return (res as unknown as { data: unknown }).data
}

export async function updatePlugin(pluginId: string): Promise<unknown> {
  const res = await http.put(`/api/plugins/update`, { name: pluginId })
  return (res as unknown as { data: unknown }).data
}