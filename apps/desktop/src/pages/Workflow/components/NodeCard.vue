<template>
  <div :class="nodeClasses" :style="nodeStyle">
    <!-- 状态光晕 -->
    <div class="node-glow" :class="`glow-${status}`"></div>

    <div class="node-header">
      <div class="node-header-left">
        <div class="node-icon-wrap" :style="{ color: currentColor }">
          <svg
            viewBox="0 0 24 24"
            width="14"
            height="14"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path :d="typeIcon" />
          </svg>
        </div>
        <span class="node-title">{{ data.label || '未命名节点' }}</span>
      </div>
      <span :class="['node-status', `node-status--${status}`]"></span>
    </div>

    <div class="node-body">
      <p v-if="data.description" class="node-description">{{ data.description }}</p>
      <!-- 关键参数显示 -->
      <div v-if="keyParams.length > 0" class="node-params">
        <div v-for="param in keyParams" :key="param.label" class="node-param-item">
          <span class="param-label">{{ param.label }}:</span>
          <span class="param-value">{{ param.value }}</span>
        </div>
      </div>
    </div>

    <div class="node-ports">
      <div class="node-ports-left">
        <div
          v-for="port in inputPorts"
          :key="port.name"
          class="node-port node-port--input"
          :title="`${port.label}: ${port.type}`"
        >
          <Handle
            type="target"
            :position="Position.Left"
            :id="port.name"
            :class="['port-handle', `port-handle--${portTypeColorClass(port.type)}`]"
          />
          <span class="port-label">{{ port.label }}</span>
        </div>
      </div>
      <div class="node-ports-right">
        <div
          v-for="port in outputPorts"
          :key="port.name"
          class="node-port node-port--output"
          :title="`${port.label}: ${port.type}`"
        >
          <span class="port-label">{{ port.label }}</span>
          <Handle
            type="source"
            :position="Position.Right"
            :id="port.name"
            :class="['port-handle', `port-handle--${portTypeColorClass(port.type)}`]"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import { nodeTemplates } from '../config/nodeTemplates'

interface Port {
  name: string
  label: string
  type: string
}

interface NodeData {
  label: string
  description?: string
  nodeType: string
  status: 'idle' | 'running' | 'success' | 'error' | 'waiting'
  inputs: Port[]
  outputs: Port[]
  params?: Record<string, any>
  paramDefinitions?: { name: string; label: string }[]
  templateKey?: string
  category?: string
}

interface Props {
  id: string
  data: NodeData
  selected?: boolean
}

const props = defineProps<Props>()

const status = computed(() => props.data.status || 'idle')
const inputPorts = computed(() => props.data.inputs || [])
const outputPorts = computed(() => props.data.outputs || [])

// 找到对应模板
const template = computed(() => {
  return nodeTemplates.find(t => t.key === props.data.templateKey)
})

// 节点颜色
const currentColor = computed(() => {
  return template.value?.color || typeColorFromNodeType(props.data.nodeType)
})

// 关键参数：取前几个显示在节点上
const keyParams = computed(() => {
  const defs = props.data.paramDefinitions || []
  const params = props.data.params || {}
  if (defs.length === 0) return []

  // 优先显示关键参数
  const keyParamNames = ['model_version', 'model_type', 'model', 'epochs', 'epoch', 'batch_size', 'device', 'gpu_select', 'task_type', 'data_format', 'input_size', 'hidden_size', 'num_layers', 'model_size', 'variant', 'format', 'confidence', 'temperature', 'max_tokens', 'condition', 'max_iterations', 'max_retries', 'server_url', 'tool_name', 'data_path', 'file_path', 'python_version', 'repo_url', 'sumo_cfg', 'ros_version', 'port']

  const selected: { label: string; value: string }[] = []

  for (const name of keyParamNames) {
    const def = defs.find(d => d.name === name)
    if (def && params[name] !== undefined && params[name] !== '') {
      let displayValue = params[name]
      if (typeof displayValue === 'boolean') {
        displayValue = displayValue ? 'ON' : 'OFF'
      } else if (def.type === 'select' && def.options) {
        const opt = (def.options as any[]).find((o: any) => o.value === displayValue)
        if (opt) displayValue = (opt as any).label
      }
      selected.push({ label: def.label, value: String(displayValue) })
      if (selected.length >= 4) break
    }
  }

  return selected
})

