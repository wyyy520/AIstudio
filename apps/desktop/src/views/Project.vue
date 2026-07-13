<template>
  <div class="project">
    <div class="project__toolbar">
      <h3 class="project__title">项目管理</h3>
      <div class="project__actions">
        <AppButton type="primary" size="small" @click="showCreateModal = true">新建项目</AppButton>
      </div>
    </div>

    <div class="project__list">
      <div v-if="projectStore.loading" class="project__loading">加载中...</div>
      <div v-else-if="projectStore.projects.length === 0" class="project__empty">
        <p>还没有项目，点击"新建项目"开始</p>
      </div>
      <div v-else class="project__cards">
        <div v-for="project in projectStore.projects" :key="project.id" class="project-card">
          <div class="project-card__header">
            <span class="project-card__name">{{ project.name }}</span>
            <AppTag :color="statusColor(project.status)" size="small">{{ project.status || 'ready' }}</AppTag>
          </div>
          <p class="project-card__desc">{{ project.description || '暂无描述' }}</p>
          <div class="project-card__meta">
            <span>{{ formatDate(project.created_at) }}</span>
          </div>
          <div class="project-card__actions">
            <AppButton type="ghost" size="small" @click="editProject(project)">编辑</AppButton>
            <AppButton type="ghost" size="small" @click="openWorkflow(project)">工作流</AppButton>
            <AppButton type="danger" size="small" @click="confirmDelete(project)">删除</AppButton>
          </div>
        </div>
      </div>
    </div>

    <AppModal v-model:visible="showConfirmModal" title="确认删除" width="400px">
      <p>确定删除项目 <strong>"{{ deletingProject?.name }}"</strong> 吗？此操作不可撤销。</p>
      <template #footer>
        <AppButton type="secondary" size="small" @click="showConfirmModal = false">取消</AppButton>
        <AppButton type="danger" size="small" :loading="deleting" @click="handleDelete">确认删除</AppButton>
      </template>
    </AppModal>

    <AppModal v-model:visible="showCreateModal" :title="editingProject ? '编辑项目' : '新建项目'" width="480px">
      <div class="form-group">
        <AppInput v-model="form.name" label="项目名称" placeholder="输入项目名称" />
      </div>
      <div class="form-group">
        <AppInput v-model="form.description" label="项目描述" placeholder="输入项目描述" />
      </div>
      <template #footer>
        <AppButton type="secondary" size="small" @click="showCreateModal = false">取消</AppButton>
        <AppButton type="primary" size="small" :loading="submitting" @click="handleSubmit">
          {{ editingProject ? '保存' : '创建' }}
        </AppButton>
      </template>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useProjectStore } from '@/stores/project'
import AppButton from '@/components/AppButton.vue'
import AppTag from '@/components/AppTag.vue'
import AppModal from '@/components/AppModal.vue'
import AppInput from '@/components/AppInput.vue'

const router = useRouter()
const projectStore = useProjectStore()

const showCreateModal = ref(false)
const showConfirmModal = ref(false)
const submitting = ref(false)
const deleting = ref(false)
const editingProject = ref<any>(null)
const deletingProject = ref<any>(null)
const form = reactive({ name: '', description: '' })

function statusColor(status: string) {
  if (status === 'running') return 'warning'
  if (status === 'completed' || status === 'success') return 'success'
  if (status === 'failed' || status === 'error') return 'error'
  return 'default'
}

function formatDate(dateStr: string) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleDateString('zh-CN')
}

function editProject(project: any) {
  editingProject.value = project
  form.name = project.name
  form.description = project.description || ''
  showCreateModal.value = true
}

function openWorkflow(project: any) {
  router.push('/workflow')
}

function confirmDelete(project: any) {
  deletingProject.value = project
  showConfirmModal.value = true
}

async function handleDelete() {
  if (!deletingProject.value) return
  deleting.value = true
  try {
    await projectStore.removeProject(deletingProject.value.id)
    showConfirmModal.value = false
    deletingProject.value = null
  } finally {
    deleting.value = false
  }
}

async function handleSubmit() {
  if (!form.name.trim()) return
  submitting.value = true
  try {
    if (editingProject.value) {
      await projectStore.editProject(editingProject.value.id, { name: form.name, description: form.description })
    } else {
      await projectStore.addProject({ name: form.name, description: form.description })
    }
    showCreateModal.value = false
    form.name = ''
    form.description = ''
    editingProject.value = null
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  projectStore.fetchProjects()
})
</script>

<style scoped>
.project {
  padding: 24px;
  height: 100%;
  overflow-y: auto;
}

.project__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.project__title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
}

.project__loading,
.project__empty {
  text-align: center;
  padding: 48px 0;
  color: var(--text-tertiary);
  font-size: 14px;
}

.project__cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}

.project-card {
  background: var(--bg-tertiary);
  border-radius: var(--radius-md);
  padding: 20px;
  border: 1px solid var(--border-subtle);
  transition: all var(--transition-fast);
}

.project-card:hover {
  border-color: var(--border-default);
  box-shadow: var(--shadow);
}

.project-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.project-card__name {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.project-card__desc {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 12px;
  line-height: 1.5;
}

.project-card__meta {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-bottom: 12px;
}

.project-card__actions {
  display: flex;
  gap: 4px;
}

.form-group {
  margin-bottom: 16px;
}
</style>