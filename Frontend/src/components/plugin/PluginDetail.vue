<template>
  <Transition name="slide-in" mode="out-in">
    <div v-if="plugin" class="plugin-detail" :key="plugin.id">
      <div class="plugin-detail-header">
        <div class="plugin-detail-icon-wrap">
          <svg
            class="plugin-detail-icon"
            viewBox="0 0 24 24"
            width="28"
            height="28"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path :d="plugin.icon" />
          </svg>
        </div>
        <div class="plugin-detail-title-area">
          <h2 class="plugin-detail-name">{{ plugin.name }}</h2>
          <div class="plugin-detail-meta">
            <span class="plugin-detail-version">v{{ plugin.version }}</span>
            <span class="meta-sep">·</span>
            <span class="plugin-detail-author">{{ plugin.author }}</span>
            <span class="meta-sep">·</span>
            <span class="plugin-detail-source">
              <svg v-if="plugin.source === 'github'" viewBox="0 0 24 24" width="12" height="12" fill="currentColor"><path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/></svg>
              <svg v-else-if="plugin.source === 'local'" viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M3 15v4c0 1.1.9 2 2 2h14a2 2 0 0 0 2-2v-4M17 8l-5-5-5 5M12 3v12" /></svg>
              <svg v-else viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" /></svg>
              <span>{{ sourceLabel }}</span>
            </span>
          </div>
        </div>
        <PluginStatus :status="plugin.status" />
      </div>

      <div class="plugin-detail-actions">
        <button
          v-if="plugin.status === 'not-installed'"
          class="action-btn action-btn--primary"
          :disabled="isInstalling"
          @click="$emit('install', plugin.id)"
        >
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3" /></svg>
          <span>Install</span>
        </button>
        <button
          v-if="plugin.status === 'installed'"
          class="action-btn action-btn--secondary"
          :disabled="isInstalling"
          @click="$emit('update', plugin.id)"
        >
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 4v6h6M23 20v-6h-6 M20.49 9A9 9 0 0 0 5.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 0 1 3.51 15" /></svg>
          <span>Update</span>
        </button>
        <button
          v-if="plugin.status === 'installed'"
          class="action-btn action-btn--danger"
          :disabled="isInstalling"
          @click="$emit('remove', plugin.id)"
        >
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 4H8l-7 8 7 8h13a2 2 0 0 0 2-2V6a2 2 0 0 0-2-2z M18 9l-6 6M12 9l6 6" /></svg>
          <span>Remove</span>
        </button>
      </div>

      <div class="plugin-detail-body">
        <section class="detail-section">
          <h3 class="detail-section-title">Description</h3>
          <p class="detail-description">{{ plugin.description }}</p>
        </section>

        <section v-if="plugin.capabilities.length > 0" class="detail-section">
          <h3 class="detail-section-title">Capabilities</h3>
          <div class="detail-capabilities">
            <span
              v-for="cap in plugin.capabilities"
              :key="cap"
              class="capability-tag"
            >
              {{ cap }}
            </span>
          </div>
        </section>

        <section v-if="plugin.workflowNodes.length > 0" class="detail-section">
          <h3 class="detail-section-title">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M6 3h3v6H6V3zm0 12h3v6H6v-6zm9-12h3v6h-3V3zm0 12h3v6h-3v-6zm-9 0V9m3 9v-3m3 3v-3m3 3V9" /></svg>
            Workflow Nodes
          </h3>
          <div class="detail-nodes">
            <div
              v-for="node in plugin.workflowNodes"
              :key="node.name"
              class="workflow-node"
            >
              <span class="node-dot" :style="{ background: nodeCategoryColor(node.category) }"></span>
              <span class="node-name">{{ node.name }}</span>
              <span class="node-type">{{ node.type }}</span>
            </div>
          </div>
        </section>

        <section v-if="plugin.dependencies.length > 0" class="detail-section">
          <DependencyPanel :dependencies="plugin.dependencies" />
        </section>

        <section v-if="plugin.agentTools.length > 0" class="detail-section">
          <AgentToolPanel :tools="plugin.agentTools" />
        </section>

        <section v-if="plugin.tags.length > 0" class="detail-section">
          <h3 class="detail-section-title">Tags</h3>
          <div class="detail-tags">
            <span
              v-for="tag in plugin.tags"
              :key="tag"
              class="detail-tag"
            >
              {{ tag }}
            </span>
          </div>
        </section>
      </div>
    </div>

    <div v-else class="plugin-detail plugin-detail--empty">
      <div class="plugin-detail-empty-content">
        <svg viewBox="0 0 24 24" width="48" height="48" fill="none" stroke="var(--text-tertiary)" stroke-width="1" stroke-linecap="round" stroke-linejoin="round">
          <path d="M3 3h7v7H3V3zm0 11h7v7H3v-7zm11-11h7v7h-7V3zm0 11h7v7h-7v-7z" />
        </svg>
        <p class="empty-title">Select a Plugin</p>
        <p class="empty-desc">Choose a plugin from the list to view details</p>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Plugin, PluginCategory } from '@/pages/PluginStore/types'
