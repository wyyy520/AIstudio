<template>
  <div class="dataset-manager">
    <div class="manager-header">
      <h2 class="section-title">Datasets</h2>
      <button class="create-btn" @click="$emit('create')">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M12 5v14M5 12h14" />
        </svg>
        导入
      </button>
    </div>

    <div v-if="datasets.length === 0" class="empty-state">
      <svg viewBox="0 0 24 24" width="48" height="48" fill="none" stroke="currentColor" stroke-width="1.5">
        <path d="M4 7V4h16v3M9 20h6M12 4v16" />
      </svg>
      <span>暂无数据集</span>
      <span class="empty-hint">点击上方按钮导入数据集</span>
    </div>

    <div v-else class="dataset-grid">
      <DatasetCard
        v-for="dataset in datasets"
        :key="dataset.id"
        :dataset="dataset"
        @preview="$emit('preview', $event)"
        @convert="$emit('convert', $event)"
        @delete="$emit('delete', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ProjectDataset } from '@/types/project'
import DatasetCard from './DatasetCard.vue'

defineProps<{
  datasets: ProjectDataset[]
}>()

defineEmits<{
  create: []
  preview: [dataset: ProjectDataset]
  convert: [dataset: ProjectDataset]
  delete: [dataset: ProjectDataset]
}>()
</script>

<style scoped>
.dataset-manager {
  padding: var(--spacing-6);
  height: 100%;
  overflow-y: auto;
}

.manager-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-4);
}

.section-title {
  font-size: var(--text-h3);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.create-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  height: 32px;
  padding: 0 var(--spacing-3);
  border-radius: var(--radius-md);
  background: var(--primary);
  border: none;
  color: white;
  font-size: var(--text-body-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.create-btn:hover {
  background: var(--primary-hover);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-3);
  height: 300px;
  color: var(--text-tertiary);
}

.empty-state span {
  font-size: var(--text-body);
}

.empty-hint {
  font-size: var(--text-caption);
  color: var(--text-disabled);
}

.dataset-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: var(--spacing-4);
}
</style>
