<template>
  <div class="project-toolbar">
    <div class="toolbar-left">
      <button
        class="toolbar-btn"
        :class="{ active: activeView === 'dashboard' }"
        @click="$emit('view-change', 'dashboard')"
      >
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M3 3h7v7H3V3zm0 11h7v7H3v-7zm11-11h7v7h-7V3zm0 11h7v7h-7v-7z" />
        </svg>
        概览
      </button>
      <button
        class="toolbar-btn"
        :class="{ active: activeView === 'workflows' }"
        @click="$emit('view-change', 'workflows')"
      >
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M6 3h3v6H6V3zm0 12h3v6H6v-6zm9-12h3v6h-3V3zm0 12h3v6h-3v-6zm-9 0V9m3 9v-3m3 3v-3m3 3V9" />
        </svg>
        Workflows
      </button>
      <button
        class="toolbar-btn"
        :class="{ active: activeView === 'datasets' }"
        @click="$emit('view-change', 'datasets')"
      >
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M4 7V4h16v3M9 20h6M12 4v16" />
        </svg>
        Datasets
      </button>
      <button
        class="toolbar-btn"
        :class="{ active: activeView === 'models' }"
        @click="$emit('view-change', 'models')"
      >
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M12 2L2 7l10 5 10-5-10-5zm0 22L2 17l10-5 10 5-10 5z" />
        </svg>
        Models
      </button>
      <button
        class="toolbar-btn"
        :class="{ active: activeView === 'experiments' }"
        @click="$emit('view-change', 'experiments')"
      >
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M9 3h6v11l-3 3-3-3V3z" />
        </svg>
        Experiments
      </button>
      <button
        class="toolbar-btn"
        :class="{ active: activeView === 'environment' }"
        @click="$emit('view-change', 'environment')"
      >
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
          <rect x="4" y="4" width="16" height="16" rx="2" />
          <path d="M9 9h6v6H9z" />
        </svg>
        Environment
      </button>
    </div>

    <div class="toolbar-right">
      <button class="toolbar-btn toolbar-btn-primary" @click="$emit('new-project')">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M12 5v14M5 12h14" />
        </svg>
        New Project
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  activeView: string
}>()

defineEmits<{
  'view-change': [view: string]
  'new-project': []
}>()
</script>

<style scoped>
.project-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 44px;
  padding: 0 var(--spacing-4);
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-subtle);
}

.toolbar-left,
.toolbar-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
}

.toolbar-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  height: 32px;
  padding: 0 var(--spacing-3);
  border-radius: var(--radius-md);
  background: transparent;
  border: none;
  color: var(--text-secondary);
  font-size: var(--text-body-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.toolbar-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.toolbar-btn.active {
  background: var(--primary-bg);
  color: var(--primary);
}

.toolbar-btn-primary {
  background: var(--primary);
  color: white;
}

.toolbar-btn-primary:hover {
  background: var(--primary-hover);
}
</style>
