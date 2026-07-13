import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'
import './assets/styles/global.css'
import 'highlight.js/styles/github-dark.css'
import { useThemeStore } from '@/stores/theme'
import { apiClient } from '@/api/client'

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)
app.use(router)

// Initialize API client
apiClient.init()

// Initialize theme
const themeStore = useThemeStore()
themeStore.init()

app.mount('#app')