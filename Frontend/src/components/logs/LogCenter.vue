<template>
  <div class="log-center">
    <LogToolbar
      :search-query="store.searchQuery"
      :filter-level="store.filterLevel"
      :selected-task-id="store.selectedTaskId"
      :is-analyzing="store.isAnalyzing"
      @update:search-query="store.setSearchQuery"
      @update:filter-level="store.setFilterLevel"
      @clear="store.clearLogs"
      @export="store.exportCurrentLogs('log')"
      @analyze="store.analyzeCurrentTask"
    />

    <div class="log-center-body">
      <div class="log-center-sidebar" :class="{ collapsed: sidebarCollapsed }">
        <TaskList
          v-if="!sidebarCollapsed"
          :tasks="store.tasks"
          :selected-task-id="store.selectedTaskId"
          :search-query="store.searchQuery"
          :running-count="store.runningCount"
          @select="store.selectTask"
          @update:search-query="store.setSearchQuery"
        />
        <button v-else class="sidebar-expand-btn" @click="sidebarCollapsed = false">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="m9 18 6-6-6-6" />
          </svg>
        </button>
      </div>

      <div class="log-center-main">
        <div v-if="store.selectedTask" class="log-workspace">
          <LogViewer
            :logs="store.filteredLogs"
            :active-tab="store.activeTab"
            @update:active-tab="store.setActiveTab"
            @download="store.exportCurrentLogs('log')"
          />
        </div>
        <div v-else class="log-workspace-empty">
          <svg viewBox="0 0 24 24" width="48" height="48" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" /><path d="M14 2v6h6" /><path d="M16 13H8" /><path d="M16 17H8" /><path d="M10 9H8" />
          </svg>
          <span class="empty-title">No Task Selected</span>
          <span class="empty-desc">Select a task from the list to view logs</span>
        </div>
      </div>

      <div class="log-center-analysis" :class="{ collapsed: analysisCollapsed }">
        <AIAnalysisPanel
          v-if="!analysisCollapsed"
          :analyses="store.errorAnalyses"
          :is-analyzing="store.isAnalyzing"
          :agent-phase="store.agentPhase"
          :selected-task-id="store.selectedTaskId"
          @analyze="store.analyzeCurrentTask"
          @apply-fix="store.startFix"
          @generate-command="handleGenerateCommand"
          @ignore="store.ignoreAnalysis"
        />
        <button v-else class="analysis-expand-btn" @click="analysisCollapsed = false">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 2a10 10 0 1 0 10 10 4 4 0 0 1-5-5 4 4 0 0 1-5-5" /><path d="M8.5 8.5v.01" />
          </svg>
        </button>
      </div>
    </div>

    <div v-if="bottomPanelVisible" class="log-center-bottom">
      <div class="bottom-header">
        <div class="bottom-tabs">
          <button
            v-if="store.selectedTask?.type === 'training' && store.trainingMetrics"
            class="bottom-tab"
            :class="{ active: bottomTab === 'training' }"
            @click="bottomTab = 'training'"
          >
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M22 12h-4l-3 9L9 3l-3 9H2" />
            </svg>
            Training Monitor
          </button>
          <button
            v-if="store.workflowTimeline"
            class="bottom-tab"
            :class="{ active: bottomTab === 'timeline' }"
            @click="bottomTab = 'timeline'"
          >
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M12 20V10" /><path d="M18 20V4" /><path d="M6 20v-4" />
            </svg>
            Workflow Timeline
          </button>
        </div>
        <button class="bottom-close" @click="bottomPanelVisible = false">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M18 6 6 18" /><path d="m6 6 12 12" />
          </svg>
        </button>
      </div>
      <div class="bottom-content">
        <TrainingMonitor v-if="bottomTab === 'training' && store.trainingMetrics" :metrics="store.trainingMetrics" />
        <Timeline v-else-if="bottomTab === 'timeline' && store.workflowTimeline" :timeline="store.workflowTimeline" />
      </div>
    </div>

    <FixDialog
      :visible="store.showFixDialog"
      :steps="store.fixSteps"
      :command="currentFixCommand"
      :is-fixing="store.isFixing"
      @cancel="store.cancelFix"
      @execute="store.executeFix"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useLogStore } from '@/store/log'
