<template>
  <div class="project-dashboard">
    <div class="dashboard-header">
      <div class="header-info">
        <h1 class="project-title">{{ project.name }}</h1>
        <div class="project-meta">
          <span class="meta-target">{{ project.target || 'python' }}</span>
          <span class="meta-status" :class="`status-${project.status}`">
            <span class="status-dot"></span>
            {{ statusLabel }}
          </span>
        </div>
      </div>
    </div>

    <div class="dashboard-grid">
      <div class="grid-section">
        <div class="section-title">Project Info</div>
        <div class="info-list">
          <div class="info-item">
            <span class="info-label">Path</span>
            <span class="info-value" :title="project.rootPath">{{ project.rootPath || '—' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Created</span>
            <span class="info-value">{{ formatDate(project.createdAt) }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Updated</span>
            <span class="info-value">{{ formatDate(project.updatedAt) }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Files</span>
            <span class="info-value">{{ project.fileCount }}</span>
          </div>
        </div>
      </div>

      <div class="grid-section">
        <div class="section-title">Quick Actions</div>
        <div class="action-list">
          <button class="action-item" @click="$emit('open-workflow')">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.5">
              <polygon points="5,3 19,12 5,21" />
            </svg>
            <span>Open Workflow</span>
          </button>
          <button class="action-item disabled" disabled title="Coming in Phase 2">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M12 2L2 7l10 5 10-5-10-5zm0 22L2 17l10-5 10 5-10 5z" />
            </svg>
            <span>Run Workflow</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ProjectSummary } from '@/api/project'

interface Props {
  project: ProjectSummary
}

defineProps<Props>()

defineEmits<{
  'open-workflow': []
  'run-workflow': []
}>()

const statusLabel = computed(() => {
  const map: Record<string, string> = {
    active: 'Active',
    archived: 'Archived',
  }
  return (s: string) => map[s] || s
})

function formatDate(dateStr: string): string {
  if (!dateStr) return '—'
  try {
    return new Date(dateStr).toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return dateStr
  }
}
</script>

<style scoped>
.project-dashboard {
  padding: var(--spacing-6);
  overflow-y: auto;
  height: 100%;
}

.dashboard-header {
  margin-bottom: var(--spacing-6);
}

.project-title {
  font-size: var(--text-h2);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  margin-bottom: var(--spacing-2);
}

.project-meta {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.meta-target {
  font-size: var(--text-body-sm);
  color: var(--text-secondary);
  background: var(--bg-tertiary);
  padding: 2px 8px;
  border-radius: var(--radius-sm);
}

.meta-status {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  font-size: var(--text-body-sm);
}

.meta-status .status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-active { color: var(--success); }
.status-active .status-dot { background: var(--success); }
.status-archived { color: var(--neutral); }
.status-archived .status-dot { background: var(--neutral); }

.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--spacing-4);
}

.grid-section {
  background: var(--bg-tertiary);
  border-radius: var(--radius-lg);
  padding: var(--spacing-4);
}

.section-title {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-secondary);
  margin-bottom: var(--spacing-3);
}

.info-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-2) var(--spacing-3);
  background: var(--bg-secondary);
  border-radius: var(--radius-sm);
}

.info-label {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.info-value {
  font-size: var(--text-body-sm);
  color: var(--text-primary);
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.action-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.action-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-3);
  background: var(--bg-secondary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-size: var(--text-body-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.action-item:hover:not(.disabled) {
  background: var(--bg-hover);
  border-color: var(--primary);
}

.action-item.disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
</style>
