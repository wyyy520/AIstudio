<template>
  <div class="runtime-page">
    <div class="page-header">
      <h1>运行时监控</h1>
      <p class="subtitle">实时监控工程运行状态、进程管理与终端输出</p>
    </div>

    <div class="runtime-grid">
      <!-- 左侧：运行控制 -->
      <div class="runtime-sidebar">
        <div class="card">
          <h3>运行控制</h3>
          <div class="runtime-actions">
            <button
              class="btn btn-primary btn-full"
              :disabled="runtimeStore.status === 'compiling' || runtimeStore.status === 'running'"
              @click="runProject"
            >
              ▶ 运行工程
            </button>
            <button
              class="btn btn-danger btn-full"
              :disabled="runtimeStore.status !== 'running'"
              @click="stopProject"
            >
              ⏹ 停止运行
            </button>
            <button class="btn btn-secondary btn-full" @click="runtimeStore.reset()">
              ↻ 重置
            </button>
          </div>
        </div>

        <div class="card">
          <h3>运行状态</h3>
          <div class="status-display">
            <div :class="['status-badge', `status-badge--${runtimeStore.status}`]">
              {{ statusLabel }}
            </div>
          </div>
          <div class="status-details">
            <div class="detail-row">
              <span class="detail-label">运行 ID</span>
              <span class="detail-value">{{ runtimeStore.currentRunId || '-' }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">项目 ID</span>
              <span class="detail-value">{{ runtimeStore.currentProjectId || '-' }}</span>
            </div>
            <div class="detail-row" v-if="elapsedTime">
              <span class="detail-label">运行时长</span>
              <span class="detail-value">{{ elapsedTime }}</span>
            </div>
          </div>
        </div>

        <!-- 进度 -->
        <div v-if="runtimeStore.status !== 'idle'" class="card">
          <h3>进度</h3>
          <div class="progress-bar">
            <div
              :class="['progress-fill', runtimeStore.status === 'failed' ? 'progress-fill--error' : '']"
              :style="{ width: runtimeStore.progress + '%' }"
            />
          </div>
          <div class="progress-text">{{ runtimeStore.progress }}%</div>
        </div>
      </div>

      <!-- 主区域：终端输出 -->
      <div class="runtime-main">
        <!-- 终端输出 -->
        <div class="card terminal-card">
          <div class="terminal-header">
            <div class="terminal-title">
              <span class="terminal-dot dot--red" />
              <span class="terminal-dot dot--yellow" />
              <span class="terminal-dot dot--green" />
              <h3>终端输出</h3>
            </div>
            <div class="terminal-actions">
              <button class="terminal-btn" @click="clearOutput" title="清空输出">🗑</button>
              <button class="terminal-btn" @click="autoScroll = !autoScroll" :title="autoScroll ? '停止滚动' : '自动滚动'">
                {{ autoScroll ? '📌' : '📋' }}
              </button>
            </div>
          </div>
          <div class="terminal-output" ref="terminalRef">
            <div
              v-for="(line, idx) in terminalLines"
              :key="idx"
              :class="['terminal-line', `terminal-line--${line.type}`]"
            >
              <span class="terminal-prefix">{{ line.prefix }}</span>
              <span class="terminal-text">{{ line.text }}</span>
            </div>
            <div v-if="terminalLines.length === 0" class="terminal-empty">
              点击"运行工程"开始查看实时输出...
            </div>
          </div>
        </div>

        <!-- 进程列表 -->
        <div class="card process-card">
          <h3>活跃进程</h3>
          <div class="process-list">
            <div
              v-for="proc in processes"
              :key="proc.pid"
              class="process-item"
            >
              <span :class="['process-status', `process-status--${proc.status}`]" />
              <span class="process-pid">PID {{ proc.pid }}</span>
              <span class="process-name">{{ proc.name }}</span>
              <span class="process-cpu">{{ proc.cpu }}% CPU</span>
              <span class="process-mem">{{ proc.mem }} MB</span>
              <button class="process-kill" @click="killProcess(proc.pid)" title="终止进程">✕</button>
            </div>
            <div v-if="processes.length === 0 && runtimeStore.status === 'idle'" class="process-empty">
              当前无运行中的进程
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onUnmounted } from 'vue'
import { useRuntimeStore } from '@/stores/runtime'

const runtimeStore = useRuntimeStore()
const autoScroll = ref(true)
const terminalRef = ref<HTMLElement>()
const startTime = ref<number | null>(null)
const now = ref(Date.now())
let timerInterval: ReturnType<typeof setInterval> | null = null

interface TerminalLine {
  type: 'stdout' | 'stderr' | 'info' | 'system'
  prefix: string
  text: string
}

const terminalLines = ref<TerminalLine[]>([])
const processes = ref<Array<{
  pid: number
  name: string
  status: 'running' | 'sleeping'
  cpu: number
  mem: number
}>>([])

