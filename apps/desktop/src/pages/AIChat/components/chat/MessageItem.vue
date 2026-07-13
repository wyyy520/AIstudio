<template>
  <div class="message-item" :class="`message-item--${message.role}`">
    <div v-if="message.role === 'assistant'" class="message-avatar">
      <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <path d="M12 2L2 7l10 5 10-5-10-5z" />
        <path d="M2 17l10 5 10-5" />
        <path d="M2 12l10 5 10-5" />
      </svg>
    </div>
    <div class="message-body">
      <div v-if="message.role === 'assistant' && message.model" class="message-meta">
        <span class="message-model">{{ message.model }}</span>
        <TokenBadge :usage="message.tokenUsage" :duration="message.duration" />
      </div>

      <div v-if="message.plan && message.plan.length > 0" class="message-plan">
        <div class="plan-header">执行计划</div>
        <div v-for="item in message.plan" :key="item.id" class="plan-item" :class="`plan-item--${item.status}`">
          <span class="plan-icon">
            <svg v-if="item.status === 'completed'" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12" /></svg>
            <svg v-else-if="item.status === 'running'" class="spin" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12a9 9 0 1 1-6.219-8.56" /></svg>
            <span v-else class="plan-dot" />
          </span>
          <span class="plan-action">{{ item.action }}</span>
          <span class="plan-desc">{{ item.description }}</span>
        </div>
      </div>

      <div class="message-bubble selectable">
        <div class="message-content" v-html="renderedContent" />
        <div v-if="message.isStreaming" class="message-cursor" />
        <button
          v-if="!message.isStreaming && message.content"
          class="message-copy-btn"
          title="复制"
          @click="handleCopy"
        >
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
            <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
          </svg>
          <span class="copy-tooltip">{{ copyTip }}</span>
        </button>
      </div>

      <div v-if="message.toolCalls && message.toolCalls.length > 0" class="message-toolcalls">
        <div class="toolcalls-header">Tool 调用</div>
        <div
          v-for="tc in message.toolCalls"
          :key="tc.id"
          class="toolcall-item"
          :class="`toolcall-item--${tc.status}`"
        >
          <span class="toolcall-icon">
            <svg v-if="tc.status === 'completed'" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12" /></svg>
            <svg v-else-if="tc.status === 'running' || tc.status === 'pending'" class="spin" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12a9 9 0 1 1-6.219-8.56" /></svg>
            <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
          </span>
          <span class="toolcall-type-badge" :class="`badge--${tc.type}`">{{ typeLabel(tc.type) }}</span>
          <span class="toolcall-name">{{ tc.name }}</span>
          <span class="toolcall-desc">{{ tc.description }}</span>
        </div>
      </div>

      <TaskCard v-if="message.task" :task="message.task" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import TokenBadge from '../shared/TokenBadge.vue'
import TaskCard from './TaskCard.vue'
import { renderMarkdown, copyToClipboard } from '../../utils/markdown'
import type { ChatMessage } from '../../types'

const props = defineProps<{
  message: ChatMessage
}>()

const copyTip = ref('')

const renderedContent = computed(() => {
  if (!props.message.content) return ''
  return renderMarkdown(props.message.content)
})

function typeLabel(type: string): string {
  const map: Record<string, string> = {
    plugin: 'Plugin',
    workflow: 'Workflow',
    task: 'Task',
    mcp: 'MCP',
    code: 'Code',
  }
  return map[type] || type
}

async function handleCopy() {
  try {
    await copyToClipboard(props.message.content)
    copyTip.value = '已复制'
    setTimeout(() => { copyTip.value = '' }, 2000)
  } catch {
    copyTip.value = '复制失败'
  }
}
</script>

<style scoped>
.message-item {
  display: flex;
  gap: var(--spacing-3);
  padding: var(--spacing-4) 0;
}

.message-item--user {
  flex-direction: row-reverse;
}

.message-avatar {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-md);
  background: var(--primary-bg);
  color: var(--primary);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.message-body {
  max-width: 720px;
  min-width: 0;
}

.message-item--user .message-body {
  text-align: right;
}

.message-meta {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-1);
}

.message-model {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  font-weight: var(--font-semibold);
}

.message-plan {
  background: var(--bg-primary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: var(--spacing-3);
  margin-bottom: var(--spacing-2);
}

.plan-header {
  font-size: var(--text-caption);
  font-weight: var(--font-semibold);
  color: var(--text-tertiary);
  margin-bottom: var(--spacing-2);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.plan-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-1) 0;
  font-size: var(--text-body-sm);
}

.plan-icon {
  display: flex;
  align-items: center;
  width: 16px;
  flex-shrink: 0;
}

.plan-item--completed .plan-icon { color: var(--success); }
.plan-item--running .plan-icon { color: var(--primary); }
.plan-item--error .plan-icon { color: var(--error); }
.plan-item--pending .plan-icon { color: var(--text-disabled); }

.plan-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  border: 1.5px solid currentColor;
}

.plan-action {
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  white-space: nowrap;
}

