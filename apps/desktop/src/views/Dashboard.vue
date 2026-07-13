<template>
  <div class="dashboard">
    <div class="dashboard__grid">
      <div class="dashboard__card">
        <h3 class="card-title">快速开始</h3>
        <div class="quick-actions">
          <button class="quick-btn" @click="$router.push('/project')">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--primary)" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            <span>新建项目</span>
          </button>
          <button class="quick-btn" @click="openProject">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--nlp)" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>
            <span>打开项目</span>
          </button>
          <button class="quick-btn" @click="$router.push('/workflow')">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--timeseries)" stroke-width="2"><polyline points="16 3 21 3 21 8"/><line x1="4" y1="20" x2="21" y2="3"/><polyline points="21 16 21 21 16 21"/><line x1="15" y1="15" x2="21" y2="21"/></svg>
            <span>新建工作流</span>
          </button>
          <button class="quick-btn" @click="$router.push('/plugins')">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--vision)" stroke-width="2"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
            <span>浏览插件</span>
          </button>
        </div>
      </div>

      <div class="dashboard__card">
        <h3 class="card-title">最近编辑</h3>
        <div v-if="recentItems.length === 0" class="empty-hint">还没有编辑记录，开始创建第一个项目吧</div>
        <div v-else class="recent-list">
          <div v-for="item in recentItems" :key="item.id" class="recent-item" @click="navigateTo(item)">
            <span class="recent-item__name">{{ item.name }}</span>
            <span class="recent-item__time">{{ item.updated_at }}</span>
          </div>
        </div>
      </div>

      <div class="dashboard__card">
        <h3 class="card-title">系统状态</h3>
        <div class="status-list">
          <div class="status-item">
            <span :class="['status-dot', healthStatus ? 'status-dot--ok' : 'status-dot--err']" />
            <span class="status-label">Backend</span>
            <span class="status-value">{{ healthStatus ? '运行中' : '离线' }}</span>
          </div>
          <div class="status-item">
            <span :class="['status-dot', envStatus.python ? 'status-dot--ok' : 'status-dot--warn']" />
            <span class="status-label">Python</span>
            <span class="status-value">{{ envStatus.python ? '正常' : '未检测' }}</span>
          </div>
          <div class="status-item">
            <span :class="['status-dot', envStatus.cuda ? 'status-dot--ok' : 'status-dot--warn']" />
            <span class="status-label">CUDA</span>
            <span class="status-value">{{ envStatus.cuda ? '已安装' : '未安装' }}</span>
          </div>
          <div class="status-item">
            <span class="status-dot status-dot--ok" />
            <span class="status-label">插件</span>
            <span class="status-value">{{ pluginCount }} 个已安装</span>
          </div>
        </div>
        <button class="check-btn" @click="checkEnvironment">检查环境</button>
      </div>

      <div class="dashboard__card">
        <h3 class="card-title">任务概览</h3>
        <div class="task-stats">
          <div class="task-stat">
            <span class="task-stat__num">{{ taskStats.running }}</span>
            <span class="task-stat__label">运行中</span>
          </div>
          <div class="task-stat">
            <span class="task-stat__num">{{ taskStats.completed }}</span>
            <span class="task-stat__label">已完成</span>
          </div>
          <div class="task-stat">
            <span class="task-stat__num">{{ taskStats.failed }}</span>
            <span class="task-stat__label">失败</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { useRouter } from 'vue-router'
import request from '@/api/request'
import { getPlugins } from '@/api/plugin'
import { getTasks } from '@/api/task'
import { getWorkflows } from '@/api/workflow'
import { getProjects } from '@/api/project'

const router = useRouter()

const healthStatus = ref(false)
const envStatus = reactive({ python: false, cuda: false })
const pluginCount = ref(0)
const taskStats = reactive({ running: 0, completed: 0, failed: 0 })
const recentItems = ref<any[]>([])

