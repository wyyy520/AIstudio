<template>
  <div class="chat-toolbar">
    <div class="chat-toolbar-left">
      <ModelSelector :providers="providers" :selected="selectedModel" @select="$emit('model-change', $event)" />
    </div>
    <div class="chat-toolbar-center">
      <span class="chat-toolbar-title">AI Chat</span>
    </div>
    <div class="chat-toolbar-right">
      <button class="toolbar-btn" @click="$emit('new-chat')" title="新建对话">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <line x1="12" y1="5" x2="12" y2="19" />
          <line x1="5" y1="12" x2="19" y2="12" />
        </svg>
      </button>
      <button class="toolbar-btn" @click="$emit('toggle-history')" title="历史记录">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10" />
          <polyline points="12 6 12 12 16 14" />
        </svg>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import ModelSelector from './ModelSelector.vue'
import type { AIProvider } from '../../types'

defineProps<{
  providers: AIProvider[]
  selectedModel: string
}>()

defineEmits<{
  'model-change': [modelId: string]
  'new-chat': []
  'toggle-history': []
}>()
</script>

<style scoped>
.chat-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 44px;
  padding: 0 var(--spacing-4);
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-subtle);
  flex-shrink: 0;
}

.chat-toolbar-left,
.chat-toolbar-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.chat-toolbar-center {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
}

.chat-toolbar-title {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-secondary);
}

.toolbar-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  color: var(--text-tertiary);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.toolbar-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}
</style>
