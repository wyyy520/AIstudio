<template>
  <div class="settings">
    <div class="settings__header">
      <h3 class="settings__title">设置</h3>
    </div>

    <div class="settings__body">
      <aside class="settings-nav">
        <div
          v-for="tab in tabs"
          :key="tab.key"
          :class="['settings-nav__item', { 'settings-nav__item--active': activeTab === tab.key }]"
          @click="activeTab = tab.key"
        >
          {{ tab.label }}
        </div>
      </aside>

      <main class="settings-content">
        <div v-if="activeTab === 'general'" class="settings-section">
          <h4 class="section-title">通用设置</h4>
          <div class="form-row">
            <label class="form-label">主题</label>
            <div class="theme-switcher">
              <button
                v-for="t in themes"
                :key="t.value"
                :class="['theme-btn', { 'theme-btn--active': settingsStore.theme === t.value }]"
                @click="settingsStore.setTheme(t.value)"
              >{{ t.label }}</button>
            </div>
          </div>
          <div class="form-row">
            <label class="form-label">字体大小</label>
            <div class="font-size-control">
              <button class="size-btn" @click="settingsStore.setFontSize(Math.max(12, settingsStore.fontSize - 1))">A-</button>
              <span class="size-value">{{ settingsStore.fontSize }}px</span>
              <button class="size-btn" @click="settingsStore.setFontSize(Math.min(20, settingsStore.fontSize + 1))">A+</button>
            </div>
          </div>
          <div class="form-row">
            <label class="form-label">语言</label>
            <select v-model="language" class="form-select">
              <option value="zh-CN">简体中文</option>
              <option value="en">English</option>
            </select>
          </div>
        </div>

        <div v-if="activeTab === 'engine'" class="settings-section">
          <h4 class="section-title">引擎配置</h4>
          <div class="form-row">
            <label class="form-label">默认 LLM 引擎</label>
            <select v-model="engineConfig.defaultEngine" class="form-select">
              <option value="openai">OpenAI</option>
              <option value="ollama">Ollama</option>
              <option value="anthropic">Anthropic</option>
            </select>
          </div>
          <div class="form-row">
            <label class="form-label">API 地址</label>
            <input v-model="engineConfig.apiUrl" class="form-input" placeholder="https://api.example.com" />
          </div>
          <div class="form-row">
            <label class="form-label">API Key</label>
            <input v-model="engineConfig.apiKey" type="password" class="form-input" placeholder="sk-..." />
          </div>
          <div class="form-row">
            <AppButton type="primary" size="small" @click="saveEngineConfig">保存配置</AppButton>
          </div>
        </div>

        <div v-if="activeTab === 'model'" class="settings-section">
          <h4 class="section-title">模型配置</h4>
          <div class="form-row">
            <label class="form-label">默认模型</label>
            <select v-model="modelConfig.defaultModel" class="form-select">
              <option value="gpt-4">GPT-4</option>
              <option value="gpt-3.5-turbo">GPT-3.5 Turbo</option>
              <option value="claude-3-opus">Claude 3 Opus</option>
              <option value="claude-3-sonnet">Claude 3 Sonnet</option>
              <option value="llama3">Llama 3</option>
              <option value="qwen2">Qwen 2</option>
            </select>
          </div>
          <div class="form-row">
            <label class="form-label">Temperature</label>
            <div class="slider-control">
              <input type="range" v-model.number="modelConfig.temperature" min="0" max="2" step="0.1" class="slider" />
              <span class="slider-value">{{ modelConfig.temperature }}</span>
            </div>
          </div>
          <div class="form-row">
            <label class="form-label">最大 Tokens</label>
            <input v-model.number="modelConfig.maxTokens" type="number" class="form-input" />
          </div>
          <div class="form-row">
            <AppButton type="primary" size="small" @click="saveModelConfig">保存配置</AppButton>
          </div>
        </div>

        <div v-if="activeTab === 'plugins'" class="settings-section">
          <h4 class="section-title">插件管理</h4>
          <p class="section-desc">在插件中心管理插件的安装与配置</p>
          <AppButton type="secondary" size="small" @click="$router.push('/plugins')">前往插件中心</AppButton>
        </div>

        <div v-if="activeTab === 'shortcuts'" class="settings-section">
          <h4 class="section-title">快捷键</h4>
          <div class="shortcut-list">
            <div class="shortcut-item">
              <span class="shortcut-label">新建项目</span>
              <kbd class="shortcut-key">Ctrl + N</kbd>
            </div>
            <div class="shortcut-item">
              <span class="shortcut-label">保存</span>
              <kbd class="shortcut-key">Ctrl + S</kbd>
            </div>
            <div class="shortcut-item">
              <span class="shortcut-label">运行工作流</span>
              <kbd class="shortcut-key">Ctrl + R</kbd>
            </div>
            <div class="shortcut-item">
              <span class="shortcut-label">搜索</span>
              <kbd class="shortcut-key">Ctrl + K</kbd>
            </div>
            <div class="shortcut-item">
              <span class="shortcut-label">切换侧栏</span>
              <kbd class="shortcut-key">Ctrl + B</kbd>
            </div>
            <div class="shortcut-item">
              <span class="shortcut-label">设置</span>
              <kbd class="shortcut-key">Ctrl + ,</kbd>
            </div>
          </div>
        </div>

        <div v-if="activeTab === 'about'" class="settings-section">
          <h4 class="section-title">关于 AIStudio</h4>
          <div class="about-info">
            <div class="about-row">
              <span class="about-label">应用名称</span>
              <span class="about-value">AIStudio</span>
            </div>
            <div class="about-row">
              <span class="about-label">版本</span>
              <span class="about-value">0.1.0 (Beta)</span>
            </div>
            <div class="about-row">
              <span class="about-label">运行环境</span>
              <span class="about-value">{{ isTauriEnv ? 'Tauri 桌面端' : 'Web 浏览器' }}</span>
            </div>
            <div class="about-row">
              <span class="about-label">运行模式</span>
              <span class="about-value">{{ isDevMode ? '开发模式' : '发布模式' }}</span>
            </div>
            <div class="about-row">
              <span class="about-label">许可证</span>
              <span class="about-value">MIT License</span>
            </div>
            <div class="about-row">
              <span class="about-label">技术栈</span>
              <span class="about-value">Vue 3 + TypeScript + Tauri + Go</span>
            </div>
            <div class="about-divider" />
            <p class="about-copyright">© 2026 AI Studio. All rights reserved.</p>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import AppButton from '@/components/AppButton.vue'

