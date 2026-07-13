<template>
  <div class="plugin-install-task">
    <div class="install-task-header">
      <div class="install-task-title">
        <svg class="install-task-icon" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="var(--primary)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 12a9 9 0 1 1-6.219-8.56" />
        </svg>
        <span>Installing {{ task.pluginName }}</span>
      </div>
      <button
        v-if="task.status === 'running'"
        class="install-task-cancel"
        @click="$emit('cancel')"
      >
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M18 6 6 18M6 6l12 12" />
        </svg>
        <span>Cancel</span>
      </button>
    </div>

    <div class="install-task-progress">
      <div class="install-task-progress-bar">
        <div
          class="install-task-progress-fill"
          :style="{ width: progressPercent + '%' }"
        ></div>
      </div>
      <span class="install-task-progress-text">{{ progressPercent }}%</span>
    </div>

    <div class="install-task-steps">
      <div
        v-for="step in task.steps"
        :key="step.id"
        class="install-step"
        :class="`step-status--${step.status}`"
      >
        <div class="install-step-header" @click="toggleStep(step.id)">
          <span class="install-step-icon">
            <svg v-if="step.status === 'completed'" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="var(--success)" stroke-width="2"><path d="M20 6L9 17l-5-5" /></svg>
            <svg v-else-if="step.status === 'in-progress'" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="var(--primary)" stroke-width="2" class="is-pulse"><path d="M5 3l14 9-14 9V3z" /></svg>
            <svg v-else-if="step.status === 'failed'" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="var(--error)" stroke-width="2"><path d="M18 6 6 18M6 6l12 12" /></svg>
            <svg v-else viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="var(--text-tertiary)" stroke-width="2"><circle cx="12" cy="12" r="10" /></svg>
          </span>
          <span class="install-step-name">{{ step.name }}</span>
          <span v-if="step.duration" class="install-step-duration">{{ step.duration }}ms</span>
          <svg
            class="install-step-chevron"
            :class="{ 'is-open': expandedSteps.has(step.id) }"
            viewBox="0 0 24 24"
            width="12"
            height="12"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
          >
            <path d="M9 18l6-6-6-6" />
          </svg>
        </div>
        <Transition name="expand">
          <div v-if="expandedSteps.has(step.id) && step.logs.length > 0" class="install-step-logs">
            <div
              v-for="(log, idx) in step.logs"
              :key="idx"
              class="log-entry"
              :class="`log-level--${log.level}`"
            >
              <span class="log-message">{{ log.message }}</span>
            </div>
          </div>
        </Transition>
      </div>
    </div>

    <div v-if="task.status === 'completed'" class="install-task-complete">
      <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="var(--success)" stroke-width="2"><path d="M20 6L9 17l-5-5" /></svg>
      <span>Installation completed successfully</span>
      <button class="install-task-done" @click="$emit('done')">Done</button>
    </div>

    <div v-if="task.status === 'cancelled'" class="install-task-cancelled">
      <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="var(--warning)" stroke-width="2"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z M12 9v4M12 17h.01" /></svg>
      <span>Installation cancelled</span>
      <button class="install-task-done" @click="$emit('done')">Close</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { InstallTask } from '@/pages/PluginStore/types'

interface Props {
  task: InstallTask
}

const props = defineProps<Props>()

defineEmits<{
  cancel: []
  done: []
}>()

const expandedSteps = ref<Set<string>>(new Set())

const progressPercent = computed(() => {
  const total = props.task.steps.length
  if (total === 0) return 0
  const completed = props.task.steps.filter(s => s.status === 'completed').length
  return Math.round((completed / total) * 100)
})

function toggleStep(stepId: string) {
  if (expandedSteps.value.has(stepId)) {
    expandedSteps.value.delete(stepId)
  } else {
    expandedSteps.value.add(stepId)
  }
}
</script>

<style scoped>
.plugin-install-task {
  padding: var(--spacing-6);
}

.install-task-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-4);
}

.install-task-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--text-h2);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.install-task-icon.is-spinning {
  animation: spin 1s linear infinite;
}

.install-task-cancel {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--text-secondary);
  font-size: var(--text-caption);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.install-task-cancel:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.install-task-progress {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  margin-bottom: var(--spacing-4);
}

.install-task-progress-bar {
  flex: 1;
  height: 4px;
  background: var(--bg-hover);
  border-radius: 2px;
  overflow: hidden;
}

.install-task-progress-fill {
  height: 100%;
  background: var(--primary);
  border-radius: 2px;
  transition: width 300ms ease;
}

.install-task-progress-text {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  min-width: 36px;
  text-align: right;
  font-family: var(--font-family-mono);
}

.install-task-steps {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.install-step {
  background: var(--bg-tertiary);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.install-step-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-3) var(--spacing-4);
  cursor: pointer;
  transition: background var(--transition-fast);
}

.install-step-header:hover {
  background: var(--bg-hover);
}

.install-step-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  flex-shrink: 0;
}

.install-step-icon .is-pulse {
  animation: pulse 1.5s ease-in-out infinite;
}

.install-step-name {
  flex: 1;
  font-size: var(--text-body-sm);
  color: var(--text-primary);
}

.step-status--pending .install-step-name {
  color: var(--text-tertiary);
}

.install-step-duration {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  font-family: var(--font-family-mono);
}

.install-step-chevron {
  color: var(--text-tertiary);
  transition: transform var(--transition-fast);
  flex-shrink: 0;
}

.install-step-chevron.is-open {
  transform: rotate(180deg);
}

.install-step-logs {
  padding: 0 var(--spacing-4) var(--spacing-3);
  padding-left: 52px;
}

.log-entry {
  font-family: var(--font-family-mono);
  font-size: var(--text-caption);
  line-height: var(--leading-code);
  color: var(--text-secondary);
  padding: 2px 0;
}

.log-level--error {
  color: var(--error);
}

.log-level--warn {
  color: var(--warning);
}

.install-task-complete,
.install-task-cancelled {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  margin-top: var(--spacing-4);
  padding: var(--spacing-3) var(--spacing-4);
  border-radius: var(--radius-lg);
  font-size: var(--text-body-sm);
}

.install-task-complete {
  background: var(--success-bg);
  color: var(--success);
}

.install-task-cancelled {
  background: var(--warning-bg);
  color: var(--warning);
}

.install-task-done {
  margin-left: auto;
  padding: 4px 12px;
  border: 1px solid currentColor;
  border-radius: var(--radius-md);
  background: transparent;
  color: inherit;
  font-size: var(--text-caption);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.install-task-done:hover {
  background: rgba(255, 255, 255, 0.1);
}

.expand-enter-active,
.expand-leave-active {
  transition: all var(--transition-normal);
  overflow: hidden;
}

.expand-enter-from,
.expand-leave-to {
  opacity: 0;
  max-height: 0;
}

.expand-enter-to,
.expand-leave-from {
  opacity: 1;
  max-height: 200px;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}
</style>