<template>
  <div class="project-explorer">
    <div class="explorer-header">
      <span class="explorer-title">Explorer</span>
      <button
        class="explorer-action"
        title="新建项目"
        @click="$emit('new-project')"
      >
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M12 5v14M5 12h14" />
        </svg>
      </button>
    </div>

    <div v-if="!currentProject" class="explorer-empty">
      <svg viewBox="0 0 24 24" width="32" height="32" fill="none" stroke="currentColor" stroke-width="1.5">
        <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" />
      </svg>
      <span>选择或创建项目</span>
    </div>

    <div v-else class="explorer-content">
      <div class="project-info">
        <div class="project-name">{{ currentProject.name }}</div>
        <div class="project-type">{{ projectTypeLabel }}</div>
      </div>

      <div class="explorer-tree">
        <div
          v-for="node in treeNodes"
          :key="node.id"
          class="tree-node"
          :class="{ active: activeExplorerNode === node.id }"
          @click="handleNodeClick(node.id)"
        >
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
            <path :d="node.icon" />
          </svg>
          <span class="node-label">{{ node.label }}</span>
          <span v-if="node.count !== undefined" class="node-count">{{ node.count }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useProjectStore, type ExplorerNodeType } from '@/stores/project'

const store = useProjectStore()

defineEmits<{
  'new-project': []
}>()

const currentProject = computed(() => store.currentProject)
const activeExplorerNode = computed(() => store.activeExplorerNode)

const projectTypeLabel = computed(() => {
  if (!currentProject.value) return ''
  const typeMap: Record<string, string> = {
    detection: '目标检测',
    classification: '图像分类',
    segmentation: '语义分割',
    timeseries: '时序预测',
    custom: '自定义',
  }
  return typeMap[currentProject.value.type] || currentProject.value.type
})

interface TreeNode {
  id: ExplorerNodeType
  label: string
  icon: string
  count?: number
}

const treeNodes = computed<TreeNode[]>(() => {
  if (!currentProject.value) return []
  const p = currentProject.value
  return [
    { id: 'dashboard', label: '概览', icon: 'M3 3h7v7H3V3zm0 11h7v7H3v-7zm11-11h7v7h-7V3zm0 11h7v7h-7v-7z' },
    { id: 'workflows', label: 'Workflows', icon: 'M6 3h3v6H6V3zm0 12h3v6H6v-6zm9-12h3v6h-3V3zm0 12h3v6h-3v-6zm-9 0V9m3 9v-3m3 3v-3m3 3V9', count: p.workflows.length },
    { id: 'datasets', label: 'Datasets', icon: 'M4 7V4h16v3M9 20h6M12 4v16', count: p.datasets.length },
    { id: 'models', label: 'Models', icon: 'M12 2L2 7l10 5 10-5-10-5zm0 22L2 17l10-5 10 5-10 5z', count: p.models.length },
    { id: 'experiments', label: 'Experiments', icon: 'M9 3h6v11l-3 3-3-3V3z', count: p.experiments.length },
    { id: 'environment', label: 'Environment', icon: 'M12 2L2 7l10 5 10-5-10-5zm0 22L2 17l10-5 10 5-10 5z' },
    { id: 'outputs', label: 'Outputs', icon: 'M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z', count: p.outputs.length },
    { id: 'logs', label: 'Logs', icon: 'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z', count: p.logs.length },
  ]
})

function handleNodeClick(nodeId: ExplorerNodeType) {
  store.setExplorerNode(nodeId)
}
</script>

<style scoped>
.project-explorer {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg-secondary);
}

.explorer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-3) var(--spacing-4);
  border-bottom: 1px solid var(--border-subtle);
}

.explorer-title {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.explorer-action {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: var(--radius-sm);
  background: transparent;
  border: none;
  color: var(--text-tertiary);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.explorer-action:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.explorer-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-3);
  color: var(--text-tertiary);
}

.explorer-empty span {
  font-size: var(--text-body-sm);
}

.explorer-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.project-info {
  padding: var(--spacing-3) var(--spacing-4);
  border-bottom: 1px solid var(--border-subtle);
}

.project-name {
  font-size: var(--text-body);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  margin-bottom: var(--spacing-1);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.project-type {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.explorer-tree {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-2);
}

.tree-node {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  height: 32px;
  padding: 0 var(--spacing-3);
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.tree-node:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.tree-node.active {
  background: var(--primary-bg);
  color: var(--primary);
}

.tree-node svg {
  flex-shrink: 0;
  opacity: 0.7;
}

.tree-node.active svg {
  opacity: 1;
}

.node-label {
  flex: 1;
  font-size: var(--text-body-sm);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-count {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  background: var(--bg-tertiary);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  min-width: 20px;
  text-align: center;
}
</style>