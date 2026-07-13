<template>
  <div
    :class="['plugin-card', { 'is-selected': selected, 'is-installed': plugin.status === 'installed' }]"
    @click="$emit('click')"
  >
    <div class="plugin-card-icon-wrap">
      <svg
        class="plugin-card-icon"
        viewBox="0 0 24 24"
        width="24"
        height="24"
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path :d="plugin.icon" />
      </svg>
    </div>
    <div class="plugin-card-info">
      <div class="plugin-card-name">{{ plugin.name }}</div>
      <div class="plugin-card-desc">{{ plugin.description }}</div>
      <div class="plugin-card-meta">
        <span class="plugin-card-category" :style="{ color: categoryColor }">{{ categoryLabel }}</span>
        <span class="plugin-card-sep">·</span>
        <span class="plugin-card-version">v{{ plugin.version }}</span>
        <span class="plugin-card-sep">·</span>
        <span class="plugin-card-author">{{ plugin.author }}</span>
      </div>
      <div class="plugin-card-stats">
        <span v-if="plugin.size" class="plugin-card-stat">
          <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
          </svg>
          {{ plugin.size }}
        </span>
        <span v-if="plugin.downloads" class="plugin-card-stat">
          <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3" />
          </svg>
          {{ formatDownloads(plugin.downloads) }}
        </span>
      </div>
    </div>
    <PluginStatus :status="plugin.status" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Plugin, PluginCategory } from '@/pages/PluginStore/types'
import PluginStatus from './PluginStatus.vue'

interface Props {
  plugin: Plugin
  selected?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  selected: false,
})

defineEmits<{
  click: []
}>()

const categoryLabels: Record<PluginCategory, string> = {
  vision: 'Vision AI',
  nlp: 'NLP',
  timeseries: 'Time Series',
  speech: 'Speech',
  simulation: 'Simulation',
  system: 'System',
  mcp: 'MCP',
}

const categoryColors: Record<PluginCategory, string> = {
  vision: 'var(--vision)',
  nlp: 'var(--nlp)',
  timeseries: 'var(--timeseries)',
  speech: 'var(--nlp)',
  simulation: 'var(--simulation)',
  system: 'var(--system)',
  mcp: 'var(--mcp)',
}

const categoryLabel = computed(() => categoryLabels[props.plugin.category])
const categoryColor = computed(() => categoryColors[props.plugin.category])

function formatDownloads(count: number): string {
  if (count >= 1000000) {
    return `${(count / 1000000).toFixed(1)}M`
  }
  if (count >= 1000) {
    return `${(count / 1000).toFixed(1)}K`
  }
  return count.toString()
}
</script>

<style scoped>
.plugin-card {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  padding: var(--spacing-3) var(--spacing-4);
  border-radius: var(--radius-xl);
  background: var(--bg-tertiary);
  cursor: pointer;
  transition: all var(--transition-fast);
  border: 1px solid transparent;
}

.plugin-card:hover {
  background: var(--bg-hover);
  box-shadow: var(--shadow-sm);
}

.plugin-card.is-selected {
  background: var(--bg-active);
  border-color: var(--primary);
  box-shadow: 0 0 0 1px var(--primary-bg);
}

.plugin-card-icon-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: var(--radius-lg);
  background: var(--bg-secondary);
  flex-shrink: 0;
}

.plugin-card-icon {
  color: var(--text-secondary);
}

.plugin-card.is-selected .plugin-card-icon {
  color: var(--primary);
}

.plugin-card-info {
  flex: 1;
  min-width: 0;
}

.plugin-card-name {
  font-size: var(--text-body);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: var(--leading-body);
}

.plugin-card-meta {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

.plugin-card-desc {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  line-height: var(--leading-caption);
  margin-top: 2px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
}

.plugin-card-category {
  font-weight: 500;
}

.plugin-card-sep {
  color: var(--text-tertiary);
}

.plugin-card-version {
  color: var(--text-tertiary);
}

.plugin-card-author {
  color: var(--text-tertiary);
}

.plugin-card-stats {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  margin-top: 2px;
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.plugin-card-stat {
  display: inline-flex;
  align-items: center;
  gap: 3px;
}
</style>