<template>
  <div class="node-panel" :class="{ 'is-collapsed': collapsed }">
    <div v-if="!collapsed" class="panel-content">
      <div class="panel-search">
        <svg
          class="search-icon"
          viewBox="0 0 24 24"
          width="14"
          height="14"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
        >
          <circle cx="11" cy="11" r="8" />
          <path d="m21 21-4.3-4.3" />
        </svg>
        <input
          v-model="searchQuery"
          class="search-input"
          type="text"
          placeholder="搜索节点..."
          @input="onSearchInput"
        />
      </div>

      <div class="panel-categories">
        <div
          v-for="category in filteredCategories"
          :key="category.key"
          class="category"
        >
          <button
            class="category-header"
            @click="toggleCategory(category.key)"
          >
            <svg
              class="category-arrow"
              :class="{ open: expandedCategories.has(category.key) }"
              viewBox="0 0 24 24"
              width="12"
              height="12"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
            >
              <path d="m9 18 6-6-6-6" />
            </svg>
            <div
              class="category-color-dot"
              :style="{ background: category.color }"
            ></div>
            <span class="category-label">{{ category.label }}</span>
            <span class="category-count">{{ category.nodes.length }}</span>
          </button>
          <div v-show="expandedCategories.has(category.key)" class="category-nodes">
            <div
              v-for="node in category.nodes"
              :key="node.key"
              class="node-item"
              draggable="true"
              @dragstart="onDragStart($event, node)"
              @click="emits('addNode', node)"
              :title="node.description"
            >
              <div class="node-color" :style="{ background: node.color }"></div>
              <div class="node-info">
                <span class="node-name">{{ node.label }}</span>
                <span class="node-desc">{{ node.description }}</span>
              </div>
              <div class="node-badges">
                <span v-if="node.inputs.length" class="node-badge node-badge-in" :title="`${node.inputs.length} 个输入`">
                  {{ node.inputs.length }}in
                </span>
                <span v-if="node.outputs.length" class="node-badge node-badge-out" :title="`${node.outputs.length} 个输出`">
                  {{ node.outputs.length }}out
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <button class="panel-toggle" @click="collapsed = !collapsed" :title="collapsed ? '展开节点面板' : '收起节点面板'">
      <svg
        viewBox="0 0 24 24"
        width="16"
        height="16"
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
      >
        <path v-if="collapsed" d="m15 18-6-6 6-6" />
        <path v-else d="m9 18 6-6-6-6" />
      </svg>
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { nodeTemplates } from '../config/nodeTemplates'
import {
  NODE_CATEGORY_LABELS,
  NODE_CATEGORY_COLORS,
  type NodeTemplate,
  type NodeCategory,
} from '../types/workflow'

const collapsed = ref(false)
const searchQuery = ref('')

// 按分类分组节点模板
const categoryOrder: NodeCategory[] = [
  'ai-vision', 'ai-nlp', 'ai-timeseries', 'ai-audio',
  'data', 'training', 'deployment',
  'logic', 'system', 'simulation', 'mcp',
]

const categoryIconMap: Record<NodeCategory, string> = {
  'ai-vision': 'M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z M12 9a3 3 0 1 0 0 6 3 3 0 0 0 0-6z',
  'ai-nlp': 'M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z',
  'ai-timeseries': 'M3 16.5 9 10.5 13 14.5 21 6.5 M21 6.5 13 14.5 9 10.5 3 16.5',
  'ai-audio': 'M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z M19 10v2a7 7 0 0 1-14 0v-2 M12 19v3',
  'data': 'M3 3v18h18 M18.5 9h.01 M15.5 17h.01 M11.5 14h.01 M8.5 11h.01',
  'training': 'M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z M2 12h3 M19 12h3 M12 2v3 M12 19v3',
  'deployment': 'M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z',
  'logic': 'M4 4h16v16H4z M8 8h8 M8 12h8 M8 16h8',
  'system': 'M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z',
  'simulation': 'M9 3v2l6 8-6 8v2h12V3H9z',
  'mcp': 'M13 2L3 14h9l-1 8 10-12h-9l1-8z',
  'input': 'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z M14 2v6h6 M16 13H8 M16 17H8 M10 9H8',
  'output': 'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z M14 2v6h6',
}

const categories = computed(() => {
  const grouped: Record<string, NodeTemplate[]> = {}
  for (const t of nodeTemplates) {
    if (!grouped[t.category]) grouped[t.category] = []
    grouped[t.category].push(t)
  }
  return categoryOrder
    .filter(cat => grouped[cat]?.length > 0)
    .map(cat => ({
      key: cat,
      label: NODE_CATEGORY_LABELS[cat] || cat,
      icon: categoryIconMap[cat],
      color: NODE_CATEGORY_COLORS[cat] || 'var(--neutral)',
      nodes: grouped[cat] || [],
    }))
})

