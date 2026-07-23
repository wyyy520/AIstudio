import request from './request'
import { isTauri, getBackendUrl } from '@/tauri'

export interface ApiResponse<T = any> {
  code: number
  message: string
  data?: T
}

class ApiClient {
  private baseUrl = (import.meta.env.VITE_API_BASE_URL || '') + '/api'

  async init() {
    if (isTauri()) {
      const backendUrl = await getBackendUrl()
      this.baseUrl = `${backendUrl}/api`
    }
  }

  async get<T = any>(url: string, config?: any): Promise<T> {
    const res = await request.get(url, { ...config, baseURL: this.baseUrl })
    return this.unwrap<T>(res)
  }

  async post<T = any>(url: string, data?: any, config?: any): Promise<T> {
    const res = await request.post(url, data, { ...config, baseURL: this.baseUrl })
    return this.unwrap<T>(res)
  }

  async put<T = any>(url: string, data?: any, config?: any): Promise<T> {
    const res = await request.put(url, data, { ...config, baseURL: this.baseUrl })
    return this.unwrap<T>(res)
  }

  async delete<T = any>(url: string, config?: any): Promise<T> {
    const res = await request.delete(url, { ...config, baseURL: this.baseUrl })
    return this.unwrap<T>(res)
  }

  private unwrap<T>(res: any): T {
    if (res && typeof res === 'object' && 'code' in res) {
      if (res.code !== 0) {
        throw new Error(res.message || 'Request failed')
      }
      return res.data as T
    }
    if (res && typeof res === 'object' && 'data' in res) {
      return res.data as T
    }
    return res as T
  }

  getBaseUrl(): string {
    return this.baseUrl
  }

  getWsUrl(): string {
    return this.baseUrl.replace(/^http/, 'ws') + '/v1/ws'
  }
}

export const apiClient = new ApiClient()

export default apiClient