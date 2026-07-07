import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export type ThemeMode = 'light' | 'dark' | 'system'

export const useThemeStore = defineStore('theme', () => {
  const mode = ref<ThemeMode>((localStorage.getItem('theme_mode') as ThemeMode) || 'dark')
  const resolvedTheme = ref<'light' | 'dark'>('dark')

  function applyTheme(theme: 'light' | 'dark') {
    document.documentElement.setAttribute('data-theme', theme)
    resolvedTheme.value = theme
  }

  function resolveSystemTheme(): 'light' | 'dark' {
    if (window.matchMedia('(prefers-color-scheme: light)').matches) {
      return 'light'
    }
    return 'dark'
  }

  function setMode(newMode: ThemeMode) {
    mode.value = newMode
    localStorage.setItem('theme_mode', newMode)

    if (newMode === 'system') {
      applyTheme(resolveSystemTheme())
    } else {
      applyTheme(newMode)
    }
  }

  // Initialize theme on store creation
  function init() {
    if (mode.value === 'system') {
      applyTheme(resolveSystemTheme())
      // Listen for system theme changes
      window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
        if (mode.value === 'system') {
          applyTheme(resolveSystemTheme())
        }
      })
    } else {
      applyTheme(mode.value)
    }
  }

  // Watch for mode changes
  watch(mode, (newMode) => {
    if (newMode === 'system') {
      applyTheme(resolveSystemTheme())
    }
  })

  return {
    mode,
    resolvedTheme,
    setMode,
    init,
  }
})