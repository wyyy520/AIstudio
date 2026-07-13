<template>
  <footer class="statusbar">
    <div class="statusbar-left">
      <span class="statusbar-item" :class="runtimeClass">
        <span class="status-dot" :class="runtimeDotClass" />
        {{ runtimeLabel }}
      </span>
      <span v-if="projectName" class="statusbar-separator">|</span>
      <span v-if="projectName" class="statusbar-item">{{ projectName }}</span>
      <span v-if="runStatus !== 'idle'" class="statusbar-separator">|</span>
      <span v-if="runStatus !== 'idle'" class="statusbar-item" :class="runStatusClass">
        {{ runStatusLabel }}
      </span>
    </div>
    <div class="statusbar-right">
      <span class="statusbar-item">AIStudio v0.1.0</span>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRuntimeStore } from '@/stores/runtime'
import { useProjectStore } from '@/stores/project'
import { getHealth } from '@/api/health'

const runtimeStore = useRuntimeStore()
const projectStore = useProjectStore()

const healthOk = ref(false)
let healthInterval: ReturnType<typeof setInterval> | null = null

async function checkHealth() {
  try {
    await getHealth()
    healthOk.value = true
  } catch {
    healthOk.value = false
  }
}

onMounted(() => {
  checkHealth()
  healthInterval = setInterval(checkHealth, 15000)
})

onUnmounted(() => {
  if (healthInterval) clearInterval(healthInterval)
})

const projectName = computed(() => projectStore.currentProject?.name || '')
const runStatus = computed(() => runtimeStore.status)

const runtimeLabel = computed(() => healthOk.value ? 'Connected' : 'Disconnected')
const runtimeClass = computed(() => healthOk.value ? 'statusbar-item--ok' : 'statusbar-item--err')
const runtimeDotClass = computed(() => healthOk.value ? 'status-dot--ok' : 'status-dot--err')

const runStatusLabel = computed(() => {
  const map: Record<string, string> = {
    idle: '', compiling: 'Compiling...', running: 'Running...',
    completed: 'Completed', failed: 'Failed',
  }
  return map[runStatus.value] || ''
})

const runStatusClass = computed(() => {
  const map: Record<string, string> = {
    compiling: 'statusbar-item--warn', running: 'statusbar-item--warn',
    completed: 'statusbar-item--ok', failed: 'statusbar-item--err',
  }
  return map[runStatus.value] || ''
})
</script>

<style scoped>
.statusbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 24px;
  padding: 0 12px;
  background: var(--bg-secondary);
  border-top: 1px solid var(--border-subtle);
  flex-shrink: 0;
}

.statusbar-left, .statusbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.statusbar-item {
  font-size: 11px;
  color: var(--text-tertiary);
  display: flex;
  align-items: center;
  gap: 4px;
}

.statusbar-item--ok { color: var(--success); }
.statusbar-item--warn { color: var(--warning); }
.statusbar-item--err { color: var(--error); }

.statusbar-separator {
  color: var(--border-subtle);
  font-size: 11px;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  display: inline-block;
  flex-shrink: 0;
}

.status-dot--ok { background: var(--success); }
.status-dot--warn { background: var(--warning); }
.status-dot--err { background: var(--error); }
</style>
