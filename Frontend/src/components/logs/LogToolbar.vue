<template>
  <div class="log-toolbar">
    <div class="toolbar-left">
      <div class="toolbar-search">
        <svg class="search-icon" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="11" cy="11" r="8" /><path d="m21 21-4.3-4.3" />
        </svg>
        <input
          type="text"
          class="search-input"
          placeholder="Search logs..."
          :value="searchQuery"
          @input="$emit('update:searchQuery', ($event.target as HTMLInputElement).value)"
        />
      </div>
      <div class="toolbar-filter">
        <button class="filter-btn" @click="showFilter = !showFilter">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3" />
          </svg>
          Filter
          <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="m6 9 6 6 6-6" />
          </svg>
        </button>
        <div v-if="showFilter" class="filter-dropdown">
          <LogFilter :model-value="filterLevel" @update:model-value="handleFilterChange" />
        </div>
      </div>
    </div>

    <div class="toolbar-right">
      <button class="toolbar-btn" title="Clear Logs" @click="$emit('clear')">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M3 6h18" /><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" /><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
        </svg>
      </button>
      <button class="toolbar-btn" title="Export Logs" @click="$emit('export')">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" /><path d="m7 10 5 5 5-5" /><path d="M12 15V3" />
        </svg>
      </button>
      <button class="toolbar-btn analyze-btn" title="AI Analyze" :disabled="!selectedTaskId || isAnalyzing" @click="$emit('analyze')">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 2a10 10 0 1 0 10 10 4 4 0 0 1-5-5 4 4 0 0 1-5-5" /><path d="M8.5 8.5v.01" /><path d="M16 15.5v.01" />
        </svg>
        AI Analyze
      </button>
      <button class="toolbar-btn fix-btn" title="Auto Fix" :disabled="!selectedTaskId" @click="$emit('analyze')">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z" />
        </svg>
        Auto Fix
      </button>
    </div>
  </div>

  <div v-if="showFilter" class="filter-backdrop" @click="showFilter = false"></div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { FilterLevel } from '@/pages/Logs/types'
import LogFilter from './LogFilter.vue'

interface Props {
  searchQuery: string
  filterLevel: FilterLevel
  selectedTaskId: string | null
  isAnalyzing: boolean
}

defineProps<Props>()

const emit = defineEmits<{
  'update:searchQuery': [value: string]
  'update:filterLevel': [value: FilterLevel]
  clear: []
  export: []
  analyze: []
}>()

const showFilter = ref(false)

function handleFilterChange(value: FilterLevel) {
  emit('update:filterLevel', value)
  showFilter.value = false
}
</script>

<style scoped>
.log-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 48px;
  padding: 0 var(--spacing-4);
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-subtle);
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.toolbar-search {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  height: 32px;
  padding: 0 var(--spacing-3);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  width: 240px;
  transition: border-color var(--transition-fast);
}

.toolbar-search:focus-within {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px var(--primary-bg);
}

.search-icon {
  flex-shrink: 0;
  color: var(--text-tertiary);
}

.search-input {
  flex: 1;
  font-size: var(--text-body-sm);
  color: var(--text-primary);
  background: transparent;
  border: none;
  outline: none;
}

.search-input::placeholder {
  color: var(--text-tertiary);
}

.toolbar-filter {
  position: relative;
}

.filter-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  height: 32px;
  padding: 0 var(--spacing-3);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: var(--text-body-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.filter-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.filter-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  z-index: 50;
}

.filter-backdrop {
  position: fixed;
  inset: 0;
  z-index: 40;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.toolbar-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  height: 32px;
  padding: 0 var(--spacing-3);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: var(--text-body-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.toolbar-btn:hover:not(:disabled) {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.toolbar-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.analyze-btn {
  background: var(--primary-bg);
  border-color: rgba(139, 92, 246, 0.2);
  color: var(--primary);
}

.analyze-btn:hover:not(:disabled) {
  background: rgba(139, 92, 246, 0.2);
}

.fix-btn {
  background: var(--primary);
  border-color: var(--primary);
  color: white;
}

.fix-btn:hover:not(:disabled) {
  background: var(--primary-hover);
}
</style>