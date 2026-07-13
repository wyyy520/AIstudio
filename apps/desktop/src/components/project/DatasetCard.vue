<template>
  <div class="dataset-card">
    <div class="card-header">
      <div class="card-icon">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M4 7V4h16v3M9 20h6M12 4v16" />
        </svg>
      </div>
      <div class="card-info">
        <div class="card-name">{{ dataset.name }}</div>
        <div class="card-format">{{ dataset.format }}</div>
      </div>
    </div>

    <div class="card-stats">
      <div class="stat">
        <span class="stat-value">{{ dataset.size }}</span>
        <span class="stat-label">大小</span>
      </div>
      <div class="stat">
        <span class="stat-value">{{ formatNumber(dataset.imageCount) }}</span>
        <span class="stat-label">图片</span>
      </div>
      <div class="stat">
        <span class="stat-value">{{ dataset.classCount }}</span>
        <span class="stat-label">类别</span>
      </div>
    </div>

    <div class="card-actions">
      <button class="action-btn" @click="$emit('preview', dataset)">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
          <circle cx="12" cy="12" r="3" />
        </svg>
        预览
      </button>
      <button class="action-btn" @click="$emit('convert', dataset)">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3" />
        </svg>
        转换
      </button>
      <button class="action-btn action-btn-danger" @click="$emit('delete', dataset)">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
        </svg>
        删除
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ProjectDataset } from '@/types/project'

defineProps<{
  dataset: ProjectDataset
}>()

defineEmits<{
  preview: [dataset: ProjectDataset]
  convert: [dataset: ProjectDataset]
  delete: [dataset: ProjectDataset]
}>()

function formatNumber(num: number): string {
  if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M'
  if (num >= 1000) return (num / 1000).toFixed(1) + 'K'
  return num.toString()
}
</script>

<style scoped>
.dataset-card {
  background: var(--bg-tertiary);
  border-radius: var(--radius-lg);
  padding: var(--spacing-4);
  transition: all var(--transition-normal);
  border: 1px solid transparent;
}

.dataset-card:hover {
  background: var(--bg-hover);
  border-color: var(--border-default);
}

.card-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  margin-bottom: var(--spacing-3);
}

.card-icon {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--primary-bg);
  color: var(--primary);
  border-radius: var(--radius-md);
}

.card-info {
  flex: 1;
  min-width: 0;
}

.card-name {
  font-size: var(--text-body);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-format {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.card-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-3);
  padding: var(--spacing-3);
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
}

.stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.stat-value {
  font-size: var(--text-body);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.stat-label {
  font-size: 10px;
  color: var(--text-tertiary);
}

.card-actions {
  display: flex;
  gap: var(--spacing-2);
}

.action-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-1);
  height: 28px;
  border-radius: var(--radius-sm);
  background: transparent;
  border: 1px solid var(--border-default);
  color: var(--text-secondary);
  font-size: var(--text-caption);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.action-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
  border-color: var(--border-strong);
}

.action-btn-danger:hover {
  background: var(--error-bg);
  color: var(--error);
  border-color: var(--error);
}
</style>
