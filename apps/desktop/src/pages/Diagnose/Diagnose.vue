<template>
  <div class="diagnose-page">
    <div class="page-header">
      <h1>诊断中心</h1>
      <p class="subtitle">日志分析、错误诊断、环境检查与修复建议</p>
    </div>

    <div class="diagnose-grid">
      <!-- 左侧：诊断面板 -->
      <div class="diagnose-sidebar">
        <div class="card">
          <h3>诊断工具</h3>
          <div class="tool-list">
            <button
              :class="['tool-item', { active: activeTool === 'log-analyze' }]"
              @click="activeTool = 'log-analyze'"
            >
              <span class="tool-icon">📊</span>
              <div class="tool-info">
                <span class="tool-name">日志分析</span>
                <span class="tool-desc">分析运行日志中的错误</span>
              </div>
            </button>
            <button
              :class="['tool-item', { active: activeTool === 'env-check' }]"
              @click="activeTool = 'env-check'"
            >
              <span class="tool-icon">🌍</span>
              <div class="tool-info">
                <span class="tool-name">环境检查</span>
                <span class="tool-desc">检测开发环境配置</span>
              </div>
            </button>
            <button
              :class="['tool-item', { active: activeTool === 'workflow-check' }]"
              @click="activeTool = 'workflow-check'"
            >
              <span class="tool-icon">🔍</span>
              <div class="tool-info">
                <span class="tool-name">工作流检查</span>
                <span class="tool-desc">校验工作流合法性</span>
              </div>
            </button>
            <button
              :class="['tool-item', { active: activeTool === 'dependency-check' }]"
              @click="activeTool = 'dependency-check'"
            >
              <span class="tool-icon">📦</span>
              <div class="tool-info">
                <span class="tool-name">依赖检查</span>
                <span class="tool-desc">检查包依赖完整性</span>
              </div>
            </button>
          </div>
        </div>

        <!-- 快速诊断 -->
        <div class="card">
          <h3>快速诊断</h3>
          <button class="btn btn-primary btn-full" @click="runQuickDiagnose">
            🩺 一键诊断
          </button>
          <div v-if="quickResult" class="quick-result" :class="`quick-result--${quickResult.severity}`">
            <div class="quick-summary">
              <span class="quick-icon">{{ severityIcon(quickResult.severity) }}</span>
              <span class="quick-text">发现 {{ quickResult.issues }} 个问题</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 主区域 -->
      <div class="diagnose-main">
        <!-- 环境检查 -->
        <div v-if="activeTool === 'env-check'" class="card">
          <h3>环境检查结果</h3>
          <div class="env-grid">
            <div v-for="env in envItems" :key="env.name" class="env-item">
              <div class="env-header">
                <span :class="['env-status-dot', env.status ? 'dot-ok' : 'dot-error']" />
                <span class="env-name">{{ env.name }}</span>
                <span :class="['env-badge', env.status ? 'badge-ok' : 'badge-error']">
                  {{ env.status ? '✅ 正常' : '❌ 异常' }}
                </span>
              </div>
              <div class="env-detail">
                <span class="env-version">{{ env.version || '未安装' }}</span>
                <span class="env-path">{{ env.path || '-' }}</span>
              </div>
              <div v-if="!env.status" class="env-fix">
                <button class="btn btn-sm btn-primary" @click="fixEnv(env.name)">
                  🔧 修复
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- 工作流检查 -->
        <div v-if="activeTool === 'workflow-check'" class="card">
          <h3>工作流校验结果</h3>
          <div class="check-list">
            <div
              v-for="check in workflowChecks"
              :key="check.id"
              :class="['check-item', `check-item--${check.level}`]"
            >
              <span class="check-icon">
                {{ check.level === 'error' ? '❌' : check.level === 'warning' ? '⚠️' : '✅' }}
              </span>
              <div class="check-content">
                <span class="check-message">{{ check.message }}</span>
                <span v-if="check.detail" class="check-detail">{{ check.detail }}</span>
              </div>
              <button v-if="check.fixable" class="btn btn-sm btn-secondary" @click="applyFix(check.id)">
                修复
              </button>
            </div>
          </div>
        </div>

        <!-- 依赖检查 -->
        <div v-if="activeTool === 'dependency-check'" class="card">
          <h3>依赖包检查</h3>
          <div class="dep-list">
            <div v-for="dep in dependencies" :key="dep.name" class="dep-item">
              <span :class="['dep-status', dep.installed ? 'installed' : 'missing']" />
              <span class="dep-name">{{ dep.name }}</span>
              <span class="dep-required">{{ dep.required }}</span>
              <span class="dep-current">{{ dep.current || '未安装' }}</span>
              <button v-if="!dep.installed" class="btn btn-sm btn-primary" @click="installDep(dep.name)">
                安装
              </button>
            </div>
          </div>
        </div>

        <!-- 日志分析 -->
        <div v-if="activeTool === 'log-analyze'" class="card">
          <h3>日志分析</h3>
          <div class="analyze-input">
            <textarea
              v-model="logInput"
              placeholder="粘贴运行日志或错误信息..."
              rows="6"
              class="input textarea"
            />
            <button class="btn btn-primary" @click="analyzeLog" :disabled="!logInput.trim()">
              分析日志
            </button>
          </div>
          <div v-if="analysisResults.length > 0" class="analysis-results">
            <div
              v-for="(result, idx) in analysisResults"
              :key="idx"
              :class="['analysis-card', `analysis-card--${result.type}`]"
            >
              <div class="analysis-header">
                <span :class="['analysis-badge', `analysis-badge--${result.type}`]">
                  {{ result.type === 'error' ? '错误' : result.type === 'warning' ? '警告' : '信息' }}
                </span>
                <span class="analysis-source">{{ result.source }}</span>
              </div>
              <p class="analysis-message">{{ result.message }}</p>
              <div v-if="result.suggestion" class="analysis-suggestion">
                <strong>建议：</strong>{{ result.suggestion }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

type ToolType = 'log-analyze' | 'env-check' | 'workflow-check' | 'dependency-check'
const activeTool = ref<ToolType>('env-check')
const logInput = ref('')
const quickResult = ref<{ severity: string; issues: number } | null>(null)

// Environment items
const envItems = ref([
  { name: 'Python 3.10+', status: true, version: '3.10.12', path: '/usr/bin/python3' },
  { name: 'CUDA Toolkit', status: true, version: '12.1.105', path: '/usr/local/cuda' },
  { name: 'PyTorch', status: true, version: '2.1.0', path: '-pip' },
  { name: 'MATLAB Runtime', status: false, version: null, path: null },
  { name: 'STM32CubeMX', status: false, version: null, path: null },
  { name: 'ANSYS', status: false, version: null, path: null },
  { name: 'Docker', status: true, version: '24.0.7', path: '/usr/bin/docker' },
  { name: 'Git', status: true, version: '2.40.1', path: '/usr/bin/git' },
])

// Workflow checks
const workflowChecks = ref([
  { id: 'wc1', level: 'ok' as const, message: '所有节点 ID 唯一', detail: '', fixable: false },
  { id: 'wc2', level: 'ok' as const, message: '无循环依赖', detail: '', fixable: false },
  { id: 'wc3', level: 'warning' as const, message: '发现孤立节点: DataLoader_2', detail: '该节点未连接到工作流中', fixable: true },
  { id: 'wc4', level: 'ok' as const, message: '节点参数校验通过', detail: '', fixable: false },
  { id: 'wc5', level: 'error' as const, message: 'Export 节点缺少输出路径', detail: '参数 output_dir 未设置', fixable: true },
  { id: 'wc6', level: 'warning' as const, message: '端口类型不兼容: Image → Tensor', detail: '可能需要插入 Convert 节点', fixable: true },
])

// Dependencies
const dependencies = ref([
  { name: 'torch', required: '>=2.0.0', current: '2.1.0', installed: true },
  { name: 'ultralytics', required: '>=8.0.0', current: '8.1.0', installed: true },
  { name: 'numpy', required: '>=1.24.0', current: '1.26.0', installed: true },
  { name: 'opencv-python', required: '>=4.8.0', current: '4.8.1', installed: true },
  { name: 'matplotlib', required: '>=3.7.0', current: '3.8.2', installed: true },
  { name: 'pandas', required: '>=2.0.0', current: '2.1.4', installed: true },
  { name: 'scikit-learn', required: '>=1.3.0', current: null, installed: false },
  { name: 'tensorboard', required: '>=2.14.0', current: null, installed: false },
])

// Analysis results
interface AnalysisResult {
  type: 'error' | 'warning' | 'info'
  source: string
  message: string
  suggestion?: string
}
const analysisResults = ref<AnalysisResult[]>([])

function severityIcon(severity: string) {
  const icons: Record<string, string> = { critical: '🔴', warning: '🟡', ok: '🟢' }
  return icons[severity] || '⚪'
}

async function runQuickDiagnose() {
  await simulateDelay(1200)
  quickResult.value = { severity: 'warning', issues: 3 }
}

function fixEnv(name: string) {
  alert(`正在修复 ${name} 环境配置...`)
}

function applyFix(id: string) {
  const check = workflowChecks.value.find(c => c.id === id)
  if (check) {
    check.level = 'ok'
    check.fixable = false
    check.detail = '已自动修复'
  }
}

function installDep(name: string) {
  const dep = dependencies.value.find(d => d.name === name)
  if (dep) {
    dep.installed = true
    dep.current = '最新版本'
  }
}

function analyzeLog() {
  if (!logInput.value.trim()) return
  const text = logInput.value.toLowerCase()

  analysisResults.value = []

  if (text.includes('cuda') || text.includes('out of memory')) {
    analysisResults.value.push({
      type: 'error',
      source: 'CUDA Runtime',
      message: '检测到 CUDA 内存不足错误',
      suggestion: '减小 batch_size 或 image_size 参数，或使用 CPU 模式',
    })
  }
  if (text.includes('modulenotfound') || text.includes('no module named')) {
    analysisResults.value.push({
      type: 'error',
      source: 'Python Import',
      message: '缺少 Python 依赖包',
      suggestion: '运行 pip install 安装所需包，或检查虚拟环境配置',
    })
  }
  if (text.includes('file not found') || text.includes('no such file')) {
    analysisResults.value.push({
      type: 'error',
      source: 'File System',
      message: '找不到指定文件或目录',
      suggestion: '检查文件路径是否正确，确认数据集或模型文件存在',
    })
  }
  if (text.includes('warn') || text.includes('deprecated')) {
    analysisResults.value.push({
      type: 'warning',
      source: 'Deprecation',
      message: '使用了已弃用的 API 或参数',
      suggestion: '参考最新文档更新 API 调用',
    })
  }

  if (analysisResults.value.length === 0) {
    analysisResults.value.push({
      type: 'info',
      source: 'Analyzer',
      message: '未检测到已知错误模式，建议查看完整日志',
    })
  }
}
</script>

<script lang="ts">
function simulateDelay(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms))
}
</script>

