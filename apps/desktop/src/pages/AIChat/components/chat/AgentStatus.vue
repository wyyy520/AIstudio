<template>
  <div class="agent-status">
    <div class="agent-status-indicator">
      <span class="agent-status-dot" :class="`agent-status-dot--${status}`" />
      <span class="agent-status-text">{{ statusText }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { AgentStatus } from '../../types'

const props = defineProps<{
  status: AgentStatus
}>()

const statusText = computed(() => {
  const map: Record<AgentStatus, string> = {
    idle: '',
    thinking: 'AI 正在分析...',
    planning: '正在制定计划...',
    calling_tool: '正在调用工具...',
    running: '正在执行任务...',
    finished: '已完成',
    error: '发生错误',
  }
  return map[props.status]
})
</script>

<style scoped>
.agent-status {
  display: flex;
  justify-content: flex-start;
  padding: var(--spacing-2) 0;
}

.agent-status-indicator {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-1) var(--spacing-3);
  background: var(--bg-tertiary);
  border-radius: var(--radius-md);
}

.agent-status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}

.agent-status-dot--thinking,
.agent-status-dot--planning,
.agent-status-dot--calling_tool,
.agent-status-dot--running {
  background: var(--primary);
  animation: pulse 1.5s ease-in-out infinite;
}

.agent-status-dot--finished {
  background: var(--success);
}

.agent-status-dot--error {
  background: var(--error);
}

.agent-status-text {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}
</style>
