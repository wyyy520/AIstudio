<template>
  <div class="dashboard-page">
    <div class="dashboard-grid">
      <div class="dashboard-card">
        <QuickStartCard @action="handleQuickAction" />
      </div>
      <div class="dashboard-card">
        <RecentProjects
          @open-project="openProject"
          @view-all="viewAllProjects"
          @context-menu="handleContextMenu"
        />
      </div>
      <div class="dashboard-card">
        <SystemStatus @check-environment="checkEnvironment" />
      </div>
      <div class="dashboard-card">
        <AnnouncementCard />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useProjectStore } from '@/store/project'
import QuickStartCard from './components/QuickStartCard.vue'
import RecentProjects from './components/RecentProjects.vue'
import SystemStatus from './components/SystemStatus.vue'
import AnnouncementCard from './components/AnnouncementCard.vue'

const router = useRouter()
const projectStore = useProjectStore()

onMounted(() => {
  projectStore.fetchProjects()
})

function handleQuickAction(key: string): void {
  switch (key) {
    case 'new-project':
      router.push('/projects')
      break
    case 'open-project':
      router.push('/projects')
      break
    case 'import-project':
      router.push('/projects')
      break
    case 'templates':
      router.push('/plugins')
      break
  }
}

function openProject(id: string): void {
  router.push(`/workflow/${id}`)
}

function viewAllProjects(): void {
  router.push('/projects')
}

function handleContextMenu(_event: MouseEvent, _project: unknown): void {
  // 右键菜单将在后续 ContextMenu 组件中实现
}

function checkEnvironment(): void {
  // 环境检查将在后续实现
}
</script>

<style scoped>
.dashboard-page {
  width: 100%;
  height: 100%;
  overflow-y: auto;
  padding: var(--spacing-8);
}

.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--spacing-4);
  max-width: 960px;
  margin: 0 auto;
}

.dashboard-card {
  background: var(--bg-tertiary);
  border-radius: var(--radius-xl);
  padding: 20px;
  transition: box-shadow var(--transition-normal);
}

.dashboard-card:hover {
  box-shadow: var(--shadow);
}

/* ===== 响应式 ===== */
@media (max-width: 1200px) {
  .dashboard-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 800px) {
  .dashboard-page {
    padding: var(--spacing-4);
  }

  .dashboard-grid {
    grid-template-columns: 1fr;
  }
}
</style>