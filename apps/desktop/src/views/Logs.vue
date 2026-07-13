<template>
  <div class="logs">
    <div class="logs__toolbar">
      <h3 class="logs__title">日志</h3>
      <div class="logs__controls">
        <select v-model="levelFilter" class="level-select">
          <option value="all">全部级别</option>
          <option value="INFO">INFO</option>
          <option value="WARN">WARN</option>
          <option value="ERROR">ERROR</option>
        </select>
        <input v-model="searchText" placeholder="搜索日志..." class="search-input" />
        <AppButton type="secondary" size="small" @click="refreshLogs">刷新</AppButton>
        <AppButton type="ghost" size="small" @click="clearLogs">清空</AppButton>
      </div>
    </div>

    <div ref="logContainerRef" class="logs__content">
      <div v-if="logEntries.length === 0" class="logs__empty">暂无日志</div>
      <div v-else class="log-list">
        <div
          v-for="entry in filteredLogs"
          :key="entry.id || entry.timestamp"
          :class="['log-entry', `log-entry--${entry.level?.toLowerCase()}`]"
        >
          <span class="log-entry__time">{{ formatTime(entry.timestamp) }}</span>
          <span :class="['log-entry__level', `log-entry__level--${entry.level?.toLowerCase()}`]">{{ entry.level }}</span>
          <span class="log-entry__source">{{ entry.source }}</span>
          <span class="log-entry__message">{{ entry.message }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { queryLogs } from '@/api/log'
import AppButton from '@/components/AppButton.vue'

const route = useRoute()

interface LogEntry {
  id?: string
  level: string
  message: string
  source: string
  timestamp: string
}

const logEntries = ref<LogEntry[]>([])
const levelFilter = ref('all')
const searchText = ref('')
const taskFilter = ref<string | undefined>()
const logContainerRef = ref<HTMLElement>()
let ws: WebSocket | null = null
let pollTimer: ReturnType<typeof setInterval> | null = null

const filteredLogs = computed(() => {
  let list = logEntries.value
  if (levelFilter.value !== 'all') {
    list = list.filter(e => e.level === levelFilter.value)
  }
  if (searchText.value) {
    const s = searchText.value.toLowerCase()
    list = list.filter(e => e.message?.toLowerCase().includes(s) || e.source?.toLowerCase().includes(s))
  }
  return list
})

function formatTime(ts: string) {
  if (!ts) return ''
  const d = new Date(ts)
  return d.toLocaleTimeString('zh-CN', { hour12: false })
}

async function refreshLogs() {
  try {
    const res: any = await queryLogs({
      level: levelFilter.value === 'all' ? undefined : levelFilter.value,
      taskId: taskFilter.value,
      limit: 200,
    })
    logEntries.value = res.data || []
  } catch {
    logEntries.value = []
  }
}

function clearLogs() {
  logEntries.value = []
}

function connectWebSocket() {
  const envWS = import.meta.env.VITE_WS_URL
  let wsBase = ''
  if (envWS) {
    wsBase = envWS
  } else if (typeof window !== 'undefined') {
    const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    wsBase = `${proto}//${window.location.host}`
  }
  const wsUrl = `${wsBase}/api/v1/ws`
  try {
    ws = new WebSocket(wsUrl)
    ws.onmessage = (event) => {
      try {
        const entry = JSON.parse(event.data)
        logEntries.value.push(entry)
        if (logEntries.value.length > 1000) {
          logEntries.value = logEntries.value.slice(-500)
        }
        scrollToBottom()
      } catch {}
    }
    ws.onerror = () => {
      ws = null
      startPolling()
    }
    ws.onclose = () => {
      ws = null
      startPolling()
    }
  } catch {
    startPolling()
  }
}

function startPolling() {
  if (pollTimer) return
  pollTimer = setInterval(() => {
    refreshLogs()
  }, 5000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

function scrollToBottom() {
  requestAnimationFrame(() => {
    if (logContainerRef.value) {
      logContainerRef.value.scrollTop = logContainerRef.value.scrollHeight
    }
  })
}

onMounted(() => {
  const taskId = route.params.taskId
  if (taskId && typeof taskId === 'string' && taskId !== 'all') {
    taskFilter.value = taskId
  }
  refreshLogs()
  connectWebSocket()
})

onUnmounted(() => {
  if (ws) {
    ws.close()
    ws = null
  }
  stopPolling()
})
</script>

<style scoped>
.logs {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.logs__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  border-bottom: 1px solid var(--border-subtle);
  background: var(--bg-secondary);
  flex-shrink: 0;
}

.logs__title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.logs__controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.level-select {
  height: 32px;
  padding: 0 10px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-xs);
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
}

.search-input {
  width: 200px;
  height: 32px;
  padding: 0 10px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-xs);
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
}

.search-input:focus {
  border-color: var(--primary);
}

.logs__content {
  flex: 1;
  overflow-y: auto;
  padding: 0;
  font-family: var(--font-mono);
  font-size: 12px;
}

.logs__empty {
  text-align: center;
  padding: 48px 0;
  color: var(--text-tertiary);
}

.log-list {
  display: flex;
  flex-direction: column;
}

.log-entry {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 6px 16px;
  border-bottom: 1px solid var(--border-subtle);
  line-height: 1.6;
}

.log-entry:hover {
  background: var(--bg-hover);
}

.log-entry--error {
  background: rgba(239, 68, 68, 0.05);
}

.log-entry--warn {
  background: rgba(245, 158, 11, 0.03);
}

.log-entry__time {
  color: var(--text-tertiary);
  white-space: nowrap;
  flex-shrink: 0;
}

.log-entry__level {
  font-weight: 600;
  min-width: 44px;
  flex-shrink: 0;
}

.log-entry__level--info { color: var(--info); }
.log-entry__level--warn { color: var(--warning); }
.log-entry__level--error { color: var(--error); }

.log-entry__source {
  color: var(--text-tertiary);
  min-width: 80px;
  flex-shrink: 0;
}

.log-entry__message {
  color: var(--text-primary);
  word-break: break-all;
}
</style>