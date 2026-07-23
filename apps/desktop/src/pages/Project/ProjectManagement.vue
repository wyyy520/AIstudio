<template>
  <div class="project-management">
    <ProjectToolbar
      :active-view="activeView"
      @view-change="handleViewChange"
      @new-project="showCreateDialog = true"
    />

    <div class="management-layout">
      <div class="layout-left">
        <div class="explorer-panel">
          <div class="explorer-header">
            <span class="explorer-title">Projects</span>
            <div class="explorer-actions">
              <button class="icon-btn" title="Scan projects directory" @click="handleScan">
                <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.2"/>
                </svg>
              </button>
            </div>
          </div>

          <div v-if="loading" class="explorer-loading">
            <span class="loading-spinner"></span>
          </div>

          <div v-else-if="projects.length === 0" class="explorer-empty">
            <p>No projects yet</p>
            <p class="empty-hint">Create a new project or open an existing folder</p>
          </div>

          <div v-else class="project-list">
            <div
              v-for="project in projects"
              :key="project.id"
              class="project-list-item"
              :class="{ active: currentProject?.id === project.id }"
              @click="handleSelectProject(project)"
              @dblclick="handleOpenWorkflow(project)"
            >
              <div class="item-info">
                <span class="item-name">{{ project.name }}</span>
                <span class="item-path" :title="project.rootPath">{{ project.rootPath }}</span>
              </div>
              <div class="item-meta">
                <span class="item-target">{{ project.target || 'python' }}</span>
                <span class="status-dot" :class="`status-${project.status}`"></span>
              </div>
            </div>
          </div>
        </div>

        <!-- Recent projects -->
        <div v-if="recentProjects.length > 0" class="recent-section">
          <div class="section-header">
            <span class="section-title">Recent</span>
          </div>
          <div class="recent-list">
            <div
              v-for="project in recentProjects.slice(0, 5)"
              :key="project.id"
              class="recent-item"
              @click="handleSelectProject(project)"
            >
              <span class="recent-name">{{ project.name }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="layout-center">
        <div v-if="!currentProject" class="welcome-state">
          <div class="welcome-content">
            <svg viewBox="0 0 24 24" width="64" height="64" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" />
            </svg>
            <h2>AIStudio Project Workspace</h2>
            <p>Select a project to start working, or create a new one</p>
            <div class="welcome-actions">
              <button class="welcome-btn primary" @click="showCreateDialog = true">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M12 5v14M5 12h14" />
                </svg>
                New Project
              </button>
              <button class="welcome-btn secondary" @click="handleOpenFolder">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" />
                </svg>
                Open Folder
              </button>
            </div>
          </div>
        </div>

        <template v-else>
          <ProjectDashboard
            :project="currentProject"
            @run-workflow="handleRunWorkflow"
            @open-workflow="handleOpenWorkflow(currentProject)"
          />
        </template>
      </div>
    </div>

    <div class="management-statusbar">
      <div class="statusbar-left">
        <span v-if="currentProject" class="statusbar-item">{{ currentProject.name }}</span>
        <span v-else class="statusbar-item muted">No project selected</span>
      </div>
      <div class="statusbar-right">
        <span class="statusbar-item">v0.1.0</span>
      </div>
    </div>

    <CreateProjectDialog
      v-model:visible="showCreateDialog"
      @create="handleCreateProject"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useProjectStore } from '@/stores/project'
import type { ProjectSummary } from '@/api/project'
import { openProject as apiOpenProject } from '@/api/project'
import ProjectToolbar from '@/components/project/ProjectToolbar.vue'
import ProjectDashboard from '@/components/project/ProjectDashboard.vue'
import CreateProjectDialog from '@/components/project/CreateProjectDialog.vue'

const store = useProjectStore()
const router = useRouter()

const activeView = ref('dashboard')
const showCreateDialog = ref(false)

const projects = computed(() => store.sortedProjects)
const currentProject = computed(() => store.currentProject)
const recentProjects = computed(() => store.recentProjects)
const loading = computed(() => store.loading)

onMounted(async () => {
  await Promise.all([
    store.fetchProjects(),
    store.fetchRecentProjects(),
  ])
})

function handleViewChange(view: string) {
  activeView.value = view
}

async function handleSelectProject(project: ProjectSummary) {
  store.selectProject(project)
}

async function handleOpenWorkflow(project: ProjectSummary) {
  store.selectProject(project)
  // Navigate to the workflow editor
  router.push(`/project/${project.id}/workflow`)
}

