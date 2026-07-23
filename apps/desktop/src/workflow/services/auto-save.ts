import { watch } from 'vue'
import { useWorkflowStore } from '../store'
import { projectManager } from './project-manager'

let saveTimer: ReturnType<typeof setTimeout> | null = null
const SAVE_DEBOUNCE_MS = 1000

export function startAutoSave() {
  const store = useWorkflowStore()

  watch(
    () => ({
      nodes: store.workflow.nodes,
      edges: store.workflow.edges,
      viewport: store.workflow.viewport,
      name: store.workflow.name,
    }),
    () => {
      store.isDirty = true
      if (saveTimer) clearTimeout(saveTimer)
      saveTimer = setTimeout(() => {
        doSave(store)
      }, SAVE_DEBOUNCE_MS)
    },
    { deep: true },
  )
}

export function stopAutoSave() {
  if (saveTimer) {
    clearTimeout(saveTimer)
    saveTimer = null
  }
}

export function flushSave(): Promise<boolean> {
  if (saveTimer) {
    clearTimeout(saveTimer)
    saveTimer = null
  }
  const store = useWorkflowStore()
  return doSave(store)
}

async function doSave(store: ReturnType<typeof useWorkflowStore>): Promise<boolean> {
  // Require a project ID to save via the backend API
  if (!store.project?.id) return false
  try {
    const ok = await projectManager.saveWorkflow(store.project.id)
    if (ok) {
      store.lastSaved = store.workflow.updated_at
      store.isDirty = false
    }
    return ok
  } catch (err) {
    console.error('[auto-save] save failed:', err)
    return false
  }
}
