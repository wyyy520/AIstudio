import http from './request'

// Backend PluginSummary matches GET /api/plugins response
export interface ApiPluginSummary {
  id: string
  name: string
  version: string
  author: string
  type: string
  description: string
  status: string
  enabled: boolean
  nodeCount: number
  createdAt: string
  updatedAt: string
}

// Backend Plugin (full detail) matches GET /api/plugin/:id response
export interface ApiPlugin {
  id: string
  name: string
  version: string
  author: string
  type: string
  description: string
  entry: string
  source: string
  path: string
  config: string
  dependencies: ApiDependency[]
  status: string
  enabled: boolean
  nodes: ApiNodeRegistration[]
  createdAt: string
  updatedAt: string
}

export interface ApiDependency {
  name: string
  version: string
  version_min: string
  optional: boolean
}

export interface ApiNodeRegistration {
  type: string
  name: string
  description: string
  inputs: Array<{ name: string; type: string; required: boolean }>
  outputs: Array<{ name: string; type: string }>
}

export interface ApiInstallResult {
  success: boolean
  message: string
  plugin: ApiPlugin | null
  dependencies: ApiDependencyCheckResult[]
}

export interface ApiDependencyCheckResult {
  name: string
  status: string
  message: string
  required: string
  installed: string
}

// GET /api/plugins — list all plugin summaries
export async function getPlugins(): Promise<ApiPluginSummary[]> {
  const res = await http.get('/api/plugins')
  return (res as unknown as { data: ApiPluginSummary[] }).data ?? []
}

// GET /api/plugin/:id — get plugin details by ID or name
export async function getPluginById(id: string): Promise<ApiPlugin> {
  const res = await http.get(`/api/plugin/${id}`)
  return (res as unknown as { data: ApiPlugin }).data
}

// POST /api/plugin/install — install a plugin
export async function installPlugin(name: string, url?: string): Promise<ApiInstallResult> {
  const res = await http.post('/api/plugin/install', { name, url })
  return (res as unknown as { data: ApiInstallResult }).data
}

// POST /api/plugin/remove — remove a plugin
export async function removePlugin(name: string): Promise<void> {
  await http.post('/api/plugin/remove', { name })
}

// Legacy aliases for store compatibility
export async function getPluginByName(name: string): Promise<ApiPlugin> {
  return getPluginById(name)
}

export async function enablePlugin(name: string): Promise<void> {
  await http.put(`/api/plugins/${name}/status`, { status: 'enabled' })
}

export async function disablePlugin(name: string): Promise<void> {
  await http.put(`/api/plugins/${name}/status`, { status: 'disabled' })
}

export async function uninstallPlugin(name: string): Promise<void> {
  await removePlugin(name)
}

export async function installPluginAction(name: string): Promise<ApiInstallResult> {
  return installPlugin(name)
}

export async function updatePluginStatus(name: string, status: string): Promise<void> {
  await http.put(`/api/plugins/${name}/status`, { status })
}

export async function executePlugin(name: string, input: Record<string, unknown>): Promise<unknown> {
  const res = await http.post(`/api/plugins/${name}/execute`, { input })
  return (res as unknown as { data: unknown }).data
}

export async function updatePlugin(pluginId: string): Promise<unknown> {
  const res = await http.put(`/api/plugins/update`, { name: pluginId })
  return (res as unknown as { data: unknown }).data
}