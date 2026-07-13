<template>
  <div class="project-management">
    <ProjectToolbar
      :active-view="activeView"
      @view-change="handleViewChange"
      @new-project="showCreateDialog = true"
    />

    <div class="management-layout">
      <div class="layout-left">
        <div class="explorer-panel">
          <ProjectExplorer @new-project="showCreateDialog = true" />

          <div v-if="!store.currentProject" class="project-list">
            <div class="list-header">
              <span class="list-title">Projects</span>
              <span class="list-count">{{ store.projectCount }}</span>
            </div>
            <div class="list-content">
              <div
                v-for="project in store.sortedProjects"
                :key="project.id"
                class="project-list-item"
                :class="{ active: store.currentProject?.id === project.id }"
                @click="store.selectProject(project)"
              >
                <div class="item-info">
                  <span class="item-name">{{ project.name }}</span>
                  <span class="item-type">{{ getTypeLabel(project.type) }}</span>
                </div>
                <div class="item-status">
                  <span class="status-dot" :class="`status-${project.status}`"></span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="layout-center">
        <div v-if="!store.currentProject" class="welcome-state">
          <div class="welcome-content">
            <svg viewBox="0 0 24 24" width="64" height="64" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" />
            </svg>
            <h2>AI Project Workspace</h2>
            <p>选择一个项目开始工作，或创建新项目</p>
            <button class="welcome-btn" @click="showCreateDialog = true">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M12 5v14M5 12h14" />
              </svg>
              New Project
            </button>
          </div>
        </div>

        <template v-else>
          <ProjectDashboard
            v-if="activeView === 'dashboard'"
            :project="store.currentProject"
            @run-workflow="handleRunWorkflow"
            @train-model="handleTrainModel"
            @open-dataset="activeView = 'datasets'"
            @deploy-model="handleDeployModel"
            @view-experiment="activeView = 'experiments'"
          />

          <WorkflowManager
            v-else-if="activeView === 'workflows'"
            :workflows="store.currentProject.workflows"
            @create="handleCreateWorkflow"
            @open="handleOpenWorkflow"
            @run="handleRunWorkflowItem"
            @clone="handleCloneWorkflow"
          />

          <DatasetManager
            v-else-if="activeView === 'datasets'"
            :datasets="store.currentProject.datasets"
            @create="handleImportDataset"
            @preview="handlePreviewDataset"
            @convert="handleConvertDataset"
            @delete="handleDeleteDataset"
          />

          <ModelManager
            v-else-if="activeView === 'models'"
            :models="store.currentProject.models"
            @load="handleLoadModel"
            @export="handleExportModel"
            @deploy="handleDeployModelItem"
          />

          <ExperimentTable
            v-else-if="activeView === 'experiments'"
            :experiments="store.currentProject.experiments"
            @compare="handleCompareExperiments"
            @detail="handleExperimentDetail"
          />

          <EnvironmentPanel
            v-else-if="activeView === 'environment'"
            :environment="store.currentProject.environment"
            @repair="handleRepairEnvironment"
            @rebuild="handleRebuildEnvironment"
            @install="handleInstallDependencies"
          />
        </template>
      </div>

      <div class="layout-right">
        <AIAssistantPanel @apply-fix="handleApplyFix" />
      </div>
    </div>

    <div class="management-statusbar">
      <ProjectStatus :project="store.currentProject" />
      <div class="statusbar-right">
        <span class="statusbar-item">AIStudio v0.1.0</span>
      </div>
    </div>

    <CreateProjectDialog
      v-model:visible="showCreateDialog"
      :templates="store.templates"
      @create="handleCreateProject"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useProjectStore } from '@/stores/project'
import type { ProjectWorkflow, ProjectDataset, ProjectModel, ProjectExperiment } from '@/types/project'
import ProjectToolbar from '@/components/project/ProjectToolbar.vue'
import ProjectExplorer from '@/components/project/ProjectExplorer.vue'
import ProjectDashboard from '@/components/project/ProjectDashboard.vue'
import WorkflowManager from '@/components/project/WorkflowManager.vue'
import DatasetManager from '@/components/project/DatasetManager.vue'
import ModelManager from '@/components/project/ModelManager.vue'
import ExperimentTable from '@/components/project/ExperimentTable.vue'
import EnvironmentPanel from '@/components/project/EnvironmentPanel.vue'
import AIAssistantPanel from '@/components/project/AIAssistantPanel.vue'
import ProjectStatus from '@/components/project/ProjectStatus.vue'
import CreateProjectDialog from '@/components/project/CreateProjectDialog.vue'

const store = useProjectStore()

const activeView = ref('dashboard')
const showCreateDialog = ref(false)

onMounted(async () => {
  await store.fetchProjects()
  await store.fetchTemplates()
})

function handleViewChange(view: string) {
  activeView.value = view
  store.setExplorerNode(view as 'dashboard' | 'workflows' | 'datasets' | 'models' | 'experiments' | 'environment')
}

