import request from './request'

export interface ApiProject {
  id: number
  name: string
  description: string
  status: string
  ownerId: number
  createdAt: string
  updatedAt: string
}

export interface Project {
  id: string
  name: string
  description: string
  status: string
  created_at: string
  updated_at: string
}

export function getProjects() {
  return request.get('/projects')
}

export function getProject(id: string) {
  return request.get(`/projects/${id}`)
}

export const getProjectById = getProject

export function createProject(data: { name: string; target?: string; description?: string }) {
  return request.post('/projects', data)
}

export function updateProject(id: string, data: { name?: string; description?: string; status?: string }) {
  return request.put(`/projects/${id}`, data)
}

export function deleteProject(id: string) {
  return request.delete(`/projects/${id}`)
}