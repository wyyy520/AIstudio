<template>
  <div class="model-selector" ref="selectorRef">
    <button class="model-selector-trigger" @click="toggle">
      <span class="model-selector-current">{{ currentModelName }}</span>
      <svg class="model-selector-arrow" :class="{ open }" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <polyline points="6 9 12 15 18 9" />
      </svg>
    </button>
    <Transition name="dropdown">
      <div v-if="open" class="model-selector-dropdown">
        <div v-for="provider in providers" :key="provider.id" class="model-selector-group">
          <div class="model-selector-group-label">{{ provider.name }}</div>
          <button
            v-for="model in provider.models"
            :key="model.id"
            class="model-selector-option"
            :class="{ active: model.id === selected }"
            @click="select(model.id)"
          >
            <span class="model-selector-option-name">{{ model.name }}</span>
            <svg v-if="model.id === selected" class="model-selector-check" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="20 6 9 17 4 12" />
            </svg>
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import type { AIProvider } from '../../types'

const props = defineProps<{
  providers: AIProvider[]
  selected: string
}>()

const emit = defineEmits<{
  select: [modelId: string]
}>()

const open = ref(false)
const selectorRef = ref<HTMLElement>()

const currentModelName = computed(() => {
  for (const p of props.providers) {
    const m = p.models.find(m => m.id === props.selected)
    if (m) return m.name
  }
  return 'Select Model'
})

function toggle() {
  open.value = !open.value
}

function select(id: string) {
  emit('select', id)
  open.value = false
}

function handleClickOutside(e: MouseEvent) {
  if (selectorRef.value && !selectorRef.value.contains(e.target as Node)) {
    open.value = false
  }
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', handleClickOutside))
</script>

<style scoped>
.model-selector {
  position: relative;
}

.model-selector-trigger {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  height: 32px;
  padding: 0 var(--spacing-3);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-size: var(--text-body-sm);
  font-family: var(--font-family-sans);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.model-selector-trigger:hover {
  border-color: var(--border-strong);
  background: var(--bg-hover);
}

.model-selector-current {
  white-space: nowrap;
}

.model-selector-arrow {
  transition: transform var(--transition-fast);
  color: var(--text-tertiary);
}

.model-selector-arrow.open {
  transform: rotate(180deg);
}

.model-selector-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  min-width: 220px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  padding: var(--spacing-1);
  z-index: 100;
}

.model-selector-group {
  padding: var(--spacing-1) 0;
}

.model-selector-group + .model-selector-group {
  border-top: 1px solid var(--border-subtle);
}

.model-selector-group-label {
  padding: var(--spacing-1) var(--spacing-2);
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  font-weight: var(--font-semibold);
}

.model-selector-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  height: 32px;
  padding: 0 var(--spacing-2);
  border: none;
  background: transparent;
  color: var(--text-primary);
  font-size: var(--text-body-sm);
  font-family: var(--font-family-sans);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: background var(--transition-fast);
}

.model-selector-option:hover {
  background: var(--bg-hover);
}

.model-selector-option.active {
  background: var(--primary-bg);
  color: var(--primary);
}

.model-selector-check {
  color: var(--primary);
}

/* Dropdown transition */
.dropdown-enter-active,
.dropdown-leave-active {
  transition: opacity 150ms ease, transform 150ms ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
