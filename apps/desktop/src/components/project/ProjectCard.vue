<template>
  <div class="project-card" @click="$emit('select', project)">
    <div class="card-header">
      <div class="card-title">{{ project.name }}</div>
      <div class="card-status">
        <span class="status-dot" :class="statusClass"></span>
        <span class="status-text">{{ statusLabel }}</span>
      </div>
    </div>

    <div class="card-body">
      <div class="card-meta">
        <div class="meta-item">
          <span class="meta-label">类型</span>
          <span class="meta-value">{{ typeLabel }}</span>
        </div>
        <div class="meta-item">
          <span class="meta-label">更新</span>
          <span class="meta-value">{{ formatDate(project.updatedAt) }}</span>
        </div>
      </div>

      <div class="card-stats">
        <div class="stat-item">
          <span class="stat-value">{{ project.workflows.length }}</span>
          <span class="stat-label">Workflows</span>
        </div>
        <div class="stat-item">
          <span class="stat-value">{{ project.datasets.length }}</span>
          <span class="stat-label">Datasets</span>
        </div>
        <div class="stat-item">
          <span class="stat-value">{{ project.models.length }}</span>
          <span class="stat-label">Models</span>
        </div>
        <div class="stat-item">
          <span class="stat-value">{{ project.experiments.length }}</span>
          <span class="stat-label">Experiments</span>
        </div>
      </div>
    </div>

    <div class="card-footer">
      <button class="card-btn" @click.stop="$emit('open', project)">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6M15 3h6v6M10 14L21 3" />
        </svg>
        打开
      </button>
      <button class="card-btn card-btn-danger" @click.stop="$emit('delete', project)">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
        </svg>
        删除
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Project } from '@/types/project'

interface Props {
  project: Project
}

const props = defineProps<Props>()

defineEmits<{
  select: [project: Project]
  open: [project: Project]
  delete: [project: Project]
}>()

const statusClass = computed(() => {
  const map: Record<string, string> = {
    active: 'status-active',
    idle: 'status-idle',
    running: 'status-running',
    error: 'status-error',
    archived: 'status-archived',
  }
  return map[props.project.status] || 'status-idle'
})

const statusLabel = computed(() => {
  const map: Record<string, string> = {
    active: '活跃',
    idle: '空闲',
    running: '运行中',
    error: '错误',
    archived: '已归档',
  }
  return map[props.project.status] || props.project.status
})

const typeLabel = computed(() => {
  const map: Record<string, string> = {
    detection: '目标检测',
    classification: '图像分类',
    segmentation: '语义分割',
    timeseries: '时序预测',
    custom: '自定义',
  }
  return map[props.project.type] || props.project.type
})

function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))

  if (days === 0) return '今天'
  if (days === 1) return '昨天'
  if (days < 7) return `${days} 天前`
  if (days < 30) return `${Math.floor(days / 7)} 周前`
  return date.toLocaleDateString('zh-CN')
}
</script>

<style scoped>
.project-card {
  background: var(--bg-tertiary);
  border-radius: var(--radius-lg);
  padding: var(--spacing-4);
  cursor: pointer;
  transition: all var(--transition-normal);
  border: 1px solid transparent;
}

.project-card:hover {
  background: var(--bg-hover);
  border-color: var(--border-default);
  box-shadow: var(--shadow);
}

.card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: var(--spacing-3);
}

.card-title {
  font-size: var(--text-body);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  margin-right: var(--spacing-2);
}

.card-status {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  flex-shrink: 0;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-active {
  background: var(--success);
}

.status-idle {
  background: var(--neutral);
}

.status-running {
  background: var(--info);
  animation: pulse 2s infinite;
}

.status-error {
  background: var(--error);
}

.status-archived {
  background: var(--text-tertiary);
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.status-text {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.card-body {
  margin-bottom: var(--spacing-3);
}

.card-meta {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-3);
}

.meta-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.meta-label {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.meta-value {
  font-size: var(--text-caption);
  color: var(--text-secondary);
}

.card-stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--spacing-2);
}

.stat-item {
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

.card-footer {
  display: flex;
  gap: var(--spacing-2);
  padding-top: var(--spacing-3);
  border-top: 1px solid var(--border-subtle);
}

.card-btn {
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

.card-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
  border-color: var(--border-strong);
}

.card-btn-danger:hover {
  background: var(--error-bg);
  color: var(--error);
  border-color: var(--error);
}
</style>
