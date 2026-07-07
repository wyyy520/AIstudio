import http from './request'

export interface ApiProject {
  id: number
  name: string
  description: string
  ownerId: number
  status: string
  createdAt: string
  updatedAt: string
}

export async function getProjects(): Promise<ApiProject[]> {
  const res = await http.get('/api/projects')
  return (res as unknown as { data: ApiProject[] }).data
}

export async function getProjectById(id: number): Promise<ApiProject> {
  const res = await http.get(`/api/projects/${id}`)
  return (res as unknown as { data: ApiProject }).data
}

export async function createProject(data: {
  name: string
  description?: string
  ownerId: number
}): Promise<ApiProject> {
  const res = await http.post('/api/projects', data)
  return (res as unknown as { data: ApiProject }).data
}

export async function updateProject(id: number, data: Partial<{ name: string; description: string; status: string }>): Promise<ApiProject> {
  const res = await http.put(`/api/projects/${id}`, data)
  return (res as unknown as { data: ApiProject }).data
}

export async function deleteProject(id: number): Promise<void> {
  await http.delete(`/api/projects/${id}`)
}

export async function getTemplates(): Promise<ApiProject[]> {
  const res = await http.get('/api/projects/templates')
  return (res as unknown as { data: ApiProject[] }).data
}

export async function runWorkflow(projectId: number, workflowId: string): Promise<unknown> {
  const res = await http.post(`/api/projects/${projectId}/workflows/${workflowId}/run`)
  return (res as unknown as { data: unknown }).data
}

export async function repairEnvironment(projectId: number): Promise<unknown> {
  const res = await http.post(`/api/projects/${projectId}/environment/repair`)
  return (res as unknown as { data: unknown }).data
}