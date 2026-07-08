import http from './request'

export interface ApiWorkflow {
  id: number
  projectId: number
  name: string
  definition: string
  status: string
  createdAt: string
  updatedAt: string
}

export interface ApiWorkflowCreateRequest {
  projectId: number
  name: string
  definition?: string
}

export interface ApiWorkflowUpdateRequest {
  name?: string
  definition?: string
  status?: string
}

export interface ApiNodeType {
  type: string
  plugin: string
  name: string
  description: string
  inputs: Array<{ name: string; type: string; required: boolean }>
  outputs: Array<{ name: string; type: string }>
}

export interface ApiWorkflowRunResult {
  task_id: string
}

// GET /api/workflows
export async function getWorkflows(projectId?: string): Promise<ApiWorkflow[]> {
  const params = projectId ? { projectId } : {}
  const res = await http.get('/api/workflows', { params })
  return (res as unknown as { data: ApiWorkflow[] }).data ?? []
}

// GET /api/workflows/:id
export async function getWorkflowById(id: number | string): Promise<ApiWorkflow> {
  const res = await http.get(`/api/workflows/${id}`)
  return (res as unknown as { data: ApiWorkflow }).data
}

// POST /api/workflows
export async function createWorkflow(data: ApiWorkflowCreateRequest): Promise<ApiWorkflow> {
  const res = await http.post('/api/workflows', data)
  return (res as unknown as { data: ApiWorkflow }).data
}

// PUT /api/workflows/:id
export async function updateWorkflow(id: number | string, data: ApiWorkflowUpdateRequest): Promise<ApiWorkflow> {
  const res = await http.put(`/api/workflows/${id}`, data)
  return (res as unknown as { data: ApiWorkflow }).data
}

// DELETE /api/workflows/:id
export async function deleteWorkflow(id: number | string): Promise<void> {
  await http.delete(`/api/workflows/${id}`)
}

// POST /api/workflows/:id/run — executes a workflow and returns a task_id
export async function runWorkflow(id: number | string): Promise<ApiWorkflowRunResult> {
  const res = await http.post(`/api/workflows/${id}/run`)
  return (res as unknown as { data: ApiWorkflowRunResult }).data
}

// GET /api/workflows/nodes — returns all registered node types
export async function getNodeTypes(): Promise<ApiNodeType[]> {
  const res = await http.get('/api/workflows/nodes')
  return (res as unknown as { data: ApiNodeType[] }).data ?? []
}