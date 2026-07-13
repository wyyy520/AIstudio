<template>
  <div class="task-list">
    <div class="task-list-header">
      <div class="task-list-search">
        <svg class="search-icon" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="11" cy="11" r="8" /><path d="m21 21-4.3-4.3" />
        </svg>
        <input
          type="text"
          class="search-input"
          placeholder="搜索任务..."
          :value="searchQuery"
          @input="$emit('update:searchQuery', ($event.target as HTMLInputElement).value)"
        />
      </div>
    </div>

    <div class="task-list-body">
      <template v-for="group in groups" :key="group.key">
        <div v-if="group.tasks.length" class="task-group">
          <div class="task-group-header" @click="toggleGroup(group.key)">
            <svg class="chevron-icon" :class="{ collapsed: collapsedGroups.has(group.key) }" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="m6 9 6 6 6-6" />
            </svg>
            <span class="task-group-label">{{ group.label }}</span>
            <span class="task-group-count">{{ group.tasks.length }}</span>
          </div>
          <div v-show="!collapsedGroups.has(group.key)" class="task-group-items">
            <div
              v-for="task in group.tasks"
              :key="task.id"
              class="task-card"
              :class="{ active: task.id === selectedTaskId }"
              @click="$emit('select', task.id)"
            >
              <div class="task-card-row">
                <span class="task-status-dot" :class="`status-${task.status}`"></span>
                <span class="task-name">{{ task.name }}</span>
              </div>
              <div class="task-card-row task-card-meta">
                <AppTag :color="typeColor(task.type)" size="small">{{ typeLabel(task.type) }}</AppTag>
                <span class="task-time">{{ formatTime(task.startedAt) }}</span>
              </div>
              <div class="task-card-row task-card-status">
                <span class="task-status-text" :class="`text-${task.status}`">{{ statusLabel(task.status) }}</span>
                <span class="task-duration">{{ formatDuration(task.duration) }}</span>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>

    <div class="task-list-footer">
      <span class="footer-stat">{{ tasks.length }} 个任务</span>
      <span class="footer-dot">·</span>
      <span class="footer-stat">{{ runningCount }} 运行中</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Task, TaskType, TaskStatus } from '@/pages/Logs/types'
import AppTag from '@/components/AppTag/AppTag.vue'

interface Props {
  tasks: Task[]
  selectedTaskId: string | null
  searchQuery: string
  runningCount: number
}

const props = defineProps<Props>()

defineEmits<{
  select: [taskId: string]
  'update:searchQuery': [value: string]
}>()

const collapsedGroups = ref(new Set<string>())

const groups = computed(() => {
  const q = props.searchQuery.toLowerCase().trim()
  let list = props.tasks
  if (q) {
    list = list.filter(t => t.name.toLowerCase().includes(q))
  }
  return [
    { key: 'running', label: '运行中', tasks: list.filter(t => t.status === 'running') },
    { key: 'failed', label: '失败', tasks: list.filter(t => t.status === 'failed') },
    { key: 'warning', label: '警告', tasks: list.filter(t => t.status === 'warning') },
    { key: 'completed', label: '已完成', tasks: list.filter(t => t.status === 'success') },
  ]
})

function toggleGroup(key: string) {
  if (collapsedGroups.value.has(key)) {
    collapsedGroups.value.delete(key)
  } else {
    collapsedGroups.value.add(key)
  }
}

function typeLabel(type: TaskType): string {
  const map: Record<TaskType, string> = {
    training: '训练',
    simulation: '仿真',
    export: '导出',
    workflow: '工作流',
    system: '系统',
    agent: 'Agent',
  }
  return map[type]
}

function typeColor(type: TaskType): string {
  const map: Record<TaskType, string> = {
    training: 'primary',
    simulation: 'info',
    export: 'default',
    workflow: 'primary',
    system: 'default',
    agent: 'error',
  }
  return map[type]
}

function statusLabel(status: TaskStatus): string {
  const map: Record<TaskStatus, string> = {
    running: '运行中',
    success: '完成',
    failed: '失败',
    warning: '警告',
  }
  return map[status]
}

function formatTime(iso: string): string {
  const d = new Date(iso)
  return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

function formatDuration(seconds: number): string {
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return m > 0 ? `${m}m ${s}s` : `${s}s`
}
</script>

<style scoped>
.task-list {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-subtle);
}

.task-list-header {
  padding: var(--spacing-3);
  border-bottom: 1px solid var(--border-subtle);
}

.task-list-search {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  height: 32px;
  padding: 0 var(--spacing-3);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  transition: border-color var(--transition-fast);
}

.task-list-search:focus-within {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px var(--primary-bg);
}

.search-icon {
  flex-shrink: 0;
  color: var(--text-tertiary);
}

.search-input {
  flex: 1;
  font-size: var(--text-caption);
  color: var(--text-primary);
  background: transparent;
  border: none;
  outline: none;
}

.search-input::placeholder {
  color: var(--text-tertiary);
}

.task-list-body {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-2);
}

.task-group {
  margin-bottom: var(--spacing-2);
}

.task-group-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  padding: var(--spacing-1) var(--spacing-2);
  cursor: pointer;
  border-radius: var(--radius-sm);
  user-select: none;
  transition: background var(--transition-fast);
}

.task-group-header:hover {
  background: var(--bg-hover);
}

.chevron-icon {
  flex-shrink: 0;
  color: var(--text-tertiary);
  transition: transform var(--transition-fast);
}

.chevron-icon.collapsed {
  transform: rotate(-90deg);
}

.task-group-label {
  font-size: var(--text-caption);
  font-weight: var(--font-semibold);
  color: var(--text-secondary);
  flex: 1;
}

.task-group-count {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.task-group-items {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1);
  padding-top: var(--spacing-1);
}

.task-card {
  padding: var(--spacing-2) var(--spacing-3);
  border-radius: var(--radius-lg);
  cursor: pointer;
  transition: background var(--transition-fast);
  border: 1px solid transparent;
}

.task-card:hover {
  background: var(--bg-hover);
}

.task-card.active {
  background: var(--bg-active);
  border-color: var(--primary);
}

.task-card-row {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.task-card-meta {
  margin-top: var(--spacing-1);
  padding-left: 20px;
}

.task-card-status {
  margin-top: 2px;
  padding-left: 20px;
}

.task-status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.task-status-dot.status-running {
  background: var(--info);
  animation: pulse 2s ease-in-out infinite;
}

.task-status-dot.status-success {
  background: var(--success);
}

.task-status-dot.status-failed {
  background: var(--error);
}

.task-status-dot.status-warning {
  background: var(--warning);
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.task-name {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.task-time {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  font-family: var(--font-family-mono);
}

.task-status-text {
  font-size: var(--text-caption);
  font-weight: var(--font-semibold);
}

.task-status-text.text-running { color: var(--info); }
.task-status-text.text-success { color: var(--success); }
.task-status-text.text-failed { color: var(--error); }
.task-status-text.text-warning { color: var(--warning); }

.task-duration {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  font-family: var(--font-family-mono);
  margin-left: auto;
}

.task-list-footer {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-2) var(--spacing-3);
  border-top: 1px solid var(--border-subtle);
}

.footer-stat {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.footer-dot {
  color: var(--text-disabled);
}
</style>