<template>
  <div class="settings-section">
    <h2 class="section-title">插件管理</h2>
    <p class="section-desc">查看和管理已安装的插件。可启用、禁用或卸载插件。</p>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-state">
      <span class="loading-text">加载插件列表...</span>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="error-state">
      <span class="error-text">{{ error }}</span>
      <AppButton size="small" @click="fetchPlugins">重试</AppButton>
    </div>

    <!-- 空状态 -->
    <div v-else-if="plugins.length === 0" class="empty-state">
      <svg viewBox="0 0 24 24" width="32" height="32" fill="none" stroke="currentColor" stroke-width="1.5" class="empty-icon">
        <path d="M3 3h7v7H3V3zm0 11h7v7H3v-7zm11-11h7v7h-7V3zm0 11h7v7h-7v-7z" />
      </svg>
      <span class="empty-text">暂无已安装的插件</span>
      <span class="empty-hint">前往插件市场安装插件</span>
    </div>

    <!-- 插件列表 -->
    <div v-else class="plugin-list">
      <div
        v-for="plugin in plugins"
        :key="plugin.name"
        class="plugin-card"
      >
        <div class="plugin-info">
          <div class="plugin-header">
            <span class="plugin-name">{{ plugin.name }}</span>
            <AppTag
              :color="statusColor(plugin.status)"
              size="small"
            >
              {{ statusLabel(plugin.status) }}
            </AppTag>
          </div>
          <div class="plugin-meta">
            <span class="plugin-version">v{{ plugin.version }}</span>
            <span v-if="plugin.description" class="plugin-desc">{{ plugin.description }}</span>
          </div>
        </div>
        <div class="plugin-actions">
          <AppButton
            v-if="plugin.status === 'enabled' || plugin.status === 'installed'"
            size="small"
            type="outline"
            @click="handleDisable(plugin)"
          >
            禁用
          </AppButton>
          <AppButton
            v-else-if="plugin.status === 'disabled'"
            size="small"
            type="primary"
            @click="handleEnable(plugin)"
          >
            启用
          </AppButton>
          <AppButton
            size="small"
            type="text"
            :disabled="plugin.status === 'enabled'"
            @click="handleUninstall(plugin)"
          >
            卸载
          </AppButton>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getPlugins, enablePlugin, disablePlugin } from '@/api/plugin'
import type { ApiPlugin } from '@/api/plugin'
import AppButton from '@/components/AppButton/AppButton.vue'
import AppTag from '@/components/AppTag/AppTag.vue'

const plugins = ref<ApiPlugin[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

const statusLabels: Record<string, string> = {
  enabled: '已启用',
  disabled: '已禁用',
  installed: '已安装',
  error: '异常',
}

const statusColors: Record<string, 'success' | 'warning' | 'error' | 'info'> = {
  enabled: 'success',
  disabled: 'warning',
  installed: 'info',
  error: 'error',
}

function statusLabel(status: string): string {
  return statusLabels[status] || status
}

function statusColor(status: string): 'success' | 'warning' | 'error' | 'info' {
  return statusColors[status] || 'default'
}

async function fetchPlugins() {
  loading.value = true
  error.value = null
  try {
    plugins.value = await getPlugins()
  } catch (e) {
    error.value = '加载插件列表失败'
    console.error('[settings] fetch plugins failed:', e)
  } finally {
    loading.value = false
  }
}

async function handleEnable(plugin: ApiPlugin) {
  try {
    await enablePlugin(plugin.name)
    plugin.status = 'enabled'
  } catch (e) {
    console.error('[settings] enable plugin failed:', e)
  }
}

async function handleDisable(plugin: ApiPlugin) {
  try {
    await disablePlugin(plugin.name)
    plugin.status = 'disabled'
  } catch (e) {
    console.error('[settings] disable plugin failed:', e)
  }
}

async function handleUninstall(plugin: ApiPlugin) {
  // Use the new API
  try {
    const { uninstallPlugin } = await import('@/api/plugin')
    await uninstallPlugin(plugin.name)
    plugins.value = plugins.value.filter(p => p.name !== plugin.name)
  } catch (e) {
    console.error('[settings] uninstall plugin failed:', e)
  }
}

onMounted(fetchPlugins)
</script>

<style scoped>
.settings-section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-6);
}

.section-title {
  font-size: var(--text-h2);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  line-height: var(--leading-h2);
}

.section-desc {
  font-size: var(--text-body);
  color: var(--text-secondary);
  line-height: var(--leading-body);
  margin-top: -12px;
}

/* Loading / Error / Empty */
.loading-state,
.error-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-3);
  padding: var(--spacing-12) var(--spacing-4);
  background: var(--bg-tertiary);
  border-radius: var(--radius-xl);
}

.loading-text,
.error-text,
.empty-text {
  font-size: var(--text-body);
  color: var(--text-secondary);
}

.empty-icon {
  color: var(--text-tertiary);
  opacity: 0.5;
}

.empty-hint {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

/* Plugin list */
.plugin-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.plugin-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-3) var(--spacing-4);
  background: var(--bg-tertiary);
  border-radius: var(--radius-lg);
  transition: background var(--transition-fast);
}

.plugin-card:hover {
  background: var(--bg-hover);
}

.plugin-info {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1);
  min-width: 0;
}

.plugin-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.plugin-name {
  font-size: var(--text-body);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  line-height: var(--leading-body);
}

.plugin-meta {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.plugin-version {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  font-family: var(--font-family-mono);
}

.plugin-desc {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.plugin-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  flex-shrink: 0;
}
</style>