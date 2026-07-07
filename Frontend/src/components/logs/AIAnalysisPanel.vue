<template>
  <div class="ai-analysis-panel">
    <div class="panel-header">
      <div class="panel-title">
        <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 2a10 10 0 1 0 10 10 4 4 0 0 1-5-5 4 4 0 0 1-5-5" /><path d="M8.5 8.5v.01" /><path d="M16 15.5v.01" /><path d="M12 12v.01" />
        </svg>
        <span>AI Analysis</span>
      </div>
      <AppButton
        type="primary"
        size="small"
        label="Analyze"
        :loading="isAnalyzing"
        :disabled="!selectedTaskId || isAnalyzing"
        @click="$emit('analyze')"
      />
    </div>

    <div class="panel-body">
      <div v-if="analyses.length" class="analysis-list">
        <ErrorCard
          v-for="analysis in analyses"
          :key="analysis.id"
          :analysis="analysis"
          @apply-fix="$emit('applyFix', $event, arguments[1])"
          @generate-command="$emit('generateCommand', $event)"
          @ignore="$emit('ignore', $event)"
        />
      </div>

      <div v-else-if="isAnalyzing" class="analysis-loading">
        <div class="loading-skeleton">
          <div class="skeleton-line skeleton-title"></div>
          <div class="skeleton-line skeleton-text"></div>
          <div class="skeleton-line skeleton-text short"></div>
          <div class="skeleton-line skeleton-text"></div>
        </div>
      </div>

      <div v-else class="analysis-empty">
        <svg viewBox="0 0 24 24" width="40" height="40" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 2a10 10 0 1 0 10 10 4 4 0 0 1-5-5 4 4 0 0 1-5-5" /><path d="M8.5 8.5v.01" /><path d="M16 15.5v.01" />
        </svg>
        <span class="empty-title">AI Analysis Ready</span>
        <span class="empty-desc">Click Analyze to start analyzing logs</span>
      </div>
    </div>

    <div class="panel-footer">
      <div class="agent-status">
        <span class="agent-dot" :class="`phase-${agentPhase}`"></span>
        <span class="agent-label">Agent:</span>
        <span class="agent-phase" :class="`phase-${agentPhase}`">{{ phaseLabel }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ErrorAnalysis, AgentPhase } from '@/pages/Logs/types'
import ErrorCard from './ErrorCard.vue'
import AppButton from '@/components/AppButton/AppButton.vue'

interface Props {
  analyses: ErrorAnalysis[]
  isAnalyzing: boolean
  agentPhase: AgentPhase
  selectedTaskId: string | null
}

const props = defineProps<Props>()

defineEmits<{
  analyze: []
  applyFix: [analysisId: string, solutionId: string]
  generateCommand: [analysis: ErrorAnalysis]
  ignore: [analysisId: string]
}>()

const phaseLabel = computed(() => {
  const map: Record<AgentPhase, string> = {
    idle: 'Idle',
    thinking: 'Thinking...',
    analyzing: 'Analyzing...',
    calling_tool: 'Calling Tool...',
    executing: 'Executing...',
    completed: 'Completed',
    failed: 'Failed',
  }
  return map[props.agentPhase]
})
</script>

<style scoped>
.ai-analysis-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--bg-secondary);
  border-left: 1px solid var(--border-subtle);
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-3) var(--spacing-4);
  border-bottom: 1px solid var(--border-subtle);
}

.panel-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.panel-title svg {
  color: var(--primary);
}

.panel-body {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-3);
}

.analysis-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.analysis-loading {
  padding: var(--spacing-4);
}

.loading-skeleton {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.skeleton-line {
  height: 14px;
  background: var(--bg-tertiary);
  border-radius: var(--radius-sm);
  animation: shimmer 1.5s ease-in-out infinite;
}

.skeleton-title {
  width: 60%;
  height: 18px;
}

.skeleton-text {
  width: 100%;
}

.skeleton-text.short {
  width: 40%;
}

@keyframes shimmer {
  0%, 100% { opacity: 0.3; }
  50% { opacity: 0.6; }
}

.analysis-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-2);
  height: 100%;
  color: var(--text-tertiary);
  padding: var(--spacing-8);
}

.analysis-empty svg {
  opacity: 0.3;
}

.empty-title {
  font-size: var(--text-body);
  font-weight: var(--font-semibold);
  color: var(--text-secondary);
}

.empty-desc {
  font-size: var(--text-body-sm);
  color: var(--text-tertiary);
}

.panel-footer {
  padding: var(--spacing-2) var(--spacing-4);
  border-top: 1px solid var(--border-subtle);
}

.agent-status {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.agent-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.agent-dot.phase-idle { background: var(--text-disabled); }
.agent-dot.phase-thinking { background: var(--primary); animation: pulse 2s ease-in-out infinite; }
.agent-dot.phase-analyzing { background: var(--info); animation: pulse 1.5s ease-in-out infinite; }
.agent-dot.phase-calling_tool { background: var(--warning); animation: pulse 1s ease-in-out infinite; }
.agent-dot.phase-executing { background: var(--success); animation: pulse 1.5s ease-in-out infinite; }
.agent-dot.phase-completed { background: var(--success); }
.agent-dot.phase-failed { background: var(--error); }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.agent-label {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.agent-phase {
  font-size: var(--text-caption);
  font-weight: var(--font-semibold);
}

.agent-phase.phase-idle { color: var(--text-tertiary); }
.agent-phase.phase-thinking { color: var(--primary); }
.agent-phase.phase-analyzing { color: var(--info); }
.agent-phase.phase-calling_tool { color: var(--warning); }
.agent-phase.phase-executing { color: var(--success); }
.agent-phase.phase-completed { color: var(--success); }
.agent-phase.phase-failed { color: var(--error); }
</style>