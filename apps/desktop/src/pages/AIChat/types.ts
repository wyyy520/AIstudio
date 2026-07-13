export interface AIModel {
  id: string
  name: string
  providerId: string
  maxTokens: number
  defaultTemperature: number
}

export type ProviderStatus = 'connected' | 'disconnected' | 'error'

export interface AIProvider {
  id: string
  name: string
  icon: string
  apiBaseUrl: string
  apiKey: string
  models: AIModel[]
  status: ProviderStatus
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

export interface ToolCall {
  id: string
  type: 'plugin' | 'workflow' | 'task' | 'mcp' | 'code'
  name: string
  description: string
  status: 'pending' | 'running' | 'completed' | 'error'
  input?: string
  output?: string
  duration?: number
}

export interface PlanItem {
  id: string
  action: string
  description: string
  status: 'pending' | 'running' | 'completed' | 'error'
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
  toolCalls?: ToolCall[]
  plan?: PlanItem[]
  isStreaming?: boolean
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

export interface AgentResponse {
  reply: string
  goal: string
  explanation: string
  plan: PlanItem[]
  steps: Record<string, any>[]
  status: string
  summary: string
}

export interface SendMessageOptions {
  signal?: AbortSignal
  onChunk?: (chunk: string) => void
  onStatus?: (status: AgentStatus) => void
  onToolCall?: (call: ToolCall) => void
  onPlan?: (plan: PlanItem[]) => void
  onDone?: (fullContent: string) => void
  onError?: (error: Error) => void
}
