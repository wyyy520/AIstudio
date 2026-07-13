<template>
  <div class="workflow-console" :class="{ 'is-collapsed': collapsed }">
    <div v-if="!collapsed" class="console-content">
      <div class="console-tabs">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          :class="['console-tab', { active: activeTab === tab.key }]"
          @click="activeTab = tab.key"
        >
          {{ tab.label }}
          <span v-if="tab.badge > 0" class="tab-badge" :class="`tab-badge--${tab.badgeType}`">
            {{ tab.badge }}
          </span>
        </button>
        <div class="console-tab-actions">
          <button
            class="console-action-btn"
            :class="{ active: autoScroll }"
            @click="autoScroll = !autoScroll"
            title="自动滚动"
          >
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="m6 9 6 6 6-6" />
            </svg>
          </button>
          <button class="console-action-btn" @click="clearLogs" title="清除">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M3 6h18 M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6 M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
            </svg>
          </button>
        </div>
      </div>

      <div class="console-body">
        <!-- 日志 Tab -->
        <div v-if="activeTab === 'logs'" class="log-list">
          <div
            v-for="log in logs"
            :key="log.id"
            :class="['log-item', `log-item--${log.level}`]"
          >
            <span class="log-time">{{ log.time }}</span>
            <span class="log-level">{{ log.level.toUpperCase() }}</span>
            <span class="log-message">{{ log.message }}</span>
          </div>
          <div v-if="logs.length === 0" class="log-empty">
            控制台输出将显示在这里
          </div>
        </div>

        <!-- 问题 Tab -->
        <div v-else-if="activeTab === 'problems'" class="problems-list">
          <div v-if="validationErrors.length === 0" class="log-empty">
            <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" class="empty-icon">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14 M22 4 12 14.01l-3-3" />
            </svg>
            <span>工作流配置正常，无错误</span>
          </div>
          <div
            v-for="(error, index) in validationErrors"
            :key="index"
            :class="['problem-item', `problem-item--${error.severity}`]"
          >
            <div class="problem-header">
              <span class="problem-severity" :class="`problem-severity--${error.severity}`">
                {{ error.severity === 'error' ? '✕' : error.severity === 'warning' ? '⚠' : 'ℹ' }}
              </span>
              <span class="problem-node-id">{{ error.nodeId }}</span>
              <span class="problem-type">{{ errorTypeLabel(error.type) }}</span>
            </div>
            <div class="problem-message">{{ error.message }}</div>
            <div v-if="error.details" class="problem-details">{{ error.details }}</div>
            <div v-if="error.autoFix" class="problem-fix">
              <span class="fix-label">建议修复:</span>
              <span class="fix-text">{{ error.autoFix }}</span>
            </div>
          </div>
        </div>

        <!-- AI 建议 Tab -->
        <div v-else-if="activeTab === 'ai'" class="ai-list">
          <div v-if="aiSuggestions.length === 0" class="log-empty">
            <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" class="empty-icon">
              <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" />
            </svg>
            <span>点击 AI Fix 按钮获取智能建议</span>
          </div>
          <div
            v-for="(suggestion, index) in aiSuggestions"
            :key="index"
            class="ai-suggestion"
          >
            <div class="ai-suggestion-header">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" class="ai-icon">
                <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" />
              </svg>
              <span class="ai-suggestion-title">AI 分析 #{{ index + 1 }}</span>
            </div>
            <div class="ai-suggestion-message">{{ suggestion.message }}</div>
            <div v-if="suggestion.actions.length > 0" class="ai-actions">
              <div
                v-for="(action, aIndex) in suggestion.actions"
                :key="aIndex"
                class="ai-action-item"
              >
                <span class="ai-action-num">{{ aIndex + 1 }}</span>
                <span class="ai-action-text">{{ action }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 节点输出 Tab -->
        <div v-else class="log-list">
          <div class="log-empty">暂无输出</div>
        </div>
      </div>
    </div>

    <button
      class="console-toggle"
      @click="collapsed = !collapsed"
      :title="collapsed ? '展开控制台' : '收起控制台'"
    >
      <span class="toggle-label">{{ collapsed ? '控制台' : '' }}</span>
      <svg
        viewBox="0 0 24 24"
        width="14"
        height="14"
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
      >
        <path v-if="collapsed" d="m18 15-6-6-6 6" />
        <path v-else d="m6 9 6 6 6-6" />
      </svg>
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { ValidationError } from '../types/workflow'

interface LogEntry {
  id: number
  time: string
  level: 'info' | 'success' | 'warning' | 'error'
  message: string
}

interface AiSuggestion {
  message: string
  actions: string[]
}

interface Props {
  validationErrors?: ValidationError[]
  aiSuggestions?: AiSuggestion[]
}

const props = withDefaults(defineProps<Props>(), {
  validationErrors: () => [],
  aiSuggestions: () => [],
})

const collapsed = ref(false)
const autoScroll = ref(true)
const activeTab = ref('logs')

const errorCount = computed(() => {
  return props.validationErrors?.filter(e => e.severity === 'error').length || 0
})

const warningCount = computed(() => {
  return props.validationErrors?.filter(e => e.severity === 'warning').length || 0
})

const tabs = computed(() => [
  {
    key: 'logs',
    label: '日志',
    badge: 0,
    badgeType: 'info',
  },
  {
    key: 'problems',
    label: '问题',
    badge: errorCount.value + warningCount.value,
    badgeType: errorCount.value > 0 ? 'error' : 'warning',
  },
  {
    key: 'ai',
    label: 'AI 建议',
    badge: props.aiSuggestions?.length || 0,
    badgeType: 'info',
  },
  {
    key: 'output',
    label: '节点输出',
    badge: 0,
    badgeType: 'info',
  },
])

const logs = ref<LogEntry[]>([
  { id: 1, time: '10:30:01', level: 'info', message: '工作流已加载: 车辆检测工作流' },
  { id: 2, time: '10:30:02', level: 'info', message: '节点初始化完成: 3 个节点' },
  { id: 3, time: '10:30:05', level: 'success', message: '环境检查通过: Python 3.11.8, CUDA 12.4' },
])

function errorTypeLabel(type: string): string {
  const labels: Record<string, string> = {
    'missing-input': '缺少输入',
    'param-error': '参数错误',
    'env-error': '环境错误',
    'connection-error': '连接错误',
    'type-mismatch': '类型不匹配',
  }
  return labels[type] || type
}

function clearLogs(): void {
  logs.value = []
}
</script>

<style scoped>
.workflow-console {
  height: 200px;
  background: var(--bg-secondary);
  border-top: 1px solid var(--border-subtle);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: height var(--transition-normal);
  flex-shrink: 0;
}

.workflow-console.is-collapsed {
  height: 28px;
}

.console-content {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

/* ===== 标签页 ===== */
.console-tabs {
  display: flex;
  align-items: center;
  height: 32px;
  padding: 0 var(--spacing-2);
  border-bottom: 1px solid var(--border-subtle);
  gap: 2px;
}

.console-tab {
  position: relative;
  height: 32px;
  padding: 0 var(--spacing-3);
  border: none;
  border-bottom: 2px solid transparent;
  background: transparent;
  color: var(--text-tertiary);
  font-size: var(--text-caption);
  font-family: var(--font-family-sans);
  cursor: pointer;
  transition: all var(--transition-fast);
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
}

.console-tab:hover {
  color: var(--text-secondary);
}

.console-tab.active {
  color: var(--text-primary);
  border-bottom-color: var(--primary);
}

.tab-badge {
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

.tab-badge--error {
  background: var(--error);
}

.tab-badge--warning {
  background: var(--warning);
}

.tab-badge--info {
  background: var(--info);
}

.console-tab-actions {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 2px;
}

.console-action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-tertiary);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.console-action-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.console-action-btn.active {
  color: var(--primary);
}

/* ===== 内容区 ===== */
.console-body {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-2) 0;
}

.console-body::-webkit-scrollbar {
  width: 4px;
}

.console-body::-webkit-scrollbar-thumb {
  background: var(--border-subtle);
  border-radius: 2px;
}

/* ===== 日志 ===== */
.log-list {
  display: flex;
  flex-direction: column;
}

.log-item {
  display: flex;
  gap: var(--spacing-3);
  padding: 2px var(--spacing-4);
  font-size: var(--text-caption);
  font-family: var(--font-family-mono);
  line-height: 20px;
}

.log-time {
  color: var(--text-tertiary);
  flex-shrink: 0;
}

.log-level {
  color: var(--text-tertiary);
  flex-shrink: 0;
  width: 36px;
  font-size: 10px;
  opacity: 0.7;
}

.log-item--success .log-level {
  color: var(--success);
}

.log-item--warning .log-level {
  color: var(--warning);
}

.log-item--error {
  background: var(--error-bg);
}

.log-item--error .log-level {
  color: var(--error);
}

.log-message {
  color: var(--text-secondary);
  word-break: break-all;
}

.log-item--error .log-message {
  color: var(--error);
}

.log-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-6) var(--spacing-4);
  color: var(--text-tertiary);
  font-size: var(--text-caption);
  text-align: center;
}

