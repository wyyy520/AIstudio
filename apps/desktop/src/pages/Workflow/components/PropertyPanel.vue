<template>
  <div class="property-panel" :class="{ 'is-collapsed': collapsed }">
    <div v-if="!collapsed" class="panel-content">
      <div class="panel-header">
        <h4 class="panel-title">{{ selectedNode ? '节点属性' : '工作流属性' }}</h4>
        <div v-if="selectedNode" class="node-status-badge" :class="`status-${nodeStatus}`">
          {{ statusLabel }}
        </div>
      </div>

      <!-- ===== 节点属性 ===== -->
      <div v-if="selectedNode" class="panel-body">
        <!-- 节点名称 -->
        <div class="prop-group">
          <label class="prop-label">节点名称</label>
          <div class="prop-input-wrapper">
            <input
              class="prop-input"
              type="text"
              :value="selectedNode.data.label"
              @input="onLabelChange"
              placeholder="节点名称"
            />
            <span class="prop-template-tag" :style="{ borderColor: templateColor, color: templateColor }">
              {{ templateLabel }}
            </span>
          </div>
        </div>

        <!-- 节点类型 -->
        <div class="prop-group">
          <label class="prop-label">节点类型</label>
          <div class="prop-type-row">
            <span class="prop-type-badge" :style="{ background: templateColor }">
              {{ nodeTypeLabel }}
            </span>
            <span class="prop-category-tag">{{ categoryLabel }}</span>
          </div>
        </div>

        <!-- 描述 -->
        <div class="prop-group">
          <label class="prop-label">描述</label>
          <div class="prop-description-text">{{ selectedNode.data.description || '无描述' }}</div>
        </div>

        <div class="prop-divider"></div>

        <!-- 动态参数区域 -->
        <template v-if="paramCategories.length > 0">
          <div
            v-for="category in paramCategories"
            :key="category.name"
            class="prop-category"
          >
            <button
              v-if="category.name !== '__default'"
              class="prop-category-header"
              @click="toggleCategory(category.name)"
            >
              <svg
                class="category-arrow"
                :class="{ open: expandedCategories.has(category.name) }"
                viewBox="0 0 24 24"
                width="12"
                height="12"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
              >
                <path d="m9 18 6-6-6-6" />
              </svg>
              <span>{{ category.name }}</span>
              <span class="category-hint">{{ category.params.length }} 项</span>
            </button>
            <div v-show="category.name === '__default' || expandedCategories.has(category.name)" class="prop-category-body">
              <div
                v-for="param in category.params"
                :key="param.name"
                class="prop-field"
                :class="{ 'has-error': getParamError(param.name) }"
              >
                <label class="prop-field-label">
                  {{ param.label }}
                  <span v-if="param.required" class="required-star">*</span>
                </label>

                <!-- Text 输入 -->
                <input
                  v-if="param.type === 'text'"
                  class="prop-input"
                  type="text"
                  :value="getParamValue(param.name)"
                  :placeholder="param.placeholder || ''"
                  @input="onParamChange(param.name, $event)"
                />

                <!-- Number 输入 -->
                <div v-else-if="param.type === 'number'" class="prop-number-row">
                  <input
                    class="prop-input prop-input-number"
                    type="number"
                    :value="getParamValue(param.name)"
                    :min="param.min"
                    :max="param.max"
                    :step="param.step || 1"
                    @input="onParamChange(param.name, $event)"
                    @blur="validateParam(param)"
                  />
                  <span v-if="param.hint" class="prop-hint">{{ param.hint }}</span>
                </div>

                <!-- Select 下拉 -->
                <select
                  v-else-if="param.type === 'select'"
                  class="prop-select"
                  :value="getParamValue(param.name)"
                  @change="onParamChange(param.name, $event)"
                >
                  <option
                    v-for="opt in param.options"
                    :key="String(opt.value)"
                    :value="opt.value"
                  >
                    {{ opt.label }}
                  </option>
                </select>

                <!-- Switch 开关 -->
                <label v-else-if="param.type === 'switch'" class="prop-switch">
                  <input
                    type="checkbox"
                    :checked="getParamValue(param.name)"
                    @change="onParamChange(param.name, $event)"
                  />
                  <span class="prop-switch-slider"></span>
                  <span class="prop-switch-label">{{ getParamValue(param.name) ? '开启' : '关闭' }}</span>
                </label>

                <!-- Slider 滑块 -->
                <div v-else-if="param.type === 'slider'" class="prop-slider-row">
                  <input
                    class="prop-slider"
                    type="range"
                    :value="getParamValue(param.name)"
                    :min="param.min || 0"
                    :max="param.max || 1"
                    :step="param.step || 0.01"
                    @input="onParamChange(param.name, $event)"
                  />
                  <span class="prop-slider-value">{{ getParamValue(param.name) }}</span>
                </div>

                <!-- Multi-Select -->
                <div v-else-if="param.type === 'multi-select'" class="prop-multiselect">
                  <label
                    v-for="opt in param.options"
                    :key="String(opt.value)"
                    class="prop-multiselect-item"
                  >
                    <input
                      type="checkbox"
                      :checked="isMultiSelected(param.name, opt.value)"
                      @change="onMultiSelectChange(param.name, opt.value, $event)"
                    />
                    <span>{{ opt.label }}</span>
                  </label>
                </div>

                <!-- Device Select -->
                <select
                  v-else-if="param.type === 'device-select'"
                  class="prop-select"
                  :value="getParamValue(param.name)"
                  @change="onParamChange(param.name, $event)"
                >
                  <option value="auto">Auto</option>
                  <option value="0">RTX 3060</option>
                  <option value="1">RTX 4090</option>
                  <option value="cpu">CPU</option>
                </select>

                <div v-if="getParamError(param.name)" class="prop-field-error">
                  {{ getParamError(param.name) }}
                </div>
              </div>
            </div>
          </div>
        </template>

        <div v-else class="prop-empty">
          <span>该节点无可配置参数</span>
        </div>

        <div class="prop-divider"></div>

        <!-- 端口信息 -->
        <div class="prop-group">
          <label class="prop-label">输入端口</label>
          <div v-if="selectedNode.data.inputs?.length" class="prop-ports">
            <div v-for="port in selectedNode.data.inputs" :key="port.name" class="prop-port-item">
              <span class="port-dot" :style="{ background: portTypeColor(port.type) }"></span>
              <span class="port-name">{{ port.label }}</span>
              <span class="port-type-tag" :style="{ background: portTypeColor(port.type) + '22', color: portTypeColor(port.type) }">
                {{ port.type }}
              </span>
            </div>
          </div>
          <div v-else class="prop-empty-small">无输入端口</div>
        </div>

        <div class="prop-group">
          <label class="prop-label">输出端口</label>
          <div v-if="selectedNode.data.outputs?.length" class="prop-ports">
            <div v-for="port in selectedNode.data.outputs" :key="port.name" class="prop-port-item">
              <span class="port-dot" :style="{ background: portTypeColor(port.type) }"></span>
              <span class="port-name">{{ port.label }}</span>
              <span class="port-type-tag" :style="{ background: portTypeColor(port.type) + '22', color: portTypeColor(port.type) }">
                {{ port.type }}
              </span>
            </div>
          </div>
          <div v-else class="prop-empty-small">无输出端口</div>
        </div>

        <div class="prop-divider"></div>

        <!-- 操作按钮 -->
        <div class="prop-actions">
          <AppButton
            type="danger"
            size="small"
            block
            @click="deleteNode"
          >
            删除节点
          </AppButton>
        </div>
      </div>

      <!-- ===== 工作流属性 ===== -->
      <div v-else class="panel-body">
        <div class="prop-group">
          <label class="prop-label">工作流名称</label>
          <input class="prop-input" type="text" v-model="workflowName" />
        </div>

        <div class="prop-group">
          <label class="prop-label">描述</label>
          <textarea class="prop-textarea" rows="3" placeholder="描述这个工作流..." v-model="workflowDescription"></textarea>
        </div>

        <div class="prop-divider"></div>

        <div class="prop-group">
          <label class="prop-label">运行模式</label>
          <select class="prop-select" v-model="runMode">
            <option value="sync">同步执行</option>
            <option value="stream">流式执行</option>
            <option value="debug">调试模式</option>
          </select>
        </div>

        <div class="prop-group">
          <label class="prop-label">超时时间</label>
          <select class="prop-select" v-model="timeout">
            <option value="300">5 分钟</option>
            <option value="600">10 分钟</option>
            <option value="1800">30 分钟</option>
            <option value="3600">1 小时</option>
          </select>
        </div>

        <div class="prop-divider"></div>

        <div class="prop-group">
          <label class="prop-label">工作流统计</label>
          <div class="prop-stats">
            <div class="stat-item">
              <span class="stat-label">节点数</span>
              <span class="stat-value">{{ totalNodes }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">连接数</span>
              <span class="stat-value">{{ totalEdges }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <button class="panel-toggle" @click="collapsed = !collapsed" :title="collapsed ? '展开属性面板' : '收起属性面板'">
      <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
        <path v-if="collapsed" d="m15 18-6-6 6-6" />
        <path v-else d="m9 18 6-6-6-6" />
      </svg>
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { Node } from '@vue-flow/core'
import {
  NODE_CATEGORY_LABELS,
  PORT_TYPE_COLORS,
  NODE_STATUS_COLORS,
  type ParamDefinition,
  type PortType,
  type WorkflowNodeData,
} from '../types/workflow'
import { nodeTemplates } from '../config/nodeTemplates'
import AppButton from '@/components/AppButton/AppButton.vue'

interface Props {
  selectedNode?: Node | null
  totalNodes?: number
  totalEdges?: number
}

const props = withDefaults(defineProps<Props>(), {
  selectedNode: null,
  totalNodes: 0,
  totalEdges: 0,
})

const collapsed = ref(false)
const workflowName = ref('未命名工作流')
const workflowDescription = ref('')
const runMode = ref('sync')
const timeout = ref('300')
const expandedCategories = ref(new Set<string>())
const paramErrors = ref<Record<string, string>>({})

const emits = defineEmits<{
  'update-node-label': [label: string]
  'update-node-params': [params: Record<string, any>]
  'delete-node': []
  'close': []
}>()

// ===== 计算属性 =====

const nodeData = computed<WorkflowNodeData | null>(() => {
  return (props.selectedNode?.data as WorkflowNodeData) || null
})

const nodeStatus = computed(() => {
  return nodeData.value?.status || 'idle'
})

const statusLabel = computed(() => {
  const labels: Record<string, string> = {
    idle: '空闲',
    running: '运行中',
    success: '成功',
    error: '失败',
    waiting: '等待中',
  }
  return labels[nodeStatus.value] || '空闲'
})

const templateKey = computed(() => {
  return nodeData.value?.templateKey || ''
})

const nodeTemplate = computed(() => {
  return nodeTemplates.find(t => t.key === templateKey.value)
})

const nodeTypeLabel = computed(() => {
  return nodeTemplate.value?.label || '未知节点'
})

const templateLabel = computed(() => {
  return nodeTemplate.value?.label || '未知'
})

const templateColor = computed(() => {
  return nodeTemplate.value?.color || 'var(--neutral)'
})

const categoryLabel = computed(() => {
  if (!nodeData.value) return ''
  return NODE_CATEGORY_LABELS[nodeData.value.category] || ''
})

// 参数按分类分组
const paramCategories = computed(() => {
  const defs = nodeData.value?.paramDefinitions || []
  if (defs.length === 0) return []

  const groups: Record<string, ParamDefinition[]> = {}
  for (const param of defs) {
    const cat = param.category || '__default'
    if (!groups[cat]) groups[cat] = []
    groups[cat].push(param)
  }

  // 保持分类顺序
  return Object.entries(groups).map(([name, params]) => ({ name, params }))
})

// ===== 方法 =====

function getParamValue(paramName: string): any {
  return nodeData.value?.params?.[paramName] ?? getDefaultValue(paramName)
}

function getDefaultValue(paramName: string): any {
  const def = nodeData.value?.paramDefinitions?.find(p => p.name === paramName)
  return def?.default
}

function getParamError(paramName: string): string | null {
  return paramErrors.value[paramName] || null
}

function onLabelChange(e: Event): void {
  const target = e.target as HTMLInputElement
  emits('update-node-label', target.value)
}

function onParamChange(paramName: string, event: Event): void {
  const target = event.target as HTMLInputElement | HTMLSelectElement
  let value: any

  if (target.type === 'checkbox') {
    value = (target as HTMLInputElement).checked
  } else if (target.type === 'range') {
    value = parseFloat(target.value)
  } else if (target.type === 'number') {
    value = target.value === '' ? '' : parseFloat(target.value)
  } else {
    value = target.value
  }

  if (nodeData.value) {
    const newParams = { ...nodeData.value.params, [paramName]: value }
    if (props.selectedNode) {
      props.selectedNode.data.params = newParams
    }
    emits('update-node-params', newParams)
  }

  // 清除该字段的错误
  if (paramErrors.value[paramName]) {
    const next = { ...paramErrors.value }
    delete next[paramName]
    paramErrors.value = next
  }
}

function validateParam(param: ParamDefinition): void {
  const value = getParamValue(param.name)

  if (param.required && (value === undefined || value === '' || value === null)) {
    paramErrors.value = { ...paramErrors.value, [param.name]: `${param.label} 为必填项` }
    return
  }

  if (param.validation) {
    const { rule, message } = param.validation
    if (rule === '> 0' && parseFloat(value) <= 0) {
      paramErrors.value = { ...paramErrors.value, [param.name]: message }
      return
    }
  }

  if (param.type === 'number') {
    const num = parseFloat(value)
    if (isNaN(num)) {
      paramErrors.value = { ...paramErrors.value, [param.name]: `${param.label} 必须是有效数字` }
      return
    }
    if (param.min !== undefined && num < param.min) {
      paramErrors.value = { ...paramErrors.value, [param.name]: `${param.label} 最小值为 ${param.min}` }
      return
    }
    if (param.max !== undefined && num > param.max) {
      paramErrors.value = { ...paramErrors.value, [param.name]: `${param.label} 最大值为 ${param.max}` }
      return
    }
  }

  const next = { ...paramErrors.value }
  delete next[param.name]
  paramErrors.value = next
}

function isMultiSelected(paramName: string, value: any): boolean {
  const current = getParamValue(paramName)
  return Array.isArray(current) && current.includes(value)
}

function onMultiSelectChange(paramName: string, value: any, event: Event): void {
  const checked = (event.target as HTMLInputElement).checked
  const current = getParamValue(paramName) || []
  let newValue: any[]
  if (checked) {
    newValue = [...current, value]
  } else {
    newValue = current.filter((v: any) => v !== value)
  }
  if (nodeData.value) {
    const newParams = { ...nodeData.value.params, [paramName]: newValue }
    if (props.selectedNode) {
      props.selectedNode.data.params = newParams
    }
    emits('update-node-params', newParams)
  }
}

function toggleCategory(name: string): void {
  const next = new Set(expandedCategories.value)
  if (next.has(name)) {
    next.delete(name)
  } else {
    next.add(name)
  }
  expandedCategories.value = next
}

function portTypeColor(type: string): string {
  return PORT_TYPE_COLORS[type as PortType] || 'var(--neutral)'
}

function deleteNode(): void {
  emits('delete-node')
}

// 当选中节点变化时，自动展开所有分类
watch(() => props.selectedNode?.id, () => {
  expandedCategories.value = new Set()
  paramErrors.value = {}
  // 如果节点使用 collapsible 布局，自动展开第一个分类
  if (nodeTemplate.value?.paramsLayout === 'collapsible') {
    const cats = paramCategories.value
    if (cats.length > 0) {
      expandedCategories.value.add(cats[0].name)
    }
  }
})
</script>

<style scoped>
.property-panel {
  position: relative;
  width: 360px;
  height: 100%;
  background: var(--bg-secondary);
  border-left: 1px solid var(--border-subtle);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: width var(--transition-normal);
  flex-shrink: 0;
}

.property-panel.is-collapsed {
  width: 36px;
}

.panel-content {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

/* ===== 头部 ===== */
.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-3) var(--spacing-4);
  border-bottom: 1px solid var(--border-subtle);
}

.panel-title {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  line-height: var(--leading-body-sm);
}

.node-status-badge {
  font-size: var(--text-caption);
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  font-weight: var(--font-medium);
}

.node-status-badge.status-idle {
  background: rgba(156, 163, 175, 0.15);
  color: var(--text-tertiary);
}

.node-status-badge.status-running {
  background: rgba(96, 165, 250, 0.15);
  color: var(--info);
}

.node-status-badge.status-success {
  background: rgba(34, 197, 94, 0.15);
  color: var(--success);
}

.node-status-badge.status-error {
  background: rgba(239, 68, 68, 0.15);
  color: var(--error);
}

.node-status-badge.status-waiting {
  background: rgba(251, 191, 36, 0.15);
  color: var(--warning);
}

/* ===== 内容 ===== */
.panel-body {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-4);
}

