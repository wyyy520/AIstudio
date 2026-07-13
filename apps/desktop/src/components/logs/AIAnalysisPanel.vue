<template>
  <div class="ai-analysis-panel">
    <div class="panel-header">
      <div class="panel-title">
        <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 2a10 10 0 1 0 10 10 4 4 0 0 1-5-5 4 4 0 0 1-5-5" /><path d="M8.5 8.5v.01" /><path d="M16 15.5v.01" /><path d="M12 12v.01" />
        </svg>
        <span>AI 错误助手</span>
      </div>
      <AppButton
        type="primary"
        size="small"
        :label="isAnalyzing ? '分析中...' : '让AI分析'"
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
          @apply-fix="(analysisId, solutionId) => $emit('applyFix', analysisId, solutionId)"
          @generate-command="$emit('generateCommand', $event)"
          @ignore="$emit('ignore', $event)"
        />
      </div>

      <div v-else-if="isAnalyzing" class="analysis-loading">
        <div class="loading-animation">
          <div class="loading-ring"></div>
          <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 2a10 10 0 1 0 10 10 4 4 0 0 1-5-5 4 4 0 0 1-5-5" /><path d="M8.5 8.5v.01" /><path d="M16 15.5v.01" />
          </svg>
        </div>
        <span class="loading-text">AI 正在分析错误日志...</span>
        <span class="loading-hint">正在识别错误模式并生成解决方案</span>
      </div>

      <div v-else class="analysis-empty">
        <div class="empty-icon">
          <svg viewBox="0 0 24 24" width="48" height="48" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 2a10 10 0 1 0 10 10 4 4 0 0 1-5-5 4 4 0 0 1-5-5" /><path d="M8.5 8.5v.01" /><path d="M16 15.5v.01" />
          </svg>
        </div>
        <span class="empty-title">AI 错误助手就绪</span>
        <span class="empty-desc">选择一个任务，然后点击"让AI分析"按钮</span>
        <span class="empty-hint">AI 将自动识别错误并提供解决方案</span>
      </div>
    </div>

    <div class="panel-footer">
      <div class="agent-status">
        <span class="agent-dot" :class="`phase-${agentPhase}`"></span>
        <span class="agent-label">Agent:</span>
        <span class="agent-phase" :class="`phase-${agentPhase}`">{{ phaseLabel }}</span>
      </div>
      <div v-if="analyses.length" class="analysis-count">
        <span class="count-value">{{ analyses.length }}</span>
        <span class="count-label">个错误</span>
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
    idle: '空闲',
    thinking: '思考中...',
    analyzing: '分析中...',
    calling_tool: '调用工具...',
    executing: '执行中...',
    completed: '已完成',
    failed: '失败',
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
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-3);
  height: 100%;
  padding: var(--spacing-8);
}

.loading-animation {
  position: relative;
  width: 64px;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.loading-ring {
  position: absolute;
  inset: 0;
  border: 2px solid var(--primary-bg);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.loading-animation svg {
  color: var(--primary);
  animation: pulse 2s ease-in-out infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.loading-text {
  font-size: var(--text-body);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.loading-hint {
  font-size: var(--text-body-sm);
  color: var(--text-tertiary);
}

.analysis-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-3);
  height: 100%;
  padding: var(--spacing-8);
}

.empty-icon {
  width: 80px;
  height: 80px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--primary-bg);
  border-radius: 50%;
  margin-bottom: var(--spacing-2);
}

.empty-icon svg {
  color: var(--primary);
  opacity: 0.6;
}

.empty-title {
  font-size: var(--text-h3);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.empty-desc {
  font-size: var(--text-body-sm);
  color: var(--text-secondary);
  text-align: center;
}

.empty-hint {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.panel-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
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
.agent-dot.phase-thinking { background: var(--primary); animation: dotPulse 2s ease-in-out infinite; }
.agent-dot.phase-analyzing { background: var(--info); animation: dotPulse 1.5s ease-in-out infinite; }
.agent-dot.phase-calling_tool { background: var(--warning); animation: dotPulse 1s ease-in-out infinite; }
.agent-dot.phase-executing { background: var(--success); animation: dotPulse 1.5s ease-in-out infinite; }
.agent-dot.phase-completed { background: var(--success); }
.agent-dot.phase-failed { background: var(--error); }

@keyframes dotPulse {
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

.analysis-count {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
}

.count-value {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--primary);
}

.count-label {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}
</style>
