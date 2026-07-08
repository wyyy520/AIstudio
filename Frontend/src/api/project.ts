import http from './request'

// Backend API response types
export interface ApiProject {
  id: number
  name: string
  description: string
  ownerId: number
  status: string
  createdAt: string
  updatedAt: string
}

export interface ApiProjectCreateRequest {
  name: string
  description?: string
  ownerId: number
}

export interface ApiProjectUpdateRequest {
  name?: string
  description?: string
  status?: string
}

// GET /api/projects
export async function getProjects(): Promise<ApiProject[]> {
  const res = await http.get('/api/projects')
  return (res as unknown as { data: ApiProject[] }).data ?? []
}

// GET /api/projects/:id
export async function getProjectById(id: number | string): Promise<ApiProject> {
  const res = await http.get(`/api/projects/${id}`)
  return (res as unknown as { data: ApiProject }).data
}

// POST /api/projects
export async function createProject(data: ApiProjectCreateRequest): Promise<ApiProject> {
  const res = await http.post('/api/projects', data)
  return (res as unknown as { data: ApiProject }).data
}

// PUT /api/projects/:id
export async function updateProject(id: number | string, data: ApiProjectUpdateRequest): Promise<ApiProject> {
  const res = await http.put(`/api/projects/${id}`, data)
  return (res as unknown as { data: ApiProject }).data
}

// DELETE /api/projects/:id
export async function deleteProject(id: number | string): Promise<void> {
  await http.delete(`/api/projects/${id}`)
}