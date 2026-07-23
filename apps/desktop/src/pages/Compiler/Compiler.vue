<template>
  <div class="compiler-page">
    <div class="page-header">
      <h1>编译器</h1>
      <p class="subtitle">工作流编译、图优化与执行计划生成</p>
    </div>

    <div class="compiler-grid">
      <!-- 左侧：控制面板 -->
      <div class="compiler-control">
        <div class="card">
          <h3>编译控制</h3>
          <div class="control-form">
            <div class="form-group">
              <label>项目文件</label>
              <div class="file-input-wrapper">
                <input
                  v-model="projectId"
                  type="text"
                  placeholder="输入项目 ID"
                  :disabled="store.isCompiling"
                  class="input"
                />
              </div>
            </div>
            <div class="form-group">
              <label>编译目标</label>
              <select v-model="target" :disabled="store.isCompiling" class="input">
                <option value="">自动检测</option>
                <option value="python">Python</option>
                <option value="matlab">MATLAB</option>
                <option value="stm32">STM32</option>
                <option value="ansys">ANSYS</option>
                <option value="cpp">C++</option>
              </select>
            </div>
            <button
              class="btn btn-primary btn-full"
              :disabled="store.isCompiling || !projectId"
              @click="startCompile"
            >
              <span v-if="store.isCompiling" class="spinner" />
              {{ store.isCompiling ? '编译中...' : '开始编译' }}
            </button>
            <button
              v-if="store.phase === 'completed' || store.phase === 'failed'"
              class="btn btn-secondary btn-full"
              @click="store.reset()"
            >
              重置
            </button>
          </div>
        </div>

        <!-- 编译阶段显示 -->
        <div class="card">
          <h3>编译阶段</h3>
          <div class="phase-list">
            <div
              v-for="p in phases"
              :key="p.key"
              :class="['phase-item', phaseClass(p.key)]"
            >
              <span class="phase-icon">{{ phaseIcon(p.key) }}</span>
              <span class="phase-label">{{ p.label }}</span>
            </div>
          </div>
        </div>

        <!-- 图优化结果 -->
        <div v-if="store.optimizerResult" class="card">
          <h3>图优化结果</h3>
          <div class="optimizer-stats">
            <div class="stat-item">
              <span class="stat-value">{{ store.optimizerResult.deadNodesRemoved }}</span>
              <span class="stat-label">死节点消除</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ store.optimizerResult.invalidEdgesRemoved }}</span>
              <span class="stat-label">无效边清理</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ store.optimizerResult.duplicateEdgesRemoved }}</span>
              <span class="stat-label">重复边去除</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ store.optimizerResult.nodesFused }}</span>
              <span class="stat-label">节点融合</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 中间：进度与日志 -->
      <div class="compiler-main">
        <!-- 进度条 -->
        <div v-if="store.phase !== 'idle'" class="card">
          <div class="progress-section">
            <div class="progress-header">
              <span>编译进度</span>
              <span class="progress-percent">{{ store.progress }}%</span>
            </div>
            <div class="progress-bar">
              <div
                :class="['progress-fill', store.phase === 'failed' ? 'progress-fill--error' : '']"
                :style="{ width: store.progress + '%' }"
              />
            </div>
          </div>
        </div>

        <!-- 错误信息 -->
        <div v-if="store.error" class="card card--error">
          <div class="error-message">
            <span class="error-icon">❌</span>
            <span>{{ store.error }}</span>
          </div>
        </div>

        <!-- 编译日志 -->
        <div class="card log-card">
          <div class="log-header">
            <h3>编译日志</h3>
            <span class="log-count">{{ store.logs.length }} 条</span>
          </div>
          <div class="log-list" ref="logContainer">
            <div
              v-for="(log, idx) in store.logs"
              :key="idx"
              :class="['log-entry', `log-entry--${log.level}`]"
            >
              <span class="log-time">{{ formatTime(log.timestamp) }}</span>
              <span :class="['log-level', `log-level--${log.level}`]">{{ log.level.toUpperCase() }}</span>
              <span class="log-phase">[{{ log.phase }}]</span>
              <span class="log-message">{{ log.message }}</span>
            </div>
            <div v-if="store.logs.length === 0" class="log-empty">
              点击"开始编译"查看编译日志
            </div>
          </div>
        </div>
      </div>

      <!-- 右侧：输出预览 -->
      <div class="compiler-output">
        <div v-if="store.phase === 'completed'" class="card">
          <h3>编译输出</h3>
          <div class="output-tabs">
            <button
              :class="['tab-btn', { active: activeTab === 'ewir' }]"
              @click="activeTab = 'ewir'"
            >EWIR</button>
            <button
              :class="['tab-btn', { active: activeTab === 'plan' }]"
              @click="activeTab = 'plan'"
            >执行计划</button>
            <button
              :class="['tab-btn', { active: activeTab === 'manifest' }]"
              @click="activeTab = 'manifest'"
            >清单</button>
          </div>
          <div class="output-content">
            <div v-if="activeTab === 'ewir'" class="code-preview">
              <pre>// ui.json - 编辑器状态
{
  "viewport": { "zoom": 0.85, "offsetX": 120 },
  "selectedNodeId": "node_002"
}

