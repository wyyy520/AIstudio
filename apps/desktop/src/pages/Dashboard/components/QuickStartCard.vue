<template>
  <div class="quick-start-card">
    <h3 class="card-title">快速开始</h3>
    <div class="quick-actions">
      <button
        v-for="action in actions"
        :key="action.key"
        class="quick-action-btn"
        @click="handleAction(action.key)"
      >
        <svg
          class="quick-action-icon"
          viewBox="0 0 24 24"
          width="32"
          height="32"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path :d="action.icon" />
        </svg>
        <span class="quick-action-label">{{ action.label }}</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
interface QuickAction {
  key: string
  label: string
  icon: string
}

const actions: QuickAction[] = [
  {
    key: 'new-project',
    label: '新建项目',
    // Lucide: file-plus
    icon: 'M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z M14 2v6h6 M12 18v-6 M9 15h6',
  },
  {
    key: 'open-project',
    label: '打开项目',
    // Lucide: folder-open
    icon: 'M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z',
  },
  {
    key: 'import-project',
    label: '导入项目',
    // Lucide: download
    icon: 'M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4 M7 10l5 5 5-5 M12 15V3',
  },
  {
    key: 'templates',
    label: '模板中心',
    // Lucide: blocks
    icon: 'M3 3h7v7H3V3zm0 11h7v7H3v-7zm11-11h7v7h-7V3zm0 11h7v7h-7v-7z',
  },
]

const emit = defineEmits<{
  action: [key: string]
}>()

function handleAction(key: string): void {
  emit('action', key)
}
</script>

<style scoped>
.quick-start-card {
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

.quick-actions {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--spacing-3);
  flex: 1;
}

.quick-action-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-2);
  width: 80px;
  height: 80px;
  border: none;
  border-radius: var(--radius-lg);
  background: var(--bg-hover);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
  font-family: var(--font-family-sans);
  justify-self: center;
}

.quick-action-btn:hover {
  background: var(--bg-active);
  color: var(--text-primary);
  transform: translateY(-2px);
  box-shadow: var(--shadow);
}

.quick-action-btn:active {
  transform: scale(0.95);
}

.quick-action-icon {
  flex-shrink: 0;
  opacity: 0.8;
}

.quick-action-btn:hover .quick-action-icon {
  opacity: 1;
  color: var(--primary);
}

.quick-action-label {
  font-size: var(--text-caption);
  font-weight: var(--font-regular);
  line-height: var(--leading-caption);
}
</style>