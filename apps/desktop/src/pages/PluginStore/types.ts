export type PluginCategory =
  | 'vision'
  | 'nlp'
  | 'timeseries'
  | 'speech'
  | 'simulation'
  | 'system'
  | 'mcp'

export type PluginStatus =
  | 'installed'
  | 'not-installed'
  | 'installing'
  | 'updating'
  | 'error'

export type DependencyStatus = 'satisfied' | 'not-installed' | 'version-mismatch' | 'checking'

export interface Dependency {
  name: string
  versionRequired: string
  versionInstalled?: string
  status: DependencyStatus
}

export interface WorkflowNodeDef {
  name: string
  type: string
  category: PluginCategory
}

export interface AgentToolDef {
  name: string
  description: string
  parameters: ToolParameter[]
  returns: string
}

export interface ToolParameter {
  name: string
  type: string
  required: boolean
  description?: string
}

export interface Plugin {
  id: string
  name: string
  version: string
  author: string
  source: 'github' | 'local' | 'registry'
  sourceUrl: string
  description: string
  category: PluginCategory
  icon: string
  status: PluginStatus
  capabilities: string[]
  workflowNodes: WorkflowNodeDef[]
  dependencies: Dependency[]
  agentTools: AgentToolDef[]
  tags: string[]
  installedAt?: string
  updatedAt?: string
  error?: string
  size?: string
  downloads?: number
  githubUrl?: string
  readme?: string
}

export type InstallStepStatus = 'completed' | 'in-progress' | 'pending' | 'failed'

export interface InstallStep {
  id: string
  name: string
  status: InstallStepStatus
  duration?: number
  logs: LogEntry[]
  error?: string
}

export interface LogEntry {
  timestamp: string
  level: 'info' | 'warn' | 'error' | 'debug'
  message: string
}

export type InstallTaskStatus = 'running' | 'completed' | 'failed' | 'cancelled'

export interface InstallTask {
  id: string
  pluginId: string
  pluginName: string
  status: InstallTaskStatus
  steps: InstallStep[]
  startedAt: string
  completedAt?: string
}

export interface PluginCategoryGroup {
  category: PluginCategory
  label: string
  icon: string
  color: string
}

export interface AgentInvocation {
  id: string
  pluginId: string
  userMessage: string
  agentResponse: string
  toolUsed?: string
  workflowGenerated?: string[]
  result?: string
  timestamp: string
}