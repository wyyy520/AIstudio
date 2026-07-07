<template>
  <div class="settings-section">
    <h2 class="section-title">主题设置</h2>
    <p class="section-desc">选择应用外观主题。支持浅色、深色和跟随系统。</p>

    <div class="settings-card">
      <h3 class="card-title">外观模式</h3>
      <div class="theme-options">
        <button
          v-for="option in themeOptions"
          :key="option.id"
          class="theme-card"
          :class="{ active: store.mode === option.id }"
          @click="store.setMode(option.id)"
        >
          <div class="theme-preview" :class="`theme-preview--${option.id}`">
            <div class="preview-toolbar">
              <span class="preview-dot" :class="option.id === 'dark' ? 'preview-dot-dark' : 'preview-dot-light'" />
              <span class="preview-dot" :class="option.id === 'dark' ? 'preview-dot-dark' : 'preview-dot-light'" />
              <span class="preview-dot" :class="option.id === 'dark' ? 'preview-dot-dark' : 'preview-dot-light'" />
            </div>
            <div class="preview-body">
              <div class="preview-sidebar" />
              <div class="preview-main">
                <div class="preview-line" />
                <div class="preview-line preview-line--short" />
              </div>
            </div>
          </div>
          <div class="theme-info">
            <span class="theme-name">{{ option.name }}</span>
            <span class="theme-desc">{{ option.desc }}</span>
          </div>
          <div class="theme-check" v-if="store.mode === option.id">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M20 6 9 17l-5-5" />
            </svg>
          </div>
        </button>
      </div>
    </div>

    <div class="settings-card">
      <h3 class="card-title">当前主题</h3>
      <div class="theme-info-row">
        <span class="info-label">当前模式</span>
        <span class="info-value">
          {{ store.mode === 'system' ? '跟随系统' : store.mode === 'dark' ? '深色' : '浅色' }}
          <span v-if="store.mode === 'system'">
            （{{ store.resolvedTheme === 'dark' ? '深色' : '浅色' }}）
          </span>
        </span>
      </div>
      <div class="theme-info-row">
        <span class="info-label">主题色</span>
        <span class="info-value">
          <span class="color-dot" style="background: var(--primary)" />
          紫色（#8b5cf6）
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useThemeStore } from '@/store/theme'

const store = useThemeStore()

interface ThemeOption {
  id: 'light' | 'dark' | 'system'
  name: string
  desc: string
}

const themeOptions: ThemeOption[] = [
  { id: 'light', name: '浅色模式', desc: '明亮的界面外观' },
  { id: 'dark', name: '深色模式', desc: '深邃的界面外观' },
  { id: 'system', name: '跟随系统', desc: '自动匹配系统主题' },
]
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

/* Theme options */
.theme-options {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--spacing-3);
}

.theme-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-3);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-default);
  background: var(--bg-primary);
  cursor: pointer;
  transition: all var(--transition-fast);
  position: relative;
  text-align: center;
}

.theme-card:hover {
  border-color: var(--primary);
}

.theme-card.active {
  border-color: var(--primary);
  box-shadow: 0 0 0 1px var(--primary);
  background: var(--primary-bg);
}

/* Theme preview */
.theme-preview {
  width: 100%;
  height: 80px;
  border-radius: var(--radius-md);
  overflow: hidden;
  border: 1px solid var(--border-subtle);
}

.theme-preview--dark {
  background: #0f0f11;
}

.theme-preview--light {
  background: #ffffff;
}

.theme-preview--system {
  background: linear-gradient(135deg, #0f0f11 50%, #ffffff 50%);
}

.preview-toolbar {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 8px;
  height: 24px;
}

.preview-dot {
  width: 4px;
  height: 4px;
  border-radius: 50%;
}

.preview-dot-dark {
  background: #363641;
}

.preview-dot-light {
  background: #d4d4d8;
}

.preview-body {
  display: flex;
  height: calc(100% - 24px);
}

.preview-sidebar {
  width: 24px;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-subtle);
}

.preview-main {
  flex: 1;
  padding: 6px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.preview-line {
  height: 4px;
  border-radius: 2px;
  background: var(--bg-hover);
}

.preview-line--short {
  width: 60%;
}

/* Theme info */
.theme-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.theme-name {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.theme-desc {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.theme-check {
  position: absolute;
  top: var(--spacing-2);
  right: var(--spacing-2);
  color: var(--primary);
}

/* Theme info rows */
.theme-info-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-2) 0;
}

.info-label {
  font-size: var(--text-body);
  color: var(--text-secondary);
}

.info-value {
  font-size: var(--text-body);
  color: var(--text-primary);
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
}

.color-dot {
  display: inline-block;
  width: 10px;
  height: 10px;
  border-radius: 50%;
}
</style>