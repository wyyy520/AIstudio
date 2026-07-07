<template>
  <div class="action-buttons">
    <button
      v-for="action in actions"
      :key="action.id"
      class="action-btn"
      @click="$emit('action', action)"
    >
      <svg class="action-btn-icon" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <path :d="action.icon" />
      </svg>
      <span class="action-btn-label">{{ action.label }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import type { QuickAction } from '../../types'

defineProps<{
  actions: QuickAction[]
}>()

defineEmits<{
  action: [action: QuickAction]
}>()
</script>

<style scoped>
.action-buttons {
  display: flex;
  gap: var(--spacing-2);
  padding: var(--spacing-2) 0;
  flex-wrap: wrap;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  height: 28px;
  padding: 0 var(--spacing-2);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: var(--text-caption);
  font-family: var(--font-family-sans);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.action-btn:hover {
  border-color: var(--border-default);
  color: var(--text-primary);
  background: var(--bg-hover);
}

.action-btn:active {
  transform: scale(0.97);
}

.action-btn-icon {
  flex-shrink: 0;
}
</style>
