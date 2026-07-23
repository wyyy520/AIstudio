<template>
  <div class="workflow-toolbar">
    <div class="toolbar-left">
      <div class="breadcrumb">
        <span class="breadcrumb-text">{{ workflowName }}</span>
      </div>
    </div>

    <div class="toolbar-center">
      <div class="toolbar-actions">
        <button class="tb-btn tb-btn-primary" @click="$emit('save')" :disabled="isSaving">
          {{ isSaving ? 'Saving...' : 'Save' }}
        </button>
        <button class="tb-btn tb-btn-run" @click="$emit('run')">
          Run
        </button>
        <div class="tb-divider"></div>
        <button class="tb-btn" @click="$emit('validate')">
          Validate
        </button>
        <button class="tb-btn" @click="$emit('exportJSON')">
          Export
        </button>
      </div>
    </div>

    <div class="toolbar-right">
      <div class="view-controls">
        <button class="view-btn" @click="$emit('zoomOut')" title="Zoom Out">−</button>
        <button class="view-btn" @click="$emit('zoomIn')" title="Zoom In">+</button>
        <button class="view-btn" @click="$emit('fitView')" title="Fit View">⊞</button>
        <button class="view-btn" @click="$emit('toggleFullscreen')" title="Fullscreen">⛶</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  workflowName?: string
  validationErrors?: any[]
  isSaving?: boolean
}

withDefaults(defineProps<Props>(), {
  workflowName: 'Untitled',
  validationErrors: () => [],
  isSaving: false,
})

defineEmits<{
  save: []
  run: []
  zoomIn: []
  zoomOut: []
  fitView: []
  toggleFullscreen: []
  validate: []
  exportJSON: []
}>()
</script>

<style scoped>
.workflow-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 44px;
  padding: 0 var(--spacing-4);
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-subtle);
  flex-shrink: 0;
}

.toolbar-left { display: flex; align-items: center; }
.toolbar-center { display: flex; align-items: center; }
.toolbar-right { display: flex; align-items: center; }

.breadcrumb-text {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.tb-btn {
  height: 30px;
  padding: 0 12px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  font-size: var(--text-body-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}
.tb-btn:hover { background: var(--bg-hover); color: var(--text-primary); }
.tb-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.tb-btn-primary { background: var(--primary); color: white; border-color: var(--primary); }
.tb-btn-primary:hover { background: var(--primary-hover); }
.tb-btn-run { background: var(--success); color: white; border-color: var(--success); }
.tb-btn-run:hover { background: #16a34a; }

.tb-divider {
  width: 1px; height: 20px; background: var(--border-subtle); margin: 0 var(--spacing-1);
}

.view-controls {
  display: flex; align-items: center; gap: 2px;
}
.view-btn {
  display: flex; align-items: center; justify-content: center;
  width: 28px; height: 28px;
  border: none; border-radius: var(--radius-sm);
  background: transparent; color: var(--text-secondary);
  cursor: pointer; font-size: 16px;
}
.view-btn:hover { background: var(--bg-hover); color: var(--text-primary); }
</style>
