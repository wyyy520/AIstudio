import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { sendChatStream } from '@/api/agent'
import { testEngineConnection } from '@/api/settings'
import type {
  AIProvider,
  AIModel,
  ChatMessage,
  Conversation,
  AgentStatus,
  Attachment,
  ToolCall,
  PlanItem,
  ProviderStatus,
  AgentResponse,
} from '@/pages/AIChat/types'

const DEFAULT_PROVIDERS: AIProvider[] = [
  {
    id: 'openai',
    name: 'OpenAI',
    icon: 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 15h-2v-2h2v2zm0-4h-2V7h2v6z',
    apiBaseUrl: 'https://api.openai.com/v1',
    apiKey: '',
    status: 'disconnected',
    models: [
      { id: 'gpt-5', name: 'GPT-5', providerId: 'openai', maxTokens: 256000, defaultTemperature: 0.7 },
      { id: 'gpt-4.1', name: 'GPT-4.1', providerId: 'openai', maxTokens: 128000, defaultTemperature: 0.7 },
      { id: 'gpt-4o', name: 'GPT-4o', providerId: 'openai', maxTokens: 128000, defaultTemperature: 0.7 },
      { id: 'gpt-4o-mini', name: 'GPT-4o Mini', providerId: 'openai', maxTokens: 128000, defaultTemperature: 0.7 },
      { id: 'gpt-4.1-mini', name: 'GPT-4.1 Mini', providerId: 'openai', maxTokens: 128000, defaultTemperature: 0.7 },
      { id: 'gpt-4.1-nano', name: 'GPT-4.1 Nano', providerId: 'openai', maxTokens: 128000, defaultTemperature: 0.7 },
      { id: 'o3', name: 'o3', providerId: 'openai', maxTokens: 200000, defaultTemperature: 0.7 },
      { id: 'o4-mini', name: 'o4-mini', providerId: 'openai', maxTokens: 200000, defaultTemperature: 0.7 },
    ],
  },
  {
    id: 'claude',
    name: 'Claude (Anthropic)',
    icon: 'M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5',
    apiBaseUrl: 'https://api.anthropic.com/v1',
    apiKey: '',
    status: 'disconnected',
    models: [
      { id: 'claude-sonnet-5', name: 'Claude Sonnet 5', providerId: 'claude', maxTokens: 256000, defaultTemperature: 0.7 },
      { id: 'claude-opus-4.8', name: 'Claude Opus 4.8', providerId: 'claude', maxTokens: 200000, defaultTemperature: 0.7 },
      { id: 'claude-haiku-4.5', name: 'Claude Haiku 4.5', providerId: 'claude', maxTokens: 200000, defaultTemperature: 0.7 },
      { id: 'claude-sonnet-4', name: 'Claude Sonnet 4', providerId: 'claude', maxTokens: 200000, defaultTemperature: 0.7 },
      { id: 'claude-opus-4', name: 'Claude Opus 4', providerId: 'claude', maxTokens: 200000, defaultTemperature: 0.7 },
    ],
  },
  {
    id: 'deepseek',
    name: 'DeepSeek',
    icon: 'M13 2L3 14h9l-1 8 10-12h-9l1-8z',
    apiBaseUrl: 'https://api.deepseek.com/v1',
    apiKey: '',
    status: 'disconnected',
    models: [
      { id: 'deepseek-chat-v4', name: 'DeepSeek V4', providerId: 'deepseek', maxTokens: 128000, defaultTemperature: 0.7 },
      { id: 'deepseek-chat', name: 'DeepSeek V3', providerId: 'deepseek', maxTokens: 64000, defaultTemperature: 0.7 },
      { id: 'deepseek-reasoner', name: 'DeepSeek R1', providerId: 'deepseek', maxTokens: 128000, defaultTemperature: 0.7 },
    ],
  },
  {
    id: 'gemini',
    name: 'Google Gemini',
    icon: 'M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20zm0 18a8 8 0 1 1 0-16 8 8 0 0 1 0 16z',
    apiBaseUrl: 'https://generativelanguage.googleapis.com/v1',
    apiKey: '',
    status: 'disconnected',
    models: [
      { id: 'gemini-2.5-pro', name: 'Gemini 2.5 Pro', providerId: 'gemini', maxTokens: 1000000, defaultTemperature: 0.7 },
      { id: 'gemini-2.5-flash', name: 'Gemini 2.5 Flash', providerId: 'gemini', maxTokens: 1000000, defaultTemperature: 0.7 },
      { id: 'gemini-2.0-flash', name: 'Gemini 2.0 Flash', providerId: 'gemini', maxTokens: 1000000, defaultTemperature: 0.7 },
    ],
  },
  {
    id: 'qwen',
    name: '通义千问 (Qwen)',
    icon: 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2z',
    apiBaseUrl: 'https://dashscope.aliyuncs.com/api/v1',
    apiKey: '',
    status: 'disconnected',
    models: [
      { id: 'qwen-max-2026', name: 'Qwen Max (最新)', providerId: 'qwen', maxTokens: 128000, defaultTemperature: 0.7 },
      { id: 'qwen-max', name: 'Qwen Max', providerId: 'qwen', maxTokens: 32000, defaultTemperature: 0.7 },
      { id: 'qwen-plus', name: 'Qwen Plus', providerId: 'qwen', maxTokens: 128000, defaultTemperature: 0.7 },
      { id: 'qwen-turbo', name: 'Qwen Turbo', providerId: 'qwen', maxTokens: 128000, defaultTemperature: 0.7 },
    ],
  },
  {
    id: 'ollama',
    name: 'Ollama (本地)',
    icon: 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8z',
    apiBaseUrl: import.meta.env.VITE_OLLAMA_URL || 'http://localhost:11434',
    apiKey: '',
    status: 'disconnected',
    models: [
      { id: 'llama-4', name: 'Llama 4', providerId: 'ollama', maxTokens: 131072, defaultTemperature: 0.7 },
      { id: 'llama3.1', name: 'Llama 3.1', providerId: 'ollama', maxTokens: 8192, defaultTemperature: 0.7 },
      { id: 'qwen2.5', name: 'Qwen 2.5', providerId: 'ollama', maxTokens: 32768, defaultTemperature: 0.7 },
      { id: 'deepseek-r1', name: 'DeepSeek R1 (本地)', providerId: 'ollama', maxTokens: 32768, defaultTemperature: 0.7 },
      { id: 'mistral', name: 'Mistral', providerId: 'ollama', maxTokens: 8192, defaultTemperature: 0.7 },
    ],
  },
]