import PluginStatus from './PluginStatus.vue'
import DependencyPanel from './DependencyPanel.vue'
import AgentToolPanel from './AgentToolPanel.vue'

interface Props {
  plugin: Plugin | null
  isInstalling?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  isInstalling: false,
})

defineEmits<{
  install: [pluginId: string]
  update: [pluginId: string]
  remove: [pluginId: string]
}>()

const sourceLabels: Record<string, string> = {
  github: 'GitHub',
  local: 'Local',
  registry: 'Registry',
}

const sourceLabel = computed(() => {
  if (!props.plugin) return ''
  return sourceLabels[props.plugin.source] ?? props.plugin.source
})

const nodeCategoryColors: Record<PluginCategory, string> = {
  vision: 'var(--vision)',
  nlp: 'var(--nlp)',
  timeseries: 'var(--timeseries)',
  speech: 'var(--nlp)',
  simulation: 'var(--simulation)',
  system: 'var(--system)',
  mcp: 'var(--mcp)',
}

function nodeCategoryColor(category: PluginCategory): string {
  return nodeCategoryColors[category] ?? 'var(--primary)'
}
</script>

<style scoped>
.plugin-detail {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow-y: auto;
}

.plugin-detail--empty {
  display: flex;
  align-items: center;
  justify-content: center;
}

.plugin-detail-empty-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-3);
  text-align: center;
  opacity: 0.5;
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

.plugin-detail-header {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-4);
  padding: var(--spacing-6);
  border-bottom: 1px solid var(--border-subtle);
}

.plugin-detail-icon-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: var(--radius-xl);
  background: var(--bg-tertiary);
  flex-shrink: 0;
}

.plugin-detail-icon {
  color: var(--primary);
}

.plugin-detail-title-area {
  flex: 1;
  min-width: 0;
}

.plugin-detail-name {
  font-size: var(--text-h1);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  line-height: var(--leading-h1);
  margin: 0;
}

.plugin-detail-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: var(--spacing-1);
  font-size: var(--text-caption);
}

.plugin-detail-version {
  color: var(--text-tertiary);
  font-family: var(--font-family-mono);
}

.plugin-detail-author {
  color: var(--text-secondary);
}

.meta-sep {
  color: var(--text-tertiary);
}

.plugin-detail-source {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--text-tertiary);
}

.plugin-detail-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-3) var(--spacing-6);
  border-bottom: 1px solid var(--border-subtle);
}

.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border-radius: var(--radius-md);
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  cursor: pointer;
  transition: all var(--transition-fast);
  border: 1px solid transparent;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.action-btn--primary {
  background: var(--primary);
  color: white;
  border-color: var(--primary);
}

.action-btn--primary:hover:not(:disabled) {
  background: var(--primary-hover);
}

.action-btn--secondary {
  background: transparent;
  color: var(--text-secondary);
  border-color: var(--border-default);
}

.action-btn--secondary:hover:not(:disabled) {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.action-btn--danger {
  background: transparent;
  color: var(--error);
  border-color: var(--error);
}

.action-btn--danger:hover:not(:disabled) {
  background: var(--error-bg);
}

.plugin-detail-body {
  flex: 1;
  padding: var(--spacing-6);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-6);
}

.detail-section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.detail-section-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--text-h3);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  margin: 0;
}

.detail-description {
  font-size: var(--text-body-sm);
  color: var(--text-secondary);
  line-height: var(--leading-body-sm);
  margin: 0;
}

.detail-capabilities {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-2);
}

.capability-tag {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: var(--radius-md);
  background: var(--primary-bg);
  color: var(--primary);
  font-size: var(--text-caption);
  font-weight: 500;
}

.detail-nodes {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.workflow-node {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-2) var(--spacing-3);
  background: var(--bg-tertiary);
  border-radius: var(--radius-lg);
}

.node-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.node-name {
  font-size: var(--text-body-sm);
  color: var(--text-primary);
  font-weight: 500;
}

.node-type {
  margin-left: auto;
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  font-family: var(--font-family-mono);
}

.detail-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-2);
}

.detail-tag {
  display: inline-flex;
  align-items: center;
  padding: 3px 8px;
  border-radius: var(--radius-sm);
  background: var(--bg-hover);
  color: var(--text-secondary);
  font-size: var(--text-caption);
}

.slide-in-enter-active,
.slide-in-leave-active {
  transition: opacity var(--transition-normal), transform var(--transition-normal);
}

.slide-in-enter-from {
  opacity: 0;
  transform: translateX(12px);
}

.slide-in-leave-to {
  opacity: 0;
  transform: translateX(-12px);
}
</style>