<template>
  <div class="project-dashboard">
    <div class="dashboard-header">
      <div class="header-info">
        <h1 class="project-title">{{ project.name }}</h1>
        <div class="project-meta">
          <span class="meta-type">{{ typeLabel }}</span>
          <span class="meta-status" :class="`status-${project.status}`">
            <span class="status-dot"></span>
            {{ statusLabel }}
          </span>
        </div>
      </div>
    </div>

    <div class="dashboard-grid">
      <div class="grid-section">
        <div class="section-title">最近任务</div>
        <div class="recent-tasks">
          <div
            v-for="task in recentTasks"
            :key="task.id"
            class="task-item"
          >
            <span class="task-status" :class="`status-${task.status}`"></span>
            <span class="task-name">{{ task.name }}</span>
            <span class="task-time">{{ task.time }}</span>
          </div>
          <div v-if="recentTasks.length === 0" class="empty-hint">暂无任务</div>
        </div>
      </div>

      <div class="grid-section">
        <div class="section-title">当前 Workflow</div>
        <div class="workflow-list">
          <div
            v-for="workflow in project.workflows.slice(0, 3)"
            :key="workflow.id"
            class="workflow-item"
          >
            <div class="workflow-info">
              <span class="workflow-name">{{ workflow.name }}</span>
              <span class="workflow-version">v{{ workflow.version }}</span>
            </div>
            <div class="workflow-meta">
              <span class="workflow-nodes">{{ workflow.nodeCount }} 节点</span>
              <span class="workflow-status" :class="`status-${workflow.status}`">{{ workflowStatusLabel(workflow.status) }}</span>
            </div>
          </div>
          <div v-if="project.workflows.length === 0" class="empty-hint">暂无 Workflow</div>
        </div>
      </div>

      <div class="grid-section">
        <div class="section-title">当前模型</div>
        <div class="model-list">
          <div
            v-for="model in project.models.slice(0, 3)"
            :key="model.id"
            class="model-item"
          >
            <div class="model-info">
              <span class="model-name">{{ model.name }}</span>
              <span class="model-framework">{{ model.framework }}</span>
            </div>
            <div class="model-accuracy">{{ (model.accuracy * 100).toFixed(1) }}%</div>
          </div>
          <div v-if="project.models.length === 0" class="empty-hint">暂无模型</div>
        </div>
      </div>

      <div class="grid-section">
        <div class="section-title">环境状态</div>
        <div class="env-status">
          <div class="env-item">
            <span class="env-label">Python</span>
            <span class="env-value">{{ project.environment.pythonVersion }}</span>
          </div>
          <div class="env-item">
            <span class="env-label">CUDA</span>
            <span class="env-value">{{ project.environment.cudaVersion }}</span>
          </div>
          <div class="env-item">
            <span class="env-label">PyTorch</span>
            <span class="env-value">{{ project.environment.pytorchVersion }}</span>
          </div>
          <div class="env-item">
            <span class="env-label">GPU</span>
            <span class="env-value" :class="`gpu-${project.environment.gpuStatus}`">{{ gpuStatusLabel }}</span>
          </div>
        </div>
      </div>
    </div>

    <div class="dashboard-actions">
      <button class="action-btn action-btn-primary" @click="$emit('run-workflow')">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
          <polygon points="5,3 19,12 5,21" />
        </svg>
        运行 Workflow
      </button>
      <button class="action-btn" @click="$emit('train-model')">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M12 2L2 7l10 5 10-5-10-5zm0 22L2 17l10-5 10 5-10 5z" />
        </svg>
        训练模型
      </button>
      <button class="action-btn" @click="$emit('open-dataset')">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M4 7V4h16v3M9 20h6M12 4v16" />
        </svg>
        打开数据集
      </button>
      <button class="action-btn" @click="$emit('deploy-model')">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M22 12h-4l-3 9L9 3l-3 9H2" />
        </svg>
        部署模型
      </button>
      <button class="action-btn" @click="$emit('view-experiment')">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M9 3h6v11l-3 3-3-3V3z" />
        </svg>
        查看实验
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
  'run-workflow': []
  'train-model': []
  'open-dataset': []
  'deploy-model': []
  'view-experiment': []
}>()

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

