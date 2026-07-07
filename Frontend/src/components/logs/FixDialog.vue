<template>
  <Teleport to="body">
    <div v-if="visible" class="fix-dialog-overlay" @click.self="$emit('cancel')">
      <div class="fix-dialog">
        <div class="fix-dialog-header">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z" />
          </svg>
          <span class="fix-dialog-title">Apply Fix</span>
        </div>

        <div class="fix-dialog-body">
          <p class="fix-dialog-desc">AI suggests the following fix:</p>

          <div v-if="command" class="fix-command selectable">
            <span class="fix-command-text">{{ command }}</span>
          </div>

          <div class="fix-steps">
            <div v-for="step in steps" :key="step.id" class="fix-step" :class="`step-${step.status}`">
              <span class="fix-step-icon">
                <svg v-if="step.status === 'completed'" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M20 6 9 17l-5-5" />
                </svg>
                <svg v-else-if="step.status === 'running'" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="spin-icon">
                  <path d="M21 12a9 9 0 1 1-6.219-8.56" />
                </svg>
                <svg v-else-if="step.status === 'failed'" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="12" cy="12" r="10" /><path d="m15 9-6 6" /><path d="m9 9 6 6" />
                </svg>
                <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="12" cy="12" r="10" />
                </svg>
              </span>
              <span class="fix-step-label">{{ step.label }}</span>
            </div>
          </div>

          <div v-if="isFixing" class="fix-warning">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z" /><path d="M12 9v4" /><path d="M12 17h.01" />
            </svg>
            <span>This action will modify your environment configuration.</span>
          </div>
        </div>

        <div class="fix-dialog-footer">
          <AppButton type="ghost" size="medium" label="Cancel" :disabled="isFixing" @click="$emit('cancel')" />
          <AppButton
            type="primary"
            size="medium"
            :label="isFixing ? 'Fixing...' : 'Apply Fix'"
            :loading="isFixing"
            :disabled="isFixing"
            @click="$emit('execute')"
          />
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import type { FixStep } from '@/pages/Logs/types'
import AppButton from '@/components/AppButton/AppButton.vue'

interface Props {
  visible: boolean
  steps: FixStep[]
  command?: string
  isFixing: boolean
}

defineProps<Props>()

defineEmits<{
  cancel: []
  execute: []
}>()
</script>

<style scoped>
.fix-dialog-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 200ms ease-in-out;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.fix-dialog {
  width: 480px;
  max-width: 90vw;
  background: var(--bg-secondary);
  border-radius: var(--radius-2xl);
  box-shadow: var(--shadow-xl);
  border: 1px solid var(--border-default);
  animation: scaleIn 250ms cubic-bezier(0.4, 0, 0.2, 1);
}

@keyframes scaleIn {
  from { opacity: 0; transform: scale(0.95); }
  to { opacity: 1; transform: scale(1); }
}

.fix-dialog-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-4) var(--spacing-6);
  border-bottom: 1px solid var(--border-subtle);
}

.fix-dialog-header svg {
  color: var(--primary);
}

.fix-dialog-title {
  font-size: var(--text-h3);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.fix-dialog-body {
  padding: var(--spacing-4) var(--spacing-6);
}

.fix-dialog-desc {
  font-size: var(--text-body-sm);
  color: var(--text-secondary);
  margin-bottom: var(--spacing-3);
}

.fix-command {
  padding: var(--spacing-3);
  background: var(--bg-tertiary);
  border-radius: var(--radius-lg);
  margin-bottom: var(--spacing-4);
  font-family: var(--font-family-mono);
  font-size: var(--text-code);
  color: var(--text-primary);
  border: 1px solid var(--border-subtle);
}

.fix-command-text {
  white-space: pre-wrap;
  word-break: break-all;
}

.fix-steps {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-3);
}

.fix-step {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.fix-step-icon {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.step-completed .fix-step-icon { color: var(--success); }
.step-running .fix-step-icon { color: var(--info); }
.step-failed .fix-step-icon { color: var(--error); }
.step-pending .fix-step-icon { color: var(--text-disabled); }

.spin-icon {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.fix-step-label {
  font-size: var(--text-body-sm);
  color: var(--text-primary);
}

.step-completed .fix-step-label {
  color: var(--success);
}

.step-running .fix-step-label {
  color: var(--info);
}

.fix-warning {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-2) var(--spacing-3);
  background: var(--warning-bg);
  border-radius: var(--radius-md);
  font-size: var(--text-caption);
  color: var(--warning);
}

.fix-dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--spacing-2);
  padding: var(--spacing-3) var(--spacing-6);
  border-top: 1px solid var(--border-subtle);
}
</style>