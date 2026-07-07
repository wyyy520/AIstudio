<template>
  <div class="provider-panel">
    <div class="provider-panel-header">
      <span class="provider-panel-title">AI Provider</span>
      <button class="provider-panel-history-btn" @click="$emit('toggle-history')" title="历史记录">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10" />
          <polyline points="12 6 12 12 16 14" />
        </svg>
      </button>
    </div>
    <div class="provider-panel-list">
      <ProviderCard
        v-for="provider in providers"
        :key="provider.id"
        :provider="provider"
        :is-active="provider.id === activeProviderId"
        @click="$emit('select-provider', provider.id)"
        @settings="$emit('settings', provider)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import ProviderCard from './ProviderCard.vue'
import type { AIProvider } from '../../types'

defineProps<{
  providers: AIProvider[]
  activeProviderId?: string
}>()

defineEmits<{
  'select-provider': [id: string]
  'settings': [provider: AIProvider]
  'toggle-history': []
}>()
</script>

<style scoped>
.provider-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-subtle);
}

.provider-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-3) var(--spacing-3) var(--spacing-2);
  flex-shrink: 0;
}

.provider-panel-title {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-secondary);
}

.provider-panel-history-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  color: var(--text-tertiary);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.provider-panel-history-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.provider-panel-list {
  flex: 1;
  overflow-y: auto;
  padding: 0 var(--spacing-2) var(--spacing-2);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}
</style>
