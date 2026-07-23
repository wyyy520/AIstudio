import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Plugin, PluginCategory, InstallTask, InstallStep } from '@/pages/PluginStore/types'
import type { ApiPluginSummary, ApiPlugin } from '@/api/plugin'
import * as pluginApi from '@/api/plugin'
import { useWorkflowStore } from '@/workflow/store'

// Map Backend ApiPluginSummary to Frontend Plugin type
function mapSummaryToPlugin(api: ApiPluginSummary): Plugin {
  return {
    id: api.id || api.name,
    name: api.name,
    version: api.version,
    author: api.author || 'AIStudio',
    source: 'registry',
    sourceUrl: api.githubUrl || '',
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
    size: api.size || '',
    downloads: api.downloads || 0,
    githubUrl: api.githubUrl || '',
    readme: api.readme || '',
  }
}

// Map Backend full ApiPlugin to Frontend Plugin type
function mapDetailToPlugin(api: ApiPlugin): Plugin {
  return {
    id: api.id || api.name,
    name: api.name,
    version: api.version,
    author: api.author || 'AIStudio',
    source: (api.source as Plugin['source']) || 'registry',
    sourceUrl: api.githubUrl || '',
    description: api.description || '',
    category: (api.type as PluginCategory) || 'system',
    icon: getCategoryIcon(api.type),
    status: mapBackendStatus(api.status, api.enabled),
    capabilities: api.nodes.map(n => n.name),
    workflowNodes: api.nodes.map(n => ({
      name: n.name,
      type: n.type,
      category: (n.category as PluginCategory) || (api.type as PluginCategory) || 'system',
    })),
    dependencies: api.dependencies.map(d => ({
      name: d.name,
      versionRequired: d.version_min || d.version || 'any',
      versionInstalled: d.version || undefined,
      status: 'satisfied' as const,
    })),
    agentTools: [],
    tags: [api.type],
    installedAt: api.createdAt,
    updatedAt: api.updatedAt,
    size: api.size || '',
    downloads: api.downloads || 0,
    githubUrl: api.githubUrl || '',
    readme: api.readme || '',
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
    case 'installed': return enabled ? 'installed' : 'installed'
    case 'updating': return 'updating'
    case 'error': return 'error'
    default: return 'installed'
  }
}

