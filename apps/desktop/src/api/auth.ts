import request from './request'

export interface LoginParams {
  username: string
  password: string
}

export interface RegisterParams {
  username: string
  email: string
  password: string
}

export function login(data: LoginParams) {
  return request.post('/auth/login', data)
}

export function register(data: RegisterParams) {
  return request.post('/auth/register', data)
}

export function refreshToken() {
  return request.post('/auth/refresh', data)
}

export function logout() {
  return request.post('/auth/logout')
}

export function getProfile() {
  return request.get('/users/profile')
}

export function updateProfile(data: any) {
  return request.put('/users/profile', data)
}