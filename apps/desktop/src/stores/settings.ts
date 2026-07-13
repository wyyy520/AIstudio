import { defineStore } from 'pinia'
import { ref, reactive, computed } from 'vue'
import * as settingsApi from '@/api/settings'
import type { AppSettings } from '@/api/settings'
import { useThemeStore } from './theme'

const DEFAULT_SHORTCUTS: Record<string, string> = {
  'search': 'Ctrl+K',
  'run-agent': 'Ctrl+Enter',
  'new-project': 'Ctrl+Shift+N',
  'save': 'Ctrl+S',
  'toggle-sidebar': 'Ctrl+B',
  'command-palette': 'Ctrl+Shift+P',
}

export const useSettingsStore = defineStore('appSettings', () => {
  const loading = ref(false)
  const saving = ref(false)
  const error = ref<string | null>(null)

  const general = reactive({
    username: 'Admin',
    email: 'admin@aistudio.dev',
    language: 'zh-CN',
    autoSave: true,
    autoSaveInterval: 30,
    startupBehavior: 'dashboard' as 'dashboard' | 'last-project' | 'empty',
  })

  const engine = reactive<AppSettings['engine']>({
    provider: 'openai',
    model: 'gpt-4',
    apiKey: '',
    endpoint: 'https://api.openai.com/v1',
    timeout: 30000,
  })

  const shortcuts = reactive<Record<string, string>>({ ...DEFAULT_SHORTCUTS })

  const availableLanguages = [
    { value: 'zh-CN', label: '简体中文' },
    { value: 'en-US', label: 'English' },
    { value: 'ja-JP', label: '日本語' },
  ]

  const engineModels: Record<string, string[]> = {
    openai: ['gpt-4', 'gpt-4-turbo', 'gpt-3.5-turbo', 'gpt-4o', 'gpt-4o-mini'],
    claude: ['claude-3-opus', 'claude-3-sonnet', 'claude-3-haiku'],
    local: ['llama3', 'mistral', 'qwen', 'deepseek'],
  }

  // Backward compatibility - theme/font/sidebar shortcuts
  const theme = computed(() => {
    const themeStore = useThemeStore()
    return themeStore.mode
  })

  const fontSize = ref<number>(Number(localStorage.getItem('aistudio_fontsize')) || 14)
  const sidebarCollapsed = ref(false)

  function setTheme(value: 'dark' | 'light' | 'system') {
    const themeStore = useThemeStore()
    themeStore.setMode(value)
  }

  function setFontSize(size: number) {
    fontSize.value = size
    localStorage.setItem('aistudio_fontsize', String(size))
    document.documentElement.style.fontSize = `${size}px`
  }

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  async function fetchSettings() {
    loading.value = true
    error.value = null
    try {
      const data = await settingsApi.getSettings()
      if (data) {
        Object.assign(general, {
          username: general.username,
          email: general.email,
          language: data.language || general.language,
          autoSave: data.autoSave ?? general.autoSave,
          autoSaveInterval: data.autoSaveInterval || general.autoSaveInterval,
          startupBehavior: data.startupBehavior || general.startupBehavior,
        })
        if (data.engine) {
          Object.assign(engine, data.engine)
        }
        if (data.shortcuts) {
          Object.assign(shortcuts, data.shortcuts)
        }
        if (data.theme) {
          const themeStore = useThemeStore()
          themeStore.setMode(data.theme)
        }
      }
    } catch (e) {
      console.warn('[settings] failed to fetch from backend, using defaults:', e)
    } finally {
      loading.value = false
    }
  }

  async function saveGeneral() {
    saving.value = true
    error.value = null
    try {
      await settingsApi.updateSettings({
        language: general.language,
        autoSave: general.autoSave,
        autoSaveInterval: general.autoSaveInterval,
        startupBehavior: general.startupBehavior,
      })
    } catch (e) {
      error.value = 'Failed to save settings'
      console.error('[settings] save failed:', e)
    } finally {
      saving.value = false
    }
  }

  async function saveEngine() {
    saving.value = true
    error.value = null
    try {
      await settingsApi.updateEngineConfig({ ...engine })
    } catch (e) {
      error.value = 'Failed to save engine config'
      console.error('[settings] save engine failed:', e)
    } finally {
      saving.value = false
    }
  }

  function saveShortcuts() {
    localStorage.setItem('shortcuts', JSON.stringify({ ...shortcuts }))
  }

  function loadShortcuts() {
    const saved = localStorage.getItem('shortcuts')
    if (saved) {
      try {
        const parsed = JSON.parse(saved)
        Object.assign(shortcuts, parsed)
      } catch {
        // ignore
      }
    }
  }

  function resetShortcuts() {
    Object.assign(shortcuts, DEFAULT_SHORTCUTS)
    saveShortcuts()
  }

  function getShortcutLabel(action: string): string {
    return shortcuts[action] || ''
  }

  return {
    loading,
    saving,
    error,
    general,
    engine,
    shortcuts,
    availableLanguages,
    engineModels,
    theme,
    fontSize,
    sidebarCollapsed,
    setTheme,
    setFontSize,
    toggleSidebar,
    fetchSettings,
    saveGeneral,
    saveEngine,
    saveShortcuts,
    loadShortcuts,
    resetShortcuts,
    getShortcutLabel,
  }
})