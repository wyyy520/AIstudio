import request from './request'

export interface ApiNodeType {
  type: string
  name: string
  category: string
  description: string
  inputs: Array<{ name: string; type: string; required: boolean }>
  outputs: Array<{ name: string; type: string }>
}

export interface ApiWorkflow {
  id: number
  name: string
  projectId: number
  definition: string
  status: string
  createdAt: string
  updatedAt: string
}

export interface ApiWorkflowRunResult {
  id: string
  workflowId: string
  status: string
  message?: string
}

export interface Workflow {
  id: string
  name: string
  description: string
  nodes: any[]
  edges: any[]
  status: string
  created_at: string
  updated_at: string
}

/** Compilation plan (dry-run) */
export interface CompilePlanData {
  generatorId: string
  generatorName: string
  projectName: string
  outputDir: string
  estimatedFiles: number
  estimatedSizeKB: number
  validated: boolean
  warnings: string[]
}

/** Generated file info */
export interface GeneratedFileInfo {
  path: string
  content: string
  mode: number
}

/** Compilation result */
export interface CompileResultData {
  target: string
  projectRoot: string
  entryPoints: string[]
  files: GeneratedFileInfo[]
  duration: string
  generatorId: string
}

/** Run result */
export interface RunResultData {
  runId: string
  status: string
  exitCode: number
  stdout: string
  stderr: string
  duration: string
  projectRoot: string
  entryPoints: string[]
  startedAt: string
  completedAt: string
}

/** Runtime environment report */
export interface EnvironmentReport {
  python: { detected: boolean; version: string }
  packages: Array<{ name: string; installed: boolean }>
  gpu: { available: boolean; cudaVersion: string }
  compatible: boolean
}

export function getWorkflows(projectId?: string) {
  const params = projectId ? { projectId } : {}
  return request.get('/workflows', { params })
}

export function getWorkflow(projectId: string) {
  return request.get(`/projects/${projectId}/workflow`)
}

export const getWorkflowById = getWorkflow

export function createWorkflow(data: { projectId?: number; name: string; definition?: string } | Partial<Workflow>) {
  return request.post('/workflows', data)
}

export function updateWorkflow(projectId: string, data: { definition?: string } | any) {
  return request.put(`/projects/${projectId}/workflow`, data)
}

export function deleteWorkflow(id: string) {
  return request.delete(`/workflows/${id}`)
}

export function runWorkflow(projectId: string, params?: any) {
  return request.post(`/projects/${projectId}/run`, params)
}

export const getNodeTypes = () => listNodeTypes()

export function listNodeTypes() {
  return request.get('/workflows/nodes')
}

/** Compile a project's workflow into generated project files */
export function compileProject(projectId: string, target?: string) {
  return request.post(`/projects/${projectId}/compile`, { target })
}

/** Execute a compiled project */
export function runCompiledProject(projectId: string, target?: string, timeout?: number) {
  return request.post(`/projects/${projectId}/run`, { target, timeout })
}

/** Detect runtime environment */
export function detectEnvironment(data?: { python?: string; packages?: string[]; gpu?: boolean }) {
  return request.post('/runtime/detect', data || {})
}

/** Stop a running project */
export function stopProject(data: { runId: string }) {
  return request.post('/runtime/stop', data)
}

/** Get runtime status */
export function getRuntimeStatus(runId: string) {
  return request.get(`/runtime/status/${runId}`)
}