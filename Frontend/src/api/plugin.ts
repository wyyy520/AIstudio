import type { Plugin, InstallTask, PluginCategory } from '@/pages/PluginStore/types'
import { mockPlugins, createMockInstallSteps } from '@/pages/PluginStore/mock'

export async function fetchPlugins(category?: PluginCategory): Promise<Plugin[]> {
  await delay(300)
  if (!category) return [...mockPlugins]
  return mockPlugins.filter(p => p.category === category)
}

export async function fetchPluginDetail(pluginId: string): Promise<Plugin | undefined> {
  await delay(200)
  return mockPlugins.find(p => p.id === pluginId)
}

export async function searchPlugins(query: string): Promise<Plugin[]> {
  await delay(200)
  const q = query.toLowerCase()
  return mockPlugins.filter(p =>
    p.name.toLowerCase().includes(q) ||
    p.description.toLowerCase().includes(q) ||
    p.tags.some(t => t.toLowerCase().includes(q)) ||
    p.capabilities.some(c => c.toLowerCase().includes(q))
  )
}

export async function installPlugin(pluginId: string): Promise<InstallTask> {
  await delay(100)
  const plugin = mockPlugins.find(p => p.id === pluginId)
  if (!plugin) throw new Error(`Plugin ${pluginId} not found`)
  return {
    id: `task-${Date.now()}`,
    pluginId,
    pluginName: plugin.name,
    status: 'running',
    steps: createMockInstallSteps(plugin.name),
    startedAt: new Date().toISOString(),
  }
}

export async function uninstallPlugin(pluginId: string): Promise<void> {
  await delay(500)
  const plugin = mockPlugins.find(p => p.id === pluginId)
  if (plugin) {
    plugin.status = 'not-installed'
    delete plugin.installedAt
  }
}

export async function updatePlugin(pluginId: string): Promise<InstallTask> {
  await delay(100)
  const plugin = mockPlugins.find(p => p.id === pluginId)
  if (!plugin) throw new Error(`Plugin ${pluginId} not found`)
  return {
    id: `task-update-${Date.now()}`,
    pluginId,
    pluginName: plugin.name,
    status: 'running',
    steps: createMockInstallSteps(plugin.name).slice(3),
    startedAt: new Date().toISOString(),
  }
}

function delay(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms))
}