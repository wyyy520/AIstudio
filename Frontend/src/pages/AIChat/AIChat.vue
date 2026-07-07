<template>
  <div class="ai-chat-page">
    <ChatToolbar
      :providers="providers"
      :selected-model="selectedModel"
      @model-change="selectedModel = $event"
      @new-chat="handleNewChat"
      @toggle-history="showHistory = true"
    />

    <div class="ai-chat-layout">
      <ProviderPanel
        class="ai-chat-provider"
        :providers="providers"
        :active-provider-id="activeProviderId"
        @select-provider="activeProviderId = $event"
        @settings="openProviderSettings"
        @toggle-history="showHistory = true"
      />

      <ChatArea
        class="ai-chat-main"
        :messages="messages"
        :agent-status="agentStatus"
        :quick-actions="quickActions"
        @send="handleSend"
        @stop="handleStop"
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
      @close="showSettings = false"
      @save="handleSaveProvider"
      @delete="handleDeleteProvider"
    />

    <HistoryDrawer
      :conversations="conversations"
      :visible="showHistory"
      :selected-id="currentConversationId"
      @close="showHistory = false"
      @select="handleSelectConversation"
      @delete="handleDeleteConversation"
      @favorite="handleFavoriteConversation"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import ChatToolbar from './components/toolbar/ChatToolbar.vue'
import ProviderPanel from './components/provider/ProviderPanel.vue'
import ProviderSettings from './components/provider/ProviderSettings.vue'
import ChatArea from './components/chat/ChatArea.vue'
import ContextPanel from './components/context/ContextPanel.vue'
import HistoryDrawer from './components/context/HistoryDrawer.vue'
import {
  mockProviders,
  mockMessages,
  mockConversations,
  mockContext,
  mockQuickActions,
} from './mock'
import type {
  AIProvider,
  ChatMessage,
  Conversation,
  AgentStatus,
  QuickAction,
  Attachment,
} from './types'

const providers = ref<AIProvider[]>(mockProviders)
const messages = ref<ChatMessage[]>(mockMessages)
const conversations = ref<Conversation[]>(mockConversations)
const context = reactive(mockContext)
const quickActions = mockQuickActions

const selectedModel = ref('claude-sonnet-4')
const activeProviderId = ref('claude')
const currentConversationId = ref('conv-1')
const agentStatus = ref<AgentStatus>('idle')

const showSettings = ref(false)
const showHistory = ref(false)
const settingsProvider = ref<AIProvider | null>(null)

function openProviderSettings(provider: AIProvider) {
  settingsProvider.value = provider
  showSettings.value = true
}

function handleSaveProvider(data: Record<string, unknown>) {
  if (settingsProvider.value) {
    const idx = providers.value.findIndex(p => p.id === settingsProvider.value!.id)
    if (idx !== -1) {
      providers.value[idx].apiBaseUrl = data.apiBaseUrl as string
      providers.value[idx].apiKey = data.apiKey as string
    }
  }
  showSettings.value = false
}

function handleDeleteProvider() {
  showSettings.value = false
}

function handleNewChat() {
  currentConversationId.value = `conv-${Date.now()}`
  messages.value = []
  agentStatus.value = 'idle'
}

function handleSend(content: string, _attachments: Attachment[]) {
  if (!content.trim()) return

  const userMsg: ChatMessage = {
    id: `msg-${Date.now()}`,
    conversationId: currentConversationId.value,
    role: 'user',
    content,
    createdAt: Date.now(),
  }
  messages.value.push(userMsg)

  agentStatus.value = 'thinking'

  setTimeout(() => {
    agentStatus.value = 'planning'
  }, 800)

  setTimeout(() => {
    agentStatus.value = 'calling_tool'
  }, 1500)

  setTimeout(() => {
    agentStatus.value = 'running'
  }, 2200)

  setTimeout(() => {
    agentStatus.value = 'finished'
    const aiMsg: ChatMessage = {
      id: `msg-${Date.now()}`,
      conversationId: currentConversationId.value,
      role: 'assistant',
      content: generateMockResponse(content),
      model: providers.value.find(p => p.models.some(m => m.id === selectedModel.value))?.name || 'AI',
      createdAt: Date.now(),
      duration: Math.floor(Math.random() * 3000) + 1000,
      tokenUsage: {
        prompt: Math.floor(Math.random() * 500) + 200,
        completion: Math.floor(Math.random() * 800) + 300,
        total: 0,
      },
    }
    aiMsg.tokenUsage!.total = aiMsg.tokenUsage!.prompt + aiMsg.tokenUsage!.completion
    messages.value.push(aiMsg)

    setTimeout(() => {
      agentStatus.value = 'idle'
    }, 500)
  }, 3000)
}

function generateMockResponse(input: string): string {
  if (input.includes('Workflow') || input.includes('workflow')) {
    return `好的，我来为你生成一个 Workflow。\n\n## 分析需求\n\n根据你的描述，我建议创建以下 Workflow 节点：\n\n1. **数据输入节点** - 接收数据源\n2. **预处理节点** - 数据清洗和转换\n3. **处理节点** - 核心逻辑处理\n4. **输出节点** - 结果输出\n\n我已经在工作流编辑器中创建了这个 Workflow，你可以切换到工作流页面查看和编辑。`
  }
  if (input.includes('训练') || input.includes('模型')) {
    return `收到，我来帮你训练模型。\n\n## 环境检查\n\n- ✅ Python 3.10\n- ✅ PyTorch 2.0\n- ✅ CUDA 11.8\n\n## 训练配置\n\n\`\`\`yaml\nepochs: 100\nbatch_size: 16\nlearning_rate: 0.01\n\`\`\`\n\n训练已准备就绪，正在启动...`
  }
  if (input.includes('错误') || input.includes('修复') || input.includes('fix')) {
    return `我来分析这个错误：\n\n## 错误分析\n\n错误发生在 \`train.py:42\` 行：\n\n\`\`\`python\nTypeError: unsupported operand type(s) for +: 'int' and 'NoneType'\n\`\`\`\n\n## 原因\n\n变量 \`learning_rate\` 未被正确初始化。\n\n## 修复方案\n\n\`\`\`python\nlearning_rate = 0.01  # 添加默认值\n\`\`\`\n\n已自动修复该错误。`
  }
  return `收到你的消息。我正在分析中...\n\n这是一个 Mock 响应。在实际使用中，AI 将根据你的输入内容提供智能回复，包括代码生成、任务执行、问题分析等功能。`
}

function handleStop() {
  agentStatus.value = 'idle'
}

function handleSelectConversation(id: string) {
  currentConversationId.value = id
  showHistory.value = false
}

function handleDeleteConversation(id: string) {
  conversations.value = conversations.value.filter(c => c.id !== id)
}

function handleFavoriteConversation(id: string) {
  const conv = conversations.value.find(c => c.id === id)
  if (conv) {
    conv.isFavorite = !conv.isFavorite
  }
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

/* Responsive */
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
