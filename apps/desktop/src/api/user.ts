import http from './request'

export interface User {
  id: number
  username: string
  email: string
  createdAt: string
  updatedAt: string
}

export async function getUsers(): Promise<User[]> {
  const res = await http.get('/users')
  return (res as unknown as { data: User[] }).data
}

export async function getUserById(id: number): Promise<User> {
  const res = await http.get(`/users/${id}`)
  return (res as unknown as { data: User }).data
}

export async function createUser(data: { username: string; email: string; password: string }): Promise<User> {
  const res = await http.post('/users', data)
  return (res as unknown as { data: User }).data
}

export async function updateUser(id: number, data: Partial<{ username: string; email: string; password: string }>): Promise<User> {
  const res = await http.put(`/users/${id}`, data)
  return (res as unknown as { data: User }).data
}

export async function deleteUser(id: number): Promise<void> {
  await http.delete(`/users/${id}`)
}