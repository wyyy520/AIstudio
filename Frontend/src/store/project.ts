import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Project, ProjectTemplate, ProjectWorkflow, ProjectDataset, ProjectModel, ProjectExperiment } from '@/types/project'
import {
  getProjects,
  getProjectById,
  createProject as apiCreateProject,
  deleteProject as apiDeleteProject,
  getTemplates,
  runWorkflow as apiRunWorkflow,
  repairEnvironment as apiRepairEnvironment,
} from '@/api/project'

export type ExplorerNodeType = 'dashboard' | 'workflows' | 'datasets' | 'models' | 'experiments' | 'environment' | 'outputs' | 'logs'

export const useProjectStore = defineStore('project', () => {
  const projects = ref<Project[]>([])
  const currentProject = ref<Project | null>(null)
  const templates = ref<ProjectTemplate[]>([])
  const loading = ref(false)
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
    try {
      projects.value = await getProjects()
    } finally {
      loading.value = false
    }
  }

  async function fetchProjectById(id: string) {
    loading.value = true
    try {
      const project = await getProjectById(id)
      currentProject.value = project || null
    } finally {
      loading.value = false
    }
  }

  async function fetchTemplates() {
    templates.value = await getTemplates()
  }

  async function createNewProject(data: {
    name: string
    template: string
    framework: string
    plugins: string[]
  }) {
    loading.value = true
    try {
      const project = await apiCreateProject(data)
      projects.value.push(project)
      currentProject.value = project
      return project
    } finally {
      loading.value = false
    }
  }

  async function removeProject(id: string) {
    loading.value = true
    try {
      const success = await apiDeleteProject(id)
      if (success) {
        projects.value = projects.value.filter(p => p.id !== id)
        if (currentProject.value?.id === id) {
          currentProject.value = null
          activeExplorerNode.value = 'dashboard'
        }
      }
      return success
    } finally {
      loading.value = false
    }
  }

  function selectProject(project: Project) {
    currentProject.value = project
    activeExplorerNode.value = 'dashboard'
  }

  function setExplorerNode(node: ExplorerNodeType) {
    activeExplorerNode.value = node
  }

  async function runProjectWorkflow(workflowId: string) {
    if (!currentProject.value) return false
    return apiRunWorkflow(currentProject.value.id, workflowId)
  }

  async function repairProjectEnvironment() {
    if (!currentProject.value) return false
    return apiRepairEnvironment(currentProject.value.id)
  }

  function updateWorkflowStatus(workflowId: string, status: ProjectWorkflow['status']) {
    if (!currentProject.value) return
    const workflow = currentProject.value.workflows.find(w => w.id === workflowId)
    if (workflow) {
      workflow.status = status
    }
  }

  return {
    projects,
    currentProject,
    templates,
    loading,
    activeExplorerNode,
    sortedProjects,
    projectCount,
    activeWorkflows,
    completedExperiments,
    fetchProjects,
    fetchProjectById,
    fetchTemplates,
    createNewProject,
    removeProject,
    selectProject,
    setExplorerNode,
    runProjectWorkflow,
    repairProjectEnvironment,
    updateWorkflowStatus,
  }
})
