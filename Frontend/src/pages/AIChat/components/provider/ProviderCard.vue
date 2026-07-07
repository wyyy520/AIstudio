<template>
  <div
    class="provider-card"
    :class="{ 'is-active': isActive }"
    @click="$emit('click')"
  >
    <div class="provider-card-header">
      <div class="provider-card-icon">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path :d="provider.icon" />
        </svg>
      </div>
      <div class="provider-card-info">
        <div class="provider-card-name">{{ provider.name }}</div>
        <div class="provider-card-models">{{ provider.models.length }} 个模型</div>
      </div>
      <StatusDot :status="provider.status" />
    </div>
    <div class="provider-card-footer">
      <span class="provider-card-status-text">{{ statusText }}</span>
      <button class="provider-card-settings" @click.stop="$emit('settings')" title="设置">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="3" />
          <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
        </svg>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import StatusDot from '../shared/StatusDot.vue'
import type { AIProvider } from '../../types'

const props = defineProps<{
  provider: AIProvider
  isActive?: boolean
}>()

defineEmits<{
  click: []
  settings: []
}>()

const statusText = computed(() => {
  const map: Record<string, string> = {
    connected: 'Connected',
    disconnected: 'Disconnected',
    error: 'Error',
  }
  return map[props.provider.status] || ''
})
</script>

<style scoped>
.provider-card {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: var(--spacing-3);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.provider-card:hover {
  border-color: var(--border-default);
  background: var(--bg-hover);
}

.provider-card.is-active {
  border-color: var(--primary);
  background: var(--primary-bg);
}

.provider-card-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.provider-card-icon {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-md);
  background: var(--bg-active);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: var(--text-secondary);
}

.provider-card-info {
  flex: 1;
  min-width: 0;
}

.provider-card-name {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.provider-card-models {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.provider-card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: var(--spacing-2);
}

.provider-card-status-text {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  font-variant-numeric: tabular-nums;
}

.provider-card-settings {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border: none;
  background: transparent;
  color: var(--text-tertiary);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.provider-card-settings:hover {
  background: var(--bg-active);
  color: var(--text-primary);
}
</style>
