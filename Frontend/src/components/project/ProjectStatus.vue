<template>
  <div class="project-status">
    <div v-if="project" class="status-content">
      <div class="status-item">
        <span class="status-label">项目</span>
        <span class="status-value">{{ project.name }}</span>
      </div>
      <div class="status-separator"></div>
      <div class="status-item">
        <span class="status-label">状态</span>
        <span class="status-value" :class="`status-${project.status}`">{{ statusLabel }}</span>
      </div>
      <div class="status-separator"></div>
      <div class="status-item">
        <span class="status-label">模型</span>
        <span class="status-value">{{ project.models.length }} 个</span>
      </div>
      <div class="status-separator"></div>
      <div class="status-item">
        <span class="status-label">数据集</span>
        <span class="status-value">{{ project.datasets.length }} 个</span>
      </div>
    </div>
    <div v-else class="status-empty">
      <span>未选择项目</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Project } from '@/types/project'

const props = defineProps<{
  project: Project | null
}>()

const statusLabel = computed(() => {
  if (!props.project) return ''
  const map: Record<string, string> = {
    active: '活跃',
    idle: '空闲',
    running: '运行中',
    error: '错误',
    archived: '已归档',
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
.status-value.status-idle { color: var(--neutral); }
.status-value.status-running { color: var(--info); }
.status-value.status-error { color: var(--error); }

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
