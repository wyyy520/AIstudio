<template>
  <div class="message-list" ref="listRef">
    <div v-if="messages.length === 0" class="message-list-empty">
      <svg viewBox="0 0 24 24" width="48" height="48" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round">
        <path d="M12 2L2 7l10 5 10-5-10-5z" />
        <path d="M2 17l10 5 10-5" />
        <path d="M2 12l10 5 10-5" />
      </svg>
      <p class="message-list-empty-title">开始新的对话</p>
      <p class="message-list-empty-desc">输入消息，开始与 AI 助手交流</p>
    </div>
    <template v-else>
      <MessageItem
        v-for="message in messages"
        :key="message.id"
        :message="message"
      />
    </template>
    <AgentStatus v-if="agentStatus !== 'idle' && agentStatus !== 'finished' && agentStatus !== 'error'" :status="agentStatus" />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import MessageItem from './MessageItem.vue'
import AgentStatus from './AgentStatus.vue'
import type { ChatMessage, AgentStatus as AgentStatusType } from '../../types'

const props = defineProps<{
  messages: ChatMessage[]
  agentStatus: AgentStatusType
}>()

const listRef = ref<HTMLElement>()

watch(
  () => props.messages.length,
  async () => {
    await nextTick()
    scrollToBottom()
  }
)

watch(
  () => props.messages[props.messages.length - 1]?.content,
  async () => {
    await nextTick()
    scrollToBottom()
  }
)

function scrollToBottom() {
  if (listRef.value) {
    listRef.value.scrollTop = listRef.value.scrollHeight
  }
}
</script>

<style scoped>
.message-list {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-4) var(--spacing-6);
}

.message-list-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-disabled);
  gap: var(--spacing-3);
}

.message-list-empty-title {
  font-size: var(--text-h3);
  color: var(--text-tertiary);
}

.message-list-empty-desc {
  font-size: var(--text-body-sm);
  color: var(--text-disabled);
}
</style>
