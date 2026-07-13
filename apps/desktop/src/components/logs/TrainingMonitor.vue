<template>
  <div class="training-monitor">
    <div class="metrics-grid">
      <div class="metric-card">
        <div class="metric-header">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M22 12h-4l-3 9L9 3l-3 9H2" />
          </svg>
          <span class="metric-label">Epoch</span>
        </div>
        <div class="metric-value">
          <span class="metric-current">{{ metrics.currentEpoch }}</span>
          <span class="metric-total">/ {{ metrics.totalEpochs }}</span>
        </div>
        <div class="metric-progress">
          <div class="progress-bar">
            <div class="progress-fill" :style="{ width: `${(metrics.currentEpoch / metrics.totalEpochs) * 100}%` }"></div>
          </div>
        </div>
      </div>

      <div class="metric-card">
        <div class="metric-header">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="22 7 13.5 15.5 8.5 10.5 2 17" /><polyline points="16 7 22 7 22 13" />
          </svg>
          <span class="metric-label">Loss</span>
        </div>
        <div class="metric-value">
          <span class="metric-number">{{ metrics.metrics.loss.toFixed(4) }}</span>
          <span class="metric-trend down">
            <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="m7 7 10 10" /><path d="M17 17V7" /><path d="M7 17h10" />
            </svg>
          </span>
        </div>
      </div>

      <div class="metric-card">
        <div class="metric-header">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="22 7 13.5 15.5 8.5 10.5 2 17" /><polyline points="16 17 22 17 22 11" />
          </svg>
          <span class="metric-label">Accuracy</span>
        </div>
        <div class="metric-value">
          <span class="metric-number">{{ (metrics.metrics.accuracy * 100).toFixed(1) }}%</span>
          <span class="metric-trend up">
            <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="m17 7-10 10" /><path d="M7 7v10" /><path d="M17 17H7" />
            </svg>
          </span>
        </div>
      </div>

      <div class="metric-card">
        <div class="metric-header">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <rect x="4" y="4" width="16" height="16" rx="2" /><path d="M9 9h6v6H9z" /><path d="M9 1v3" /><path d="M15 1v3" /><path d="M9 20v3" /><path d="M15 20v3" /><path d="M20 9h3" /><path d="M20 14h3" /><path d="M1 9h3" /><path d="M1 14h3" />
          </svg>
          <span class="metric-label">GPU</span>
        </div>
        <div class="metric-value">
          <span class="metric-number" :class="{ warning: metrics.metrics.gpuUsage > 0.9 }">{{ (metrics.metrics.gpuUsage * 100).toFixed(0) }}%</span>
        </div>
        <div class="metric-progress">
          <div class="progress-bar">
            <div
              class="progress-fill"
              :class="{ warning: metrics.metrics.gpuUsage > 0.9 }"
              :style="{ width: `${metrics.metrics.gpuUsage * 100}%` }"
            ></div>
          </div>
        </div>
      </div>
    </div>

    <div class="charts-row">
      <div class="chart-card">
        <span class="chart-title">Loss Curve</span>
        <div class="chart-area">
          <svg class="chart-svg" viewBox="0 0 300 100" preserveAspectRatio="none">
            <polyline
              v-if="metrics.history.length > 1"
              :points="lossPoints"
              fill="none"
              stroke="var(--primary)"
              stroke-width="1.5"
              vector-effect="non-scaling-stroke"
            />
          </svg>
        </div>
      </div>
      <div class="chart-card">
        <span class="chart-title">Accuracy Curve</span>
        <div class="chart-area">
          <svg class="chart-svg" viewBox="0 0 300 100" preserveAspectRatio="none">
            <polyline
              v-if="metrics.history.length > 1"
              :points="accuracyPoints"
              fill="none"
              stroke="var(--success)"
              stroke-width="1.5"
              vector-effect="non-scaling-stroke"
            />
          </svg>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { TrainingMetrics } from '@/pages/Logs/types'

interface Props {
  metrics: TrainingMetrics
}

const props = defineProps<Props>()

const lossPoints = computed(() => {
  const h = props.metrics.history
  if (h.length < 2) return ''
  const maxLoss = Math.max(...h.map(e => e.loss))
  const minLoss = Math.min(...h.map(e => e.loss))
  const range = maxLoss - minLoss || 1
  return h.map((e, i) => {
    const x = (i / (h.length - 1)) * 300
    const y = 95 - ((e.loss - minLoss) / range) * 90
    return `${x},${y}`
  }).join(' ')
})

const accuracyPoints = computed(() => {
  const h = props.metrics.history
  if (h.length < 2) return ''
  const maxAcc = Math.max(...h.map(e => e.accuracy))
  const minAcc = Math.min(...h.map(e => e.accuracy))
  const range = maxAcc - minAcc || 1
  return h.map((e, i) => {
    const x = (i / (h.length - 1)) * 300
    const y = 95 - ((e.accuracy - minAcc) / range) * 90
    return `${x},${y}`
  }).join(' ')
})
</script>

<style scoped>
.training-monitor {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
  padding: var(--spacing-3);
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--spacing-3);
}

.metric-card {
  background: var(--bg-tertiary);
  border-radius: var(--radius-xl);
  padding: var(--spacing-3);
  border: 1px solid var(--border-subtle);
}

.metric-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  margin-bottom: var(--spacing-2);
}

.metric-header svg {
  color: var(--text-tertiary);
}

.metric-label {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.metric-value {
  display: flex;
  align-items: baseline;
  gap: var(--spacing-1);
  margin-bottom: var(--spacing-2);
}

.metric-current {
  font-size: var(--text-h2);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  font-family: var(--font-family-mono);
}

.metric-total {
  font-size: var(--text-body-sm);
  color: var(--text-tertiary);
  font-family: var(--font-family-mono);
}

.metric-number {
  font-size: var(--text-h2);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  font-family: var(--font-family-mono);
}

.metric-number.warning {
  color: var(--warning);
}

.metric-trend {
  display: flex;
  align-items: center;
}

.metric-trend.up { color: var(--success); }
.metric-trend.down { color: var(--success); }

.metric-progress {
  margin-top: 2px;
}

.progress-bar {
  height: 4px;
  background: var(--bg-secondary);
  border-radius: 2px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: var(--primary);
  border-radius: 2px;
  transition: width 300ms ease-out;
}

.progress-fill.warning {
  background: var(--warning);
}

.charts-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--spacing-3);
}

.chart-card {
  background: var(--bg-tertiary);
  border-radius: var(--radius-xl);
  padding: var(--spacing-3);
  border: 1px solid var(--border-subtle);
}

.chart-title {
  font-size: var(--text-caption);
  font-weight: var(--font-semibold);
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  display: block;
  margin-bottom: var(--spacing-2);
}

.chart-area {
  height: 80px;
  position: relative;
}

.chart-svg {
  width: 100%;
  height: 100%;
}
</style>