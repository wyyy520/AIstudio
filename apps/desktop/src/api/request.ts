import axios from 'axios'
import { useUserStore } from '@/stores/user'

const _request = axios.create({
  baseURL: (import.meta.env.VITE_API_BASE_URL || '') + '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

_request.interceptors.request.use(
  (config) => {
    const userStore = useUserStore()
    if (userStore.token) {
      config.headers.Authorization = `Bearer ${userStore.token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

_request.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error) => {
    if (error.response?.status === 401) {
      const userStore = useUserStore()
      userStore.logout()
    }
    const message = error.response?.data?.message || error.message || 'Request failed'
    return Promise.reject(new Error(message))
  }
)

/**
 * Typed wrapper around the axios instance.
 *
 * The response interceptor strips the AxiosResponse wrapper at runtime and
 * returns the response body directly, but axios's type declarations still
 * say `Promise<AxiosResponse<T>>`.  This wrapper overrides the return type
 * to match actual runtime behavior.
 */
const request = {
  get<T = any>(url: string, config?: any): Promise<T> {
    return _request.get(url, config) as unknown as Promise<T>
  },
  post<T = any>(url: string, data?: any, config?: any): Promise<T> {
    return _request.post(url, data, config) as unknown as Promise<T>
  },
  put<T = any>(url: string, data?: any, config?: any): Promise<T> {
    return _request.put(url, data, config) as unknown as Promise<T>
  },
  delete<T = any>(url: string, config?: any): Promise<T> {
    return _request.delete(url, config) as unknown as Promise<T>
  },
}

export default request