const typeIconMap: Record<string, string> = {
  dataset: 'M3 3v18h18 M18.5 9h.01 M15.5 17h.01 M11.5 14h.01 M8.5 11h.01',
  vision: 'M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z M12 9a3 3 0 1 0 0 6 3 3 0 0 0 0-6z',
  nlp: 'M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z',
  speech: 'M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z M19 10v2a7 7 0 0 1-14 0v-2 M12 19v3',
  timeseries: 'M3 16.5 9 10.5 13 14.5 21 6.5 M21 6.5 13 14.5 9 10.5 3 16.5',
  logic: 'M4 4h16v16H4z M8 8h8 M8 12h8 M8 16h8',
  system: 'M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z',
  simulation: 'M9 3v2l6 8-6 8v2h12V3H9z',
  mcp: 'M13 2L3 14h9l-1 8 10-12h-9l1-8z',
  agent: 'M12 2a3 3 0 0 0-3 3v7a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3z M19 10v2a7 7 0 0 1-14 0v-2 M12 19v3',
  input: 'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z M14 2v6h6 M16 13H8 M16 17H8 M10 9H8',
  output: 'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z M14 2v6h6',
  deployment: 'M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z',
  training: 'M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z M2 12h3 M19 12h3 M12 2v3 M12 19v3',
  default: 'M12 2L2 7l10 5 10-5-10-5z M2 17l10 5 10-5 M2 12l10 5 10-5',
}

const typeIcon = computed(() => typeIconMap[props.data.nodeType] || typeIconMap.default)

function typeColorFromNodeType(nodeType: string): string {
  const colorMap: Record<string, string> = {
    dataset: 'var(--timeseries)',
    vision: 'var(--vision)',
    nlp: 'var(--nlp)',
    speech: 'var(--info)',
    timeseries: 'var(--timeseries)',
    logic: 'var(--logic)',
    system: 'var(--neutral)',
    simulation: 'var(--simulation)',
    mcp: 'var(--mcp)',
    agent: 'var(--agent)',
    deployment: 'var(--mcp)',
    training: 'var(--success)',
    input: 'var(--timeseries)',
    output: 'var(--mcp)',
  }
  return colorMap[nodeType] || 'var(--neutral)'
}

const nodeTypeColor = computed(() => {
  const colorMap: Record<string, string> = {
    dataset: 'timeseries',
    vision: 'vision',
    nlp: 'nlp',
    speech: 'info',
    timeseries: 'timeseries',
    logic: 'logic',
    system: 'system',
    simulation: 'simulation',
    mcp: 'mcp',
    agent: 'agent',
    deployment: 'mcp',
    training: 'success',
    input: 'timeseries',
    output: 'mcp',
  }
  return colorMap[props.data.nodeType] || 'default'
})

function portTypeColorClass(type: string): string {
  const map: Record<string, string> = {
    image: 'vision',
    tensor: 'nlp',
    dataset: 'timeseries',
    model: 'mcp',
    text: 'info',
    audio: 'warning',
    result: 'success',
    number: 'logic',
    json: 'system',
    trigger: 'agent',
    service: 'simulation',
    any: 'default',
  }
  return map[type] || 'default'
}

const nodeClasses = computed(() => [
  'node-card',
  `node-card--${nodeTypeColor.value}`,
  {
    'is-selected': props.selected,
    'is-running': status.value === 'running',
    'is-error': status.value === 'error',
    'is-success': status.value === 'success',
    'is-waiting': status.value === 'waiting',
  },
])

const nodeStyle = computed(() => ({
  '--node-color': currentColor.value,
}))
</script>

<style scoped>
.node-card {
  position: relative;
  min-width: 200px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-left: 3px solid var(--node-color, var(--neutral));
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-node);
  transition: all var(--transition-fast);
  font-family: var(--font-family-sans);
  cursor: grab;
  overflow: hidden;
}

.node-card:active {
  cursor: grabbing;
}

.node-card.is-selected {
  border-color: var(--primary);
  border-left-width: 3px;
  box-shadow: 0 0 0 2px rgba(139, 92, 246, 0.3), var(--shadow-node);
  transform: translateY(-1px);
}

.node-card.is-running {
  border-color: var(--info);
  border-left-color: var(--info);
  animation: node-pulse 2s ease-in-out infinite;
}

.node-card.is-error {
  border-color: var(--error);
  border-left-color: var(--error);
  animation: node-shake 0.5s ease-in-out;
}

.node-card.is-success {
  border-color: var(--success);
  border-left-color: var(--success);
}

