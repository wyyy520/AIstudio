import type { WorkflowSchema } from '../schema'
import { useWorkflowStore } from '../store'
import {
  createProject as apiCreateProject,
  openProject as apiOpenProject,
  readWorkflow,
  saveWorkflow as apiSaveWorkflow,
} from '@/api/project'

export interface ProjectInfo {
  id: string
  path: string
  name: string
}

class ProjectManager {
  /**
   * Create a new AIStudio project via the backend API.
   * The backend creates the real directory structure on the filesystem.
   */
  async newProject(name: string, description = ''): Promise<ProjectInfo | null> {
    try {
      const project = await apiCreateProject({
        name,
        description,
        target: 'python',
      })

      const store = useWorkflowStore()
      store.project = {
        id: project.id,
        path: project.rootPath,
        name: project.name,
      }
      store.resetWorkflow(project.name)

      return {
        id: project.id,
        path: project.rootPath,
        name: project.name,
      }
    } catch (err) {
      console.error('[project-manager] newProject failed:', err)
      return null
    }
  }

  /**
   * Open a project directory by its backend project ID.
   * Loads the workflow.json from disk via the API.
   */
  async openProjectById(projectId: string): Promise<ProjectInfo | null> {
    try {
      const wf = await readWorkflow(projectId) as WorkflowSchema
      const store = useWorkflowStore()
      store.loadWorkflowJSON(wf)

      // Store project info — project path is the root containing workflow.json
      store.project = {
        id: projectId,
        path: '',
        name: wf.name || 'Untitled',
      }

      return {
        id: projectId,
        path: store.project.path,
        name: store.project.name,
      }
    } catch (err) {
      console.error('[project-manager] openProjectById failed:', err)
      return null
    }
  }

  /**
   * Open a project from the filesystem by opening a real folder.
   * Uses the backend API to register the directory as a project.
   */
  async openProjectAt(path: string): Promise<ProjectInfo | null> {
    try {
      const project = await apiOpenProject(path)
      const store = useWorkflowStore()

      // Load workflow
      const wfContent = await readWorkflow(project.id) as WorkflowSchema
      store.loadWorkflowJSON(wfContent)

      store.project = {
        id: project.id,
        path: project.rootPath,
        name: project.name,
      }

      return {
        id: project.id,
        path: project.rootPath,
        name: project.name,
      }
    } catch (err) {
      console.error('[project-manager] openProjectAt failed:', err)
      return null
    }
  }

  /**
   * Save the current workflow via the backend API,
   * which writes workflow.json to the project directory.
   */
  async saveWorkflow(projectId: string): Promise<boolean> {
    try {
      const store = useWorkflowStore()
      const wf = store.getWorkflowJSON()
      await apiSaveWorkflow(projectId, wf)
      store.isDirty = false
      store.lastSaved = wf.updated_at || new Date().toISOString()
      return true
    } catch (err) {
      console.error('[project-manager] saveWorkflow failed:', err)
      return false
    }
  }
}

export const projectManager = new ProjectManager()
