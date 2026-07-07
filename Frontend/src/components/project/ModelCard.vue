<template>
  <div class="model-card">
    <div class="card-header">
      <div class="card-icon" :class="`icon-${model.framework}`">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M12 2L2 7l10 5 10-5-10-5zm0 22L2 17l10-5 10 5-10 5z" />
        </svg>
      </div>
      <div class="card-info">
        <div class="card-name">{{ model.name }}</div>
        <div class="card-version">v{{ model.version }}</div>
      </div>
      <div class="card-framework">{{ frameworkLabel }}</div>
    </div>

    <div class="card-details">
      <div class="detail-row">
        <span class="detail-label">大小</span>
        <span class="detail-value">{{ model.size }}</span>
      </div>
      <div class="detail-row">
        <span class="detail-label">来源</span>
        <span class="detail-value">{{ model.source }}</span>
      </div>
      <div class="detail-row">
        <span class="detail-label">训练时间</span>
        <span class="detail-value">{{ formatDate(model.trainedAt) }}</span>
      </div>
      <div class="detail-row">
        <span class="detail-label">准确率</span>
        <span class="detail-value accuracy">{{ (model.accuracy * 100).toFixed(1) }}%</span>
      </div>
    </div>

    <div class="card-actions">
      <button class="action-btn" @click="$emit('load', model)">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M17 8l-5-5-5 5M12 3v12" />
        </svg>
        加载
      </button>
      <button class="action-btn" @click="$emit('export', model)">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3" />
        </svg>
        导出
      </button>
      <button class="action-btn action-btn-primary" @click="$emit('deploy', model)">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M22 12h-4l-3 9L9 3l-3 9H2" />
        </svg>
        部署
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ProjectModel } from '@/types/project'

const props = defineProps<{
  model: ProjectModel
}>()

defineEmits<{
  load: [model: ProjectModel]
  export: [model: ProjectModel]
  deploy: [model: ProjectModel]
}>()

const frameworkLabel = computed(() => {
  const map: Record<string, string> = {
    pytorch: 'PyTorch',
    tensorflow: 'TensorFlow',
    onnx: 'ONNX',
    tensorrt: 'TensorRT',
    auto: 'Auto',
  }
  return map[props.model.framework] || props.model.framework
})

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('zh-CN')
}
</script>

<style scoped>
.model-card {
  background: var(--bg-tertiary);
  border-radius: var(--radius-lg);
  padding: var(--spacing-4);
  transition: all var(--transition-normal);
  border: 1px solid transparent;
}

.model-card:hover {
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
  border-radius: var(--radius-md);
}

.icon-pytorch {
  background: rgba(238, 76, 44, 0.1);
  color: #ee4c2c;
}

.icon-tensorflow {
  background: rgba(255, 160, 0, 0.1);
  color: #ffa000;
}

.icon-onnx {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
}

.icon-tensorrt {
  background: rgba(118, 185, 0, 0.1);
  color: #76b900;
}

.icon-auto {
  background: var(--primary-bg);
  color: var(--primary);
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

.card-version {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.card-framework {
  font-size: var(--text-caption);
  color: var(--text-secondary);
  background: var(--bg-secondary);
  padding: var(--spacing-1) var(--spacing-2);
  border-radius: var(--radius-sm);
}

.card-details {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-3);
  padding: var(--spacing-3);
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
}

.detail-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.detail-label {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.detail-value {
  font-size: var(--text-caption);
  color: var(--text-secondary);
}

.detail-value.accuracy {
  color: var(--success);
  font-weight: var(--font-semibold);
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

.action-btn-primary {
  background: var(--primary-bg);
  color: var(--primary);
  border-color: var(--primary);
}

.action-btn-primary:hover {
  background: var(--primary);
  color: white;
}
</style>