// workflow.ir.json - 工程中间表示
{
  "version": "1.0.0",
  "project": { "name": "Traffic_Detection" },
  "nodes": [...],
  "edges": [...],
  "domains": ["ai"]
}</pre>
            </div>
            <div v-if="activeTab === 'plan'" class="code-preview">
              <pre>// execution_plan.json
{
  "executionOrder": ["node_001", "node_002"],
  "steps": [
    {
      "nodeId": "node_001",
      "type": "dataset",
      "domain": "ai",
      "dependencies": []
    },
    {
      "nodeId": "node_002",
      "type": "yolo",
      "domain": "ai",
      "dependencies": ["node_001"],
      "generator": "python"
    }
  ]
}</pre>
            </div>
            <div v-if="activeTab === 'manifest'" class="code-preview">
              <pre>// plugin_manifest.json
{
  "plugins": [
    {
      "name": "python-runtime",
      "version": "1.0.0",
      "type": "runtime"
    },
    {
      "name": "yolo-plugin",
      "version": "2.0.0",
      "type": "node"
    }
  ]
}</pre>
            </div>
          </div>
        </div>

        <div v-else-if="store.phase === 'idle'" class="card output-placeholder">
          <div class="placeholder-content">
            <span class="placeholder-icon">📦</span>
            <p>编译完成后，EWIR、执行计划和插件清单将在此显示</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { useCompilerStore, type CompilerPhase } from '@/stores/compiler'

const store = useCompilerStore()
const projectId = ref('')
const target = ref('')
const activeTab = ref<'ewir' | 'plan' | 'manifest'>('ewir')
const logContainer = ref<HTMLElement>()

const phases = [
  { key: 'parsing' as CompilerPhase, label: '1. 解析 Workflow JSON' },
  { key: 'validating' as CompilerPhase, label: '2. 校验节点参数' },
  { key: 'optimizing' as CompilerPhase, label: '3. 图优化' },
  { key: 'building_ewir' as CompilerPhase, label: '4. 构建 EWIR' },
  { key: 'building_plan' as CompilerPhase, label: '5. 生成执行计划' },
  { key: 'generating_manifest' as CompilerPhase, label: '6. 生成插件清单' },
]

const phaseOrder = phases.map(p => p.key)

function phaseClass(key: CompilerPhase) {
  const idx = phaseOrder.indexOf(key)
  const currentIdx = phaseOrder.indexOf(store.phase)
  if (currentIdx > idx) return 'phase-item--done'
  if (currentIdx === idx) return 'phase-item--active'
  return ''
}

function phaseIcon(key: CompilerPhase) {
  const idx = phaseOrder.indexOf(key)
  const currentIdx = phaseOrder.indexOf(store.phase)
  if (currentIdx > idx) return '✅'
  if (currentIdx === idx) return '⏳'
  return '○'
}

function formatTime(ts: string) {
  return new Date(ts).toLocaleTimeString()
}

async function startCompile() {
  if (!projectId.value) return
  await store.compile(projectId.value, target.value || undefined)
}

// Auto-scroll logs
watch(() => store.logs.length, async () => {
  await nextTick()
  if (logContainer.value) {
    logContainer.value.scrollTop = logContainer.value.scrollHeight
  }
})
</script>

<style scoped>
.compiler-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.page-header h1 {
  font-size: 22px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.subtitle {
  font-size: 13px;
  color: var(--text-secondary);
  margin: 4px 0 0;
}

.compiler-grid {
  flex: 1;
  display: grid;
  grid-template-columns: 260px 1fr 320px;
  gap: 16px;
  min-height: 0;
}

/* Card base */
.card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  padding: 16px;
  margin-bottom: 12px;
}

