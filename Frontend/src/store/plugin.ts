import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Plugin, PluginCategory, InstallTask, InstallStep } from '@/pages/PluginStore/types'
import type { ApiPluginSummary, ApiPlugin } from '@/api/plugin'
import * as pluginApi from '@/api/plugin'

// Map Backend ApiPluginSummary to Frontend Plugin type
function mapSummaryToPlugin(api: ApiPluginSummary): Plugin {
  return {
    id: api.id || api.name,
    name: api.name,
    version: api.version,
    author: api.author || 'AIStudio',
    source: 'local',
    sourceUrl: '',
    description: api.description || '',
    category: (api.type as PluginCategory) || 'system',
    icon: getCategoryIcon(api.type),
    status: mapBackendStatus(api.status, api.enabled),
    capabilities: [],
    workflowNodes: [],
    dependencies: [],
    agentTools: [],
    tags: [api.type],
    installedAt: api.createdAt,
    updatedAt: api.updatedAt,
  }
}

// Map Backend full ApiPlugin to Frontend Plugin type
function mapDetailToPlugin(api: ApiPlugin): Plugin {
  return {
    id: api.id || api.name,
    name: api.name,
    version: api.version,
    author: api.author || 'AIStudio',
    source: (api.source as Plugin['source']) || 'local',
    sourceUrl: '',
    description: api.description || '',
    category: (api.type as PluginCategory) || 'system',
    icon: getCategoryIcon(api.type),
    status: mapBackendStatus(api.status, api.enabled),
    capabilities: api.nodes.map(n => n.name),
    workflowNodes: api.nodes.map(n => ({
      name: n.name,
      type: n.type,
      category: (api.type as PluginCategory) || 'system',
    })),
    dependencies: api.dependencies.map(d => ({
      name: d.name,
      versionRequired: d.version_min || d.version || 'any',
      versionInstalled: d.version || undefined,
      status: 'satisfied',
    })),
    agentTools: [],
    tags: [api.type],
    installedAt: api.createdAt,
    updatedAt: api.updatedAt,
  }
}

function getCategoryIcon(type: string): string {
  const icons: Record<string, string> = {
    vision: 'M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z M11 12a2 2 0 1 0 4 0 2 2 0 0 0-4 0z',
    nlp: 'M4 7V4h16v3M9 20h6M12 4v16',
    timeseries: 'M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z M12 6v6l4 2',
    simulation: 'M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z',
    mcp: 'M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71 M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71',
    system: 'M4 17l6-6-6-6M12 19h8',
    speech: 'M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z M19 10v2a7 7 0 0 1-14 0v-2 M12 19v4M8 23h8',
  }
  return icons[type] || icons.system
}

function mapBackendStatus(status: string, enabled: boolean): Plugin['status'] {
  switch (status) {
    case 'not_installed': return 'not-installed'
    case 'installing': return 'installing'
    case 'installed': return enabled ? 'installed' : 'not-installed'
    case 'updating': return 'updating'
    case 'error': return 'error'
    default: return 'not-installed'
  }
}

