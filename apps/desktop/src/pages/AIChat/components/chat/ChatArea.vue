<template>
  <div class="chat-area">
    <MessageList :messages="messages" :agent-status="agentStatus" />
    <div class="chat-area-input">
      <ActionButtons :actions="quickActions" @action="handleAction" />
      <InputBox
        v-model="inputValue"
        :disabled="agentStatus !== 'idle' && agentStatus !== 'finished' && agentStatus !== 'error'"
        @send="handleSend"
        @stop="$emit('stop')"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import MessageList from './MessageList.vue'
import ActionButtons from './ActionButtons.vue'
import InputBox from './InputBox.vue'
import type { ChatMessage, AgentStatus, QuickAction, Attachment } from '../../types'

defineProps<{
  messages: ChatMessage[]
  agentStatus: AgentStatus
  quickActions: QuickAction[]
}>()

const emit = defineEmits<{
  send: [content: string, attachments: Attachment[]]
  stop: []
}>()

const inputValue = ref('')

function handleSend(content: string, attachments: Attachment[]) {
  emit('send', content, attachments)
}

function handleAction(action: QuickAction) {
  inputValue.value = action.prompt
  emit('send', action.prompt, [])
  inputValue.value = ''
}
</script>

<style scoped>
.chat-area {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-width: 0;
}

.chat-area-input {
  flex-shrink: 0;
  padding: 0 var(--spacing-6) var(--spacing-4);
}
</style>
