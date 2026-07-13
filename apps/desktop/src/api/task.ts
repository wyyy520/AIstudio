import request from './request'

export interface ApiTask {
  id: string
  name: string
  description?: string
  priority?: number
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
  progress?: number
  result?: unknown
  error?: string
  startedAt?: string
  completedAt?: string
}

export interface ApiTaskCreateRequest {
  name: string
  description?: string
  priority?: number
  handler: string
  payload?: unknown
}

export interface Task {
  id: string
  name: string
  type: string
  status: string
  progress: number
  created_at: string
  updated_at: string
}

export function getTasks() {
  return request.get('/tasks')
}

export const getTaskById = getTask

export function getTask(id: string) {
  return request.get(`/tasks/${id}`)
}

export function createTask(data: ApiTaskCreateRequest | Partial<Task>) {
  return request.post('/tasks', data)
}

export function cancelTask(id: string) {
  return request.put(`/tasks/${id}/cancel`)
}

export function deleteTask(id: string) {
  return request.delete(`/tasks/${id}`)
}

export function getTaskStatus(id: string) {
  return request.get(`/tasks/${id}/status`)
}