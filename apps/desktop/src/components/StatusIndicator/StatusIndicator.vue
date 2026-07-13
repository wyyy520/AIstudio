<template>
  <span :class="indicatorClasses">
    <span class="status-dot"></span>
    <span v-if="label" class="status-label">{{ label }}</span>
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

type StatusType = 'success' | 'warning' | 'error' | 'info' | 'neutral'

interface Props {
  type?: StatusType
  label?: string
}

const props = withDefaults(defineProps<Props>(), {
  type: 'neutral',
})

const indicatorClasses = computed(() => {
  return [
    'status-indicator',
    `status-indicator--${props.type}`,
  ]
})
</script>

<style scoped>
.status-indicator {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-2);
  flex-shrink: 0;
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-label {
  font-size: var(--text-body-sm);
  line-height: var(--leading-body-sm);
  color: var(--text-secondary);
  white-space: nowrap;
}

/* ===== 颜色 ===== */
.status-indicator--success .status-dot {
  background: var(--success);
}

.status-indicator--success .status-label {
  color: var(--success);
}

.status-indicator--warning .status-dot {
  background: var(--warning);
}

.status-indicator--warning .status-label {
  color: var(--warning);
}

.status-indicator--error .status-dot {
  background: var(--error);
}

.status-indicator--error .status-label {
  color: var(--error);
}

.status-indicator--info .status-dot {
  background: var(--info);
}

.status-indicator--info .status-label {
  color: var(--info);
}

.status-indicator--neutral .status-dot {
  background: var(--neutral);
}

.status-indicator--neutral .status-label {
  color: var(--text-secondary);
}
</style>