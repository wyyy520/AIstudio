<template>
  <AppModal v-model:visible="showConfirmModal" title="确认删除" width="400px">
      <p>确定删除插件 <strong>"{{ deletingPlugin?.name }}"</strong> 吗？此操作不可撤销。</p>
      <template #footer>
        <AppButton type="secondary" size="small" @click="showConfirmModal = false">取消</AppButton>
        <AppButton type="danger" size="small" :loading="deleting" @click="handleDelete">确认删除</AppButton>
      </template>
    </AppModal>

    <div class="plugin-store">
    <div class="plugin-store__toolbar">
      <h3 class="plugin-store__title">插件中心</h3>
      <div class="plugin-store__search">
        <input v-model="searchText" placeholder="搜索插件..." class="search-input" />
      </div>
    </div>

    <div class="plugin-store__body">
      <aside class="plugin-sidebar">
        <div
          v-for="cat in categories"
          :key="cat.key"
          :class="['category-item', { 'category-item--active': activeCategory === cat.key }]"
          @click="activeCategory = cat.key"
        >
          <span :class="['category-dot', `category-dot--${cat.key}`]" />
          <span class="category-label">{{ cat.label }}</span>
          <span class="category-count">{{ getCategoryCount(cat.key) }}</span>
        </div>
        <div class="category-divider" />
        <div
          :class="['category-item', { 'category-item--active': activeCategory === 'installed' }]"
          @click="activeCategory = 'installed'"
        >
          <span class="category-dot category-dot--success" />
          <span class="category-label">已安装</span>
        </div>
      </aside>

      <main class="plugin-main">
        <div v-if="loading" class="plugin-loading">加载中...</div>
        <div v-else-if="filteredPlugins.length === 0" class="plugin-empty">没有找到插件</div>
        <div v-else class="plugin-grid">
          <div v-for="plugin in filteredPlugins" :key="plugin.name" class="plugin-card">
            <div class="plugin-card__header">
              <span :class="['plugin-card__dot', `plugin-card__dot--${plugin.category}`]" />
              <span class="plugin-card__name">{{ plugin.name }}</span>
              <AppTag :color="statusColor(plugin.status)" size="small">{{ plugin.status }}</AppTag>
            </div>
            <p class="plugin-card__desc">{{ plugin.description || '暂无描述' }}</p>
            <div class="plugin-card__meta">
              <span class="plugin-card__version">v{{ plugin.version }}</span>
              <span :class="['plugin-card__cat', `plugin-card__cat--${plugin.category}`]">{{ getCategoryLabel(plugin.category) }}</span>
            </div>
            <div class="plugin-card__actions">
              <AppButton
                v-if="plugin.status !== 'installed'"
                type="primary"
                size="small"
                :loading="installing === plugin.name"
                @click="installPluginByName(plugin.name)"
              >安装</AppButton>
              <template v-else>
                <AppButton type="secondary" size="small" @click="togglePlugin(plugin)">禁用</AppButton>
                <AppButton type="danger" size="small" @click="uninstallPluginByName(plugin.name)">删除</AppButton>
              </template>
            </div>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getPlugins, installPlugin, uninstallPlugin, enablePlugin, disablePlugin } from '@/api/plugin'
import AppButton from '@/components/AppButton.vue'
import AppTag from '@/components/AppTag.vue'
import AppModal from '@/components/AppModal.vue'

const searchText = ref('')
const activeCategory = ref('all')
const plugins = ref<any[]>([])
const loading = ref(false)
const installing = ref<string | null>(null)
const showConfirmModal = ref(false)
const deleting = ref(false)
const deletingPlugin = ref<any>(null)

const categories = [
  { key: 'all', label: '全部' },
  { key: 'vision', label: '视觉处理' },
  { key: 'nlp', label: '自然语言' },
  { key: 'timeseries', label: '时序处理' },
  { key: 'logic', label: '逻辑控制' },
  { key: 'system', label: '系统工具' },
  { key: 'mcp', label: 'MCP 服务' },
  { key: 'simulation', label: '仿真联动' },
  { key: 'agent', label: 'Agent' },
]

function getCategoryLabel(key: string) {
  return categories.find(c => c.key === key)?.label || key
}

function getCategoryCount(key: string) {
  if (key === 'all') return plugins.value.length
  return plugins.value.filter(p => p.category === key).length
}

const filteredPlugins = computed(() => {
  let list = plugins.value
  if (activeCategory.value === 'installed') {
    list = list.filter(p => p.status === 'installed')
  } else if (activeCategory.value !== 'all') {
    list = list.filter(p => p.category === activeCategory.value)
  }
  if (searchText.value) {
    const s = searchText.value.toLowerCase()
    list = list.filter(p => p.name.toLowerCase().includes(s) || (p.description || '').toLowerCase().includes(s))
  }
  return list
})

function statusColor(status: string) {
  if (status === 'installed') return 'success'
  if (status === 'error') return 'error'
  if (status === 'running') return 'warning'
  return 'default'
}

