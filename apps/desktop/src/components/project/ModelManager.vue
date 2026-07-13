<template>
  <div class="model-manager">
    <div class="manager-header">
      <h2 class="section-title">Models</h2>
    </div>

    <div v-if="models.length === 0" class="empty-state">
      <svg viewBox="0 0 24 24" width="48" height="48" fill="none" stroke="currentColor" stroke-width="1.5">
        <path d="M12 2L2 7l10 5 10-5-10-5zm0 22L2 17l10-5 10 5-10 5z" />
      </svg>
      <span>暂无模型</span>
      <span class="empty-hint">训练模型后将在此显示</span>
    </div>

    <div v-else class="model-grid">
      <ModelCard
        v-for="model in models"
        :key="model.id"
        :model="model"
        @load="$emit('load', $event)"
        @export="$emit('export', $event)"
        @deploy="$emit('deploy', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ProjectModel } from '@/types/project'
import ModelCard from './ModelCard.vue'

defineProps<{
  models: ProjectModel[]
}>()

defineEmits<{
  load: [model: ProjectModel]
  export: [model: ProjectModel]
  deploy: [model: ProjectModel]
}>()
</script>

<style scoped>
.model-manager {
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

.model-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: var(--spacing-4);
}
</style>
