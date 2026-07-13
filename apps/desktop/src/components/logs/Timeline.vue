<template>
  <div class="timeline">
    <div class="timeline-nodes">
      <div v-for="(node, idx) in timeline.nodes" :key="node.nodeId" class="timeline-node">
        <div class="node-indicator">
          <span class="node-dot" :class="`dot-${node.status}`"></span>
          <span v-if="idx < timeline.nodes.length - 1" class="node-line" :class="`line-${node.status}`"></span>
        </div>
        <div class="node-content">
          <div class="node-header">
            <span class="node-name">{{ node.name }}</span>
            <span class="node-status" :class="`status-${node.status}`">{{ statusLabel(node.status) }}</span>
          </div>
          <div v-if="node.status === 'running' && node.progress != null" class="node-progress">
            <div class="progress-bar">
              <div class="progress-fill" :style="{ width: `${node.progress * 100}%` }"></div>
            </div>
            <span class="progress-text">{{ Math.round(node.progress * 100) }}%</span>
          </div>
          <div v-if="node.duration != null" class="node-duration">
            {{ node.duration.toFixed(1) }}s
          </div>
        </div>
      </div>
    </div>
    <div v-if="!timeline.nodes.length" class="timeline-empty">
      No timeline data available
    </div>
  </div>
</template>

<script setup lang="ts">
import type { WorkflowTimeline, StepStatus } from '@/pages/Logs/types'

interface Props {
  timeline: WorkflowTimeline
}

defineProps<Props>()

function statusLabel(status: StepStatus): string {
  const map: Record<StepStatus, string> = {
    completed: 'Completed',
    running: 'Running',
    failed: 'Failed',
    pending: 'Waiting',
  }
  return map[status]
}
</script>

<style scoped>
.timeline {
  padding: var(--spacing-3);
}

.timeline-nodes {
  display: flex;
  flex-direction: column;
}

.timeline-node {
  display: flex;
  gap: var(--spacing-3);
  min-height: 48px;
}

.node-indicator {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 20px;
  flex-shrink: 0;
}

.node-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  flex-shrink: 0;
  border: 2px solid;
}

.dot-completed { border-color: var(--success); background: var(--success); }
.dot-running { border-color: var(--info); background: var(--info); animation: pulse 2s ease-in-out infinite; }
.dot-failed { border-color: var(--error); background: var(--error); }
.dot-pending { border-color: var(--text-disabled); background: transparent; }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.node-line {
  width: 2px;
  flex: 1;
  min-height: 16px;
}

.line-completed { background: var(--success); }
.line-running { background: var(--info); }
.line-failed { background: var(--error); }
.line-pending { background: var(--border-default); }

.node-content {
  flex: 1;
  padding-bottom: var(--spacing-3);
}

.node-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.node-name {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.node-status {
  font-size: var(--text-caption);
  font-weight: var(--font-semibold);
}

.status-completed { color: var(--success); }
.status-running { color: var(--info); }
.status-failed { color: var(--error); }
.status-pending { color: var(--text-disabled); }

.node-progress {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  margin-top: var(--spacing-1);
}

.progress-bar {
  flex: 1;
  height: 4px;
  background: var(--bg-tertiary);
  border-radius: 2px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: var(--primary);
  border-radius: 2px;
  transition: width 300ms ease-out;
}

.progress-text {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  font-family: var(--font-family-mono);
  min-width: 32px;
}

.node-duration {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  font-family: var(--font-family-mono);
  margin-top: 2px;
}

.timeline-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-8);
  color: var(--text-tertiary);
  font-size: var(--text-body-sm);
}
</style>