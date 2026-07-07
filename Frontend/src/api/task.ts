import http from './request'

export interface Task {
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

export async function getTasks(): Promise<Task[]> {
  const res = await http.get('/api/tasks')
  return (res as unknown as { data: Task[] }).data
}

export async function getTaskById(id: string): Promise<Task> {
  const res = await http.get(`/api/tasks/${id}`)
  return (res as unknown as { data: Task }).data
}

export async function createTask(data: {
  name: string
  description?: string
  handler: string
  priority?: number
  payload?: unknown
}): Promise<{ taskId: string }> {
  const res = await http.post('/api/tasks', data)
  return (res as unknown as { data: { taskId: string } }).data
}

export async function cancelTask(id: string): Promise<void> {
  await http.put(`/api/tasks/${id}/cancel`)
}