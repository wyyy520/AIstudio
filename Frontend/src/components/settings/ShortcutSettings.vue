<template>
  <div class="settings-section">
    <h2 class="section-title">快捷键</h2>
    <p class="section-desc">查看和自定义键盘快捷键。点击快捷键可修改绑定。</p>

    <div class="settings-card">
      <div class="shortcut-list">
        <div
          v-for="item in shortcutItems"
          :key="item.action"
          class="shortcut-row"
        >
          <div class="shortcut-info">
            <span class="shortcut-label">{{ item.label }}</span>
            <span class="shortcut-desc">{{ item.desc }}</span>
          </div>
          <div class="shortcut-key-wrapper">
            <button
              class="shortcut-key"
              :class="{ 'is-editing': editingAction === item.action }"
              @click="startEdit(item.action)"
              @blur="finishEdit"
              @keydown="handleKeydown"
              tabindex="0"
            >
              <template v-if="editingAction === item.action">
                按下快捷键...
              </template>
              <template v-else>
                {{ store.shortcuts[item.action] || '未设置' }}
              </template>
            </button>
            <button
              class="shortcut-reset"
              title="重置"
              @click="store.resetShortcuts(); editingAction = null"
            >
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
                <path d="M3 3v5h5" />
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="settings-actions">
      <AppButton size="medium" type="outline" @click="store.resetShortcuts()">
        重置所有快捷键
      </AppButton>
      <AppButton size="medium" type="primary" @click="handleSave">
        保存快捷键
      </AppButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useSettingsStore } from '@/store/settings'
import AppButton from '@/components/AppButton/AppButton.vue'

const store = useSettingsStore()
const editingAction = ref<string | null>(null)

interface ShortcutItem {
  action: string
  label: string
  desc: string
}

const shortcutItems: ShortcutItem[] = [
  { action: 'search', label: '搜索', desc: '打开全局搜索面板' },
  { action: 'run-agent', label: '运行 Agent', desc: '在当前上下文执行 Agent' },
  { action: 'new-project', label: '新建项目', desc: '快速创建新项目' },
  { action: 'save', label: '保存', desc: '保存当前内容' },
  { action: 'toggle-sidebar', label: '切换侧边栏', desc: '显示或隐藏侧边栏' },
  { action: 'command-palette', label: '命令面板', desc: '打开命令面板' },
]

function startEdit(action: string) {
  editingAction.value = action
}

function finishEdit() {
  editingAction.value = null
}

function handleKeydown(e: KeyboardEvent) {
  if (!editingAction.value) return
  e.preventDefault()

  const parts: string[] = []
  if (e.ctrlKey || e.metaKey) parts.push('Ctrl')
  if (e.altKey) parts.push('Alt')
  if (e.shiftKey) parts.push('Shift')

  const key = e.key
  // Exclude modifier-only presses
  if (['Control', 'Alt', 'Shift', 'Meta'].includes(key)) return

  // Normalize key
  const normalizedKey = key.length === 1 ? key.toUpperCase() : key
  parts.push(normalizedKey)

  const shortcut = parts.join('+')
  store.shortcuts[editingAction.value] = shortcut
  editingAction.value = null
}

function handleSave() {
  store.saveShortcuts()
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
  overflow: hidden;
}

.shortcut-list {
  display: flex;
  flex-direction: column;
}

.shortcut-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-3) var(--spacing-4);
  border-bottom: 1px solid var(--border-subtle);
}

.shortcut-row:last-child {
  border-bottom: none;
}

.shortcut-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.shortcut-label {
  font-size: var(--text-body);
  color: var(--text-primary);
  line-height: var(--leading-body);
}

.shortcut-desc {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.shortcut-key-wrapper {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
}

.shortcut-key {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 100px;
  height: 28px;
  padding: 0 var(--spacing-3);
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-default);
  background: var(--bg-primary);
  color: var(--text-secondary);
  font-size: var(--text-body-sm);
  font-family: var(--font-family-mono);
  cursor: pointer;
  transition: all var(--transition-fast);
  user-select: none;
}

.shortcut-key:hover {
  border-color: var(--primary);
  color: var(--text-primary);
}

.shortcut-key.is-editing {
  border-color: var(--primary);
  color: var(--primary);
  box-shadow: 0 0 0 1px var(--primary);
  background: var(--primary-bg);
  animation: pulse 1s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

.shortcut-reset {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  color: var(--text-tertiary);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.shortcut-reset:hover {
  color: var(--text-primary);
  background: var(--bg-hover);
}

.settings-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}
</style>