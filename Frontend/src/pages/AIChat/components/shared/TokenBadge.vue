<template>
  <span class="token-badge">
    <span class="token-badge-item" v-if="duration">{{ formatDuration(duration) }}</span>
    <span class="token-badge-sep" v-if="duration && usage">·</span>
    <span class="token-badge-item" v-if="usage">{{ usage.total.toLocaleString() }} tokens</span>
  </span>
</template>

<script setup lang="ts">
import type { TokenUsage } from '../../types'

defineProps<{
  usage?: TokenUsage
  duration?: number
}>()

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(1)}s`
}
</script>

<style scoped>
.token-badge {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-1);
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  font-variant-numeric: tabular-nums;
}

.token-badge-sep {
  opacity: 0.5;
}
</style>
