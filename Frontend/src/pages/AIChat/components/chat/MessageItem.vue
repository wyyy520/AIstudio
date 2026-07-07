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
      <div class="message-bubble selectable">
        <div class="message-content" v-html="renderedContent" />
      </div>
      <TaskCard v-if="message.task" :task="message.task" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import TokenBadge from '../shared/TokenBadge.vue'
import TaskCard from './TaskCard.vue'
import type { ChatMessage } from '../../types'

const props = defineProps<{
  message: ChatMessage
}>()

const renderedContent = computed(() => {
  let content = props.message.content
  // Basic markdown rendering for code blocks
  content = content.replace(/```(\w*)\n([\s\S]*?)```/g, '<pre class="code-block"><code>$2</code></pre>')
  // Inline code
  content = content.replace(/`([^`]+)`/g, '<code class="inline-code">$1</code>')
  // Bold
  content = content.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
  // Headers
  content = content.replace(/^### (.+)$/gm, '<h4>$1</h4>')
  content = content.replace(/^## (.+)$/gm, '<h3>$1</h3>')
  content = content.replace(/^# (.+)$/gm, '<h2>$1</h2>')
  // Lists
  content = content.replace(/^- (.+)$/gm, '<li>$1</li>')
  content = content.replace(/(<li>.*<\/li>\n?)+/g, '<ul>$&</ul>')
  // Line breaks
  content = content.replace(/\n/g, '<br>')
  return content
})
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

.message-bubble {
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

.message-content :deep(h2) {
  font-size: var(--text-h3);
  font-weight: var(--font-semibold);
  margin: var(--spacing-3) 0 var(--spacing-2);
  color: var(--text-primary);
}

.message-content :deep(h3) {
  font-size: var(--text-body);
  font-weight: var(--font-semibold);
  margin: var(--spacing-2) 0 var(--spacing-1);
  color: var(--text-primary);
}

.message-content :deep(h4) {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  margin: var(--spacing-2) 0 var(--spacing-1);
  color: var(--text-primary);
}

.message-content :deep(.code-block) {
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

.message-content :deep(.inline-code) {
  background: var(--bg-active);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  font-family: var(--font-family-mono);
  font-size: var(--text-code);
}

.message-content :deep(ul) {
  padding-left: var(--spacing-4);
  margin: var(--spacing-2) 0;
}

.message-content :deep(li) {
  margin: var(--spacing-1) 0;
}

.message-content :deep(strong) {
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}
</style>
