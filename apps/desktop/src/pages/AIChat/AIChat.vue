<template>
  <div class="ai-chat-page">
    <ChatToolbar
      :providers="store.providers"
      :selected-model="store.currentModelId"
      @model-change="store.selectModel"
      @new-chat="store.newConversation"
      @toggle-history="showHistory = true"
    />

    <div class="ai-chat-layout">
      <ProviderPanel
        class="ai-chat-provider"
        :providers="store.providers"
        :active-provider-id="store.currentProviderId"
        @select-provider="store.selectProvider"
        @settings="openProviderSettings"
        @toggle-history="showHistory = true"
      />

      <ChatArea
        class="ai-chat-main"
        :messages="store.messages"
        :agent-status="store.agentStatus"
        :quick-actions="quickActions"
        @send="handleSend"
        @stop="store.stopStreaming"
      />

      <ContextPanel
        class="ai-chat-context"
        :context="context"
      />
    </div>

    <ProviderSettings
      v-if="settingsProvider"
      :provider="settingsProvider"
      :visible="showSettings"
      :test-connection="handleTestProvider"
      @close="showSettings = false"
      @save="handleSaveProvider"
      @delete="handleDeleteProvider"
    />

    <HistoryDrawer
      :conversations="store.conversations"
      :visible="showHistory"
      :selected-id="store.currentConversationId"
      @close="showHistory = false"
      @select="store.selectConversation"
      @delete="store.deleteConversation"
      @favorite="store.toggleFavorite"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useChatStore } from '@/stores/chat'
import { useProjectStore } from '@/stores/project'
import ChatToolbar from './components/toolbar/ChatToolbar.vue'
import ProviderPanel from './components/provider/ProviderPanel.vue'
import ProviderSettings from './components/provider/ProviderSettings.vue'
import ChatArea from './components/chat/ChatArea.vue'
import ContextPanel from './components/context/ContextPanel.vue'
import HistoryDrawer from './components/context/HistoryDrawer.vue'
import { mockQuickActions } from './mock'
import type {
  AIProvider,
  QuickAction,
  Attachment,
  ChatContext,
} from './types'

const store = useChatStore()
const projectStore = useProjectStore()

const settingsProvider = ref<AIProvider | null>(null)
const showSettings = ref(false)
const showHistory = ref(false)

const context = reactive<ChatContext>({
  project: undefined,
  workflow: undefined,
  files: [],
  plugins: [],
  mcpServers: [],
})

const quickActions: QuickAction[] = mockQuickActions

onMounted(() => {
  if (!store.currentConversationId) {
    store.newConversation()
  } else {
    store.loadConversationMessages(store.currentConversationId)
  }

  if (projectStore.projects.length > 0) {
    const p = projectStore.projects[0]
    context.project = { id: p.id, name: p.name }
  }
})

watch(() => store.messages.length, () => {
  if (store.messages.length === 1 && store.messages[0].role === 'assistant') {
    const title = store.messages[0].content.slice(0, 30)
    store.updateConversationTitle(title)
  }
})

function openProviderSettings(provider: AIProvider) {
  settingsProvider.value = provider
  showSettings.value = true
}

function handleSaveProvider(data: Record<string, unknown>) {
  if (settingsProvider.value) {
    store.updateProvider(settingsProvider.value.id, {
      apiBaseUrl: data.apiBaseUrl as string,
      apiKey: data.apiKey as string,
    })
  }
  showSettings.value = false
}

async function handleTestProvider(data: Record<string, unknown>) {
  if (!settingsProvider.value) return { success: false, message: 'No provider selected' }
  store.updateProvider(settingsProvider.value.id, {
    apiBaseUrl: data.apiBaseUrl as string,
    apiKey: data.apiKey as string,
  })
  return store.testProviderConnection(settingsProvider.value.id)
}

function handleDeleteProvider() {
  if (settingsProvider.value) {
    store.updateProvider(settingsProvider.value.id, {
      apiKey: '',
      status: 'disconnected',
    })
  }
  showSettings.value = false
}

function handleSend(content: string, attachments: Attachment[]) {
  store.sendMessage(content, attachments)
}
</script>

<style scoped>
.ai-chat-page {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  overflow: hidden;
  background: var(--bg-primary);
}

.ai-chat-layout {
  display: flex;
  flex: 1;
  min-height: 0;
}

.ai-chat-provider {
  width: 240px;
  flex-shrink: 0;
}

.ai-chat-main {
  flex: 1;
  min-width: 0;
}

.ai-chat-context {
  width: 260px;
  flex-shrink: 0;
}

@media (max-width: 1200px) {
  .ai-chat-context {
    display: none;
  }
}

@media (max-width: 900px) {
  .ai-chat-provider {
    display: none;
  }
}
</style>