const settingsStore = useSettingsStore()
const activeTab = ref('general')
const language = ref('zh-CN')
const isTauriEnv = false
const isDevMode = ref(true)

onMounted(() => {
  // Detect environment mode
  isDevMode.value = import.meta.env.DEV || import.meta.env.VITE_APP_ENV === 'development'
})

const tabs = [
  { key: 'general', label: '通用' },
  { key: 'engine', label: '引擎配置' },
  { key: 'model', label: '模型配置' },
  { key: 'plugins', label: '插件管理' },
  { key: 'shortcuts', label: '快捷键' },
  { key: 'about', label: '关于' },
]

const themes = [
  { value: 'dark' as const, label: '深色' },
  { value: 'light' as const, label: '浅色' },
  { value: 'system' as const, label: '跟随系统' },
]

const engineConfig = reactive({
  defaultEngine: localStorage.getItem('aistudio_engine') || 'openai',
  apiUrl: localStorage.getItem('aistudio_api_url') || import.meta.env.VITE_API_BASE_URL || '',
  apiKey: localStorage.getItem('aistudio_api_key') || '',
})

const modelConfig = reactive({
  defaultModel: localStorage.getItem('aistudio_model') || 'gpt-4',
  temperature: Number(localStorage.getItem('aistudio_temp')) || 0.7,
  maxTokens: Number(localStorage.getItem('aistudio_max_tokens')) || 4096,
})