.panel-body::-webkit-scrollbar {
  width: 4px;
}

.panel-body::-webkit-scrollbar-thumb {
  background: var(--border-subtle);
  border-radius: 2px;
}

/* ===== 属性组 ===== */
.prop-group {
  margin-bottom: var(--spacing-4);
}

.prop-label {
  display: block;
  font-size: var(--text-caption);
  font-weight: var(--font-semibold);
  color: var(--text-secondary);
  margin-bottom: var(--spacing-2);
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.prop-input-wrapper {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.prop-input {
  width: 100%;
  height: 32px;
  padding: 0 10px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  font-size: var(--text-body-sm);
  font-family: var(--font-family-sans);
  outline: none;
  transition: border-color var(--transition-fast);
}

.prop-input:focus {
  border-color: var(--primary);
}

.prop-input-number {
  font-family: var(--font-family-mono);
}

.prop-template-tag {
  display: inline-flex;
  align-self: flex-start;
  font-size: var(--text-caption);
  padding: 1px 8px;
  border: 1px solid;
  border-radius: var(--radius-sm);
  font-weight: var(--font-medium);
}

.prop-type-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.prop-type-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  border-radius: var(--radius-sm);
  font-size: var(--text-caption);
  color: #fff;
  font-weight: var(--font-medium);
}

.prop-category-tag {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.prop-description-text {
  font-size: var(--text-body-sm);
  color: var(--text-tertiary);
  line-height: 1.5;
  padding: 8px;
  background: var(--bg-tertiary);
  border-radius: var(--radius-sm);
}

.prop-textarea {
  width: 100%;
  padding: 8px 10px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  font-size: var(--text-body-sm);
  font-family: var(--font-family-sans);
  outline: none;
  resize: vertical;
}

.prop-textarea:focus {
  border-color: var(--primary);
}

.prop-select {
  width: 100%;
  height: 32px;
  padding: 0 10px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  font-size: var(--text-body-sm);
  font-family: var(--font-family-sans);
  outline: none;
  cursor: pointer;
  transition: border-color var(--transition-fast);
}

.prop-select:focus {
  border-color: var(--primary);
}

.prop-divider {
  height: 1px;
  background: var(--border-subtle);
  margin: var(--spacing-4) 0;
}

/* ===== 参数分类 ===== */
.prop-category {
  margin-bottom: var(--spacing-3);
}

.prop-category-header {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
  padding: 8px 10px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  cursor: pointer;
  transition: all var(--transition-fast);
  font-family: var(--font-family-sans);
}

.prop-category-header:hover {
  background: var(--bg-hover);
}

.category-arrow {
  transition: transform var(--transition-fast);
  flex-shrink: 0;
}

.category-arrow.open {
  transform: rotate(90deg);
}

.category-hint {
  margin-left: auto;
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  font-weight: var(--font-regular);
}

.prop-category-body {
  padding: var(--spacing-2) 0 var(--spacing-2) var(--spacing-2);
}

/* ===== 参数字段 ===== */
.prop-field {
  margin-bottom: var(--spacing-3);
}

.prop-field:last-child {
  margin-bottom: 0;
}

.prop-field.has-error .prop-input,
.prop-field.has-error .prop-select {
  border-color: var(--error);
}

.prop-field-label {
  display: block;
  font-size: var(--text-caption);
  color: var(--text-secondary);
  margin-bottom: 4px;
  font-weight: var(--font-medium);
}

.required-star {
  color: var(--error);
  margin-left: 2px;
}

.prop-field-error {
  font-size: var(--text-caption);
  color: var(--error);
  margin-top: 4px;
  line-height: 1.4;
}

.prop-hint {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  margin-top: 2px;
}

.prop-number-row {
  display: flex;
  flex-direction: column;
}

/* ===== Switch ===== */
.prop-switch {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  user-select: none;
}

.prop-switch input {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}

.prop-switch-slider {
  position: relative;
  width: 36px;
  height: 20px;
  background: var(--border-default);
  border-radius: 10px;
  transition: all var(--transition-fast);
  flex-shrink: 0;
}

.prop-switch-slider::after {
  content: '';
  position: absolute;
  top: 2px;
  left: 2px;
  width: 16px;
  height: 16px;
  background: #fff;
  border-radius: 50%;
  transition: all var(--transition-fast);
}

.prop-switch input:checked + .prop-switch-slider {
  background: var(--primary);
}

.prop-switch input:checked + .prop-switch-slider::after {
  transform: translateX(16px);
}

.prop-switch-label {
  font-size: var(--text-caption);
  color: var(--text-secondary);
}

/* ===== Slider ===== */
.prop-slider-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.prop-slider {
  flex: 1;
  height: 4px;
  -webkit-appearance: none;
  appearance: none;
  background: var(--border-default);
  border-radius: 2px;
  outline: none;
  cursor: pointer;
}

.prop-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 14px;
  height: 14px;
  background: var(--primary);
  border-radius: 50%;
  cursor: pointer;
}

.prop-slider-value {
  font-size: var(--text-caption);
  color: var(--text-primary);
  font-family: var(--font-family-mono);
  min-width: 40px;
  text-align: right;
}

/* ===== Multi-Select ===== */
.prop-multiselect {
  display: flex;
  flex-direction: column;
  gap: 4px;
  max-height: 150px;
  overflow-y: auto;
  padding: 4px;
  background: var(--bg-tertiary);
  border-radius: var(--radius-sm);
}

.prop-multiselect-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 6px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: var(--text-body-sm);
  color: var(--text-primary);
  transition: background var(--transition-fast);
}

