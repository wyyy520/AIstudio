import { ref } from 'vue'
import type { AppInfo, SystemInfo } from '@/api/tauri'

export function useTauri() {
  const isTauri = ref(typeof window !== 'undefined' && window.__TAURI__ !== undefined)

  return { isTauri }
}

export function useAppInfo() {
  const info = ref<AppInfo | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetch() {
    if (!window.__TAURI__) {
      error.value = 'Not running in Tauri'
      return
    }
    loading.value = true
    error.value = null
    try {
      const { getAppInfo } = await import('@/api/tauri')
      info.value = await getAppInfo()
    } catch (e: any) {
      error.value = e?.message ?? String(e)
    } finally {
      loading.value = false
    }
  }

  return { info, loading, error, fetch }
}

export function useSystemInfo() {
  const info = ref<SystemInfo | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetch() {
    if (!window.__TAURI__) {
      error.value = 'Not running in Tauri'
      return
    }
    loading.value = true
    error.value = null
    try {
      const { getSystemInfo } = await import('@/api/tauri')
      info.value = await getSystemInfo()
    } catch (e: any) {
      error.value = e?.message ?? String(e)
    } finally {
      loading.value = false
    }
  }

  return { info, loading, error, fetch }
}