const statusLabel = computed(() => {
  const labels: Record<string, string> = {
    idle: '空闲', compiling: '编译中', running: '运行中',
    completed: '已完成', failed: '失败',
  }
  return labels[runtimeStore.status] || runtimeStore.status
})

const elapsedTime = computed(() => {
  if (!startTime.value) return null
  const diff = Math.floor((now.value - startTime.value) / 1000)
  const mins = Math.floor(diff / 60)
  const secs = diff % 60
  return `${mins}:${secs.toString().padStart(2, '0')}`
})

function addLine(type: TerminalLine['type'], text: string) {
  const prefixes: Record<string, string> = {
    stdout: '→', stderr: '✕', info: 'ℹ', system: '◆',
  }
  terminalLines.value.push({ type, prefix: prefixes[type], text })
}

async function runProject() {
  const projectId = runtimeStore.currentProjectId || 'demo-project'
  terminalLines.value = []
  processes.value = []
  startTime.value = Date.now()
  now.value = Date.now()

  addLine('system', '正在启动工程环境...')
  addLine('system', `项目: ${projectId}`)

  timerInterval = setInterval(() => { now.value = Date.now() }, 1000)

  try {
    await runtimeStore.compileAndRun(projectId)

    // Simulate process list
    if (runtimeStore.status === 'running') {
      processes.value = [
        { pid: 12345, name: 'python main.py', status: 'running', cpu: 45.2, mem: 256.8 },
        { pid: 12346, name: 'yolo_executor', status: 'running', cpu: 32.1, mem: 512.4 },
        { pid: 12347, name: 'dataset_loader', status: 'sleeping', cpu: 0.5, mem: 128.3 },
      ]
    }

    // Simulate terminal output
    addLine('info', 'Workflow 解析完成')
    addLine('info', '执行计划加载: 3 个步骤')
    addLine('stdout', 'Step 1/3: 加载数据集...')
    await simulateDelay(800)
    addLine('stdout', '  数据集路径: ./datasets/road_crack')
    addLine('stdout', '  图片数量: 1200')
    addLine('stdout', 'Step 1/3: ✅ 数据集加载完成 (0.8s)')
    await simulateDelay(400)
    addLine('stdout', 'Step 2/3: 开始训练模型...')
    addLine('stdout', '  Epoch 1/100: loss=2.345, mAP=0.452')
    await simulateDelay(600)
    addLine('stdout', '  Epoch 2/100: loss=1.876, mAP=0.523')
    addLine('stdout', '  ...')
    addLine('stdout', '  Epoch 100/100: loss=0.234, mAP=0.892')
    addLine('stdout', 'Step 2/3: ✅ 模型训练完成 (45.2s)')
    await simulateDelay(400)
    addLine('stdout', 'Step 3/3: 导出模型...')
    addLine('stdout', '  输出: ./outputs/best.pt')
    addLine('stdout', '  格式: PyTorch')
    addLine('stdout', 'Step 3/3: ✅ 模型导出完成 (0.3s)')
    addLine('info', '所有步骤执行完成')
    addLine('system', '运行结束，退出码: 0')

    if (processes.value.length > 0) {
      processes.value.forEach(p => p.status = 'sleeping')
    }
  } catch (e) {
    addLine('stderr', `运行失败: ${e instanceof Error ? e.message : '未知错误'}`)
  } finally {
    if (timerInterval) {
      clearInterval(timerInterval)
      timerInterval = null
    }
  }
}

async function stopProject() {
  addLine('system', '正在终止运行...')
  await runtimeStore.stop()
  processes.value = []
  addLine('system', '运行已停止')
}

function killProcess(pid: number) {
  processes.value = processes.value.filter(p => p.pid !== pid)
  addLine('system', `进程 ${pid} 已终止`)
}

function clearOutput() {
  terminalLines.value = []
}

// Auto-scroll
watch(() => terminalLines.value.length, async () => {
  if (!autoScroll.value) return
  await nextTick()
  if (terminalRef.value) {
    terminalRef.value.scrollTop = terminalRef.value.scrollHeight
  }
})

onUnmounted(() => {
  if (timerInterval) clearInterval(timerInterval)
})
</script>

<script lang="ts">
function simulateDelay(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms))
}
</script>

<style scoped>
.runtime-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.page-header h1 {
  font-size: 22px; font-weight: 700;
  color: var(--text-primary); margin: 0;
}
.subtitle {
  font-size: 13px; color: var(--text-secondary); margin: 4px 0 0;
}

.runtime-grid {
  flex: 1;
  display: grid;
  grid-template-columns: 240px 1fr;
  gap: 16px; min-height: 0;
}

.card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  padding: 16px; margin-bottom: 12px;
}
.card h3 {
  font-size: 14px; font-weight: 600;
  color: var(--text-primary); margin: 0 0 12px;
}

