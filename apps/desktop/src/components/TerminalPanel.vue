<template>
  <div class="terminal-panel">
    <div class="terminal-header">
      <div class="terminal-tabs">
        <button
          v-for="tab in tabs" :key="tab.key"
          :class="['terminal-tab', { active: activeTab === tab.key }]"
          @click="activeTab = tab.key"
        >
          {{ tab.label }}
        </button>
      </div>
      <div class="terminal-actions">
        <label class="raw-toggle">
          <input type="checkbox" v-model="showRaw" />
          <span>Raw</span>
        </label>
        <button class="terminal-action-btn" @click="clearLogs" title="Clear">✕</button>
      </div>
    </div>
    <div ref="logContainer" class="terminal-body">
      <div v-if="displayLogs.length === 0" class="terminal-empty">Output will appear here...</div>
      <div v-for="(log, i) in displayLogs" :key="i" :class="['log-line', `log-line--${log.level}`]">
        <span class="log-time">{{ log.time }}</span>
        <span class="log-level">{{ log.level.toUpperCase() }}</span>
        <span class="log-msg">{{ log.message }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { wsClient, type WebSocketEvent } from '@/api/websocket'

interface LogLine {
  time: string
  level: 'info' | 'warn' | 'error' | 'debug'
  message: string
}

const props = withDefaults(defineProps<{
  height?: string
}>(), {
  height: '200px',
})

const activeTab = ref<'output' | 'logs'>('logs')
const showRaw = ref(false)
const logs = ref<LogLine[]>([])
const logContainer = ref<HTMLElement>()

const tabs = [
  { key: 'logs', label: 'Logs' },
  { key: 'output', label: 'Output' },
]

const displayLogs = computed(() => {
  if (showRaw.value) return logs.value
  return logs.value.filter(l => l.level !== 'debug')
})

function formatTime(ts: string) {
  const d = new Date(ts)
  return d.toLocaleTimeString('zh-CN', { hour12: false })
}

function addLog(level: LogLine['level'], message: string, timestamp?: string) {
  logs.value.push({
    time: timestamp ? formatTime(timestamp) : new Date().toLocaleTimeString('zh-CN', { hour12: false }),
    level,
    message,
  })
  if (logs.value.length > 1000) {
    logs.value = logs.value.slice(-500)
  }
  scrollToBottom()
}

function scrollToBottom() {
  nextTick(() => {
    if (logContainer.value) {
      logContainer.value.scrollTop = logContainer.value.scrollHeight
    }
  })
}

function clearLogs() {
  logs.value = []
}

let unsub: (() => void) | null = null

function startListening() {
  wsClient.connect()
  unsub = wsClient.subscribe((event: WebSocketEvent) => {
    if (event.type === 'task_log' && event.data.message) {
      const level: LogLine['level'] = 
        event.data.level === 'error' ? 'error' :
        event.data.level === 'warn' ? 'warn' :
        event.data.level === 'debug' ? 'debug' : 'info'
      addLog(level, event.data.message, event.data.timestamp)
    }
    if (event.type === 'runtime:log' && (event.data as any).message) {
      addLog('info', (event.data as any).message, event.data.timestamp)
    }
  })
}

function stopListening() {
  if (unsub) {
    unsub()
    unsub = null
  }
}

defineExpose({ startListening, stopListening, clearLogs, addLog })
</script>

<style scoped>
.terminal-panel {
  height: v-bind(height);
  background: var(--bg-secondary);
  border-top: 1px solid var(--border-subtle);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.terminal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 32px;
  padding: 0 8px;
  border-bottom: 1px solid var(--border-subtle);
  flex-shrink: 0;
}

.terminal-tabs {
  display: flex;
  gap: 2px;
}

.terminal-tab {
  height: 32px;
  padding: 0 12px;
  border: none;
  border-bottom: 2px solid transparent;
  background: transparent;
  color: var(--text-tertiary);
  font-size: 12px;
  cursor: pointer;
}

.terminal-tab.active {
  color: var(--text-primary);
  border-bottom-color: var(--primary);
}

.terminal-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.raw-toggle {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: var(--text-tertiary);
  cursor: pointer;
}

.terminal-action-btn {
  width: 20px;
  height: 20px;
  border: none;
  border-radius: 4px;
  background: transparent;
  color: var(--text-tertiary);
  cursor: pointer;
  font-size: 11px;
}

.terminal-action-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.terminal-body {
  flex: 1;
  overflow-y: auto;
  padding: 4px 0;
  font-family: var(--font-mono);
  font-size: 12px;
  line-height: 1.5;
}

.terminal-empty {
  padding: 24px 16px;
  color: var(--text-tertiary);
  text-align: center;
}

.log-line {
  display: flex;
  gap: 8px;
  padding: 1px 12px;
}

.log-line:hover {
  background: var(--bg-hover);
}

.log-time {
  color: var(--text-tertiary);
  flex-shrink: 0;
}

.log-level {
  width: 36px;
  flex-shrink: 0;
  font-weight: 600;
}

.log-line--info .log-level { color: var(--info); }
.log-line--warn .log-level { color: var(--warning); }
.log-line--error { background: var(--error-bg); }
.log-line--error .log-level { color: var(--error); }
.log-line--debug .log-level { color: var(--text-tertiary); }

.log-msg {
  color: var(--text-secondary);
  word-break: break-all;
}

.log-line--error .log-msg {
  color: var(--error);
}
</style>
