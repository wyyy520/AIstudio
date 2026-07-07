<template>
  <span class="status-dot" :class="`status-dot--${status}`" :title="statusLabel" />
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ status: string }>()

const statusLabel = computed(() => {
  const map: Record<string, string> = {
    connected: '已连接',
    disconnected: '未连接',
    error: '连接错误',
    active: '活跃',
    inactive: '未激活',
  }
  return map[props.status] || props.status
})
</script>

<style scoped>
.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-dot--connected {
  background: var(--success);
  box-shadow: 0 0 6px rgba(34, 197, 94, 0.4);
}

.status-dot--disconnected {
  background: var(--text-disabled);
}

.status-dot--error {
  background: var(--error);
  box-shadow: 0 0 6px rgba(239, 68, 68, 0.4);
}

.status-dot--active {
  background: var(--success);
  box-shadow: 0 0 6px rgba(34, 197, 94, 0.4);
}

.status-dot--inactive {
  background: var(--text-disabled);
}
</style>
