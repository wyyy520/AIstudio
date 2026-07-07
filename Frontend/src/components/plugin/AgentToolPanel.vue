<template>
  <div class="agent-tool-panel">
    <div class="agent-tool-header">
      <svg class="agent-tool-header-icon" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71 M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
      </svg>
      <span class="agent-tool-header-title">Agent Tools</span>
    </div>
    <div class="agent-tool-list">
      <div
        v-for="tool in tools"
        :key="tool.name"
        class="agent-tool-item"
      >
        <div class="agent-tool-item-header">
          <svg class="agent-tool-item-icon" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="var(--primary)" stroke-width="1.5">
            <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z" />
          </svg>
          <span class="agent-tool-name">{{ tool.name }}</span>
          <span class="agent-tool-returns">{{ tool.returns }}</span>
        </div>
        <div class="agent-tool-desc">{{ tool.description }}</div>
        <div v-if="tool.parameters.length > 0" class="agent-tool-params">
          <div
            v-for="param in tool.parameters"
            :key="param.name"
            class="agent-tool-param"
          >
            <span class="param-name">{{ param.name }}</span>
            <span class="param-type">{{ param.type }}</span>
            <span v-if="param.required" class="param-required">required</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { AgentToolDef } from '@/pages/PluginStore/types'

interface Props {
  tools: AgentToolDef[]
}

defineProps<Props>()
</script>

<style scoped>
.agent-tool-panel {
  background: var(--bg-tertiary);
  border-radius: var(--radius-xl);
  padding: var(--spacing-4);
}

.agent-tool-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-3);
}

.agent-tool-header-icon {
  color: var(--text-tertiary);
  flex-shrink: 0;
}

.agent-tool-header-title {
  font-size: var(--text-h3);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.agent-tool-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.agent-tool-item {
  background: var(--bg-secondary);
  border-radius: var(--radius-lg);
  padding: var(--spacing-3);
}

.agent-tool-item-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-1);
}

.agent-tool-item-icon {
  flex-shrink: 0;
}

.agent-tool-name {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.agent-tool-returns {
  margin-left: auto;
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  font-family: var(--font-family-mono);
  background: var(--bg-tertiary);
  padding: 2px 8px;
  border-radius: var(--radius-sm);
}

.agent-tool-desc {
  font-size: var(--text-caption);
  color: var(--text-secondary);
  margin-bottom: var(--spacing-2);
}

.agent-tool-params {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.agent-tool-param {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--text-caption);
  font-family: var(--font-family-mono);
}

.param-name {
  color: var(--text-primary);
}

.param-type {
  color: var(--primary);
}

.param-required {
  font-size: 10px;
  color: var(--warning);
  background: var(--warning-bg);
  padding: 1px 6px;
  border-radius: var(--radius-sm);
  font-family: var(--font-family-sans);
}
</style>