.plan-desc {
  color: var(--text-tertiary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.message-bubble {
  position: relative;
  background: var(--bg-tertiary);
  border-radius: var(--radius-lg);
  padding: var(--spacing-3) var(--spacing-4);
  line-height: 1.6;
}

.message-item--user .message-bubble {
  background: var(--primary);
  color: white;
  border-radius: var(--radius-lg) var(--radius-lg) var(--radius-sm) var(--radius-lg);
}

.message-item--assistant .message-bubble {
  border-radius: var(--radius-lg) var(--radius-lg) var(--radius-lg) var(--radius-sm);
}

.message-content :deep(h1),
.message-content :deep(h2),
.message-content :deep(h3),
.message-content :deep(h4) {
  font-weight: var(--font-semibold);
  margin: var(--spacing-3) 0 var(--spacing-2);
  color: var(--text-primary);
}

.message-content :deep(h1) { font-size: 1.3em; }
.message-content :deep(h2) { font-size: 1.2em; }
.message-content :deep(h3) { font-size: 1.1em; }
.message-content :deep(h4) { font-size: 1em; }

.message-content :deep(p) {
  margin: var(--spacing-2) 0;
}

.message-content :deep(ul),
.message-content :deep(ol) {
  padding-left: var(--spacing-5);
  margin: var(--spacing-2) 0;
}

.message-content :deep(li) {
  margin: var(--spacing-1) 0;
}

.message-content :deep(pre) {
  background: var(--bg-primary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: var(--spacing-3);
  margin: var(--spacing-2) 0;
  overflow-x: auto;
  font-family: var(--font-family-mono);
  font-size: var(--text-code);
  line-height: var(--leading-code);
}

.message-content :deep(code) {
  font-family: var(--font-family-mono);
  font-size: 0.9em;
  background: var(--bg-active);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
}

.message-content :deep(pre code) {
  background: none;
  padding: 0;
  border-radius: 0;
}

.message-content :deep(blockquote) {
  border-left: 3px solid var(--primary);
  padding-left: var(--spacing-3);
  margin: var(--spacing-2) 0;
  color: var(--text-secondary);
}

.message-content :deep(strong) {
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.message-content :deep(a) {
  color: var(--primary);
  text-decoration: none;
}

.message-content :deep(a:hover) {
  text-decoration: underline;
}

.message-content :deep(table) {
  border-collapse: collapse;
  width: 100%;
  margin: var(--spacing-2) 0;
  font-size: var(--text-body-sm);
}

.message-content :deep(th),
.message-content :deep(td) {
  border: 1px solid var(--border-subtle);
  padding: var(--spacing-1) var(--spacing-2);
  text-align: left;
}

.message-content :deep(th) {
  background: var(--bg-hover);
  font-weight: var(--font-semibold);
}

.message-content :deep(hr) {
  border: none;
  border-top: 1px solid var(--border-subtle);
  margin: var(--spacing-3) 0;
}

.message-cursor {
  display: inline-block;
  width: 2px;
  height: 1em;
  background: var(--primary);
  margin-left: 2px;
  animation: cursorPulse 1s ease-in-out infinite;
  vertical-align: text-bottom;
}

@keyframes cursorPulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

.message-copy-btn {
  position: absolute;
  top: var(--spacing-2);
  right: var(--spacing-2);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  background: var(--bg-hover);
  color: var(--text-tertiary);
  border-radius: var(--radius-sm);
  cursor: pointer;
  opacity: 0;
  transition: all var(--transition-fast);
}

.message-bubble:hover .message-copy-btn {
  opacity: 1;
}

.message-copy-btn:hover {
  background: var(--bg-active);
  color: var(--text-primary);
}

.copy-tooltip {
  position: absolute;
  top: -24px;
  left: 50%;
  transform: translateX(-50%);
  padding: 2px 8px;
  background: var(--bg-primary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  font-size: var(--text-caption);
  color: var(--text-primary);
  white-space: nowrap;
  pointer-events: none;
}

.message-toolcalls {
  margin-top: var(--spacing-2);
  background: var(--bg-primary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: var(--spacing-3);
}

.toolcalls-header {
  font-size: var(--text-caption);
  font-weight: var(--font-semibold);
  color: var(--text-tertiary);
  margin-bottom: var(--spacing-2);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.toolcall-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-1) 0;
  font-size: var(--text-body-sm);
}

.toolcall-icon {
  display: flex;
  align-items: center;
  width: 16px;
  flex-shrink: 0;
}

.toolcall-item--completed .toolcall-icon { color: var(--success); }
.toolcall-item--running .toolcall-icon,
.toolcall-item--pending .toolcall-icon { color: var(--primary); }
.toolcall-item--error .toolcall-icon { color: var(--error); }

.toolcall-type-badge {
  display: inline-flex;
  align-items: center;
  height: 18px;
  padding: 0 6px;
  border-radius: var(--radius-sm);
  font-size: 10px;
  font-weight: var(--font-semibold);
  text-transform: uppercase;
  letter-spacing: 0.03em;
  flex-shrink: 0;
}

.badge--plugin { background: #e0f2fe; color: #0369a1; }
.badge--workflow { background: #fae8ff; color: #a21caf; }
.badge--task { background: #fef3c7; color: #b45309; }
.badge--mcp { background: #d1fae5; color: #065f46; }
.badge--code { background: #dbeafe; color: #1d4ed8; }

.toolcall-name {
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  white-space: nowrap;
}

.toolcall-desc {
  color: var(--text-tertiary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
