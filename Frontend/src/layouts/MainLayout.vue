<template>
  <div class="main-layout">
    <!-- ===== 自定义标题栏 (窗口拖拽) ===== -->
    <TitleBar />

    <!-- ===== 左侧导航栏 ===== -->
    <aside class="sidebar">
      <!-- 顶部: Logo 区域 -->
      <div class="sidebar-header">
        <div class="sidebar-logo">
          <svg
            class="sidebar-logo-icon"
            viewBox="0 0 24 24"
            width="22"
            height="22"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M12 2L2 7l10 5 10-5-10-5z" />
            <path d="M2 17l10 5 10-5" />
            <path d="M2 12l10 5 10-5" />
          </svg>
          <span class="sidebar-logo-text">AIStudio</span>
        </div>
      </div>

      <!-- 中间: 导航菜单 -->
      <nav class="sidebar-nav">
        <div
          v-for="item in navItems"
          :key="item.path"
          class="sidebar-nav-item"
          :class="{ active: isActive(item.path) }"
          :title="item.label"
          @click="navigateTo(item.path)"
        >
          <svg
            class="sidebar-nav-icon"
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
          <span class="sidebar-nav-label">{{ item.label }}</span>
        </div>
      </nav>

      <!-- 底部: 设置/用户区域 -->
      <div class="sidebar-footer">
        <div
          class="sidebar-nav-item"
          :class="{ active: isActive('/settings') }"
          title="设置"
          @click="navigateTo('/settings')"
        >
          <svg
            class="sidebar-nav-icon"
            viewBox="0 0 24 24"
            width="18"
            height="18"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <circle cx="12" cy="12" r="3" />
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
          </svg>
          <span class="sidebar-nav-label">设置</span>
        </div>
      </div>
    </aside>

    <!-- ===== 主内容区域 ===== -->
    <div class="main-content">
      <router-view />
    </div>

    <!-- ===== 底部状态栏 ===== -->
    <footer class="statusbar">
      <div class="statusbar-left">
        <span class="statusbar-item">就绪</span>
      </div>
      <div class="statusbar-right">
        <span class="statusbar-item">AIStudio v0.1.0</span>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { useRouter, useRoute } from 'vue-router'
import TitleBar from '@/components/titlebar/TitleBar.vue'

interface NavItem {
  path: string
  label: string
  icon: string
}

const router = useRouter()
const route = useRoute()

const navItems: NavItem[] = [
  {
    path: '/',
    label: '仪表盘',
    // Lucide: layout-dashboard
    icon: 'M3 3h7v7H3V3zm0 11h7v7H3v-7zm11-11h7v7h-7V3zm0 11h7v7h-7v-7z',
  },
  {
    path: '/workflow',
    label: '工作流',
    // Lucide: workflow
    icon: 'M6 3h3v6H6V3zm0 12h3v6H6v-6zm9-12h3v6h-3V3zm0 12h3v6h-3v-6zm-9 0V9m3 9v-3m3 3v-3m3 3V9',
  },
  {
    path: '/chat',
    label: 'AI 对话',
    // Lucide: message-square
    icon: 'M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z',
  },
  {
    path: '/plugins',
    label: '插件市场',
    // Lucide: blocks
    icon: 'M3 3h7v7H3V3zm0 11h7v7H3v-7zm11-11h7v7h-7V3zm0 11h7v7h-7v-7z',
  },
  {
    path: '/projects',
    label: '项目管理',
    // Lucide: folder-open
    icon: 'M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z',
  },
  {
    path: '/logs',
    label: '日志',
    // Lucide: scroll-text
    icon: 'M8 21h12a2 2 0 0 0 2-2v-2H10v2a2 2 0 1 1-4 0V5a2 2 0 1 1 4 0v2h12v2H8a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2z',
  },
]

function isActive(path: string): boolean {
  if (path === '/') {
    return route.path === '/'
  }
  return route.path.startsWith(path)
}

function navigateTo(path: string): void {
  router.push(path)
}
</script>

<style scoped>
.main-layout {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

/* ===== 侧边栏 ===== */
.sidebar {
  position: fixed;
  top: var(--titlebar-height);
  left: 0;
  bottom: var(--statusbar-height);
  width: var(--sidebar-width);
  display: flex;
  flex-direction: column;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-subtle);
  z-index: 100;
  overflow: hidden;
}

/* ===== 侧边栏头部 ===== */
.sidebar-header {
  padding: var(--spacing-4) var(--spacing-4) var(--spacing-3);
  border-bottom: 1px solid var(--border-subtle);
}

.sidebar-logo {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.sidebar-logo-icon {
  color: var(--primary);
  flex-shrink: 0;
}

.sidebar-logo-text {
  font-size: var(--text-h3);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

/* ===== 导航列表 ===== */
.sidebar-nav {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-2) var(--spacing-2);
}

.sidebar-nav-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  height: 36px;
  padding: 0 var(--spacing-3);
  margin-bottom: 2px;
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  cursor: pointer;
  transition: background var(--transition-fast), color var(--transition-fast);
  position: relative;
}

.sidebar-nav-item:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.sidebar-nav-item:active {
  transform: scale(0.98);
}

.sidebar-nav-item.active {
  background: var(--bg-active);
  color: var(--text-primary);
}

.sidebar-nav-item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 18px;
  border-radius: 0 2px 2px 0;
  background: var(--primary);
}

.sidebar-nav-icon {
  flex-shrink: 0;
  opacity: 0.8;
}

.sidebar-nav-item.active .sidebar-nav-icon {
  opacity: 1;
  color: var(--primary);
}

.sidebar-nav-label {
  font-size: var(--text-body-sm);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* ===== 侧边栏底部 ===== */
.sidebar-footer {
  padding: var(--spacing-2);
  border-top: 1px solid var(--border-subtle);
}

/* ===== 主内容区域 ===== */
.main-content {
  margin-top: var(--titlebar-height);
  margin-left: var(--sidebar-width);
  margin-bottom: var(--statusbar-height);
  flex: 1;
  overflow: hidden;
  background: var(--bg-primary);
}

/* ===== 底部状态栏 ===== */
.statusbar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: var(--statusbar-height);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--spacing-3);
  background: var(--bg-secondary);
  border-top: 1px solid var(--border-subtle);
  z-index: 200;
}

.statusbar-left,
.statusbar-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.statusbar-item {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  line-height: var(--statusbar-height);
  cursor: default;
}
</style>