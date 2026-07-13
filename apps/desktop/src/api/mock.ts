/**
 * Mock Adapter for Workflow and Task APIs
 * Provides simulated responses when backend is not available
 */

import type { ApiWorkflow, ApiWorkflowRunResult } from './workflow'

// Mock data storage
const mockWorkflows: Map<string, ApiWorkflow> = new Map()
const mockTasks: Map<string, { id: string; status: string; progress: number }> = new Map()

let mockWorkflowIdCounter = 1
let mockTaskIdCounter = 1

// Check if backend is available
async function isBackendAvailable(): Promise<boolean> {
  try {
    const response = await fetch('/api/health', { method: 'GET' })
    return response.ok
  } catch {
    return false
  }
}

// Generate mock workflow
function generateMockWorkflow(id?: string): ApiWorkflow {
  const workflowId = id || `workflow_${mockWorkflowIdCounter++}`
  return {
    id: parseInt(workflowId.replace(/\D/g, '') || String(mockWorkflowIdCounter++)),
    name: `Workflow ${workflowId}`,
    projectId: 1,
    definition: JSON.stringify({ nodes: [], edges: [] }),
    status: 'idle',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  }
}

// Generate mock task ID
function generateMockTaskId(): string {
  return `task_${Date.now()}_${mockTaskIdCounter++}`
}

// Mock API responses
export const mockApi = {
  workflows: {
    list: async (): Promise<ApiWorkflow[]> => {
      return Array.from(mockWorkflows.values())
    },

    get: async (id: string): Promise<ApiWorkflow | null> => {
      return mockWorkflows.get(id) || generateMockWorkflow(id)
    },

    create: async (data: any): Promise<ApiWorkflow> => {
      const workflow = generateMockWorkflow()
      workflow.name = data.name || workflow.name
      workflow.definition = data.definition || workflow.definition
      mockWorkflows.set(String(workflow.id), workflow)
      return workflow
    },

    update: async (id: string, data: any): Promise<ApiWorkflow> => {
      const workflow = mockWorkflows.get(id) || generateMockWorkflow(id)
      Object.assign(workflow, data)
      workflow.updatedAt = new Date().toISOString()
      mockWorkflows.set(id, workflow)
      return workflow
    },

    delete: async (id: string): Promise<boolean> => {
      mockWorkflows.delete(id)
      return true
    },

    run: async (id: string): Promise<ApiWorkflowRunResult> => {
      const taskId = generateMockTaskId()
      mockTasks.set(taskId, {
        id: taskId,
        status: 'running',
        progress: 0,
      })

      // Simulate task progress
      simulateTaskProgress(taskId)

      return {
        id: taskId,
        workflowId: id,
        status: 'running',
        message: 'Workflow execution started',
      }
    },
  },

  tasks: {
    getStatus: async (taskId: string) => {
      const task = mockTasks.get(taskId)
      if (!task) {
        return { status: 'not_found', progress: 0 }
      }
      return {
        status: task.status,
        progress: task.progress,
      }
    },
  },
}

// Simulate task progress
function simulateTaskProgress(taskId: string): void {
  let progress = 0
  const interval = setInterval(() => {
    const task = mockTasks.get(taskId)
    if (!task) {
      clearInterval(interval)
      return
    }

    progress += Math.random() * 15
    if (progress >= 100) {
      progress = 100
      task.status = 'completed'
      task.progress = 100
      clearInterval(interval)

      // Dispatch WebSocket event for completion
      dispatchMockWebSocketEvent({
        type: 'task_complete',
        taskId,
        data: {
          status: 'completed',
          progress: 100,
          timestamp: new Date().toISOString(),
        },
      })
    } else {
      task.progress = Math.round(progress)

      // Dispatch WebSocket event for progress
      dispatchMockWebSocketEvent({
        type: 'task_progress',
        taskId,
        data: {
          status: 'running',
          progress: task.progress,
          timestamp: new Date().toISOString(),
        },
      })
    }
  }, 800)
}

// Mock WebSocket event dispatcher
function dispatchMockWebSocketEvent(event: any): void {
  // Placeholder for future WebSocket event dispatching
  void event
}

// Export mock adapter
export const mockAdapter = {
  isBackendAvailable,
  mockApi,
}

export default mockAdapter