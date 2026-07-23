import request from './request'

// ============================================================================
// Types — matching the backend project.ProjectSummary
// ============================================================================

export interface ProjectSummary {
  id: string
  name: string
  description?: string
  target?: string
  status: string
  rootPath: string
  version: number
  fileCount: number
  createdAt: string
  updatedAt: string
}

export interface ApiResponse<T = any> {
  code: number
  message: string
  data?: T
}

// ============================================================================
// Helpers
// ============================================================================

function unwrapData<T>(res: ApiResponse<T>): T {
  if (res.code !== 0) {
    throw new Error(res.message || 'Request failed')
  }
  return res.data as T
}

// ============================================================================
// Project CRUD
// ============================================================================

/** List all projects. Returns array directly. */
export async function getProjects(): Promise<ProjectSummary[]> {
  const res = await request.get('/projects') as ApiResponse<ProjectSummary[]>
  return unwrapData(res)
}

/** Get a single project by ID. */
export async function getProjectById(id: string): Promise<ProjectSummary> {
  const res = await request.get(`/projects/${id}`) as ApiResponse<ProjectSummary>
  return unwrapData(res)
}

/** Create a new project on the filesystem. */
export async function createProject(data: {
  name: string
  description?: string
  target?: string
}): Promise<ProjectSummary> {
  const res = await request.post('/projects', data) as ApiResponse<ProjectSummary>
  return unwrapData(res)
}

/** Update project metadata. */
export async function updateProject(
  id: string,
  data: { name?: string; description?: string; target?: string; status?: string }
): Promise<ProjectSummary> {
  const res = await request.put(`/projects/${id}`, data) as ApiResponse<ProjectSummary>
  return unwrapData(res)
}

/** Soft-delete a project by ID. */
export async function deleteProject(id: string): Promise<void> {
  const res = await request.delete(`/projects/${id}`) as ApiResponse
  if (res.code !== 0) {
    throw new Error(res.message || 'Delete failed')
  }
}

// ============================================================================
// Open / Recent / Scan
// ============================================================================

/** Open any real filesystem directory as an AIStudio project. */
export async function openProject(path: string): Promise<ProjectSummary> {
  const res = await request.post('/projects/open', { path }) as ApiResponse<ProjectSummary>
  return unwrapData(res)
}

/** Get recently opened projects. */
export async function getRecentProjects(): Promise<ProjectSummary[]> {
  const res = await request.get('/projects/recent') as ApiResponse<ProjectSummary[]>
  return unwrapData(res)
}

/** Re-scan the projects directory for new/removed projects. */
export async function scanProjects(): Promise<ProjectSummary[]> {
  const res = await request.post('/projects/scan') as ApiResponse<ProjectSummary[]>
  return unwrapData(res)
}

// ============================================================================
// Workflow I/O (stored as workflow.json inside the project directory)
// ============================================================================

/** Read the workflow.json for a project. */
export async function readWorkflow(id: string): Promise<any> {
  const res = await request.get(`/projects/${id}/workflow`) as ApiResponse
  return unwrapData(res)
}

/** Save/replace the workflow.json for a project. */
export async function saveWorkflow(id: string, data: any): Promise<void> {
  const res = await request.put(`/projects/${id}/workflow`, data) as ApiResponse
  if (res.code !== 0) {
    throw new Error(res.message || 'Save workflow failed')
  }
}
