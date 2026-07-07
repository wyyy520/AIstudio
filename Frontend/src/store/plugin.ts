import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Plugin, InstallTask, PluginCategory, InstallStep } from '@/pages/PluginStore/types'
import { mockPlugins, mockCategoryGroups } from '@/pages/PluginStore/mock'
import * as pluginApi from '@/api/plugin'

export const usePluginStore = defineStore('plugin', () => {
  const plugins = ref<Plugin[]>([...mockPlugins])
  const selectedPluginId = ref<string | null>(null)
  const activeCategory = ref<PluginCategory | 'all'>('all')
  const searchQuery = ref('')
  const installTask = ref<InstallTask | null>(null)
  const isInstalling = ref(false)

  const categories = ref(mockCategoryGroups)

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
        p.capabilities.some(c => c.toLowerCase().includes(q)) ||
        p.workflowNodes.some(n => n.name.toLowerCase().includes(q)) ||
        p.agentTools.some(t => t.name.toLowerCase().includes(q))
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

    try {
      const task = await pluginApi.installPlugin(pluginId)
      installTask.value = task
      await simulateInstallSteps(task, pluginId)
    } catch {
      plugin.status = 'error'
    } finally {
      isInstalling.value = false
    }
  }

  async function uninstallPluginAction(pluginId: string) {
    const plugin = plugins.value.find(p => p.id === pluginId)
    if (!plugin) return

    await pluginApi.uninstallPlugin(pluginId)
    plugin.status = 'not-installed'
    delete plugin.installedAt
  }

  async function updatePluginAction(pluginId: string) {
    const plugin = plugins.value.find(p => p.id === pluginId)
    if (!plugin) return

    plugin.status = 'updating'
    isInstalling.value = true

    try {
      const task = await pluginApi.updatePlugin(pluginId)
      installTask.value = task
      await simulateInstallSteps(task, pluginId)
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

  async function simulateInstallSteps(task: InstallTask, pluginId: string) {
    const plugin = plugins.value.find(p => p.id === pluginId)
    if (!plugin) return

    for (let i = 0; i < task.steps.length; i++) {
      if (task.status === 'cancelled') break

      const step = task.steps[i]
      step.status = 'in-progress'
      step.logs = [{ timestamp: new Date().toISOString(), level: 'info', message: `Starting ${step.name.toLowerCase()}...` }]

      const duration = 400 + Math.random() * 800
      await new Promise(resolve => setTimeout(resolve, duration))

      if (task.status === 'cancelled') break

      step.status = 'completed'
      step.duration = Math.round(duration)
      step.logs.push({ timestamp: new Date().toISOString(), level: 'info', message: `${step.name} completed.` })
    }

    if (task.status !== 'cancelled') {
      task.status = 'completed'
      task.completedAt = new Date().toISOString()
      plugin.status = 'installed'
      plugin.installedAt = new Date().toISOString()
    }
  }

  return {
    plugins,
    selectedPluginId,
    activeCategory,
    searchQuery,
    installTask,
    isInstalling,
    categories,
    filteredPlugins,
    selectedPlugin,
    installedCount,
    updatesAvailableCount,
    errorCount,
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