async function checkHealth() {
  try {
    await request.get('/health')
    healthStatus.value = true
  } catch {
    healthStatus.value = false
  }
}

async function checkEnvironment() {
  try {
    const res: any = await request.get('/environment')
    envStatus.python = res.data?.python?.installed || false
    envStatus.cuda = res.data?.cuda?.installed || false
  } catch {
    envStatus.python = false
    envStatus.cuda = false
  }
}

async function loadPlugins() {
  try {
    const res: any = await getPlugins()
    pluginCount.value = (res.data || []).length
  } catch {
    pluginCount.value = 0
  }
}

async function loadTasks() {
  try {
    const res: any = await getTasks()
    const tasks = res.data || []
    taskStats.running = tasks.filter((t: any) => t.status === 'running').length
    taskStats.completed = tasks.filter((t: any) => t.status === 'completed' || t.status === 'success').length
    taskStats.failed = tasks.filter((t: any) => t.status === 'failed' || t.status === 'error').length
  } catch {}
}

async function loadRecent() {
  try {
    const [projRes, wfRes]: any[] = await Promise.all([getProjects(), getWorkflows()])
    const projects = (projRes.data || []).map((p: any) => ({ ...p, type: 'project' }))
    const workflows = (wfRes.data || []).map((w: any) => ({ ...w, type: 'workflow' }))
    recentItems.value = [...projects, ...workflows]
      .sort((a: any, b: any) => new Date(b.updated_at || 0).getTime() - new Date(a.updated_at || 0).getTime())
      .slice(0, 5)
  } catch {}
}

function navigateTo(item: any) {
  if (item.type === 'workflow') router.push(`/workflow/${item.id}`)
  else router.push('/project')
}

function openProject() {
  router.push('/project')
}

onMounted(() => {
  checkHealth()
  checkEnvironment()
  loadPlugins()
  loadTasks()
  loadRecent()
})
</script>

<style scoped>
.dashboard {
  padding: 24px;
  height: 100%;
  overflow-y: auto;
}

.dashboard__grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.dashboard__card {
  background: var(--bg-tertiary);
  border-radius: var(--radius-md);
  padding: 20px;
  border: 1px solid var(--border-subtle);
}

.card-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 16px;
}

.quick-actions {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.quick-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 16px 8px;
  border-radius: var(--radius-sm);
  background: var(--bg-secondary);
  color: var(--text-secondary);
  transition: all var(--transition-fast);
}

.quick-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
  transform: translateY(-2px);
}

.quick-btn span {
  font-size: 12px;
}

.empty-hint {
  color: var(--text-tertiary);
  font-size: 13px;
  text-align: center;
  padding: 24px 0;
}

.recent-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.recent-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  border-radius: var(--radius-xs);
  cursor: pointer;
  transition: background var(--transition-fast);
}

.recent-item:hover {
  background: var(--bg-hover);
}

.recent-item__name {
  font-size: 13px;
  color: var(--text-primary);
}

.recent-item__time {
  font-size: 12px;
  color: var(--text-tertiary);
}

.status-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-dot--ok { background: var(--success); }
.status-dot--warn { background: var(--warning); }
.status-dot--err { background: var(--error); }

.status-label {
  font-size: 13px;
  color: var(--text-secondary);
  min-width: 60px;
}

.status-value {
  font-size: 13px;
  color: var(--text-primary);
}

.check-btn {
  margin-top: 16px;
  width: 100%;
  height: 32px;
  border-radius: var(--radius-xs);
  background: var(--bg-secondary);
  color: var(--text-secondary);
  font-size: 13px;
  transition: all var(--transition-fast);
}

.check-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.task-stats {
  display: flex;
  gap: 24px;
  padding: 8px 0;
}

.task-stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.task-stat__num {
  font-size: 28px;
  font-weight: 600;
  color: var(--text-primary);
  font-variant-numeric: tabular-nums;
}

.task-stat__label {
  font-size: 12px;
  color: var(--text-tertiary);
}
</style>