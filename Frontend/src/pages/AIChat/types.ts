export interface AIModel {
  id: string
  name: string
  providerId: string
  maxTokens: number
  defaultTemperature: number
}

export interface AIProvider {
  id: string
  name: string
  icon: string
  apiBaseUrl: string
  apiKey: string
  models: AIModel[]
  status: 'connected' | 'disconnected' | 'error'
  connectionError?: string
}

export type AgentStatus = 'idle' | 'thinking' | 'planning' | 'calling_tool' | 'running' | 'finished' | 'error'

export interface TokenUsage {
  prompt: number
  completion: number
  total: number
}

export interface Attachment {
  id: string
  type: 'file' | 'image'
  name: string
  size: number
  url?: string
  mimeType: string
}

export interface TaskStep {
  id: string
  label: string
  status: 'pending' | 'running' | 'completed' | 'error'
}

export interface TaskExecution {
  id: string
  title: string
  steps: TaskStep[]
  progress: number
  status: AgentStatus
}

export interface ChatMessage {
  id: string
  conversationId: string
  role: 'user' | 'assistant' | 'system'
  content: string
  model?: string
  createdAt: number
  duration?: number
  tokenUsage?: TokenUsage
  attachments?: Attachment[]
  task?: TaskExecution
  agentStatus?: AgentStatus
}

export interface Conversation {
  id: string
  title: string
  createdAt: number
  updatedAt: number
  messageCount: number
  model: string
  isFavorite: boolean
}

export interface ChatContext {
  project?: { id: string; name: string }
  workflow?: { id: string; name: string; nodes: string[] }
  files?: Array<{ name: string; path: string }>
  plugins?: Array<{ name: string; status: 'active' | 'inactive' }>
  mcpServers?: Array<{ name: string; status: 'connected' | 'disconnected' }>
}

export interface QuickAction {
  id: string
  label: string
  icon: string
  prompt: string
}
