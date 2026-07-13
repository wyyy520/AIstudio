import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as apiLogin, getProfile } from '@/api/auth'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('aistudio_token') || '')
  const user = ref<any>(null)

  const isLoggedIn = computed(() => !!token.value)
  const username = computed(() => user.value?.username || '')

  async function login(usernameVal: string, password: string) {
    const res: any = await apiLogin({ username: usernameVal, password })
    if (res.data?.token) {
      token.value = res.data.token
      localStorage.setItem('aistudio_token', res.data.token)
      await fetchProfile()
    }
    return res
  }

  async function fetchProfile() {
    try {
      const res: any = await getProfile()
      user.value = res.data
    } catch {
      user.value = null
    }
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('aistudio_token')
  }

  return { token, user, isLoggedIn, username, login, fetchProfile, logout }
})