.empty-icon {
  opacity: 0.4;
}

/* ===== 问题列表 ===== */
.problems-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.problem-item {
  padding: var(--spacing-2) var(--spacing-4);
  border-left: 3px solid transparent;
  transition: background var(--transition-fast);
}

.problem-item:hover {
  background: var(--bg-hover);
}

.problem-item--error {
  border-left-color: var(--error);
}

.problem-item--warning {
  border-left-color: var(--warning);
}

.problem-item--info {
  border-left-color: var(--info);
}

.problem-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  margin-bottom: 2px;
}

.problem-severity {
  font-size: 12px;
  width: 16px;
  height: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  flex-shrink: 0;
}

.problem-severity--error {
  background: rgba(239, 68, 68, 0.15);
  color: var(--error);
}

.problem-severity--warning {
  background: rgba(251, 191, 36, 0.15);
  color: var(--warning);
}

.problem-severity--info {
  background: rgba(96, 165, 250, 0.15);
  color: var(--info);
}

.problem-node-id {
  font-size: 10px;
  font-family: var(--font-family-mono);
  color: var(--text-tertiary);
  background: var(--bg-tertiary);
  padding: 1px 6px;
  border-radius: var(--radius-sm);
}

.problem-type {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  margin-left: auto;
}

.problem-message {
  font-size: var(--text-caption);
  color: var(--text-primary);
  line-height: 1.5;
  margin-left: 24px;
}

