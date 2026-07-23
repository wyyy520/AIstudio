<template>
  <div class="project-status">
    <div v-if="project" class="status-content">
      <div class="status-item">
        <span class="status-label">Project</span>
        <span class="status-value">{{ project.name }}</span>
      </div>
      <div class="status-separator"></div>
      <div class="status-item">
        <span class="status-label">Status</span>
        <span class="status-value" :class="`status-${project.status}`">{{ statusLabel }}</span>
      </div>
      <div class="status-separator"></div>
      <div class="status-item">
        <span class="status-label">Target</span>
        <span class="status-value">{{ project.target || 'python' }}</span>
      </div>
    </div>
    <div v-else class="status-empty">
      <span>No project selected</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ProjectSummary } from '@/api/project'

const props = defineProps<{
  project: ProjectSummary | null
}>()

const statusLabel = computed(() => {
  if (!props.project) return ''
  const map: Record<string, string> = {
    active: 'Active',
    archived: 'Archived',
  }
  return map[props.project.status] || props.project.status
})
</script>

<style scoped>
.project-status {
  display: flex;
  align-items: center;
  height: 100%;
}

.status-content {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.status-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
}

.status-label {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.status-value {
  font-size: var(--text-caption);
  color: var(--text-secondary);
}

.status-value.status-active { color: var(--success); }
.status-value.status-archived { color: var(--neutral); }

.status-separator {
  width: 1px;
  height: 12px;
  background: var(--border-subtle);
}

.status-empty {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}
</style>