const expandedCategories = ref(new Set<string>(categoryOrder))

const filteredCategories = computed(() => {
  if (!searchQuery.value.trim()) {
    return categories.value
  }
  const query = searchQuery.value.toLowerCase()
  return categories.value
    .map(cat => ({
      ...cat,
      nodes: cat.nodes.filter(n =>
        n.label.toLowerCase().includes(query) ||
        n.description.toLowerCase().includes(query) ||
        n.key.toLowerCase().includes(query)
      ),
    }))
    .filter(cat => cat.nodes.length > 0)
})

const emits = defineEmits<{
  addNode: [node: NodeTemplate]
}>()

function toggleCategory(key: string): void {
  const next = new Set(expandedCategories.value)
  if (next.has(key)) {
    next.delete(key)
  } else {
    next.add(key)
  }
  expandedCategories.value = next
}

function onDragStart(event: DragEvent, node: NodeTemplate): void {
  event.dataTransfer!.setData('application/json', JSON.stringify(node))
  event.dataTransfer!.effectAllowed = 'copy'
}

function onSearchInput(): void {
  if (searchQuery.value.trim()) {
    const next = new Set<string>()
    for (const cat of filteredCategories.value) {
      next.add(cat.key)
    }
    expandedCategories.value = next
  }
}
</script>

<style scoped>
.node-panel {
  position: relative;
  width: 260px;
  height: 100%;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-subtle);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: width var(--transition-normal);
  flex-shrink: 0;
}

.node-panel.is-collapsed {
  width: 36px;
}

.panel-content {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

/* ===== 搜索 ===== */
.panel-search {
  position: relative;
  padding: var(--spacing-3) var(--spacing-3);
  border-bottom: 1px solid var(--border-subtle);
}

.search-icon {
  position: absolute;
  left: 20px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-tertiary);
  pointer-events: none;
}

.search-input {
  width: 100%;
  height: 32px;
  padding: 0 12px 0 32px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  font-size: var(--text-body-sm);
  font-family: var(--font-family-sans);
  outline: none;
  transition: border-color var(--transition-fast);
}

.search-input:focus {
  border-color: var(--primary);
}

.search-input::placeholder {
  color: var(--text-tertiary);
}

/* ===== 分类列表 ===== */
.panel-categories {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-2) 0;
}

.panel-categories::-webkit-scrollbar {
  width: 4px;
}

.panel-categories::-webkit-scrollbar-thumb {
  background: var(--border-subtle);
  border-radius: 2px;
}

.category {
  border-bottom: 1px solid var(--border-subtle);
}

.category:last-child {
  border-bottom: none;
}

.category-header {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
  padding: 8px var(--spacing-3);
  background: none;
  border: none;
  color: var(--text-secondary);
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  cursor: pointer;
  transition: all var(--transition-fast);
  font-family: var(--font-family-sans);
  text-align: left;
}

.category-header:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.category-arrow {
  transition: transform var(--transition-fast);
  flex-shrink: 0;
  opacity: 0.5;
}

.category-arrow.open {
  transform: rotate(90deg);
}

.category-color-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}

.category-label {
  flex: 1;
}

.category-count {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  font-weight: var(--font-regular);
  background: var(--bg-tertiary);
  padding: 1px 6px;
  border-radius: 8px;
}

/* ===== 节点列表 ===== */
.category-nodes {
  padding: 0 var(--spacing-2) var(--spacing-2);
}

.node-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-radius: var(--radius-sm);
  cursor: grab;
  transition: all var(--transition-fast);
  user-select: none;
}

.node-item:hover {
  background: var(--bg-hover);
}

.node-item:active {
  cursor: grabbing;
  background: var(--bg-active);
}

.node-color {
  width: 4px;
  height: 28px;
  border-radius: 2px;
  flex-shrink: 0;
}

.node-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.node-name {
  font-size: var(--text-body-sm);
  color: var(--text-primary);
  font-weight: var(--font-medium);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.node-desc {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.node-badges {
  display: flex;
  gap: 3px;
  flex-shrink: 0;
}

.node-badge {
  font-size: 9px;
  padding: 1px 5px;
  border-radius: 4px;
  font-weight: var(--font-medium);
}

.node-badge-in {
  background: rgba(96, 165, 250, 0.15);
  color: var(--info);
}

.node-badge-out {
  background: rgba(34, 197, 94, 0.15);
  color: var(--success);
}

/* ===== 折叠按钮 ===== */
.panel-toggle {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  width: 20px;
  height: 48px;
  border: 1px solid var(--border-subtle);
  border-radius: 4px;
  background: var(--bg-secondary);
  color: var(--text-tertiary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-fast);
  z-index: 10;
}

.panel-toggle:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}
</style>