function getTypeLabel(type: string): string {
  const map: Record<string, string> = {
    detection: '目标检测',
    classification: '图像分类',
    segmentation: '语义分割',
    timeseries: '时序预测',
    custom: '自定义',
  }
  return map[type] || type
}

async function handleCreateProject(data: {
  name: string
  template: string
  framework: string
  plugins: string[]
}) {
  await store.createNewProject(data)
}

function handleRunWorkflow() {
  if (!store.currentProject) return
  const running = store.currentProject.workflows.find(w => w.status === 'running')
  if (running) return
  const idle = store.currentProject.workflows.find(w => w.status === 'completed' || w.status === 'idle')
  if (idle) {
    store.runProjectWorkflow(idle.id)
  }
}

function showNotImplemented(feature: string) {
  alert(`"${feature}" 功能将在后续版本中实现`)
}

function handleTrainModel() {
  showNotImplemented('模型训练')
}

function handleDeployModel() {
  showNotImplemented('模型部署')
}

function handleCreateWorkflow() {
  showNotImplemented('创建工作流')
}

function handleOpenWorkflow(workflow: ProjectWorkflow) {
  store.selectProject(store.currentProject!)
  // Navigate to workflow editor
  const router = useRouter()
  router.push(`/workflow/${workflow.id}`)
}

function handleRunWorkflowItem(workflow: ProjectWorkflow) {
  store.runProjectWorkflow(workflow.id)
}

function handleCloneWorkflow(workflow: ProjectWorkflow) {
  showNotImplemented('克隆工作流')
}

function handleImportDataset() {
  showNotImplemented('导入数据集')
}

function handlePreviewDataset(dataset: ProjectDataset) {
  showNotImplemented('数据集预览')
}

function handleConvertDataset(dataset: ProjectDataset) {
  showNotImplemented('数据集格式转换')
}

function handleDeleteDataset(dataset: ProjectDataset) {
  if (confirm(`确定删除数据集 "${dataset.name}" 吗？`)) {
    store.deleteDataset(dataset.id)
  }
}

function handleLoadModel(model: ProjectModel) {
  showNotImplemented('加载模型')
}

function handleExportModel(model: ProjectModel) {
  showNotImplemented('导出模型')
}

function handleDeployModelItem(model: ProjectModel) {
  showNotImplemented('模型部署')
}

function handleCompareExperiments() {
  showNotImplemented('实验对比')
}

function handleExperimentDetail(experiment: ProjectExperiment) {
  showNotImplemented('实验详情')
}

async function handleRepairEnvironment() {
  await store.repairProjectEnvironment()
}

function handleRebuildEnvironment() {
  showNotImplemented('重建环境')
}

function handleInstallDependencies() {
  showNotImplemented('安装依赖')
}

function handleApplyFix() {
  showNotImplemented('自动修复')
}
</script>

<style scoped>
.project-management {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.management-layout {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.layout-left {
  width: 280px;
  flex-shrink: 0;
  border-right: 1px solid var(--border-subtle);
  overflow: hidden;
}

.explorer-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.project-list {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-top: 1px solid var(--border-subtle);
}

.list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-3) var(--spacing-4);
}

.list-title {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.list-count {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  background: var(--bg-tertiary);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
}

.list-content {
  flex: 1;
  overflow-y: auto;
  padding: 0 var(--spacing-2) var(--spacing-2);
}

.project-list-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-2) var(--spacing-3);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.project-list-item:hover {
  background: var(--bg-hover);
}

.project-list-item.active {
  background: var(--primary-bg);
}

.item-info {
  flex: 1;
  min-width: 0;
}

.item-name {
  display: block;
  font-size: var(--text-body-sm);
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-type {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.item-status {
  flex-shrink: 0;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-dot.status-active { background: var(--success); }
.status-dot.status-idle { background: var(--neutral); }
.status-dot.status-running { background: var(--info); animation: pulse 2s infinite; }
.status-dot.status-error { background: var(--error); }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.layout-center {
  flex: 1;
  overflow: hidden;
  background: var(--bg-primary);
}

.welcome-state {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.welcome-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-4);
  text-align: center;
  color: var(--text-tertiary);
}

.welcome-content svg {
  color: var(--text-disabled);
}

.welcome-content h2 {
  font-size: var(--text-h2);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.welcome-content p {
  font-size: var(--text-body);
  color: var(--text-secondary);
}

.welcome-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  height: 44px;
  padding: 0 var(--spacing-6);
  border-radius: var(--radius-lg);
  background: var(--primary);
  border: none;
  color: white;
  font-size: var(--text-body);
  font-weight: var(--font-semibold);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.welcome-btn:hover {
  background: var(--primary-hover);
  transform: translateY(-1px);
  box-shadow: var(--shadow-lg);
}

.layout-right {
  width: 320px;
  flex-shrink: 0;
  overflow: hidden;
}

.management-statusbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: var(--statusbar-height);
  padding: 0 var(--spacing-3);
  background: var(--bg-secondary);
  border-top: 1px solid var(--border-subtle);
}

.statusbar-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.statusbar-item {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}
</style>