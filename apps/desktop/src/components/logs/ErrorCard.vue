<template>
  <div class="error-card" :class="`severity-${analysis.severity}`">
    <div class="error-card-header">
      <div class="error-card-severity">
        <svg v-if="analysis.severity === 'critical'" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10" /><path d="m15 9-6 6" /><path d="m9 9 6 6" />
        </svg>
        <svg v-else-if="analysis.severity === 'warning'" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z" /><path d="M12 9v4" /><path d="M12 17h.01" />
        </svg>
        <svg v-else viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10" /><path d="M12 16v-4" /><path d="M12 8h.01" />
        </svg>
        <span class="severity-badge" :class="`badge-${analysis.severity}`">{{ severityLabel }}</span>
      </div>
      <span class="error-type">{{ analysis.errorType }}</span>
    </div>

    <div class="error-card-impact">
      <span class="impact-label">影响模块</span>
      <span class="impact-value" :class="`impact-${analysis.impactModule}`">{{ impactLabel }}</span>
    </div>

    <div class="error-card-body">
      <div class="error-section">
        <span class="error-label">问题</span>
        <p class="error-text">{{ analysis.problem }}</p>
      </div>
      <div class="error-section">
        <span class="error-label">原因</span>
        <p class="error-text">{{ analysis.cause }}</p>
      </div>
    </div>

    <div v-if="analysis.solutions.length" class="error-card-solutions">
      <span class="error-label">解决方案</span>
      <div class="solution-list">
        <div v-for="(sol, idx) in analysis.solutions" :key="sol.id" class="solution-card">
          <div class="solution-header">
            <span class="solution-index">{{ idx + 1 }}</span>
            <div class="solution-info">
              <span class="solution-title">{{ sol.title }}</span>
              <span class="solution-desc">{{ sol.description }}</span>
            </div>
          </div>
          <div v-if="sol.command" class="solution-command selectable">
            <span class="solution-cmd-text">{{ sol.command }}</span>
            <button class="cmd-copy-btn" title="复制" @click="copyCommand(sol.command!)">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                <rect width="14" height="14" x="8" y="8" rx="2" ry="2" /><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2" />
              </svg>
            </button>
          </div>
          <div class="solution-meta">
            <span class="solution-time">{{ sol.estimatedTime }}</span>
            <span class="solution-risk" :class="`risk-${sol.risk}`">风险: {{ riskLabel(sol.risk) }}</span>
          </div>
          <div class="solution-actions">
            <button v-if="sol.autoFixable && analysis.status !== 'fixed'" class="fix-btn" @click="$emit('applyFix', analysis.id, sol.id)">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z" />
              </svg>
              一键修复
            </button>
            <span v-if="analysis.status === 'fixed'" class="fixed-badge">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                <path d="M20 6 9 17l-5-5" />
              </svg>
              已修复
            </span>
          </div>
        </div>
      </div>
    </div>

    <div class="error-card-footer">
      <button class="footer-btn footer-btn--secondary" @click="$emit('generateCommand', analysis)">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="m7 4-4 4 4 4" /><path d="m17 4 4 4-4 4" /><path d="m14 4-4 16" />
        </svg>
        复制命令
      </button>
      <button v-if="analysis.status === 'pending'" class="footer-btn footer-btn--ghost" @click="$emit('ignore', analysis.id)">
        忽略
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ErrorAnalysis, ImpactModule } from '@/pages/Logs/types'

interface Props {
  analysis: ErrorAnalysis
}

const props = defineProps<Props>()

defineEmits<{
  applyFix: [analysisId: string, solutionId: string]
  generateCommand: [analysis: ErrorAnalysis]
  ignore: [analysisId: string]
}>()

const severityLabel = computed(() => {
  const map: Record<string, string> = {
    critical: '严重',
    warning: '警告',
    info: '提示',
  }
  return map[props.analysis.severity] || props.analysis.severity
})

const impactLabel = computed(() => {
  const map: Record<ImpactModule, string> = {
    cuda: 'CUDA',
    pytorch: 'PyTorch',
    tensorflow: 'TensorFlow',
    dataset: '数据集',
    config: '配置',
    memory: '内存',
    network: '网络',
    dependency: '依赖',
    unknown: '未知',
  }
  return map[props.analysis.impactModule] || props.analysis.impactModule
})

function riskLabel(risk: string): string {
  const map: Record<string, string> = {
    low: '低',
    medium: '中',
    high: '高',
  }
  return map[risk] || risk
}

function copyCommand(cmd: string) {
  navigator.clipboard.writeText(cmd)
}
</script>

<style scoped>
.error-card {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-xl);
  padding: var(--spacing-4);
  transition: border-color var(--transition-fast);
}

.error-card.severity-critical {
  border-left: 3px solid var(--error);
}

