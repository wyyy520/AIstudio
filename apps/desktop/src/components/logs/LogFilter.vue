<template>
  <div class="log-filter">
    <div class="filter-section">
      <span class="filter-label">日志级别</span>
      <div class="filter-options">
        <label
          v-for="opt in levelOptions"
          :key="opt.value"
          class="filter-option"
          :class="{ active: modelValue === opt.value }"
          @click="$emit('update:modelValue', opt.value)"
        >
          <span class="filter-dot" :style="{ background: opt.color }"></span>
          <span class="filter-text">{{ opt.label }}</span>
        </label>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { FilterLevel } from '@/pages/Logs/types'

interface Props {
  modelValue: FilterLevel
}

defineProps<Props>()

defineEmits<{
  'update:modelValue': [value: FilterLevel]
}>()

const levelOptions: Array<{ value: FilterLevel; label: string; color: string }> = [
  { value: 'all', label: '全部', color: 'var(--text-secondary)' },
  { value: 'error', label: '错误', color: 'var(--error)' },
  { value: 'warning', label: '警告', color: 'var(--warning)' },
  { value: 'info', label: '信息', color: 'var(--info)' },
  { value: 'debug', label: '调试', color: 'var(--text-tertiary)' },
]
</script>

<style scoped>
.log-filter {
  padding: var(--spacing-3);
  background: var(--bg-secondary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  min-width: 160px;
}

.filter-section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.filter-label {
  font-size: var(--text-caption);
  font-weight: var(--font-semibold);
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.filter-options {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.filter-option {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-1) var(--spacing-2);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: background var(--transition-fast);
}

.filter-option:hover {
  background: var(--bg-hover);
}

.filter-option.active {
  background: var(--bg-active);
}

.filter-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.filter-text {
  font-size: var(--text-body-sm);
  color: var(--text-primary);
}
</style>