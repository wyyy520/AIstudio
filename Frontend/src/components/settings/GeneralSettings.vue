<template>
  <div class="settings-section">
    <h2 class="section-title">通用设置</h2>
    <p class="section-desc">配置用户信息、界面语言和应用启动行为。</p>

    <!-- 用户信息 -->
    <div class="settings-card">
      <h3 class="card-title">用户信息</h3>
      <div class="form-grid">
        <div class="form-field">
          <label class="field-label">用户名</label>
          <input
            class="field-input"
            v-model="store.general.username"
            disabled
            placeholder="用户名"
          />
        </div>
        <div class="form-field">
          <label class="field-label">邮箱</label>
          <input
            class="field-input"
            v-model="store.general.email"
            disabled
            placeholder="邮箱"
          />
        </div>
      </div>
    </div>

    <!-- 语言与区域 -->
    <div class="settings-card">
      <h3 class="card-title">语言与区域</h3>
      <div class="form-field">
        <label class="field-label">界面语言</label>
        <select class="field-select" v-model="store.general.language">
          <option
            v-for="lang in store.availableLanguages"
            :key="lang.value"
            :value="lang.value"
          >
            {{ lang.label }}
          </option>
        </select>
      </div>
    </div>

    <!-- 编辑器行为 -->
    <div class="settings-card">
      <h3 class="card-title">编辑器行为</h3>

      <div class="form-row">
        <div class="form-row-content">
          <span class="form-row-label">自动保存</span>
          <span class="form-row-hint">编辑内容时自动保存更改</span>
        </div>
        <AppSwitch v-model="store.general.autoSave" />
      </div>

      <div v-if="store.general.autoSave" class="form-field">
        <label class="field-label">自动保存间隔（秒）</label>
        <input
          class="field-input"
          v-model.number="store.general.autoSaveInterval"
          type="number"
          min="5"
          max="300"
          step="5"
        />
      </div>

      <div class="form-field">
        <label class="field-label">启动行为</label>
        <select class="field-select" v-model="store.general.startupBehavior">
          <option value="dashboard">显示仪表盘</option>
          <option value="last-project">打开上次项目</option>
          <option value="empty">空白界面</option>
        </select>
      </div>
    </div>

    <!-- 操作按钮 -->
    <div class="settings-actions">
      <AppButton
        type="primary"
        size="medium"
        :loading="store.saving"
        @click="store.saveGeneral()"
      >
        保存更改
      </AppButton>
      <span v-if="store.error" class="save-error">{{ store.error }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useSettingsStore } from '@/store/settings'
import AppSwitch from '@/components/AppSwitch/AppSwitch.vue'
import AppButton from '@/components/AppButton/AppButton.vue'

const store = useSettingsStore()
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

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--spacing-3);
}

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

.field-input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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

.form-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-2) 0;
}

.form-row-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.form-row-label {
  font-size: var(--text-body);
  color: var(--text-primary);
  line-height: var(--leading-body);
}

.form-row-hint {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.settings-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  padding-top: var(--spacing-2);
}

.save-error {
  font-size: var(--text-body-sm);
  color: var(--error);
}
</style>