.node-card.is-waiting {
  border-color: var(--warning);
  border-left-color: var(--warning);
  opacity: 0.85;
}

/* ===== 状态光晕 ===== */
.node-glow {
  position: absolute;
  inset: -1px;
  border-radius: var(--radius-lg);
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.3s ease;
}

.glow-running {
  opacity: 0.4;
  box-shadow: 0 0 12px 2px rgba(96, 165, 250, 0.5);
}

.glow-error {
  opacity: 0.5;
  box-shadow: 0 0 12px 2px rgba(239, 68, 68, 0.5);
}

.glow-success {
  opacity: 0.3;
  box-shadow: 0 0 12px 2px rgba(34, 197, 94, 0.4);
}

/* ===== 头部 ===== */
.node-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: rgba(0, 0, 0, 0.15);
  border-radius: 0 6px 0 0;
  position: relative;
  z-index: 1;
}

.node-header-left {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.node-icon-wrap {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.node-title {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* ===== 状态指示器 ===== */
.node-status {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
  position: relative;
  z-index: 1;
}

.node-status--idle {
  background: var(--neutral);
}

.node-status--running {
  background: var(--info);
  animation: status-pulse 1s ease-in-out infinite;
}

.node-status--success {
  background: var(--success);
}

.node-status--error {
  background: var(--error);
}

.node-status--waiting {
  background: var(--warning);
  opacity: 0.5;
}

/* ===== 内容区 ===== */
.node-body {
  padding: 8px 12px;
  position: relative;
  z-index: 1;
}

.node-description {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  margin: 0 0 6px 0;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* ===== 关键参数 ===== */
.node-params {
  display: flex;
  flex-direction: column;
  gap: 3px;
  padding-top: 4px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}

.node-param-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  line-height: 1.5;
}

.param-label {
  color: var(--text-tertiary);
  flex-shrink: 0;
}

.param-value {
  color: var(--node-color, var(--text-secondary));
  font-weight: var(--font-medium);
  font-family: var(--font-family-mono);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* ===== 端口 ===== */
.node-ports {
  display: flex;
  justify-content: space-between;
  padding: 4px 0;
  position: relative;
  z-index: 1;
}

.node-ports-left,
.node-ports-right {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.node-port {
  position: relative;
  display: flex;
  align-items: center;
  height: 20px;
}

.node-port--input {
  padding-left: 4px;
  flex-direction: row;
}

.node-port--output {
  padding-right: 4px;
  flex-direction: row-reverse;
}

.port-label {
  font-size: 10px;
  color: var(--text-tertiary);
  user-select: none;
  padding: 0 4px;
  white-space: nowrap;
}

/* ===== Handle 样式 ===== */
:deep(.port-handle) {
  width: 8px !important;
  height: 8px !important;
  border: 2px solid var(--bg-tertiary) !important;
  background: var(--neutral) !important;
  border-radius: 50% !important;
  transition: all var(--transition-fast) !important;
}

:deep(.port-handle:hover) {
  transform: scale(1.5);
  cursor: crosshair;
}

:deep(.port-handle--vision) { background: var(--vision) !important; }
:deep(.port-handle--nlp) { background: var(--nlp) !important; }
:deep(.port-handle--timeseries) { background: var(--timeseries) !important; }
:deep(.port-handle--info) { background: var(--info) !important; }
:deep(.port-handle--warning) { background: var(--warning) !important; }
:deep(.port-handle--success) { background: var(--success) !important; }
:deep(.port-handle--error) { background: var(--error) !important; }
:deep(.port-handle--logic) { background: var(--logic) !important; }
:deep(.port-handle--system) { background: var(--neutral) !important; }
:deep(.port-handle--simulation) { background: var(--simulation) !important; }
:deep(.port-handle--mcp) { background: var(--mcp) !important; }
:deep(.port-handle--agent) { background: var(--agent) !important; }
:deep(.port-handle--default) { background: var(--neutral) !important; }

/* ===== 动画 ===== */
@keyframes node-pulse {
  0%, 100% {
    box-shadow: 0 0 4px rgba(96, 165, 250, 0.3), var(--shadow-node);
  }
  50% {
    box-shadow: 0 0 14px rgba(96, 165, 250, 0.5), var(--shadow-node);
  }
}

@keyframes node-shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-3px); }
  75% { transform: translateX(3px); }
}

@keyframes status-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}
</style>