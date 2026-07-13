import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Project, ProjectWorkflow, ProjectDataset, ProjectModel, ProjectExperiment } from '@/types/project'
import {
  getProjects,
  getProjectById,
  createProject as apiCreateProject,
  deleteProject as apiDeleteProject,
  updateProject,
  type ApiProject,
} from '@/api/project'
import { getWorkflows, runWorkflow as apiRunWorkflow } from '@/api/workflow'

export interface ProjectTemplate {
  id: string
  name: string
  description: string
  icon: string
  category: string
}

const mockTemplates: ProjectTemplate[] = [
  { id: 'vision', name: 'AI Vision', description: 'Image classification, detection, segmentation', icon: 'eye', category: 'vision' },
  { id: 'nlp', name: 'NLP', description: 'Text classification, NER, sentiment analysis', icon: 'message-square', category: 'nlp' },
  { id: 'timeseries', name: 'Time Series', description: 'Forecasting, anomaly detection', icon: 'trending-up', category: 'timeseries' },
  { id: 'custom', name: 'Custom', description: 'Empty project, build from scratch', icon: 'code', category: 'custom' },
]

export type ExplorerNodeType = 'dashboard' | 'workflows' | 'datasets' | 'models' | 'experiments' | 'environment' | 'outputs' | 'logs'

// Map from Backend ApiProject to Frontend Project type
function mapApiToProject(api: ApiProject): Project {
  return {
    id: String(api.id),
    name: api.name,
    type: 'custom',
    status: (api.status as Project['status']) || 'idle',
    createdAt: api.createdAt,
    updatedAt: api.updatedAt,
    description: api.description || '',
    template: '',
    framework: 'pytorch',
    plugins: [],
    workflows: [],
    datasets: [],
    models: [],
    experiments: [],
    environment: {
      pythonVersion: '3.10.12',
      cudaVersion: '12.1',
      pytorchVersion: '2.1.0',
      gpuStatus: 'ready',
      dependencies: [],
    },
    outputs: [],
    logs: [],
  }
}

export const useProjectStore = defineStore('appProject', () => {
  const projects = ref<Project[]>([])
  const currentProject = ref<Project | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const activeExplorerNode = ref<ExplorerNodeType>('dashboard')

  const sortedProjects = computed(() => {
    return [...projects.value].sort((a, b) =>
      new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
    )
  })

  const projectCount = computed(() => projects.value.length)

  const activeWorkflows = computed(() => {
    if (!currentProject.value) return []
    return currentProject.value.workflows.filter(w => w.status === 'running')
  })

  const completedExperiments = computed(() => {
    if (!currentProject.value) return []
    return currentProject.value.experiments.filter(e => e.status === 'completed')
  })

  async function fetchProjects() {
    loading.value = true
    error.value = null
    try {
      const apiProjects = await getProjects()
      projects.value = apiProjects.map(mapApiToProject)
    } catch (e) {
      error.value = e instanceof Error ? e.message : '获取项目列表失败'
      console.error('[project-store] fetchProjects failed:', e)
    } finally {
      loading.value = false
    }
  }

  async function fetchProjectById(id: string) {
    loading.value = true
    error.value = null
    try {
      const apiProject = await getProjectById(id)
      currentProject.value = mapApiToProject(apiProject)

      // Fetch associated workflows
      try {
        const apiWorkflows = await getWorkflows(id)
        currentProject.value.workflows = apiWorkflows.map(w => ({
          id: String(w.id),
          name: w.name,
          version: '1.0.0',
          nodeCount: 0,
          updatedAt: w.updatedAt,
          status: (w.status as ProjectWorkflow['status']) || 'idle',
        }))
      } catch {
        // workflows may not be available
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : '获取项目详情失败'
      console.error('[project-store] fetchProjectById failed:', e)
    } finally {
      loading.value = false
    }
  }

  async function createNewProject(data: {
    name: string
    template: string
    framework: string
    plugins: string[]
  }) {
    loading.value = true
    error.value = null
    try {
      const apiProject = await apiCreateProject({
        name: data.name,
        target: data.template || 'custom',
        description: `Template: ${data.template}, Framework: ${data.framework}`,
      })
      const project = mapApiToProject(apiProject)
      projects.value.push(project)
      currentProject.value = project
      return project
    } catch (e) {
      error.value = e instanceof Error ? e.message : '创建项目失败'
      console.error('[project-store] createNewProject failed:', e)
      return null
    } finally {
      loading.value = false
    }
  }

  async function removeProject(id: string) {
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
    } catch (e) {
      error.value = e instanceof Error ? e.message : '删除项目失败'
      console.error('[project-store] removeProject failed:', e)
      return false
    } finally {
      loading.value = false
    }
  }

  async function updateCurrentProject(updates: { name?: string; description?: string; status?: string }) {
    if (!currentProject.value) return false
    error.value = null
    try {
      await updateProject(currentProject.value.id, updates)
      if (updates.name) currentProject.value.name = updates.name
      if (updates.description) currentProject.value.description = updates.description
      if (updates.status) currentProject.value.status = updates.status as Project['status']
      return true
    } catch (e) {
      error.value = e instanceof Error ? e.message : '更新项目失败'
      return false
    }
  }

  function selectProject(project: Project) {
    currentProject.value = project
    activeExplorerNode.value = 'dashboard'
  }

  function setExplorerNode(node: ExplorerNodeType) {
    activeExplorerNode.value = node
  }

  function updateWorkflowStatus(workflowId: string, status: ProjectWorkflow['status']) {
    if (!currentProject.value) return
    const workflow = currentProject.value.workflows.find(w => w.id === workflowId)
    if (workflow) {
      workflow.status = status
    }
  }

  const templates = ref<ProjectTemplate[]>(mockTemplates)

  function fetchTemplates() {
    // Templates are static for now, can be fetched from API later
    return templates.value
  }

  async function runProjectWorkflow(workflowId: string) {
    if (!currentProject.value) return null
    try {
      const result = await apiRunWorkflow(workflowId)
      updateWorkflowStatus(workflowId, 'running')
      return result.id
    } catch (e) {
      error.value = e instanceof Error ? e.message : '运行工作流失败'
      return null
    }
  }

  // Backward compatibility aliases for views using simpler API
  async function addProject(data: any) {
    return createNewProject({
      name: data.name,
      template: 'custom',
      framework: 'pytorch',
      plugins: [],
    })
  }

  async function editProject(id: string, data: any) {
    return updateCurrentProject(data)
  }

  return {
    projects,
    currentProject,
    loading,
    error,
    activeExplorerNode,
    sortedProjects,
    projectCount,
    activeWorkflows,
    completedExperiments,
    templates,
    fetchProjects,
    fetchProjectById,
    createNewProject,
    removeProject,
    updateCurrentProject,
    selectProject,
    setExplorerNode,
    updateWorkflowStatus,
    fetchTemplates,
    runProjectWorkflow,
    addProject,
    editProject,
  }
})