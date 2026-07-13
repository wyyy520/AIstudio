<template>
  <div class="context-section" :class="{ collapsed: isCollapsed }">
    <button v-if="collapsible" class="context-section-header" @click="isCollapsed = !isCollapsed">
      <svg class="context-section-arrow" viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <polyline points="6 9 12 15 18 9" />
      </svg>
      <span class="context-section-title">{{ title }}</span>
    </button>
    <div v-else class="context-section-header context-section-header--static">
      <span class="context-section-title">{{ title }}</span>
    </div>
    <Transition name="collapse">
      <div v-if="!isCollapsed" class="context-section-body">
        <slot />
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

withDefaults(defineProps<{
  title: string
  collapsible?: boolean
  defaultCollapsed?: boolean
}>(), {
  collapsible: true,
  defaultCollapsed: false,
})

const isCollapsed = ref(false)
</script>

<style scoped>
.context-section {
  border-bottom: 1px solid var(--border-subtle);
}

.context-section-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  width: 100%;
  height: 36px;
  padding: 0 var(--spacing-3);
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  transition: background var(--transition-fast);
}

.context-section-header:hover {
  background: var(--bg-hover);
}

.context-section-header--static {
  cursor: default;
  padding-top: var(--spacing-2);
  height: auto;
}

.context-section-header--static:hover {
  background: transparent;
}

.context-section-arrow {
  transition: transform var(--transition-fast);
  flex-shrink: 0;
  color: var(--text-tertiary);
}

.context-section.collapsed .context-section-arrow {
  transform: rotate(-90deg);
}

.context-section-title {
  font-size: var(--text-caption);
  font-weight: var(--font-semibold);
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.context-section-body {
  padding: 0 var(--spacing-3) var(--spacing-3);
}

/* Collapse transition */
.collapse-enter-active,
.collapse-leave-active {
  transition: opacity 200ms ease, height 200ms ease;
  overflow: hidden;
}

.collapse-enter-from,
.collapse-leave-to {
  opacity: 0;
}
</style>
