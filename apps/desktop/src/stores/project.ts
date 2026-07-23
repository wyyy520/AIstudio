import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ProjectSummary } from '@/api/project'
import {
  getProjects,
  getProjectById,
  createProject as apiCreateProject,
  updateProject as apiUpdateProject,
  deleteProject as apiDeleteProject,
  openProject as apiOpenProject,
  getRecentProjects,
  scanProjects,
  readWorkflow,
  saveWorkflow,
} from '@/api/project'

// ============================================================================
// Types
// ============================================================================

export type ExplorerNodeType =
  | 'dashboard'
  | 'workflows'
  | 'datasets'
  | 'models'
  | 'experiments'
  | 'environment'
  | 'outputs'
  | 'logs'

// ============================================================================
// Store
// ============================================================================

export const useProjectStore = defineStore('project', () => {
  // --- State ---
  const projects = ref<ProjectSummary[]>([])
  const currentProject = ref<ProjectSummary | null>(null)
  const recentProjects = ref<ProjectSummary[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const activeExplorerNode = ref<ExplorerNodeType>('dashboard')

  // --- Computed ---
  const sortedProjects = computed(() => {
    return [...projects.value].sort(
      (a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
    )
  })

  const projectCount = computed(() => projects.value.length)

  // --- Fetch ---

  async function fetchProjects() {
    loading.value = true
    error.value = null
    try {
      projects.value = await getProjects()
    } catch (e: any) {
      error.value = e.message || '获取项目列表失败'
      console.error('[project-store] fetchProjects:', e)
    } finally {
      loading.value = false
    }
  }

  async function fetchProjectById(id: string) {
    loading.value = true
    error.value = null
    try {
      currentProject.value = await getProjectById(id)
    } catch (e: any) {
      error.value = e.message || '获取项目详情失败'
      console.error('[project-store] fetchProjectById:', e)
    } finally {
      loading.value = false
    }
  }

  async function fetchRecentProjects() {
    try {
      recentProjects.value = await getRecentProjects()
    } catch (e) {
      console.error('[project-store] fetchRecentProjects:', e)
    }
  }

  // --- Create ---

  async function createNewProject(data: {
    name: string
    description?: string
    target?: string
  }): Promise<ProjectSummary | null> {
    loading.value = true
    error.value = null
    try {
      const project = await apiCreateProject(data)
      projects.value.push(project)
      currentProject.value = project
      return project
    } catch (e: any) {
      error.value = e.message || '创建项目失败'
      console.error('[project-store] createNewProject:', e)
      return null
    } finally {
      loading.value = false
    }
  }

  // --- Open (real folder) ---

  async function openProject(path: string): Promise<ProjectSummary | null> {
    loading.value = true
    error.value = null
    try {
      const project = await apiOpenProject(path)
      // Add to local list if not already there
      const exists = projects.value.find(p => p.id === project.id)
      if (!exists) {
        projects.value.push(project)
      } else {
        Object.assign(exists, project)
      }
      currentProject.value = project
      return project
    } catch (e: any) {
      error.value = e.message || '打开项目失败'
      console.error('[project-store] openProject:', e)
      return null
    } finally {
      loading.value = false
    }
  }

  // --- Update ---

  async function updateCurrentProject(updates: {
    name?: string
    description?: string
    target?: string
    status?: string
  }): Promise<boolean> {
    if (!currentProject.value) return false
    error.value = null
    try {
      const updated = await apiUpdateProject(currentProject.value.id, updates)
      Object.assign(currentProject.value, updated)
      return true
    } catch (e: any) {
      error.value = e.message || '更新项目失败'
      return false
    }
  }

  // --- Delete ---

  async function removeProject(id: string): Promise<boolean> {
    loading.value = true
    error.value = null
    try {
      await apiDeleteProject(id)
      projects.value = projects.value.filter(p => p.id !== id)
      if (currentProject.value?.id === id) {
        currentProject.value = null
        activeExplorerNode.value = 'dashboard'
      }
      return true
    } catch (e: any) {
      error.value = e.message || '删除项目失败'
      return false
    } finally {
      loading.value = false
    }
  }

  // --- Selection ---

  function selectProject(project: ProjectSummary) {
    currentProject.value = project
    activeExplorerNode.value = 'dashboard'
  }

  function setExplorerNode(node: ExplorerNodeType) {
    activeExplorerNode.value = node
  }

  // --- Workflow I/O ---

  async function loadWorkflow(projectId: string): Promise<any> {
    try {
      return await readWorkflow(projectId)
    } catch (e: any) {
      error.value = e.message || '读取工作流失败'
      return null
    }
  }

  async function persistWorkflow(projectId: string, data: any): Promise<boolean> {
    try {
      await saveWorkflow(projectId, data)
      return true
    } catch (e: any) {
      error.value = e.message || '保存工作流失败'
      return false
    }
  }

  // --- Scan ---

  async function rescanProjects(): Promise<boolean> {
    error.value = null
    try {
      projects.value = await scanProjects()
      return true
    } catch (e: any) {
      error.value = e.message || '扫描项目失败'
      return false
    }
  }

  return {
    // State
    projects,
    currentProject,
    recentProjects,
    loading,
    error,
    activeExplorerNode,
    // Computed
    sortedProjects,
    projectCount,
    // Actions
    fetchProjects,
    fetchProjectById,
    fetchRecentProjects,
    createNewProject,
    openProject,
    updateCurrentProject,
    removeProject,
    selectProject,
    setExplorerNode,
    loadWorkflow,
    persistWorkflow,
    rescanProjects,
  }
})
