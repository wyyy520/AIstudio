<template>
  <div class="task-card">
    <div class="task-card-header">
      <div class="task-card-icon">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M9 11l3 3L22 4" />
          <path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11" />
        </svg>
      </div>
      <span class="task-card-title">{{ task.title }}</span>
    </div>

    <div class="task-card-steps">
      <div
        v-for="step in task.steps"
        :key="step.id"
        class="task-step"
        :class="`task-step--${step.status}`"
      >
        <span class="task-step-icon">
          <svg v-if="step.status === 'completed'" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="20 6 9 17 4 12" />
          </svg>
          <svg v-else-if="step.status === 'running'" class="task-step-spinner" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 12a9 9 0 1 1-6.219-8.56" />
          </svg>
          <svg v-else-if="step.status === 'error'" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10" />
            <line x1="15" y1="9" x2="9" y2="15" />
            <line x1="9" y1="9" x2="15" y2="15" />
          </svg>
          <span v-else class="task-step-pending" />
        </span>
        <span class="task-step-label">{{ step.label }}</span>
      </div>
    </div>

    <div class="task-card-progress">
      <div class="task-progress-bar">
        <div class="task-progress-fill" :style="{ width: `${task.progress}%` }" />
      </div>
      <span class="task-progress-text">{{ task.progress }}%</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { TaskExecution } from '../../types'

defineProps<{
  task: TaskExecution
}>()
</script>

<style scoped>
.task-card {
  background: var(--bg-primary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: var(--spacing-3);
  margin-top: var(--spacing-3);
}

.task-card-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-3);
}

.task-card-icon {
  color: var(--primary);
  display: flex;
  align-items: center;
}

.task-card-title {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.task-card-steps {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-3);
}

.task-step {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.task-step-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

.task-step--completed .task-step-icon {
  color: var(--success);
}

.task-step--running .task-step-icon {
  color: var(--primary);
}

.task-step--error .task-step-icon {
  color: var(--error);
}

.task-step--pending .task-step-icon {
  color: var(--text-disabled);
}

.task-step-pending {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  border: 1.5px solid var(--text-disabled);
}

.task-step-spinner {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.task-step-label {
  font-size: var(--text-body-sm);
  color: var(--text-secondary);
}

.task-step--completed .task-step-label {
  color: var(--text-tertiary);
}

.task-step--running .task-step-label {
  color: var(--text-primary);
  font-weight: var(--font-semibold);
}

.task-card-progress {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.task-progress-bar {
  flex: 1;
  height: 4px;
  background: var(--bg-hover);
  border-radius: 2px;
  overflow: hidden;
}

.task-progress-fill {
  height: 100%;
  background: var(--primary);
  border-radius: 2px;
  transition: width 500ms ease;
}

.task-progress-text {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  font-variant-numeric: tabular-nums;
  min-width: 32px;
  text-align: right;
}
</style>
