<template>
  <div class="settings-section">
    <h2 class="section-title">Engine 配置</h2>
    <p class="section-desc">配置 AI 模型提供商、API 密钥和连接参数。</p>

    <!-- Provider 选择 -->
    <div class="settings-card">
      <h3 class="card-title">模型提供商</h3>
      <div class="provider-tabs">
        <button
          v-for="p in providers"
          :key="p.id"
          class="provider-tab"
          :class="{ active: store.engine.provider === p.id }"
          @click="store.engine.provider = p.id"
        >
          <span class="provider-tab-name">{{ p.name }}</span>
          <span class="provider-tab-desc">{{ p.desc }}</span>
        </button>
      </div>
    </div>

    <!-- Connection 配置 -->
    <div class="settings-card">
      <h3 class="card-title">连接配置</h3>

      <div class="form-field">
        <label class="field-label">Endpoint</label>
        <input
          class="field-input"
          v-model="store.engine.endpoint"
          placeholder="https://api.openai.com/v1"
        />
      </div>

      <div class="form-field">
        <label class="field-label">API Key</label>
        <div class="password-field">
          <input
            class="field-input"
            :type="showKey ? 'text' : 'password'"
            v-model="store.engine.apiKey"
            placeholder="sk-..."
          />
          <button class="password-toggle" @click="showKey = !showKey" :title="showKey ? '隐藏' : '显示'">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
              <path v-if="showKey" d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
              <circle v-if="showKey" cx="12" cy="12" r="3" />
              <template v-else>
                <path d="M9.88 9.88a3 3 0 1 0 4.24 4.24" />
                <path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68" />
                <path d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.62 9.62 0 0 0 4.29-.89" />
                <path d="M4 4l16 16" />
              </template>
            </svg>
          </button>
        </div>
      </div>

      <div class="form-field">
        <label class="field-label">模型</label>
        <select class="field-select" v-model="store.engine.model">
          <option
            v-for="m in currentModels"
            :key="m"
            :value="m"
          >
            {{ m }}
          </option>
        </select>
      </div>

      <div class="form-field">
        <label class="field-label">Timeout (ms)</label>
        <input
          class="field-input"
          v-model.number="store.engine.timeout"
          type="number"
          min="1000"
          step="1000"
        />
      </div>
    </div>

    <!-- 操作 -->
    <div class="settings-actions">
      <AppButton
        type="secondary"
        size="medium"
        :loading="testing"
        @click="handleTest"
      >
        测试连接
      </AppButton>
      <AppButton
        type="primary"
        size="medium"
        :loading="store.saving"
        @click="store.saveEngine()"
      >
        保存配置
      </AppButton>
      <span v-if="testResult" class="test-result" :class="testResult.success ? 'test-success' : 'test-fail'">
        {{ testResult.message }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useSettingsStore } from '@/store/settings'
import { testEngineConnection } from '@/api/settings'
import AppButton from '@/components/AppButton/AppButton.vue'

const store = useSettingsStore()
const showKey = ref(false)
const testing = ref(false)
const testResult = ref<{ success: boolean; message: string } | null>(null)

interface ProviderOption {
  id: 'openai' | 'claude' | 'local'
  name: string
  desc: string
}

const providers: ProviderOption[] = [
  { id: 'openai', name: 'OpenAI', desc: 'GPT-4 / GPT-3.5' },
  { id: 'claude', name: 'Claude', desc: 'Claude 3 Opus / Sonnet' },
  { id: 'local', name: '本地模型', desc: 'Llama / Mistral / Qwen' },
]

const currentModels = computed(() => {
  return store.engineModels[store.engine.provider] || []
})

async function handleTest() {
  testing.value = true
  testResult.value = null
  try {
    const result = await testEngineConnection({
      provider: store.engine.provider,
      endpoint: store.engine.endpoint,
      apiKey: store.engine.apiKey,
    })
    testResult.value = result
  } catch {
    testResult.value = { success: false, message: '连接失败：无法访问服务器' }
  } finally {
    testing.value = false
  }
}
</script>

<style scoped>
.settings-section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-6);
}

.section-title {
  font-size: var(--text-h2);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  line-height: var(--leading-h2);
}

.section-desc {
  font-size: var(--text-body);
  color: var(--text-secondary);
  line-height: var(--leading-body);
  margin-top: -12px;
}

.settings-card {
  background: var(--bg-tertiary);
  border-radius: var(--radius-xl);
  padding: var(--spacing-4);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.card-title {
  font-size: var(--text-body);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  line-height: var(--leading-body);
  padding-bottom: var(--spacing-2);
  border-bottom: 1px solid var(--border-subtle);
}

/* Provider Tabs */
.provider-tabs {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--spacing-2);
}

.provider-tab {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: var(--spacing-3) var(--spacing-2);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-default);
  background: var(--bg-primary);
  cursor: pointer;
  transition: all var(--transition-fast);
  text-align: center;
}

.provider-tab:hover {
  border-color: var(--primary);
  background: var(--primary-bg);
}

.provider-tab.active {
  border-color: var(--primary);
  background: var(--primary-bg);
  box-shadow: 0 0 0 1px var(--primary);
}

.provider-tab-name {
  font-size: var(--text-body);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.provider-tab-desc {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

/* Form fields */
.form-field {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1);
}

.field-label {
  font-size: var(--text-body-sm);
  color: var(--text-secondary);
  line-height: var(--leading-body-sm);
}

.field-input {
  height: 36px;
  padding: 0 var(--spacing-3);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-default);
  background: var(--bg-primary);
  color: var(--text-primary);
  font-size: var(--text-body);
  transition: border-color var(--transition-fast);
}

.field-input:focus {
  border-color: var(--primary);
  box-shadow: 0 0 0 1px var(--primary);
}

.field-select {
  height: 36px;
  padding: 0 var(--spacing-3);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-default);
  background: var(--bg-primary);
  color: var(--text-primary);
  font-size: var(--text-body);
  cursor: pointer;
  transition: border-color var(--transition-fast);
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 24 24' width='16' height='16' fill='none' stroke='%2371717a' stroke-width='1.5' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 10px center;
  padding-right: 36px;
}

.field-select:focus {
  border-color: var(--primary);
  box-shadow: 0 0 0 1px var(--primary);
}

/* Password field */
.password-field {
  position: relative;
}

.password-field .field-input {
  width: 100%;
  padding-right: 40px;
}

.password-toggle {
  position: absolute;
  right: 4px;
  top: 50%;
  transform: translateY(-50%);
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
  transition: color var(--transition-fast);
}

.password-toggle:hover {
  color: var(--text-primary);
}

/* Actions */
.settings-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  padding-top: var(--spacing-2);
  flex-wrap: wrap;
}

.test-result {
  font-size: var(--text-body-sm);
  line-height: var(--leading-body-sm);
}

.test-success {
  color: var(--success);
}

.test-fail {
  color: var(--error);
}
</style>