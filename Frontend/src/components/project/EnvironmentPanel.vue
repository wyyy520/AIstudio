<template>
  <div class="environment-panel">
    <div class="panel-header">
      <h2 class="section-title">Environment</h2>
      <div class="panel-status">
        <span class="status-dot" :class="`status-${environment.gpuStatus}`"></span>
        <span class="status-text">{{ gpuStatusLabel }}</span>
      </div>
    </div>

    <div class="env-grid">
      <div class="env-card">
        <div class="env-icon">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M12 2L2 7l10 5 10-5-10-5zm0 22L2 17l10-5 10 5-10 5z" />
          </svg>
        </div>
        <div class="env-info">
          <span class="env-label">Python</span>
          <span class="env-value">{{ environment.pythonVersion }}</span>
        </div>
      </div>

      <div class="env-card">
        <div class="env-icon">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M22 12h-4l-3 9L9 3l-3 9H2" />
          </svg>
        </div>
        <div class="env-info">
          <span class="env-label">CUDA</span>
          <span class="env-value">{{ environment.cudaVersion }}</span>
        </div>
      </div>

      <div class="env-card">
        <div class="env-icon">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M12 2L2 7l10 5 10-5-10-5zm0 22L2 17l10-5 10 5-10 5z" />
          </svg>
        </div>
        <div class="env-info">
          <span class="env-label">PyTorch</span>
          <span class="env-value">{{ environment.pytorchVersion }}</span>
        </div>
      </div>

      <div class="env-card">
        <div class="env-icon" :class="`icon-${environment.gpuStatus}`">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.5">
            <rect x="4" y="4" width="16" height="16" rx="2" />
            <path d="M9 9h6v6H9z" />
          </svg>
        </div>
        <div class="env-info">
          <span class="env-label">GPU 状态</span>
          <span class="env-value" :class="`gpu-${environment.gpuStatus}`">{{ gpuStatusLabel }}</span>
        </div>
      </div>
    </div>

    <div class="dependencies-section">
      <div class="section-header">
        <span class="section-subtitle">依赖状态</span>
        <span class="dep-count">{{ environment.dependencies.length }} 项</span>
      </div>

      <div class="dep-list">
        <div
          v-for="dep in environment.dependencies"
          :key="dep.name"
          class="dep-item"
        >
          <div class="dep-info">
            <span class="dep-name">{{ dep.name }}</span>
            <span class="dep-version">v{{ dep.version }}</span>
          </div>
          <span class="dep-status" :class="`dep-${dep.status}`">{{ depStatusLabel(dep.status) }}</span>
        </div>
      </div>
    </div>

    <div class="panel-actions">
      <button class="action-btn" @click="$emit('repair')">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z" />
        </svg>
        修复环境
      </button>
      <button class="action-btn" @click="$emit('rebuild')">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
        </svg>
        重建环境
      </button>
      <button class="action-btn" @click="$emit('install')">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3" />
        </svg>
        安装依赖
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ProjectEnvironment } from '@/types/project'

const props = defineProps<{
  environment: ProjectEnvironment
}>()

defineEmits<{
  repair: []
  rebuild: []
  install: []
}>()

const gpuStatusLabel = computed(() => {
  const map: Record<string, string> = {
    ready: '就绪',
    warning: '警告',
    error: '错误',
  }
  return map[props.environment.gpuStatus] || props.environment.gpuStatus
})

function depStatusLabel(status: string): string {
  const map: Record<string, string> = {
    installed: '已安装',
    outdated: '需更新',
    missing: '缺失',
  }
  return map[status] || status
}
</script>

<style scoped>
.environment-panel {
  padding: var(--spacing-6);
  height: 100%;
  overflow-y: auto;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-4);
}

.section-title {
  font-size: var(--text-h3);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.panel-status {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-dot.status-ready { background: var(--success); }
.status-dot.status-warning { background: var(--warning); }
.status-dot.status-error { background: var(--error); }

.status-text {
  font-size: var(--text-body-sm);
  color: var(--text-secondary);
}

.env-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--spacing-3);
  margin-bottom: var(--spacing-6);
}

.env-card {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  padding: var(--spacing-3) var(--spacing-4);
  background: var(--bg-tertiary);
  border-radius: var(--radius-md);
}

.env-icon {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--primary-bg);
  color: var(--primary);
  border-radius: var(--radius-md);
}

.env-icon.icon-ready { background: var(--success-bg); color: var(--success); }
.env-icon.icon-warning { background: var(--warning-bg); color: var(--warning); }
.env-icon.icon-error { background: var(--error-bg); color: var(--error); }

.env-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.env-label {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.env-value {
  font-size: var(--text-body);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.env-value.gpu-ready { color: var(--success); }
.env-value.gpu-warning { color: var(--warning); }
.env-value.gpu-error { color: var(--error); }

.dependencies-section {
  margin-bottom: var(--spacing-6);
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-3);
}

.section-subtitle {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-secondary);
}

.dep-count {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.dep-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.dep-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-3);
  background: var(--bg-tertiary);
  border-radius: var(--radius-md);
}

.dep-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.dep-name {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.dep-version {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.dep-status {
  font-size: var(--text-caption);
  padding: 2px 8px;
  border-radius: var(--radius-sm);
}

.dep-installed {
  background: var(--success-bg);
  color: var(--success);
}

.dep-outdated {
  background: var(--warning-bg);
  color: var(--warning);
}

.dep-missing {
  background: var(--error-bg);
  color: var(--error);
}

.panel-actions {
  display: flex;
  gap: var(--spacing-3);
}

.action-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-2);
  height: 36px;
  border-radius: var(--radius-md);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  color: var(--text-secondary);
  font-size: var(--text-body-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.action-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
  border-color: var(--border-strong);
}
</style>