// Fallback plugin list when backend API is unavailable
const FALLBACK_PLUGINS: Plugin[] = [
  { id: 'yolo-vision', name: 'YOLO 目标检测', version: '1.2.0', author: 'AIStudio', source: 'registry' as const, sourceUrl: '', description: '基于 YOLOv8/v9/v10/v11 的高性能目标检测，支持训练、推理和部署全流程', category: 'vision' as PluginCategory, icon: getCategoryIcon('vision'), status: 'not-installed' as const, capabilities: ['yolo_training', 'yolo_inference', 'image_classification'], workflowNodes: [{ name: 'YOLO 训练', type: 'vision', category: 'vision' as const }, { name: 'YOLO 推理', type: 'vision', category: 'vision' as const }], dependencies: [], agentTools: [], tags: ['vision', 'detection', 'yolo'], size: '1.2GB', downloads: 15800, githubUrl: '', readme: '' },
  { id: 'nlp-llm', name: 'LLM 大语言模型', version: '1.1.0', author: 'AIStudio', source: 'registry' as const, sourceUrl: '', description: '集成 GPT、Claude、LLaMA 等大语言模型，支持文本生成、对话、总结等 NLP 任务', category: 'nlp' as PluginCategory, icon: getCategoryIcon('nlp'), status: 'not-installed' as const, capabilities: ['text_generation', 'chat_completion', 'text_summarization'], workflowNodes: [{ name: 'LLM 对话', type: 'nlp', category: 'nlp' as const }, { name: '文本总结', type: 'nlp', category: 'nlp' as const }], dependencies: [], agentTools: [], tags: ['nlp', 'llm', 'chat'], size: '3.5GB', downloads: 23400, githubUrl: '', readme: '' },
  { id: 'resnet-classifier', name: 'ResNet/EfficientNet 分类', version: '1.0.0', author: 'AIStudio', source: 'registry' as const, sourceUrl: '', description: 'ResNet18/34/50/101 和 EfficientNet 系列图像分类模型，支持迁移学习和训练', category: 'vision' as PluginCategory, icon: getCategoryIcon('vision'), status: 'not-installed' as const, capabilities: ['resnet_classify', 'efficientnet_classify'], workflowNodes: [{ name: 'ResNet 分类', type: 'vision', category: 'vision' as const }, { name: 'EfficientNet 分类', type: 'vision', category: 'vision' as const }], dependencies: [], agentTools: [], tags: ['vision', 'classification', 'resnet'], size: '450MB', downloads: 8900, githubUrl: '', readme: '' },
  { id: 'lstm-transformer', name: 'LSTM/Transformer 文本', version: '1.0.1', author: 'AIStudio', source: 'registry' as const, sourceUrl: '', description: 'LSTM 和 Transformer 序列模型，用于文本分类、情感分析、序列预测', category: 'nlp' as PluginCategory, icon: getCategoryIcon('nlp'), status: 'not-installed' as const, capabilities: ['lstm_text', 'transformer_text'], workflowNodes: [{ name: 'LSTM 文本', type: 'nlp', category: 'nlp' as const }, { name: 'Transformer', type: 'nlp', category: 'nlp' as const }], dependencies: [], agentTools: [], tags: ['nlp', 'lstm', 'transformer'], size: '280MB', downloads: 6200, githubUrl: '', readme: '' },
  { id: 'faster-rcnn-ssd', name: 'Faster-RCNN/SSD 检测', version: '1.0.0', author: 'AIStudio', source: 'registry' as const, sourceUrl: '', description: '经典两阶段和一阶段目标检测算法 Faster-RCNN、SSD，ResNet/MobileNetV2 主干', category: 'vision' as PluginCategory, icon: getCategoryIcon('vision'), status: 'not-installed' as const, capabilities: ['faster_rcnn', 'ssd_detect'], workflowNodes: [{ name: 'Faster-RCNN', type: 'vision', category: 'vision' as const }, { name: 'SSD 检测', type: 'vision', category: 'vision' as const }], dependencies: [], agentTools: [], tags: ['vision', 'detection', 'rcnn', 'ssd'], size: '560MB', downloads: 7200, githubUrl: '', readme: '' },
  { id: 'unet-mask-rcnn', name: 'U-Net/Mask-RCNN 分割', version: '1.0.0', author: 'AIStudio', source: 'registry' as const, sourceUrl: '', description: '语义分割 U-Net 和实例分割 Mask-RCNN，支持自定义训练', category: 'vision' as PluginCategory, icon: getCategoryIcon('vision'), status: 'not-installed' as const, capabilities: ['unet_segment', 'mask_rcnn'], workflowNodes: [{ name: 'U-Net 分割', type: 'vision', category: 'vision' as const }, { name: 'Mask-RCNN', type: 'vision', category: 'vision' as const }], dependencies: [], agentTools: [], tags: ['vision', 'segmentation', 'unet'], size: '480MB', downloads: 5500, githubUrl: '', readme: '' },
  { id: 'simulation-traffic', name: 'SUMO 交通仿真', version: '1.0.2', author: 'AIStudio', source: 'registry' as const, sourceUrl: '', description: 'SUMO 开源交通仿真平台集成，支持道路网络建模、车流仿真和数据导出', category: 'simulation' as PluginCategory, icon: getCategoryIcon('simulation'), status: 'not-installed' as const, capabilities: ['sumo_traffic'], workflowNodes: [{ name: 'SUMO 仿真', type: 'simulation', category: 'simulation' as const }], dependencies: [], agentTools: [], tags: ['simulation', 'traffic', 'sumo'], size: '320MB', downloads: 3800, githubUrl: '', readme: '' },
  { id: 'data-tools', name: '数据处理工具集', version: '1.0.0', author: 'AIStudio', source: 'registry' as const, sourceUrl: '', description: '数据加载、清洗、归一化、分割等数据预处理工具集', category: 'system' as PluginCategory, icon: getCategoryIcon('system'), status: 'not-installed' as const, capabilities: ['dataset_loader', 'data_cleaning', 'normalize', 'data_split'], workflowNodes: [{ name: 'Dataset 加载', type: 'dataset', category: 'system' as const }, { name: '数据清洗', type: 'data', category: 'system' as const }, { name: '归一化', type: 'data', category: 'system' as const }], dependencies: [], agentTools: [], tags: ['data', 'preprocessing', 'normalize'], size: '120MB', downloads: 12500, githubUrl: '', readme: '' },
  { id: 'mcp-server', name: 'MCP 服务连接器', version: '1.1.0', author: 'AIStudio', source: 'registry' as const, sourceUrl: '', description: 'Model Context Protocol 服务连接器，连接外部 MCP 服务扩展 AI Studio 功能', category: 'mcp' as PluginCategory, icon: getCategoryIcon('mcp'), status: 'not-installed' as const, capabilities: ['mcp_connect'], workflowNodes: [{ name: 'MCP 连接', type: 'mcp', category: 'mcp' as const }], dependencies: [], agentTools: [], tags: ['mcp', 'connect', 'service'], size: '60MB', downloads: 2100, githubUrl: '', readme: '' },
]

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
    { category: 'all' as const, label: 'All Plugins', icon: 'M3 3h7v7H3V3zm0 11h7v7H3v-7zm11-11h7v7h-7V3zm0 11h7v7h-7v-7z', color: 'var(--primary)' },
    { category: 'vision' as const, label: 'AI Vision', icon: 'M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z M11 12a2 2 0 1 0 4 0 2 2 0 0 0-4 0z', color: 'var(--vision)' },
    { category: 'nlp' as const, label: 'NLP', icon: 'M4 7V4h16v3M9 20h6M12 4v16', color: 'var(--nlp)' },
    { category: 'timeseries' as const, label: 'Time Series', icon: 'M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z M12 6v6l4 2', color: 'var(--timeseries)' },
    { category: 'speech' as const, label: 'Speech', icon: 'M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z M19 10v2a7 7 0 0 1-14 0v-2 M12 19v4M8 23h8', color: 'var(--info)' },
    { category: 'simulation' as const, label: 'Simulation', icon: 'M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z', color: 'var(--simulation)' },
    { category: 'system' as const, label: 'System Tools', icon: 'M4 17l6-6-6-6M12 19h8', color: 'var(--system)' },
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
      console.warn('[plugin-store] Backend unavailable, using fallback plugin list:', e)
      // Use fallback data when backend API is not available
      plugins.value = FALLBACK_PLUGINS.map(p => ({ ...p, id: p.id }))
      // Don't set error for fallback — plugins still show
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
        { id: 'step-1', name: 'Downloading', status: 'pending', logs: [] },
        { id: 'step-2', name: 'Installing', status: 'pending', logs: [] },
        { id: 'step-3', name: 'Checking', status: 'pending', logs: [] },
        { id: 'step-4', name: 'Ready', status: 'pending', logs: [] },
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

        if (result.plugin) {
          const updated = mapDetailToPlugin(result.plugin)
          const idx = plugins.value.findIndex(p => p.id === pluginId)
          if (idx >= 0) plugins.value[idx] = updated
        }

        await refreshWorkflowNodes()
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
      await pluginApi.updatePlugin(plugin.name)
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

  async function refreshWorkflowNodes() {
    try {
      const workflowStore = useWorkflowStore()
      await workflowStore.fetchNodeTypes()
    } catch (e) {
      console.error('[plugin-store] refreshWorkflowNodes failed:', e)
    }
  }

  async function simulateInstallProgress(task: InstallTask, plugin: Plugin) {
    for (let i = 0; i < task.steps.length; i++) {
      if ((task as InstallTask).status === 'cancelled') break

      const step = task.steps[i]
      step.status = 'in-progress'
      step.logs = [{ timestamp: new Date().toISOString(), level: 'info', message: `正在${step.name}...` }]

      const duration = 300 + Math.random() * 500
      await new Promise(resolve => setTimeout(resolve, duration))

      if ((task as InstallTask).status === 'cancelled') break

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
    refreshWorkflowNodes,
  }
})