import LogToolbar from '@/components/logs/LogToolbar.vue'
import TaskList from '@/components/logs/TaskList.vue'
import LogViewer from '@/components/logs/LogViewer.vue'
import AIAnalysisPanel from '@/components/logs/AIAnalysisPanel.vue'
import TrainingMonitor from '@/components/logs/TrainingMonitor.vue'
import Timeline from '@/components/logs/Timeline.vue'
import FixDialog from '@/components/logs/FixDialog.vue'
import type { ErrorAnalysis } from '@/pages/Logs/types'

const store = useLogStore()

const sidebarCollapsed = ref(false)
const analysisCollapsed = ref(false)
const bottomPanelVisible = ref(false)
const bottomTab = ref<'training' | 'timeline'>('training')

const currentFixCommand = computed(() => {
  if (!store.currentFixSolutionId) return undefined
  for (const analysis of store.errorAnalyses) {
    const sol = analysis.solutions.find(s => s.id === store.currentFixSolutionId)
    if (sol?.command) return sol.command
  }
  return undefined
})

watch(() => store.selectedTask, (task) => {
  if (task?.type === 'training' && store.trainingMetrics) {
    bottomTab.value = 'training'
    bottomPanelVisible.value = true
  } else if (store.workflowTimeline) {
    bottomTab.value = 'timeline'
    bottomPanelVisible.value = true
  } else {
    bottomPanelVisible.value = false
  }
})

function handleGenerateCommand(analysis: ErrorAnalysis) {
  if (analysis.solutions.length > 0 && analysis.solutions[0].command) {
    navigator.clipboard.writeText(analysis.solutions[0].command)
  }
}

onMounted(() => {
  store.loadTasks()
  store.connectWebSocket()
})

onUnmounted(() => {
  store.disconnectWebSocket()
})
</script>

<style scoped>
.log-center {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.log-center-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.log-center-sidebar {
  width: 240px;
  flex-shrink: 0;
  transition: width var(--transition-normal);
  overflow: hidden;
}

.log-center-sidebar.collapsed {
  width: 48px;
}

.sidebar-expand-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 48px;
  background: var(--bg-secondary);
  border: none;
  border-right: 1px solid var(--border-subtle);
  color: var(--text-tertiary);
  cursor: pointer;
  transition: color var(--transition-fast);
}

.sidebar-expand-btn:hover {
  color: var(--text-primary);
}

.log-center-main {
  flex: 1;
  overflow: hidden;
  min-width: 0;
}

.log-workspace {
  height: 100%;
}

.log-workspace-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: var(--spacing-3);
  color: var(--text-tertiary);
}

.log-workspace-empty svg {
  opacity: 0.3;
}

.empty-title {
  font-size: var(--text-h3);
  font-weight: var(--font-semibold);
  color: var(--text-secondary);
}

.empty-desc {
  font-size: var(--text-body-sm);
  color: var(--text-tertiary);
}

.log-center-analysis {
  width: 360px;
  flex-shrink: 0;
  transition: width var(--transition-normal);
  overflow: hidden;
}

.log-center-analysis.collapsed {
  width: 48px;
}

.analysis-expand-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 48px;
  background: var(--bg-secondary);
  border: none;
  border-left: 1px solid var(--border-subtle);
  color: var(--primary);
  cursor: pointer;
  transition: color var(--transition-fast);
}

.analysis-expand-btn:hover {
  color: var(--primary-hover);
}

.log-center-bottom {
  border-top: 1px solid var(--border-subtle);
  background: var(--bg-secondary);
  max-height: 300px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.bottom-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--spacing-3);
  border-bottom: 1px solid var(--border-subtle);
  height: 36px;
}

.bottom-tabs {
  display: flex;
  gap: 0;
}

.bottom-tab {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  height: 36px;
  padding: 0 var(--spacing-3);
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--text-secondary);
  font-size: var(--text-body-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.bottom-tab:hover {
  color: var(--text-primary);
}

.bottom-tab.active {
  color: var(--text-primary);
  border-bottom-color: var(--primary);
}

.bottom-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  color: var(--text-tertiary);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.bottom-close:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.bottom-content {
  flex: 1;
  overflow: auto;
}
</style>