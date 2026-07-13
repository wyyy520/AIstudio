<template>
  <div
    :class="['plugin-category', { 'is-active': active, 'is-collapsed': collapsed }]"
    @click="$emit('select')"
  >
    <div class="plugin-category-header">
      <svg
        class="plugin-category-icon"
        viewBox="0 0 24 24"
        width="16"
        height="16"
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path :d="icon" />
      </svg>
      <span class="plugin-category-label">{{ label }}</span>
      <span class="plugin-category-count">{{ count }}</span>
      <svg
        class="plugin-category-chevron"
        viewBox="0 0 24 24"
        width="14"
        height="14"
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path d="M9 18l6-6-6-6" />
      </svg>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  label: string
  icon: string
  count: number
  active?: boolean
  collapsed?: boolean
}

withDefaults(defineProps<Props>(), {
  active: false,
  collapsed: false,
})

defineEmits<{
  select: []
}>()
</script>

<style scoped>
.plugin-category {
  cursor: pointer;
  border-radius: var(--radius-md);
  transition: background var(--transition-fast), color var(--transition-fast);
}

.plugin-category:hover {
  background: var(--bg-hover);
}

.plugin-category.is-active {
  background: var(--bg-active);
}

.plugin-category.is-active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 18px;
  border-radius: 0 2px 2px 0;
  background: var(--primary);
}

.plugin-category-header {
  position: relative;
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  height: 32px;
  padding: 0 var(--spacing-3);
}

.plugin-category-icon {
  flex-shrink: 0;
  opacity: 0.7;
}

.plugin-category.is-active .plugin-category-icon {
  opacity: 1;
}

.plugin-category-label {
  flex: 1;
  font-size: var(--text-body-sm);
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.plugin-category.is-active .plugin-category-label {
  color: var(--text-primary);
}

.plugin-category-count {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  min-width: 18px;
  text-align: center;
}

.plugin-category-chevron {
  flex-shrink: 0;
  color: var(--text-tertiary);
  transition: transform var(--transition-fast);
}

.plugin-category.is-collapsed .plugin-category-chevron {
  transform: rotate(-90deg);
}
</style>