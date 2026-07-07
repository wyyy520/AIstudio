<template>
  <header class="titlebar" data-tauri-drag-region>
    <div class="titlebar-left">
      <svg class="titlebar-icon" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <path d="M12 2L2 7l10 5 10-5-10-5z" />
        <path d="M2 17l10 5 10-5" />
        <path d="M2 12l10 5 10-5" />
      </svg>
      <span class="titlebar-title">AI Studio</span>
    </div>
    <div class="titlebar-center" data-tauri-drag-region />
    <div class="titlebar-right">
      <button class="titlebar-btn" title="最小化" @click="minimizeWindow">
        <svg viewBox="0 0 12 12" width="12" height="12">
          <rect x="1" y="5.5" width="10" height="1" fill="currentColor" />
        </svg>
      </button>
      <button class="titlebar-btn" title="最大化" @click="maximizeWindow">
        <svg viewBox="0 0 12 12" width="12" height="12">
          <rect x="1.5" y="1.5" width="9" height="9" rx="1" fill="none" stroke="currentColor" stroke-width="1" />
        </svg>
      </button>
      <button class="titlebar-btn titlebar-close" title="关闭" @click="closeWindow">
        <svg viewBox="0 0 12 12" width="12" height="12">
          <path d="M2 2l8 8M10 2l-8 8" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
        </svg>
      </button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { appWindow } from '@tauri-apps/api/window'

async function minimizeWindow() {
  try {
    await appWindow.minimize()
  } catch { }
}

async function maximizeWindow() {
  try {
    await appWindow.toggleMaximize()
  } catch { }
}

async function closeWindow() {
  try {
    await appWindow.close()
  } catch { }
}
</script>

<style scoped>
.titlebar {
  display: flex;
  align-items: center;
  height: var(--titlebar-height);
  padding: 0 var(--spacing-2);
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-subtle);
  user-select: none;
  flex-shrink: 0;
  position: relative;
  z-index: 1000;
}

.titlebar-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding-left: var(--spacing-2);
  min-width: 200px;
}

.titlebar-icon {
  color: var(--primary);
  flex-shrink: 0;
}

.titlebar-title {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-secondary);
  letter-spacing: -0.01em;
}

.titlebar-center {
  flex: 1;
  height: 100%;
}

.titlebar-right {
  display: flex;
  align-items: center;
  gap: 2px;
  padding-right: var(--spacing-1);
}

.titlebar-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 28px;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: background var(--transition-fast), color var(--transition-fast);
}

.titlebar-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.titlebar-btn:active {
  background: var(--bg-active);
}

.titlebar-close:hover {
  background: var(--error);
  color: white;
}

.titlebar-close:active {
  background: var(--error-bg);
}
</style>
