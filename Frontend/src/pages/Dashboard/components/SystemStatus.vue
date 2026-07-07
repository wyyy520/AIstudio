<template>
  <div class="system-status-card">
    <h3 class="card-title">系统状态</h3>
    <div class="status-list">
      <div v-for="item in statusItems" :key="item.key" class="status-item">
        <StatusIndicator :type="item.status" :label="item.label" />
        <span class="status-value">{{ item.value }}</span>
      </div>
    </div>
    <div class="card-footer">
      <AppButton type="secondary" size="small" @click="handleCheck">
        检查环境
      </AppButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import StatusIndicator from '@/components/StatusIndicator/StatusIndicator.vue'
import AppButton from '@/components/AppButton/AppButton.vue'

type StatusType = 'success' | 'warning' | 'error' | 'info' | 'neutral'

interface StatusItem {
  key: string
  label: string
  value: string
  status: StatusType
}

const statusItems: StatusItem[] = [
  {
    key: 'python',
    label: 'Python 环境',
    value: 'Python 3.11.8',
    status: 'success',
  },
  {
    key: 'gpu',
    label: 'GPU 状态',
    value: '空闲 · 显存 2.1GB / 8.0GB',
    status: 'success',
  },
  {
    key: 'cuda',
    label: 'CUDA 版本',
    value: 'CUDA 12.4',
    status: 'success',
  },
  {
    key: 'storage',
    label: '磁盘使用',
    value: '128GB / 512GB',
    status: 'warning',
  },
  {
    key: 'plugins',
    label: '已安装插件',
    value: '12 个已安装',
    status: 'neutral',
  },
]

const emit = defineEmits<{
  'check-environment': []
}>()

function handleCheck(): void {
  emit('check-environment')
}
</script>

<style scoped>
.system-status-card {
  display: flex;
  flex-direction: column;
}

.card-title {
  font-size: var(--text-h3);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  line-height: var(--leading-h3);
  margin-bottom: var(--spacing-4);
}

.status-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
  flex: 1;
}

.status-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-2) 0;
}

.status-value {
  font-size: var(--text-body-sm);
  color: var(--text-secondary);
  line-height: var(--leading-body-sm);
}

.card-footer {
  margin-top: var(--spacing-4);
  padding-top: var(--spacing-4);
  border-top: 1px solid var(--border-subtle);
}
</style>