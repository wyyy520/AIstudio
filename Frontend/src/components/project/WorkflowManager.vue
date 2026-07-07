<template>
  <div class="workflow-manager">
    <div class="manager-header">
      <h2 class="section-title">Workflows</h2>
      <button class="create-btn" @click="$emit('create')">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M12 5v14M5 12h14" />
        </svg>
        新建
      </button>
    </div>

    <div v-if="workflows.length === 0" class="empty-state">
      <svg viewBox="0 0 24 24" width="48" height="48" fill="none" stroke="currentColor" stroke-width="1.5">
        <path d="M6 3h3v6H6V3zm0 12h3v6H6v-6zm9-12h3v6h-3V3zm0 12h3v6h-3v-6zm-9 0V9m3 9v-3m3 3v-3m3 3V9" />
      </svg>
      <span>暂无 Workflow</span>
      <span class="empty-hint">点击上方按钮创建新的 Workflow</span>
    </div>

    <div v-else class="workflow-list">
      <div
        v-for="workflow in workflows"
        :key="workflow.id"
        class="workflow-item"
      >
        <div class="workflow-main">
          <div class="workflow-icon">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M6 3h3v6H6V3zm0 12h3v6H6v-6zm9-12h3v6h-3V3zm0 12h3v6h-3v-6zm-9 0V9m3 9v-3m3 3v-3m3 3V9" />
            </svg>
          </div>
          <div class="workflow-info">
            <div class="workflow-name">{{ workflow.name }}</div>
            <div class="workflow-meta">
              <span class="meta-item">v{{ workflow.version }}</span>
              <span class="meta-separator">·</span>
              <span class="meta-item">{{ workflow.nodeCount }} 节点</span>
              <span class="meta-separator">·</span>
              <span class="meta-item">{{ formatDate(workflow.updatedAt) }}</span>
            </div>
          </div>
          <div class="workflow-status" :class="`status-${workflow.status}`">
            <span class="status-dot"></span>
            {{ statusLabel(workflow.status) }}
          </div>
        </div>

        <div class="workflow-actions">
          <button class="action-btn" @click="$emit('open', workflow)">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6M15 3h6v6M10 14L21 3" />
            </svg>
            打开
          </button>
          <button
            class="action-btn action-btn-run"
            :disabled="workflow.status === 'running'"
            @click="$emit('run', workflow)"
          >
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
              <polygon points="5,3 19,12 5,21" />
            </svg>
            {{ workflow.status === 'running' ? '运行中...' : '运行' }}
          </button>
          <button class="action-btn" @click="$emit('clone', workflow)">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
              <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
              <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
            </svg>
            克隆
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ProjectWorkflow } from '@/types/project'

defineProps<{
  workflows: ProjectWorkflow[]
}>()

defineEmits<{
  create: []
  open: [workflow: ProjectWorkflow]
  run: [workflow: ProjectWorkflow]
  clone: [workflow: ProjectWorkflow]
}>()

function statusLabel(status: string): string {
  const map: Record<string, string> = {
    pending: '等待中',
    running: '运行中',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消',
    idle: '空闲',
  }
  return map[status] || status
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('zh-CN')
}
</script>

<style scoped>
.workflow-manager {
  padding: var(--spacing-6);
  height: 100%;
  overflow-y: auto;
}

.manager-header {
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

.create-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  height: 32px;
  padding: 0 var(--spacing-3);
  border-radius: var(--radius-md);
  background: var(--primary);
  border: none;
  color: white;
  font-size: var(--text-body-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.create-btn:hover {
  background: var(--primary-hover);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-3);
  height: 300px;
  color: var(--text-tertiary);
}

.empty-state span {
  font-size: var(--text-body);
}

.empty-hint {
  font-size: var(--text-caption);
  color: var(--text-disabled);
}

.workflow-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.workflow-item {
  background: var(--bg-tertiary);
  border-radius: var(--radius-lg);
  padding: var(--spacing-4);
  transition: all var(--transition-normal);
  border: 1px solid transparent;
}

.workflow-item:hover {
  background: var(--bg-hover);
  border-color: var(--border-default);
}

.workflow-main {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  margin-bottom: var(--spacing-3);
}

.workflow-icon {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--primary-bg);
  color: var(--primary);
  border-radius: var(--radius-md);
}

.workflow-info {
  flex: 1;
  min-width: 0;
}

.workflow-name {
  font-size: var(--text-body);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  margin-bottom: var(--spacing-1);
}

.workflow-meta {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.meta-separator {
  color: var(--border-default);
}

.workflow-status {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  font-size: var(--text-caption);
  padding: var(--spacing-1) var(--spacing-2);
  border-radius: var(--radius-sm);
}

.workflow-status .status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.status-completed {
  background: var(--success-bg);
  color: var(--success);
}
.status-completed .status-dot { background: var(--success); }

.status-running {
  background: var(--info-bg);
  color: var(--info);
}
.status-running .status-dot { background: var(--info); animation: pulse 2s infinite; }

.status-pending {
  background: var(--warning-bg);
  color: var(--warning);
}
.status-pending .status-dot { background: var(--warning); }

.status-failed {
  background: var(--error-bg);
  color: var(--error);
}
.status-failed .status-dot { background: var(--error); }

.status-idle,
.status-cancelled {
  background: var(--bg-hover);
  color: var(--text-tertiary);
}
.status-idle .status-dot,
.status-cancelled .status-dot { background: var(--text-tertiary); }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.workflow-actions {
  display: flex;
  gap: var(--spacing-2);
  padding-top: var(--spacing-3);
  border-top: 1px solid var(--border-subtle);
}

.action-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  height: 28px;
  padding: 0 var(--spacing-3);
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

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.action-btn-run {
  background: var(--success-bg);
  color: var(--success);
  border-color: var(--success);
}

.action-btn-run:hover:not(:disabled) {
  background: var(--success);
  color: white;
}
</style>