async function handleCreateProject(data: { name: string; description?: string; target?: string }) {
  const project = await store.createNewProject(data)
  if (project) {
    showCreateDialog.value = false
    router.push(`/project/${project.id}/workflow`)
  }
}

async function handleOpenFolder() {
  // Prompt for a directory path (in browser mode, fall back to prompt)
  let path = ''
  try {
    path = prompt('Enter the absolute path to your project folder:') || ''
  } catch {
    return
  }
  if (!path.trim()) return

  const project = await store.openProject(path.trim())
  if (project) {
    router.push(`/project/${project.id}/workflow`)
  }
}

async function handleScan() {
  await store.rescanProjects()
}

function handleRunWorkflow() {
  if (!currentProject.value) return
  router.push(`/project/${currentProject.value.id}/workflow`)
}
</script>

<style scoped>
.project-management {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.management-layout {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.layout-left {
  width: 300px;
  flex-shrink: 0;
  border-right: 1px solid var(--border-subtle);
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.explorer-panel {
  padding: var(--spacing-3);
}

.explorer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-3);
}

.explorer-title {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.explorer-actions {
  display: flex;
  gap: var(--spacing-1);
}

.icon-btn {
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
}

.icon-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.explorer-loading {
  padding: var(--spacing-6);
  text-align: center;
  color: var(--text-tertiary);
}

.loading-spinner {
  display: inline-block;
  width: 20px;
  height: 20px;
  border: 2px solid var(--border-subtle);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.explorer-empty {
  padding: var(--spacing-6);
  text-align: center;
  color: var(--text-tertiary);
}

.explorer-empty .empty-hint {
  font-size: var(--text-caption);
  margin-top: var(--spacing-2);
  color: var(--text-disabled);
}

.project-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.project-list-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-2) var(--spacing-3);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.project-list-item:hover {
  background: var(--bg-hover);
}

.project-list-item.active {
  background: var(--primary-bg);
}

.item-info {
  flex: 1;
  min-width: 0;
}

.item-name {
  display: block;
  font-size: var(--text-body-sm);
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: var(--font-medium);
}

.item-path {
  display: block;
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 220px;
}

.item-meta {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  flex-shrink: 0;
}

.item-target {
  font-size: var(--text-caption);
  padding: 1px 6px;
  border-radius: var(--radius-sm);
  background: var(--bg-tertiary);
  color: var(--text-tertiary);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-dot.status-active { background: var(--success); }
.status-dot.status-archived { background: var(--neutral); }

.recent-section {
  border-top: 1px solid var(--border-subtle);
  padding: var(--spacing-3);
}

.section-header {
  margin-bottom: var(--spacing-2);
}

.section-title {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.recent-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.recent-item {
  padding: var(--spacing-1) var(--spacing-2);
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: var(--text-body-sm);
  color: var(--text-secondary);
}

.recent-item:hover {
  background: var(--bg-hover);
}

.layout-center {
  flex: 1;
  overflow: hidden;
  background: var(--bg-primary);
}

.welcome-state {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.welcome-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-4);
  text-align: center;
  color: var(--text-tertiary);
}

.welcome-content svg {
  color: var(--text-disabled);
}

.welcome-content h2 {
  font-size: var(--text-h2);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.welcome-content p {
  font-size: var(--text-body);
  color: var(--text-secondary);
}

.welcome-actions {
  display: flex;
  gap: var(--spacing-3);
}

.welcome-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  height: 44px;
  padding: 0 var(--spacing-6);
  border-radius: var(--radius-lg);
  border: none;
  font-size: var(--text-body);
  font-weight: var(--font-semibold);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.welcome-btn.primary {
  background: var(--primary);
  color: white;
}

.welcome-btn.primary:hover {
  background: var(--primary-hover);
  transform: translateY(-1px);
  box-shadow: var(--shadow-lg);
}

.welcome-btn.secondary {
  background: var(--bg-secondary);
  color: var(--text-primary);
  border: 1px solid var(--border-subtle);
}

.welcome-btn.secondary:hover {
  background: var(--bg-hover);
  transform: translateY(-1px);
}

.management-statusbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: var(--statusbar-height);
  padding: 0 var(--spacing-3);
  background: var(--bg-secondary);
  border-top: 1px solid var(--border-subtle);
}

.statusbar-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.statusbar-item {
  font-size: var(--text-caption);
  color: var(--text-primary);
}

.statusbar-item.muted {
  color: var(--text-tertiary);
}

.statusbar-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}
</style>
