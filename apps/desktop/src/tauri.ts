import { invoke } from '@tauri-apps/api/core'

export function isTauri(): boolean {
  return typeof window !== 'undefined' && window.__TAURI__ !== undefined
}

export async function getBackendUrl(): Promise<string> {
  if (isTauri()) {
    return await invoke('get_backend_url')
  }
  if (typeof window !== 'undefined') {
    return window.location.origin
  }
  return ''
}