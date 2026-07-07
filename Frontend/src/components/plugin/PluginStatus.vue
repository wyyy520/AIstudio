<template>
  <span :class="badgeClasses">
    <svg
      v-if="statusIcon"
      class="status-badge-icon"
      :class="{ 'is-spinning': isSpinning }"
      viewBox="0 0 24 24"
      width="12"
      height="12"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
      stroke-linecap="round"
      stroke-linejoin="round"
    >
      <path :d="statusIcon" />
    </svg>
    <span class="status-badge-text">{{ statusLabel }}</span>
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { PluginStatus } from '@/pages/PluginStore/types'

interface Props {
  status: PluginStatus
}

const props = defineProps<Props>()

const statusConfig: Record<PluginStatus, { label: string; icon: string; bg: string; color: string; spin: boolean }> = {
  'installed': { label: 'Installed', icon: 'M20 6L9 17l-5-5', bg: 'var(--success-bg)', color: 'var(--success)', spin: false },
  'not-installed': { label: 'Not Installed', icon: 'M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3', bg: 'var(--bg-hover)', color: 'var(--text-secondary)', spin: false },
  'installing': { label: 'Installing', icon: 'M21 12a9 9 0 1 1-6.219-8.56', bg: 'var(--primary-bg)', color: 'var(--primary)', spin: true },
  'updating': { label: 'Updating', icon: 'M1 4v6h6M23 20v-6h-6 M20.49 9A9 9 0 0 0 5.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 0 1 3.51 15', bg: 'var(--info-bg)', color: 'var(--info)', spin: true },
  'error': { label: 'Error', icon: 'M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z M12 9v4M12 17h.01', bg: 'var(--error-bg)', color: 'var(--error)', spin: false },
}

const config = computed(() => statusConfig[props.status])

const badgeClasses = computed(() => [
  'plugin-status-badge',
  `plugin-status-badge--${props.status}`,
])

const statusIcon = computed(() => config.value.icon)
const statusLabel = computed(() => config.value.label)
const isSpinning = computed(() => config.value.spin)
</script>

<style scoped>
.plugin-status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border-radius: 8px;
  padding: 4px 10px;
  font-size: var(--text-caption);
  font-weight: 500;
  white-space: nowrap;
  flex-shrink: 0;
}

.plugin-status-badge--installed {
  background: var(--success-bg);
  color: var(--success);
}

.plugin-status-badge--not-installed {
  background: var(--bg-hover);
  color: var(--text-secondary);
}

.plugin-status-badge--installing {
  background: var(--primary-bg);
  color: var(--primary);
}

.plugin-status-badge--updating {
  background: var(--info-bg);
  color: var(--info);
}

.plugin-status-badge--error {
  background: var(--error-bg);
  color: var(--error);
}

.status-badge-icon {
  flex-shrink: 0;
}

.status-badge-icon.is-spinning {
  animation: spin 1s linear infinite;
}

.status-badge-text {
  line-height: 1;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>