.error-card.severity-warning {
  border-left: 3px solid var(--warning);
}

.error-card.severity-info {
  border-left: 3px solid var(--info);
}

.error-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-3);
}

.error-card-severity {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.error-card.severity-critical .error-card-severity svg { color: var(--error); }
.error-card.severity-warning .error-card-severity svg { color: var(--warning); }
.error-card.severity-info .error-card-severity svg { color: var(--info); }

.severity-badge {
  font-size: 11px;
  font-weight: var(--font-semibold);
  text-transform: uppercase;
  padding: 2px 8px;
  border-radius: var(--radius-sm);
}

.badge-critical { background: var(--error-bg); color: var(--error); }
.badge-warning { background: var(--warning-bg); color: var(--warning); }
.badge-info { background: var(--info-bg); color: var(--info); }

.error-type {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  font-family: var(--font-family-mono);
}

.error-card-impact {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-3);
  padding: var(--spacing-2) var(--spacing-3);
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
}

.impact-label {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.impact-value {
  font-size: var(--text-caption);
  font-weight: var(--font-semibold);
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  background: var(--primary-bg);
  color: var(--primary);
}

.impact-cuda { background: #fef3c7; color: #d97706; }
.impact-pytorch { background: #fee2e2; color: #dc2626; }
.impact-tensorflow { background: #fef3c7; color: #d97706; }
.impact-dataset { background: #dbeafe; color: #2563eb; }
.impact-config { background: #e0e7ff; color: #4f46e5; }
.impact-memory { background: #fce7f3; color: #db2777; }
.impact-network { background: #d1fae5; color: #059669; }
.impact-dependency { background: #f3e8ff; color: #9333ea; }

.error-card-body {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-3);
}

.error-section {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.error-label {
  font-size: var(--text-caption);
  font-weight: var(--font-semibold);
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.error-text {
  font-size: var(--text-body-sm);
  color: var(--text-secondary);
  line-height: var(--leading-body-sm);
}

.error-card-solutions {
  margin-bottom: var(--spacing-3);
}

.solution-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
  margin-top: var(--spacing-2);
}

.solution-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: var(--spacing-3);
}

.solution-header {
  display: flex;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-2);
}

.solution-index {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  background: var(--primary-bg);
  color: var(--primary);
  font-size: 11px;
  font-weight: var(--font-semibold);
  flex-shrink: 0;
}

.solution-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.solution-title {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.solution-desc {
  font-size: var(--text-caption);
  color: var(--text-secondary);
}

.solution-command {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-2) var(--spacing-3);
  background: var(--bg-primary);
  border-radius: var(--radius-md);
  margin-bottom: var(--spacing-2);
  font-family: var(--font-family-mono);
  font-size: var(--text-caption);
  color: var(--text-secondary);
}

.solution-cmd-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cmd-copy-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2px;
  border: none;
  background: transparent;
  color: var(--text-tertiary);
  cursor: pointer;
  border-radius: 4px;
  transition: color var(--transition-fast), background var(--transition-fast);
}

.cmd-copy-btn:hover {
  color: var(--text-primary);
  background: var(--bg-hover);
}

.solution-meta {
  display: flex;
  gap: var(--spacing-3);
  margin-bottom: var(--spacing-2);
}

.solution-time {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.solution-risk {
  font-size: var(--text-caption);
  font-weight: var(--font-semibold);
}

.risk-low { color: var(--success); }
.risk-medium { color: var(--warning); }
.risk-high { color: var(--error); }

.solution-actions {
  display: flex;
  gap: var(--spacing-2);
}

.fix-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  height: 32px;
  padding: 0 var(--spacing-4);
  background: var(--primary);
  color: white;
  border: none;
  border-radius: var(--radius-md);
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  cursor: pointer;
  transition: background var(--transition-fast);
}

.fix-btn:hover {
  background: var(--primary-hover);
}

.fixed-badge {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  font-size: var(--text-caption);
  font-weight: var(--font-semibold);
  color: var(--success);
}

.error-card-footer {
  display: flex;
  gap: var(--spacing-2);
  padding-top: var(--spacing-3);
  border-top: 1px solid var(--border-subtle);
}

.footer-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  height: 28px;
  padding: 0 var(--spacing-3);
  border: none;
  border-radius: var(--radius-md);
  font-size: var(--text-caption);
  cursor: pointer;
  transition: background var(--transition-fast), color var(--transition-fast);
}

.footer-btn--secondary {
  background: var(--bg-secondary);
  color: var(--text-secondary);
  border: 1px solid var(--border-default);
}

.footer-btn--secondary:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.footer-btn--ghost {
  background: transparent;
  color: var(--text-tertiary);
}

.footer-btn--ghost:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}
</style>
