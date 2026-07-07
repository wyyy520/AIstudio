<template>
  <div class="raw-log-viewer">
    <div class="raw-log-toolbar">
      <button class="toolbar-btn" :class="{ active: wrapEnabled }" @click="wrapEnabled = !wrapEnabled" title="Toggle Wrap">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M3 6h18M3 12h15a3 3 0 1 1 0 6h-4" /><path d="m14 15-2 2 2 2" /><path d="M3 18h7" />
        </svg>
        Wrap
      </button>
      <button class="toolbar-btn" @click="copyAll" title="Copy All">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <rect width="14" height="14" x="8" y="8" rx="2" ry="2" /><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2" />
        </svg>
        Copy All
      </button>
      <button class="toolbar-btn" @click="$emit('download')" title="Download">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" /><path d="m7 10 5 5 5-5" /><path d="M12 15V3" />
        </svg>
        Download
      </button>
    </div>
    <div class="raw-log-content selectable" ref="logContainer">
      <div
        v-for="(entry, idx) in logs"
        :key="entry.id"
        class="raw-log-line"
        :class="`level-${entry.level}`"
      >
        <span class="line-number">{{ idx + 1 }}</span>
        <span class="line-timestamp">{{ formatTimestamp(entry.timestamp) }}</span>
        <span class="line-level" :class="`level-tag-${entry.level}`">[{{ entry.level.toUpperCase() }}]</span>
        <span class="line-message" :style="{ whiteSpace: wrapEnabled ? 'pre-wrap' : 'nowrap' }">{{ entry.rawMessage }}</span>
      </div>
      <div v-if="!logs.length" class="raw-log-empty">
        <svg viewBox="0 0 24 24" width="32" height="32" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" /><path d="M14 2v6h6" /><path d="M16 13H8" /><path d="M16 17H8" /><path d="M10 9H8" />
        </svg>
        <span>No raw logs available</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { LogEntry } from '@/pages/Logs/types'

interface Props {
  logs: LogEntry[]
}

const props = defineProps<Props>()

defineEmits<{
  download: []
}>()

const wrapEnabled = ref(false)
const logContainer = ref<HTMLElement>()

function formatTimestamp(iso: string): string {
  const d = new Date(iso)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}:${String(d.getSeconds()).padStart(2, '0')}`
}

function copyAll() {
  const text = props.logs.map(l => `[${formatTimestamp(l.timestamp)}] [${l.level.toUpperCase()}] ${l.rawMessage}`).join('\n')
  navigator.clipboard.writeText(text)
}
</script>

<style scoped>
.raw-log-viewer {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.raw-log-toolbar {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-2) var(--spacing-3);
  border-bottom: 1px solid var(--border-subtle);
}

.toolbar-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  height: 26px;
  padding: 0 var(--spacing-2);
  background: transparent;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-size: var(--text-caption);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.toolbar-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.toolbar-btn.active {
  background: var(--primary-bg);
  border-color: var(--primary);
  color: var(--primary);
}

.raw-log-content {
  flex: 1;
  overflow: auto;
  padding: var(--spacing-2) 0;
  font-family: var(--font-family-mono);
  font-size: var(--text-code);
  line-height: var(--leading-code);
}

.raw-log-line {
  display: flex;
  align-items: flex-start;
  padding: 1px var(--spacing-3);
  transition: background var(--transition-fast);
}

.raw-log-line:hover {
  background: var(--bg-hover);
}

.raw-log-line.level-error {
  background: var(--error-bg);
}

.raw-log-line.level-warning {
  background: var(--warning-bg);
}

.line-number {
  width: 40px;
  text-align: right;
  color: var(--text-disabled);
  padding-right: var(--spacing-3);
  flex-shrink: 0;
  user-select: none;
}

.line-timestamp {
  color: var(--text-tertiary);
  margin-right: var(--spacing-2);
  flex-shrink: 0;
}

.line-level {
  margin-right: var(--spacing-2);
  flex-shrink: 0;
  font-weight: var(--font-semibold);
}

.level-tag-info { color: var(--info); }
.level-tag-warning { color: var(--warning); }
.level-tag-error { color: var(--error); }
.level-tag-debug { color: var(--text-tertiary); }

.line-message {
  color: var(--text-primary);
  overflow: hidden;
}

.raw-log-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-3);
  height: 100%;
  color: var(--text-tertiary);
  font-family: var(--font-family-sans);
  font-size: var(--text-body-sm);
}
</style>