import request from './request'

export interface ApiPluginSummary {
  id: string
  name: string
  version: string
  author?: string
  description: string
  type: string
  status: string
  enabled: boolean
  createdAt: string
  updatedAt: string
  size?: string
  downloads?: number
  githubUrl?: string
  readme?: string
}

export interface ApiPluginNode {
  name: string
  type: string
  category: string
}

export interface ApiPluginDependency {
  name: string
  version?: string
  version_min?: string
}

export interface ApiPlugin {
  id: string
  name: string
  version: string
  author?: string
  source?: string
  description: string
  type: string
  status: string
  enabled: boolean
  nodes: ApiPluginNode[]
  dependencies: ApiPluginDependency[]
  createdAt: string
  updatedAt: string
  size?: string
  downloads?: number
  githubUrl?: string
  readme?: string
}

export interface PluginInstallResponse {
  success: boolean
  message: string
  plugin?: ApiPlugin
  taskId?: string
}

export function getPlugins() {
  return request.get<ApiPluginSummary[]>('/plugins')
}

export function getPluginById(name: string) {
  return request.get<ApiPlugin>(`/plugins/${name}`)
}

export function installPlugin(name: string): Promise<PluginInstallResponse> {
  return request.post('/plugins/install', { name })
}

export function uninstallPlugin(name: string) {
  return request.delete(`/plugins/${name}`)
}

export function updatePlugin(name: string) {
  return request.put(`/plugins/${name}/update`)
}

export function enablePlugin(name: string) {
  return request.put(`/plugins/${name}/enable`, {})
}

export function disablePlugin(name: string) {
  return request.put(`/plugins/${name}/disable`, {})
}