<style scoped>
.diagnose-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.page-header h1 { font-size: 22px; font-weight: 700; color: var(--text-primary); margin: 0; }
.subtitle { font-size: 13px; color: var(--text-secondary); margin: 4px 0 0; }

.diagnose-grid {
  flex: 1;
  display: grid;
  grid-template-columns: 240px 1fr;
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

.card h3 { font-size: 14px; font-weight: 600; color: var(--text-primary); margin: 0 0 12px; }

/* Tools */
.tool-list { display: flex; flex-direction: column; gap: 4px; }
.tool-item {
  display: flex; align-items: center; gap: 10px;
  padding: 10px; border-radius: var(--radius-xs);
  background: var(--bg-primary); border: 2px solid transparent;
  cursor: pointer; text-align: left; color: var(--text-primary);
  transition: all var(--transition-fast);
}
.tool-item:hover { border-color: var(--border-hover); }
.tool-item.active { border-color: var(--primary); background: var(--bg-active); }
.tool-icon { font-size: 20px; }
.tool-name { font-size: 13px; font-weight: 600; display: block; }
.tool-desc { font-size: 11px; color: var(--text-tertiary); }

/* Buttons */
.btn {
  padding: 8px 16px; border-radius: var(--radius-xs);
  font-size: 13px; font-weight: 500; cursor: pointer;
  transition: all var(--transition-fast); border: none;
  display: flex; align-items: center; justify-content: center; gap: 6px;
}
.btn-full { width: 100%; margin-top: 8px; }
.btn-primary { background: var(--primary); color: #fff; }
.btn-primary:hover:not(:disabled) { opacity: 0.9; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { background: var(--bg-hover); color: var(--text-primary); }
.btn-sm { padding: 4px 10px; font-size: 12px; }

/* Quick result */
.quick-result { margin-top: 10px; padding: 10px; border-radius: var(--radius-xs); }
.quick-result--critical { background: #fbe9e7; border: 1px solid #ffcdd2; }
.quick-result--warning { background: #fff8e1; border: 1px solid #ffecb3; }
.quick-result--ok { background: #e8f5e9; border: 1px solid #c8e6c9; }
.quick-summary { display: flex; align-items: center; gap: 8px; font-size: 13px; }

.diagnose-main { overflow-y: auto; }

/* Environment */
.env-grid { display: flex; flex-direction: column; gap: 6px; }
.env-item {
  padding: 10px 12px; border-radius: var(--radius-xs);
  background: var(--bg-primary); border: 1px solid var(--border-subtle);
}
.env-header { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; }
.env-status-dot { width: 8px; height: 8px; border-radius: 50%; }
.dot-ok { background: #27c93f; }
.dot-error { background: #ff5f56; }
.env-name { font-weight: 600; font-size: 13px; flex: 1; }
.env-badge { font-size: 11px; padding: 1px 8px; border-radius: 10px; }
.badge-ok { background: #e8f5e9; color: #2e7d32; }
.badge-error { background: #fbe9e7; color: #c62828; }
.env-detail { display: flex; gap: 12px; font-size: 11px; color: var(--text-tertiary); padding-left: 16px; }
.env-fix { margin-top: 6px; }

/* Workflow check */
.check-list { display: flex; flex-direction: column; gap: 6px; }
.check-item {
  display: flex; align-items: center; gap: 10px;
  padding: 10px 12px; border-radius: var(--radius-xs);
  font-size: 13px;
}
.check-item--ok { background: var(--bg-primary); }
.check-item--warning { background: #fff8e1; border: 1px solid #ffecb3; }
.check-item--error { background: #fbe9e7; border: 1px solid #ffcdd2; }
.check-icon { font-size: 14px; }
.check-content { flex: 1; }
.check-message { font-weight: 500; }
.check-detail { display: block; font-size: 11px; color: var(--text-tertiary); margin-top: 2px; }

/* Dependencies */
.dep-list { display: flex; flex-direction: column; gap: 4px; }
.dep-item {
  display: flex; align-items: center; gap: 10px;
  padding: 8px 10px; border-radius: var(--radius-xs);
  background: var(--bg-primary); font-size: 12px;
}
.dep-status { width: 8px; height: 8px; border-radius: 50%; }
.installed { background: #27c93f; }
.missing { background: #ff5f56; }
.dep-name { font-weight: 600; font-family: monospace; min-width: 130px; }
.dep-required { color: var(--text-tertiary); }
.dep-current { color: var(--text-secondary); flex: 1; }

/* Log analysis */
.analyze-input { margin-bottom: 16px; }
.textarea { min-height: 100px; resize: vertical; }
.input {
  width: 100%;
  padding: 10px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-xs);
  background: var(--bg-primary);
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
  box-sizing: border-box;
  font-family: 'Consolas', 'Monaco', monospace;
}
.input:focus { border-color: var(--primary); }

.analysis-results { display: flex; flex-direction: column; gap: 8px; margin-top: 12px; }
.analysis-card {
  padding: 12px; border-radius: var(--radius-xs); border-left: 4px solid;
}
.analysis-card--error { border-color: var(--error); background: #fbe9e7; }
.analysis-card--warning { border-color: var(--warning, #ff9800); background: #fff8e1; }
.analysis-card--info { border-color: var(--info, #2196f3); background: #e3f2fd; }
.analysis-header { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; }
.analysis-badge {
  font-size: 11px; padding: 2px 8px; border-radius: 10px; font-weight: 600;
}
.analysis-badge--error { background: #ffcdd2; color: #c62828; }
.analysis-badge--warning { background: #ffecb3; color: #e65100; }
.analysis-badge--info { background: #bbdefb; color: #1565c0; }
.analysis-source { font-size: 12px; color: var(--text-tertiary); }
.analysis-message { font-size: 13px; margin: 4px 0; }
.analysis-suggestion { font-size: 12px; padding: 6px 8px; background: rgba(0,0,0,0.05); border-radius: 4px; margin-top: 6px; }
</style>
