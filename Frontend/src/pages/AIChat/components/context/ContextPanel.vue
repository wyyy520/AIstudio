<template>
  <div class="context-panel">
    <div class="context-panel-header">
      <span class="context-panel-title">Context</span>
    </div>

    <div class="context-panel-content">
      <ContextSection title="当前项目" :collapsible="false">
        <div v-if="context.project" class="context-item">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" />
          </svg>
          <span class="context-item-text">{{ context.project.name }}</span>
        </div>
        <div v-else class="context-empty">未选择项目</div>
      </ContextSection>

      <ContextSection title="Workflow" v-if="context.workflow">
        <div class="context-item">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M6 3h3v6H6V3zm0 12h3v6H6v-6zm9-12h3v6h-3V3zm0 12h3v6h-3v-6zm-9 0V9m3 9v-3m3 3v-3m3 3V9" />
          </svg>
          <span class="context-item-text">{{ context.workflow.name }}</span>
        </div>
        <div class="context-nodes">
          <span v-for="(node, i) in context.workflow.nodes" :key="i" class="context-node-tag">
            {{ node }}
          </span>
        </div>
      </ContextSection>

      <ContextSection title="文件" v-if="context.files && context.files.length > 0">
        <div v-for="file in context.files" :key="file.path" class="context-item context-item-file">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
            <polyline points="14 2 14 8 20 8" />
          </svg>
          <span class="context-item-text">{{ file.name }}</span>
        </div>
      </ContextSection>

      <ContextSection title="插件" v-if="context.plugins && context.plugins.length > 0">
        <div v-for="plugin in context.plugins" :key="plugin.name" class="context-item">
          <StatusDot :status="plugin.status" />
          <span class="context-item-text">{{ plugin.name }}</span>
        </div>
      </ContextSection>

      <ContextSection title="MCP" v-if="context.mcpServers && context.mcpServers.length > 0">
        <div v-for="server in context.mcpServers" :key="server.name" class="context-item">
          <StatusDot :status="server.status" />
          <span class="context-item-text">{{ server.name }}</span>
        </div>
      </ContextSection>
    </div>
  </div>
</template>

<script setup lang="ts">
import ContextSection from './ContextSection.vue'
import StatusDot from '../shared/StatusDot.vue'
import type { ChatContext } from '../../types'

defineProps<{
  context: ChatContext
}>()
</script>

<style scoped>
.context-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--bg-secondary);
  border-left: 1px solid var(--border-subtle);
}

.context-panel-header {
  display: flex;
  align-items: center;
  padding: var(--spacing-3) var(--spacing-3) var(--spacing-2);
  flex-shrink: 0;
}

.context-panel-title {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-secondary);
}

.context-panel-content {
  flex: 1;
  overflow-y: auto;
}

.context-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-1) 0;
  color: var(--text-secondary);
}

.context-item-file {
  cursor: pointer;
  transition: color var(--transition-fast);
}

.context-item-file:hover {
  color: var(--text-primary);
}

.context-item-text {
  font-size: var(--text-body-sm);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.context-empty {
  font-size: var(--text-body-sm);
  color: var(--text-disabled);
  font-style: italic;
}

.context-nodes {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-1);
  margin-top: var(--spacing-1);
}

.context-node-tag {
  display: inline-flex;
  align-items: center;
  height: 20px;
  padding: 0 var(--spacing-2);
  background: var(--primary-bg);
  color: var(--primary);
  border-radius: var(--radius-sm);
  font-size: var(--text-caption);
}
</style>
