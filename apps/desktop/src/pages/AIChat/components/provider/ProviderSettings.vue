<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="visible" class="provider-settings-overlay" @click.self="$emit('close')">
        <div class="provider-settings-modal">
          <div class="provider-settings-header">
            <h3 class="provider-settings-title">{{ provider.name }} 设置</h3>
            <button class="provider-settings-close" @click="$emit('close')">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                <line x1="18" y1="6" x2="6" y2="18" />
                <line x1="6" y1="6" x2="18" y2="18" />
              </svg>
            </button>
          </div>

          <div class="provider-settings-body">
            <div class="settings-field">
              <label class="settings-label">Provider</label>
              <input class="settings-input" :value="provider.name" disabled />
            </div>

            <div class="settings-field">
              <label class="settings-label">API Base URL</label>
              <input class="settings-input" v-model="form.apiBaseUrl" placeholder="https://api.example.com/v1" />
            </div>

            <div class="settings-field">
              <label class="settings-label">API Key</label>
              <div class="api-key-input">
                <input class="settings-input" v-model="form.apiKey" :type="showKey ? 'text' : 'password'" placeholder="sk-..." />
                <button class="toggle-key-btn" @click="showKey = !showKey" :title="showKey ? '隐藏' : '显示'">
                  <svg v-if="showKey" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24" />
                    <line x1="1" y1="1" x2="23" y2="23" />
                  </svg>
                  <svg v-else viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                    <circle cx="12" cy="12" r="3" />
                  </svg>
                </button>
              </div>
            </div>

            <div class="settings-field">
              <label class="settings-label">默认模型</label>
              <select class="settings-input" v-model="form.model">
                <option v-for="m in provider.models" :key="m.id" :value="m.id">{{ m.name }}</option>
              </select>
            </div>

            <div class="settings-row">
              <div class="settings-field settings-field-half">
                <label class="settings-label">Temperature</label>
                <input class="settings-input" v-model.number="form.temperature" type="number" min="0" max="2" step="0.1" />
              </div>
              <div class="settings-field settings-field-half">
                <label class="settings-label">Max Tokens</label>
                <input class="settings-input" v-model.number="form.maxTokens" type="number" min="1" />
              </div>
            </div>

            <div v-if="testResult" class="settings-test-result" :class="`settings-test-result--${testResult.status}`">
              {{ testResult.message }}
            </div>
          </div>

          <div class="provider-settings-footer">
            <button class="settings-btn settings-btn-danger" @click="$emit('delete')">清除 API Key</button>
            <div class="settings-footer-right">
              <button class="settings-btn settings-btn-outline" @click="handleTest" :disabled="testing">
                <svg v-if="testing" class="spin" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12a9 9 0 1 1-6.219-8.56" /></svg>
                {{ testing ? '测试中...' : '测试连接' }}
              </button>
              <button class="settings-btn settings-btn-primary" @click="handleSave">保存</button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import type { AIProvider } from '../../types'

const props = defineProps<{
  provider: AIProvider
  visible: boolean
  testConnection?: (data: Record<string, unknown>) => Promise<{ success: boolean; message: string }>
}>()

const emit = defineEmits<{
  close: []
  save: [data: Record<string, unknown>]
  delete: []
}>()

const form = reactive({
  apiBaseUrl: '',
  apiKey: '',
  model: '',
  temperature: 0.7,
  maxTokens: 4096,
})

const showKey = ref(false)
const testing = ref(false)
const testResult = ref<{ status: 'success' | 'error'; message: string } | null>(null)

watch(() => props.visible, (v) => {
  if (v) {
    form.apiBaseUrl = props.provider.apiBaseUrl
    form.apiKey = props.provider.apiKey
    form.model = props.provider.models[0]?.id || ''
    showKey.value = false
    testResult.value = null
  }
})

async function handleTest() {
  if (!props.testConnection) {
    testResult.value = { status: 'error', message: '未提供测试功能' }
    return
  }
  testing.value = true
  testResult.value = null
  try {
    const result = await props.testConnection({ ...form })
    testResult.value = {
      status: result.success ? 'success' : 'error',
      message: result.message,
    }
  } catch {
    testResult.value = { status: 'error', message: '测试连接失败' }
  } finally {
    testing.value = false
  }
}

function handleSave() {
  emit('save', { ...form })
}
</script>

<style scoped>
.provider-settings-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.provider-settings-modal {
  width: 480px;
  max-height: 80vh;
  background: var(--bg-secondary);
  border-radius: var(--radius-2xl);
  box-shadow: var(--shadow-xl);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.provider-settings-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-4) var(--spacing-4) var(--spacing-3);
  border-bottom: 1px solid var(--border-subtle);
}

.provider-settings-title {
  font-size: var(--text-h3);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.provider-settings-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  color: var(--text-tertiary);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.provider-settings-close:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.provider-settings-body {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-4);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.settings-field {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1);
}

.settings-row {
  display: flex;
  gap: var(--spacing-3);
}

.settings-field-half {
  flex: 1;
}

.settings-label {
  font-size: var(--text-caption);
  color: var(--text-secondary);
}

.settings-input {
  height: 36px;
  padding: 0 var(--spacing-3);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-size: var(--text-body);
  font-family: var(--font-family-sans);
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
  outline: none;
}

.settings-input:focus {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px var(--primary-bg);
}

.settings-input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

select.settings-input {
  cursor: pointer;
}

.api-key-input {
  display: flex;
  gap: var(--spacing-1);
}

.api-key-input .settings-input {
  flex: 1;
}

.toggle-key-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--bg-tertiary);
  color: var(--text-tertiary);
  cursor: pointer;
  flex-shrink: 0;
  transition: all var(--transition-fast);
}

.toggle-key-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.settings-test-result {
  padding: var(--spacing-2) var(--spacing-3);
  border-radius: var(--radius-md);
  font-size: var(--text-body-sm);
}

.settings-test-result--success {
  background: var(--success-bg);
  color: var(--success);
}

.settings-test-result--error {
  background: var(--error-bg);
  color: var(--error);
}

.provider-settings-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-3) var(--spacing-4);
  border-top: 1px solid var(--border-subtle);
}

.settings-footer-right {
  display: flex;
  gap: var(--spacing-2);
}

.settings-btn {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-1);
  height: 32px;
  padding: 0 var(--spacing-3);
  border: none;
  border-radius: var(--radius-md);
  font-size: var(--text-body-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
  font-family: var(--font-family-sans);
}

.settings-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.settings-btn-primary {
  background: var(--primary);
  color: white;
}

.settings-btn-primary:hover:not(:disabled) {
  background: var(--primary-hover);
}

.settings-btn-outline {
  background: transparent;
  color: var(--text-primary);
  border: 1px solid var(--border-default);
}

.settings-btn-outline:hover:not(:disabled) {
  background: var(--bg-hover);
}

.settings-btn-danger {
  background: transparent;
  color: var(--error);
  border: 1px solid var(--error);
}

.settings-btn-danger:hover {
  background: var(--error-bg);
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.modal-enter-active,
.modal-leave-active {
  transition: opacity 200ms ease;
}
.modal-enter-active .provider-settings-modal,
.modal-leave-active .provider-settings-modal {
  transition: transform 250ms cubic-bezier(0.4, 0, 0.2, 1), opacity 200ms ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
.modal-enter-from .provider-settings-modal {
  transform: scale(0.95);
  opacity: 0;
}
.modal-leave-to .provider-settings-modal {
  transform: scale(0.95);
  opacity: 0;
}
</style>