export const usePluginStore = defineStore('plugin', () => {
  const plugins = ref<Plugin[]>([])
  const selectedPluginId = ref<string | null>(null)
  const activeCategory = ref<PluginCategory | 'all'>('all')
  const searchQuery = ref('')
  const installTask = ref<InstallTask | null>(null)
  const isInstalling = ref(false)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const categories = ref([
    { category: 'vision' as const, label: 'AI Vision', icon: 'M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z M11 12a2 2 0 1 0 4 0 2 2 0 0 0-4 0z', color: 'var(--vision)' },
    { category: 'nlp' as const, label: 'NLP', icon: 'M4 7V4h16v3M9 20h6M12 4v16', color: 'var(--nlp)' },
    { category: 'timeseries' as const, label: 'Time Series', icon: 'M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z M12 6v6l4 2', color: 'var(--timeseries)' },
    { category: 'speech' as const, label: 'Speech', icon: 'M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z M19 10v2a7 7 0 0 1-14 0v-2 M12 19v4M8 23h8', color: 'var(--nlp)' },
    { category: 'simulation' as const, label: 'Simulation', icon: 'M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z', color: 'var(--simulation)' },
    { category: 'system' as const, label: 'System', icon: 'M4 17l6-6-6-6M12 19h8', color: 'var(--system)' },
    { category: 'mcp' as const, label: 'MCP', icon: 'M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71 M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71', color: 'var(--mcp)' },
  ])

  const filteredPlugins = computed(() => {
    let result = plugins.value

    if (activeCategory.value !== 'all') {
      result = result.filter(p => p.category === activeCategory.value)
    }

    if (searchQuery.value.trim()) {
      const q = searchQuery.value.toLowerCase().trim()
      result = result.filter(p =>
        p.name.toLowerCase().includes(q) ||
        p.description.toLowerCase().includes(q) ||
        p.tags.some(t => t.toLowerCase().includes(q)) ||
        p.capabilities.some(c => c.toLowerCase().includes(q))
      )
    }

    return result
  })

  const selectedPlugin = computed(() => {
    if (!selectedPluginId.value) return null
    return plugins.value.find(p => p.id === selectedPluginId.value) ?? null
  })

  const installedCount = computed(() => {
    return plugins.value.filter(p => p.status === 'installed').length
  })

  const updatesAvailableCount = computed(() => {
    return plugins.value.filter(p => p.status === 'updating').length
  })

  const errorCount = computed(() => {
    return plugins.value.filter(p => p.status === 'error').length
  })

  async function fetchPlugins() {
    loading.value = true
    error.value = null
    try {
      const summaries = await pluginApi.getPlugins()
      plugins.value = summaries.map(mapSummaryToPlugin)
    } catch (e) {
      error.value = e instanceof Error ? e.message : '获取插件列表失败'
      console.error('[plugin-store] fetchPlugins failed:', e)
    } finally {
      loading.value = false
    }
  }

  async function fetchPluginDetail(id: string) {
    error.value = null
    try {
      const detail = await pluginApi.getPluginById(id)
      const mapped = mapDetailToPlugin(detail)
      const idx = plugins.value.findIndex(p => p.id === id || p.name === id)
      if (idx >= 0) {
        plugins.value[idx] = mapped
      }
      return mapped
    } catch (e) {
      error.value = e instanceof Error ? e.message : '获取插件详情失败'
      return null
    }
  }

  function selectPlugin(pluginId: string) {
    selectedPluginId.value = pluginId
  }

  function setCategory(category: PluginCategory | 'all') {
    activeCategory.value = category
  }

  function setSearchQuery(query: string) {
    searchQuery.value = query
  }

  async function installPluginAction(pluginId: string) {
    const plugin = plugins.value.find(p => p.id === pluginId)
    if (!plugin) return

    plugin.status = 'installing'
    isInstalling.value = true
    error.value = null

    installTask.value = {
      id: `install-${Date.now()}`,
      pluginId: plugin.id,
      pluginName: plugin.name,
      status: 'running',
      steps: [
        { id: 'step-1', name: '验证依赖', status: 'pending', logs: [] },
        { id: 'step-2', name: '下载插件', status: 'pending', logs: [] },
        { id: 'step-3', name: '安装依赖', status: 'pending', logs: [] },
        { id: 'step-4', name: '注册节点', status: 'pending', logs: [] },
      ],
      startedAt: new Date().toISOString(),
    }

    try {
      await simulateInstallProgress(installTask.value, plugin)
      const result = await pluginApi.installPlugin(plugin.name)
      if (result.success) {
        installTask.value!.status = 'completed'
        installTask.value!.completedAt = new Date().toISOString()
        plugin.status = 'installed'
        plugin.installedAt = new Date().toISOString()

        // Refresh plugin detail to get nodes
        if (result.plugin) {
          const updated = mapDetailToPlugin(result.plugin)
          const idx = plugins.value.findIndex(p => p.id === pluginId)
          if (idx >= 0) plugins.value[idx] = updated
        }
      } else {
        throw new Error(result.message || '安装失败')
      }
    } catch (e) {
      installTask.value!.status = 'failed'
      plugin.status = 'error'
      error.value = e instanceof Error ? e.message : '安装失败'
    } finally {
      isInstalling.value = false
    }
  }

  async function uninstallPluginAction(pluginId: string) {
    const plugin = plugins.value.find(p => p.id === pluginId)
    if (!plugin) return

    error.value = null
    try {
      await pluginApi.uninstallPlugin(plugin.name)
      plugin.status = 'not-installed'
      delete plugin.installedAt
    } catch (e) {
      error.value = e instanceof Error ? e.message : '卸载失败'
    }
  }

  async function updatePluginAction(pluginId: string) {
    const plugin = plugins.value.find(p => p.id === pluginId)
    if (!plugin) return

    plugin.status = 'updating'
    isInstalling.value = true
    error.value = null

    try {
      await pluginApi.updatePlugin(pluginId)
      plugin.status = 'installed'
      plugin.updatedAt = new Date().toISOString()
    } catch {
      plugin.status = 'error'
    } finally {
      isInstalling.value = false
    }
  }

  function cancelInstall() {
    if (installTask.value) {
      installTask.value.status = 'cancelled'
      const plugin = plugins.value.find(p => p.id === installTask.value!.pluginId)
      if (plugin) {
        plugin.status = 'not-installed'
      }
      isInstalling.value = false
    }
  }

  function clearInstallTask() {
    installTask.value = null
  }

  async function simulateInstallProgress(task: InstallTask, plugin: Plugin) {
    for (let i = 0; i < task.steps.length; i++) {
      if (task.status === 'cancelled') break

      const step = task.steps[i]
      step.status = 'in-progress'
      step.logs = [{ timestamp: new Date().toISOString(), level: 'info', message: `正在${step.name}...` }]

      const duration = 300 + Math.random() * 500
      await new Promise(resolve => setTimeout(resolve, duration))

      if (task.status === 'cancelled') break

      step.status = 'completed'
      step.duration = Math.round(duration)
      step.logs.push({ timestamp: new Date().toISOString(), level: 'info', message: `${step.name}完成.` })
    }
  }

  return {
    plugins,
    selectedPluginId,
    activeCategory,
    searchQuery,
    installTask,
    isInstalling,
    loading,
    error,
    categories,
    filteredPlugins,
    selectedPlugin,
    installedCount,
    updatesAvailableCount,
    errorCount,
    fetchPlugins,
    fetchPluginDetail,
    selectPlugin,
    setCategory,
    setSearchQuery,
    installPluginAction,
    uninstallPluginAction,
    updatePluginAction,
    cancelInstall,
    clearInstallTask,
  }
})