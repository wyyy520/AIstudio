<template>
  <div class="log-viewer">
    <div class="log-viewer-tabs">
      <button
        class="tab-btn"
        :class="{ active: activeTab === 'human' }"
        @click="$emit('update:activeTab', 'human')"
      >
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" /><circle cx="12" cy="7" r="4" />
        </svg>
        Human Log
      </button>
      <button
        class="tab-btn"
        :class="{ active: activeTab === 'raw' }"
        @click="$emit('update:activeTab', 'raw')"
      >
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="m16 18 2-2-2-2" /><path d="m8 18-2-2 2-2" /><path d="m14 4-4 16" />
        </svg>
        Raw Log
      </button>
    </div>

    <div class="log-viewer-content">
      <div v-if="activeTab === 'human'" class="human-log">
        <div
          v-for="entry in logs"
          :key="entry.id"
          class="human-log-step"
          :class="`step-${entry.stepStatus}`"
        >
          <div class="step-row">
            <span class="step-time">{{ formatTime(entry.timestamp) }}</span>
            <span class="step-icon">
              <svg v-if="entry.stepStatus === 'completed'" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M20 6 9 17l-5-5" />
              </svg>
              <svg v-else-if="entry.stepStatus === 'running'" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="10" /><path d="M12 8v4l3 3" />
              </svg>
              <svg v-else-if="entry.stepStatus === 'failed'" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="10" /><path d="m15 9-6 6" /><path d="m9 9 6 6" />
              </svg>
              <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="10" />
              </svg>
            </span>
            <span class="step-message">{{ entry.humanMessage }}</span>
          </div>
        </div>

        <div v-if="!logs.length" class="log-empty">
          <svg viewBox="0 0 24 24" width="40" height="40" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" /><path d="M14 2v6h6" /><path d="M16 13H8" /><path d="M16 17H8" /><path d="M10 9H8" />
          </svg>
          <span class="log-empty-title">No Logs</span>
          <span class="log-empty-desc">Select a task to view logs</span>
        </div>
      </div>

      <RawLogViewer
        v-else
        :logs="logs"
        @download="$emit('download')"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { LogEntry, LogTab } from '@/pages/Logs/types'
import RawLogViewer from './RawLogViewer.vue'

interface Props {
  logs: LogEntry[]
  activeTab: LogTab
}

defineProps<Props>()

defineEmits<{
  'update:activeTab': [tab: LogTab]
  download: []
}>()

function formatTime(iso: string): string {
  const d = new Date(iso)
  return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}
</script>

<style scoped>
.log-viewer {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.log-viewer-tabs {
  display: flex;
  gap: 0;
  background: var(--bg-tertiary);
  border-bottom: 1px solid var(--border-subtle);
}

.tab-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  height: 36px;
  padding: 0 var(--spacing-4);
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--text-secondary);
  font-size: var(--text-body-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.tab-btn:hover {
  color: var(--text-primary);
  background: var(--bg-hover);
}

.tab-btn.active {
  color: var(--text-primary);
  border-bottom-color: var(--primary);
}

.log-viewer-content {
  flex: 1;
  overflow: auto;
}

.human-log {
  padding: var(--spacing-3);
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.human-log-step {
  padding: var(--spacing-2) var(--spacing-3);
  border-radius: var(--radius-md);
  border-left: 2px solid transparent;
  transition: background var(--transition-fast);
}

.human-log-step:hover {
  background: var(--bg-hover);
}

.human-log-step.step-running {
  border-left-color: var(--info);
}

.human-log-step.step-failed {
  border-left-color: var(--error);
  background: var(--error-bg);
}

.step-row {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.step-time {
  font-size: var(--text-caption);
  font-family: var(--font-family-mono);
  color: var(--text-tertiary);
  flex-shrink: 0;
}

.step-icon {
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

.step-completed .step-icon { color: var(--success); }
.step-running .step-icon { color: var(--info); animation: pulse 2s ease-in-out infinite; }
.step-failed .step-icon { color: var(--error); }
.step-pending .step-icon { color: var(--text-disabled); }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.step-message {
  font-size: var(--text-body-sm);
  color: var(--text-primary);
}

.log-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-2);
  height: 100%;
  color: var(--text-tertiary);
  padding: var(--spacing-8);
}

.log-empty svg {
  opacity: 0.4;
}

.log-empty-title {
  font-size: var(--text-h3);
  font-weight: var(--font-semibold);
  color: var(--text-secondary);
}

.log-empty-desc {
  font-size: var(--text-body-sm);
  color: var(--text-tertiary);
}
</style>