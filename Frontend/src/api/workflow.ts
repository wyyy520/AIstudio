import http from './request'

export interface Workflow {
  id: number
  projectId: number
  name: string
  definition: string
  status: string
  createdAt: string
  updatedAt: string
}

export async function getWorkflows(projectId?: string): Promise<Workflow[]> {
  const params = projectId ? { projectId } : {}
  const res = await http.get('/api/workflows', { params })
  return (res as unknown as { data: Workflow[] }).data
}

export async function getWorkflowById(id: number): Promise<Workflow> {
  const res = await http.get(`/api/workflows/${id}`)
  return (res as unknown as { data: Workflow }).data
}

export async function createWorkflow(data: {
  projectId: number
  name: string
  definition?: string
}): Promise<Workflow> {
  const res = await http.post('/api/workflows', data)
  return (res as unknown as { data: Workflow }).data
}

export async function updateWorkflow(id: number, data: Partial<{ name: string; definition: string; status: string }>): Promise<Workflow> {
  const res = await http.put(`/api/workflows/${id}`, data)
  return (res as unknown as { data: Workflow }).data
}

export async function deleteWorkflow(id: number): Promise<void> {
  await http.delete(`/api/workflows/${id}`)
}