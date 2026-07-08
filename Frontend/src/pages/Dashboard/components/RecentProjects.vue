<template>
  <div class="recent-projects-card">
    <div class="card-header">
      <h3 class="card-title">最近编辑</h3>
      <button class="card-link" @click="viewAll">查看全部</button>
    </div>
    <div class="project-list">
      <div
        v-for="project in projects"
        :key="project.id"
        class="project-item"
        @click="openProject(project.id)"
        @contextmenu.prevent="handleContextMenu($event, project)"
      >
        <svg
          class="project-icon"
          viewBox="0 0 24 24"
          width="20"
          height="20"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path :d="project.icon" />
        </svg>
        <div class="project-info">
          <span class="project-name">{{ project.name }}</span>
          <span class="project-time">{{ project.time }}</span>
        </div>
        <button
          class="project-more"
          type="button"
          @click.stop="handleContextMenu($event, project)"
          title="更多操作"
        >
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="12" cy="12" r="1" fill="currentColor" />
            <circle cx="12" cy="5" r="1" fill="currentColor" />
            <circle cx="12" cy="19" r="1" fill="currentColor" />
          </svg>
        </button>
      </div>

      <div v-if="projects.length === 0" class="project-empty">
        <span class="empty-text">暂无最近编辑的项目</span>
        <span class="empty-hint">创建第一个项目开始使用</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useProjectStore } from '@/store/project'

interface RecentProject {
  id: string
  name: string
  time: string
  icon: string
}

const projectStore = useProjectStore()

const projects = computed<RecentProject[]>(() => {
  return projectStore.sortedProjects.slice(0, 5).map(p => ({
    id: p.id,
    name: p.name,
    time: formatTime(p.updatedAt),
    icon: 'M6 3h3v6H6V3zm0 12h3v6H6v-6zm9-12h3v6h-3V3zm0 12h3v6h-3v-6zm-9 0V9m3 9v-3m3 3v-3m3 3V9',
  }))
})

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  if (days === 0) return '今天'
  if (days === 1) return '昨天'
  if (days < 7) return `${days} 天前`
  return date.toLocaleDateString('zh-CN')
}

const emit = defineEmits<{
  'open-project': [id: string]
  'view-all': []
  'context-menu': [event: MouseEvent, project: RecentProject]
}>()

function openProject(id: string): void {
  emit('open-project', id)
}

function viewAll(): void {
  emit('view-all')
}

function handleContextMenu(event: MouseEvent, project: RecentProject): void {
  emit('context-menu', event, project)
}
</script>

<style scoped>
.recent-projects-card {
  display: flex;
  flex-direction: column;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-4);
}

.card-title {
  font-size: var(--text-h3);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  line-height: var(--leading-h3);
}

.card-link {
  border: none;
  background: transparent;
  color: var(--primary);
  font-size: var(--text-body-sm);
  font-family: var(--font-family-sans);
  cursor: pointer;
  padding: 0;
  transition: color var(--transition-fast);
}

.card-link:hover {
  color: var(--primary-hover);
}

/* ===== 列表 ===== */
.project-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
}

.project-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  height: 48px;
  padding: 0 var(--spacing-3);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: background var(--transition-fast);
}

.project-item:hover {
  background: var(--bg-hover);
}

.project-item:active {
  background: var(--bg-active);
  transform: scale(0.99);
}

.project-icon {
  flex-shrink: 0;
  color: var(--text-tertiary);
  opacity: 0.7;
}

.project-item:hover .project-icon {
  opacity: 1;
  color: var(--primary);
}

.project-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.project-name {
  font-size: var(--text-body-sm);
  color: var(--text-primary);
  line-height: var(--leading-body-sm);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.project-time {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  line-height: var(--leading-caption);
}

.project-more {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-tertiary);
  cursor: pointer;
  opacity: 0;
  transition: all var(--transition-fast);
  flex-shrink: 0;
}

.project-item:hover .project-more {
  opacity: 1;
}

.project-more:hover {
  background: var(--bg-active);
  color: var(--text-primary);
}

/* ===== 空状态 ===== */
.project-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-1);
  flex: 1;
  padding: var(--spacing-8) 0;
}

.empty-text {
  font-size: var(--text-body-sm);
  color: var(--text-secondary);
}

.empty-hint {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}
</style>