function saveEngineConfig() {
  localStorage.setItem('aistudio_engine', engineConfig.defaultEngine)
  localStorage.setItem('aistudio_api_url', engineConfig.apiUrl)
  localStorage.setItem('aistudio_api_key', engineConfig.apiKey)
}

function saveModelConfig() {
  localStorage.setItem('aistudio_model', modelConfig.defaultModel)
  localStorage.setItem('aistudio_temp', String(modelConfig.temperature))
  localStorage.setItem('aistudio_max_tokens', String(modelConfig.maxTokens))
}
</script>

<style scoped>
.settings {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.settings__header {
  padding: 16px 24px;
  border-bottom: 1px solid var(--border-subtle);
  background: var(--bg-secondary);
  flex-shrink: 0;
}

.settings__title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.settings__body {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.settings-nav {
  width: 180px;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-subtle);
  padding: 12px 8px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex-shrink: 0;
}

.settings-nav__item {
  padding: 8px 12px;
  border-radius: var(--radius-xs);
  font-size: 13px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.settings-nav__item:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.settings-nav__item--active {
  background: var(--bg-active);
  color: var(--primary);
}

.settings-content {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
}

.settings-section {
  max-width: 600px;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 20px;
}

.section-desc {
  font-size: 13px;
  color: var(--text-tertiary);
  margin-bottom: 16px;
}

.form-row {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
  gap: 16px;
}

.form-label {
  width: 100px;
  font-size: 13px;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.form-input {
  flex: 1;
  height: 36px;
  padding: 0 12px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-xs);
  color: var(--text-primary);
  font-size: 14px;
  outline: none;
}

.form-input:focus {
  border-color: var(--primary);
}

.form-select {
  flex: 1;
  height: 36px;
  padding: 0 12px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-xs);
  color: var(--text-primary);
  font-size: 14px;
  outline: none;
}

.theme-switcher {
  display: flex;
  gap: 4px;
}

.theme-btn {
  padding: 6px 16px;
  border-radius: var(--radius-xs);
  font-size: 13px;
  color: var(--text-secondary);
  background: var(--bg-tertiary);
  transition: all var(--transition-fast);
}

.theme-btn:hover {
  background: var(--bg-hover);
}

.theme-btn--active {
  background: var(--primary);
  color: #fff;
}

.font-size-control {
  display: flex;
  align-items: center;
  gap: 12px;
}

.size-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-xs);
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  font-size: 14px;
  transition: all var(--transition-fast);
}

.size-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.size-value {
  font-size: 14px;
  color: var(--text-primary);
  min-width: 40px;
  text-align: center;
  font-variant-numeric: tabular-nums;
}

.slider-control {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.slider {
  flex: 1;
  accent-color: var(--primary);
}

.slider-value {
  font-size: 14px;
  color: var(--text-primary);
  min-width: 30px;
  text-align: center;
  font-variant-numeric: tabular-nums;
}

.shortcut-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.shortcut-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 0;
  border-bottom: 1px solid var(--border-subtle);
}

.shortcut-label {
  font-size: 13px;
  color: var(--text-primary);
}

.shortcut-key {
  padding: 4px 10px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: 4px;
  font-size: 12px;
  font-family: var(--font-mono);
  color: var(--text-secondary);
}

.about-info {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.about-row {
  display: flex;
  align-items: center;
  gap: 16px;
}

.about-label {
  width: 100px;
  font-size: 13px;
  color: var(--text-secondary);
}

.about-value {
  font-size: 14px;
  color: var(--text-primary);
}

.about-divider {
  height: 1px;
  background: var(--border-subtle);
  margin: 8px 0;
}

.about-copyright {
  font-size: 12px;
  color: var(--text-tertiary);
  text-align: center;
  margin-top: 8px;
}
</style>