.problem-details {
  font-size: 11px;
  color: var(--text-tertiary);
  margin-left: 24px;
  margin-top: 2px;
  font-family: var(--font-family-mono);
}

.problem-fix {
  margin-left: 24px;
  margin-top: 4px;
  padding: 4px 8px;
  background: var(--bg-tertiary);
  border-radius: var(--radius-sm);
  display: flex;
  gap: 4px;
  font-size: 11px;
}

.fix-label {
  color: var(--warning);
  font-weight: var(--font-semibold);
  flex-shrink: 0;
}

.fix-text {
  color: var(--text-secondary);
}

/* ===== AI 建议 ===== */
.ai-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
  padding: 0 var(--spacing-4);
}

.ai-suggestion {
  padding: var(--spacing-3);
  background: var(--bg-tertiary);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
}

.ai-suggestion-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-2);
}

.ai-icon {
  color: var(--warning);
  flex-shrink: 0;
}

.ai-suggestion-title {
  font-size: var(--text-caption);
  font-weight: var(--font-semibold);
  color: var(--warning);
}

.ai-suggestion-message {
  font-size: var(--text-caption);
  color: var(--text-primary);
  line-height: 1.5;
  margin-bottom: var(--spacing-2);
}

.ai-actions {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.ai-action-item {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-2);
  font-size: var(--text-caption);
  color: var(--text-secondary);
  padding: 4px 8px;
  background: var(--bg-secondary);
  border-radius: var(--radius-sm);
}

.ai-action-num {
  width: 16px;
  height: 16px;
  background: var(--primary);
  color: #fff;
  border-radius: 50%;
  font-size: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-top: 1px;
}

.ai-action-text {
  line-height: 1.5;
}

/* ===== 折叠按钮 ===== */
.console-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-1);
  height: 28px;
  border: none;
  background: var(--bg-secondary);
  border-top: 1px solid var(--border-subtle);
  color: var(--text-tertiary);
  cursor: pointer;
  font-family: var(--font-family-sans);
  font-size: var(--text-caption);
  transition: all var(--transition-fast);
}

.console-toggle:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.toggle-label {
  font-size: 11px;
}
</style>