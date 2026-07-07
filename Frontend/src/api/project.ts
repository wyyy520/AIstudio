import type { Project, ProjectTemplate } from '@/types/project'
import { mockProjects, mockTemplates } from '@/mock/projects'

const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

export async function getProjects(): Promise<Project[]> {
  await delay(300)
  return [...mockProjects]
}

export async function getProjectById(id: string): Promise<Project | undefined> {
  await delay(200)
  return mockProjects.find(p => p.id === id)
}

export async function createProject(data: {
  name: string
  template: string
  framework: string
  plugins: string[]
}): Promise<Project> {
  await delay(500)
  const newProject: Project = {
    id: `proj-${Date.now()}`,
    name: data.name,
    type: 'custom',
    status: 'idle',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    description: '',
    template: data.template,
    framework: data.framework as Project['framework'],
    plugins: data.plugins,
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
  mockProjects.push(newProject)
  return newProject
}

export async function deleteProject(id: string): Promise<boolean> {
  await delay(300)
  const index = mockProjects.findIndex(p => p.id === id)
  if (index !== -1) {
    mockProjects.splice(index, 1)
    return true
  }
  return false
}

export async function getTemplates(): Promise<ProjectTemplate[]> {
  await delay(200)
  return [...mockTemplates]
}

export async function runWorkflow(projectId: string, workflowId: string): Promise<boolean> {
  await delay(1000)
  const project = mockProjects.find(p => p.id === projectId)
  if (project) {
    const workflow = project.workflows.find(w => w.id === workflowId)
    if (workflow) {
      workflow.status = 'running'
      setTimeout(() => {
        workflow.status = 'completed'
      }, 3000)
      return true
    }
  }
  return false
}

export async function trainModel(projectId: string, modelId: string): Promise<boolean> {
  await delay(1000)
  return true
}

export async function repairEnvironment(projectId: string): Promise<boolean> {
  await delay(2000)
  const project = mockProjects.find(p => p.id === projectId)
  if (project) {
    project.environment.gpuStatus = 'ready'
    project.environment.dependencies.forEach(dep => {
      dep.status = 'installed'
    })
  }
  return true
}
