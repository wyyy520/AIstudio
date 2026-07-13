<template>
  <aside class="plugin-sidebar">
    <div class="plugin-sidebar-header">
      <svg class="plugin-sidebar-header-icon" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <path d="M3 3h7v7H3V3zm0 11h7v7H3v-7zm11-11h7v7h-7V3zm0 11h7v7h-7v-7z" />
      </svg>
      <span class="plugin-sidebar-header-title">Plugin Explorer</span>
    </div>

    <div class="plugin-sidebar-search">
      <PluginSearch
        :model-value="searchQuery"
        placeholder="Filter plugins..."
        @update:model-value="handleSearch"
        @search="handleSearch"
        @clear="handleSearch('')"
      />
    </div>

    <div class="plugin-sidebar-categories">
      <div
        class="plugin-sidebar-category"
        :class="{ 'is-active': activeCategory === 'all' }"
        @click="handleCategorySelect('all')"
      >
        <div class="category-item">
          <svg class="category-item-icon" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M4 4h16v16H4z" />
          </svg>
          <span class="category-item-label">All Plugins</span>
          <span class="category-item-count">{{ totalPlugins }}</span>
        </div>
      </div>

      <div class="plugin-sidebar-section-title">Categories</div>

      <div
        v-for="cat in categories"
        :key="cat.category"
        class="plugin-sidebar-category"
        :class="{ 'is-active': activeCategory === cat.category }"
        @click="handleCategorySelect(cat.category)"
      >
        <div class="category-item">
          <svg class="category-item-icon" :style="{ color: cat.color }" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path :d="cat.icon" />
          </svg>
          <span class="category-item-label">{{ cat.label }}</span>
          <span class="category-item-count">{{ getCategoryCount(cat.category) }}</span>
        </div>
        <div class="category-item-installed">
          {{ getCategoryInstalledCount(cat.category) }} installed
        </div>
      </div>
    </div>

    <div class="plugin-sidebar-footer">
      <div class="plugin-sidebar-stat">
        <span class="stat-value">{{ installedCount }}</span>
        <span class="stat-label">Installed</span>
      </div>
      <div class="plugin-sidebar-stat">
        <span class="stat-value">{{ totalPlugins }}</span>
        <span class="stat-label">Total</span>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { PluginCategory, PluginCategoryGroup, Plugin } from '@/pages/PluginStore/types'
import PluginSearch from './PluginSearch.vue'

interface Props {
  categories: PluginCategoryGroup[]
  plugins: Plugin[]
  activeCategory: PluginCategory | 'all'
  searchQuery: string
  installedCount: number
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:activeCategory': [value: PluginCategory | 'all']
  'update:searchQuery': [value: string]
}>()

const totalPlugins = computed(() => props.plugins.length)

function getCategoryCount(category: PluginCategory): number {
  return props.plugins.filter(p => p.category === category).length
}

function getCategoryInstalledCount(category: PluginCategory): number {
  return props.plugins.filter(p => p.category === category && p.status === 'installed').length
}

function handleCategorySelect(category: PluginCategory | 'all') {
  emit('update:activeCategory', category)
}

function handleSearch(query: string) {
  emit('update:searchQuery', query)
}
</script>

<style scoped>
.plugin-sidebar {
  display: flex;
  flex-direction: column;
  width: 240px;
  min-width: 240px;
  height: 100%;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-subtle);
  overflow: hidden;
}

.plugin-sidebar-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-4) var(--spacing-4) var(--spacing-3);
  border-bottom: 1px solid var(--border-subtle);
}

.plugin-sidebar-header-icon {
  color: var(--primary);
  flex-shrink: 0;
}

.plugin-sidebar-header-title {
  font-size: var(--text-body);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.plugin-sidebar-search {
  padding: var(--spacing-3) var(--spacing-3) var(--spacing-2);
}

.plugin-sidebar-categories {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-2) var(--spacing-2);
}

.plugin-sidebar-category {
  position: relative;
  cursor: pointer;
  border-radius: var(--radius-md);
  transition: background var(--transition-fast);
  margin-bottom: 2px;
}

.plugin-sidebar-category:hover {
  background: var(--bg-hover);
}

.plugin-sidebar-category.is-active {
  background: var(--bg-active);
}

.plugin-sidebar-category.is-active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 18px;
  border-radius: 0 2px 2px 0;
  background: var(--primary);
}

.category-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  height: 32px;
  padding: 0 var(--spacing-3);
}

.category-item-icon {
  flex-shrink: 0;
  color: var(--text-tertiary);
}

.plugin-sidebar-category.is-active .category-item-icon {
  opacity: 1;
}

.category-item-label {
  flex: 1;
  font-size: var(--text-body-sm);
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.plugin-sidebar-category.is-active .category-item-label {
  color: var(--text-primary);
}

.category-item-count {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  min-width: 18px;
  text-align: center;
}

.category-item-installed {
  font-size: 10px;
  color: var(--text-tertiary);
  padding: 0 var(--spacing-3) var(--spacing-1) 32px;
  margin-top: -4px;
}

.plugin-sidebar-section-title {
  font-size: 10px;
  font-weight: var(--font-semibold);
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: var(--spacing-3) var(--spacing-3) var(--spacing-1);
}

.plugin-sidebar-footer {
  display: flex;
  align-items: center;
  gap: var(--spacing-4);
  padding: var(--spacing-3) var(--spacing-4);
  border-top: 1px solid var(--border-subtle);
}

.plugin-sidebar-stat {
  display: flex;
  align-items: baseline;
  gap: var(--spacing-1);
}

.stat-value {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  font-family: var(--font-family-mono);
}

.stat-label {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}
</style>