/* Buttons */
.runtime-actions { display: flex; flex-direction: column; gap: 6px; }
.btn {
  padding: 10px 16px; border-radius: var(--radius-xs);
  font-size: 13px; font-weight: 500; cursor: pointer;
  transition: all var(--transition-fast); border: none;
  display: flex; align-items: center; justify-content: center; gap: 6px;
}
.btn-full { width: 100%; }
.btn-primary { background: var(--primary); color: #fff; }
.btn-primary:hover:not(:disabled) { opacity: 0.9; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-danger { background: var(--error); color: #fff; }
.btn-danger:hover:not(:disabled) { opacity: 0.9; }
.btn-danger:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { background: var(--bg-hover); color: var(--text-primary); }

/* Status */
.status-display { text-align: center; margin-bottom: 12px; }
.status-badge {
  display: inline-block; padding: 6px 20px; border-radius: 20px;
  font-size: 14px; font-weight: 600;
}
.status-badge--idle { background: var(--bg-hover); color: var(--text-secondary); }
.status-badge--compiling { background: #e3f2fd; color: #1976d2; }
.status-badge--running { background: #e8f5e9; color: #388e3c; }
.status-badge--completed { background: #e8f5e9; color: #2e7d32; }
.status-badge--failed { background: #fbe9e7; color: #d32f2f; }

.status-details { display: flex; flex-direction: column; gap: 6px; }
.detail-row { display: flex; justify-content: space-between; font-size: 12px; }
.detail-label { color: var(--text-tertiary); }
.detail-value { color: var(--text-primary); font-family: monospace; font-size: 11px; }

/* Progress */
.progress-bar {
  height: 6px; background: var(--bg-hover);
  border-radius: 3px; overflow: hidden; margin-top: 8px;
}
.progress-fill {
  height: 100%; background: var(--primary);
  border-radius: 3px; transition: width 0.3s ease;
}
.progress-fill--error { background: var(--error); }
.progress-text { text-align: center; font-size: 12px; color: var(--text-secondary); margin-top: 4px; }

/* Terminal */
.runtime-main { display: flex; flex-direction: column; min-height: 0; }
.terminal-card { flex: 1; display: flex; flex-direction: column; min-height: 0; margin-bottom: 12px; }
.terminal-header {
  display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px;
}
.terminal-title { display: flex; align-items: center; gap: 6px; }
.terminal-title h3 { margin: 0; }
.terminal-dot {
  width: 10px; height: 10px; border-radius: 50%; flex-shrink: 0;
}
.dot--red { background: #ff5f56; }
.dot--yellow { background: #ffbd2e; }
.dot--green { background: #27c93f; }
.terminal-actions { display: flex; gap: 4px; }
.terminal-btn {
  background: none; border: none; cursor: pointer;
  font-size: 14px; padding: 2px 6px; border-radius: 3px;
}
.terminal-btn:hover { background: var(--bg-hover); }

.terminal-output {
  flex: 1; overflow-y: auto;
  background: #1a1a2e; border-radius: var(--radius-xs);
  padding: 12px; font-family: 'Consolas', 'Monaco', monospace; font-size: 12px;
  line-height: 1.5; min-height: 200px;
}

.terminal-line { display: flex; gap: 8px; align-items: baseline; }
.terminal-line--stdout .terminal-text { color: #e0e0e0; }
.terminal-line--stderr .terminal-text { color: #ff6b6b; }
.terminal-line--info .terminal-text { color: #64b5f6; }
.terminal-line--system .terminal-text { color: #81c784; }
.terminal-prefix { color: #757575; width: 16px; text-align: center; flex-shrink: 0; }
.terminal-text { word-break: break-all; }

.terminal-empty {
  color: #616161; text-align: center; padding: 40px 0;
}

/* Process list */
.process-card { min-height: 120px; }
.process-list { display: flex; flex-direction: column; gap: 4px; }
.process-item {
  display: flex; align-items: center; gap: 10px;
  padding: 8px 10px; border-radius: var(--radius-xs);
  background: var(--bg-primary); font-size: 12px;
}
.process-status {
  width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0;
}
.process-status--running { background: #27c93f; }
.process-status--sleeping { background: #ffbd2e; }
.process-pid { font-family: monospace; color: var(--text-tertiary); }
.process-name { flex: 1; color: var(--text-primary); font-weight: 500; }
.process-cpu, .process-mem { color: var(--text-tertiary); font-size: 11px; white-space: nowrap; }
.process-kill {
  background: none; border: none; color: var(--error); cursor: pointer;
  font-size: 14px; padding: 0 4px; opacity: 0.6;
}
.process-kill:hover { opacity: 1; }
.process-empty {
  text-align: center; color: var(--text-tertiary); padding: 20px 0; font-size: 13px;
}
</style>
