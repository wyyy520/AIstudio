import axios from 'axios'
import { useUserStore } from '@/stores/user'

const request = axios.create({
  baseURL: (import.meta.env.VITE_API_BASE_URL || '') + '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

request.interceptors.request.use(
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

request.interceptors.response.use(
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

export default request