import axios, { type AxiosInstance, type AxiosResponse } from 'axios'

export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data?: T
}

const http: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor: attach token if available
http.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error),
)

// Response interceptor: unwrap data and handle errors
http.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const { data, status } = response
    if (status === 200 || status === 201) {
      return data as unknown as AxiosResponse
    }
    return response
  },
  (error) => {
    if (error.response) {
      const { status, data } = error.response
      switch (status) {
        case 401:
          localStorage.removeItem('auth_token')
          window.location.href = '/login'
          break
        case 403:
          console.error('[api] forbidden:', data?.message)
          break
        case 404:
          console.error('[api] not found:', data?.message)
          break
        case 500:
          console.error('[api] server error:', data?.message)
          break
      }
      return Promise.reject(new Error(data?.message || `HTTP ${status}`))
    }
    if (error.code === 'ECONNABORTED') {
      return Promise.reject(new Error('请求超时'))
    }
    return Promise.reject(error)
  },
)

export default http