async function loadPlugins() {
  loading.value = true
  try {
    const res: any = await getPlugins()
    plugins.value = res.data || []
  } catch {
    plugins.value = []
  } finally {
    loading.value = false
  }
}

async function installPluginByName(name: string) {
  installing.value = name
  try {
    await installPlugin(name)
    await loadPlugins()
  } finally {
    installing.value = null
  }
}

async function uninstallPluginByName(name: string) {
  deletingPlugin.value = plugins.value.find(p => p.name === name)
  showConfirmModal.value = true
}

async function handleDelete() {
  if (!deletingPlugin.value) return
  deleting.value = true
  try {
    await uninstallPlugin(deletingPlugin.value.name)
    await loadPlugins()
    showConfirmModal.value = false
    deletingPlugin.value = null
  } finally {
    deleting.value = false
  }
}

async function togglePlugin(plugin: any) {
  if (plugin.enabled) {
    await disablePlugin(plugin.name)
  } else {
    await enablePlugin(plugin.name)
  }
  await loadPlugins()
}

onMounted(() => {
  loadPlugins()
})
</script>

<style scoped>
.plugin-store {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.plugin-store__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  border-bottom: 1px solid var(--border-subtle);
  background: var(--bg-secondary);
  flex-shrink: 0;
}

.plugin-store__title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.search-input {
  width: 240px;
  height: 32px;
  padding: 0 12px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-xs);
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
}

.search-input:focus {
  border-color: var(--primary);
}

.plugin-store__body {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.plugin-sidebar {
  width: 200px;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-subtle);
  padding: 12px 8px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex-shrink: 0;
  overflow-y: auto;
}

.category-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: var(--radius-xs);
  cursor: pointer;
  transition: all var(--transition-fast);
  font-size: 13px;
}

.category-item:hover {
  background: var(--bg-hover);
}

.category-item--active {
  background: var(--bg-active);
  color: var(--primary);
}

.category-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.category-dot--all { background: var(--text-tertiary); }
.category-dot--vision { background: var(--vision); }
.category-dot--nlp { background: var(--nlp); }
.category-dot--timeseries { background: var(--timeseries); }
.category-dot--logic { background: var(--logic); }
.category-dot--system { background: var(--system); }
.category-dot--mcp { background: var(--mcp); }
.category-dot--simulation { background: var(--simulation); }
.category-dot--agent { background: var(--agent); }
.category-dot--success { background: var(--success); }

.category-label {
  flex: 1;
  color: var(--text-secondary);
}

.category-item--active .category-label {
  color: var(--text-primary);
}

.category-count {
  font-size: 11px;
  color: var(--text-tertiary);
}

.category-divider {
  height: 1px;
  background: var(--border-subtle);
  margin: 8px 0;
}

.plugin-main {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}

.plugin-loading,
.plugin-empty {
  text-align: center;
  padding: 48px 0;
  color: var(--text-tertiary);
}

.plugin-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.plugin-card {
  background: var(--bg-tertiary);
  border-radius: var(--radius-md);
  padding: 16px;
  border: 1px solid var(--border-subtle);
  transition: all var(--transition-fast);
}

.plugin-card:hover {
  border-color: var(--border-default);
  box-shadow: var(--shadow);
}

.plugin-card__header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.plugin-card__dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.plugin-card__dot--vision { background: var(--vision); }
.plugin-card__dot--nlp { background: var(--nlp); }
.plugin-card__dot--timeseries { background: var(--timeseries); }
.plugin-card__dot--logic { background: var(--logic); }
.plugin-card__dot--system { background: var(--system); }
.plugin-card__dot--mcp { background: var(--mcp); }
.plugin-card__dot--simulation { background: var(--simulation); }
.plugin-card__dot--agent { background: var(--agent); }

.plugin-card__name {
  flex: 1;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.plugin-card__desc {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 10px;
  line-height: 1.5;
}

.plugin-card__meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.plugin-card__version {
  font-size: 12px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}

.plugin-card__cat {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
}

.plugin-card__cat--vision { background: rgba(236,72,153,0.1); color: var(--vision); }
.plugin-card__cat--nlp { background: rgba(59,130,246,0.1); color: var(--nlp); }
.plugin-card__cat--timeseries { background: rgba(16,185,129,0.1); color: var(--timeseries); }
.plugin-card__cat--logic { background: rgba(245,158,11,0.1); color: var(--logic); }
.plugin-card__cat--system { background: rgba(107,114,128,0.1); color: var(--system); }
.plugin-card__cat--mcp { background: rgba(139,92,246,0.1); color: var(--mcp); }
.plugin-card__cat--simulation { background: rgba(6,182,212,0.1); color: var(--simulation); }
.plugin-card__cat--agent { background: rgba(239,68,68,0.1); color: var(--agent); }

.plugin-card__actions {
  display: flex;
  gap: 6px;
}
</style>