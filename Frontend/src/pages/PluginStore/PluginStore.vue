<template>
  <div class="plugin-market">
    <PluginSidebar
      :categories="store.categories"
      :plugins="store.plugins"
      :active-category="store.activeCategory"
      :search-query="store.searchQuery"
      :installed-count="store.installedCount"
      @update:active-category="store.setCategory"
      @update:search-query="store.setSearchQuery"
    />

    <div class="plugin-market-main">
      <div class="plugin-market-toolbar">
        <div class="toolbar-left">
          <h1 class="toolbar-title">
            {{ activeCategoryLabel }}
          </h1>
          <span class="toolbar-count">{{ store.filteredPlugins.length }} plugins</span>
        </div>
        <div class="toolbar-right">
          <PluginSearch
            :model-value="store.searchQuery"
            placeholder="Search plugins..."
            @update:model-value="store.setSearchQuery"
            @search="store.setSearchQuery"
            @clear="store.setSearchQuery('')"
          />
        </div>
      </div>

      <div class="plugin-market-list">
        <TransitionGroup name="card-list" tag="div" class="plugin-card-grid">
          <PluginCard
            v-for="plugin in store.filteredPlugins"
            :key="plugin.id"
            :plugin="plugin"
            :selected="plugin.id === store.selectedPluginId"
            @click="store.selectPlugin(plugin.id)"
          />
        </TransitionGroup>

        <div v-if="store.filteredPlugins.length === 0" class="plugin-market-empty">
          <svg viewBox="0 0 24 24" width="40" height="40" fill="none" stroke="var(--text-tertiary)" stroke-width="1">
            <circle cx="11" cy="11" r="8" />
            <path d="M21 21l-4.35-4.35" />
          </svg>
          <p class="empty-title">No plugins found</p>
          <p class="empty-desc">Try adjusting your search or category filter</p>
        </div>
      </div>
    </div>

    <div class="plugin-market-detail">
      <PluginDetail
        :plugin="store.selectedPlugin"
        :is-installing="store.isInstalling"
        @install="handleInstall"
        @update="handleUpdate"
        @remove="handleRemove"
      />
    </div>

    <Transition name="overlay">
      <div v-if="store.installTask && store.installTask.status === 'running'" class="plugin-market-overlay" @click.self>
        <div class="plugin-market-install-panel">
          <PluginInstallTask
            :task="store.installTask"
            @cancel="store.cancelInstall"
            @done="store.clearInstallTask"
          />
        </div>
      </div>
    </Transition>

    <Transition name="overlay">
      <div v-if="store.installTask && (store.installTask.status === 'completed' || store.installTask.status === 'cancelled')" class="plugin-market-overlay" @click.self="store.clearInstallTask">
        <div class="plugin-market-install-panel">
          <PluginInstallTask
            :task="store.installTask"
            @cancel="store.cancelInstall"
            @done="store.clearInstallTask"
          />
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { usePluginStore } from '@/store/plugin'
import type { PluginCategory } from './types'
import PluginSidebar from '@/components/plugin/PluginSidebar.vue'
import PluginCard from '@/components/plugin/PluginCard.vue'
import PluginDetail from '@/components/plugin/PluginDetail.vue'
import PluginSearch from '@/components/plugin/PluginSearch.vue'
import PluginInstallTask from '@/components/plugin/PluginInstallTask.vue'

const store = usePluginStore()

onMounted(() => {
  store.fetchPlugins()
})

const categoryLabels: Record<PluginCategory | 'all', string> = {
  all: 'All Plugins',
  vision: 'AI Vision',
  nlp: 'NLP',
  timeseries: 'Time Series',
  speech: 'Speech',
  simulation: 'Simulation',
  system: 'System',
  mcp: 'MCP',
}

const activeCategoryLabel = computed(() => categoryLabels[store.activeCategory])

async function handleInstall(pluginId: string) {
  await store.installPluginAction(pluginId)
}

async function handleUpdate(pluginId: string) {
  await store.updatePluginAction(pluginId)
}

async function handleRemove(pluginId: string) {
  await store.uninstallPluginAction(pluginId)
}
</script>

<style scoped>
.plugin-market {
  display: flex;
  height: 100%;
  overflow: hidden;
  background: var(--bg-primary);
}

.plugin-market-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  border-right: 1px solid var(--border-subtle);
}

.plugin-market-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-4) var(--spacing-6);
  border-bottom: 1px solid var(--border-subtle);
  background: var(--bg-secondary);
}

.toolbar-left {
  display: flex;
  align-items: baseline;
  gap: var(--spacing-3);
}

.toolbar-title {
  font-size: var(--text-h2);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  margin: 0;
}

.toolbar-count {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.toolbar-right {
  width: 240px;
}

.plugin-market-list {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-4) var(--spacing-6);
}

.plugin-card-grid {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.plugin-market-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-3);
  padding: var(--spacing-12) 0;
  opacity: 0.5;
}

.empty-title {
  font-size: var(--text-h3);
  font-weight: var(--font-semibold);
  color: var(--text-secondary);
}

.empty-desc {
  font-size: var(--text-body-sm);
  color: var(--text-tertiary);
}

.plugin-market-detail {
  width: 380px;
  min-width: 380px;
  height: 100%;
  overflow: hidden;
  background: var(--bg-secondary);
}

.plugin-market-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.plugin-market-install-panel {
  width: 560px;
  max-height: 80vh;
  background: var(--bg-secondary);
  border-radius: var(--radius-2xl);
  border: 1px solid var(--border-default);
  box-shadow: var(--shadow-xl);
  overflow-y: auto;
}

.card-list-enter-active,
.card-list-leave-active {
  transition: all var(--transition-normal);
}

.card-list-enter-from {
  opacity: 0;
  transform: translateY(8px);
}

.card-list-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

.card-list-move {
  transition: transform var(--transition-normal);
}

.overlay-enter-active,
.overlay-leave-active {
  transition: opacity var(--transition-normal);
}

.overlay-enter-from,
.overlay-leave-to {
  opacity: 0;
}
</style>