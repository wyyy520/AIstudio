/**
 * Unified API Client for AIStudio Frontend
 *
 * This is the central entry point for all API requests.
 * All API modules should use this client for consistency.
 */

import http from './request'
import type { AxiosRequestConfig, AxiosResponse } from 'axios'

export interface ApiResponse<T = any> {
  code: number
  message: string
  data?: T
}

class ApiClient {
  /**
   * Generic GET request
   */
  async get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const res = await http.get<ApiResponse<T>>(url, config)
    return this.unwrap<T>(res)
  }

  /**
   * Generic POST request
   */
  async post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const res = await http.post<ApiResponse<T>>(url, data, config)
    return this.unwrap<T>(res)
  }

  /**
   * Generic PUT request
   */
  async put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const res = await http.put<ApiResponse<T>>(url, data, config)
    return this.unwrap<T>(res)
  }

  /**
   * Generic DELETE request
   */
  async delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const res = await http.delete<ApiResponse<T>>(url, config)
    return this.unwrap<T>(res)
  }

  /**
   * Unwrap the API response
   */
  private unwrap<T>(res: AxiosResponse<ApiResponse<T>>): T {
    const data = res.data
    if (data.code !== 0) {
      throw new Error(data.message || 'Request failed')
    }
    return data.data as T
  }

  /**
   * Get the full base URL
   */
  getBaseUrl(): string {
    return http.defaults.baseURL || 'http://localhost:8081'
  }

  /**
   * Get the WebSocket URL
   */
  getWsUrl(): string {
    const base = this.getBaseUrl()
    return base.replace(/^http/, 'ws')
  }
}

/**
 * Global API client instance
 */
export const apiClient = new ApiClient()

export default apiClient