const gpuStatusLabel = computed(() => {
  const map: Record<string, string> = {
    ready: '就绪',
    warning: '警告',
    error: '错误',
  }
  return map[props.project.environment.gpuStatus] || props.project.environment.gpuStatus
})

interface RecentTask {
  id: string
  name: string
  status: string
  time: string
}

const recentTasks = computed<RecentTask[]>(() => {
  const tasks: RecentTask[] = []
  props.project.workflows.forEach(w => {
    if (w.status === 'completed' || w.status === 'running') {
      tasks.push({
        id: w.id,
        name: w.name,
        status: w.status,
        time: new Date(w.updatedAt).toLocaleDateString('zh-CN'),
      })
    }
  })
  props.project.experiments.forEach(e => {
    if (e.status === 'completed' || e.status === 'running') {
      tasks.push({
        id: e.id,
        name: `${e.modelName} - Epoch ${e.epoch}`,
        status: e.status,
        time: e.duration,
      })
    }
  })
  return tasks.slice(0, 5)
})

function workflowStatusLabel(status: string): string {
  const map: Record<string, string> = {
    pending: '等待中',
    running: '运行中',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消',
  }
  return map[status] || status
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

.meta-type {
  font-size: var(--text-body-sm);
  color: var(--text-secondary);
  background: var(--bg-tertiary);
  padding: var(--spacing-1) var(--spacing-2);
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
.status-idle { color: var(--neutral); }
.status-idle .status-dot { background: var(--neutral); }
.status-running { color: var(--info); }
.status-running .status-dot { background: var(--info); animation: pulse 2s infinite; }
.status-error { color: var(--error); }
.status-error .status-dot { background: var(--error); }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--spacing-4);
  margin-bottom: var(--spacing-6);
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

.recent-tasks,
.workflow-list,
.model-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.task-item,
.workflow-item,
.model-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-2) var(--spacing-3);
  background: var(--bg-secondary);
  border-radius: var(--radius-sm);
}

.task-status {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: var(--spacing-2);
}

.task-status.status-completed { background: var(--success); }
.task-status.status-running { background: var(--info); }

.task-name,
.workflow-name,
.model-name {
  flex: 1;
  font-size: var(--text-body-sm);
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.task-time,
.workflow-version {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  margin-left: var(--spacing-2);
}

.workflow-info,
.model-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.workflow-version,
.model-framework {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.workflow-meta {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.workflow-nodes {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.workflow-status {
  font-size: var(--text-caption);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
}

.workflow-status.status-completed {
  background: var(--success-bg);
  color: var(--success);
}

.workflow-status.status-running {
  background: var(--info-bg);
  color: var(--info);
}

.workflow-status.status-idle {
  background: var(--bg-hover);
  color: var(--text-tertiary);
}

.model-accuracy {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--success);
}

.env-status {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--spacing-2);
}

.env-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: var(--spacing-2) var(--spacing-3);
  background: var(--bg-secondary);
  border-radius: var(--radius-sm);
}

.env-label {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.env-value {
  font-size: var(--text-body-sm);
  color: var(--text-primary);
}

.env-value.gpu-ready { color: var(--success); }
.env-value.gpu-warning { color: var(--warning); }
.env-value.gpu-error { color: var(--error); }

.empty-hint {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  text-align: center;
  padding: var(--spacing-4);
}

.dashboard-actions {
  display: flex;
  gap: var(--spacing-3);
  flex-wrap: wrap;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  height: 36px;
  padding: 0 var(--spacing-4);
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

.action-btn-primary {
  background: var(--primary);
  color: white;
  border-color: var(--primary);
}

.action-btn-primary:hover {
  background: var(--primary-hover);
}
</style>
