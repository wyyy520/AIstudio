<template>
  <div class="settings-page">
    <!-- 左侧菜单 -->
    <aside class="settings-sidebar">
      <div class="settings-sidebar-header">
        <h2 class="settings-sidebar-title">设置</h2>
      </div>
      <nav class="settings-sidebar-nav">
        <button
          v-for="item in menuItems"
          :key="item.key"
          class="settings-nav-item"
          :class="{ active: activeTab === item.key }"
          @click="activeTab = item.key"
        >
          <svg
            class="settings-nav-icon"
            viewBox="0 0 24 24"
            width="18"
            height="18"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path :d="item.icon" />
          </svg>
          <span class="settings-nav-label">{{ item.label }}</span>
        </button>
      </nav>
    </aside>

    <!-- 右侧内容 -->
    <div class="settings-content">
      <transition name="fade" mode="out-in">
        <GeneralSettings v-if="activeTab === 'general'" key="general" />
        <EngineConfig v-else-if="activeTab === 'engine'" key="engine" />
        <PluginManager v-else-if="activeTab === 'plugins'" key="plugins" />
        <ShortcutSettings v-else-if="activeTab === 'shortcuts'" key="shortcuts" />
        <ThemeSettings v-else-if="activeTab === 'theme'" key="theme" />
      </transition>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import GeneralSettings from '@/components/settings/GeneralSettings.vue'
import EngineConfig from '@/components/settings/EngineConfig.vue'
import PluginManager from '@/components/settings/PluginManager.vue'
import ShortcutSettings from '@/components/settings/ShortcutSettings.vue'
import ThemeSettings from '@/components/settings/ThemeSettings.vue'

const store = useSettingsStore()
const activeTab = ref('general')

interface MenuItem {
  key: string
  label: string
  icon: string
}

const menuItems: MenuItem[] = [
  {
    key: 'general',
    label: '通用设置',
    // Lucide: settings
    icon: 'M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z',
  },
  {
    key: 'engine',
    label: 'Engine 配置',
    // Lucide: cpu
    icon: 'M9 2v4M15 2v4M9 18v4M15 18v4M5 8h14M5 16h14M3 10h2M3 14h2M19 10h2M19 14h2M7 8v8M17 8v8',
  },
  {
    key: 'plugins',
    label: '插件管理',
    // Lucide: puzzle
    icon: 'M19.439 15.39c-.537.3-1.116.473-1.689.526a5.066 5.066 0 0 1-1.609-.21 5.04 5.04 0 0 1-1.639-.91 5.04 5.04 0 0 1-1.148-1.44 5.066 5.066 0 0 1-.21-1.609 5.04 5.04 0 0 1 .91-1.639 5.04 5.04 0 0 1 1.44-1.148 5.066 5.066 0 0 1 1.609-.21c.574.053 1.152.226 1.69.526l1.661-1.661a7.003 7.003 0 0 0-3.338-1.7 7.01 7.01 0 0 0-3.724.292 7.004 7.004 0 0 0-3.04 2.148 7.004 7.004 0 0 0-1.148 1.44 7.01 7.01 0 0 0 1.519 8.764 7.01 7.01 0 0 0 3.04 1.424 7.004 7.004 0 0 0 3.724-.292 7.003 7.003 0 0 0 3.038-2.148z',
  },
  {
    key: 'shortcuts',
    label: '快捷键',
    // Lucide: keyboard
    icon: 'M2 7v10a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2H4a2 2 0 0 0-2 2zm4 6h12M6 9h2M10 9h2M14 9h2M18 9h2M6 13h2M10 13h2M14 13h2M18 13h2',
  },
  {
    key: 'theme',
    label: '主题设置',
    // Lucide: palette
    icon: 'M12 2C6.49 2 2 6.49 2 12s4.49 10 10 10c1.38 0 2.5-1.12 2.5-2.5 0-.61-.23-1.16-.6-1.58a2.5 2.5 0 0 1-.4-1.42c0-1.38 1.12-2.5 2.5-2.5H16c3.31 0 6-2.69 6-6 0-4.96-4.49-9-10-9zM7.5 12a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3zm3-4a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3zm5 0a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3zm3 4a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3z',
  },
]

onMounted(() => {
  store.loadShortcuts()
})
</script>

<style scoped>
.settings-page {
  display: flex;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

/* ===== 左侧菜单 ===== */
.settings-sidebar {
  width: 200px;
  min-width: 200px;
  display: flex;
  flex-direction: column;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-subtle);
  overflow-y: auto;
}

.settings-sidebar-header {
  padding: var(--spacing-4) var(--spacing-4) var(--spacing-3);
  border-bottom: 1px solid var(--border-subtle);
}

.settings-sidebar-title {
  font-size: var(--text-h3);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  line-height: var(--leading-h3);
}

.settings-sidebar-nav {
  flex: 1;
  padding: var(--spacing-2);
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.settings-nav-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  height: 36px;
  padding: 0 var(--spacing-3);
  border: none;
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
  font-family: var(--font-family-sans);
  font-size: var(--text-body-sm);
  text-align: left;
  width: 100%;
}

.settings-nav-item:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.settings-nav-item.active {
  background: var(--bg-active);
  color: var(--text-primary);
}

.settings-nav-item.active .settings-nav-icon {
  color: var(--primary);
}

.settings-nav-icon {
  flex-shrink: 0;
  opacity: 0.8;
}

.settings-nav-item.active .settings-nav-icon {
  opacity: 1;
}

.settings-nav-label {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* ===== 右侧内容 ===== */
.settings-content {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-8);
  padding-top: var(--spacing-6);
  max-width: 720px;
}

/* ===== 过渡动画 ===== */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>