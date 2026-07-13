<template>
  <div class="workflow-toolbar">
    <div class="toolbar-left">
      <div class="breadcrumb">
        <svg
          class="breadcrumb-icon"
          viewBox="0 0 24 24"
          width="14"
          height="14"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
        >
          <path d="M6 3h3v6H6V3zm0 12h3v6H6v-6zm9-12h3v6h-3V3zm0 12h3v6h-3v-6zm-9 0V9m3 9v-3m3 3v-3m3 3V9" />
        </svg>
        <span class="breadcrumb-text">{{ workflowName }}</span>
      </div>
    </div>

    <div class="toolbar-center">
      <div class="toolbar-actions">
        <AppButton
          type="secondary"
          size="small"
          icon-left="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z M17 21v-8H7v8 M7 3v5h8"
          @click="save"
        >
          保存
        </AppButton>
        <AppButton
          type="primary"
          size="small"
          icon-left="M5 12h14 M12 5l7 7-7 7"
          @click="run"
          :disabled="isRunning"
        >
          运行
        </AppButton>
        <AppButton
          type="secondary"
          size="small"
          icon-left="M10 15v-6 M14 15v-6"
          @click="pause"
          :disabled="!isRunning"
        >
          暂停
        </AppButton>
        <AppButton
          type="danger"
          size="small"
          icon-left="M10 4v16 M4 4h16"
          @click="stop"
          :disabled="!isRunning"
        >
          停止
        </AppButton>

        <div class="toolbar-divider"></div>

        <AppButton
          type="secondary"
          size="small"
          :class="{ 'has-alert': errorCount > 0 }"
          icon-left="M22 11.08V12a10 10 0 1 1-5.93-9.14 M22 4 12 14.01l-3-3"
          @click="validate"
        >
          校验
          <span v-if="errorCount > 0" class="btn-badge btn-badge-error">{{ errorCount }}</span>
        </AppButton>
        <AppButton
          type="warning"
          size="small"
          :class="{ 'has-alert': errorCount > 0 }"
          icon-left="M13 2L3 14h9l-1 8 10-12h-9l1-8z"
          @click="aiFix"
        >
          AI Fix
        </AppButton>
        <AppButton
          type="secondary"
          size="small"
          icon-left="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z M14 2v6h6"
          @click="exportJSON"
        >
          导出
        </AppButton>
      </div>
    </div>

    <div class="toolbar-right">
      <div class="view-controls">
        <button class="view-btn" @click="zoomOut" title="缩小">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="11" cy="11" r="8" />
            <path d="m21 21-4.3-4.3" />
            <path d="M8 11h6" />
          </svg>
        </button>
        <span class="view-zoom">{{ zoomPercent }}%</span>
        <button class="view-btn" @click="zoomIn" title="放大">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="11" cy="11" r="8" />
            <path d="m21 21-4.3-4.3" />
            <path d="M11 8v6" />
            <path d="M8 11h6" />
          </svg>
        </button>
        <button class="view-btn" @click="fitView" title="自适应">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3" />
          </svg>
        </button>
        <div class="view-divider"></div>
        <button class="view-btn" @click="toggleFullscreen" title="全屏">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3" />
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import AppButton from '@/components/AppButton/AppButton.vue'
import type { ValidationError } from '../types/workflow'

interface Props {
  workflowName?: string
  validationErrors?: ValidationError[]
}

const props = withDefaults(defineProps<Props>(), {
  workflowName: '未命名工作流',
  validationErrors: () => [],
})

const isRunning = ref(false)
const zoomPercent = ref(100)

const errorCount = computed(() => {
  return props.validationErrors?.filter(e => e.severity === 'error').length || 0
})

const emits = defineEmits<{
  save: []
  run: []
  pause: []
  stop: []
  zoomIn: []
  zoomOut: []
  fitView: []
  toggleFullscreen: []
  validate: []
  aiFix: []
  exportJSON: []
}>()

function save(): void {
  emits('save')
}

function run(): void {
  isRunning.value = true
  emits('run')
}

function pause(): void {
  isRunning.value = false
  emits('pause')
}

function stop(): void {
  isRunning.value = false
  emits('stop')
}

function validate(): void {
  emits('validate')
}

function aiFix(): void {
  emits('aiFix')
}

function exportJSON(): void {
  emits('exportJSON')
}

function zoomIn(): void {
  zoomPercent.value = Math.min(200, zoomPercent.value + 10)
  emits('zoomIn')
}

function zoomOut(): void {
  zoomPercent.value = Math.max(25, zoomPercent.value - 10)
  emits('zoomOut')
}

function fitView(): void {
  zoomPercent.value = 100
  emits('fitView')
}

function toggleFullscreen(): void {
  emits('toggleFullscreen')
}
</script>

<style scoped>
.workflow-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 48px;
  padding: 0 var(--spacing-4);
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-subtle);
  flex-shrink: 0;
}

/* ===== 左侧面包屑 ===== */
.toolbar-left {
  display: flex;
  align-items: center;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  color: var(--text-secondary);
}

.breadcrumb-icon {
  opacity: 0.7;
}

.breadcrumb-text {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
}

/* ===== 中间操作按钮 ===== */
.toolbar-center {
  display: flex;
  align-items: center;
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.toolbar-divider {
  width: 1px;
  height: 24px;
  background: var(--border-subtle);
  margin: 0 var(--spacing-1);
}

.has-alert {
  position: relative;
}

.btn-badge {
  position: absolute;
  top: -6px;
  right: -6px;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  border-radius: 8px;
  font-size: 10px;
  font-weight: var(--font-bold);
  line-height: 16px;
  text-align: center;
  color: #fff;
}

.btn-badge-error {
  background: var(--error);
}

/* ===== 右侧视图控制 ===== */
.toolbar-right {
  display: flex;
  align-items: center;
}

.view-controls {
  display: flex;
  align-items: center;
  gap: 2px;
}

.view-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.view-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.view-zoom {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  font-family: var(--font-family-mono);
  min-width: 40px;
  text-align: center;
}

.view-divider {
  width: 1px;
  height: 20px;
  background: var(--border-subtle);
  margin: 0 var(--spacing-1);
}
</style>