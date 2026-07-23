<template>
  <div class="generator-page">
    <div class="page-header">
      <h1>工程生成器</h1>
      <p class="subtitle">根据执行计划生成 Python、MATLAB、STM32、ANSYS 等真实工程</p>
    </div>

    <div class="generator-grid">
      <!-- 左侧：目标选择 -->
      <div class="generator-sidebar">
        <div class="card">
          <h3>选择生成目标</h3>
          <div class="target-list">
            <button
              v-for="t in store.availableTargets"
              :key="t.id"
              :class="['target-item', { active: selectedTarget === t.id, disabled: store.isGenerating }]"
              :disabled="store.isGenerating"
              @click="selectedTarget = t.id"
            >
              <span class="target-icon">{{ t.icon }}</span>
              <div class="target-info">
                <span class="target-name">{{ t.name }}</span>
                <span class="target-version">{{ t.version }}</span>
              </div>
              <span class="target-desc">{{ t.description }}</span>
            </button>
          </div>
        </div>

        <div class="card">
          <h3>生成控制</h3>
          <div class="control-form">
            <div class="form-group">
              <label>项目 ID</label>
              <input v-model="projectId" type="text" placeholder="输入项目 ID" :disabled="store.isGenerating" class="input" />
            </div>
            <button
              class="btn btn-primary btn-full"
              :disabled="store.isGenerating || !selectedTarget"
              @click="startGenerate"
            >
              <span v-if="store.isGenerating" class="spinner" />
              {{ store.isGenerating ? '生成中...' : `生成 ${targetName}` }}
            </button>
            <button
              v-if="store.status === 'completed' || store.status === 'failed'"
              class="btn btn-secondary btn-full"
              @click="store.reset()"
            >
              重新生成
            </button>
          </div>
        </div>

        <!-- 项目信息 -->
        <div v-if="store.projectInfo" class="card">
          <h3>项目概览</h3>
          <div class="project-summary">
            <div class="summary-row">
              <span class="summary-label">目标平台</span>
              <span class="summary-value">{{ store.projectInfo.target }}</span>
            </div>
            <div class="summary-row">
              <span class="summary-label">输出目录</span>
              <span class="summary-value code">{{ store.projectInfo.outputDir }}</span>
            </div>
            <div class="summary-row">
              <span class="summary-label">文件数量</span>
              <span class="summary-value">{{ store.projectInfo.fileCount }}</span>
            </div>
            <div class="summary-row">
              <span class="summary-label">预估大小</span>
              <span class="summary-value">{{ store.projectInfo.estimatedSize }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 中间：进度和文件列表 -->
      <div class="generator-main">
        <div v-if="store.status !== 'idle'" class="card">
          <div class="progress-section">
            <div class="progress-header">
              <span>生成进度 - {{ statusText }}</span>
              <span class="progress-percent">{{ store.progress }}%</span>
            </div>
            <div class="progress-bar">
              <div
                :class="['progress-fill', store.status === 'failed' ? 'progress-fill--error' : '']"
                :style="{ width: store.progress + '%' }"
              />
            </div>
          </div>
        </div>

        <div v-if="store.error" class="card card--error">
          <div class="error-message">
            <span class="error-icon">❌</span>
            <span>{{ store.error }}</span>
          </div>
        </div>

        <!-- 文件列表 -->
        <div class="card file-card">
          <div class="file-header">
            <h3>生成文件列表</h3>
            <div class="file-stats">
              <span class="stat-badge">代码 {{ store.generatedCodeFileCount }}</span>
              <span class="stat-badge">配置 {{ store.generatedConfigFileCount }}</span>
            </div>
          </div>
          <div class="file-list">
            <div
              v-for="(file, idx) in store.generatedFiles"
              :key="idx"
              :class="['file-item', `file-item--${file.type}`]"
            >
              <span :class="['file-icon', `file-icon--${file.type}`]">{{ fileTypeIcon(file.type) }}</span>
              <span class="file-path">{{ file.path }}</span>
              <span class="file-lang">{{ file.language || '' }}</span>
              <span class="file-size">{{ formatSize(file.size) }}</span>
            </div>
            <div v-if="store.generatedFiles.length === 0" class="file-empty">
              选择目标后开始生成工程文件
            </div>
          </div>
        </div>
      </div>

      <!-- 右侧：生成日志 -->
      <div class="generator-log-panel">
        <div class="card log-card">
          <div class="log-header">
            <h3>生成日志</h3>
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
              <span class="log-message">{{ log.message }}</span>
            </div>
            <div v-if="store.logs.length === 0" class="log-empty">
              点击生成查看实时日志
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useGeneratorStore } from '@/stores/generator'

const store = useGeneratorStore()
const selectedTarget = ref<string | null>(null)
const projectId = ref('')
const logContainer = ref<HTMLElement>()

const targetName = computed(() => {
  const t = store.availableTargets.find(t => t.id === selectedTarget.value)
  return t?.name || ''
})

const statusText = computed(() => {
  const map: Record<string, string> = {
    planning: '规划中...',
    loading_template: '加载模板...',
    generating: '生成文件中...',
    completed: '完成',
    failed: '失败',
  }
  return map[store.status] || store.status
})

function fileTypeIcon(type: string) {
  const icons: Record<string, string> = { code: '📄', config: '⚙️', resource: '📦', template: '📋' }
  return icons[type] || '📁'
}

function formatSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`
  return `${(bytes / 1024).toFixed(1)} KB`
}

function formatTime(ts: string) {
  return new Date(ts).toLocaleTimeString()
}

async function startGenerate() {
  if (!selectedTarget.value) return
  await store.generate(selectedTarget.value, projectId.value || 'demo-project')
}

watch(() => store.logs.length, async () => {
  await nextTick()
  if (logContainer.value) {
    logContainer.value.scrollTop = logContainer.value.scrollHeight
  }
})
</script>

<style scoped>
.generator-page {
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

.generator-grid {
  flex: 1;
  display: grid;
  grid-template-columns: 280px 1fr 280px;
  gap: 16px;
  min-height: 0;
}

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

/* Target list */
.target-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
  max-height: 280px;
  overflow-y: auto;
}

.target-item {
  display: grid;
  grid-template-columns: auto 1fr auto;
  grid-template-rows: auto auto;
  align-items: center;
  gap: 2px 8px;
  padding: 10px;
  border-radius: var(--radius-xs);
  background: var(--bg-primary);
  border: 2px solid transparent;
  cursor: pointer;
  text-align: left;
  transition: all var(--transition-fast);
  color: var(--text-primary);
}

.target-item:hover:not(.disabled) { border-color: var(--border-hover); }
.target-item.active { border-color: var(--primary); background: var(--bg-active); }
.target-item.disabled { opacity: 0.5; cursor: not-allowed; }

.target-icon { font-size: 20px; grid-row: span 2; }
.target-name { font-size: 13px; font-weight: 600; }
.target-version { font-size: 11px; color: var(--text-tertiary); }
.target-desc { font-size: 11px; color: var(--text-tertiary); grid-column: 3; }

/* Controls */
.form-group { margin-bottom: 12px; }
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

.input:focus { border-color: var(--primary); }

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

/* Project summary */
.project-summary { display: flex; flex-direction: column; gap: 6px; }
.summary-row { display: flex; justify-content: space-between; align-items: center; font-size: 12px; }
.summary-label { color: var(--text-tertiary); }
.summary-value { color: var(--text-primary); font-weight: 500; }
.summary-value.code { font-family: monospace; font-size: 11px; }

/* Progress */
.progress-header {
  display: flex; justify-content: space-between;
  margin-bottom: 8px; font-size: 13px; color: var(--text-secondary);
}
.progress-percent { font-weight: 600; color: var(--primary); }
.progress-bar {
  height: 6px; background: var(--bg-hover);
  border-radius: 3px; overflow: hidden;
}
.progress-fill {
  height: 100%; background: var(--primary);
  border-radius: 3px; transition: width 0.3s ease;
}
.progress-fill--error { background: var(--error); }

/* Error */
.error-message {
  display: flex; align-items: center; gap: 8px;
  color: var(--error); font-size: 13px;
}

/* File list */
.generator-main { display: flex; flex-direction: column; min-height: 0; }
.file-card { flex: 1; display: flex; flex-direction: column; min-height: 0; margin-bottom: 0; }
.file-header {
  display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px;
}
.file-header h3 { margin: 0; }
.file-stats { display: flex; gap: 6px; }
.stat-badge {
  font-size: 11px; padding: 2px 8px;
  border-radius: 10px; background: var(--bg-hover); color: var(--text-secondary);
}

.file-list { flex: 1; overflow-y: auto; }
.file-item {
  display: flex; align-items: center; gap: 8px;
  padding: 6px 4px; font-size: 12px;
  border-bottom: 1px solid var(--border-subtle);
}
.file-item:last-child { border-bottom: none; }
.file-item--code { border-left: 3px solid var(--info, #2196f3); }
.file-item--config { border-left: 3px solid var(--warning, #ff9800); }
.file-item--resource { border-left: 3px solid var(--success); }

.file-icon { font-size: 14px; width: 20px; text-align: center; }
.file-path { flex: 1; font-family: monospace; color: var(--text-primary); font-size: 11px; }
.file-lang { font-size: 10px; color: var(--text-tertiary); background: var(--bg-hover); padding: 1px 6px; border-radius: 3px; }
.file-size { font-size: 11px; color: var(--text-tertiary); white-space: nowrap; }

.file-empty {
  text-align: center; color: var(--text-tertiary); padding: 40px 0;
}

/* Log panel */
.generator-log-panel { display: flex; flex-direction: column; }
.log-card { flex: 1; display: flex; flex-direction: column; min-height: 0; margin-bottom: 0; }
.log-header {
  display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px;
}
.log-header h3 { margin: 0; }
.log-count { font-size: 11px; color: var(--text-tertiary); }

.log-list {
  flex: 1; overflow-y: auto;
  font-family: 'Consolas', 'Monaco', monospace; font-size: 12px; line-height: 1.6;
}
.log-entry { display: flex; gap: 6px; padding: 2px 0; align-items: baseline; }
.log-entry--warning { background: var(--warning-bg, #fff8e1); }
.log-entry--error { background: var(--error-bg, #fff0f0); }
.log-time { color: var(--text-tertiary); white-space: nowrap; }
.log-level { font-weight: 600; white-space: nowrap; width: 40px; font-size: 11px; }
.log-level--info { color: var(--info, #2196f3); }
.log-level--warning { color: var(--warning, #ff9800); }
.log-level--error { color: var(--error); }
.log-message { color: var(--text-primary); word-break: break-all; }

.log-empty {
  text-align: center; color: var(--text-tertiary); padding: 40px 0;
}
</style>