.card--error {
  border-color: var(--error);
  background: var(--error-bg, #fff0f0);
}

.card h3 {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 12px;
}

/* Controls */
.compiler-control {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.form-group {
  margin-bottom: 12px;
}

.form-group label {
  display: block;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 4px;
}

.input {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-xs);
  background: var(--bg-primary);
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
  box-sizing: border-box;
}

.input:focus {
  border-color: var(--primary);
}

.btn {
  padding: 8px 16px;
  border-radius: var(--radius-xs);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  border: none;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.btn-full { width: 100%; margin-top: 8px; }
.btn-primary { background: var(--primary); color: #fff; }
.btn-primary:hover:not(:disabled) { opacity: 0.9; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { background: var(--bg-hover); color: var(--text-primary); margin-top: 4px; }

.spinner {
  width: 14px; height: 14px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }

/* Phase list */
.phase-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.phase-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border-radius: var(--radius-xs);
  font-size: 12px;
  color: var(--text-tertiary);
}

.phase-item--active {
  background: var(--bg-active);
  color: var(--primary);
  font-weight: 600;
}

.phase-item--done {
  color: var(--success);
}

.phase-icon { font-size: 14px; width: 20px; text-align: center; }

/* Optimizer stats */
.optimizer-stats {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}

.stat-item {
  text-align: center;
  padding: 8px;
  background: var(--bg-primary);
  border-radius: var(--radius-xs);
}

.stat-value {
  font-size: 20px;
  font-weight: 700;
  color: var(--primary);
  display: block;
}

.stat-label {
  font-size: 11px;
  color: var(--text-tertiary);
}

/* Progress */
.progress-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 13px;
  color: var(--text-secondary);
}

.progress-percent {
  font-weight: 600;
  color: var(--primary);
}

.progress-bar {
  height: 6px;
  background: var(--bg-hover);
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: var(--primary);
  border-radius: 3px;
  transition: width 0.3s ease;
}

.progress-fill--error {
  background: var(--error);
}

/* Error */
.error-message {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--error);
  font-size: 13px;
}

/* Log */
.compiler-main {
  display: flex;
  flex-direction: column;
  gap: 0;
  min-height: 0;
}

.log-card {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  margin-bottom: 0;
}

.log-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.log-header h3 { margin: 0; }
.log-count { font-size: 11px; color: var(--text-tertiary); }

.log-list {
  flex: 1;
  overflow-y: auto;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  line-height: 1.6;
}

.log-entry {
  display: flex;
  gap: 8px;
  padding: 3px 0;
  align-items: baseline;
}

.log-entry--warning { background: var(--warning-bg, #fff8e1); }
.log-entry--error { background: var(--error-bg, #fff0f0); }

.log-time { color: var(--text-tertiary); white-space: nowrap; }
.log-level { font-weight: 600; white-space: nowrap; width: 45px; }
.log-level--info { color: var(--info, #2196f3); }
.log-level--warning { color: var(--warning, #ff9800); }
.log-level--error { color: var(--error); }
.log-phase { color: var(--text-tertiary); white-space: nowrap; }
.log-message { color: var(--text-primary); word-break: break-all; }

.log-empty {
  text-align: center;
  color: var(--text-tertiary);
  padding: 40px 0;
}

/* Output */
.compiler-output {
  display: flex;
  flex-direction: column;
}

.output-tabs {
  display: flex;
  gap: 4px;
  margin-bottom: 12px;
  border-bottom: 1px solid var(--border-subtle);
  padding-bottom: 8px;
}

.tab-btn {
  padding: 4px 12px;
  font-size: 12px;
  border-radius: var(--radius-xs);
  background: transparent;
  color: var(--text-secondary);
  border: none;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.tab-btn.active {
  background: var(--bg-active);
  color: var(--primary);
  font-weight: 600;
}

.output-content {
  max-height: 450px;
  overflow-y: auto;
}

.code-preview {
  background: var(--bg-primary);
  border-radius: var(--radius-xs);
  padding: 10px;
  overflow-x: auto;
}

.code-preview pre {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 11px;
  line-height: 1.5;
  color: var(--text-secondary);
  margin: 0;
  white-space: pre-wrap;
}

.output-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 200px;
}

.placeholder-content {
  text-align: center;
  color: var(--text-tertiary);
}

.placeholder-icon { font-size: 40px; display: block; margin-bottom: 12px; }
</style>
