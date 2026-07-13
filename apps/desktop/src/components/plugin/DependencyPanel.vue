<template>
  <div class="dependency-panel">
    <div class="dependency-header">
      <svg class="dependency-header-icon" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <path d="M16.5 9.4l-9-5.19M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
      </svg>
      <span class="dependency-header-title">Dependencies</span>
    </div>
    <div class="dependency-list">
      <div
        v-for="dep in dependencies"
        :key="dep.name"
        class="dependency-item"
        :class="`dep-status--${dep.status}`"
      >
        <span class="dep-status-icon">
          <svg v-if="dep.status === 'satisfied'" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="var(--success)" stroke-width="2"><path d="M20 6L9 17l-5-5" /></svg>
          <svg v-else-if="dep.status === 'not-installed'" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="var(--error)" stroke-width="2"><path d="M18 6 6 18M6 6l12 12" /></svg>
          <svg v-else-if="dep.status === 'version-mismatch'" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="var(--warning)" stroke-width="2"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z M12 9v4M12 17h.01" /></svg>
          <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="var(--primary)" stroke-width="2" class="is-spinning"><path d="M21 12a9 9 0 1 1-6.219-8.56" /></svg>
        </span>
        <span class="dep-name">{{ dep.name }}</span>
        <span class="dep-version-req">{{ dep.versionRequired }}</span>
        <span class="dep-status-text">
          <template v-if="dep.status === 'satisfied'">{{ dep.versionInstalled }} Installed</template>
          <template v-else-if="dep.status === 'not-installed'">Not Installed</template>
          <template v-else-if="dep.status === 'version-mismatch'">{{ dep.versionInstalled }} Mismatch</template>
          <template v-else>Checking...</template>
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Dependency } from '@/pages/PluginStore/types'

interface Props {
  dependencies: Dependency[]
}

defineProps<Props>()
</script>

<style scoped>
.dependency-panel {
  background: var(--bg-tertiary);
  border-radius: var(--radius-xl);
  padding: var(--spacing-4);
}

.dependency-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-3);
}

.dependency-header-icon {
  color: var(--text-tertiary);
  flex-shrink: 0;
}

.dependency-header-title {
  font-size: var(--text-h3);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.dependency-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.dependency-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  height: 32px;
  padding: 0 var(--spacing-3);
  border-radius: var(--radius-sm);
  font-size: var(--text-body-sm);
}

.dependency-item:nth-child(odd) {
  background: var(--bg-tertiary);
}

.dependency-item:nth-child(even) {
  background: var(--bg-secondary);
}

.dep-status-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  flex-shrink: 0;
}

.dep-status-icon .is-spinning {
  animation: spin 1s linear infinite;
}

.dep-name {
  width: 100px;
  color: var(--text-primary);
  flex-shrink: 0;
}

.dep-version-req {
  width: 60px;
  color: var(--text-tertiary);
  flex-shrink: 0;
  font-family: var(--font-family-mono);
  font-size: var(--text-caption);
}

.dep-status-text {
  color: var(--text-secondary);
  font-size: var(--text-caption);
}

.dep-status--satisfied .dep-status-text {
  color: var(--success);
}

.dep-status--not-installed .dep-status-text {
  color: var(--error);
}

.dep-status--version-mismatch .dep-status-text {
  color: var(--warning);
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>