function loadProviders(): AIProvider[] {
  try {
    const saved = localStorage.getItem('aistudio_chat_providers')
    if (saved) return JSON.parse(saved)
  } catch { }
  return DEFAULT_PROVIDERS.map(p => ({ ...p }))
}

function saveProviders(providers: AIProvider[]) {
  localStorage.setItem('aistudio_chat_providers', JSON.stringify(providers))
}

function loadConversations(): Conversation[] {
  try {
    const saved = localStorage.getItem('aistudio_chat_conversations')
    if (saved) return JSON.parse(saved)
  } catch { }
  return []
}

function saveConversations(conversations: Conversation[]) {
  localStorage.setItem('aistudio_chat_conversations', JSON.stringify(conversations))
}

function loadMessages(convId: string): ChatMessage[] {
  try {
    const saved = localStorage.getItem(`aistudio_chat_messages_${convId}`)
    if (saved) return JSON.parse(saved)
  } catch { }
  return []
}

function saveMessages(convId: string, messages: ChatMessage[]) {
  localStorage.setItem(`aistudio_chat_messages_${convId}`, JSON.stringify(messages))
}

function genId(): string {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
}

export const useChatStore = defineStore('chat', () => {
  const providers = ref<AIProvider[]>(loadProviders())
  const conversations = ref<Conversation[]>(loadConversations())
  const messages = ref<ChatMessage[]>([])
  const currentConversationId = ref<string>('')
  const currentProviderId = ref<string>('claude')
  const currentModelId = ref<string>('claude-sonnet-5')

  const streamingContent = ref('')
  const streamingToolCalls = ref<ToolCall[]>([])
  const streamingPlan = ref<PlanItem[]>([])
  const agentStatus = ref<AgentStatus>('idle')
  const isStreaming = ref(false)
  const abortController = ref<AbortController | null>(null)

  const currentProvider = computed(() =>
    providers.value.find(p => p.id === currentProviderId.value) || providers.value[0]
  )

  const currentModel = computed(() => {
    const p = currentProvider.value
    return p?.models.find(m => m.id === currentModelId.value)
  })

  const currentConversation = computed(() =>
    conversations.value.find(c => c.id === currentConversationId.value)
  )

  function loadConversationMessages(convId: string) {
    messages.value = loadMessages(convId)
  }

  function saveConversationMessages() {
    if (currentConversationId.value) {
      saveMessages(currentConversationId.value, messages.value)
    }
  }

  function newConversation() {
    const id = genId()
    const conv: Conversation = {
      id,
      title: '新对话',
      createdAt: Date.now(),
      updatedAt: Date.now(),
      messageCount: 0,
      model: currentModel.value?.name || currentModelId.value,
      isFavorite: false,
    }
    conversations.value.unshift(conv)
    currentConversationId.value = id
    messages.value = []
    saveConversations(conversations.value)
    streamingContent.value = ''
    agentStatus.value = 'idle'
    isStreaming.value = false
    streamingToolCalls.value = []
    streamingPlan.value = []
  }

  function deleteConversation(id: string) {
    conversations.value = conversations.value.filter(c => c.id !== id)
    saveConversations(conversations.value)
    localStorage.removeItem(`aistudio_chat_messages_${id}`)
    if (currentConversationId.value === id && conversations.value.length > 0) {
      selectConversation(conversations.value[0].id)
    } else if (conversations.value.length === 0) {
      newConversation()
    }
  }

  function toggleFavorite(id: string) {
    const conv = conversations.value.find(c => c.id === id)
    if (conv) {
      conv.isFavorite = !conv.isFavorite
      saveConversations(conversations.value)
    }
  }

  function selectConversation(id: string) {
    if (isStreaming.value) stopStreaming()
    currentConversationId.value = id
    loadConversationMessages(id)
    streamingContent.value = ''
    streamingToolCalls.value = []
    streamingPlan.value = []
    agentStatus.value = 'idle'
  }

  function selectProvider(providerId: string) {
    currentProviderId.value = providerId
    const provider = providers.value.find(p => p.id === providerId)
    if (provider?.models.length) {
      currentModelId.value = provider.models[0].id
    }
  }

  function selectModel(modelId: string) {
    currentModelId.value = modelId
  }

  function updateProvider(id: string, data: Partial<AIProvider>) {
    const idx = providers.value.findIndex(p => p.id === id)
    if (idx !== -1) {
      Object.assign(providers.value[idx], data)
      saveProviders(providers.value)
    }
  }

  function setProviderStatus(id: string, status: ProviderStatus, error?: string) {
    const provider = providers.value.find(p => p.id === id)
    if (provider) {
      provider.status = status
      provider.connectionError = error
      saveProviders(providers.value)
    }
  }

  async function testProviderConnection(id: string): Promise<{ success: boolean; message: string }> {
    const provider = providers.value.find(p => p.id === id)
    if (!provider) return { success: false, message: 'Provider not found' }

    setProviderStatus(id, 'disconnected')
    try {
      const result = await testEngineConnection({
        provider: provider.id,
        endpoint: provider.apiBaseUrl,
        apiKey: provider.apiKey,
      })
      if (result.success) {
        setProviderStatus(id, 'connected')
      } else {
        setProviderStatus(id, 'error', result.message)
      }
      return result
    } catch (e: any) {
      const msg = e.message || 'Connection failed'
      setProviderStatus(id, 'error', msg)
      return { success: false, message: msg }
    }
  }

  function updateConversationTitle(title: string) {
    const conv = conversations.value.find(c => c.id === currentConversationId.value)
    if (conv) {
      conv.title = title
      conv.updatedAt = Date.now()
      saveConversations(conversations.value)
    }
  }

  async function sendMessage(content: string, attachments?: Attachment[]) {
    if (!content.trim() || isStreaming.value) return

    if (!currentConversationId.value) {
      newConversation()
    }

    const userMsg: ChatMessage = {
      id: genId(),
      conversationId: currentConversationId.value,
      role: 'user',
      content,
      createdAt: Date.now(),
      attachments: attachments?.length ? attachments : undefined,
    }
    messages.value.push(userMsg)
    updateConversationTitle(content.slice(0, 50))

    streamingContent.value = ''
    streamingToolCalls.value = []
    streamingPlan.value = []

    const assistantMsgId = genId()
    const assistantMsg: ChatMessage = {
      id: assistantMsgId,
      conversationId: currentConversationId.value,
      role: 'assistant',
      content: '',
      model: currentModel.value?.name || currentModelId.value,
      createdAt: Date.now(),
      isStreaming: true,
    }
    messages.value.push(assistantMsg)
    isStreaming.value = true

    const controller = new AbortController()
    abortController.value = controller

    const conv = conversations.value.find(c => c.id === currentConversationId.value)
    if (conv) {
      conv.messageCount++
      conv.updatedAt = Date.now()
      saveConversations(conversations.value)
    }

    await sendChatStream(
      {
        message: content,
        projectId: conv?.id || '',
      },
      {
        signal: controller.signal,
        onChunk: (text) => {
          streamingContent.value = text
          const idx = messages.value.findIndex(m => m.id === assistantMsgId)
          if (idx !== -1) {
            messages.value[idx] = { ...messages.value[idx], content: text }
          }
        },
        onStatus: (status) => {
          agentStatus.value = status as AgentStatus
          const idx = messages.value.findIndex(m => m.id === assistantMsgId)
          if (idx !== -1) {
            messages.value[idx] = { ...messages.value[idx], agentStatus: status as AgentStatus }
          }
        },
        onToolCall: (call) => {
          const existingIdx = streamingToolCalls.value.findIndex(t => t.id === call.id)
          if (existingIdx !== -1) {
            streamingToolCalls.value[existingIdx] = call
          } else {
            streamingToolCalls.value.push(call)
          }
          const idx = messages.value.findIndex(m => m.id === assistantMsgId)
          if (idx !== -1) {
            messages.value[idx] = {
              ...messages.value[idx],
              toolCalls: [...streamingToolCalls.value],
            }
          }
        },
        onPlan: (plan) => {
          streamingPlan.value = plan
          const idx = messages.value.findIndex(m => m.id === assistantMsgId)
          if (idx !== -1) {
            messages.value[idx] = { ...messages.value[idx], plan }
          }
        },
        onDone: (_resp: AgentResponse) => {
          const idx = messages.value.findIndex(m => m.id === assistantMsgId)
          if (idx !== -1) {
            const finalContent = streamingContent.value || _resp.reply || _resp.summary || ''
            messages.value[idx] = {
              ...messages.value[idx],
              content: finalContent,
              isStreaming: false,
              agentStatus: 'finished',
              toolCalls: streamingToolCalls.value.length > 0 ? [...streamingToolCalls.value] : undefined,
              plan: streamingPlan.value.length > 0 ? [...streamingPlan.value] : undefined,
            }
          }
          saveConversationMessages()
          finish()
        },
        onError: (error) => {
          const idx = messages.value.findIndex(m => m.id === assistantMsgId)
          if (idx !== -1) {
            messages.value[idx] = {
              ...messages.value[idx],
              content: error.message || '请求失败',
              isStreaming: false,
              agentStatus: 'error',
            }
          }
          saveConversationMessages()
          finish()
        },
      }
    )
  }

  function stopStreaming() {
    if (abortController.value) {
      abortController.value.abort()
      abortController.value = null
    }
    const idx = messages.value.findIndex(m => m.isStreaming)
    if (idx !== -1) {
      messages.value[idx] = {
        ...messages.value[idx],
        content: messages.value[idx].content || '(已停止)',
        isStreaming: false,
        agentStatus: 'idle',
      }
    }
    finish()
  }

  function finish() {
    isStreaming.value = false
    agentStatus.value = 'idle'
    abortController.value = null
    streamingContent.value = ''
    streamingToolCalls.value = []
    streamingPlan.value = []
  }

  return {
    providers,
    conversations,
    messages,
    currentConversationId,
    currentProviderId,
    currentModelId,
    streamingContent,
    streamingToolCalls,
    streamingPlan,
    agentStatus,
    isStreaming,
    currentProvider,
    currentModel,
    currentConversation,
    newConversation,
    deleteConversation,
    toggleFavorite,
    selectConversation,
    selectProvider,
    selectModel,
    updateProvider,
    setProviderStatus,
    testProviderConnection,
    sendMessage,
    stopStreaming,
    updateConversationTitle,
    loadConversationMessages,
  }
})