export type TaskType = 'training' | 'simulation' | 'export' | 'workflow' | 'system' | 'agent'

export type TaskStatus = 'running' | 'success' | 'failed' | 'warning'

export type LogLevel = 'info' | 'warning' | 'error' | 'debug'

export type LogSource = 'system' | 'workflow' | 'plugin' | 'training' | 'agent'

export type StepStatus = 'completed' | 'running' | 'failed' | 'pending'

export type Severity = 'critical' | 'warning' | 'info'

export type AnalysisStatus = 'pending' | 'fixing' | 'fixed' | 'ignored'

export type RiskLevel = 'low' | 'medium' | 'high'

export type FixStepStatus = 'pending' | 'running' | 'completed' | 'failed'

export type AgentPhase = 'idle' | 'thinking' | 'analyzing' | 'calling_tool' | 'executing' | 'completed' | 'failed'

export type ImpactModule = 'cuda' | 'pytorch' | 'tensorflow' | 'dataset' | 'config' | 'memory' | 'network' | 'dependency' | 'unknown'

export interface Task {
  id: string
  name: string
  type: TaskType
  status: TaskStatus
  startedAt: string
  completedAt?: string
  duration: number
  projectId: string
  workflowId?: string
  metadata?: Record<string, unknown>
}

export interface LogEntry {
  id: string
  taskId: string
  timestamp: string
  level: LogLevel
  source: LogSource
  message: string
  rawMessage: string
  humanMessage: string
  stepName: string
  stepStatus: StepStatus
  metadata?: {
    file?: string
    line?: number
    function?: string
  }
}

export interface Solution {
  id: string
  title: string
  description: string
  command?: string
  estimatedTime: string
  risk: RiskLevel
  autoFixable: boolean
}

export interface ErrorAnalysis {
  id: string
  taskId: string
  logEntryIds: string[]
  severity: Severity
  errorType: string
  problem: string
  cause: string
  detail: string
  impactModule: ImpactModule
  solutions: Solution[]
  status: AnalysisStatus
  analyzedAt: string
}

export interface FixStep {
  id: string
  label: string
  status: FixStepStatus
}

export interface TrainingMetrics {
  taskId: string
  currentEpoch: number
  totalEpochs: number
  metrics: {
    loss: number
    accuracy: number
    learningRate: number
    gpuUsage: number
    memoryUsage: number
  }
  history: Array<{
    epoch: number
    loss: number
    accuracy: number
    gpuUsage: number
  }>
  updatedAt: string
}

export interface TimelineNode {
  nodeId: string
  name: string
  type: string
  status: StepStatus
  startedAt?: string
  completedAt?: string
  duration?: number
  progress?: number
}

export interface WorkflowTimeline {
  taskId: string
  workflowId: string
  nodes: TimelineNode[]
}

export type LogTab = 'human' | 'raw'

export type FilterLevel = 'all' | LogLevel