.prop-multiselect-item:hover {
  background: var(--bg-hover);
}

.prop-multiselect-item input[type="checkbox"] {
  accent-color: var(--primary);
}

/* ===== 端口列表 ===== */
.prop-ports {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.prop-port-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  background: var(--bg-tertiary);
  border-radius: var(--radius-sm);
  transition: background var(--transition-fast);
}

.prop-port-item:hover {
  background: var(--bg-hover);
}

.port-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.port-name {
  font-size: var(--text-body-sm);
  color: var(--text-primary);
  flex: 1;
}

.port-type-tag {
  font-size: var(--text-caption);
  padding: 1px 6px;
  border-radius: var(--radius-sm);
  font-family: var(--font-family-mono);
  font-weight: var(--font-medium);
}

.prop-empty {
  text-align: center;
  padding: var(--spacing-4);
  color: var(--text-tertiary);
  font-size: var(--text-body-sm);
}

.prop-empty-small {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  padding: 4px 0;
}

/* ===== 操作按钮 ===== */
.prop-actions {
  padding-top: var(--spacing-2);
}

/* ===== 统计 ===== */
.prop-stats {
  display: flex;
  gap: var(--spacing-4);
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.stat-label {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.stat-value {
  font-size: var(--text-h3);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  font-family: var(--font-family-mono);
}

/* ===== 折叠按钮 ===== */
.panel-toggle {
  position: absolute;
  left: 8px;
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