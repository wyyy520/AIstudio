import http from './request'

export interface ApiTask {
  id: string
  name: string
  description: string
  priority: number
  status: string
  handler: string
  payload?: unknown
  result?: unknown
  error?: string
  createdAt: string
  updatedAt: string
  startedAt?: string
  completedAt?: string
}

export interface ApiTaskStatus {
  id: string
  status: string
  progress: number
  result?: unknown
  error?: string
  startedAt?: string
  completedAt?: string
}

export interface ApiTaskCreateRequest {
  name: string
  description?: string
  handler: string
  priority?: number
  payload?: unknown
}

export interface ApiTaskCreateResponse {
  task_id: string
}

// GET /api/tasks — list all tasks
export async function getTasks(): Promise<ApiTask[]> {
  const res = await http.get('/api/tasks')
  return (res as unknown as { data: ApiTask[] }).data ?? []
}

// GET /api/tasks/:id — get task detail
export async function getTaskById(id: string): Promise<ApiTask> {
  const res = await http.get(`/api/tasks/${id}`)
  return (res as unknown as { data: ApiTask }).data
}

// GET /api/task/:id/status — get task status (real-time)
export async function getTaskStatus(id: string): Promise<ApiTaskStatus> {
  const res = await http.get(`/api/task/${id}/status`)
  return (res as unknown as { data: ApiTaskStatus }).data
}

// POST /api/task/create — create a new task
export async function createTask(data: ApiTaskCreateRequest): Promise<ApiTaskCreateResponse> {
  const res = await http.post('/api/task/create', data)
  return (res as unknown as { data: ApiTaskCreateResponse }).data
}

// POST /api/tasks — create a task via the tasks group
export async function createTaskAlt(data: ApiTaskCreateRequest): Promise<ApiTask> {
  const res = await http.post('/api/tasks', data)
  return (res as unknown as { data: ApiTask }).data
}

// PUT /api/tasks/:id/cancel — cancel a task
export async function cancelTask(id: string): Promise<void> {
  await http.put(`/api/tasks/${id}/cancel`)
}

// DELETE /api/tasks/:id — delete a task
export async function deleteTask(id: string): Promise<void> {
  await http.delete(`/api/tasks/${id}`)
}